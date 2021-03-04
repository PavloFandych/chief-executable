package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fmt.Println("Chief Executable starting...")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	team := make(map[string]*Microservice)
	prepareTeam(team)
	runTeam(team)

	time.Sleep(5 * time.Second)
	if err := team["event"].Command.Process.Signal(os.Interrupt); err != nil {
		log.Fatal(err)
	}
	fmt.Println("\n", <-signals)
}

type Microservice struct {
	Name    string
	Command *exec.Cmd
}

func (m *Microservice) New(name string, arg ...string) {
	m.Name = name
	m.Command = exec.Command(m.Name, arg...)
	m.Command.Stdout = os.Stdout
}

func (m *Microservice) Run() {
	if err := m.Command.Run(); err != nil {
		log.Fatal(err)
	}
}

func prepareTeam(team map[string]*Microservice) {
	var data Microservice
	data.New("ping", "google.com.ua")
	team["data"] = &data
	var event Microservice
	event.New("ping", "www.youtube.com")
	team["event"] = &event
}

func runTeam(team map[string]*Microservice) {
	for key, value := range team {
		fmt.Println("Starting", key)
		go value.Run()
	}
}
