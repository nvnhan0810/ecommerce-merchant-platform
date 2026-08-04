package queries

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type GetCurrentUserQuery struct {
	UserID domain.AccountID
}

type GetCurrentUserHandler struct {
	accounts domain.AccountRepository
	role     domain.Role
}

func NewGetCurrentUserHandler(accounts domain.AccountRepository, role domain.Role) *GetCurrentUserHandler {
	return &GetCurrentUserHandler{accounts: accounts, role: role}
}

func (h *GetCurrentUserHandler) Handle(_ context.Context, q GetCurrentUserQuery) (AccountDTO, error) {
	account, err := h.accounts.FindByID(q.UserID)
	if err != nil {
		return AccountDTO{}, err
	}
	return AccountDTO{
		ID:          string(account.ID),
		Email:       account.Email,
		DisplayName: account.DisplayName,
		Role:        string(h.role),
	}, nil
}
