package runner

import (
	"bufio"
	"io"
	"log"
	"nicograshoff.de/x/optask/archive"
	"nicograshoff.de/x/optask/config"
	"os/exec"
	"sync"
	"time"
)

type job struct {
	task   *config.Task
	id     string
	stdout []string
	node   *archive.Node
}

var jobs map[string]*job
var jobsMutex *sync.Mutex
var runMutex *sync.Mutex
var jobChannel chan *job
var logs map[string]*archive.FileSystem

func Init(l map[string]*archive.FileSystem) {
	logs = l
	jobs = make(map[string]*job)
	jobsMutex = &sync.Mutex{}
	jobChannel = make(chan *job)
	runMutex = &sync.Mutex{}
	go dispatchLoop()
}

func Run(task *config.Task) string {
	id := archive.NewIdentifierNow()

	runMutex.Lock()
	defer runMutex.Unlock()

	fs := logs[task.ID]
	if fs == nil {
		log.Panic("No log fs")
	}
	node := fs.CreateNode(id)
	jobID := node.String()
	job := &job{task, jobID, make([]string, 100), node}

	jobsMutex.Lock()
	jobs[jobID] = job
	jobsMutex.Unlock()

	jobChannel <- job

	return jobID
}

func GetStdout(task string, jobID string, line int) []string {
	jobsMutex.Lock()
	job := jobs[jobID]
	jobsMutex.Unlock()

	var stdout []string
	if job == nil {
		node := archive.ParseNode(jobID)
		file, err := logs[task].Open(&node, "stdout.txt")
		if err != nil {
			log.Panic(err)
		}

		stdout = make([]string, 100)
		r := bufio.NewReader(file)
		for {
			line, err := r.ReadString('\n')
			stdout = append(stdout, line)
			if err == io.EOF {
				break
			}
		}
	} else {
		// the job is still running so we wait a bit to collect some lines
		time.Sleep(250 * time.Millisecond)

		stdout = job.stdout
	}
	return stdout[line:]
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
		// We would have to read stdin from the file system otherwise.
		time.Sleep(1 * time.Second)

		fs := logs[job.task.ID]
		file, err := fs.Create(job.node, "stdout.txt")
		if err != nil {
			log.Println(err)
		}
		defer file.Close()

		w := bufio.NewWriter(file)
		for _, line := range job.stdout {
			w.WriteString(line)
		}
		w.Flush()

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
