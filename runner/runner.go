package runner

import (
	"bufio"
	"io"
	"log"
	"nicograshoff.de/x/optask/config"
	"os/exec"
)

func Start() chan *config.Task {
	channel := make(chan *config.Task)
	go serve(channel)
	return channel
}

func serve(channel chan *config.Task) {
	for {
		task := <-channel
		go handle(task)
	}
}

func handle(task *config.Task) {
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
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}

		if line != "" {
			log.Print(task.Name + ": " + line)
		}

		if err == io.EOF {
			break
		}
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}
