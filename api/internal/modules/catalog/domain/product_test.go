package domain

import "testing"

func TestNewMoney_rejectsNonPositive(t *testing.T) {
	t.Parallel()
	_, err := NewMoney(0, "VND")
	if err != ErrInvalidProductPrice {
		t.Fatalf("expected ErrInvalidProductPrice, got %v", err)
	}
}

func TestNewProduct_requiresName(t *testing.T) {
	t.Parallel()
	price, err := NewMoney(10000, "VND")
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewProduct("m1", "  ", "desc", price, 1)
	if err != ErrInvalidProductName {
		t.Fatalf("expected ErrInvalidProductName, got %v", err)
	}
}

func TestNewProduct_ok(t *testing.T) {
	t.Parallel()
	price, err := NewMoney(99000, "vnd")
	if err != nil {
		t.Fatal(err)
	}
	p, err := NewProduct("merchant-1", "Áo thun", "Cotton", price, 10)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "Áo thun" {
		t.Fatalf("unexpected name: %s", p.Name)
	}
	if p.Price.Currency != "VND" {
		t.Fatalf("expected VND, got %s", p.Price.Currency)
	}
	if p.ID == "" {
		t.Fatal("expected product id")
	}
}
