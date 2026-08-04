package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidEmail       = errors.New("email is required")
	ErrAccountNotFound    = errors.New("account not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrPasswordNotSet     = errors.New("password is not set for this account")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrEmailTaken         = errors.New("email is already taken")
	ErrInvalidAccountID   = errors.New("account id is required")
	ErrInvalidAvatar      = errors.New("avatar must be jpeg, png, webp, or gif and at most 5MB")
)

// Role is used only in JWT claims / API responses — not stored on account tables.
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
		return "", errors.New("invalid role")
	}
}

type AccountID string

func NewAccountID() AccountID {
	return AccountID(uuid.NewString())
}

func ParseAccountID(raw string) (AccountID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidAccountID
	}
	return AccountID(raw), nil
}

type PasswordHasher interface {
	Hash(plain string) (string, error)
	Compare(hash, plain string) error
}

// Account is the shared shape for rows in users / merchants / admins tables.
// Avatar and shop address fields apply to merchants only (empty for users/admins).
type Account struct {
	ID           AccountID
	Email        string
	DisplayName  string
	PasswordHash string
	CreatedAt    time.Time
	AvatarKey    string
	AddressLine  string
	CountryCode  string
	ProvinceCode string
	WardCode     string
}

func NewAccount(email, displayName string) (Account, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return Account{}, ErrInvalidEmail
	}
	return Account{
		ID:          NewAccountID(),
		Email:       email,
		DisplayName: strings.TrimSpace(displayName),
		CreatedAt:   time.Now().UTC(),
	}, nil
}

func (a *Account) SetPassword(hasher PasswordHasher, plain string) error {
	if len(plain) < 8 {
		return ErrWeakPassword
	}
	hash, err := hasher.Hash(plain)
	if err != nil {
		return err
	}
	a.PasswordHash = hash
	return nil
}

func (a Account) Authenticate(hasher PasswordHasher, plain string) error {
	if strings.TrimSpace(a.PasswordHash) == "" {
		return ErrPasswordNotSet
	}
	if err := hasher.Compare(a.PasswordHash, plain); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

func (a *Account) Rename(displayName string) {
	a.DisplayName = strings.TrimSpace(displayName)
}

func (a *Account) ChangeEmail(email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return ErrInvalidEmail
	}
	a.Email = email
	return nil
}

func (a *Account) SetAvatarKey(key string) {
	a.AvatarKey = strings.TrimSpace(strings.TrimPrefix(key, "/"))
}

func (a *Account) ClearShopAddress() {
	a.AddressLine = ""
	a.CountryCode = ""
	a.ProvinceCode = ""
	a.WardCode = ""
}

func (a *Account) SetShopAddress(addressLine, countryCode, provinceCode, wardCode string) {
	a.AddressLine = strings.TrimSpace(addressLine)
	a.CountryCode = strings.ToUpper(strings.TrimSpace(countryCode))
	a.ProvinceCode = strings.TrimSpace(provinceCode)
	a.WardCode = strings.TrimSpace(wardCode)
}

// ShopAddressProvided reports whether any shop address field is set.
func (a Account) ShopAddressProvided() bool {
	return a.AddressLine != "" || a.CountryCode != "" || a.ProvinceCode != "" || a.WardCode != ""
}

type AccountRepository interface {
	Save(account Account) error
	FindByEmail(email string) (Account, error)
	FindByID(id AccountID) (Account, error)
	List() ([]Account, error)
	Delete(id AccountID) error
	Count() (int, error)
}

type TokenClaims struct {
	UserID AccountID
	Email  string
	Role   Role
}

type TokenService interface {
	Issue(claims TokenClaims) (token string, err error)
	Parse(token string) (TokenClaims, error)
}

// Deprecated aliases kept temporarily for fewer mechanical renames in JWT package.
type UserID = AccountID
