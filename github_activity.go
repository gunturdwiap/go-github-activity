package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type GithubEvent struct {
	ID         string          `json:"id"`
	Type       GithubEventType `json:"type"`
	Actor      Actor           `json:"actor"`
	Repo       Repo            `json:"repo"`
	RawPayload json.RawMessage `json:"payload"`
	Public     bool            `json:"public"`
	CreatedAt  time.Time       `json:"created_at"`
}

type GithubEventType string

const (
	PushEvent        GithubEventType = "PushEvent"
	IssuesEvent      GithubEventType = "IssuesEvent"
	WatchEvent       GithubEventType = "WatchEvent"
	CreateEvent      GithubEventType = "CreateEvent"
	ForkEvent        GithubEventType = "ForkEvent"
	PullRequestEvent GithubEventType = "PullRequestEvent"
)

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

func (e *GithubEvent) String() string {
	switch e.Type {
	case PushEvent:
		return fmt.Sprintf("Pushed to %s", e.Repo.Name)
	case IssuesEvent:
		payload, err := extractEventPayload[IssuesEventPayload](e)
		if err != nil {
			return err.Error()
		}
		return fmt.Sprintf("%s an issue in %s", capitalize(payload.Action), e.Repo.Name)
	case WatchEvent:
		return fmt.Sprintf("Starred %s", e.Repo.Name)
	case CreateEvent:
		payload, err := extractEventPayload[CreateEventPayload](e)
		if err != nil {
			return err.Error()
		}
		return fmt.Sprintf("Created a %s in %s", payload.RefType, e.Repo.Name)
	case ForkEvent:
		return fmt.Sprintf("Forked %s", e.Repo.Name)
	case PullRequestEvent:
		return fmt.Sprintf("Opened a pull request in %s", e.Repo.Name)
	default:
		return fmt.Sprintf("%s on %s", e.Type, e.Repo.Name)
	}
}

type githubActivity struct {
	username string
	events   []GithubEvent
}

func NewGithubActivity(username string) (*githubActivity, error) {
	g := &githubActivity{
		username: username,
	}

	endpoint := fmt.Sprintf("https://api.github.com/users/%v/events", g.username)

	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf(`user "%v" not found`, g.username)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&g.events); err != nil {
		return nil, err
	}

	return g, nil
}

func (g *githubActivity) DisplayEvents() {
	for _, event := range g.events {
		fmt.Println("-", &event)
	}
}

func extractEventPayload[T any](e *GithubEvent) (*T, error) {
	var payload T
	if err := json.Unmarshal(e.RawPayload, &payload); err != nil {
		return nil, fmt.Errorf("error parsing %T payload: %v", payload, err)
	}

	return &payload, nil
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
