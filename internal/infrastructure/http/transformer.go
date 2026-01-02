package http

import (
	"fmt"

	"github.com/bsko/casino-transaction-system/internal/entity"
	"github.com/google/uuid"
)

func TransformRequestToFilter(req TransactionSearchRequest) (entity.TransactionEventFilter, error) {
	filter := entity.TransactionEventFilter{
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	if req.UserID != nil && *req.UserID != "" {
		parsedUUID, err := uuid.Parse(*req.UserID)
		if err != nil {
			return filter, fmt.Errorf("invalid user_id format: %w", err)
		}
		filter.UserID = entity.NewUserID(parsedUUID)
	}

	if req.TransactionType != nil && *req.TransactionType != "" && *req.TransactionType != "all" {
		transactionType := entity.TransactionType(*req.TransactionType)
		filter.TransactionType = &transactionType
	}

	if req.AmountFrom != nil {
		amount := entity.ToMoney(*req.AmountFrom)
		filter.AmountFrom = &amount
	}

	if req.AmountTo != nil {
		amount := entity.ToMoney(*req.AmountTo)
		filter.AmountTo = &amount
	}

	if req.CreatedFrom != nil {
		filter.CreatedFrom = req.CreatedFrom
	}

	if req.CreatedTo != nil {
		filter.CreatedTo = req.CreatedTo
	}

	return filter, nil
}

func TransformTransactionsToResponse(transactions []entity.TransactionEvent, limit, offset int) TransactionListResponse {
	resultList := make([]TransactionDTO, 0, len(transactions))

	for _, transaction := range transactions {
		resultList = append(resultList, TransactionDTO{
			UserID:          transaction.UserID.UUID.String(),
			TransactionType: string(transaction.TransactionType),
			Amount:          transaction.Amount.ToFloat(),
			Timestamp:       transaction.CreatedAt,
		})
	}

	return TransactionListResponse{
		Total:        len(resultList),
		Limit:        limit,
		Offset:       offset,
		Transactions: resultList,
	}
}
