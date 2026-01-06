package integration_tests

import (
	"context"
	"testing"
	"time"

	"github.com/bsko/casino-transaction-system/internal/entity"
	"github.com/bsko/casino-transaction-system/internal/infrastructure/repositories"
	"github.com/bsko/casino-transaction-system/internal/services/consumer"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGetTransactions(t *testing.T) {
	CleanupDB(t)

	ctx := context.Background()
	db := GetTestDB()
	dbInstance := repositories.NewDB(db)
	repo := repositories.NewTransactionEventRepository(dbInstance, dbInstance)
	service := consumer.NewGetListProcessor(repo)

	user1 := uuid.New()
	user2 := uuid.New()
	now := time.Now()
	baseTime := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())

	testEvents := []entity.TransactionEvent{
		{
			UserID:          entity.UserID{UUID: user1},
			TransactionType: entity.TransactionTypeBet,
			Amount:          entity.Money(10000), // $100.00
			CreatedAt:       baseTime.Add(-2 * time.Hour),
		},
		{
			UserID:          entity.UserID{UUID: user1},
			TransactionType: entity.TransactionTypeWin,
			Amount:          entity.Money(50000), // $500.00
			CreatedAt:       baseTime.Add(-1 * time.Hour),
		},
		{
			UserID:          entity.UserID{UUID: user2},
			TransactionType: entity.TransactionTypeBet,
			Amount:          entity.Money(20000), // $200.00
			CreatedAt:       baseTime,
		},
		{
			UserID:          entity.UserID{UUID: user2},
			TransactionType: entity.TransactionTypeWin,
			Amount:          entity.Money(100000), // $1000.00
			CreatedAt:       baseTime.Add(1 * time.Hour),
		},
		{
			UserID:          entity.UserID{UUID: user1},
			TransactionType: entity.TransactionTypeBet,
			Amount:          entity.Money(30000), // $300.00
			CreatedAt:       baseTime.Add(2 * time.Hour),
		},
	}

	err := repo.BatchStore(ctx, testEvents)
	require.NoError(t, err)

	t.Run("Get all events without filters", func(t *testing.T) {
		events, err := service.GetListByFilter(entity.TransactionEventFilter{
			Limit: 100,
		})
		require.NoError(t, err)
		require.Equal(t, 5, len(events), "Should return all 5 events")
	})

	t.Run("Filter by UserID", func(t *testing.T) {
		user1ID := entity.NewUserID(user1)
		events, err := service.GetListByFilter(entity.TransactionEventFilter{
			UserID: user1ID,
			Limit:  100,
		})
		require.NoError(t, err)
		require.Equal(t, 3, len(events), "Should return 3 events for user1")
		for _, event := range events {
			require.Equal(t, user1, event.UserID.UUID, "All events should belong to user1")
		}
	})

	t.Run("Filter by TransactionType", func(t *testing.T) {
		betType := entity.TransactionTypeBet
		events, err := service.GetListByFilter(entity.TransactionEventFilter{
			TransactionType: &betType,
			Limit:           100,
		})
		require.NoError(t, err)
		require.Equal(t, 3, len(events), "Should return 3 bet events")
		for _, event := range events {
			require.Equal(t, entity.TransactionTypeBet, event.TransactionType, "All events should be bets")
		}

		winType := entity.TransactionTypeWin
		events, err = service.GetListByFilter(entity.TransactionEventFilter{
			TransactionType: &winType,
			Limit:           100,
		})
		require.NoError(t, err)
		require.Equal(t, 2, len(events), "Should return 2 win events")
		for _, event := range events {
			require.Equal(t, entity.TransactionTypeWin, event.TransactionType, "All events should be wins")
		}
	})

	t.Run("Filter by Amount range", func(t *testing.T) {
		amountFrom := entity.Money(15000) // $150.00
		amountTo := entity.Money(40000)   // $400.00
		events, err := service.GetListByFilter(entity.TransactionEventFilter{
			AmountFrom: &amountFrom,
			AmountTo:   &amountTo,
			Limit:      100,
		})
		require.NoError(t, err)
		require.Equal(t, 2, len(events), "Should return 2 events with amount between $150 and $400")
		for _, event := range events {
			require.GreaterOrEqual(t, int64(event.Amount), int64(amountFrom), "Amount should be >= amountFrom")
			require.LessOrEqual(t, int64(event.Amount), int64(amountTo), "Amount should be <= amountTo")
		}
	})

	t.Run("Filter by CreatedAt range", func(t *testing.T) {
		createdFrom := baseTime.Add(-30 * time.Minute)
		createdTo := baseTime.Add(30 * time.Minute)
		events, err := service.GetListByFilter(entity.TransactionEventFilter{
			CreatedFrom: &createdFrom,
			CreatedTo:   &createdTo,
			Limit:       100,
		})
		require.NoError(t, err)
		require.Equal(t, 1, len(events), "Should return 1 event in time range")
		require.Equal(t, user2, events[0].UserID.UUID, "Should return user2's bet event")
	})

	t.Run("Filter with Limit", func(t *testing.T) {
		events, err := service.GetListByFilter(entity.TransactionEventFilter{
			Limit: 2,
		})
		require.NoError(t, err)
		require.Equal(t, 2, len(events), "Should return only 2 events due to limit")
	})

	t.Run("Filter with Offset", func(t *testing.T) {
		events1, err := service.GetListByFilter(entity.TransactionEventFilter{
			Limit:  2,
			Offset: 0,
		})
		require.NoError(t, err)

		events2, err := service.GetListByFilter(entity.TransactionEventFilter{
			Limit:  2,
			Offset: 2,
		})
		require.NoError(t, err)

		require.Equal(t, 2, len(events1), "First page should have 2 events")
		require.Equal(t, 2, len(events2), "Second page should have 2 events")
		if len(events1) > 0 && len(events2) > 0 {
			require.NotEqual(t, events1[0].CreatedAt, events2[0].CreatedAt, "Events should be different")
		}
	})

	t.Run("Combined filters: UserID and TransactionType", func(t *testing.T) {
		user1ID := entity.NewUserID(user1)
		betType := entity.TransactionTypeBet
		events, err := service.GetListByFilter(entity.TransactionEventFilter{
			UserID:          user1ID,
			TransactionType: &betType,
			Limit:           100,
		})
		require.NoError(t, err)
		require.Equal(t, 2, len(events), "Should return 2 bet events for user1")
		for _, event := range events {
			require.Equal(t, user1, event.UserID.UUID, "All events should belong to user1")
			require.Equal(t, entity.TransactionTypeBet, event.TransactionType, "All events should be bets")
		}
	})

	t.Run("Combined filters: UserID, TransactionType and Amount range", func(t *testing.T) {
		user2ID := entity.NewUserID(user2)
		winType := entity.TransactionTypeWin
		amountFrom := entity.Money(50000) // $500.00
		events, err := service.GetListByFilter(entity.TransactionEventFilter{
			UserID:          user2ID,
			TransactionType: &winType,
			AmountFrom:      &amountFrom,
			Limit:           100,
		})
		require.NoError(t, err)
		require.Equal(t, 1, len(events), "Should return 1 win event for user2 with amount >= $500")
		require.Equal(t, user2, events[0].UserID.UUID)
		require.Equal(t, entity.TransactionTypeWin, events[0].TransactionType)
		require.GreaterOrEqual(t, int64(events[0].Amount), int64(amountFrom))
	})
}
