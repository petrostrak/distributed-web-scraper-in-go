package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/anthdm/hollywood/actor"
	"golang.org/x/net/html"
)

type VisitRequest struct {
	links []string
}

type Manager struct {
	visitors map[*actor.PID]bool
}

func NewManager() actor.Producer {
	return func() actor.Receiver {
		return &Manager{
			visitors: make(map[*actor.PID]bool),
		}
	}
}

func (m *Manager) Receive(c *actor.Context) {
	switch msg := c.Message().(type) {
	case VisitRequest:
		m.handleVisitRequest(c, msg)
	case actor.Started:
		slog.Info("manager started")
	case actor.Stopped:
	}
}

func (m *Manager) handleVisitRequest(c *actor.Context, msg VisitRequest) error {
	for _, link := range msg.links {
		slog.Info("visiting url", "url", link)
		c.SpawnChild(NewVisitor(link, c.PID()), fmt.Sprintf("visitor/%s", link))
	}
	return nil
}

type Visitor struct {
	managerPID *actor.PID
	URL        string
}

func NewVisitor(url string, mpid *actor.PID) actor.Producer {
	return func() actor.Receiver {
		return &Visitor{
			managerPID: mpid,
			URL:        url,
		}
	}
}

func (v *Visitor) Receive(c *actor.Context) {
	switch c.Message().(type) {
	case actor.Started:
		slog.Info("visitor started working on url", "url", v.URL)
		links, err := visit(v.URL)
		if err != nil {
			slog.Error("visit error", "err", err)
			return
		}
		c.Send(v.managerPID, VisitRequest{links: links})
	case actor.Stopped:
	}
}

func extractLinks(body io.Reader) []string {
	links := make([]string, 0)
	tokenizer := html.NewTokenizer(body)

	for {
		tokenType := tokenizer.Next()

		if tokenType == html.ErrorToken {
			return links
		}
		if tokenType == html.StartTagToken {
			token := tokenizer.Token()

			if token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						links = append(links, attr.Val)
					}
				}
			}
		}
	}
}

func visit(link string) ([]string, error) {
	baseURL, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(baseURL.String())
	if err != nil {
		return nil, err
	}

	urls := make([]string, 0)
	for _, link := range extractLinks(resp.Body) {
		linkUrl, err := url.Parse(link)
		if err != nil {
			fmt.Println(err)
		}
		urls = append(urls, baseURL.ResolveReference(linkUrl).String())
	}

	return urls, nil
}

func main() {
	engine, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		log.Fatal(err)
	}

	pid := engine.Spawn(NewManager(), "manager")
	engine.Send(pid, VisitRequest{links: []string{"https://petrostrak.netlify.app"}})

	time.Sleep(5 * time.Second)
}
