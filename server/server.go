package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"nicograshoff.de/x/optask/config"
	"nicograshoff.de/x/optask/task"
	"strings"
)

type Server struct {
	addr  string
	proj  *config.Project
	tasks *task.Service
}

func NewServer(addr string, proj *config.Project, tasks *task.Service) *Server {
	s := new(Server)
	s.addr = addr
	s.proj = proj
	s.tasks = tasks
	return s
}

func (srv *Server) ListenAndServe() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		template, err := template.ParseFiles("templates/index.html")
		if err != nil {
			log.Fatal(err)
		}

		template.Execute(w, srv.proj)
	})

	http.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		taskID := r.Form.Get("task_id")
		if taskID == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		runID, err := srv.tasks.Run(task.TaskID(taskID))
		if err != nil {
			log.Panic(err)
		}

		http.Redirect(w, r, "/details?task_id="+taskID+"&run_id="+string(runID), http.StatusSeeOther)
	})

	http.HandleFunc("/details", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		taskID := task.TaskID(r.Form.Get("task_id"))
		runID := task.RunID(r.Form.Get("run_id"))

		template, err := template.ParseFiles("templates/details.html")
		if err != nil {
			log.Fatal(err)
		}

		stream, err := srv.tasks.OpenStdout(taskID, runID)
		defer stream.Close()

		task, err := srv.tasks.Task(taskID)

		type viewModel struct {
			Title     string
			CmdLine   string
			Stdout    []string
			Stderr    []string
			IsRunning bool
		}

		cmdLine := fmt.Sprintf("%v %v", task.Command, strings.Join(task.Args, " "))
		isRunning := srv.tasks.IsRunning(taskID, runID)
		vm := &viewModel{task.Name, cmdLine, stream.Lines(), make([]string, 0), isRunning}
		template.Execute(w, vm)
	})

	http.HandleFunc("/output", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		taskID := task.TaskID(r.Form.Get("task_id"))
		runID := task.RunID(r.Form.Get("run_id"))

		stdout, err := srv.tasks.OpenStdout(taskID, runID)
		if err != nil {
			panic(err)
		}
		defer stdout.Close()

		if srv.tasks.IsRunning(taskID, runID) {
			w.Header().Set("Optask-Running", "1")
		} else {
			w.Header().Set("Optask-Running", "0")
		}

		stdout.WriteTo(w)
	})

	log.Printf("Serving project " + srv.proj.Name)
	log.Fatal(http.ListenAndServe(srv.addr, nil))
}
