package queries

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type AccountDTO struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type ListAccountsHandler struct {
	repo domain.AccountRepository
	role domain.Role
}

func NewListUsersHandler(repo domain.AccountRepository) *ListAccountsHandler {
	return &ListAccountsHandler{repo: repo, role: domain.RoleUser}
}

func NewListMerchantsHandler(repo domain.AccountRepository) *ListAccountsHandler {
	return &ListAccountsHandler{repo: repo, role: domain.RoleMerchant}
}

func (h *ListAccountsHandler) Handle(_ context.Context) ([]AccountDTO, error) {
	items, err := h.repo.List()
	if err != nil {
		return nil, err
	}
	out := make([]AccountDTO, 0, len(items))
	for _, a := range items {
		out = append(out, AccountDTO{
			ID:          string(a.ID),
			Email:       a.Email,
			DisplayName: a.DisplayName,
			Role:        string(h.role),
		})
	}
	return out, nil
}
