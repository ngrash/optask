package config

import (
	"encoding/json"
	"errors"
	"log"
	"os"

	"nicograshoff.de/x/optask/internal/model"
)

var errInvalidConfig = errors.New("invalid config")

func ReadConfig(path string) (*model.Project, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Panic(err)
	}

	defer file.Close()

	project := &model.Project{}
	decoder := json.NewDecoder(file)
	decoder.Decode(project)

	err = validateConfig(project)

	return project, err
}

func validateConfig(p *model.Project) error {
	hasErr := false
	for i, t := range p.Tasks {
		if t.ID == "" {
			hasErr = logMissing("ID", i, t)
		}

		if t.Name == "" {
			hasErr = logMissing("Name", i, t)
		}

		if t.Cmd == "" {
			hasErr = logMissing("Cmd", i, t)
		}
	}

	if hasErr {
		return errInvalidConfig
	} else {
		return nil
	}
}

func logMissing(attr string, i int, t model.Task) bool {
	j, _ := json.Marshal(t)
	log.Printf("Task (index: %v) missing attribute '%v': %v\n", i, attr, string(j))
	return true
}
