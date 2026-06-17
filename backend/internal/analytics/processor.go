package analytics

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"realtime-session-coordination/backend/internal/programitem"
	"realtime-session-coordination/backend/internal/session"
)

// ProcessorConfig defines runtime behavior for the analytics outbox processor.
type ProcessorConfig struct {
	WorkerName    string
	PollInterval  time.Duration
	LeaseDuration time.Duration
	BatchSize     int
	MaxAttempts   int
	RetryPolicy   RetryPolicy
	CleanupInterval          time.Duration
	ProcessedOutboxRetention time.Duration
	DeadLetterRetention      time.Duration
	EventRetention           time.Duration

	ProjectionBuilder        *ProjectionBuilder
	GetSessionSnapshot       func(sessionID string) (session.Snapshot, error)
	ListProgramItemSnapshots func(sessionID string) ([]programitem.Snapshot, error)
	ListSessionSnapshots     func() ([]session.Snapshot, error)
}

// Processor consumes outbox records and advances processing checkpoints.
type Processor struct {
	store         ProcessorStore
	logger        *slog.Logger
	cfg           ProcessorConfig
	lastCleanupAt time.Time
	nowFn         func() time.Time

	processedCount       int64
	failedCount          int64
	deadLetterCount      int64
	projectionErrorCount int64
	checkpointErrorCount int64
	lastBatchAtUnix      int64
	lastBatchDurationMs  int64
}

