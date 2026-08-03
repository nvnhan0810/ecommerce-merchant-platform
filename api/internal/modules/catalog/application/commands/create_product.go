package commands

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
)

type CreateProductCommand struct {
	MerchantID  string
	Name        string
	Description string
	PriceCents  int64
	Currency    string
	Stock       int
}

type ProductResult struct {
	ID          string `json:"id"`
	MerchantID  string `json:"merchant_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents"`
	Currency    string `json:"currency"`
	Stock       int    `json:"stock"`
}

type CreateProductHandler struct {
	repo      domain.ProductRepository
	merchants domain.MerchantChecker
}

func NewCreateProductHandler(repo domain.ProductRepository, merchants domain.MerchantChecker) *CreateProductHandler {
	return &CreateProductHandler{repo: repo, merchants: merchants}
}

func (h *CreateProductHandler) Handle(_ context.Context, cmd CreateProductCommand) (ProductResult, error) {
	if err := h.merchants.EnsureExists(cmd.MerchantID); err != nil {
		return ProductResult{}, err
	}
	price, err := domain.NewMoney(cmd.PriceCents, cmd.Currency)
	if err != nil {
		return ProductResult{}, err
	}
	product, err := domain.NewProduct(cmd.MerchantID, cmd.Name, cmd.Description, price, cmd.Stock)
	if err != nil {
		return ProductResult{}, err
	}
	if err := h.repo.Save(product); err != nil {
		return ProductResult{}, err
	}
	return toProductResult(product), nil
}

func toProductResult(product domain.Product) ProductResult {
	return ProductResult{
		ID:          string(product.ID),
		MerchantID:  product.MerchantID,
		Name:        product.Name,
		Description: product.Description,
		PriceCents:  product.Price.AmountCents,
		Currency:    product.Price.Currency,
		Stock:       product.Stock,
	}
}
