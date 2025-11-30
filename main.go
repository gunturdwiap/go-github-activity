package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}
}

func run() error {
	username := os.Args[1]
	endpoint := fmt.Sprintf("https://api.github.com/users/%v/events", username)

	resp, err := http.Get(endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return fmt.Errorf("user %v not found", username)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	var events []GithubEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return err
	}

	for _, event := range events {
		event.Print()
	}
	return nil
}

type GithubEvent struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Actor      Actor           `json:"actor"`
	Repo       Repo            `json:"repo"`
	RawPayload json.RawMessage `json:"payload"`
	Public     bool            `json:"public"`
	CreatedAt  time.Time       `json:"created_at"`
}

type Repo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Actor struct {
	ID           int    `json:"id"`
	Login        string `json:"login"`
	DisplayLogin string `json:"display_login"`
	GravatarID   string `json:"gravatar_id"`
	URL          string `json:"url"`
	AvatarURL    string `json:"avatar_url"`
}

type CreateEventPayload struct {
	Ref          string `json:"ref"`
	RefType      string `json:"ref_type"`
	FullRef      string `json:"full_ref"`
	MasterBranch string `json:"master_branch"`
	Description  string `json:"description"`
	PusherType   string `json:"pusher_type"`
}

type IssuesEventPayload struct {
	Action    string        `json:"action"`
	Issue     interface{}   `json:"issue"`
	Assignee  interface{}   `json:"assignee"`
	Assignees []interface{} `json:"assignees"`
	Label     interface{}   `json:"label"`
	Labels    []interface{} `json:"labels"`
}

func (e *GithubEvent) Print() {
	var builder strings.Builder
	builder.WriteString("- ")
	switch e.Type {
	case "PushEvent":
		builder.WriteString(fmt.Sprintf("Pushed to %s", e.Repo.Name))
	case "IssuesEvent":
		var payload IssuesEventPayload
		if err := json.Unmarshal(e.RawPayload, &payload); err != nil {
			builder.WriteString("Error parsing IssuesEvent payload")
			break
		}
		builder.WriteString(fmt.Sprintf("%s an issue in %s", capitalize(payload.Action), e.Repo.Name))
	case "WatchEvent":
		builder.WriteString(fmt.Sprintf("Starred %s", e.Repo.Name))
	case "CreateEvent":
		var payload CreateEventPayload
		if err := json.Unmarshal(e.RawPayload, &payload); err != nil {
			builder.WriteString("Error parsing CreateEvent payload")
			break
		}
		builder.WriteString(fmt.Sprintf("Created a %s in %s", payload.RefType, e.Repo.Name))
	case "ForkEvent":
		builder.WriteString(fmt.Sprintf("Forked %s", e.Repo.Name))
	case "PullRequestEvent":
		builder.WriteString(fmt.Sprintf("Opened a pull request in %s", e.Repo.Name))
	default:
		builder.WriteString(fmt.Sprintf("%s on %s", e.Type, e.Repo.Name))
	}

	fmt.Println(builder.String())
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
