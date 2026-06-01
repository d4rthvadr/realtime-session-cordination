package sessionlog

import "fmt"

// EventType is the canonical taxonomy for append-only session log entries.
type EventType string

const (
	// Session events
	SessionCreated      EventType = "SESSION_CREATED"
	SessionStarted      EventType = "SESSION_STARTED"
	SessionPaused       EventType = "SESSION_PAUSED"
	SessionResumed      EventType = "SESSION_RESUMED"
	SessionEnded        EventType = "SESSION_ENDED"
	SessionTimeAdjusted EventType = "SESSION_TIME_ADJUSTED"

	// Program item events
	ProgramItemCreated      EventType = "PROGRAM_ITEM_CREATED"
	ProgramItemUpdated      EventType = "PROGRAM_ITEM_UPDATED"
	ProgramItemsReordered   EventType = "PROGRAM_ITEMS_REORDERED"
	ProgramItemCanceled     EventType = "PROGRAM_ITEM_CANCELED"
	ProgramItemStarted      EventType = "PROGRAM_ITEM_STARTED"
	ProgramItemPaused       EventType = "PROGRAM_ITEM_PAUSED"
	ProgramItemResumed      EventType = "PROGRAM_ITEM_RESUMED"
	ProgramItemEnded        EventType = "PROGRAM_ITEM_ENDED"
	ProgramItemTimeAdjusted EventType = "PROGRAM_ITEM_TIME_ADJUSTED"

	// Cascade events are automatically generated when a session-level action affects program items.
	CascadeProgramItemPausedBySession       EventType = "CASCADE_PROGRAM_ITEM_PAUSED_BY_SESSION"
	CascadeProgramItemResumedBySession      EventType = "CASCADE_PROGRAM_ITEM_RESUMED_BY_SESSION"
	CascadeProgramItemEndedBySession        EventType = "CASCADE_PROGRAM_ITEM_ENDED_BY_SESSION"
	CascadeProgramItemTimeAdjustedBySession EventType = "CASCADE_PROGRAM_ITEM_TIME_ADJUSTED_BY_SESSION"
)

// MessageInput contains optional values used to render human-readable trail lines.
type MessageInput struct {
	SessionTitle       string
	ProgramItemTitle   string
	DeltaSeconds       int
	ReorderedItemCount int
}

func (e EventType) String() string {
	return string(e)
}

// RenderMessage turns canonical events into stable, human-readable log text.
func RenderMessage(eventType EventType, in MessageInput) string {
	sessionLabel := in.SessionTitle
	if sessionLabel == "" {
		sessionLabel = "Session"
	}

	itemLabel := in.ProgramItemTitle
	if itemLabel == "" {
		itemLabel = "Program item"
	}

	switch eventType {
	case SessionCreated:
		return fmt.Sprintf("Created session \"%s\"", sessionLabel)
	case SessionStarted:
		return fmt.Sprintf("Started session \"%s\"", sessionLabel)
	case SessionPaused:
		return fmt.Sprintf("Paused session \"%s\"", sessionLabel)
	case SessionResumed:
		return fmt.Sprintf("Resumed session \"%s\"", sessionLabel)
	case SessionEnded:
		return fmt.Sprintf("Ended session \"%s\"", sessionLabel)
	case SessionTimeAdjusted:
		return fmt.Sprintf("Adjusted session time by %s", signedSeconds(in.DeltaSeconds))
	case ProgramItemCreated:
		return fmt.Sprintf("Added program item \"%s\"", itemLabel)
	case ProgramItemUpdated:
		return fmt.Sprintf("Updated program item \"%s\"", itemLabel)
	case ProgramItemsReordered:
		if in.ReorderedItemCount > 0 {
			return fmt.Sprintf("Reordered %d program items", in.ReorderedItemCount)
		}
		return "Reordered program items"
	case ProgramItemCanceled:
		return fmt.Sprintf("Canceled program item \"%s\"", itemLabel)
	case ProgramItemStarted:
		return fmt.Sprintf("Started program item \"%s\"", itemLabel)
	case ProgramItemPaused:
		return fmt.Sprintf("Paused program item \"%s\"", itemLabel)
	case ProgramItemResumed:
		return fmt.Sprintf("Resumed program item \"%s\"", itemLabel)
	case ProgramItemEnded:
		return fmt.Sprintf("Ended program item \"%s\"", itemLabel)
	case ProgramItemTimeAdjusted:
		return fmt.Sprintf("Adjusted program item \"%s\" by %s", itemLabel, signedSeconds(in.DeltaSeconds))
	case CascadeProgramItemPausedBySession:
		return fmt.Sprintf("Auto-paused program item \"%s\" because session was paused", itemLabel)
	case CascadeProgramItemResumedBySession:
		return fmt.Sprintf("Auto-resumed program item \"%s\" because session was resumed", itemLabel)
	case CascadeProgramItemEndedBySession:
		return fmt.Sprintf("Auto-ended program item \"%s\" because session was ended", itemLabel)
	case CascadeProgramItemTimeAdjustedBySession:
		return fmt.Sprintf("Auto-adjusted program item \"%s\" by %s from session adjustment", itemLabel, signedSeconds(in.DeltaSeconds))
	default:
		return fmt.Sprintf("Recorded event %s", eventType)
	}
}

func signedSeconds(delta int) string {
	if delta > 0 {
		return fmt.Sprintf("+%ds", delta)
	}
	return fmt.Sprintf("%ds", delta)
}
