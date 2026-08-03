package commands

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type CreateMerchantCommand struct {
	Email       string
	DisplayName string
	Password    string
}

type MerchantResult struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type CreateMerchantHandler struct {
	merchants domain.AccountRepository
	hasher    domain.PasswordHasher
}

func NewCreateMerchantHandler(merchants domain.AccountRepository, hasher domain.PasswordHasher) *CreateMerchantHandler {
	return &CreateMerchantHandler{merchants: merchants, hasher: hasher}
}

func (h *CreateMerchantHandler) Handle(_ context.Context, cmd CreateMerchantCommand) (MerchantResult, error) {
	if _, err := h.merchants.FindByEmail(cmd.Email); err == nil {
		return MerchantResult{}, domain.ErrEmailTaken
	} else if err != domain.ErrAccountNotFound {
		return MerchantResult{}, err
	}

	account, err := domain.NewAccount(cmd.Email, cmd.DisplayName)
	if err != nil {
		return MerchantResult{}, err
	}
	if err := account.SetPassword(h.hasher, cmd.Password); err != nil {
		return MerchantResult{}, err
	}
	if err := h.merchants.Save(account); err != nil {
		return MerchantResult{}, err
	}
	return toMerchantResult(account), nil
}

func toMerchantResult(account domain.Account) MerchantResult {
	return MerchantResult{
		ID:          string(account.ID),
		Email:       account.Email,
		DisplayName: account.DisplayName,
		Role:        string(domain.RoleMerchant),
	}
}
