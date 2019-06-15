package runner

import (
	"bufio"
	"io"
	"log"
	"nicograshoff.de/x/optask/config"
	"os/exec"
	"sync"
	"time"
)

type job struct {
	task     *config.Task
	id       int
	stdout   []string
	finished bool
}

var jobs map[int]*job
var jobsMutex *sync.Mutex
var runCounter int
var runCounterMutex *sync.Mutex
var jobChannel chan *job

func Init() {
	jobs = make(map[int]*job)
	jobsMutex = &sync.Mutex{}
	runCounter = 0
	runCounterMutex = &sync.Mutex{}
	jobChannel = make(chan *job)
	go dispatchLoop()
}

func Run(task *config.Task) int {
	runCounterMutex.Lock()
	runCounter++
	jobID := runCounter
	runCounterMutex.Unlock()

	jobsMutex.Lock()
	job := &job{task, jobID, make([]string, 100), false}
	jobs[jobID] = job
	jobsMutex.Unlock()
	jobChannel <- job

	return jobID
}

func IsRunning(jobID int) bool {
	return jobs[jobID] != nil
}

func GetStdout(jobID int, line int) []string {
	jobsMutex.Lock()
	job := jobs[jobID]
	jobsMutex.Unlock()
	return job.stdout[line:]
}

func dispatchLoop() {
	for {
		job := <-jobChannel
		go run(job)
	}
}

func run(job *job) {
	defer func() {
		// Wait a second in case someone is still interested in reading stdin
		// We would have to read stdin from the logfile otherwise.
		time.Sleep(1 * time.Second)

		// TODO: Write to logfile

		jobsMutex.Lock()
		delete(jobs, job.id)
		jobsMutex.Unlock()
	}()

	task := job.task
	cmd := exec.Command(task.Command, task.Args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(stdout)
	for {
		line, err := reader.ReadString('\n')
		job.stdout = append(job.stdout, line)
		if err == io.EOF {
			break
		}
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}
