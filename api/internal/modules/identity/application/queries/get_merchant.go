package queries

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type GetMerchantQuery struct {
	ID domain.AccountID
}

type GetMerchantHandler struct {
	merchants domain.AccountRepository
}

func NewGetMerchantHandler(merchants domain.AccountRepository) *GetMerchantHandler {
	return &GetMerchantHandler{merchants: merchants}
}

func (h *GetMerchantHandler) Handle(_ context.Context, q GetMerchantQuery) (AccountDTO, error) {
	account, err := h.merchants.FindByID(q.ID)
	if err != nil {
		return AccountDTO{}, err
	}
	return AccountDTO{
		ID:          string(account.ID),
		Email:       account.Email,
		DisplayName: account.DisplayName,
		Role:        string(domain.RoleMerchant),
	}, nil
}
