package main

import (
	"log"
	"net/http"

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

	s, err := web.NewServer(project, runner)
	if err != nil {
		log.Fatalf("initializing http server: %v", err)
	}

	log.Fatal(http.ListenAndServe(":8080", s))
}
