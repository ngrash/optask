package main

import (
	"nicograshoff.de/x/optask/archive"
	"nicograshoff.de/x/optask/config"
	"nicograshoff.de/x/optask/runner"
	"nicograshoff.de/x/optask/server"
)

func main() {
	logs := make(map[string]*archive.FileSystem)
	project := config.ReadConfig("config.json")
	for _, task := range project.Tasks {
		logs[task.ID] = archive.NewFileSystem("logs/" + task.ID)
	}

	runner.Init(logs)
	server.ListenAndServe(":8080", project)
}
