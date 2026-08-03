package commands

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
)

type UpdateProductCommand struct {
	ID          domain.ProductID
	MerchantID  string
	Name        string
	Description string
	PriceCents  int64
	Currency    string
	Stock       int
}

type UpdateProductHandler struct {
	repo      domain.ProductRepository
	merchants domain.MerchantChecker
}

func NewUpdateProductHandler(repo domain.ProductRepository, merchants domain.MerchantChecker) *UpdateProductHandler {
	return &UpdateProductHandler{repo: repo, merchants: merchants}
}

func (h *UpdateProductHandler) Handle(_ context.Context, cmd UpdateProductCommand) (ProductResult, error) {
	product, err := h.repo.FindByID(cmd.ID)
	if err != nil {
		return ProductResult{}, err
	}
	if err := h.merchants.EnsureExists(cmd.MerchantID); err != nil {
		return ProductResult{}, err
	}
	price, err := domain.NewMoney(cmd.PriceCents, cmd.Currency)
	if err != nil {
		return ProductResult{}, err
	}
	if err := product.Update(cmd.MerchantID, cmd.Name, cmd.Description, price, cmd.Stock); err != nil {
		return ProductResult{}, err
	}
	if err := h.repo.Save(product); err != nil {
		return ProductResult{}, err
	}
	return toProductResult(product), nil
}
