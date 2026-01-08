package consumer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bsko/casino-transaction-system/internal/entity"
	"github.com/bsko/casino-transaction-system/internal/services/consumer/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestConsumer_Start(t *testing.T) {
	t.Run("successful processing and context cancellation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockReader := mocks.NewMockkafkaReader(ctrl)
		mockRepo := mocks.NewMocktransactionEventSaveRepository(ctrl)

		consumer := NewConsumer(mockReader, mockRepo)

		event := &entity.TransactionEvent{
			UserID:          *entity.NewUserID(uuid.New()),
			TransactionType: entity.TransactionTypeBet,
			Amount:          entity.ToMoney(10.0),
			CreatedAt:       time.Now(),
		}

		ctx, cancel := context.WithCancel(context.Background())

		mockReader.EXPECT().
			Read(gomock.Any()).
			Return(event, nil).
			AnyTimes()

		mockRepo.EXPECT().
			BatchStore(gomock.Any(), gomock.Any()).
			Return(nil).
			AnyTimes()

		mockReader.EXPECT().
			Commit(gomock.Any()).
			Return(nil).
			AnyTimes()

		errChan := make(chan error, 1)
		go func() {
			errChan <- consumer.Start(ctx)
		}()

		time.Sleep(50 * time.Millisecond)
		cancel()

		select {
		case err := <-errChan:
			assert.NoError(t, err)
		case <-time.After(1 * time.Second):
			t.Fatal("test timeout")
		}
	})

	t.Run("error on read returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockReader := mocks.NewMockkafkaReader(ctrl)
		mockRepo := mocks.NewMocktransactionEventSaveRepository(ctrl)

		consumer := NewConsumer(mockReader, mockRepo)

		ctx, cancel := context.WithCancel(context.Background())
		expectedErr := errors.New("kafka read error")

		mockReader.EXPECT().
			Read(gomock.Any()).
			Return(nil, expectedErr).
			Do(func(_ context.Context) {
				go func() {
					time.Sleep(10 * time.Millisecond)
					cancel()
				}()
			}).
			AnyTimes()

		err := consumer.Start(ctx)
		assert.NoError(t, err)
	})

	t.Run("error on batcher add returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockReader := mocks.NewMockkafkaReader(ctrl)
		mockRepo := mocks.NewMocktransactionEventSaveRepository(ctrl)

		consumer := NewConsumer(mockReader, mockRepo)

		event := &entity.TransactionEvent{
			UserID:          *entity.NewUserID(uuid.New()),
			TransactionType: entity.TransactionTypeBet,
			Amount:          entity.ToMoney(10.0),
			CreatedAt:       time.Now(),
		}

		ctx := context.Background()
		expectedErr := errors.New("batch store error")

		mockReader.EXPECT().
			Read(gomock.Any()).
			Return(event, nil).
			Times(100)

		mockRepo.EXPECT().
			BatchStore(gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, batch []entity.TransactionEvent) {
				assert.Equal(t, 100, len(batch))
			}).
			Return(expectedErr).
			Times(1)

		err := consumer.Start(ctx)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}
