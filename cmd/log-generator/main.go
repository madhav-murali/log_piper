package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/nats-io/nats.go"
)

type LogMessage struct {
	Service   string `json:"service"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	Level     string `json:"level"`
}

func main() {
	nc, _ := nats.Connect(nats.DefaultURL)
	js, _ := nc.JetStream()

	services := []string{"auth", "billing", "inventory", "gateway"}
	messages := []string{
		"user login succeeded",
		"payment initiated",
		"order dispatched",
		"service started",
		"cache miss",
		"timeout occurred",
	}

	for {
		logMsg := LogMessage{
			Service:   services[rand.Intn(len(services))],
			Message:   messages[rand.Intn(len(messages))],
			Timestamp: time.Now().Unix(),
			Level:     []string{"INFO", "WARN", "ERROR"}[rand.Intn(3)],
		}

		data, _ := json.Marshal(logMsg)
		subject := fmt.Sprintf("logs.%s", logMsg.Service)
		_, err := js.Publish(subject, data)
		if err != nil {
			fmt.Println("Error publishing log:", err)
		} else {
			fmt.Printf("Published → %s | %s\n", subject, logMsg.Message)
		}

		time.Sleep(2 * time.Second)
	}
}
