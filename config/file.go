package config

import (
	"encoding/json"
	"log"
	"os"
)

type Project struct {
	ID    string
	Name  string
	Tasks []Task
}

type Task struct {
	ID      string
	Name    string
	Command string
	Args    []string
}

func ReadConfig(path string) *Project {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	project := &Project{}
	decoder := json.NewDecoder(file)
	decoder.Decode(project)

	return project
}
