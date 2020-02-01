package web

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"nicograshoff.de/x/optask/internal/model"
	"nicograshoff.de/x/optask/internal/runner"
	"nicograshoff.de/x/optask/internal/stdstreams"
)

// Context is passed to each web handler function
type Context struct {
	Project *model.Project
	Runner  *runner.Service
}

// ListenAndServe starts the web server
func ListenAndServe(c *Context, addr string) {
	handle(c, "/", indexHandler)
	handle(c, "/run", runHandler)
	handle(c, "/details", detailsHandler)
	handle(c, "/output", outputHandler)
	handle(c, "/history", historyHandler)

	log.Printf("Serving project " + c.Project.Name + " on " + addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func indexHandler(c *Context, w http.ResponseWriter, r *http.Request) {
	template, err := template.ParseFiles("web/templates/index.html")
	if err != nil {
		log.Panic(err)
	}

	runs, err := c.Runner.LatestRuns()
	if err != nil {
		log.Panic(err)
	}

	type runView struct {
		ID       string
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

	tasks := make([]taskView, len(c.Project.Tasks))
	for i, t := range c.Project.Tasks {
		tasks[i] = taskView{ID: string(t.ID), Name: t.Name}
		r := runs[t.ID]
		if r != nil {
			lr := runView{
				ID:      string(r.ID),
				Running: c.Runner.IsRunning(t.ID, r.ID),
				Exists:  true,
			}

			if lr.Running {
				lr.Duration = time.Since(r.Started)
			} else {
				lr.Duration = time.Since(r.Completed)
			}

			lr.Duration = lr.Duration.Truncate(time.Second)

			tasks[i].LastRun = lr
		}
	}

	template.Execute(w, view{
		c.Project.Name,
		tasks,
	})
}

func runHandler(c *Context, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	tID := r.Form.Get("t")
	rID, err := c.Runner.Run(model.TaskID(tID))
	if err != nil {
		log.Panic(err)
	}

	http.Redirect(w, r, "/details?t="+tID+"&r="+string(rID), http.StatusSeeOther)
}

func detailsHandler(c *Context, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	type viewModel struct {
		Title     string
		CmdLine   string
		Lines     []stdstreams.Line
		NumLines  int
		IsRunning bool
		RunID     string
		TaskID    string
	}

	tID := model.TaskID(r.Form.Get("t"))
	rID := model.RunID(r.Form.Get("r"))

	template, err := template.ParseFiles("web/templates/details.html")
	if err != nil {
		log.Panic(err)
	}

	streams, err := c.Runner.StdStreams(tID, rID)
	if err != nil {
		log.Panic(err)
	}

	task, err := c.Runner.Task(tID)
	if err != nil {
		log.Panic(err)
	}

	cmdLine := fmt.Sprintf("%v %v", task.Cmd, strings.Join(task.Args, " "))
	isRunning := c.Runner.IsRunning(tID, rID)

	lines := streams.Lines()
	vm := &viewModel{task.Name, cmdLine, lines, len(lines), isRunning, string(rID), string(tID)}
	template.Execute(w, vm)
}

func outputHandler(c *Context, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	tID := model.TaskID(r.Form.Get("t"))
	rID := model.RunID(r.Form.Get("r"))
	skip, err := strconv.Atoi(r.Form.Get("s"))
	if err != nil {
		log.Panic(err)
	}

	streams, err := c.Runner.StdStreams(tID, rID)
	if err != nil {
		log.Panic(err)
	}

	if c.Runner.IsRunning(tID, rID) {
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

func historyHandler(c *Context, w http.ResponseWriter, r *http.Request) {
	template, err := template.ParseFiles("web/templates/history.html")
	if err != nil {
		log.Panic(err)
	}

	r.ParseForm()

	tID := model.TaskID(r.Form.Get("t"))
	before := model.RunID(r.Form.Get("b"))

	runs, err := c.Runner.Runs(tID, before, 50)
	if err != nil {
		log.Panic(err)
	}

	type runView struct {
		ID       string
		Running  bool
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
		v := runView{
			ID:      string(r.ID),
			Running: c.Runner.IsRunning(tID, r.ID),
		}

		if v.Running {
			v.Duration = time.Since(r.Started)
		} else {
			v.Duration = time.Since(r.Completed)
		}

		v.Duration = v.Duration.Truncate(time.Second)

		runViews[i] = v
	}

	t, _ := c.Runner.Task(tID)

	template.Execute(w, view{
		Title: c.Project.Name,
		Task: taskView{
			ID:   string(tID),
			Name: t.Name,
		},
		Runs: runViews,
	})
}

type handleFunc func(http.ResponseWriter, *http.Request)
type handleContextFunc func(*Context, http.ResponseWriter, *http.Request)

func handle(c *Context, pattern string, fn handleContextFunc) {
	http.HandleFunc(pattern, makeHandler(fn, c))
}

func makeHandler(fn handleContextFunc, c *Context) handleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(c, w, r)
	}
}
