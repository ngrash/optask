package server

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"nicograshoff.de/x/optask/archive"
	"nicograshoff.de/x/optask/config"
	"nicograshoff.de/x/optask/exec"
	"nicograshoff.de/x/optask/exec/archivesink"
	"strconv"
)

type RunnerInfo struct {
	FS      *archive.FileSystem
	SinkFac *archivesink.Factory
	Runner  *exec.Runner
}

func ListenAndServe(addr string, project *config.Project, runners map[string]*RunnerInfo) {
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
				sinkID := runners[task.ID].Runner.Run(task.Command, task.Args...)
				http.Redirect(w, r, "/listen?job="+sinkID.String()+"&task="+task.ID, http.StatusSeeOther)
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
		r.ParseForm()
		jobID := r.Form.Get("job")
		task := r.Form.Get("task")

		line := 0
		lineParam := r.Form.Get("line")
		if lineParam != "" {
			line, _ = strconv.Atoi(lineParam)
		}

		sinkID := exec.NewSinkID(jobID)
		sink := runners[task].SinkFac.GetOpenSink(sinkID)
		if sink != nil {
			stdout := sink.StdoutLines()[line:]
			for _, line := range stdout {
				fmt.Fprintln(w, line)
			}
		} else {
			node := archive.ParseNode(jobID)
			file, err := runners[task].FS.Open(&node, "stdout.txt")
			if err != nil {
				log.Panic(err)
			}
			fileR := bufio.NewReader(file)
			fileR.WriteTo(w)
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
