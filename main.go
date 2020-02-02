package main

import (
	"log"

	"nicograshoff.de/x/optask/internal/config"
	"nicograshoff.de/x/optask/internal/runner"
	"nicograshoff.de/x/optask/internal/web"
)

func main() {
	project, err := config.Read("config.json")
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	runner := runner.NewService(project)

	c := web.Context{Project: project, Runner: runner}
	web.ListenAndServe(&c, ":8080")
}
