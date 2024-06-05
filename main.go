package main

import (
	"log"
	"log/slog"
	"time"

	"github.com/anthdm/hollywood/actor"
)

type VisitRequest struct {
	links []string
}

type Manager struct{}

func NewManager() actor.Producer {
	return func() actor.Receiver {
		return &Manager{}
	}
}

func (m *Manager) Receive(c *actor.Context) {
	switch msg := c.Message().(type) {
	case VisitRequest:
		m.handleVisitRequest(msg)
	case actor.Started:
		slog.Info("manager started")
	case actor.Stopped:
	}
}

func (m *Manager) handleVisitRequest(msg VisitRequest) error {
	for _, link := range msg.links {
		slog.Info("visiting url", "url", link)
	}
	return nil
}

func main() {
	engine, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		log.Fatal(err)
	}

	pid := engine.Spawn(NewManager(), "manager")

	engine.Send(pid, VisitRequest{links: []string{"https://petrostrak.app/"}})

	time.Sleep(5 * time.Second)
}
