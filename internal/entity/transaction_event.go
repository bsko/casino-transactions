package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	TransactionTypeBet TransactionType = "bet"
	TransactionTypeWin TransactionType = "win"
)

type UserID struct {
	UUID uuid.UUID
}

func NewUserID(uuid uuid.UUID) *UserID {
	return &UserID{
		UUID: uuid,
	}
}

type TransactionType string

type Money int64

func (m Money) String() string {
	x := float64(m) / 100
	return fmt.Sprintf("$%.2f", x)
}

func (m Money) ToFloat() float64 {
	return float64(m) / 100.0
}

func ToMoney(m float64) Money {
	return Money((m * 100) + 0.5)
}

type TransactionEvent struct {
	UserID          UserID
	TransactionType TransactionType
	Amount          Money
	CreatedAt       time.Time
}
