package commands

import (
	"context"
	"io"
	"strings"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/mediaurl"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/storage"
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
	ImageKey    string `json:"image_key"`
	ImageURL    string `json:"image_url"`
}

type CreateProductHandler struct {
	repo       domain.ProductRepository
	merchants  domain.MerchantChecker
	publicBase string
}

func NewCreateProductHandler(repo domain.ProductRepository, merchants domain.MerchantChecker, publicBase string) *CreateProductHandler {
	return &CreateProductHandler{repo: repo, merchants: merchants, publicBase: publicBase}
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
	return toProductResult(product, h.publicBase), nil
}

func toProductResult(product domain.Product, publicBase string) ProductResult {
	return ProductResult{
		ID:          string(product.ID),
		MerchantID:  product.MerchantID,
		Name:        product.Name,
		Description: product.Description,
		PriceCents:  product.Price.AmountCents,
		Currency:    product.Price.Currency,
		Stock:       product.Stock,
		ImageKey:    product.ImageKey,
		ImageURL:    mediaurl.Absolute(publicBase, product.ImageKey),
	}
}

type UploadProductImageCommand struct {
	ID              domain.ProductID
	OwnerMerchantID string
	Filename        string
	ContentType     string
	Size            int64
	Body            io.Reader
}

type UploadProductImageHandler struct {
	repo       domain.ProductRepository
	store      storage.ObjectStore
	publicBase string
}

func NewUploadProductImageHandler(repo domain.ProductRepository, store storage.ObjectStore, publicBase string) *UploadProductImageHandler {
	return &UploadProductImageHandler{repo: repo, store: store, publicBase: publicBase}
}

func (h *UploadProductImageHandler) Handle(ctx context.Context, cmd UploadProductImageCommand) (ProductResult, error) {
	if h.store == nil || !h.store.Enabled() {
		return ProductResult{}, storage.ErrObjectStoreDisabled
	}
	product, err := h.repo.FindByID(cmd.ID)
	if err != nil {
		return ProductResult{}, err
	}
	if cmd.OwnerMerchantID != "" && product.MerchantID != cmd.OwnerMerchantID {
		return ProductResult{}, domain.ErrProductNotFound
	}
	ct := strings.ToLower(strings.TrimSpace(cmd.ContentType))
	switch ct {
	case "image/jpeg", "image/jpg", "image/png", "image/webp", "image/gif":
	default:
		return ProductResult{}, domain.ErrInvalidImage
	}
	if cmd.Size <= 0 || cmd.Size > 5*1024*1024 {
		return ProductResult{}, domain.ErrInvalidImage
	}

	key := h.store.NewProductImageKey(product.MerchantID, string(product.ID), cmd.Filename)
	if err := h.store.Upload(ctx, key, cmd.Body, ct, cmd.Size); err != nil {
		return ProductResult{}, err
	}
	old := product.ImageKey
	product.SetImageKey(key)
	if err := h.repo.Save(product); err != nil {
		_ = h.store.Delete(ctx, key)
		return ProductResult{}, err
	}
	if old != "" && old != key {
		_ = h.store.Delete(ctx, old)
	}
	return toProductResult(product, h.publicBase), nil
}

type DeleteProductImageHandler struct {
	repo       domain.ProductRepository
	store      storage.ObjectStore
	publicBase string
}

func NewDeleteProductImageHandler(repo domain.ProductRepository, store storage.ObjectStore, publicBase string) *DeleteProductImageHandler {
	return &DeleteProductImageHandler{repo: repo, store: store, publicBase: publicBase}
}

type DeleteProductImageCommand struct {
	ID              domain.ProductID
	OwnerMerchantID string
}

func (h *DeleteProductImageHandler) Handle(ctx context.Context, cmd DeleteProductImageCommand) (ProductResult, error) {
	product, err := h.repo.FindByID(cmd.ID)
	if err != nil {
		return ProductResult{}, err
	}
	if cmd.OwnerMerchantID != "" && product.MerchantID != cmd.OwnerMerchantID {
		return ProductResult{}, domain.ErrProductNotFound
	}
	old := product.ImageKey
	product.ClearImage()
	if err := h.repo.Save(product); err != nil {
		return ProductResult{}, err
	}
	if old != "" && h.store != nil && h.store.Enabled() {
		_ = h.store.Delete(ctx, old)
	}
	return toProductResult(product, h.publicBase), nil
}
