package main

import (
	"log"
	"net/http"

	"github.com/ngrash/optask/internal/config"
	"github.com/ngrash/optask/internal/runner"
	"github.com/ngrash/optask/internal/web"
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
