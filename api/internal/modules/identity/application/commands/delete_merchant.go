package commands

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type DeleteMerchantCommand struct {
	ID domain.AccountID
}

type DeleteMerchantHandler struct {
	merchants domain.AccountRepository
}

func NewDeleteMerchantHandler(merchants domain.AccountRepository) *DeleteMerchantHandler {
	return &DeleteMerchantHandler{merchants: merchants}
}

func (h *DeleteMerchantHandler) Handle(_ context.Context, cmd DeleteMerchantCommand) error {
	if _, err := h.merchants.FindByID(cmd.ID); err != nil {
		return err
	}
	return h.merchants.Delete(cmd.ID)
}
