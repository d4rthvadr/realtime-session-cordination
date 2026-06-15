package analytics

import (
	"context"
	"fmt"
	"log/slog"
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

	ProjectionBuilder        *ProjectionBuilder
	GetSessionSnapshot       func(sessionID string) (session.Snapshot, error)
	ListProgramItemSnapshots func(sessionID string) ([]programitem.Snapshot, error)
	ListSessionSnapshots     func() ([]session.Snapshot, error)
}

// Processor consumes outbox records and advances processing checkpoints.
type Processor struct {
	store  ProcessorStore
	logger *slog.Logger
	cfg    ProcessorConfig
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

	if logger == nil {
		logger = slog.Default()
	}

	return &Processor{store: store, logger: logger, cfg: cfg}
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

		select {
		case <-ctx.Done():
			p.logger.Info("analytics_processor_stopped", "worker", p.cfg.WorkerName)
			return
		case <-ticker.C:
		}
	}
}

func (p *Processor) processBatch(ctx context.Context) error {
	now := time.Now().UTC()
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

		event, eventErr := p.store.GetEvent(row.EventID)
		if eventErr != nil {
			markErr := p.store.MarkFailed(row.ID, eventErr.Error(), row.Attempt >= p.cfg.MaxAttempts, time.Now().UTC())
			if markErr != nil {
				p.logger.Error("analytics_processor_mark_failed_error", "outbox_id", row.ID, "error", markErr)
			}
			continue
		}

		if projectionErr := p.rebuildProjections(event, time.Now().UTC()); projectionErr != nil {
			markErr := p.store.MarkFailed(row.ID, fmt.Sprintf("projection rebuild failed: %v", projectionErr), row.Attempt >= p.cfg.MaxAttempts, time.Now().UTC())
			if markErr != nil {
				p.logger.Error("analytics_processor_projection_mark_failed_error", "outbox_id", row.ID, "error", markErr)
			}
			continue
		}

		// Phase 2C processing contract: consume outbox, record idempotent checkpoint.
		if saveErr := p.store.SaveCheckpoint(ProcessorCheckpoint{
			WorkerName:  p.cfg.WorkerName,
			LastEventID: event.ID,
			UpdatedAt:   time.Now().UTC(),
		}); saveErr != nil {
			markErr := p.store.MarkFailed(row.ID, fmt.Sprintf("checkpoint save failed: %v", saveErr), row.Attempt >= p.cfg.MaxAttempts, time.Now().UTC())
			if markErr != nil {
				p.logger.Error("analytics_processor_checkpoint_mark_failed_error", "outbox_id", row.ID, "error", markErr)
			}
			continue
		}

		if markErr := p.store.MarkProcessed(row.ID, time.Now().UTC()); markErr != nil {
			p.logger.Error("analytics_processor_mark_processed_error", "outbox_id", row.ID, "error", markErr)
			continue
		}
	}

	return nil
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
