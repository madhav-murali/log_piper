package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type LogMessage struct {
	Service   string `json:"service"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	Level     string `json:"level"`
}

type EnrichedLog struct {
	LogMessage
	Host     string    `json:"host"`
	Received time.Time `json:"received_at"`
}

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Drain()

	js, _ := nc.JetStream()

	sub, err := js.PullSubscribe("logs.>", "log_ingestor", nats.BindStream("LOG_STREAM"))
	if err != nil {
		log.Fatal(err)
	}

	//ctx := context.Background()

	for {
		msgs, err := sub.Fetch(5, nats.MaxWait(5*time.Second))
		if err != nil {
			continue
		}
		for _, msg := range msgs {
			var l LogMessage
			if err := json.Unmarshal(msg.Data, &l); err != nil {
				log.Println("Bad message:", err)
				msg.Ack()
				continue
			}

			enriched := EnrichedLog{
				LogMessage: l,
				Host:       "local-dev",
				Received:   time.Now(),
			}

			// Print or persist
			fmt.Printf("[%s] %s | %s | %s| %s\n",
				enriched.Received.Format(time.RFC3339),
				enriched.Service,
				enriched.Host,
				enriched.Level,
				enriched.Message)

			msg.Ack()
		}
	}
}
