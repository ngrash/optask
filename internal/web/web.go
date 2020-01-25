package web

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

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
	handle(c, "/latest", latestHandler)
	handle(c, "/details", detailsHandler)
	//handle(c, "/output", outputHandler)

	log.Printf("Serving project " + c.Project.Name + " on " + addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func indexHandler(c *Context, w http.ResponseWriter, r *http.Request) {
	template, err := template.ParseFiles("web/templates/index.html")
	if err != nil {
		log.Panic(err)
	}

	template.Execute(w, c.Project)
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

func latestHandler(c *Context, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	tID := model.TaskID(r.Form.Get("t"))
	rID, err := c.Runner.LatestRun(tID)
	if err != nil {
		log.Panic(err)
	}
	http.Redirect(w, r, "/details?t="+string(tID)+"&r="+string(rID), http.StatusSeeOther)
}

func detailsHandler(c *Context, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	type viewModel struct {
		Title     string
		CmdLine   string
		Output    []stdstreams.Line
		IsRunning bool
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

	cmdLine := fmt.Sprintf("%v %v", task.Command, strings.Join(task.Args, " "))
	isRunning := c.Runner.IsRunning(tID, rID)

	vm := &viewModel{task.Name, cmdLine, streams.Lines(), isRunning}
	template.Execute(w, vm)
}

/*func outputHandler(c *Context, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	tID := model.TaskID(r.Form.Get("t"))
	rID := model.RunID(r.Form.Get("r"))

	streams, err := c.runner.StdStreams(tID, rID)
	if err != nil {
		log.Panic(err)
	}

	if c.runner.IsRunning(tID, rID) {
		w.Header().Set("Optask-Running", "1")
	} else {
		w.Header().Set("Optask-Running", "0")
	}

	streams.WriteTo(w)
}*/

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
