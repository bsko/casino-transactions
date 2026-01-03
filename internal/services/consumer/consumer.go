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

	// Создаем новый контекст для flush операций, чтобы они могли завершиться даже при отмене основного контекста
	flushCtx := context.Background()

	batcher := NewBatcher(s.batchSize, batchTimeout, func(events []entity.TransactionEvent) error {
		log.Println("Saving batch of events")
		// Используем flushCtx вместо ctx, чтобы flush мог завершиться даже при отмене основного контекста
		return s.transactionEventRepository.BatchStore(flushCtx, events)
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
				// Если ошибка из-за отмены контекста, flush-нем батчер перед возвратом
				if ctx.Err() != nil {
					log.Println("Consumer stopped by context cancellation during read")
					_ = batcher.Close()
					return nil
				}
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

func (s *Consumer) SetBatchSize(batchSize int) {
	s.batchSize = batchSize
}
