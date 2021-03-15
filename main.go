package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	log.Println("Chief Executable starting...")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	var wgDoneAll sync.WaitGroup
	team := make(map[string]*Microservice, 2)
	team["google"] = newMicroservice(&wgDoneAll, "www.google.com.ua")
	team["youtube"] = newMicroservice(&wgDoneAll, "www.youtube.com")
	//run all team
	for _, value := range team {
		runMicroservice(value)
	}

	time.Sleep(5 * time.Second)
	interruptProgrammatically("youtube", team)
	//main goroutine is blocked until system event happens
	<-signals
	//sending message into each Done channel
	for _, value := range team {
		value.Shutdown <- true
	}
	//blocked until all shutdown actions happen
	wgDoneAll.Wait()
	log.Println("Chief Executable completed.")
}

type Microservice struct {
	ShutdownWaitGroup *sync.WaitGroup
	Args              []string
	Command           *exec.Cmd
	Shutdown          chan bool
}

func newMicroservice(wgDone *sync.WaitGroup, args ...string) *Microservice {
	var service Microservice
	service.ShutdownWaitGroup = wgDone
	service.ShutdownWaitGroup.Add(1)
	service.Args = append(service.Args, args...)
	service.Command = exec.Command("ping", args...)
	service.Command.Stdout = os.Stdout
	service.Shutdown = make(chan bool, 1)
	return &service
}

func runMicroservice(service *Microservice) {
	go func() {
		log.Println("Starting:", service.Args)
		if err := service.Command.Run(); err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		defer service.ShutdownWaitGroup.Done()
		<-service.Shutdown
		//graceful shutdown actions
		log.Println("Shutting down:", service.Args)
	}()
}

func interruptProgrammatically(service string, team map[string]*Microservice) {
	if err := team[service].Command.Process.Signal(os.Interrupt); err != nil {
		log.Fatal(err)
	}
	close(team[service].Shutdown)
	delete(team, service)
}
