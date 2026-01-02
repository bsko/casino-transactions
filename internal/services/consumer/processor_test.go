package consumer

import (
	"testing"
	"time"

	"github.com/bsko/casino-transaction-system/internal/entity"
	"github.com/bsko/casino-transaction-system/internal/services/consumer/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetListProcessor_GetListByFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMocktransactionEventReadRepository(ctrl)
	processor := NewGetListProcessor(mockRepo)

	filter := entity.TransactionEventFilter{
		Limit: 10,
	}

	expectedEvents := []entity.TransactionEvent{
		{
			UserID:          *entity.NewUserID(uuid.New()),
			TransactionType: entity.TransactionTypeBet,
			Amount:          entity.ToMoney(100.0),
			CreatedAt:       time.Now(),
		},
	}

	mockRepo.EXPECT().
		GetListByFilter(filter).
		Return(expectedEvents, nil).
		Times(1)

	result, err := processor.GetListByFilter(filter)

	assert.NoError(t, err)
	assert.Equal(t, expectedEvents, result)
}
