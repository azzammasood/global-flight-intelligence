package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/learning/flight-intel-pipeline/internal/kafka"
	"github.com/learning/flight-intel-pipeline/internal/opensky"
)

type Config struct {
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

func loadConfig(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {
	cfg, err := loadConfig("credentials.json")
	if err != nil {
		log.Printf("Warning: Could not load credentials.json: %v. Using unauthenticated requests...", err)
		cfg = &Config{}
	}

	// For Opensky API, basic username and password correspond to clientId and clientSecret
	client := opensky.NewClient(cfg.ClientId, cfg.ClientSecret)
	
	brokers := []string{"localhost:9092"}
	topic := "flights-raw"
	
	producer := kafka.NewProducer(brokers, topic)
	defer producer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down producer...")
		cancel()
	}()

	ticker := time.NewTicker(10 * time.Second) // OpenSky free tier allows 1 req / 10s anon, 1 req / 5s with auth
	defer ticker.Stop()

	log.Println("Starting OpenSky Data Producer...")

	// First immediate fetch
	fetchAndPublish(ctx, client, producer)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fetchAndPublish(ctx, client, producer)
		}
	}
}

func fetchAndPublish(ctx context.Context, client *opensky.Client, producer *kafka.Producer) {
	log.Println("Fetching flights from OpenSky API...")
	flights, err := client.FetchFlights()
	if err != nil {
		log.Printf("Error fetching flights: %v", err)
		return
	}

	log.Printf("Fetched %d flights. Publishing to Kafka...", len(flights))
	if err := producer.PublishFlights(ctx, flights); err != nil {
		log.Printf("Error publishing to Kafka: %v", err)
		return
	}

	log.Println("Successfully published batch to Kafka.")
}
