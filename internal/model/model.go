package model

type RunID string

type TaskID string

type Project struct {
	ID    string
	Name  string
	Tasks []Task
}

type Task struct {
	ID      TaskID
	Name    string
	Command string
	Args    []string
}
