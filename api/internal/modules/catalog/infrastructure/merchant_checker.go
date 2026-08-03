package infrastructure

import (
	"strings"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
	identitydomain "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

// AccountMerchantChecker ensures merchant_id points at a row in merchants.
type AccountMerchantChecker struct {
	merchants identitydomain.AccountRepository
}

func NewAccountMerchantChecker(merchants identitydomain.AccountRepository) *AccountMerchantChecker {
	return &AccountMerchantChecker{merchants: merchants}
}

func (c *AccountMerchantChecker) EnsureExists(merchantID string) error {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return domain.ErrMerchantRequired
	}
	_, err := c.merchants.FindByID(identitydomain.AccountID(merchantID))
	if err == identitydomain.ErrAccountNotFound {
		return domain.ErrMerchantNotFound
	}
	return err
}
