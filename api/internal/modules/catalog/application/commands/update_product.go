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
	CategoryIDs     []string
	SetCategories   bool // true when CategoryIDs was present in request (including empty)
}

type UpdateProductHandler struct {
	repo       domain.ProductRepository
	categories domain.CategoryRepository
	merchants  domain.MerchantChecker
	publicBase string
	setCats    *SetProductCategoriesHandler
}

func NewUpdateProductHandler(
	repo domain.ProductRepository,
	categories domain.CategoryRepository,
	merchants domain.MerchantChecker,
	publicBase string,
) *UpdateProductHandler {
	return &UpdateProductHandler{
		repo:       repo,
		categories: categories,
		merchants:  merchants,
		publicBase: publicBase,
		setCats:    NewSetProductCategoriesHandler(repo, categories),
	}
}

func (h *UpdateProductHandler) Handle(ctx context.Context, cmd UpdateProductCommand) (ProductResult, error) {
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
	if cmd.SetCategories {
		if err := h.setCats.Handle(ctx, SetProductCategoriesCommand{
			ProductID:       product.ID,
			CategoryIDs:     cmd.CategoryIDs,
			OwnerMerchantID: cmd.OwnerMerchantID,
		}); err != nil {
			return ProductResult{}, err
		}
	}
	return productResultWithCategories(product, h.categories, h.publicBase)
}
