package producer

import (
	"context"
	"testing"
	"time"

	"github.com/bsko/casino-transaction-system/internal/config"
	"github.com/bsko/casino-transaction-system/internal/services/producer/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestProducer_Start(t *testing.T) {
	t.Run("successful start with initial batch and ticker", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockKafka := mocks.NewMockKafkaWriter(ctrl)
		conf := config.Producer{
			InitialBatchSize: 2,
			CreationRPS:      1,
			DistinctUsers:    2,
			AmountFrom:       10,
			AmountTo:         100,
		}

		producer := NewProducer(conf, mockKafka)

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		mockKafka.EXPECT().
			Publish(gomock.Any(), gomock.Any()).
			Return(nil).
			AnyTimes()

		errChan := make(chan error, 1)
		go func() {
			errChan <- producer.Start(ctx)
		}()

		select {
		case err := <-errChan:
			assert.NoError(t, err)
		case <-time.After(500 * time.Millisecond):
			t.Fatal("test timeout")
		}

		assert.Equal(t, conf.DistinctUsers, len(producer.usersMap))
	})

	t.Run("context cancellation stops producer", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockKafka := mocks.NewMockKafkaWriter(ctrl)
		conf := config.Producer{
			InitialBatchSize: 1,
			CreationRPS:      2,
			DistinctUsers:    1,
			AmountFrom:       10,
			AmountTo:         100,
		}

		producer := NewProducer(conf, mockKafka)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		mockKafka.EXPECT().
			Publish(gomock.Any(), gomock.Any()).
			Return(nil).
			AnyTimes()

		err := producer.Start(ctx)
		assert.NoError(t, err)
		assert.Equal(t, conf.DistinctUsers, len(producer.usersMap))
	})
}
