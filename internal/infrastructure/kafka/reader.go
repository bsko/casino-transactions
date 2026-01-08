package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bsko/casino-transaction-system/api"
	"github.com/bsko/casino-transaction-system/internal/config"
	"github.com/bsko/casino-transaction-system/internal/entity"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"google.golang.org/protobuf/proto"
)

type KafkaReader struct {
	conf        config.Kafka
	reader      *kafka.Reader
	lastMessage kafka.Message
	mu          sync.Mutex
}

func NewKafkaReader(conf config.Kafka) *KafkaReader {
	return &KafkaReader{
		conf: conf,
	}
}

func (k *KafkaReader) Connect(_ context.Context) error {
	var dialer kafka.Dialer
	if k.conf.User != "" && k.conf.Password != "" {
		dialer.SASLMechanism = plain.Mechanism{
			Username: k.conf.User,
			Password: k.conf.Password,
		}
	}
	k.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{k.conf.ConnectionString},
		GroupID:        k.conf.GroupID,
		Topic:          k.conf.Topic,
		MinBytes:       1,
		MaxBytes:       10e6,
		MaxWait:        1 * time.Second,
		ReadBackoffMin: 100 * time.Millisecond,
		ReadBackoffMax: 1 * time.Second,
		Dialer:         &dialer,
	})
	return nil
}

func (k *KafkaReader) Read(ctx context.Context) (*entity.TransactionEvent, error) {
	if k.reader == nil {
		return nil, fmt.Errorf("kafka reader is not initialized, call Connect() first")
	}

	msg, err := k.reader.ReadMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read message from kafka: %w", err)
	}

	k.mu.Lock()
	k.lastMessage = msg
	k.mu.Unlock()

	var dto api.TransactionEvent
	if err = proto.Unmarshal(msg.Value, &dto); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protobuf: %w", err)
	}

	event, err := TransformFromDTO(&dto)
	if err != nil {
		return nil, fmt.Errorf("failed to transform DTO to event: %w", err)
	}

	return event, nil
}

func (k *KafkaReader) Commit(ctx context.Context) error {
	if k.reader == nil {
		return fmt.Errorf("kafka reader is not initialized")
	}

	k.mu.Lock()
	msg := k.lastMessage
	k.mu.Unlock()

	if msg.Topic == "" {
		return nil
	}

	return k.reader.CommitMessages(ctx, msg)
}

func (k *KafkaReader) Close() error {
	if k.reader != nil {
		if readerErr := k.reader.Close(); readerErr != nil {
			return fmt.Errorf("failed to close reader: %w", readerErr)
		}
	}
	return nil
}
