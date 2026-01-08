//go:generate go run go.uber.org/mock/mockgen@latest -source=types.go -destination=mocks/mocks.go -package=mocks
package consumer

import (
	"context"

	"github.com/bsko/casino-transaction-system/internal/entity"
)

type kafkaReader interface {
	Read(ctx context.Context) (*entity.TransactionEvent, error)
	Commit(ctx context.Context) error
}

type transactionEventReadRepository interface {
	GetListByFilter(filter entity.TransactionEventFilter) ([]entity.TransactionEvent, error)
}

type transactionEventSaveRepository interface {
	BatchStore(ctx context.Context, batch []entity.TransactionEvent) error
}
