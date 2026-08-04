package commands

import (
	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type CreateUserAddressCommand struct {
	UserID        domain.AccountID
	RecipientName string
	PhoneNumber   string
	AddressLine   string
	IsDefault     bool
}

type CreateUserAddressHandler struct {
	repo domain.UserAddressRepository
}

func NewCreateUserAddressHandler(repo domain.UserAddressRepository) *CreateUserAddressHandler {
	return &CreateUserAddressHandler{repo: repo}
}

func (h *CreateUserAddressHandler) Handle(cmd CreateUserAddressCommand) (domain.UserAddress, error) {
	if cmd.IsDefault {
		if err := h.repo.ClearDefault(cmd.UserID); err != nil {
			return domain.UserAddress{}, err
		}
	} else {
		// If it's the first address, make it default
		existing, err := h.repo.ListByUserID(cmd.UserID)
		if err != nil {
			return domain.UserAddress{}, err
		}
		if len(existing) == 0 {
			cmd.IsDefault = true
		}
	}

	a, err := domain.NewUserAddress(cmd.UserID, cmd.RecipientName, cmd.PhoneNumber, cmd.AddressLine, cmd.IsDefault)
	if err != nil {
		return domain.UserAddress{}, err
	}
	if err := h.repo.Save(a); err != nil {
		return domain.UserAddress{}, err
	}
	return a, nil
}

type UpdateUserAddressCommand struct {
	ID            domain.AddressID
	UserID        domain.AccountID
	RecipientName string
	PhoneNumber   string
	AddressLine   string
	IsDefault     bool
}

type UpdateUserAddressHandler struct {
	repo domain.UserAddressRepository
}

func NewUpdateUserAddressHandler(repo domain.UserAddressRepository) *UpdateUserAddressHandler {
	return &UpdateUserAddressHandler{repo: repo}
}

func (h *UpdateUserAddressHandler) Handle(cmd UpdateUserAddressCommand) (domain.UserAddress, error) {
	a, err := h.repo.FindByID(cmd.ID)
	if err != nil {
		return domain.UserAddress{}, err
	}
	if a.UserID != cmd.UserID {
		return domain.UserAddress{}, domain.ErrAddressNotFound // or unauthorized
	}

	if cmd.IsDefault && !a.IsDefault {
		if err := h.repo.ClearDefault(cmd.UserID); err != nil {
			return domain.UserAddress{}, err
		}
	} else if !cmd.IsDefault && a.IsDefault {
		// Cannot unset default if it's the only one, but let's just allow it or pick another.
		// For simplicity, just allow it.
	}

	if err := a.Update(cmd.RecipientName, cmd.PhoneNumber, cmd.AddressLine, cmd.IsDefault); err != nil {
		return domain.UserAddress{}, err
	}
	if err := h.repo.Save(a); err != nil {
		return domain.UserAddress{}, err
	}
	return a, nil
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
