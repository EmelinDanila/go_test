package nats

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

var Conn *nats.Conn

func Init() {
	var err error
	Conn, err = nats.Connect("nats://nats:4222")
	if err != nil {
		log.Fatalf("NATS connection error: %v", err)
	}
	log.Println("Connected to NATS")
}

type Event struct {
	ID          uint32    `json:"id"`
	ProjectID   uint32    `json:"project_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"`
	Removed     bool      `json:"removed"`
	EventTime   time.Time `json:"event_time"`
}

func Publish(ctx context.Context, evt Event) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	return Conn.Publish("logs", data)
}
