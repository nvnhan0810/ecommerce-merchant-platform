package commands

import (
	"context"
	"testing"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
)

type memoryRepo struct {
	items map[domain.ProductID]domain.Product
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{items: map[domain.ProductID]domain.Product{}}
}

func (r *memoryRepo) Save(p domain.Product) error {
	r.items[p.ID] = p
	return nil
}

func (r *memoryRepo) FindByID(id domain.ProductID) (domain.Product, error) {
	p, ok := r.items[id]
	if !ok {
		return domain.Product{}, domain.ErrProductNotFound
	}
	return p, nil
}

func (r *memoryRepo) List(limit, offset int) ([]domain.Product, error) {
	out := make([]domain.Product, 0, len(r.items))
	for _, p := range r.items {
		out = append(out, p)
	}
	if offset >= len(out) {
		return []domain.Product{}, nil
	}
	end := offset + limit
	if end > len(out) {
		end = len(out)
	}
	return out[offset:end], nil
}

func TestCreateProductHandler_should_create_when_valid(t *testing.T) {
	t.Parallel()
	repo := newMemoryRepo()
	h := NewCreateProductHandler(repo)
	res, err := h.Handle(context.Background(), CreateProductCommand{
		MerchantID:  "m1",
		Name:        "Giày chạy",
		Description: "Light",
		PriceCents:  499000,
		Currency:    "VND",
		Stock:       5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ID == "" {
		t.Fatal("expected id")
	}
	if len(repo.items) != 1 {
		t.Fatalf("expected 1 product, got %d", len(repo.items))
	}
}
