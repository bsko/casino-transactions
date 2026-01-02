package producer

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"sync"
	"time"

	"github.com/bsko/casino-transaction-system/internal/entity"
)

func (p *Producer) Worker(ctx context.Context, id int, jobs <-chan struct{}, wg *sync.WaitGroup, cancel context.CancelFunc) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("finishing by context finished event: %d", id)
			return
		case _ = <-jobs:
			err := p.generateSingleMessage(ctx)
			if err != nil {
				fmt.Printf("failed to generate single message: %s", err)
				cancel()
			}
		}
	}
}

func (p *Producer) generateSingleMessage(ctx context.Context) error {
	err := p.kafka.Publish(ctx, *p.generateData())
	if err != nil {
		return fmt.Errorf("producer failed to publish message: %w", err)
	}
	return nil
}

func (p *Producer) generateData() *entity.TransactionEvent {
	randomInt, _ := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	simpleInt := randomInt.Int64()
	transactionType := entity.TransactionTypeBet
	if simpleInt%2 == 0 {
		transactionType = entity.TransactionTypeWin
	}
	userID := p.usersMap[int(simpleInt%int64(p.conf.DistinctUsers))]
	return &entity.TransactionEvent{
		UserID:          *entity.NewUserID(userID),
		TransactionType: transactionType,
		Amount:          p.calculateAmount(simpleInt, p.conf.AmountFrom, p.conf.AmountTo),
		CreatedAt:       time.Now(),
	}
}

func (p *Producer) calculateAmount(simpleInt int64, from int, to int) entity.Money {
	steps := (to - from) / 5
	index := int64(float64(simpleInt) / float64(math.MaxInt64) * float64(steps))
	return entity.ToMoney(float64(int64(from) + index*5))
}
