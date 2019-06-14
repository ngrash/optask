package main

import (
	"nicograshoff.de/x/optask/config"
	"nicograshoff.de/x/optask/runner"
	"nicograshoff.de/x/optask/server"
)

func main() {
	project := config.ReadConfig("config.json")
	channel := runner.Start()
	server.ListenAndServe(":8080", project, channel)
}
