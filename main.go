package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("please provide a GitHub username")
	}

	username := os.Args[1]
	app, err := NewGithubActivity(username)
	if err != nil {
		return err
	}

	app.DisplayEvents()
	return nil
}
