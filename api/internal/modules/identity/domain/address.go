package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidAddressID = errors.New("invalid address id")
	ErrAddressNotFound  = errors.New("address not found")
	ErrMissingAddressFields = errors.New("recipient name, phone number, and address line are required")
)

type AddressID string

func NewAddressID() AddressID {
	return AddressID(uuid.NewString())
}

func ParseAddressID(raw string) (AddressID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidAddressID
	}
	return AddressID(raw), nil
}

type UserAddress struct {
	ID            AddressID
	UserID        AccountID
	RecipientName string
	PhoneNumber   string
	AddressLine   string
	IsDefault     bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewUserAddress(userID AccountID, recipientName, phoneNumber, addressLine string, isDefault bool) (UserAddress, error) {
	recipientName = strings.TrimSpace(recipientName)
	phoneNumber = strings.TrimSpace(phoneNumber)
	addressLine = strings.TrimSpace(addressLine)

	if recipientName == "" || phoneNumber == "" || addressLine == "" {
		return UserAddress{}, ErrMissingAddressFields
	}

	now := time.Now().UTC()
	return UserAddress{
		ID:            NewAddressID(),
		UserID:        userID,
		RecipientName: recipientName,
		PhoneNumber:   phoneNumber,
		AddressLine:   addressLine,
		IsDefault:     isDefault,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func (a *UserAddress) Update(recipientName, phoneNumber, addressLine string, isDefault bool) error {
	recipientName = strings.TrimSpace(recipientName)
	phoneNumber = strings.TrimSpace(phoneNumber)
	addressLine = strings.TrimSpace(addressLine)

	if recipientName == "" || phoneNumber == "" || addressLine == "" {
		return ErrMissingAddressFields
	}

	a.RecipientName = recipientName
	a.PhoneNumber = phoneNumber
	a.AddressLine = addressLine
	a.IsDefault = isDefault
	a.UpdatedAt = time.Now().UTC()
	return nil
}

type UserAddressRepository interface {
	Save(address UserAddress) error
	FindByID(id AddressID) (UserAddress, error)
	ListByUserID(userID AccountID) ([]UserAddress, error)
	Delete(id AddressID) error
	ClearDefault(userID AccountID) error
}
