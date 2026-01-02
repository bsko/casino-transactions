package consumer

import "github.com/bsko/casino-transaction-system/internal/entity"

type GetListProcessor struct {
	transactionEventRepository transactionEventReadRepository
}

func NewGetListProcessor(transactionEventRepository transactionEventReadRepository) *GetListProcessor {
	return &GetListProcessor{
		transactionEventRepository: transactionEventRepository,
	}
}

func (s *GetListProcessor) GetListByFilter(filter entity.TransactionEventFilter) ([]entity.TransactionEvent, error) {
	// some business logic here: validation? & transform db errors to service layer
	return s.transactionEventRepository.GetListByFilter(filter)
}
