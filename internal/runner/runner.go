package runner

import (
	"os/exec"
)

// Runner is responsible for asynchronously running commands and writing their
// output to a OutputSink.
type runner struct {
	jobs jobChan
}

// Callback is called after the command completed and the sink was closed.
type callback func()

// jobChan represents the channel that is used to pass jobInfo objects.
type jobChan chan *jobInfo

// jobInfo represents a scheduled job.
type jobInfo struct {
	sink     *sink
	cmd      *exec.Cmd
	callback callback
}

// NewRunner creates a new Runner and starts it's dispatch loop.
func newRunner() *runner {
	r := new(runner)
	r.jobs = make(jobChan)
	go r.dispatchAndLoop()
	return r
}

// Run dispatches a command with a given name and args writing standard
// streams to the sink.
func (r *runner) Run(name string, args []string, s *sink, cb callback) {
	cmd := exec.Command(name, args...)
	r.jobs <- &jobInfo{s, cmd, cb}
}

// dispatchAndLoop reads jobInfo objects from the Runner's jobChan and run
// them asynchronously, then repeats.
func (r *runner) dispatchAndLoop() {
	for {
		job := <-r.jobs
		go r.run(job)
	}
}

// run connects the standard streams to the sink and runs the command.
func (r *runner) run(job *jobInfo) {
	defer job.sink.close()
	defer func() {
		go job.callback()
	}()

	job.cmd.Stdout = job.sink.openStdout()
	job.cmd.Stderr = job.sink.openStderr()

	if err := job.cmd.Start(); err != nil {
		panic(err)
	}

	if err := job.cmd.Wait(); err != nil {
		panic(err)
	}
}
