package queries

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type GetCurrentUserQuery struct {
	UserID domain.AccountID
}

type GetCurrentUserHandler struct {
	accounts   domain.AccountRepository
	role       domain.Role
	publicBase string
	geo        domain.GeoRepository
}

func NewGetCurrentUserHandler(accounts domain.AccountRepository, role domain.Role) *GetCurrentUserHandler {
	return &GetCurrentUserHandler{accounts: accounts, role: role}
}

func NewGetCurrentMerchantHandler(accounts domain.AccountRepository, publicBase string, geo domain.GeoRepository) *GetCurrentUserHandler {
	return &GetCurrentUserHandler{
		accounts:   accounts,
		role:       domain.RoleMerchant,
		publicBase: publicBase,
		geo:        geo,
	}
}

func (h *GetCurrentUserHandler) Handle(_ context.Context, q GetCurrentUserQuery) (AccountDTO, error) {
	account, err := h.accounts.FindByID(q.UserID)
	if err != nil {
		return AccountDTO{}, err
	}
	return ToAccountDTO(account, h.role, h.publicBase, h.geo), nil
}
