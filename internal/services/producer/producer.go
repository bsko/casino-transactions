package producer

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/bsko/casino-transaction-system/internal/config"
	"github.com/google/uuid"
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

	workersCount := getWorkersCount(p.conf.CreationRPS)
	jobs := make(chan struct{}, p.conf.CreationRPS*2)
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
				log.Printf("Shutting down producer goroutine\n")
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

func getWorkersCount(creationRPS int) int {
	if creationRPS > 10 {
		return creationRPS
	}
	return 10
}
