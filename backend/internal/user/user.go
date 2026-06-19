package user

import "time"

const (
	TypeGuest  = "guest"
	TypeUser   = "user"
	RoleUser   = "user"
	RoleAdmin  = "admin"
)

type User struct {
	ID              string
	Name            *string
	Email           *string
	Type            string
	Role            string
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
	IsVisible       bool
	AvatarURL       *string
	Bio             *string
	IsActive        bool
}

type Snapshot struct {
	ID              string     `json:"id"`
	Name            *string    `json:"name,omitempty"`
	Email           *string    `json:"email,omitempty"`
	Type            string     `json:"type"`
	Role            string     `json:"role"`
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	DeletedAt       *time.Time `json:"deletedAt,omitempty"`
	IsVisible       bool       `json:"isVisible"`
	AvatarURL       *string    `json:"avatarUrl,omitempty"`
	Bio             *string    `json:"bio,omitempty"`
	IsActive        bool       `json:"isActive"`
}

func ToSnapshot(u *User) Snapshot {
	return Snapshot{
		ID:              u.ID,
		Name:            u.Name,
		Email:           u.Email,
		Type:            u.Type,
		Role:            u.Role,
		EmailVerifiedAt: u.EmailVerifiedAt,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
		DeletedAt:       u.DeletedAt,
		IsVisible:       u.IsVisible,
		AvatarURL:       u.AvatarURL,
		Bio:             u.Bio,
		IsActive:        u.IsActive,
	}
}
