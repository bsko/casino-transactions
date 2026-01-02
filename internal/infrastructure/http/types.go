//go:generate go run go.uber.org/mock/mockgen@latest -source=types.go -destination=mocks/mocks.go -package=mocks
package http

import "github.com/bsko/casino-transaction-system/internal/entity"

type postTransactionsMessageHandler interface {
	GetListByFilter(filter entity.TransactionEventFilter) ([]entity.TransactionEvent, error)
}
