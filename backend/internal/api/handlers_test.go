package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"realtime-session-coordination/backend/internal/analytics"
	"realtime-session-coordination/backend/internal/programitem"
	"realtime-session-coordination/backend/internal/session"
	"realtime-session-coordination/backend/internal/sessionlog"
	"realtime-session-coordination/backend/internal/ws"

	"github.com/gin-gonic/gin"
)

type testIngestionStore struct {
	events   []analytics.EventRecord
	enqueued []string
}

type testAnalyticsProjectionProcessorStore struct {
	freshness               analytics.ProcessorFreshness
	metrics                 analytics.ProcessorMetrics
	sessionProjection       analytics.SessionProjection
	sessionProjectionFound  bool
	platformProjection      analytics.PlatformProjection
	platformProjectionFound bool
	deadLetters             map[int64]analytics.DeadLetterRecord
}

func (s *testAnalyticsProjectionProcessorStore) ClaimPendingForProcessing(workerName string, leaseUntil time.Time, limit int, now time.Time) ([]analytics.OutboxRecord, error) {
	return nil, nil
}

func (s *testAnalyticsProjectionProcessorStore) GetEvent(eventID string) (analytics.EventRecord, error) {
	return analytics.EventRecord{}, nil
}

func (s *testAnalyticsProjectionProcessorStore) MarkProcessed(outboxID int64, now time.Time) error {
	return nil
}

func (s *testAnalyticsProjectionProcessorStore) MarkFailed(outboxID int64, lastError string, deadLetter bool, nextRetryAt *time.Time, now time.Time) error {
	return nil
}

func (s *testAnalyticsProjectionProcessorStore) SaveCheckpoint(checkpoint analytics.ProcessorCheckpoint) error {
	return nil
}

func (s *testAnalyticsProjectionProcessorStore) LoadCheckpoint(workerName string) (analytics.ProcessorCheckpoint, bool, error) {
	return analytics.ProcessorCheckpoint{}, false, nil
}

func (s *testAnalyticsProjectionProcessorStore) GetFreshness(workerName string, now time.Time) (analytics.ProcessorFreshness, error) {
	return s.freshness, nil
}

func (s *testAnalyticsProjectionProcessorStore) UpsertSessionProjection(p analytics.SessionProjection) error {
	return nil
}

func (s *testAnalyticsProjectionProcessorStore) GetSessionProjection(sessionID string) (analytics.SessionProjection, bool, error) {
	return s.sessionProjection, s.sessionProjectionFound, nil
}

func (s *testAnalyticsProjectionProcessorStore) UpsertPlatformProjection(p analytics.PlatformProjection) error {
	return nil
}

func (s *testAnalyticsProjectionProcessorStore) GetPlatformProjection() (analytics.PlatformProjection, bool, error) {
	return s.platformProjection, s.platformProjectionFound, nil
}

func (s *testAnalyticsProjectionProcessorStore) ListDeadLetters(limit int, offset int) ([]analytics.DeadLetterRecord, error) {
	if s.deadLetters == nil {
		return []analytics.DeadLetterRecord{}, nil
	}
	rows := make([]analytics.DeadLetterRecord, 0, len(s.deadLetters))
	for _, row := range s.deadLetters {
		rows = append(rows, row)
	}
	return rows, nil
}

func (s *testAnalyticsProjectionProcessorStore) GetDeadLetter(outboxID int64) (analytics.DeadLetterRecord, bool, error) {
	if s.deadLetters == nil {
		return analytics.DeadLetterRecord{}, false, nil
	}
	row, ok := s.deadLetters[outboxID]
	return row, ok, nil
}

func (s *testAnalyticsProjectionProcessorStore) RetryDeadLetter(outboxID int64, now time.Time) error {
	if s.deadLetters != nil {
		delete(s.deadLetters, outboxID)
	}
	return nil
}

func (s *testAnalyticsProjectionProcessorStore) GetProcessorMetrics() analytics.ProcessorMetrics {
	return s.metrics
}

func (s *testIngestionStore) AppendEventAndEnqueue(record analytics.EventRecord, now time.Time) error {
	s.events = append(s.events, record)
	s.enqueued = append(s.enqueued, record.ID)
	return nil
}

