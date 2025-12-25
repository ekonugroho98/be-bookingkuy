package worker

import (
	"context"
	"sync"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Job represents a background job
type Job struct {
	ID       string
	Name     string
	Handler  func(ctx context.Context) error
	Interval time.Duration
}

// Worker manages background jobs
type Worker struct {
	jobs     map[string]*Job
	running  bool
	mu       sync.RWMutex
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// New creates a new worker
func New() *Worker {
	return &Worker{
		jobs: make(map[string]*Job),
	}
}

// Register registers a job
func (w *Worker) Register(job *Job) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.jobs[job.ID] = job
	logger.Infof("Job registered: %s", job.Name)
}

// Start starts the worker
func (w *Worker) Start(ctx context.Context) {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}
	w.running = true
	ctx, w.cancel = context.WithCancel(ctx)
	w.mu.Unlock()

	logger.Info("Worker started")

	// Start each job
	for _, job := range w.jobs {
		w.wg.Add(1)
		go w.runJob(ctx, job)
	}
}

// Stop stops the worker
func (w *Worker) Stop() {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return
	}
	w.running = false
	w.mu.Unlock()

	logger.Info("Stopping worker...")
	w.cancel()
	w.wg.Wait()
	logger.Info("Worker stopped")
}

// runJob runs a job at interval
func (w *Worker) runJob(ctx context.Context, job *Job) {
	defer w.wg.Done()

	ticker := time.NewTicker(job.Interval)
	defer ticker.Stop()

	// Run immediately on start
	if err := job.Handler(ctx); err != nil {
		logger.Errorf("Job %s failed: %v", job.Name, err)
	}

	for {
		select {
		case <-ctx.Done():
			logger.Infof("Job %s stopped", job.Name)
			return
		case <-ticker.C:
			if err := job.Handler(ctx); err != nil {
				logger.Errorf("Job %s failed: %v", job.Name, err)
			} else {
				logger.Infof("Job %s completed successfully", job.Name)
			}
		}
	}
}

// ExecuteOnce executes a job once immediately
func (w *Worker) ExecuteOnce(ctx context.Context, jobID string) error {
	w.mu.RLock()
	job, exists := w.jobs[jobID]
	w.mu.RUnlock()

	if !exists {
		return logger.Errorf("Job not found: %s", jobID)
	}

	return job.Handler(ctx)
}
