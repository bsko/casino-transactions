package consumer

import (
	"sync"
	"testing"
	"time"

	"github.com/bsko/casino-transaction-system/internal/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatcher_Add(t *testing.T) {
	t.Run("successful flush when batch size reached", func(t *testing.T) {
		var flushedBatches [][]entity.TransactionEvent
		var mu sync.Mutex

		flushFunc := func(batch []entity.TransactionEvent) error {
			mu.Lock()
			defer mu.Unlock()
			flushedBatches = append(flushedBatches, append([]entity.TransactionEvent{}, batch...))
			return nil
		}

		batcher := NewBatcher(3, 100*time.Millisecond, flushFunc)

		events := []entity.TransactionEvent{
			{UserID: *entity.NewUserID(uuid.New()), TransactionType: entity.TransactionTypeBet, Amount: entity.ToMoney(10.0), CreatedAt: time.Now()},
			{UserID: *entity.NewUserID(uuid.New()), TransactionType: entity.TransactionTypeWin, Amount: entity.ToMoney(20.0), CreatedAt: time.Now()},
			{UserID: *entity.NewUserID(uuid.New()), TransactionType: entity.TransactionTypeBet, Amount: entity.ToMoney(30.0), CreatedAt: time.Now()},
		}

		for i, event := range events {
			err := batcher.Add(event)
			require.NoError(t, err)

			mu.Lock()
			flushedCount := len(flushedBatches)
			mu.Unlock()

			if i == len(events)-1 {
				assert.Equal(t, 1, flushedCount, "batch should be flushed when size reached")
				assert.Equal(t, 3, len(flushedBatches[0]), "flushed batch should contain all events")
			}
		}

		err := batcher.Close()
		require.NoError(t, err)
	})

	t.Run("flush on timeout and close with error handling", func(t *testing.T) {
		var flushedBatches [][]entity.TransactionEvent
		var mu sync.Mutex

		flushFunc := func(batch []entity.TransactionEvent) error {
			mu.Lock()
			defer mu.Unlock()
			flushedBatches = append(flushedBatches, append([]entity.TransactionEvent{}, batch...))
			return nil
		}

		batcher := NewBatcher(10, 50*time.Millisecond, flushFunc)

		event := entity.TransactionEvent{
			UserID:          *entity.NewUserID(uuid.New()),
			TransactionType: entity.TransactionTypeBet,
			Amount:          entity.ToMoney(10.0),
			CreatedAt:       time.Now(),
		}

		err := batcher.Add(event)
		require.NoError(t, err)

		mu.Lock()
		initialFlushedCount := len(flushedBatches)
		mu.Unlock()

		assert.Equal(t, 0, initialFlushedCount, "batch should not be flushed immediately")

		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		afterTimeoutFlushedCount := len(flushedBatches)
		mu.Unlock()

		assert.GreaterOrEqual(t, afterTimeoutFlushedCount, 1, "batch should be flushed after timeout")
		assert.Equal(t, 1, len(flushedBatches[0]), "flushed batch should contain the event")

		err = batcher.Close()
		assert.NoError(t, err, "close should succeed with empty batch")
	})
}
