// Package exec provides types and methods to asynchronously run commands and
// capture standard streams (Stdout and Stderr)
package exec

import (
	"io"
	"os/exec"
)

// Runner is responsible for asynchronously running commands and writing their
// output to a OutputSink.
type Runner struct {
	jobs jobChan
}

// OutputSink provides methods to handle standard streams of a command.
type OutputSink interface {
	OpenStdout() io.Writer
	OpenStderr() io.Writer
	Close()
}

// Callback is called after the command completed and the sink was closed.
type Callback func()

// jobChan represents the channel that is used to pass jobInfo objects.
type jobChan chan *jobInfo

// jobInfo represents a scheduled job.
type jobInfo struct {
	sink     OutputSink
	cmd      *exec.Cmd
	callback Callback
}

// NewRunner creates a new Runner and starts it's dispatch loop.
func NewRunner() *Runner {
	r := new(Runner)
	r.jobs = make(jobChan)
	go r.dispatchAndLoop()
	return r
}

// Run dispatches a command with a given name and args writing standard
// streams to the sink.
func (r *Runner) Run(name string, args []string, sink OutputSink, callback Callback) {
	cmd := exec.Command(name, args...)
	r.jobs <- &jobInfo{sink, cmd, callback}
}

// dispatchAndLoop reads jobInfo objects from the Runner's jobChan and run
// them asynchronously, then repeats.
func (r *Runner) dispatchAndLoop() {
	for {
		job := <-r.jobs
		go r.run(job)
	}
}

// run connects the standard streams to the sink and runs the command.
func (r *Runner) run(job *jobInfo) {
	defer job.sink.Close()
	defer func() {
		go job.callback()
	}()

	job.cmd.Stdout = job.sink.OpenStdout()
	job.cmd.Stderr = job.sink.OpenStderr()

	if err := job.cmd.Start(); err != nil {
		panic(err)
	}

	if err := job.cmd.Wait(); err != nil {
		panic(err)
	}
}
