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
	batchSize                  int
}

func NewConsumer(reader kafkaReader, repository transactionEventSaveRepository) *Consumer {
	return &Consumer{
		reader:                     reader,
		transactionEventRepository: repository,
		batchSize:                  batchSize,
	}
}

func (s *Consumer) Start(ctx context.Context) error {
	log.Println("Starting consumer")
	defer func() { log.Println("Stopping consumer") }()

	flushCtx := context.Background()

	batcher := NewBatcher(s.batchSize, batchTimeout, func(events []entity.TransactionEvent) error {
		log.Println("Saving batch of events")
		err := s.transactionEventRepository.BatchStore(flushCtx, events)
		if err != nil {
			return err
		}
		if commitErr := s.reader.Commit(ctx); commitErr != nil {
			log.Printf("Failed to commit offsets: %v", commitErr)
		} else {
			log.Println("Offsets committed successfully")
		}
		return nil
	})
	defer func() { _ = batcher.Close() }()

	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer stopped by context cancellation")
			_ = batcher.Close()
			return nil
		default:
			msg, err := s.reader.Read(ctx)
			if err != nil {
				if ctx.Err() != nil {
					log.Println("Consumer stopped by context cancellation during read")
					_ = batcher.Close()
					return nil
				} else {
					select {
					case <-ctx.Done():
						log.Println("Consumer stopped by context cancellation during retry")
						_ = batcher.Close()
						return nil
					default:
						log.Println("Consumer stopped during read, repeating")
						continue
					}
				}
			}
			if err = batcher.Add(*msg); err != nil {
				log.Printf("Failed to add message to batcher: %v", err)
				return err
			}
		}
	}
}

func (s *Consumer) SetBatchSize(batchSize int) {
	s.batchSize = batchSize
}
