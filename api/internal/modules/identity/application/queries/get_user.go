package queries

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type GetUserQuery struct {
	ID domain.AccountID
}

type GetUserHandler struct {
	users domain.AccountRepository
}

func NewGetUserHandler(users domain.AccountRepository) *GetUserHandler {
	return &GetUserHandler{users: users}
}

func (h *GetUserHandler) Handle(_ context.Context, q GetUserQuery) (AccountDTO, error) {
	account, err := h.users.FindByID(q.ID)
	if err != nil {
		return AccountDTO{}, err
	}
	return AccountDTO{
		ID:          string(account.ID),
		Email:       account.Email,
		DisplayName: account.DisplayName,
		Role:        string(domain.RoleUser),
	}, nil
}
