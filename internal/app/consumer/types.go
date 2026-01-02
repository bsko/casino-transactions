package consumer

import (
	"context"

	"github.com/bsko/casino-transaction-system/internal/entity"
)

type consumerInterface interface {
	Start(ctx context.Context) error
}

type kafkaReaderInterface interface {
	Connect(ctx context.Context) error
	Read(ctx context.Context) (*entity.TransactionEvent, error)
	Close() error
}

type transactionEventRepositoryInterface interface {
	Connect() error
	Close() error
	BatchStore(batch []entity.TransactionEvent) error
}

type httpServer interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}
