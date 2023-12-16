package threadpool

import (
	"context"
	"sync"
	"time"
)

// ThreadPool represents a simple thread pool.
type ThreadPool struct {
	maxWorkers int                              // Maximum number of workers in the pool.
	wg         sync.WaitGroup                   // WaitGroup to wait for all tasks to complete.
	queue      chan func(context.Context) error // Queue of tasks to be executed.

	stopOnError bool       // If true, stop the pool on the first encountered error.
	mu          sync.Mutex // Mutex to lock err
	err         error      // Holds the first encountered error.

	ctx    context.Context // Context with deadline for the pool.
	cancel context.CancelFunc

	// Internal state
	closed bool // threadpool already closed
}

type PoolOption func(*ThreadPool)

func StopOnError(stop bool) PoolOption {
	return func(tp *ThreadPool) {
		tp.stopOnError = stop
	}
}

func WithContext(ctx context.Context) PoolOption {
	return func(tp *ThreadPool) {
		tp.WithContext(ctx)
	}
}

func Timeout(timeout time.Duration) PoolOption {
	return func(tp *ThreadPool) {
		tp.SetTimeout(timeout)
	}
}

// NewThreadPool creates a new ThreadPool with the specified number of workers.
// If StopOnError is true(default is true), the threadpool will cancel all pending tasks
// if an error occurs.
func NewThreadPool(maxWorkers int, options ...PoolOption) *ThreadPool {
	tp := &ThreadPool{
		maxWorkers:  maxWorkers,
		queue:       make(chan func(context.Context) error, maxWorkers),
		stopOnError: true,
	}

	// Initialize the threadpool context.
	tp.ctx, tp.cancel = context.WithCancel(context.Background())

	// Apply options
	for _, option := range options {
		option(tp)
	}

	tp.start()
	return tp
}

// Start initializes and starts the worker pool.
func (tp *ThreadPool) start() {
	for i := 0; i < tp.maxWorkers; i++ {
		go tp.worker()
	}
}

func (tp *ThreadPool) worker() {
	for {
		select {
		case task, ok := <-tp.queue:
			if !ok {
				// Channel closed, terminate the worker
				return
			}
			err := task(tp.ctx) // Execute the task
			tp.wg.Done()

			tp.mu.Lock()
			if err != nil {
				// Set the first error and terminate the pool if stopOnError is true
				if tp.stopOnError && tp.err == nil {
					tp.err = err
					tp.cancel() // Cancel the context to signal other workers to terminate
				}
			}
			tp.mu.Unlock()

		case <-tp.ctx.Done():
			// Context canceled, terminate the worker
			return
		}
	}
}

// AddTask adds a task to the thread pool for execution.
func (tp *ThreadPool) AddTask(task func(context.Context) error) {
	select {
	case <-tp.ctx.Done():
		// Context canceled, don't add new tasks
		return
	default:
		tp.wg.Add(1)
		tp.queue <- task
	}
}

// Wait waits for all tasks in the thread pool to complete and returns the first error encountered.
func (tp *ThreadPool) Wait() error {
	// Do not close an already closed channel
	if tp.closed {
		return nil
	}
	tp.closed = true

	// close the channel, signal -> no more tasks accepted
	close(tp.queue)

	// wait for current tasks to complete
	tp.wg.Wait()

	tp.mu.Lock()
	defer tp.mu.Unlock()

	return tp.err
}

func (tp *ThreadPool) WithContext(ctx context.Context) {
	tp.ctx, tp.cancel = context.WithCancel(ctx)
}

// SetDeadline sets a deadline for the thread pool.
// After the deadline, the pool will cancel all pending tasks.
func (tp *ThreadPool) SetDeadline(deadline time.Time) {
	tp.ctx, tp.cancel = context.WithDeadline(tp.ctx, deadline)
}

func (tp *ThreadPool) SetTimeout(timeout time.Duration) {
	tp.ctx, tp.cancel = context.WithTimeout(tp.ctx, timeout)
}

func (tp *ThreadPool) NumWorkers() int {
	return tp.maxWorkers
}

// Cancel cancels the thread pool, terminating all workers and pending tasks.
func (tp *ThreadPool) Cancel() {
	tp.cancel()
}
