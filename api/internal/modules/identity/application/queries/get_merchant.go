package queries

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type GetMerchantQuery struct {
	ID domain.AccountID
}

type GetMerchantHandler struct {
	merchants  domain.AccountRepository
	publicBase string
	geo        domain.GeoRepository
}

func NewGetMerchantHandler(merchants domain.AccountRepository, publicBase string, geo domain.GeoRepository) *GetMerchantHandler {
	return &GetMerchantHandler{merchants: merchants, publicBase: publicBase, geo: geo}
}

func (h *GetMerchantHandler) Handle(_ context.Context, q GetMerchantQuery) (AccountDTO, error) {
	account, err := h.merchants.FindByID(q.ID)
	if err != nil {
		return AccountDTO{}, err
	}
	return ToAccountDTO(account, domain.RoleMerchant, h.publicBase, h.geo), nil
}

func (h *GetMerchantHandler) HandlePublic(_ context.Context, q GetMerchantQuery) (PublicMerchantDTO, error) {
	account, err := h.merchants.FindByID(q.ID)
	if err != nil {
		return PublicMerchantDTO{}, err
	}
	return ToPublicMerchantDTO(account, h.publicBase, h.geo), nil
}
