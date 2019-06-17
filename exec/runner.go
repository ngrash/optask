package exec

import (
	"log"
	"os/exec"
	"sync"
)

type jobChan chan *jobInfo

type Runner struct {
	orders  jobChan
	sinkFac SinkFactory
	runMu   *sync.Mutex
}

type jobInfo struct {
	sink Sink
	cmd  *exec.Cmd
}

// NewRunner initializes a new Runner and starts the dispatch loop.
func NewRunner(sinkFac SinkFactory) *Runner {
	r := new(Runner)
	r.orders = make(jobChan)
	r.sinkFac = sinkFac
	r.runMu = new(sync.Mutex)

	go r.dispatchAndLoop()

	return r
}

func (r *Runner) Run(name string, arg ...string) SinkID {
	r.runMu.Lock()
	defer r.runMu.Unlock()

	job := new(jobInfo)
	job.sink = r.sinkFac.NewSink()
	job.cmd = exec.Command(name, arg...)

	r.orders <- job

	return job.sink.ID()
}

// dispatchAndLoop starts a goroute for each job received through the jobChan.
func (r *Runner) dispatchAndLoop() {
	for {
		job := <-r.orders
		go r.run(job)
	}
}

// run is called in a goroutine by dispatchAndLoop for each job.
func (r *Runner) run(job *jobInfo) {
	defer job.sink.Close()

	job.cmd.Stdout = job.sink.OpenStdout()
	job.cmd.Stderr = job.sink.OpenStderr()

	if err := job.cmd.Start(); err != nil {
		log.Panic(err)
	}

	if err := job.cmd.Wait(); err != nil {
		log.Panic(err)
	}
}
