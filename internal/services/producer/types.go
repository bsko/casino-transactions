//go:generate go run go.uber.org/mock/mockgen@latest -source=types.go -destination=mocks/mocks.go -package=mocks
package producer

import (
	"context"

	"github.com/bsko/casino-transaction-system/internal/entity"
)

type KafkaWriter interface {
	Publish(ctx context.Context, event entity.TransactionEvent) error
}
