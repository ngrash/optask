package main

import (
	"nicograshoff.de/x/optask/config"
	"nicograshoff.de/x/optask/runner"
	"nicograshoff.de/x/optask/server"
)

func main() {
	project := config.ReadConfig("config.json")
	runner.Init()
	server.ListenAndServe(":8080", project)
}
