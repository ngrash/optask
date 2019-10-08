package main

import (
	"nicograshoff.de/x/optask/internal/config"
	"nicograshoff.de/x/optask/internal/runner"
	"nicograshoff.de/x/optask/internal/web"
)

func main() {
	project := config.ReadConfig("config.json")
	tasks := runner.NewService(project)
	server := web.NewServer(":8080", project, tasks)
	server.ListenAndServe()
}
