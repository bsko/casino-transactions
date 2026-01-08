package producer

import (
	"context"
	"fmt"

	"github.com/bsko/casino-transaction-system/internal/config"
	"github.com/bsko/casino-transaction-system/internal/infrastructure/kafka"
	"github.com/bsko/casino-transaction-system/internal/services/producer"
)

const (
	producerConfigFilename = "configs/producer/config.yaml"
)

type ProducerApp struct {
	conf     *config.App
	kafka    kafkaAdapterInterface
	producer producerInterface
}

func (p *ProducerApp) Initialize(ctx context.Context) error {
	configReader := config.NewReader()
	conf, err := configReader.Read(producerConfigFilename)
	if err != nil {
		return fmt.Errorf("failed to init config: %w", err)
	}
	if conf.Kafka == nil {
		return fmt.Errorf("no kafka config provided")
	}

	kafkaAdapter := kafka.NewKafkaWriter(*conf.Kafka)
	err = kafkaAdapter.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to kafka: %w", err)
	}

	if conf.Producer == nil {
		return fmt.Errorf("no producer config provided")
	}
	producer := producer.NewProducer(*conf.Producer, kafkaAdapter)

	p.conf = conf
	p.kafka = kafkaAdapter
	p.producer = producer
	return nil
}

func (p *ProducerApp) Exec(ctx context.Context) error {
	return p.producer.Start(ctx)
}

func (p *ProducerApp) Shutdown(_ context.Context) error {
	var errs []error

	if p.kafka != nil {
		if err := p.kafka.Close(); err != nil {
			errs = append(errs, fmt.Errorf("kafka close error: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}
