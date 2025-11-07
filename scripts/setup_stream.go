package main

import (
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, _ := nats.Connect(nats.DefaultURL)
	js, _ := nc.JetStream()

	streamName := "LOG_STREAM"

	_, err := js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{"logs.>"},
		Storage:  nats.FileStorage,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created stream:", streamName)
}
