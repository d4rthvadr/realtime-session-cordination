package user

import "time"

const (
	TypeGuest  = "guest"
	TypeNormal = "normal"
)

type User struct {
	ID        string
	Name      *string
	Type      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	IsVisible bool
	AvatarURL *string
	Bio       *string
	IsActive  bool
}

type Snapshot struct {
	ID        string     `json:"id"`
	Name      *string    `json:"name,omitempty"`
	Type      string     `json:"type"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
	IsVisible bool       `json:"isVisible"`
	AvatarURL *string    `json:"avatarUrl,omitempty"`
	Bio       *string    `json:"bio,omitempty"`
	IsActive  bool       `json:"isActive"`
}

func ToSnapshot(u *User) Snapshot {
	return Snapshot{
		ID:        u.ID,
		Name:      u.Name,
		Type:      u.Type,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		DeletedAt: u.DeletedAt,
		IsVisible: u.IsVisible,
		AvatarURL: u.AvatarURL,
		Bio:       u.Bio,
		IsActive:  u.IsActive,
	}
}
