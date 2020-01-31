package model

import "time"

type RunID string

type TaskID string

type Project struct {
	ID    string
	Name  string
	Tasks []Task
}

type Task struct {
	ID   TaskID
	Name string
	Cmd  string
	Args []string
}

type Run struct {
	ID        RunID
	Started   time.Time
	Completed time.Time
	ExitCode  int
}
