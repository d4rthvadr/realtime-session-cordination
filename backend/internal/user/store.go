package user

import "errors"

var (
	ErrNotFound = errors.New("user not found")
)

type Store interface {
	Create(user *User) (*User, error)
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
}
