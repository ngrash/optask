package main

import (
	"nicograshoff.de/x/optask/config"
	"nicograshoff.de/x/optask/server"
	"nicograshoff.de/x/optask/task"
)

func main() {
	project := config.ReadConfig("config.json")
	tasks := task.NewService(project)
	server := server.NewServer(":8080", project, tasks)
	server.ListenAndServe()
}
