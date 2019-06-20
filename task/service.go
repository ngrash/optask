// task package contains types and methods to handle asynchronous task running and logging
package task

import (
	"errors"
	"nicograshoff.de/x/optask/archive"
	"nicograshoff.de/x/optask/config"
	"nicograshoff.de/x/optask/exec"
	"nicograshoff.de/x/optask/exec/archivesink"
	"path"
)

type RunID string

type TaskID string

type Service struct {
	tasks  []Task
	runner *exec.Runner
	fs     map[TaskID]*archive.FileSystem
	sinks  map[TaskID]map[RunID]*archivesink.Sink
}

type Task struct {
	ID      TaskID
	Name    string
	Command string
	Args    []string
}

func NewService(pr *config.Project) *Service {
	sv := new(Service)
	sv.fs = make(map[TaskID]*archive.FileSystem)
	sv.sinks = make(map[TaskID]map[RunID]*archivesink.Sink)
	sv.tasks = make([]Task, len(pr.Tasks))
	for i, t := range pr.Tasks {
		sv.fs[TaskID(t.ID)] = archive.NewFileSystem(path.Join(pr.Logs, t.ID))
		sv.tasks[i] = Task{TaskID(t.ID), t.Name, t.Command, t.Args}
	}

	sv.runner = exec.NewRunner()

	return sv
}

func (sv *Service) ListTasks() []Task {
	return sv.tasks
}

func (sv *Service) Task(id TaskID) (Task, error) {
	for _, task := range sv.tasks {
		if task.ID == id {
			return task, nil
		}
	}

	return Task{}, errors.New("No task with ID " + string(id))
}

func (sv *Service) Run(taskID TaskID) (RunID, error) {
	// Prepare sink
	fs := sv.fs[taskID]
	sink := archivesink.NewSink(fs)
	runID := RunID(sink.NodeID())

	if sv.sinks[taskID] == nil {
		sv.sinks[taskID] = make(map[RunID]*archivesink.Sink)
	}
	sv.sinks[taskID][runID] = sink

	// Run task
	task, err := sv.Task(taskID)
	if err != nil {
		return RunID(""), err
	}

	sv.runner.Run(task.Command, task.Args, sink, func() {
		delete(sv.sinks[taskID], runID)
	})

	return runID, nil
}

func (sv *Service) OpenStdout(taskID TaskID, runID RunID) (StdStream, error) {
	stream, err := sv.openStream(taskID, runID, "stdout")
	return stream, err
}

func (sv *Service) OpenStderr(taskID TaskID, runID RunID) (StdStream, error) {
	stream, err := sv.openStream(taskID, runID, "stderr")
	return stream, err
}

func (sv *Service) openStream(taskID TaskID, runID RunID, stream string) (StdStream, error) {
	sink := sv.sinks[taskID][runID]
	if sink != nil {
		switch stream {
		case "stdout":
			return newSliceStream(sink.StdoutLines()), nil
		case "stderr":
			return newSliceStream(sink.StderrLines()), nil
		}

		return nil, errors.New("Invalid stream requests: " + stream)
	} else {
		fs := sv.fs[taskID]
		node := fs.Node(string(runID))
		file, err := fs.Open(node, stream+".txt")
		return newFileStream(file), err
	}
}
