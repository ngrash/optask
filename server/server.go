package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"nicograshoff.de/x/optask/config"
	"nicograshoff.de/x/optask/runner"
	"strconv"
	"time"
)

func ListenAndServe(addr string, project *config.Project) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		template, err := template.ParseFiles("templates/index.html")
		if err != nil {
			log.Fatal(err)
		}

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
				log.Println("Task: " + task.Name)
				jobID := runner.Run(task)
				http.Redirect(w, r, "/listen?job="+strconv.Itoa(jobID), http.StatusSeeOther)
			} else {
				log.Println("No such task: " + taskID + ". Check your config and request.")
				w.WriteHeader(http.StatusNotFound)
			}
		}
	})

	http.HandleFunc("/job", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		jobIDstr := r.Form.Get("job")
		jobID, err := strconv.Atoi(jobIDstr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		template, err := template.ParseFiles("templates/job.html")
		if err != nil {
			log.Fatal(err)
		}

		template.Execute(w, jobID)
	})

	http.HandleFunc("/listen", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}

		job := r.Form.Get("job")
		if job == "" {
			log.Println("Request with empty job parameter")
		} else {
			line := 0
			lineParam := r.Form.Get("line")
			if lineParam != "" {
				line, _ = strconv.Atoi(lineParam)
			}

			jobID, err := strconv.Atoi(job)
			if err != nil {
				log.Println("Cannot convert job ID to integer: " + job)
			} else {
				if runner.IsRunning(jobID) {
					// wait a bit to collect some lines
					time.Sleep(250 * time.Millisecond)
					stdoutLines := runner.GetStdout(jobID, line)
					for _, stdoutLine := range stdoutLines {
						fmt.Fprint(w, stdoutLine)
					}
				} else {
					log.Println("Job " + job + " not running")
					w.WriteHeader(http.StatusNotFound)
				}
			}
		}

		log.Println("Listen")
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
