package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidAddressID     = errors.New("invalid address id")
	ErrAddressNotFound      = errors.New("address not found")
	ErrMissingAddressFields = errors.New("address line, province, and ward are required")
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
	ID           AddressID `json:"id"`
	UserID       AccountID `json:"user_id"`
	AddressLine  string    `json:"address_line"`
	CountryCode  string    `json:"country_code"`
	ProvinceCode string    `json:"province_code"`
	WardCode     string    `json:"ward_code"`
	CountryName  string    `json:"country_name,omitempty"`
	ProvinceName string    `json:"province_name,omitempty"`
	WardName     string    `json:"ward_name,omitempty"`
	Latitude     *float64  `json:"latitude,omitempty"`
	Longitude    *float64  `json:"longitude,omitempty"`
	IsDefault    bool      `json:"is_default"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AddressFields struct {
	AddressLine  string
	CountryCode  string
	ProvinceCode string
	WardCode     string
	Latitude     *float64
	Longitude    *float64
	IsDefault    bool
}

func NormalizeAddressFields(f AddressFields) (AddressFields, error) {
	f.AddressLine = strings.TrimSpace(f.AddressLine)
	f.CountryCode = strings.ToUpper(strings.TrimSpace(f.CountryCode))
	f.ProvinceCode = strings.TrimSpace(f.ProvinceCode)
	f.WardCode = strings.TrimSpace(f.WardCode)

	if f.CountryCode == "" {
		f.CountryCode = "VN"
	}
	if f.AddressLine == "" || f.ProvinceCode == "" || f.WardCode == "" {
		return AddressFields{}, ErrMissingAddressFields
	}
	return f, nil
}

func NewUserAddress(userID AccountID, fields AddressFields) (UserAddress, error) {
	fields, err := NormalizeAddressFields(fields)
	if err != nil {
		return UserAddress{}, err
	}

	now := time.Now().UTC()
	return UserAddress{
		ID:           NewAddressID(),
		UserID:       userID,
		AddressLine:  fields.AddressLine,
		CountryCode:  fields.CountryCode,
		ProvinceCode: fields.ProvinceCode,
		WardCode:     fields.WardCode,
		Latitude:     fields.Latitude,
		Longitude:    fields.Longitude,
		IsDefault:    fields.IsDefault,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (a *UserAddress) Update(fields AddressFields) error {
	fields, err := NormalizeAddressFields(fields)
	if err != nil {
		return err
	}

	a.AddressLine = fields.AddressLine
	a.CountryCode = fields.CountryCode
	a.ProvinceCode = fields.ProvinceCode
	a.WardCode = fields.WardCode
	a.Latitude = fields.Latitude
	a.Longitude = fields.Longitude
	a.IsDefault = fields.IsDefault
	a.UpdatedAt = time.Now().UTC()
	return nil
}

func (a UserAddress) FullAddress() string {
	parts := make([]string, 0, 4)
	for _, p := range []string{a.AddressLine, a.WardName, a.ProvinceName, a.CountryName} {
		p = strings.TrimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, ", ")
}

type UserAddressRepository interface {
	Save(address UserAddress) error
	FindByID(id AddressID) (UserAddress, error)
	ListByUserID(userID AccountID) ([]UserAddress, error)
	Delete(id AddressID) error
	ClearDefault(userID AccountID) error
}
