package producer

import (
	"context"

	"github.com/bsko/casino-transaction-system/internal/entity"
)

type producerInterface interface {
	Start(ctx context.Context) error
}

type kafkaAdapterInterface interface {
	Connect(ctx context.Context) error
	Publish(ctx context.Context, event entity.TransactionEvent) error
	Close() error
}
