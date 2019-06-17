package main

import (
	"nicograshoff.de/x/optask/archive"
	"nicograshoff.de/x/optask/config"
	"nicograshoff.de/x/optask/exec"
	"nicograshoff.de/x/optask/exec/archivesink"
	"nicograshoff.de/x/optask/server"
)

func main() {
	project := config.ReadConfig("config.json")

	runners := make(map[string]*server.RunnerInfo)
	for _, task := range project.Tasks {
		fs := archive.NewFileSystem("logs/" + task.ID)
		sinkFac := archivesink.NewFactory(fs)
		runner := exec.NewRunner(sinkFac)
		runners[task.ID] = &server.RunnerInfo{fs, sinkFac, runner}
	}

	server.ListenAndServe(":8080", project, runners)
}
