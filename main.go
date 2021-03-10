package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	fmt.Println("Chief Executable starting...")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	var wgDoneAll sync.WaitGroup
	team := make(map[string]*Microservice)
	prepareTeam(&wgDoneAll, team)
	runTeam(team)

	time.Sleep(5 * time.Second)
	interruptProgrammatically("event", team)

	fmt.Println("Signal has been caught:", <-signals)

	shutdownTeam(team)

	wgDoneAll.Wait()
	fmt.Println("Chief Executable completed.")
}

type Microservice struct {
	ShutdownWaitGroup *sync.WaitGroup
	Name              string
	Command           *exec.Cmd
	Shutdown          chan bool
}

func (m *Microservice) Init(wgDone *sync.WaitGroup, name string, args ...string) {
	m.ShutdownWaitGroup = wgDone
	m.ShutdownWaitGroup.Add(1)
	m.Name = name
	m.Command = exec.Command(m.Name, args...)
	m.Command.Stdout = os.Stdout
	m.Shutdown = make(chan bool, 1)
}

func (m *Microservice) Run() {
	go func() {
		if err := m.Command.Run(); err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		defer m.ShutdownWaitGroup.Done()
		<-m.Shutdown
		//graceful shutdown actions
		fmt.Println("Shutting down: ", m)
	}()
}

func prepareTeam(wgDone *sync.WaitGroup, team map[string]*Microservice) {
	var data Microservice
	data.Init(wgDone, "ping", "google.com.ua")
	team["data"] = &data
	var event Microservice
	event.Init(wgDone, "ping", "www.youtube.com")
	team["event"] = &event
}

func runTeam(team map[string]*Microservice) {
	for key, value := range team {
		fmt.Println("Starting:", key)
		value.Run()
	}
}

func shutdownTeam(team map[string]*Microservice) {
	for _, value := range team {
		value.Shutdown <- true
	}
}

func interruptProgrammatically(service string, team map[string]*Microservice) {
	if err := team[service].Command.Process.Signal(os.Interrupt); err != nil {
		log.Fatal(err)
	}
	close(team[service].Shutdown)
	delete(team, service)
}
