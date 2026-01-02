package entity

import "time"

type TransactionEventFilter struct {
	UserID          *UserID
	TransactionType *TransactionType
	AmountFrom      *Money
	AmountTo        *Money
	CreatedFrom     *time.Time
	CreatedTo       *time.Time
	Limit           int
	Offset          int
}
