package http

import "github.com/bsko/casino-transaction-system/internal/entity"

type postTransactionsMessageHandler interface {
	GetListByFilter(filter entity.TransactionEventFilter) ([]entity.TransactionEvent, error)
}
