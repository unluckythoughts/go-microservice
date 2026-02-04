package worker

import (
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/unluckythoughts/go-microservice/v2/tools/web"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TaskFunc is a function type for background tasks
type TaskFunc func(ctx web.Context) error

// Worker manages background tasks and cron jobs
type Worker struct {
	cron      *cron.Cron
	ctx       web.Context
	db        *gorm.DB
	wg        sync.WaitGroup
	enableDL  bool // enable distributed locking
	tasks     map[string]cron.EntryID
	taskMutex sync.RWMutex
}

// New creates a new Worker instance
func New(c web.Context, db *gorm.DB) *Worker {
	w := &Worker{
		cron:  cron.New(cron.WithSeconds()),
		ctx:   c,
		tasks: make(map[string]cron.EntryID),
	}

	if db != nil {
		// Enable distributed locking if a database is postgres
		if db.Dialector.Name() == "postgres" {
			c.Logger().Info("Distributed locking enabled for worker tasks")
			w.db = db
			w.enableDL = true
		}
	}
	return w
}

func (w *Worker) Logger(name string) *zap.SugaredLogger {
	l := w.ctx.Logger().With("task", name)
	return l
}

// acquireLock tries to acquire a PostgreSQL advisory lock for the given task
// Returns true if lock was acquired, false otherwise
func (w *Worker) acquireLock(taskName string) bool {
	if w.db == nil {
		// No database available, allow task to run (single instance mode)
		return true
	}

	// Generate a unique lock ID based on task name using hash
	lockID := int64(hashString(taskName))

	var acquired bool
	err := w.db.Raw("SELECT pg_try_advisory_lock(?)", lockID).Scan(&acquired).Error
	if err != nil {
		w.Logger(taskName).Warnf("Failed to acquire lock: %v", err)
		return false
	}

	return acquired
}

// releaseLock releases the PostgreSQL advisory lock for the given task
func (w *Worker) releaseLock(taskName string) {
	if w.db == nil {
		return
	}

	lockID := int64(hashString(taskName))

	var released bool
	err := w.db.Raw("SELECT pg_advisory_unlock(?)", lockID).Scan(&released).Error
	if err != nil {
		w.Logger(taskName).Warnf("Failed to release lock: %v", err)
	}
}

// hashString generates a hash for the task name to use as lock ID
func hashString(s string) uint32 {
	h := uint32(0)
	for i := 0; i < len(s); i++ {
		h = 31*h + uint32(s[i])
	}
	return h
}

// Start starts the worker's cron scheduler
func (w *Worker) Start() {
	w.cron.Start()
}

// Stop gracefully stops the worker, waiting for background tasks to complete
func (w *Worker) Stop() {
	w.ctx.Cancel()
	w.cron.Stop()
	w.wg.Wait()
}

// RunInBackground runs a function in a background goroutine
// It tracks the goroutine and handles context cancellation
func (w *Worker) RunInBackground(name string, fn TaskFunc) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				w.Logger(name).Error("Background task panicked", r)
			}
		}()

		if err := fn(w.ctx); err != nil {
			w.Logger(name).Error("Background task error", err)
		}
	}()
}

func validateCronSchedule(schedule string) error {
	// Parse the schedule to validate minimum interval
	sched, err := cron.ParseStandard(schedule)
	if err != nil {
		return fmt.Errorf("invalid cron schedule '%s': %w", schedule, err)
	}

	// Check if the recurring interval is less than 1 minute
	// It's okay if the first run is less than a minute from now,
	// but the schedule must not repeat more frequently than once per minute
	now := time.Now()
	next1 := sched.Next(now)
	next2 := sched.Next(next1)
	if next2.Sub(next1) < time.Minute {
		return fmt.Errorf("schedule interval must be at least 1 minute, got %v", next2.Sub(next1))
	}

	return nil
}

// ScheduleCron schedules a function to run at specified cron intervals
// The schedule format follows cron syntax with seconds support:
// - "0 */5 * * * *" - every 5 minutes
// - "0 0 * * * *" - every hour
// - "0 0 0 * * *" - every day at midnight
// Returns an error if the schedule is invalid or the task name already exists
func (w *Worker) ScheduleCron(name string, schedule string, fn TaskFunc) error {
	if err := validateCronSchedule(schedule); err != nil {
		return err
	}

	w.taskMutex.Lock()
	defer w.taskMutex.Unlock()

	// Check if task already exists
	if _, exists := w.tasks[name]; exists {
		return fmt.Errorf("task '%s' already scheduled", name)
	}

	// Add the cron job with distributed locking
	entryID, err := w.cron.AddFunc(schedule, func() {
		if w.enableDL {
			// Try to acquire distributed lock
			if !w.acquireLock(name) {
				w.Logger(name).Debug("Task already running on another instance, skipping")
				return
			}
			defer w.releaseLock(name)
		}

		// Run the task
		if err := fn(w.ctx); err != nil {
			w.Logger(name).Error("Cron task error", err)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to schedule task '%s': %w", name, err)
	}

	w.tasks[name] = entryID
	return nil
}

// RemoveCronTask removes a scheduled cron task by name
func (w *Worker) RemoveCronTask(name string) error {
	w.taskMutex.Lock()
	defer w.taskMutex.Unlock()

	entryID, exists := w.tasks[name]
	if !exists {
		return fmt.Errorf("task '%s' not found", name)
	}

	w.cron.Remove(entryID)
	delete(w.tasks, name)
	return nil
}

// GetScheduledTasks returns the names of all scheduled cron tasks
func (w *Worker) GetScheduledTasks() []string {
	w.taskMutex.RLock()
	defer w.taskMutex.RUnlock()

	tasks := make([]string, 0, len(w.tasks))
	for name := range w.tasks {
		tasks = append(tasks, name)
	}
	return tasks
}

// RunWithTimeout runs a task with a specified timeout
func (w *Worker) RunWithTimeout(name string, timeout time.Duration, fn TaskFunc) error {
	ctx := w.ctx.WithTimeout(timeout)
	errChan := make(chan error, 1)

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("task '%s' panicked: %v", name, r)
			}
		}()
		errChan <- fn(ctx)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("task '%s' timed out after %v", name, timeout)
	}
}
