package concurrency

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

// Task represents a unit of work to be executed
type Task interface {
	Execute(ctx context.Context) (interface{}, error)
	GetID() string
	GetPriority() int
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	TaskID string
	Result interface{}
	Error  error
	DoneAt time.Time
}

// Worker represents a goroutine that processes tasks
type Worker struct {
	id        int
	taskQueue chan Task
	resultCh  chan TaskResult
	quit      chan bool
	wg        *sync.WaitGroup
}

// NewWorker creates a new worker
func NewWorker(id int, taskQueue chan Task, resultCh chan TaskResult, wg *sync.WaitGroup) *Worker {
	return &Worker{
		id:        id,
		taskQueue: taskQueue,
		resultCh:  resultCh,
		quit:      make(chan bool),
		wg:        wg,
	}
}

// Start begins the worker's main loop
func (w *Worker) Start(ctx context.Context) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			select {
			case task := <-w.taskQueue:
				startTime := time.Now()
				result, err := task.Execute(ctx)
				w.resultCh <- TaskResult{
					TaskID: task.GetID(),
					Result: result,
					Error:  err,
					DoneAt: startTime,
				}
			case <-w.quit:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop signals the worker to stop
func (w *Worker) Stop() {
	w.quit <- true
}

// WorkerPool manages a pool of workers
type WorkerPool struct {
	workers     []*Worker
	taskQueue   chan Task
	resultCh    chan TaskResult
	workerCount int
	wg          *sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	metrics     *PoolMetrics
}

// PoolMetrics tracks worker pool performance
type PoolMetrics struct {
	TasksSubmitted  int64
	TasksCompleted  int64
	TasksFailed     int64
	AverageTaskTime time.Duration
	ActiveWorkers   int
	mu              sync.RWMutex
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workerCount int, queueSize int) *WorkerPool {
	if workerCount <= 0 {
		workerCount = runtime.NumCPU()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		workers:     make([]*Worker, workerCount),
		taskQueue:   make(chan Task, queueSize),
		resultCh:    make(chan TaskResult, queueSize),
		workerCount: workerCount,
		wg:          &sync.WaitGroup{},
		ctx:         ctx,
		cancel:      cancel,
		metrics:     &PoolMetrics{},
	}
}

// Start initializes and starts all workers
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workerCount; i++ {
		worker := NewWorker(i, wp.taskQueue, wp.resultCh, wp.wg)
		wp.workers[i] = worker
		worker.Start(wp.ctx)
	}

	// Start metrics collector
	go wp.collectMetrics()
}

// Stop gracefully shuts down the worker pool
func (wp *WorkerPool) Stop() {
	wp.cancel()

	// Stop all workers
	for _, worker := range wp.workers {
		worker.Stop()
	}

	// Wait for all workers to finish
	wp.wg.Wait()

	close(wp.taskQueue)
	close(wp.resultCh)
}

// SubmitTask adds a task to the queue
func (wp *WorkerPool) SubmitTask(task Task) error {
	select {
	case wp.taskQueue <- task:
		wp.metrics.incrementSubmitted()
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is shutting down")
	default:
		return fmt.Errorf("task queue is full")
	}
}

// SubmitTaskWithTimeout adds a task with timeout
func (wp *WorkerPool) SubmitTaskWithTimeout(task Task, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(wp.ctx, timeout)
	defer cancel()

	select {
	case wp.taskQueue <- task:
		wp.metrics.incrementSubmitted()
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout submitting task: %w", ctx.Err())
	}
}

// GetResultChannel returns the result channel
func (wp *WorkerPool) GetResultChannel() <-chan TaskResult {
	return wp.resultCh
}

// GetMetrics returns current pool metrics
func (wp *WorkerPool) GetMetrics() PoolMetrics {
	wp.metrics.mu.RLock()
	defer wp.metrics.mu.RUnlock()
	return *wp.metrics
}

// collectMetrics updates pool metrics
func (wp *WorkerPool) collectMetrics() {
	for result := range wp.resultCh {
		wp.metrics.recordResult(result)
	}
}

// incrementSubmitted increments the submitted task count
func (pm *PoolMetrics) incrementSubmitted() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.TasksSubmitted++
}

// recordResult records a completed task result
func (pm *PoolMetrics) recordResult(result TaskResult) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if result.Error != nil {
		pm.TasksFailed++
	} else {
		pm.TasksCompleted++
	}
}

// FanOutFanIn implements the fan-out/fan-in pattern
type FanOutFanIn struct {
	input       chan interface{}
	output      chan interface{}
	workers     int
	processFunc func(context.Context, interface{}) (interface{}, error)
	ctx         context.Context
	cancel      context.CancelFunc
	wg          *sync.WaitGroup
}

