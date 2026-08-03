package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidEmail         = errors.New("email is required")
	ErrInvalidRole          = errors.New("role must be user, merchant, or admin")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrForbiddenRole        = errors.New("user role is not allowed")
	ErrPasswordNotSet       = errors.New("password is not set for this account")
	ErrWeakPassword         = errors.New("password must be at least 8 characters")
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

type PasswordHasher interface {
	Hash(plain string) (string, error)
	Compare(hash, plain string) error
}

type User struct {
	ID           UserID
	Email        string
	DisplayName  string
	Role         Role
	PasswordHash string
	CreatedAt    time.Time
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

func (u *User) SetPassword(hasher PasswordHasher, plain string) error {
	if len(plain) < 8 {
		return ErrWeakPassword
	}
	hash, err := hasher.Hash(plain)
	if err != nil {
		return err
	}
	u.PasswordHash = hash
	return nil
}

func (u User) Authenticate(hasher PasswordHasher, plain string) error {
	if strings.TrimSpace(u.PasswordHash) == "" {
		return ErrPasswordNotSet
	}
	if err := hasher.Compare(u.PasswordHash, plain); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

func (u User) RequireRole(role Role) error {
	if u.Role != role {
		return ErrForbiddenRole
	}
	return nil
}

type UserRepository interface {
	Save(user User) error
	FindByEmail(email string) (User, error)
	FindByID(id UserID) (User, error)
	ListByRole(role Role) ([]User, error)
}

type TokenClaims struct {
	UserID UserID
	Email  string
	Role   Role
}

type TokenService interface {
	Issue(claims TokenClaims) (token string, err error)
	Parse(token string) (TokenClaims, error)
}
