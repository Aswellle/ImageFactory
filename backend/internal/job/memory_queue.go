package job

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// It is NOT for production: jobs are lost on restart and not shared across
// instances. Phase 2 replaces this with a Redis-backed queue.
type MemoryQueue struct {
	ch        chan *Task
	logger    *zap.Logger
	mu        sync.RWMutex
	isStopped bool
}

// NewMemoryQueue builds a bounded in-memory queue.
func NewMemoryQueue(capacity int, logger *zap.Logger) *MemoryQueue {
	if capacity <= 0 {
		capacity = 1000
	}
	return &MemoryQueue{
		ch:     make(chan *Task, capacity),
		logger: logger,
	}
}

// Submit enqueues a task. Returns error if the queue is closed or context done.
func (q *MemoryQueue) Submit(ctx context.Context, t *Task) error {
	q.mu.RLock()
	if q.isStopped {
		q.mu.RUnlock()
		return ErrQueueStopped
	}
	q.mu.RUnlock()

	select {
	case q.ch <- t:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Stop closes the queue channel.
func (q *MemoryQueue) Stop() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if !q.isStopped {
		q.isStopped = true
		close(q.ch)
	}
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
	wg      sync.WaitGroup
}

var _ Processor = (*QueueProcessor)(nil)

// NewProcessor builds a QueueProcessor with N worker goroutines.
func NewProcessor(queue *MemoryQueue, workers int) Processor {
	return &QueueProcessor{queue: queue, workers: workers}
}

// Start launches worker goroutines.
func (p *QueueProcessor) Start(ctx context.Context) error {
	if p.workers <= 0 {
		p.workers = 1
	}
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(ctx)
	}
	return nil
}

// Stop signals workers to finish and waits.
func (p *QueueProcessor) Stop() error {
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
	defer func() {
		if r := recover(); r != nil {
			p.queue.logger.Error("task panicked",
				zap.Any("panic", r),
				zap.String("job_id", func() string {
					if t != nil && t.Job != nil {
						return t.Job.ID
					}
					return "unknown"
				}()),
			)
		}
	}()

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

// ErrQueueStopped is returned when submitting to a stopped queue.
var ErrQueueStopped = context.DeadlineExceeded
