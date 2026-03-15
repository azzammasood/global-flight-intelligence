package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/learning/flight-intel-pipeline/pkg/models"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.Hash{},
		AllowAutoTopicCreation: true,
	}

	return &Producer{
		writer: w,
	}
}

func (p *Producer) PublishFlights(ctx context.Context, flights []models.Flight) error {
	var msgs []kafka.Message

	for _, f := range flights {
		b, err := json.Marshal(f)
		if err != nil {
			log.Printf("Failed to marshal flight %s: %v", f.Icao24, err)
			continue
		}

		msgs = append(msgs, kafka.Message{
			Key:   []byte(f.Icao24), // partition by aircraft ID
			Value: b,
		})
	}

	if len(msgs) == 0 {
		return nil
	}

	err := p.writer.WriteMessages(ctx, msgs...)
	if err != nil {
		return fmt.Errorf("failed to write messages: %w", err)
	}

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
