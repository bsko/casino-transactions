package consumer

import (
	"context"

	"github.com/bsko/casino-transaction-system/internal/entity"
)

type kafkaReader interface {
	Read(ctx context.Context) (*entity.TransactionEvent, error)
}

type transactionEventReadRepository interface {
	GetListByFilter(filter entity.TransactionEventFilter) ([]entity.TransactionEvent, error)
}

type transactionEventSaveRepository interface {
	BatchStore(batch []entity.TransactionEvent) error
}
