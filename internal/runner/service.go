package runner

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"nicograshoff.de/x/optask/internal/config"
	"nicograshoff.de/x/optask/internal/fs"
)

type RunID string

type TaskID string

type Service struct {
	tasks  []Task
	runner *runner
	fs     map[TaskID]*fs.FileSystem
	sinks  map[TaskID]map[RunID]*sink
}

type Task struct {
	ID      TaskID
	Name    string
	Command string
	Args    []string
}

type StdStream interface {
	WriteTo(w io.Writer)
	Lines() []string
	Close()
}

func NewService(pr *config.Project) *Service {
	sv := new(Service)
	sv.fs = make(map[TaskID]*fs.FileSystem)
	sv.sinks = make(map[TaskID]map[RunID]*sink)
	sv.tasks = make([]Task, len(pr.Tasks))
	for i, t := range pr.Tasks {
		sv.fs[TaskID(t.ID)] = fs.NewFileSystem(path.Join(pr.Logs, t.ID))
		sv.tasks[i] = Task{TaskID(t.ID), t.Name, t.Command, t.Args}
	}

	sv.runner = newRunner()

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
	s := newSink(fs)
	runID := RunID(s.NodeID())

	if sv.sinks[taskID] == nil {
		sv.sinks[taskID] = make(map[RunID]*sink)
	}
	sv.sinks[taskID][runID] = s

	// Run task
	task, err := sv.Task(taskID)
	if err != nil {
		return RunID(""), err
	}

	sv.runner.Run(task.Command, task.Args, s, func() {
		delete(sv.sinks[taskID], runID)
	})

	return runID, nil
}

func (sv *Service) LatestRun(taskID TaskID) RunID {
	node := sv.fs[taskID].LatestNode()
	nodeID := sv.fs[taskID].NodeID(node)
	return RunID(nodeID)
}

func (sv *Service) IsRunning(taskID TaskID, runID RunID) bool {
	return sv.sinks[taskID] != nil && sv.sinks[taskID][runID] != nil
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
			return newSliceStream(sink.stdoutLines()), nil
		case "stderr":
			return newSliceStream(sink.stderrLines()), nil
		}

		return nil, errors.New("Invalid stream requests: " + stream)
	} else {
		fs := sv.fs[taskID]
		node := fs.Node(string(runID))
		file, err := fs.Open(node, stream+".txt")
		return newFileStream(file), err
	}
}

// A StdStream implementation that reads from a buffer
type SliceStream struct {
	buf []string
}

func newSliceStream(buf []string) *SliceStream {
	return &SliceStream{buf}
}

func (s *SliceStream) WriteTo(w io.Writer) {
	for _, line := range s.buf {
		fmt.Fprint(w, line)
	}
}

func (s *SliceStream) Lines() []string {
	return s.buf
}

func (s *SliceStream) Close() {}

// A StdStrean implementation that reads from a file
type FileStream struct {
	file *os.File
}

func newFileStream(f *os.File) *FileStream {
	return &FileStream{f}
}

func (s *FileStream) WriteTo(w io.Writer) {
	r := bufio.NewReader(s.file)
	r.WriteTo(w)
}

func (s *FileStream) Lines() []string {
	buf := make([]string, 0)
	scanner := bufio.NewScanner(s.file)
	for scanner.Scan() {
		line := scanner.Text()
		buf = append(buf, line)
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return buf
}

func (s *FileStream) Close() {
	s.file.Close()
}
