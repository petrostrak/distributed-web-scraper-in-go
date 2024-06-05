package main

import (
	"fmt"
	"log"
	"time"

	"github.com/anthdm/hollywood/actor"
)

func main() {
	engine, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		log.Fatal(err)
	}
	engine.SpawnFunc(func(ctx *actor.Context) {
		switch msg := ctx.Message().(type) {
		case actor.Started:
			fmt.Println("started")
			_ = msg
		}
	}, "foo")

	time.Sleep(5 * time.Second)
}
