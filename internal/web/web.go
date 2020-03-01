package web

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ngrash/optask/internal/model"
	"github.com/ngrash/optask/internal/runner"
	"github.com/ngrash/optask/internal/stdstreams"
)

const TemplatePath = "web/tmpl"
const ReloadTemplates = true

// Context is passed to each web handler function
type Server struct {
	proj     *model.Project
	runner   *runner.Service
	mux      *http.ServeMux
	template struct {
		index, show, history *template.Template
	}
}

func NewServer(p *model.Project, r *runner.Service) (*Server, error) {
	s := &Server{proj: p, runner: r, mux: http.NewServeMux()}
	if err := s.loadTemplates(); err != nil {
		return nil, err
	}

	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	s.mux.HandleFunc("/", s.serveIndex)
	s.mux.HandleFunc("/exec", s.serveExec)
	s.mux.HandleFunc("/status", s.serveStatus)
	s.mux.HandleFunc("/show", s.serveShow)
	s.mux.HandleFunc("/history", s.serveHistory)
	s.mux.HandleFunc("/stdstreams", s.serveStdstreams)

	return s, nil
}

func exist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (s *Server) loadTemplates() error {
	root := filepath.Join(TemplatePath, "root.tmpl")

	common := filepath.Join(TemplatePath, "common.tmpl")
	hasCommon := exist(common)

	parse := func(name string) (*template.Template, error) {
		path := filepath.Join(TemplatePath, name)
		files := []string{root, path}
		if hasCommon {
			files = append(files, common)
		}
		return template.ParseFiles(files...)
	}

	var err error
	s.template.index, err = parse("index.tmpl")
	if err != nil {
		return err
	}
	s.template.show, err = parse("show.tmpl")
	if err != nil {
		return err
	}
	s.template.history, err = parse("history.tmpl")
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) renderTemplate(w http.ResponseWriter, t *template.Template, data interface{}) {
	err := t.ExecuteTemplate(w, "root.tmpl", data)
	handleErrorMaybe(w, err)
}

func handleErrorMaybe(w http.ResponseWriter, err error) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if ReloadTemplates {
		s.loadTemplates()
	}

	s.mux.ServeHTTP(w, r)
}

func (s *Server) serveIndex(w http.ResponseWriter, r *http.Request) {
	runs, err := s.runner.LatestRuns()
	if err != nil {
		log.Panic(err)
	}

	type runView struct {
		ID       string
		TaskID   string
		ExitCode int
		Running  bool
		Duration time.Duration
		Exists   bool
	}

	type taskView struct {
		ID, Name string
		LastRun  runView
	}

	type view struct {
		Title string
		Tasks []taskView
	}

	tasks := make([]taskView, len(s.proj.Tasks))
	for i, t := range s.proj.Tasks {
		tasks[i] = taskView{ID: string(t.ID), Name: t.Name}
		r := runs[t.ID]
		if r != nil {
			tasks[i].LastRun = runView{
				ID:       string(r.ID),
				TaskID:   string(t.ID),
				Running:  s.runner.IsRunning(t.ID, r.ID),
				Exists:   true,
				ExitCode: r.ExitCode,
				Duration: s.duration(t.ID, r),
			}
		}
	}

	v := view{s.proj.Name, tasks}

	s.renderTemplate(w, s.template.index, v)
}

func (s *Server) duration(tID model.TaskID, r *model.Run) time.Duration {
	var d time.Duration
	if s.runner.IsRunning(tID, r.ID) {
		d = time.Since(r.Started)
	} else {
		d = time.Since(r.Completed)
	}
	return d.Truncate(time.Second)
}

