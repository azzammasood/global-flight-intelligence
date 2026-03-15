package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/learning/flight-intel-pipeline/internal/clickhouse"
	"github.com/learning/flight-intel-pipeline/internal/kafka"
	"github.com/learning/flight-intel-pipeline/pkg/models"
)

func main() {
	brokers := []string{"localhost:9092"}
	topic := "flights-raw"
	groupID := "flight-clickhouse-consumer"

	consumer := kafka.NewConsumer(brokers, topic, groupID)
	defer consumer.Close()

	chClient, err := clickhouse.NewClient([]string{"localhost:9000"})
	if err != nil {
		log.Fatalf("ClickHouse connection failed: %v", err)
	}
	defer chClient.Close()
	log.Println("Connected to ClickHouse.")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down consumer...")
		cancel()
	}()

	var batch []models.Flight
	batchTicker := time.NewTicker(2 * time.Second)
	defer batchTicker.Stop()

	log.Println("Starting Kafka Consumer...")

	msgChan := make(chan []byte, 1000)

	// Background reader routine
	go func() {
		for {
			msg, err := consumer.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Error reading message: %v", err)
				continue
			}
			msgChan <- msg.Value
		}
	}()

	for {
		select {
		case <-ctx.Done():
			flushBatch(context.Background(), chClient, batch)
			return
		case msgVal := <-msgChan:
			var f models.Flight
			if err := json.Unmarshal(msgVal, &f); err != nil {
				log.Printf("JSON unmarshal error: %v", err)
				continue
			}
			batch = append(batch, f)
			if len(batch) >= 1000 {
				flushBatch(ctx, chClient, batch)
				batch = nil
			}
		case <-batchTicker.C:
			if len(batch) > 0 {
				flushBatch(ctx, chClient, batch)
				batch = nil
			}
		}
	}
}

func flushBatch(ctx context.Context, chClient *clickhouse.Client, batch []models.Flight) {
	log.Printf("Flushing batch of %d flights to ClickHouse...", len(batch))
	if err := chClient.InsertFlights(ctx, batch); err != nil {
		log.Printf("Error inserting to ClickHouse: %v", err)
	}
}
