package server

import (
	"html/template"
	"log"
	"net/http"
	"nicograshoff.de/x/optask/config"
)

func ListenAndServe(addr string, project *config.Project, channel chan *config.Task) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		template := parseTemplate()
		template.Execute(w, project)
	})

	http.HandleFunc("/trigger", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}

		taskID := r.Form.Get("task")
		if taskID == "" {
			log.Println("Request with empty task parameter")
			w.WriteHeader(http.StatusBadRequest)
		} else {
			task := findTask(taskID, project)
			if task != nil {
				channel <- task
			} else {
				log.Println("No such task: " + taskID + ". Check your config and request.")
				w.WriteHeader(http.StatusNotFound)
			}
		}
	})

	log.Printf("Serving project " + project.Name)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func findTask(id string, project *config.Project) *config.Task {
	for _, task := range project.Tasks {
		if task.ID == id {
			return &task
		}
	}

	return nil
}

func parseTemplate() *template.Template {
	template, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal(err)
	}

	return template
}
