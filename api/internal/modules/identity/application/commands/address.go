package commands

import (
	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type CreateUserAddressCommand struct {
	UserID       domain.AccountID
	AddressLine  string
	CountryCode  string
	ProvinceCode string
	WardCode     string
	Latitude     *float64
	Longitude    *float64
	IsDefault    bool
}

type CreateUserAddressHandler struct {
	repo domain.UserAddressRepository
	geo  domain.GeoRepository
}

func NewCreateUserAddressHandler(repo domain.UserAddressRepository, geo domain.GeoRepository) *CreateUserAddressHandler {
	return &CreateUserAddressHandler{repo: repo, geo: geo}
}

func (h *CreateUserAddressHandler) Handle(cmd CreateUserAddressCommand) (domain.UserAddress, error) {
	fields, err := validateAddressGeo(h.geo, domain.AddressFields{
		AddressLine:  cmd.AddressLine,
		CountryCode:  cmd.CountryCode,
		ProvinceCode: cmd.ProvinceCode,
		WardCode:     cmd.WardCode,
		Latitude:     cmd.Latitude,
		Longitude:    cmd.Longitude,
		IsDefault:    cmd.IsDefault,
	})
	if err != nil {
		return domain.UserAddress{}, err
	}

	if fields.IsDefault {
		if err := h.repo.ClearDefault(cmd.UserID); err != nil {
			return domain.UserAddress{}, err
		}
	} else {
		existing, err := h.repo.ListByUserID(cmd.UserID)
		if err != nil {
			return domain.UserAddress{}, err
		}
		if len(existing) == 0 {
			fields.IsDefault = true
		}
	}

	a, err := domain.NewUserAddress(cmd.UserID, fields)
	if err != nil {
		return domain.UserAddress{}, err
	}
	if err := h.repo.Save(a); err != nil {
		return domain.UserAddress{}, err
	}
	return h.repo.FindByID(a.ID)
}

type UpdateUserAddressCommand struct {
	ID           domain.AddressID
	UserID       domain.AccountID
	AddressLine  string
	CountryCode  string
	ProvinceCode string
	WardCode     string
	Latitude     *float64
	Longitude    *float64
	IsDefault    bool
}

type UpdateUserAddressHandler struct {
	repo domain.UserAddressRepository
	geo  domain.GeoRepository
}

func NewUpdateUserAddressHandler(repo domain.UserAddressRepository, geo domain.GeoRepository) *UpdateUserAddressHandler {
	return &UpdateUserAddressHandler{repo: repo, geo: geo}
}

func (h *UpdateUserAddressHandler) Handle(cmd UpdateUserAddressCommand) (domain.UserAddress, error) {
	a, err := h.repo.FindByID(cmd.ID)
	if err != nil {
		return domain.UserAddress{}, err
	}
	if a.UserID != cmd.UserID {
		return domain.UserAddress{}, domain.ErrAddressNotFound
	}

	fields, err := validateAddressGeo(h.geo, domain.AddressFields{
		AddressLine:  cmd.AddressLine,
		CountryCode:  cmd.CountryCode,
		ProvinceCode: cmd.ProvinceCode,
		WardCode:     cmd.WardCode,
		Latitude:     cmd.Latitude,
		Longitude:    cmd.Longitude,
		IsDefault:    cmd.IsDefault,
	})
	if err != nil {
		return domain.UserAddress{}, err
	}

	if fields.IsDefault && !a.IsDefault {
		if err := h.repo.ClearDefault(cmd.UserID); err != nil {
			return domain.UserAddress{}, err
		}
	}

	if err := a.Update(fields); err != nil {
		return domain.UserAddress{}, err
	}
	if err := h.repo.Save(a); err != nil {
		return domain.UserAddress{}, err
	}
	return h.repo.FindByID(a.ID)
}

type DeleteUserAddressCommand struct {
	ID     domain.AddressID
	UserID domain.AccountID
}

type DeleteUserAddressHandler struct {
	repo domain.UserAddressRepository
}

func NewDeleteUserAddressHandler(repo domain.UserAddressRepository) *DeleteUserAddressHandler {
	return &DeleteUserAddressHandler{repo: repo}
}

func (h *DeleteUserAddressHandler) Handle(cmd DeleteUserAddressCommand) error {
	a, err := h.repo.FindByID(cmd.ID)
	if err != nil {
		return err
	}
	if a.UserID != cmd.UserID {
		return domain.ErrAddressNotFound
	}
	return h.repo.Delete(cmd.ID)
}

func validateAddressGeo(geo domain.GeoRepository, fields domain.AddressFields) (domain.AddressFields, error) {
	fields, err := domain.NormalizeAddressFields(fields)
	if err != nil {
		return domain.AddressFields{}, err
	}

	province, err := geo.GetProvince(fields.ProvinceCode)
	if err != nil {
		return domain.AddressFields{}, domain.ErrInvalidGeoRef
	}
	if province.CountryCode != fields.CountryCode {
		return domain.AddressFields{}, domain.ErrInvalidGeoRef
	}
	ward, err := geo.GetWard(fields.WardCode)
	if err != nil {
		return domain.AddressFields{}, domain.ErrInvalidGeoRef
	}
	if ward.ProvinceCode != fields.ProvinceCode {
		return domain.AddressFields{}, domain.ErrInvalidGeoRef
	}

	if fields.Latitude == nil {
		fields.Latitude = ward.Latitude
	}
	if fields.Longitude == nil {
		fields.Longitude = ward.Longitude
	}
	return fields, nil
}