func NewProcessor(store ProcessorStore, logger *slog.Logger, cfg ProcessorConfig) *Processor {
	if cfg.WorkerName == "" {
		cfg.WorkerName = "analytics_processor"
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 2 * time.Second
	}
	if cfg.LeaseDuration <= 0 {
		cfg.LeaseDuration = 15 * time.Second
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 25
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 5
	}
	if cfg.RetryPolicy.BaseDelay <= 0 {
		cfg.RetryPolicy = DefaultRetryPolicy()
	}
	if cfg.RetryPolicy.MaxDelay <= 0 {
		cfg.RetryPolicy.MaxDelay = DefaultRetryPolicy().MaxDelay
	}
	if cfg.CleanupInterval <= 0 {
		cfg.CleanupInterval = 10 * time.Minute
	}
	if cfg.ProcessedOutboxRetention <= 0 {
		cfg.ProcessedOutboxRetention = 24 * time.Hour
	}
	if cfg.DeadLetterRetention <= 0 {
		cfg.DeadLetterRetention = 7 * 24 * time.Hour
	}
	if cfg.EventRetention <= 0 {
		cfg.EventRetention = 14 * 24 * time.Hour
	}

	if logger == nil {
		logger = slog.Default()
	}

	return &Processor{
		store:  store,
		logger: logger,
		cfg:    cfg,
		nowFn: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (p *Processor) Start(ctx context.Context) {
	if p == nil || p.store == nil {
		return
	}

	p.logger.Info("analytics_processor_started", "worker", p.cfg.WorkerName)
	ticker := time.NewTicker(p.cfg.PollInterval)
	defer ticker.Stop()

	for {
		if err := p.processBatch(ctx); err != nil {
			p.logger.Error("analytics_processor_batch_failed", "worker", p.cfg.WorkerName, "error", err)
		}
		if err := p.runScheduledCleanup(); err != nil {
			p.logger.Error("analytics_processor_cleanup_failed", "worker", p.cfg.WorkerName, "error", err)
		}

		select {
		case <-ctx.Done():
			p.logger.Info("analytics_processor_stopped", "worker", p.cfg.WorkerName)
			return
		case <-ticker.C:
		}
	}
}

func (p *Processor) processBatch(ctx context.Context) error {
	batchStart := p.nowUTC()
	defer func() {
		now := p.nowUTC()
		atomic.StoreInt64(&p.lastBatchAtUnix, now.Unix())
		atomic.StoreInt64(&p.lastBatchDurationMs, now.Sub(batchStart).Milliseconds())
	}()

	now := p.nowUTC()
	leaseUntil := now.Add(p.cfg.LeaseDuration)

	rows, err := p.store.ClaimPendingForProcessing(p.cfg.WorkerName, leaseUntil, p.cfg.BatchSize, now)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	for _, row := range rows {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		batchNow := p.nowUTC()

		event, eventErr := p.store.GetEvent(row.EventID)
		if eventErr != nil {
			atomic.AddInt64(&p.failedCount, 1)
			if row.Attempt >= p.cfg.MaxAttempts {
				atomic.AddInt64(&p.deadLetterCount, 1)
			}
			markErr := p.store.MarkFailed(row.ID, eventErr.Error(), row.Attempt >= p.cfg.MaxAttempts, retryAtForAttempt(row.Attempt, batchNow, p.cfg.RetryPolicy, row.Attempt >= p.cfg.MaxAttempts), batchNow)
			if markErr != nil {
				p.logger.Error("analytics_processor_mark_failed_error", "outbox_id", row.ID, "error", markErr)
			}
			continue
		}

		if projectionErr := p.rebuildProjections(event, batchNow); projectionErr != nil {
			atomic.AddInt64(&p.failedCount, 1)
			atomic.AddInt64(&p.projectionErrorCount, 1)
			if row.Attempt >= p.cfg.MaxAttempts {
				atomic.AddInt64(&p.deadLetterCount, 1)
			}
			markErr := p.store.MarkFailed(row.ID, fmt.Sprintf("projection rebuild failed: %v", projectionErr), row.Attempt >= p.cfg.MaxAttempts, retryAtForAttempt(row.Attempt, batchNow, p.cfg.RetryPolicy, row.Attempt >= p.cfg.MaxAttempts), batchNow)
			if markErr != nil {
				p.logger.Error("analytics_processor_projection_mark_failed_error", "outbox_id", row.ID, "error", markErr)
			}
			continue
		}

		// Phase 2C processing contract: consume outbox, record idempotent checkpoint.
		if saveErr := p.store.SaveCheckpoint(ProcessorCheckpoint{
			WorkerName:  p.cfg.WorkerName,
			LastEventID: event.ID,
			UpdatedAt:   batchNow,
		}); saveErr != nil {
			atomic.AddInt64(&p.failedCount, 1)
			atomic.AddInt64(&p.checkpointErrorCount, 1)
			if row.Attempt >= p.cfg.MaxAttempts {
				atomic.AddInt64(&p.deadLetterCount, 1)
			}
			markErr := p.store.MarkFailed(row.ID, fmt.Sprintf("checkpoint save failed: %v", saveErr), row.Attempt >= p.cfg.MaxAttempts, retryAtForAttempt(row.Attempt, batchNow, p.cfg.RetryPolicy, row.Attempt >= p.cfg.MaxAttempts), batchNow)
			if markErr != nil {
				p.logger.Error("analytics_processor_checkpoint_mark_failed_error", "outbox_id", row.ID, "error", markErr)
			}
			continue
		}

		if markErr := p.store.MarkProcessed(row.ID, p.nowUTC()); markErr != nil {
			p.logger.Error("analytics_processor_mark_processed_error", "outbox_id", row.ID, "error", markErr)
			continue
		}
		atomic.AddInt64(&p.processedCount, 1)
	}

	return nil
}

func (p *Processor) GetProcessorMetrics() ProcessorMetrics {
	metrics := ProcessorMetrics{
		ProcessedCount:          atomic.LoadInt64(&p.processedCount),
		FailedCount:             atomic.LoadInt64(&p.failedCount),
		DeadLetterCount:         atomic.LoadInt64(&p.deadLetterCount),
		ProjectionErrorCount:    atomic.LoadInt64(&p.projectionErrorCount),
		CheckpointErrorCount:    atomic.LoadInt64(&p.checkpointErrorCount),
		LastBatchDurationMillis: atomic.LoadInt64(&p.lastBatchDurationMs),
	}
	lastAt := atomic.LoadInt64(&p.lastBatchAtUnix)
	if lastAt > 0 {
		t := time.Unix(lastAt, 0).UTC()
		metrics.LastBatchAt = &t
	}
	return metrics
}

func (p *Processor) nowUTC() time.Time {
	if p.nowFn != nil {
		return p.nowFn().UTC()
	}
	return time.Now().UTC()
}

func (p *Processor) runScheduledCleanup() error {
	cleanupStore, ok := p.store.(CleanupStore)
	if !ok || cleanupStore == nil {
		return nil
	}

	now := p.nowUTC()
	if !p.lastCleanupAt.IsZero() && now.Before(p.lastCleanupAt.Add(p.cfg.CleanupInterval)) {
		return nil
	}

	processedCutoff := now.Add(-p.cfg.ProcessedOutboxRetention)
	deadLetterCutoff := now.Add(-p.cfg.DeadLetterRetention)
	eventsCutoff := now.Add(-p.cfg.EventRetention)

	processedDeleted, err := cleanupStore.CleanupProcessedOutbox(processedCutoff)
	if err != nil {
		return err
	}
	deadLettersDeleted, err := cleanupStore.CleanupDeadLetters(deadLetterCutoff)
	if err != nil {
		return err
	}
	eventsDeleted, err := cleanupStore.CleanupEvents(eventsCutoff)
	if err != nil {
		return err
	}

	p.lastCleanupAt = now
	p.logger.Info(
		"analytics_processor_cleanup_completed",
		"worker", p.cfg.WorkerName,
		"processed_outbox_deleted", processedDeleted,
		"dead_letters_deleted", deadLettersDeleted,
		"events_deleted", eventsDeleted,
	)

	return nil
}

func retryAtForAttempt(attempt int, now time.Time, policy RetryPolicy, deadLetter bool) *time.Time {
	if deadLetter {
		return nil
	}
	retryAt := NextRetryAt(attempt, now, policy)
	return &retryAt
}

func (p *Processor) rebuildProjections(event EventRecord, now time.Time) error {
	if p.cfg.ProjectionBuilder == nil {
		return nil
	}
	if p.cfg.GetSessionSnapshot == nil {
		return fmt.Errorf("session snapshot provider is required")
	}
	if p.cfg.ListProgramItemSnapshots == nil {
		return fmt.Errorf("program item snapshot provider is required")
	}
	if p.cfg.ListSessionSnapshots == nil {
		return fmt.Errorf("session list provider is required")
	}

	sessionSnap, err := p.cfg.GetSessionSnapshot(event.SessionID)
	if err != nil {
		return fmt.Errorf("load session snapshot: %w", err)
	}
	items, err := p.cfg.ListProgramItemSnapshots(event.SessionID)
	if err != nil {
		return fmt.Errorf("load session program items: %w", err)
	}
	if err = p.cfg.ProjectionBuilder.RebuildSessionProjection(event.SessionID, sessionSnap, items, now); err != nil {
		return fmt.Errorf("rebuild session projection: %w", err)
	}

	sessions, err := p.cfg.ListSessionSnapshots()
	if err != nil {
		return fmt.Errorf("load sessions: %w", err)
	}
	itemsBySession := make(map[string][]programitem.Snapshot, len(sessions))
	for _, s := range sessions {
		if s.ID == event.SessionID {
			itemsBySession[s.ID] = items
			continue
		}

		sessionItems, listErr := p.cfg.ListProgramItemSnapshots(s.ID)
		if listErr != nil {
			return fmt.Errorf("load session program items for %s: %w", s.ID, listErr)
		}
		itemsBySession[s.ID] = sessionItems
	}

	if err = p.cfg.ProjectionBuilder.RebuildPlatformProjection(sessions, itemsBySession, now); err != nil {
		return fmt.Errorf("rebuild platform projection: %w", err)
	}

	return nil
}
