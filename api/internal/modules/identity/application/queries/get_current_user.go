package queries

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type GetCurrentUserQuery struct {
	UserID domain.AccountID
}

type GetCurrentUserHandler struct {
	admins domain.AccountRepository
}

func NewGetCurrentUserHandler(admins domain.AccountRepository) *GetCurrentUserHandler {
	return &GetCurrentUserHandler{admins: admins}
}

func (h *GetCurrentUserHandler) Handle(_ context.Context, q GetCurrentUserQuery) (AccountDTO, error) {
	account, err := h.admins.FindByID(q.UserID)
	if err != nil {
		return AccountDTO{}, err
	}
	return AccountDTO{
		ID:          string(account.ID),
		Email:       account.Email,
		DisplayName: account.DisplayName,
		Role:        string(domain.RoleAdmin),
	}, nil
}