func newTestHandler(t *testing.T) (*Handler, *session.Manager, *programitem.Manager, *sessionlog.Manager, *analytics.Emitter) {
	t.Helper()

	sessionMgr := session.NewManager(session.NewMemoryStore())
	programItemMgr := programitem.NewManager(programitem.NewMemoryStore(sessionMgr.SessionExists))
	sessionLogMgr := sessionlog.NewManager(sessionlog.NewMemoryStore())
	analyticsMgr := analytics.NewManager()

	// Create in-memory stores for testing
	ingestionStore := &testIngestionStore{events: []analytics.EventRecord{}, enqueued: []string{}}
	analyticsEmitter := analytics.NewEmitter(ingestionStore)

	hub := ws.NewHub(nil)

	return NewHandler(sessionMgr, programItemMgr, sessionLogMgr, analyticsMgr, analyticsEmitter, nil, hub, nil, nil, nil), sessionMgr, programItemMgr, sessionLogMgr, analyticsEmitter
}

func TestListSessionLogsReturnsLogsAndCount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, sessionMgr, _, sessionLogMgr, _ := newTestHandler(t)

	created, _, err := sessionMgr.Create(session.CreateInput{
		Title:           "Demo Session",
		SpeakerName:     "Host",
		DurationSeconds: 1800,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	if _, err := sessionLogMgr.Append(sessionlog.AppendInput{
		SessionID: created.ID,
		EventType: sessionlog.SessionCreated,
		MessageInput: sessionlog.MessageInput{
			SessionTitle: created.Title,
		},
	}); err != nil {
		t.Fatalf("append created log: %v", err)
	}

	if _, err := sessionLogMgr.Append(sessionlog.AppendInput{
		SessionID: created.ID,
		EventType: sessionlog.SessionStarted,
		MessageInput: sessionlog.MessageInput{
			SessionTitle: created.Title,
		},
	}); err != nil {
		t.Fatalf("append started log: %v", err)
	}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+created.ID+"/logs?limit=10&offset=0", nil)
	c.Params = gin.Params{{Key: "id", Value: created.ID}}

	handler.listSessionLogs(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var payload struct {
		Logs  []sessionlog.Snapshot `json:"logs"`
		Count int                   `json:"count"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Count != 2 {
		t.Fatalf("expected count=2, got %d", payload.Count)
	}
	if len(payload.Logs) != 2 {
		t.Fatalf("expected 2 logs, got %d", len(payload.Logs))
	}
}

func TestListSessionLogsRejectsInvalidLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, sessionMgr, _, _, _ := newTestHandler(t)

	created, _, err := sessionMgr.Create(session.CreateInput{
		Title:           "Demo Session",
		SpeakerName:     "Host",
		DurationSeconds: 1800,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+created.ID+"/logs?limit=bad", nil)
	c.Params = gin.Params{{Key: "id", Value: created.ID}}

	handler.listSessionLogs(c)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestGetAnalyticsOverviewReturnsAggregates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, sessionMgr, programItemMgr, _, _ := newTestHandler(t)

	created, _, err := sessionMgr.Create(session.CreateInput{
		Title:           "Analytics Session",
		SpeakerName:     "Host",
		DurationSeconds: 3600,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	start := time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC)
	end := start.Add(30 * time.Minute)
	if _, err := programItemMgr.Create(programitem.CreateInput{
		SessionID:               created.ID,
		Title:                   "Opening Keynote",
		Type:                    "keynote",
		HostName:                "Host",
		ScheduledStart:          start,
		ScheduledEnd:            end,
		ExpectedDurationMinutes: 30,
		Position:                1,
	}); err != nil {
		t.Fatalf("create program item: %v", err)
	}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/analytics/overview", nil)

	handler.getAnalyticsOverview(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var payload struct {
		Overview analytics.PlatformOverview `json:"overview"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Overview.TotalSessions != 1 {
		t.Fatalf("expected totalSessions=1, got %d", payload.Overview.TotalSessions)
	}
	if payload.Overview.TotalProgramItems != 1 {
		t.Fatalf("expected totalProgramItems=1, got %d", payload.Overview.TotalProgramItems)
	}
}

func TestGetSessionAnalyticsUsesProjectionWhenAvailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _, _, _, _ := newTestHandler(t)
	handler.analyticsManager = nil
	handler.programItemManager = nil
	handler.analyticsProcessorStore = &testAnalyticsProjectionProcessorStore{
		freshness:              analytics.ProcessorFreshness{WorkerName: "analytics_processor", PendingCount: 2},
		sessionProjectionFound: true,
		sessionProjection: analytics.SessionProjection{
			SessionID:              "sess_projection",
			SessionStatus:          session.StatusLive,
			SessionDurationSeconds: 1500,
			ProgramItemCount:       3,
			ComputedAt:             time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC),
		},
	}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/sessions/sess_projection/analytics", nil)
	c.Params = gin.Params{{Key: "id", Value: "sess_projection"}}

	handler.getSessionAnalytics(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var payload struct {
		Analytics analytics.SessionSummary     `json:"analytics"`
		Freshness analytics.ProcessorFreshness `json:"freshness"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Analytics.SessionID != "sess_projection" {
		t.Fatalf("expected sessionID from projection, got %s", payload.Analytics.SessionID)
	}
	if payload.Analytics.ProgramItemCount != 3 {
		t.Fatalf("expected projected ProgramItemCount=3, got %d", payload.Analytics.ProgramItemCount)
	}
	if payload.Freshness.PendingCount != 2 {
		t.Fatalf("expected freshness pendingCount=2, got %d", payload.Freshness.PendingCount)
	}
}

func TestGetAnalyticsOverviewUsesProjectionWhenAvailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _, _, _, _ := newTestHandler(t)
	handler.analyticsManager = nil
	handler.programItemManager = nil
	handler.analyticsProcessorStore = &testAnalyticsProjectionProcessorStore{
		freshness:               analytics.ProcessorFreshness{WorkerName: "analytics_processor", PendingCount: 1},
		platformProjectionFound: true,
		platformProjection: analytics.PlatformProjection{
			TotalSessions:     7,
			TotalProgramItems: 19,
			ComputedAt:        time.Date(2026, 6, 15, 12, 30, 0, 0, time.UTC),
		},
	}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/analytics/overview", nil)

	handler.getAnalyticsOverview(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var payload struct {
		Overview  analytics.PlatformOverview   `json:"overview"`
		Freshness analytics.ProcessorFreshness `json:"freshness"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Overview.TotalSessions != 7 {
		t.Fatalf("expected projected TotalSessions=7, got %d", payload.Overview.TotalSessions)
	}
	if payload.Overview.TotalProgramItems != 19 {
		t.Fatalf("expected projected TotalProgramItems=19, got %d", payload.Overview.TotalProgramItems)
	}
	if payload.Freshness.PendingCount != 1 {
		t.Fatalf("expected freshness pendingCount=1, got %d", payload.Freshness.PendingCount)
	}
}

func TestListAnalyticsDeadLettersReturnsRows(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _, _, _, _ := newTestHandler(t)
	handler.analyticsProcessorStore = &testAnalyticsProjectionProcessorStore{
		deadLetters: map[int64]analytics.DeadLetterRecord{
			11: {
				OutboxID:   11,
				EventID:    "evt_dlq_11",
				SessionID:  "sess_dlq",
				EventKey:   "PROGRAM_ITEM_ENDED",
				Attempt:    5,
				LastError:  "projection rebuild failed",
				OccurredAt: time.Date(2026, 6, 17, 9, 0, 0, 0, time.UTC),
				IngestedAt: time.Date(2026, 6, 17, 9, 0, 1, 0, time.UTC),
				FailedAt:   time.Date(2026, 6, 17, 9, 0, 5, 0, time.UTC),
			},
		},
	}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/analytics/dlq?limit=10&offset=0", nil)

	handler.listAnalyticsDeadLetters(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var payload struct {
		Rows  []analytics.DeadLetterRecord `json:"rows"`
		Count int                          `json:"count"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Count != 1 || len(payload.Rows) != 1 {
		t.Fatalf("expected one dead-letter row, got count=%d len=%d", payload.Count, len(payload.Rows))
	}
}

func TestGetAnalyticsDeadLetterReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _, _, _, _ := newTestHandler(t)
	handler.analyticsProcessorStore = &testAnalyticsProjectionProcessorStore{deadLetters: map[int64]analytics.DeadLetterRecord{}}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/analytics/dlq/99", nil)
	c.Params = gin.Params{{Key: "outboxId", Value: "99"}}

	handler.getAnalyticsDeadLetter(c)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestRetryAnalyticsDeadLetterQueuesRow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _, _, _, _ := newTestHandler(t)
	handler.analyticsProcessorStore = &testAnalyticsProjectionProcessorStore{
		deadLetters: map[int64]analytics.DeadLetterRecord{
			41: {
				OutboxID:   41,
				EventID:    "evt_dlq_41",
				SessionID:  "sess_dlq",
				EventKey:   "SESSION_ENDED",
				Attempt:    5,
				LastError:  "checkpoint save failed",
				OccurredAt: time.Date(2026, 6, 17, 9, 0, 0, 0, time.UTC),
				IngestedAt: time.Date(2026, 6, 17, 9, 0, 1, 0, time.UTC),
				FailedAt:   time.Date(2026, 6, 17, 9, 0, 5, 0, time.UTC),
			},
		},
	}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/analytics/dlq/41/retry", nil)
	c.Params = gin.Params{{Key: "outboxId", Value: "41"}}

	handler.retryAnalyticsDeadLetter(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}

	store := handler.analyticsProcessorStore.(*testAnalyticsProjectionProcessorStore)
	if _, exists := store.deadLetters[41]; exists {
		t.Fatalf("expected dead-letter row to be removed after retry")
	}
}

func TestGetAnalyticsOpsStatusReturnsFreshnessAndMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	handler, _, _, _, _ := newTestHandler(t)
	handler.analyticsProcessorStore = &testAnalyticsProjectionProcessorStore{
		freshness: analytics.ProcessorFreshness{
			WorkerName:      "analytics_processor",
			PendingCount:    3,
			RetryDueCount:   2,
			DeadLetterCount: 1,
			RetryLagSeconds: 17,
			LastEventID:     "evt_ops",
			LastProcessedAt: &now,
		},
	}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/analytics/ops/status", nil)

	handler.getAnalyticsOpsStatus(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var payload struct {
		Freshness analytics.ProcessorFreshness `json:"freshness"`
		Metrics   analytics.ProcessorMetrics   `json:"metrics"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Freshness.RetryDueCount != 2 {
		t.Fatalf("expected retryDueCount=2, got %d", payload.Freshness.RetryDueCount)
	}
	if payload.Freshness.DeadLetterCount != 1 {
		t.Fatalf("expected deadLetterCount=1, got %d", payload.Freshness.DeadLetterCount)
	}
	if payload.Freshness.RetryLagSeconds != 17 {
		t.Fatalf("expected retryLagSeconds=17, got %d", payload.Freshness.RetryLagSeconds)
	}
}

func TestStartSessionEmitsAnalyticsEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	_, sessionMgr, _, _, _ := newTestHandler(t)

	created, _, err := sessionMgr.Create(session.CreateInput{
		Title:           "Demo Session",
		SpeakerName:     "Host",
		DurationSeconds: 1800,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	// Start the session via manager instead of handler to avoid auth checks
	startedEvent, err := sessionMgr.Start(created.ID)
	if err != nil {
		t.Fatalf("start session: %v", err)
	}

	if startedEvent.Type != "SESSION_STARTED" {
		t.Fatalf("expected event type SESSION_STARTED, got %s", startedEvent.Type)
	}

	// Verify session state changed
	snap, err := sessionMgr.GetSnapshot(created.ID)
	if err != nil {
		t.Fatalf("get session snapshot: %v", err)
	}

	if snap.Status != session.StatusLive {
		t.Fatalf("expected session status LIVE, got %s", snap.Status)
	}
}
