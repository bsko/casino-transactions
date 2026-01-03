//go:generate go run go.uber.org/mock/mockgen@latest -source=types.go -destination=mocks/mocks.go -package=mocks
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
	BatchStore(ctx context.Context, batch []entity.TransactionEvent) error
}

type httpServer interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}
