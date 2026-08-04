package commands

import (
	"context"
	"testing"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/storage"
)

type memoryRepo struct {
	items   map[domain.ProductID]domain.Product
	ordered map[domain.ProductID]bool
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{
		items:   map[domain.ProductID]domain.Product{},
		ordered: map[domain.ProductID]bool{},
	}
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

func (r *memoryRepo) ListByMerchant(merchantID string, limit, offset int) ([]domain.Product, error) {
	out := make([]domain.Product, 0)
	for _, p := range r.items {
		if p.MerchantID == merchantID {
			out = append(out, p)
		}
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

func (r *memoryRepo) ListByIDs(ids []domain.ProductID) ([]domain.Product, error) {
	out := make([]domain.Product, 0, len(ids))
	for _, id := range ids {
		if p, ok := r.items[id]; ok {
			out = append(out, p)
		}
	}
	return out, nil
}

func (r *memoryRepo) HasOrderItems(id domain.ProductID) (bool, error) {
	return r.ordered[id], nil
}

func (r *memoryRepo) Delete(id domain.ProductID) error {
	if _, ok := r.items[id]; !ok {
		return domain.ErrProductNotFound
	}
	delete(r.items, id)
	return nil
}

type stubMerchants struct {
	ids map[string]struct{}
}

func (s stubMerchants) EnsureExists(merchantID string) error {
	if merchantID == "" {
		return domain.ErrMerchantRequired
	}
	if _, ok := s.ids[merchantID]; !ok {
		return domain.ErrMerchantNotFound
	}
	return nil
}

func TestCreateProductHandler_should_create_when_valid(t *testing.T) {
	t.Parallel()
	repo := newMemoryRepo()
	h := NewCreateProductHandler(repo, nil, stubMerchants{ids: map[string]struct{}{"m1": {}}}, "https://ecomerce-api.nvnhan0810.com")
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
	if res.ID == "" || res.MerchantID != "m1" {
		t.Fatalf("unexpected: %+v", res)
	}
	if len(repo.items) != 1 {
		t.Fatalf("expected 1 product, got %d", len(repo.items))
	}
}

func TestCreateProductHandler_should_reject_unknown_merchant(t *testing.T) {
	t.Parallel()
	h := NewCreateProductHandler(newMemoryRepo(), nil, stubMerchants{ids: map[string]struct{}{}}, "")
	_, err := h.Handle(context.Background(), CreateProductCommand{
		MerchantID: "missing", Name: "X", PriceCents: 1000, Stock: 1,
	})
	if err != domain.ErrMerchantNotFound {
		t.Fatalf("expected ErrMerchantNotFound, got %v", err)
	}
}

func TestUpdateAndDeleteProduct(t *testing.T) {
	t.Parallel()
	repo := newMemoryRepo()
	merchants := stubMerchants{ids: map[string]struct{}{"m1": {}, "m2": {}}}
	create := NewCreateProductHandler(repo, nil, merchants, "")
	created, err := create.Handle(context.Background(), CreateProductCommand{
		MerchantID: "m1", Name: "Old", PriceCents: 10000, Stock: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	id := domain.ProductID(created.ID)

	update := NewUpdateProductHandler(repo, nil, merchants, "")
	updated, err := update.Handle(context.Background(), UpdateProductCommand{
		ID: id, MerchantID: "m2", Name: "New", Description: "d", PriceCents: 20000, Currency: "VND", Stock: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.MerchantID != "m2" || updated.Name != "New" || updated.Stock != 3 {
		t.Fatalf("unexpected update: %+v", updated)
	}

	del := NewDeleteProductHandler(repo, storage.NopStore{})
	if err := del.Handle(context.Background(), DeleteProductCommand{ID: id}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.FindByID(id); err != domain.ErrProductNotFound {
		t.Fatalf("expected deleted, got %v", err)
	}
}

func TestDeleteProduct_should_reject_when_has_orders(t *testing.T) {
	t.Parallel()
	repo := newMemoryRepo()
	merchants := stubMerchants{ids: map[string]struct{}{"m1": {}}}
	create := NewCreateProductHandler(repo, nil, merchants, "")
	created, err := create.Handle(context.Background(), CreateProductCommand{
		MerchantID: "m1", Name: "Ordered", PriceCents: 10000, Stock: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	id := domain.ProductID(created.ID)
	repo.ordered[id] = true

	del := NewDeleteProductHandler(repo, storage.NopStore{})
	if err := del.Handle(context.Background(), DeleteProductCommand{ID: id}); err != domain.ErrProductHasOrders {
		t.Fatalf("expected ErrProductHasOrders, got %v", err)
	}
	if _, err := repo.FindByID(id); err != nil {
		t.Fatalf("product should still exist, got %v", err)
	}
}

func TestDeleteProduct_should_hide_other_merchant_product(t *testing.T) {
	t.Parallel()
	repo := newMemoryRepo()
	merchants := stubMerchants{ids: map[string]struct{}{"m1": {}, "m2": {}}}
	create := NewCreateProductHandler(repo, nil, merchants, "")
	created, err := create.Handle(context.Background(), CreateProductCommand{
		MerchantID: "m1", Name: "Mine", PriceCents: 10000, Stock: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	id := domain.ProductID(created.ID)

	del := NewDeleteProductHandler(repo, storage.NopStore{})
	if err := del.Handle(context.Background(), DeleteProductCommand{ID: id, OwnerMerchantID: "m2"}); err != domain.ErrProductNotFound {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}
