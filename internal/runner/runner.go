package runner

import (
	"os/exec"

	"github.com/ngrash/optask/internal/stdstreams"
)

type runner struct {
	jobs chan *jobInfo
}

type doneFunc func(exit int)

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

	err := job.cmd.Start()
	exitErr, isExitErr := err.(*exec.ExitError)
	if err != nil && !isExitErr {
		panic(err)
	}

	err = job.cmd.Wait()
	exitErr, isExitErr = err.(*exec.ExitError)
	if err != nil && !isExitErr {
		panic(err)
	}

	job.log.Flush()

	exit := 0
	if isExitErr {
		exit = exitErr.ExitCode()
	}

	job.doneFn(exit)
}
