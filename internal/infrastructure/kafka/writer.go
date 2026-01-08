package kafka

import (
	"context"
	"fmt"

	"github.com/bsko/casino-transaction-system/internal/config"
	"github.com/bsko/casino-transaction-system/internal/entity"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"google.golang.org/protobuf/proto"
)

type KafkaWriter struct {
	conf   config.Kafka
	writer *kafka.Writer
}

func NewKafkaWriter(conf config.Kafka) *KafkaWriter {
	return &KafkaWriter{
		conf: conf,
	}
}

func (k *KafkaWriter) Connect(ctx context.Context) error {
	var dialer kafka.Dialer
	if k.conf.User != "" && k.conf.Password != "" {
		dialer.SASLMechanism = plain.Mechanism{
			Username: k.conf.User,
			Password: k.conf.Password,
		}
	}

	k.writer = &kafka.Writer{
		Addr:         kafka.TCP(k.conf.ConnectionString),
		Topic:        k.conf.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequiredAcks(k.conf.RequiredAcks),
		MaxAttempts:  k.conf.MaxAttempts,
	}
	return nil
}

func (k *KafkaWriter) Publish(ctx context.Context, event entity.TransactionEvent) error {
	if k.writer == nil {
		return fmt.Errorf("kafka writer is not initialized, call Connect() first")
	}

	dto, err := TransformToDTO(event)
	if err != nil {
		return fmt.Errorf("failed to transform event to DTO: %w", err)
	}

	data, err := proto.Marshal(dto)
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf: %w", err)
	}

	err = k.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.UserID.UUID.String()),
		Value: data,
	})
	if err != nil {
		return fmt.Errorf("failed to write message to kafka: %w", err)
	}

	return nil
}

func (k *KafkaWriter) Close() error {
	if k.writer != nil {
		if writerErr := k.writer.Close(); writerErr != nil {
			return fmt.Errorf("failed to close writer: %w", writerErr)
		}
	}
	return nil
}
