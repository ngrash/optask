package main

import (
	"nicograshoff.de/x/optask/archive"
	"nicograshoff.de/x/optask/config"
	"nicograshoff.de/x/optask/runner"
	"nicograshoff.de/x/optask/server"
)

func main() {
	runners := make(map[string]*server.RunnerInfo)
	project := config.ReadConfig("config.json")
	for _, task := range project.Tasks {
		fs := archive.NewFileSystem("logs/" + task.ID)
		sinkFac := runner.NewArchiveSinkFactory(fs)
		runner := runner.NewRunner(sinkFac)

		runners[task.ID] = &server.RunnerInfo{fs, sinkFac, runner}
	}

	server.ListenAndServe(":8080", project, runners)
}
