package collector

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"distributed-tracing-system/internal/model"
	"distributed-tracing-system/internal/storage"
)

var ErrQueueFull = errors.New("collector queue is full")
var ErrClosed = errors.New("collector is closed")

type Metrics struct {
	Received    uint64 `json:"received"`
	Dropped     uint64 `json:"dropped"`
	Written     uint64 `json:"written"`
	Failed      uint64 `json:"failed"`
	QueueLength int    `json:"queue_length"`
	LastUpload  int64  `json:"last_upload"`
}

type Collector struct {
	store    storage.Storage
	queue    chan []model.Span
	done     chan struct{}
	workers  sync.WaitGroup
	close    sync.Once
	gate     sync.RWMutex
	received atomic.Uint64
	dropped  atomic.Uint64
	written  atomic.Uint64
	failed   atomic.Uint64
	last     atomic.Int64
}

func New(store storage.Storage, workers, queueSize int) *Collector {
	c := &Collector{
		store: store,
		queue: make(chan []model.Span, queueSize),
		done:  make(chan struct{}),
	}
	for i := 0; i < workers; i++ {
		c.workers.Add(1)
		go c.run()
	}
	return c
}

func (c *Collector) Submit(spans []model.Span) error {
	if err := ValidateBatch(spans); err != nil {
		return err
	}
	batch := Deduplicate(spans)
	c.gate.RLock()
	defer c.gate.RUnlock()
	select {
	case <-c.done:
		return ErrClosed
	default:
	}
	select {
	case c.queue <- batch:
		c.received.Add(uint64(len(batch)))
		c.last.Store(time.Now().UnixNano())
		return nil
	default:
		c.dropped.Add(uint64(len(batch)))
		return nil
	}
}

func (c *Collector) run() {
	defer c.workers.Done()
	for batch := range c.queue {
		if err := c.writeWithRetry(batch); err != nil {
			c.failed.Add(uint64(len(batch)))
			continue
		}
		c.written.Add(uint64(len(batch)))
	}
}

func (c *Collector) writeWithRetry(batch []model.Span) error {
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = c.store.Put(ctx, batch)
		cancel()
		if err == nil {
			return nil
		}
		time.Sleep(time.Duration(1<<attempt) * 20 * time.Millisecond)
	}
	return err
}

func (c *Collector) Metrics() Metrics {
	return Metrics{
		Received:    c.received.Load(),
		Dropped:     c.dropped.Load(),
		Written:     c.written.Load(),
		Failed:      c.failed.Load(),
		QueueLength: len(c.queue),
		LastUpload:  c.last.Load(),
	}
}

func (c *Collector) Close(ctx context.Context) error {
	c.close.Do(func() {
		c.gate.Lock()
		defer c.gate.Unlock()
		close(c.done)
		close(c.queue)
	})
	waited := make(chan struct{})
	go func() {
		c.workers.Wait()
		close(waited)
	}()
	select {
	case <-waited:
		return c.store.Close()
	case <-ctx.Done():
		return ctx.Err()
	}
}
