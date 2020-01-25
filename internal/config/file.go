package config

import (
	"encoding/json"
	"log"
	"os"

	"nicograshoff.de/x/optask/internal/model"
)

func ReadConfig(path string) *model.Project {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	project := &model.Project{}
	decoder := json.NewDecoder(file)
	decoder.Decode(project)

	return project
}
