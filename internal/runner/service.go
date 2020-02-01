package runner

import (
	"errors"
	"time"

	"nicograshoff.de/x/optask/internal/db"
	"nicograshoff.de/x/optask/internal/model"
	"nicograshoff.de/x/optask/internal/stdstreams"
)

// Service is the domain context for running tasks
type Service struct {
	project *model.Project
	runner  *runner
	db      *db.Adapter
	runs    map[model.TaskID]map[model.RunID]runData
}

type runData struct {
	r *model.Run
	l *stdstreams.Log
}

func NewService(p *model.Project) *Service {
	r := newRunner()
	db := db.NewAdapter(p.ID+".db", p)

	runs := make(map[model.TaskID]map[model.RunID]runData)
	for _, t := range p.Tasks {
		runs[t.ID] = make(map[model.RunID]runData)
	}

	return &Service{p, r, db, runs}
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

	log := stdstreams.NewLog()

	r := model.Run{Started: time.Now()}
	if err := s.db.CreateRun(tID, &r); err != nil {
		return "", err
	}

	s.runs[tID][r.ID] = runData{&r, log}

	s.runner.Run(task.Cmd, task.Args, log, func(exit int) {
		r.Completed = time.Now()
		r.ExitCode = exit

		if err := s.db.SaveRun(tID, &r); err != nil {
			panic(err)
		}

		if err := s.db.SaveLog(tID, r.ID, log); err != nil {
			panic(err)
		}

		delete(s.runs[tID], r.ID)
	})

	return r.ID, nil
}

func (s *Service) Runs(tID model.TaskID, before model.RunID, count int) ([]*model.Run, error) {
	return s.db.Runs(tID, before, count)
}

func (s *Service) IsRunning(tID model.TaskID, rID model.RunID) bool {
	_, ok := s.runs[tID][rID]
	return ok
}

func (s *Service) LatestRuns() (map[model.TaskID]*model.Run, error) {
	return s.db.LatestRuns()
}

func (s *Service) StdStreams(tID model.TaskID, rID model.RunID) (*stdstreams.Log, error) {
	run, ok := s.runs[tID][rID]
	if ok {
		return run.l, nil
	} else {
		return s.db.Log(tID, rID)
	}
}
