package http

import (
	"testing"
	"time"

	"github.com/bsko/casino-transaction-system/internal/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTransformRequestToFilter(t *testing.T) {
	t.Run("successful transformation with all fields", func(t *testing.T) {
		userIDStr := uuid.New().String()
		transactionType := "bet"
		amountFrom := 10.5
		amountTo := 100.0
		createdFrom := time.Now().Add(-24 * time.Hour)
		createdTo := time.Now()

		req := TransactionSearchRequest{
			UserID:          &userIDStr,
			TransactionType: &transactionType,
			AmountFrom:      &amountFrom,
			AmountTo:        &amountTo,
			CreatedFrom:     &createdFrom,
			CreatedTo:       &createdTo,
			Limit:           50,
			Offset:          10,
		}

		filter, err := TransformRequestToFilter(req)

		assert.NoError(t, err)
		assert.NotNil(t, filter.UserID)
		assert.Equal(t, userIDStr, filter.UserID.UUID.String())
		assert.NotNil(t, filter.TransactionType)
		assert.Equal(t, entity.TransactionType("bet"), *filter.TransactionType)
		assert.NotNil(t, filter.AmountFrom)
		assert.NotNil(t, filter.AmountTo)
		assert.Equal(t, &createdFrom, filter.CreatedFrom)
		assert.Equal(t, &createdTo, filter.CreatedTo)
		assert.Equal(t, 50, filter.Limit)
		assert.Equal(t, 10, filter.Offset)
	})
}

func TestTransformTransactionsToResponse(t *testing.T) {
	t.Run("successful transformation to response", func(t *testing.T) {
		userID1 := uuid.New()
		userID2 := uuid.New()
		now := time.Now()

		transactions := []entity.TransactionEvent{
			{
				UserID:          *entity.NewUserID(userID1),
				TransactionType: entity.TransactionTypeBet,
				Amount:          entity.ToMoney(50.0),
				CreatedAt:       now,
			},
			{
				UserID:          *entity.NewUserID(userID2),
				TransactionType: entity.TransactionTypeWin,
				Amount:          entity.ToMoney(100.0),
				CreatedAt:       now.Add(time.Hour),
			},
		}

		limit := 10
		offset := 0

		response := TransformTransactionsToResponse(transactions, limit, offset)

		assert.Equal(t, 2, response.Total)
		assert.Equal(t, limit, response.Limit)
		assert.Equal(t, offset, response.Offset)
		assert.Len(t, response.Transactions, 2)
		assert.Equal(t, userID1.String(), response.Transactions[0].UserID)
		assert.Equal(t, "bet", response.Transactions[0].TransactionType)
		assert.Equal(t, 50.0, response.Transactions[0].Amount)
		assert.Equal(t, now, response.Transactions[0].Timestamp)
		assert.Equal(t, userID2.String(), response.Transactions[1].UserID)
		assert.Equal(t, "win", response.Transactions[1].TransactionType)
		assert.Equal(t, 100.0, response.Transactions[1].Amount)
	})
}
