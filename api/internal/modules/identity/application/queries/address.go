package queries

import (
	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type ListUserAddressesHandler struct {
	repo domain.UserAddressRepository
}

func NewListUserAddressesHandler(repo domain.UserAddressRepository) *ListUserAddressesHandler {
	return &ListUserAddressesHandler{repo: repo}
}

func (h *ListUserAddressesHandler) Handle(userID domain.AccountID) ([]domain.UserAddress, error) {
	return h.repo.ListByUserID(userID)
}

type GetUserAddressHandler struct {
	repo domain.UserAddressRepository
}

func NewGetUserAddressHandler(repo domain.UserAddressRepository) *GetUserAddressHandler {
	return &GetUserAddressHandler{repo: repo}
}

func (h *GetUserAddressHandler) Handle(id domain.AddressID, userID domain.AccountID) (domain.UserAddress, error) {
	a, err := h.repo.FindByID(id)
	if err != nil {
		return domain.UserAddress{}, err
	}
	if a.UserID != userID {
		return domain.UserAddress{}, domain.ErrAddressNotFound
	}
	return a, nil
}
