package runner

import (
	"errors"
	"fmt"

	"nicograshoff.de/x/optask/internal/db"
	"nicograshoff.de/x/optask/internal/model"
	"nicograshoff.de/x/optask/internal/stdstreams"
)

// Service is the domain context for running tasks
type Service struct {
	project *model.Project
	runner  *runner
	db      *db.Adapter
	logs    map[model.TaskID]map[model.RunID]*stdstreams.Log
}

func NewService(p *model.Project) *Service {
	r := newRunner()
	db := db.NewAdapter(p.ID+".db", p)

	logs := make(map[model.TaskID]map[model.RunID]*stdstreams.Log)
	for _, t := range p.Tasks {
		logs[t.ID] = make(map[model.RunID]*stdstreams.Log)
	}

	return &Service{p, r, db, logs}
}

func (s *Service) ListTasks() []model.Task {
	return s.project.Tasks
}

func (s *Service) Task(tID model.TaskID) (model.Task, error) {
	for _, task := range s.project.Tasks {
		if task.ID == tID {
			return task, nil
		}
	}

	return model.Task{}, errors.New("No task with ID " + string(tID))
}

func (s *Service) Run(tID model.TaskID) (model.RunID, error) {
	task, err := s.Task(tID)
	if err != nil {
		return "", err
	}

	rID, err := s.db.CreateRun(tID)
	if err != nil {
		return "", err
	}

	log := stdstreams.NewLog()
	s.logs[tID][rID] = log

	s.runner.Run(task.Command, task.Args, log, func() {
		if err := s.db.SaveRunLog(tID, rID, log); err != nil {
			fmt.Println(err)
		}

		delete(s.logs[tID], rID)
	})

	return rID, nil
}

func (s *Service) IsRunning(tID model.TaskID, rID model.RunID) bool {
	_, ok := s.logs[tID][rID]
	return ok
}

func (s *Service) LatestRun(tID model.TaskID) (model.RunID, error) {
	return s.db.LatestRun(tID)
}

func (s *Service) StdStreams(tID model.TaskID, rID model.RunID) (*stdstreams.Log, error) {
	log, ok := s.logs[tID][rID]
	if ok {
		return log, nil
	} else {
		return s.db.RunLog(tID, rID)
	}
}
