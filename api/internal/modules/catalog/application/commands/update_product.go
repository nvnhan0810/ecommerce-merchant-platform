package commands

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
)

type UpdateProductCommand struct {
	ID              domain.ProductID
	MerchantID      string
	OwnerMerchantID string
	Name            string
	Description     string
	PriceCents      int64
	Currency        string
	Stock           int
}

type UpdateProductHandler struct {
	repo       domain.ProductRepository
	merchants  domain.MerchantChecker
	publicBase string
}

func NewUpdateProductHandler(repo domain.ProductRepository, merchants domain.MerchantChecker, publicBase string) *UpdateProductHandler {
	return &UpdateProductHandler{repo: repo, merchants: merchants, publicBase: publicBase}
}

func (h *UpdateProductHandler) Handle(_ context.Context, cmd UpdateProductCommand) (ProductResult, error) {
	product, err := h.repo.FindByID(cmd.ID)
	if err != nil {
		return ProductResult{}, err
	}
	if cmd.OwnerMerchantID != "" && product.MerchantID != cmd.OwnerMerchantID {
		return ProductResult{}, domain.ErrProductNotFound
	}
	merchantID := cmd.MerchantID
	if cmd.OwnerMerchantID != "" {
		merchantID = product.MerchantID
	}
	if err := h.merchants.EnsureExists(merchantID); err != nil {
		return ProductResult{}, err
	}
	price, err := domain.NewMoney(cmd.PriceCents, cmd.Currency)
	if err != nil {
		return ProductResult{}, err
	}
	if err := product.Update(merchantID, cmd.Name, cmd.Description, price, cmd.Stock); err != nil {
		return ProductResult{}, err
	}
	if err := h.repo.Save(product); err != nil {
		return ProductResult{}, err
	}
	return toProductResult(product, h.publicBase), nil
}
