package database

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// BatchConfig holds batch operation configuration
type BatchConfig struct {
	StepResultBatchSize    int           // Flush after N step results
	ProgressUpdateInterval time.Duration // Batch progress updates interval
	MaxPendingStepResults  int           // Max buffered before force flush
}

// DefaultBatchConfig returns sensible defaults
func DefaultBatchConfig() *BatchConfig {
	return &BatchConfig{
		StepResultBatchSize:    10,
		ProgressUpdateInterval: 5 * time.Second,
		MaxPendingStepResults:  50,
	}
}

// StepResultBuffer buffers step results for batch insertion
type StepResultBuffer struct {
	mu     sync.Mutex
	buffer []*StepResult
	runID  int64
	config *BatchConfig
}

// NewStepResultBuffer creates a new step result buffer
func NewStepResultBuffer(runID int64, cfg *BatchConfig) *StepResultBuffer {
	if cfg == nil {
		cfg = DefaultBatchConfig()
	}
	return &StepResultBuffer{
		buffer: make([]*StepResult, 0, cfg.StepResultBatchSize),
		runID:  runID,
		config: cfg,
	}
}

// Add adds a step result to the buffer and flushes if threshold reached
func (b *StepResultBuffer) Add(ctx context.Context, stepName, stepType, status, command, output, errorMsg string, exports map[string]interface{}, durationMs int64, startedAt, completedAt *time.Time) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	result := &StepResult{
		ID:           uuid.New().String(),
		RunID:        b.runID,
		StepName:     stepName,
		StepType:     stepType,
		Status:       status,
		Command:      command,
		Output:       output,
		ErrorMessage: errorMsg,
		Exports:      exports,
		DurationMs:   durationMs,
		StartedAt:    startedAt,
		CompletedAt:  completedAt,
		CreatedAt:    time.Now(),
	}

	b.buffer = append(b.buffer, result)

	// Flush if we've reached the batch size or max pending
	if len(b.buffer) >= b.config.StepResultBatchSize || len(b.buffer) >= b.config.MaxPendingStepResults {
		return b.flushLocked(ctx)
	}

	return nil
}

// Flush writes all buffered step results to the database
func (b *StepResultBuffer) Flush(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.flushLocked(ctx)
}

// flushLocked performs the actual flush (must be called with lock held)
// In distributed worker mode, sends to Redis instead of local DB.
func (b *StepResultBuffer) flushLocked(ctx context.Context) error {
	if len(b.buffer) == 0 {
		return nil
	}

	// In distributed worker mode, send to Redis instead of local DB
	if shouldUseDistributedHooks() {
		for _, step := range b.buffer {
			trySendStepResultToRedis(ctx, step)
		}
		b.buffer = b.buffer[:0]
		return nil
	}

	if db == nil {
		// Clear buffer even if no db connection to prevent memory growth
		b.buffer = b.buffer[:0]
		return nil
	}

	// Batch insert all buffered results
	_, err := db.NewInsert().Model(&b.buffer).Exec(ctx)
	if err != nil {
		return err
	}

	// Clear the buffer
	b.buffer = b.buffer[:0]
	return nil
}

// Len returns the number of buffered items
func (b *StepResultBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.buffer)
}

// ProgressTracker batches progress updates
type ProgressTracker struct {
	mu           sync.Mutex
	runID        string
	pendingSteps int
	config       *BatchConfig
	stopCh       chan struct{}
	wg           sync.WaitGroup
	stopped      bool
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(runID string, cfg *BatchConfig) *ProgressTracker {
	if cfg == nil {
		cfg = DefaultBatchConfig()
	}

	pt := &ProgressTracker{
		runID:  runID,
		config: cfg,
		stopCh: make(chan struct{}),
	}

	// Start background ticker for periodic updates
	pt.wg.Add(1)
	go pt.runTicker()

	return pt
}

// IncrementSteps increments the pending step count
func (pt *ProgressTracker) IncrementSteps(count int) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.pendingSteps += count
}

// runTicker periodically flushes progress updates
func (pt *ProgressTracker) runTicker() {
	defer pt.wg.Done()

	ticker := time.NewTicker(pt.config.ProgressUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pt.stopCh:
			// Final flush before stopping
			pt.flush()
			return
		case <-ticker.C:
			pt.flush()
		}
	}
}

// flush writes pending progress to the database
func (pt *ProgressTracker) flush() {
	pt.mu.Lock()
	steps := pt.pendingSteps
	pt.pendingSteps = 0
	pt.mu.Unlock()

	if steps > 0 {
		_ = BatchUpdateRunProgress(context.Background(), pt.runID, steps)
	}
}

// Stop stops the progress tracker and flushes remaining updates
func (pt *ProgressTracker) Stop() {
	pt.mu.Lock()
	if pt.stopped {
		pt.mu.Unlock()
		return
	}
	pt.stopped = true
	pt.mu.Unlock()

	close(pt.stopCh)
	pt.wg.Wait()
}

// BatchUpdateRunProgress performs a single update for multiple steps
func BatchUpdateRunProgress(ctx context.Context, runUUID string, steps int) error {
	if db == nil || runUUID == "" || steps == 0 {
		return nil
	}

	_, err := db.NewUpdate().
		Model((*Run)(nil)).
		Set("completed_steps = completed_steps + ?", steps).
		Set("updated_at = ?", time.Now()).
		Where("run_uuid = ?", runUUID).
		Exec(ctx)
	return err
}
