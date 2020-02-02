// Package model contains core domain models.
package model

import "time"

// RunID represents the task-unique identifier of a run.
type RunID string

// TaskID represents the unique identifier of a task.
type TaskID string

// Project represents a project.
type Project struct {
	ID    string
	Name  string
	Tasks []Task
}

// Task represents a task.
type Task struct {
	ID   TaskID
	Name string
	Cmd  string
	Args []string
}

// Run represents a run, i.e. an instance of a task.
type Run struct {
	ID        RunID
	Started   time.Time
	Completed time.Time
	ExitCode  int
}
