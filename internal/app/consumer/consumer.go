package consumer

import (
	"context"
	"fmt"

	"github.com/bsko/casino-transaction-system/internal/config"
	"github.com/bsko/casino-transaction-system/internal/infrastructure/http"
	"github.com/bsko/casino-transaction-system/internal/infrastructure/kafka"
	"github.com/bsko/casino-transaction-system/internal/infrastructure/repositories"
	"github.com/bsko/casino-transaction-system/internal/services/consumer"
)

const (
	consumerConfigFilename = "configs/consumer-config.yaml"
)

type ConsumerApp struct {
	conf       *config.App
	kafka      kafkaReaderInterface
	db         transactionEventRepositoryInterface
	consumer   consumerInterface
	httpServer httpServer
}

func (p *ConsumerApp) Initialize(ctx context.Context) error {
	configReader := config.NewReader()
	conf, err := configReader.Read(consumerConfigFilename)
	if err != nil {
		return fmt.Errorf("failed to init config: %w", err)
	}
	if conf.Kafka == nil {
		return fmt.Errorf("no kafka config provided")
	}

	kafkaAdapter := kafka.NewKafkaReader(*conf.Kafka)
	err = kafkaAdapter.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to kafka: %w", err)
	}

	if conf.Postgres == nil {
		return fmt.Errorf("no postgres config provided")
	}

	dbAdapter := repositories.NewTransactionEventRepository(*conf.Postgres)
	err = dbAdapter.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}

	consumerService := consumer.NewConsumer(kafkaAdapter, dbAdapter)
	transactionsHandler := consumer.NewGetListProcessor(dbAdapter)
	httpServerInstance := http.NewHttpServer(transactionsHandler, conf.Http)

	p.conf = conf
	p.kafka = kafkaAdapter
	p.db = dbAdapter
	p.consumer = consumerService
	p.httpServer = httpServerInstance
	return nil
}

func (p *ConsumerApp) Exec(ctx context.Context) error {
	if p.consumer == nil {
		return fmt.Errorf("consumer is not initialized")
	}
	if p.httpServer == nil {
		return fmt.Errorf("http server is not initialized")
	}

	errChan := make(chan error, 1)

	go func() {
		if err := p.httpServer.Start(ctx); err != nil {
			errChan <- fmt.Errorf("http server error: %w", err)
		}
	}()

	go func() {
		if err := p.consumer.Start(ctx); err != nil {
			errChan <- fmt.Errorf("consumer error: %w", err)
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *ConsumerApp) Shutdown(ctx context.Context) error {
	var errs []error

	if p.httpServer != nil {
		if err := p.httpServer.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("htpp server close error: %w", err))
		}
	}

	if p.kafka != nil {
		if err := p.kafka.Close(); err != nil {
			errs = append(errs, fmt.Errorf("kafka close error: %w", err))
		}
	}

	if p.db != nil {
		if err := p.db.Close(); err != nil {
			errs = append(errs, fmt.Errorf("db close error: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}
