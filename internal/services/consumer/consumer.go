package consumer

import (
	"context"
	"log"
	"time"

	"github.com/bsko/casino-transaction-system/internal/entity"
)

const (
	batchSize    = 100
	batchTimeout = 10 * time.Second
)

type Consumer struct {
	reader                     kafkaReader
	transactionEventRepository transactionEventSaveRepository
}

func NewConsumer(reader kafkaReader, repository transactionEventSaveRepository) *Consumer {
	return &Consumer{
		reader:                     reader,
		transactionEventRepository: repository,
	}
}

func (s *Consumer) Start(ctx context.Context) error {
	log.Println("Starting consumer")
	defer func() { log.Println("Stopping consumer") }()

	batcher := NewBatcher(batchSize, batchTimeout, func(events []entity.TransactionEvent) error {
		log.Println("Saving batch of events")
		return s.transactionEventRepository.BatchStore(events)
	})
	defer func() { _ = batcher.Close() }()

	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer stopped by context cancellation")
			return nil
		default:
			msg, err := s.reader.Read(ctx)
			if err != nil {
				log.Printf("Failed to read message from Kafka: %v", err)
				return err
			}
			if err = batcher.Add(*msg); err != nil {
				log.Printf("Failed to add message to batcher: %v", err)
				return err
			}
		}
	}
}
