package main

import (
	"nicograshoff.de/x/optask/internal/config"
	"nicograshoff.de/x/optask/internal/runner"
	"nicograshoff.de/x/optask/internal/web"
)

func main() {
	project := config.ReadConfig("config.json")
	runner := runner.NewService(project)

	c := &web.Context{project, runner}
	web.ListenAndServe(c, ":8080")
}
