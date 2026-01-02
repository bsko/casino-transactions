package producer

import (
	"context"

	"github.com/bsko/casino-transaction-system/internal/entity"
)

type KafkaWriter interface {
	Publish(ctx context.Context, event entity.TransactionEvent) error
}
