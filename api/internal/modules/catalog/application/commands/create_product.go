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

type CreateProductResult struct {
	ID string `json:"id"`
}

type CreateProductHandler struct {
	repo domain.ProductRepository
}

func NewCreateProductHandler(repo domain.ProductRepository) *CreateProductHandler {
	return &CreateProductHandler{repo: repo}
}

func (h *CreateProductHandler) Handle(_ context.Context, cmd CreateProductCommand) (CreateProductResult, error) {
	price, err := domain.NewMoney(cmd.PriceCents, cmd.Currency)
	if err != nil {
		return CreateProductResult{}, err
	}
	product, err := domain.NewProduct(cmd.MerchantID, cmd.Name, cmd.Description, price, cmd.Stock)
	if err != nil {
		return CreateProductResult{}, err
	}
	if err := h.repo.Save(product); err != nil {
		return CreateProductResult{}, err
	}
	return CreateProductResult{ID: string(product.ID)}, nil
}
