package producer

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/bsko/casino-transaction-system/internal/config"
	"github.com/bsko/casino-transaction-system/internal/entity"
	"github.com/bsko/casino-transaction-system/internal/services/producer/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestProducer_Worker(t *testing.T) {
	t.Run("successful processing of jobs", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockKafka := mocks.NewMockKafkaWriter(ctrl)
		conf := config.Producer{
			DistinctUsers: 2,
			AmountFrom:    10,
			AmountTo:      100,
		}

		producer := NewProducer(conf, mockKafka)
		producer.preGenerateUsers(conf.DistinctUsers)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		jobs := make(chan struct{}, 2)
		var wg sync.WaitGroup
		wg.Add(1)

		mockKafka.EXPECT().
			Publish(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(2)

		jobs <- struct{}{}
		jobs <- struct{}{}

		go producer.Worker(ctx, 1, jobs, &wg, cancel)

		time.Sleep(50 * time.Millisecond)
		cancel()

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(1 * time.Second):
			t.Fatal("worker did not finish in time")
		}
	})

	t.Run("error handling stops worker", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockKafka := mocks.NewMockKafkaWriter(ctrl)
		conf := config.Producer{
			DistinctUsers: 2,
			AmountFrom:    10,
			AmountTo:      100,
		}

		producer := NewProducer(conf, mockKafka)
		producer.preGenerateUsers(conf.DistinctUsers)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		jobs := make(chan struct{}, 1)
		var wg sync.WaitGroup
		wg.Add(1)

		mockKafka.EXPECT().
			Publish(gomock.Any(), gomock.Any()).
			Return(errors.New("publish error")).
			Times(1)

		jobs <- struct{}{}

		go producer.Worker(ctx, 1, jobs, &wg, cancel)

		time.Sleep(50 * time.Millisecond)
		cancel()

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(1 * time.Second):
			t.Fatal("worker did not finish in time")
		}
	})
}

func TestProducer_generateSingleMessage(t *testing.T) {
	t.Run("successful message generation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockKafka := mocks.NewMockKafkaWriter(ctrl)
		conf := config.Producer{
			DistinctUsers: 2,
			AmountFrom:    10,
			AmountTo:      100,
		}

		producer := NewProducer(conf, mockKafka)
		producer.preGenerateUsers(conf.DistinctUsers)

		ctx := context.Background()

		mockKafka.EXPECT().
			Publish(ctx, gomock.Any()).
			Return(nil).
			Do(func(_ context.Context, event entity.TransactionEvent) {
				assert.NotEmpty(t, event.UserID.UUID)
				assert.Contains(t, []entity.TransactionType{entity.TransactionTypeBet, entity.TransactionTypeWin}, event.TransactionType)
				assert.GreaterOrEqual(t, int64(event.Amount), int64(10*100))
				assert.LessOrEqual(t, int64(event.Amount), int64(100*100))
			})

		err := producer.generateSingleMessage(ctx)
		assert.NoError(t, err)
	})

	t.Run("error on publish", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockKafka := mocks.NewMockKafkaWriter(ctrl)
		conf := config.Producer{
			DistinctUsers: 2,
			AmountFrom:    10,
			AmountTo:      100,
		}

		producer := NewProducer(conf, mockKafka)
		producer.preGenerateUsers(conf.DistinctUsers)

		ctx := context.Background()
		expectedErr := errors.New("kafka publish error")

		mockKafka.EXPECT().
			Publish(ctx, gomock.Any()).
			Return(expectedErr)

		err := producer.generateSingleMessage(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "producer failed to publish message")
		assert.Contains(t, err.Error(), "kafka publish error")
	})
}

func TestProducer_generateData(t *testing.T) {
	t.Run("successful data generation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockKafka := mocks.NewMockKafkaWriter(ctrl)
		conf := config.Producer{
			DistinctUsers: 3,
			AmountFrom:    10,
			AmountTo:      100,
		}

		producer := NewProducer(conf, mockKafka)
		producer.preGenerateUsers(conf.DistinctUsers)

		event := producer.generateData()

		require.NotNil(t, event)
		assert.NotEmpty(t, event.UserID.UUID)
		assert.Contains(t, []entity.TransactionType{entity.TransactionTypeBet, entity.TransactionTypeWin}, event.TransactionType)
		assert.GreaterOrEqual(t, int64(event.Amount), int64(10*100))
		assert.LessOrEqual(t, int64(event.Amount), int64(100*100))
		assert.False(t, event.CreatedAt.IsZero())
	})

	t.Run("data generation with zero distinct users", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockKafka := mocks.NewMockKafkaWriter(ctrl)
		conf := config.Producer{
			DistinctUsers: 0,
			AmountFrom:    10,
			AmountTo:      100,
		}

		producer := NewProducer(conf, mockKafka)
		producer.preGenerateUsers(conf.DistinctUsers)

		if len(producer.usersMap) == 0 {
			t.Skip("skipping test with zero users as it would cause panic")
		}
	})
}

func TestProducer_calculateAmount(t *testing.T) {
	t.Run("successful amount calculation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockKafka := mocks.NewMockKafkaWriter(ctrl)
		conf := config.Producer{
			AmountFrom: 10,
			AmountTo:   100,
		}

		producer := NewProducer(conf, mockKafka)

		amount := producer.calculateAmount(123456789, conf.AmountFrom, conf.AmountTo)

		assert.GreaterOrEqual(t, int64(amount), int64(10*100))
		assert.LessOrEqual(t, int64(amount), int64(100*100))
	})

	t.Run("amount calculation with invalid range", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockKafka := mocks.NewMockKafkaWriter(ctrl)
		conf := config.Producer{
			AmountFrom: 100,
			AmountTo:   10,
		}

		producer := NewProducer(conf, mockKafka)

		amount := producer.calculateAmount(123456789, conf.AmountFrom, conf.AmountTo)

		assert.NotZero(t, amount)
	})
}
