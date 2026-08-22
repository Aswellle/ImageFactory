package job

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// MemoryQueue is a channel-backed queue for single-instance Phase 1.
// It is NOT for production: jobs are lost on restart and not shared across
// instances. Phase 2 replaces this with a Redis-backed queue.
type MemoryQueue struct {
	ch     chan *Task
	logger *zap.Logger
}

// NewMemoryQueue builds a bounded in-memory queue.
func NewMemoryQueue(capacity int, logger *zap.Logger) *MemoryQueue {
	return &MemoryQueue{
		ch:     make(chan *Task, capacity),
		logger: logger,
	}
}

// Submit enqueues a task. Returns error if the queue is closed or context done.
func (q *MemoryQueue) Submit(ctx context.Context, t *Task) error {
	select {
	case q.ch <- t:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Stop closes the queue channel.
func (q *MemoryQueue) Stop() error {
	close(q.ch)
	return nil
}

// Tasks returns the receive channel for the processor.
func (q *MemoryQueue) Tasks() <-chan *Task {
	return q.ch
}

// QueueProcessor pulls tasks from a MemoryQueue and runs them concurrently.
type QueueProcessor struct {
	queue   *MemoryQueue
	workers int
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

var _ Processor = (*QueueProcessor)(nil)

// NewProcessor builds a QueueProcessor with N worker goroutines.
func NewProcessor(queue *MemoryQueue, workers int) Processor {
	return &QueueProcessor{queue: queue, workers: workers}
}

// Start launches worker goroutines.
func (p *QueueProcessor) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	for range p.workers {
		p.wg.Add(1)
		go p.worker(ctx)
	}
	return nil
}

// Stop signals workers to finish and waits.
func (p *QueueProcessor) Stop() error {
	if p.cancel != nil {
		p.cancel()
	}
	p.wg.Wait()
	return nil
}

func (p *QueueProcessor) worker(ctx context.Context) {
	defer p.wg.Done()
	for {
		select {
		case t, ok := <-p.queue.Tasks():
			if !ok {
				return
			}
			p.runTask(ctx, t)
		case <-ctx.Done():
			return
		}
	}
}

func (p *QueueProcessor) runTask(ctx context.Context, t *Task) {
	if t == nil || t.Run == nil {
		return
	}
	if err := t.Run(ctx); err != nil {
		p.queue.logger.Error("task failed",
			zap.String("job_id", t.Job.ID),
			zap.String("status", t.Job.Status),
			zap.Error(err),
		)
	}
}
