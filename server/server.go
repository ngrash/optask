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
)

type Server struct {
	proj     *config.Project
	runner   *exec.Runner
	archives map[string]*archive.FileSystem
	sinks    map[string]map[string]*archivesink.Sink
	addr     string
}

func NewServer(addr string, proj *config.Project) *Server {
	s := new(Server)
	s.addr = addr
	s.proj = proj
	s.runner = exec.NewRunner()
	s.archives = make(map[string]*archive.FileSystem)
	s.sinks = make(map[string]map[string]*archivesink.Sink)
	for _, task := range proj.Tasks {
		s.archives[task.ID] = archive.NewFileSystem("logs/" + task.ID)
		s.sinks[task.ID] = make(map[string]*archivesink.Sink)
	}
	return s
}

func (s *Server) ListenAndServe() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		template, err := template.ParseFiles("templates/index.html")
		if err != nil {
			log.Fatal(err)
		}

		template.Execute(w, s.proj)
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
			task := findTask(taskID, s.proj)
			if task != nil {
				log.Println("Task: " + task.Name)
				sinkID := s.run(task)
				http.Redirect(w, r, "/output?job="+sinkID+"&task="+task.ID, http.StatusSeeOther)
			} else {
				log.Println("No such task: " + taskID + ". Check your config and request.")
				w.WriteHeader(http.StatusNotFound)
			}
		}
	})

	http.HandleFunc("/output", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		sinkID := r.Form.Get("job")
		task := r.Form.Get("task")

		fs := s.archives[task]
		sink := s.sinks[task][sinkID]
		if sink != nil {
			log.Println("Serving from memory")
			stdout := sink.StdoutLines()
			for _, line := range stdout {
				fmt.Fprint(w, line)
			}
		} else {
			log.Println("Serving from filesystem")
			node := fs.Node(sinkID)
			file, err := fs.Open(node, "stdout.txt")
			if err != nil {
				log.Panic(err)
			}

			bufio.NewReader(file).WriteTo(w)
		}
	})

	log.Printf("Serving project " + s.proj.Name)
	log.Fatal(http.ListenAndServe(s.addr, nil))
}

func (s *Server) run(task *config.Task) string {
	fs := s.archives[task.ID]
	sink := archivesink.NewSink(fs)
	sinkID := sink.NodeID()
	s.sinks[task.ID][sinkID] = sink

	s.runner.Run(task.Command, task.Args, sink, func() {
		delete(s.sinks[task.ID], sinkID)
	})

	return sinkID
}

func findTask(id string, project *config.Project) *config.Task {
	for _, task := range project.Tasks {
		if task.ID == id {
			return &task
		}
	}

	return nil
}
