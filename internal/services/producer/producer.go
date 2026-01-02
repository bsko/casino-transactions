package producer

import (
	"context"
	"sync"
	"time"

	"github.com/bsko/casino-transaction-system/internal/config"
	"github.com/google/uuid"
)

const (
	workersCount = 3
)

type Producer struct {
	conf     config.Producer
	kafka    KafkaWriter
	usersMap map[int]uuid.UUID
}

func NewProducer(conf config.Producer, kafka KafkaWriter) *Producer {
	return &Producer{
		conf:     conf,
		kafka:    kafka,
		usersMap: make(map[int]uuid.UUID),
	}
}

func (p *Producer) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	p.preGenerateUsers(p.conf.DistinctUsers)

	jobs := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(workersCount)
	for w := 1; w <= workersCount; w++ {
		go p.Worker(ctx, w, jobs, &wg, cancel)
	}

	if p.conf.InitialBatchSize > 0 {
		for i := 0; i < p.conf.InitialBatchSize; i++ {
			jobs <- struct{}{}
		}
	}
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				for i := 0; i < p.conf.CreationRPS; i++ {
					jobs <- struct{}{}
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	wg.Wait()
	return nil
}

func (p *Producer) preGenerateUsers(distinctUsers int) {
	for i := 0; i < distinctUsers; i++ {
		p.usersMap[i] = uuid.New()
	}
}
