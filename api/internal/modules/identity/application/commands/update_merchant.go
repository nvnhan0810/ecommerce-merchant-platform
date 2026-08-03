package commands

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type UpdateMerchantCommand struct {
	ID          domain.AccountID
	Email       string
	DisplayName string
	Password    string
}

type UpdateMerchantHandler struct {
	merchants domain.AccountRepository
	hasher    domain.PasswordHasher
}

func NewUpdateMerchantHandler(merchants domain.AccountRepository, hasher domain.PasswordHasher) *UpdateMerchantHandler {
	return &UpdateMerchantHandler{merchants: merchants, hasher: hasher}
}

func (h *UpdateMerchantHandler) Handle(_ context.Context, cmd UpdateMerchantCommand) (MerchantResult, error) {
	account, err := h.merchants.FindByID(cmd.ID)
	if err != nil {
		return MerchantResult{}, err
	}

	if err := account.ChangeEmail(cmd.Email); err != nil {
		return MerchantResult{}, err
	}
	account.Rename(cmd.DisplayName)

	existing, err := h.merchants.FindByEmail(account.Email)
	if err == nil && existing.ID != account.ID {
		return MerchantResult{}, domain.ErrEmailTaken
	}
	if err != nil && err != domain.ErrAccountNotFound {
		return MerchantResult{}, err
	}

	if cmd.Password != "" {
		if err := account.SetPassword(h.hasher, cmd.Password); err != nil {
			return MerchantResult{}, err
		}
	}

	if err := h.merchants.Save(account); err != nil {
		return MerchantResult{}, err
	}
	return toMerchantResult(account), nil
}
