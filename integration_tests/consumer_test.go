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

func TestConsumer(t *testing.T) {
	CleanupDB(t)

	ctx := context.Background()
	db := GetTestDB()
	dbInstance := repositories.NewDB(db)
	repo := repositories.NewTransactionEventRepository(dbInstance, dbInstance)
	kafkaReaderMock := newKafkaReader()
	consumer := consumer.NewConsumer(kafkaReaderMock, repo)
	consumer.SetBatchSize(2)

	t.Run("Successful saving", func(t *testing.T) {
		cctx, cancel := context.WithCancel(ctx)
		defer cancel()

		errCh := make(chan error, 1)
		go func() {
			errCh <- consumer.Start(cctx)
		}()

		events := []entity.TransactionEvent{
			{
				UserID:          entity.UserID{UUID: uuid.New()},
				TransactionType: entity.TransactionTypeBet,
				Amount:          100,
				CreatedAt:       time.Now(),
			},
			{
				UserID:          entity.UserID{UUID: uuid.New()},
				TransactionType: entity.TransactionTypeBet,
				Amount:          150,
				CreatedAt:       time.Now(),
			},
			{
				UserID:          entity.UserID{UUID: uuid.New()},
				TransactionType: entity.TransactionTypeBet,
				Amount:          200,
				CreatedAt:       time.Now(),
			},
			{
				UserID:          entity.UserID{UUID: uuid.New()},
				TransactionType: entity.TransactionTypeWin,
				Amount:          1000,
				CreatedAt:       time.Now(),
			},
			{
				UserID:          entity.UserID{UUID: uuid.New()},
				TransactionType: entity.TransactionTypeWin,
				Amount:          2000,
				CreatedAt:       time.Now(),
			},
		}

		for _, event := range events {
			kafkaReaderMock.Send(event)
		}

		time.Sleep(100 * time.Millisecond)

		cancel()

		err := <-errCh
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		allEvents, err := repo.GetListByFilter(entity.TransactionEventFilter{Limit: 100})
		require.NoError(t, err)
		require.Equal(t, 5, len(allEvents), "Expected 5 events in database")
	})
}

type kafkaReader struct {
	events chan entity.TransactionEvent
}

func newKafkaReader() *kafkaReader {
	return &kafkaReader{
		events: make(chan entity.TransactionEvent, 100),
	}
}

func (k *kafkaReader) Read(ctx context.Context) (*entity.TransactionEvent, error) {
	select {
	case event := <-k.events:
		return &event, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (k *kafkaReader) Commit(_ context.Context) error {
	return nil
}

func (k *kafkaReader) Send(event entity.TransactionEvent) {
	k.events <- event
}
