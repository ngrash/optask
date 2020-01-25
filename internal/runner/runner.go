package runner

import (
	"os/exec"

	"nicograshoff.de/x/optask/internal/stdstreams"
)

type runner struct {
	jobs chan *jobInfo
}

type doneFunc func()

type jobInfo struct {
	cmd    *exec.Cmd
	log    *stdstreams.Log
	doneFn doneFunc
}

func newRunner() *runner {
	r := new(runner)
	r.jobs = make(chan *jobInfo)
	go r.dispatchAndLoop()
	return r
}

func (r *runner) Run(name string, args []string, log *stdstreams.Log, doneFn doneFunc) {
	cmd := exec.Command(name, args...)
	r.jobs <- &jobInfo{cmd, log, doneFn}
}

func (r *runner) dispatchAndLoop() {
	for {
		job := <-r.jobs
		go r.run(job)
	}
}

func (r *runner) run(job *jobInfo) {
	job.cmd.Stdout = job.log.Stdout()
	job.cmd.Stderr = job.log.Stderr()

	if err := job.cmd.Start(); err != nil {
		panic(err)
	}

	if err := job.cmd.Wait(); err != nil {
		panic(err)
	}

	job.log.Flush()
	job.doneFn()
}
