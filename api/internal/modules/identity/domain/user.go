package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidEmail = errors.New("email is required")
	ErrInvalidRole  = errors.New("role must be user, merchant, or admin")
	ErrUserNotFound = errors.New("user not found")
)

type Role string

const (
	RoleUser     Role = "user"
	RoleMerchant Role = "merchant"
	RoleAdmin    Role = "admin"
)

func ParseRole(raw string) (Role, error) {
	switch Role(strings.ToLower(strings.TrimSpace(raw))) {
	case RoleUser, RoleMerchant, RoleAdmin:
		return Role(strings.ToLower(strings.TrimSpace(raw))), nil
	default:
		return "", ErrInvalidRole
	}
}

type UserID string

func NewUserID() UserID {
	return UserID(uuid.NewString())
}

type User struct {
	ID        UserID
	Email     string
	DisplayName string
	Role      Role
	CreatedAt time.Time
}

func NewUser(email, displayName string, role Role) (User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return User{}, ErrInvalidEmail
	}
	return User{
		ID:          NewUserID(),
		Email:       email,
		DisplayName: strings.TrimSpace(displayName),
		Role:        role,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

type UserRepository interface {
	Save(user User) error
	FindByEmail(email string) (User, error)
	ListByRole(role Role) ([]User, error)
}
