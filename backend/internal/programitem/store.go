package programitem

import "time"

// Store defines persistence operations for program items.
type Store interface {
	Create(item *ProgramItem) (*ProgramItem, error)
	Get(id string) (*ProgramItem, error)
	ListBySession(sessionID string) ([]*ProgramItem, error)
	Update(item *ProgramItem) error
	Reorder(sessionID string, positions map[string]int) error
	HasOverlap(sessionID string, start, end time.Time, excludeID string) (bool, error)
	PositionExists(sessionID string, position int, excludeID string) (bool, error)
	HasInProgressItem(sessionID string, excludeID string) (bool, error)
	SessionExists(sessionID string) bool
}
