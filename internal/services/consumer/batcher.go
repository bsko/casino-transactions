package consumer

import (
	"sync"
	"time"

	"github.com/bsko/casino-transaction-system/internal/entity"
)

type Batcher struct {
	batch     []entity.TransactionEvent
	batchSize int
	timeout   time.Duration
	flushFunc func([]entity.TransactionEvent) error
	mu        sync.Mutex
	timer     *time.Timer
}

func NewBatcher(size int, timeout time.Duration, fn func([]entity.TransactionEvent) error) *Batcher {
	b := &Batcher{
		batch:     make([]entity.TransactionEvent, 0, size),
		batchSize: size,
		timeout:   timeout,
		flushFunc: fn,
	}
	b.resetTimer()
	return b
}

func (b *Batcher) Add(msg entity.TransactionEvent) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.batch = append(b.batch, msg)

	if len(b.batch) >= b.batchSize {
		return b.flush()
	}

	return nil
}

func (b *Batcher) flush() error {
	defer b.resetTimer()
	if len(b.batch) == 0 {
		return nil
	}

	err := b.flushFunc(b.batch)
	b.batch = b.batch[:0]
	return err
}

func (b *Batcher) resetTimer() {
	if b.timer != nil {
		b.timer.Stop()
	}
	b.timer = time.AfterFunc(b.timeout, func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		_ = b.flush()
	})
}

func (b *Batcher) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.timer != nil {
		b.timer.Stop()
	}
	return b.flush()
}
