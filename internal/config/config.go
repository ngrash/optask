// Package config provides functions to read and validate project configurations.
package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"nicograshoff.de/x/optask/internal/model"
)

// Read reads a project from the given file path and validates it.
func Read(path string) (*model.Project, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Panic(err)
	}

	defer file.Close()

	p := &model.Project{}
	dec := json.NewDecoder(file)
	dec.Decode(p)

	return p, validate(p)
}

func validate(p *model.Project) error {
	hasErr := false
	for i, t := range p.Tasks {
		hasErr = hasErr || logEmpty("ID", string(t.ID), i, t)
		hasErr = hasErr || logEmpty("Name", t.Name, i, t)
		hasErr = hasErr || logEmpty("Cmd", t.Cmd, i, t)
	}

	if hasErr {
		return fmt.Errorf("invalid config")
	}

	return nil
}

func logEmpty(attr string, v string, i int, t model.Task) bool {
	if v == "" {
		j, _ := json.Marshal(t)
		log.Printf("Task (index: %v) missing attribute '%v': %v\n", i, attr, string(j))
		return true
	}

	return false
}