func (s *Server) serveShow(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	type viewModel struct {
		Title     string
		Name      string
		CmdLine   string
		Lines     []stdstreams.Line
		ExitCode  int
		Duration  time.Duration
		Skip      int
		Running   bool
		ID        string
		TaskID    string
		Started   time.Time
		Completed time.Time
	}

	tID := model.TaskID(r.Form.Get("t"))
	rID := model.RunID(r.Form.Get("r"))

	streams, err := s.runner.StdStreams(tID, rID)
	if err != nil {
		log.Panic(err)
	}

	task, err := s.runner.Task(tID)
	if err != nil {
		log.Panic(err)
	}

	run, err := s.runner.Run(tID, rID)
	if err != nil {
		log.Panic(err)
	}

	cmdLine := fmt.Sprintf("%v %v", task.Cmd, strings.Join(task.Args, " "))
	isRunning := s.runner.IsRunning(tID, rID)

	lines := streams.Lines()

	v := &viewModel{
		Title:     s.proj.Name,
		Name:      task.Name,
		CmdLine:   cmdLine,
		Lines:     lines,
		Skip:      len(lines),
		Running:   isRunning,
		Duration:  s.duration(tID, run),
		ExitCode:  run.ExitCode,
		ID:        string(rID),
		TaskID:    string(tID),
		Started:   run.Started,
		Completed: run.Completed,
	}

	s.renderTemplate(w, s.template.show, v)
}

func (s *Server) serveStatus(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	tID := model.TaskID(r.Form.Get("t"))
	rID := model.RunID(r.Form.Get("r"))

	run, err := s.runner.Run(tID, rID)
	if err != nil {
		log.Panic(err)
	}

	type data struct {
		Running   bool
		Started   time.Time
		Completed time.Time
		ExitCode  int
	}

	d := data{
		Running:   s.runner.IsRunning(tID, rID),
		Started:   run.Started,
		Completed: run.Completed,
		ExitCode:  run.ExitCode,
	}

	handleErrorMaybe(w, s.template.show.ExecuteTemplate(w, "status", d))
}

func (s *Server) serveExec(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	tID := r.Form.Get("t")
	rID, err := s.runner.Exec(model.TaskID(tID))
	if err != nil {
		log.Panic(err)
	}

	http.Redirect(w, r, "/show?t="+tID+"&r="+string(rID), http.StatusSeeOther)
}

func (s *Server) serveHistory(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	tID := model.TaskID(r.Form.Get("t"))
	before := model.RunID(r.Form.Get("b"))

	runs, err := s.runner.Runs(tID, before, 50)
	if err != nil {
		log.Panic(err)
	}

	type runView struct {
		ID       string
		TaskID   string
		Running  bool
		ExitCode int
		Duration time.Duration
	}

	type taskView struct {
		ID   string
		Name string
	}

	type view struct {
		Title string
		Task  taskView
		Runs  []runView
	}

	runViews := make([]runView, len(runs))
	for i, r := range runs {
		runViews[i] = runView{
			ID:       string(r.ID),
			TaskID:   string(tID),
			ExitCode: r.ExitCode,
			Running:  s.runner.IsRunning(tID, r.ID),
			Duration: s.duration(tID, r),
		}
	}

	t, _ := s.runner.Task(tID)

	v := view{
		Title: s.proj.Name,
		Task: taskView{
			ID:   string(tID),
			Name: t.Name,
		},
		Runs: runViews,
	}

	s.renderTemplate(w, s.template.history, v)
}

func (s *Server) serveStdstreams(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	tID := model.TaskID(r.Form.Get("t"))
	rID := model.RunID(r.Form.Get("r"))
	skip, err := strconv.Atoi(r.Form.Get("s"))
	if err != nil {
		log.Panic(err)
	}

	streams, err := s.runner.StdStreams(tID, rID)
	if err != nil {
		log.Panic(err)
	}

	if s.runner.IsRunning(tID, rID) {
		w.Header().Set("Optask-Running", "1")
	} else {
		w.Header().Set("Optask-Running", "0")
	}

	json, err := streams.JSON(skip)
	if err != nil {
		log.Panic(err)
	}

	buf := bytes.NewBuffer(json)
	buf.WriteTo(w)
}
