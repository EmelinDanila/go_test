package logger

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	appnats "github.com/EmelinDanila/go_test/internal/nats"
	natspkg "github.com/nats-io/nats.go"
)

var buffer = make([]appnats.Event, 0, 100)
var mu sync.Mutex

func StartConsumer() {
	_, err := appnats.Conn.Subscribe("logs", handle)
	if err != nil {
		log.Fatalf("NATS subscribe error: %v", err)
	}
	log.Println("Logger: consumer started")

	go func() {
		for {
			time.Sleep(5 * time.Second)
			flush()
		}
	}()

	select {} // блокируем горутину
}

func handle(msg *natspkg.Msg) {
	var evt appnats.Event
	if err := json.Unmarshal(msg.Data, &evt); err != nil {
		log.Printf("Failed to decode message: %v", err)
		return
	}

	mu.Lock()
	buffer = append(buffer, evt)
	mu.Unlock()
}

func flush() {
	mu.Lock()
	defer mu.Unlock()

	if len(buffer) == 0 {
		return
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"clickhouse:9000"},
		Auth: clickhouse.Auth{Database: "default"},
	})
	if err != nil {
		log.Printf("ClickHouse connection error: %v", err)
		return
	}
	defer conn.Close()

	batch, err := conn.PrepareBatch(context.Background(), `
		INSERT INTO logs (Id, ProjectId, Name, Description, Priority, Removed, EventTime)
	`)
	if err != nil {
		log.Printf("Prepare batch error: %v", err)
		return
	}

	for _, e := range buffer {
		err := batch.Append(
			e.ID,
			e.ProjectID,
			e.Name,
			e.Description,
			e.Priority,
			e.Removed,
			e.EventTime,
		)
		if err != nil {
			log.Printf("Batch append error: %v", err)
		}
	}

	if err := batch.Send(); err != nil {
		log.Printf("Batch send error: %v", err)
	}

	log.Printf("Flushed %d events to ClickHouse", len(buffer))
	buffer = buffer[:0]
}
