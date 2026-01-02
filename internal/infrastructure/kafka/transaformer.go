package kafka

import (
	"fmt"

	"github.com/bsko/casino-transaction-system/api"
	"github.com/bsko/casino-transaction-system/internal/entity"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	dtoToEntityTypeMap = map[api.TransactionType]entity.TransactionType{
		api.TransactionType_TRANSACTION_TYPE_BET: entity.TransactionTypeBet,
		api.TransactionType_TRANSACTION_TYPE_WIN: entity.TransactionTypeWin,
	}

	entityToDTOTypeMap = map[entity.TransactionType]api.TransactionType{
		entity.TransactionTypeBet: api.TransactionType_TRANSACTION_TYPE_BET,
		entity.TransactionTypeWin: api.TransactionType_TRANSACTION_TYPE_WIN,
	}
)

func TransformFromDTO(dto *api.TransactionEvent) (*entity.TransactionEvent, error) {
	userId, err := uuid.Parse(dto.UserId)
	if err != nil {
		return nil, err
	}

	transactionType, ok := dtoToEntityTypeMap[dto.TransactionType]
	if !ok {
		return nil, fmt.Errorf("invalid transaction type: %s", dto.TransactionType)
	}

	return &entity.TransactionEvent{
		UserID: entity.UserID{
			UUID: userId,
		},
		TransactionType: transactionType,
		Amount:          entity.ToMoney(dto.Amount),
		CreatedAt:       dto.Timestamp.AsTime(),
	}, nil
}

func TransformToDTO(event entity.TransactionEvent) (*api.TransactionEvent, error) {
	transactionType, ok := entityToDTOTypeMap[event.TransactionType]
	if !ok {
		return nil, fmt.Errorf("invalid transaction type: %s", transactionType)
	}

	return &api.TransactionEvent{
		UserId:          event.UserID.UUID.String(),
		TransactionType: transactionType,
		Amount:          event.Amount.ToFloat(),
		Timestamp:       timestamppb.New(event.CreatedAt),
	}, nil
}
