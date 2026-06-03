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

func newTestHandler(t *testing.T) (*Handler, *session.Manager, *programitem.Manager, *sessionlog.Manager) {
  t.Helper()

  sessionMgr := session.NewManager(session.NewMemoryStore())
  programItemMgr := programitem.NewManager(programitem.NewMemoryStore(sessionMgr.SessionExists))
  sessionLogMgr := sessionlog.NewManager(sessionlog.NewMemoryStore())
  analyticsMgr := analytics.NewManager()
  hub := ws.NewHub(nil)

  return NewHandler(sessionMgr, programItemMgr, sessionLogMgr, analyticsMgr, hub, nil, nil), sessionMgr, programItemMgr, sessionLogMgr
}

func TestListSessionLogsReturnsLogsAndCount(t *testing.T) {
  gin.SetMode(gin.TestMode)

  handler, sessionMgr, _, sessionLogMgr := newTestHandler(t)

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

  handler, sessionMgr, _, _ := newTestHandler(t)

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

  handler, sessionMgr, programItemMgr, _ := newTestHandler(t)

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