// NewFanOutFanIn creates a new fan-out/fan-in processor
func NewFanOutFanIn(workers int, bufferSize int, processFunc func(context.Context, interface{}) (interface{}, error)) *FanOutFanIn {
	ctx, cancel := context.WithCancel(context.Background())

	return &FanOutFanIn{
		input:       make(chan interface{}, bufferSize),
		output:      make(chan interface{}, bufferSize),
		workers:     workers,
		processFunc: processFunc,
		ctx:         ctx,
		cancel:      cancel,
		wg:          &sync.WaitGroup{},
	}
}

// Start begins the fan-out/fan-in processing
func (fofi *FanOutFanIn) Start() {
	for i := 0; i < fofi.workers; i++ {
		fofi.wg.Add(1)
		go fofi.worker(i)
	}
}

// Stop stops the processor
func (fofi *FanOutFanIn) Stop() {
	fofi.cancel()
	fofi.wg.Wait()
	close(fofi.input)
	close(fofi.output)
}

// Input returns the input channel
func (fofi *FanOutFanIn) Input() chan<- interface{} {
	return fofi.input
}

// Output returns the output channel
func (fofi *FanOutFanIn) Output() <-chan interface{} {
	return fofi.output
}

// worker processes items from input to output
func (fofi *FanOutFanIn) worker(id int) {
	defer fofi.wg.Done()

	for {
		select {
		case item, ok := <-fofi.input:
			if !ok {
				return
			}

			result, err := fofi.processFunc(fofi.ctx, item)
			if err != nil {
				log.Printf("Worker %d failed to process item: %v", id, err)
				continue
			}

			select {
			case fofi.output <- result:
			case <-fofi.ctx.Done():
				return
			}

		case <-fofi.ctx.Done():
			return
		}
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	maxFailures   int
	resetTimeout  time.Duration
	failures      int
	lastFailTime  time.Time
	state         CircuitState
	mu            sync.RWMutex
	onStateChange func(CircuitState)
}

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

func (cs CircuitState) String() string {
	switch cs {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}

// Execute runs a function through the circuit breaker
func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	if !cb.allowRequest() {
		return nil, fmt.Errorf("circuit breaker is OPEN")
	}

	result, err := fn()
	if err != nil {
		cb.recordFailure()
		return nil, err
	}

	cb.recordSuccess()
	return result, nil
}

// allowRequest determines if a request should be allowed
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		return time.Since(cb.lastFailTime) > cb.resetTimeout
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// recordFailure records a failure
func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.failures >= cb.maxFailures {
		cb.setState(StateOpen)
	}
}

// recordSuccess records a success
func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		cb.setState(StateClosed)
	}
	cb.failures = 0
}

// setState changes the circuit breaker state
func (cb *CircuitBreaker) setState(state CircuitState) {
	if cb.state != state {
		cb.state = state
		if cb.onStateChange != nil {
			cb.onStateChange(state)
		}
	}
}

// GetState returns the current state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// SetStateChangeCallback sets a callback for state changes
func (cb *CircuitBreaker) SetStateChangeCallback(callback func(CircuitState)) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.onStateChange = callback
}

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	tokens     chan struct{}
	refillRate time.Duration
	tokenCount int
	maxTokens  int
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens int, refillRate time.Duration) *RateLimiter {
	rl := &RateLimiter{
		tokens:     make(chan struct{}, maxTokens),
		refillRate: refillRate,
		tokenCount: maxTokens,
		maxTokens:  maxTokens,
		lastRefill: time.Now(),
	}

	// Initialize tokens
	for i := 0; i < maxTokens; i++ {
		rl.tokens <- struct{}{}
	}

	// Start refill goroutine
	go rl.refill()

	return rl
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

// Wait blocks until a token is available
func (rl *RateLimiter) Wait() {
	<-rl.tokens
}

// WaitWithTimeout blocks until a token is available or timeout
func (rl *RateLimiter) WaitWithTimeout(timeout time.Duration) bool {
	select {
	case <-rl.tokens:
		return true
	case <-time.After(timeout):
		return false
	}
}

// refill adds tokens to the bucket
func (rl *RateLimiter) refill() {
	ticker := time.NewTicker(rl.refillRate)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		if rl.tokenCount < rl.maxTokens {
			select {
			case rl.tokens <- struct{}{}:
				rl.tokenCount++
			default:
				// Channel is full
			}
		}
		rl.mu.Unlock()
	}
}
