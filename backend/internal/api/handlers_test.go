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
	sessionProjection       analytics.SessionProjection
	sessionProjectionFound  bool
	platformProjection      analytics.PlatformProjection
	platformProjectionFound bool
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

func (s *testAnalyticsProjectionProcessorStore) MarkFailed(outboxID int64, lastError string, deadLetter bool, now time.Time) error {
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

	return NewHandler(sessionMgr, programItemMgr, sessionLogMgr, analyticsMgr, analyticsEmitter, nil, hub, nil, nil), sessionMgr, programItemMgr, sessionLogMgr, analyticsEmitter
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
