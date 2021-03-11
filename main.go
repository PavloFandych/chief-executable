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
	team := make(map[string]*Microservice)
	prepareTeam(&wgDoneAll, team)
	runTeam(team)

	time.Sleep(5 * time.Second)
	interruptProgrammatically("event", team)
	//main goroutine is blocked until system event happens
	log.Println("Signal has been caught:", <-signals)
	//sending message into each Done channel
	shutdownTeam(team)
	//blocked until all shutdown actions happen
	wgDoneAll.Wait()
	log.Println("Chief Executable completed.")
}

type Microservice struct {
	ShutdownWaitGroup *sync.WaitGroup
	Name              string
	Args              []string
	Command           *exec.Cmd
	Shutdown          chan bool
}

func (m *Microservice) Init(wgDone *sync.WaitGroup, name string, args ...string) {
	m.ShutdownWaitGroup = wgDone
	m.ShutdownWaitGroup.Add(1)
	m.Name = name
	m.Args = make([]string, len(args))
	m.Args = append(m.Args, args...)
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
		log.Println("Shutting down:", m.Name, m.Args)
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
	for _, value := range team {
		log.Println("Starting:", value.Name, value.Args)
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
