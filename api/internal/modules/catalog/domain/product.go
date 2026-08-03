package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidProductName  = errors.New("product name is required")
	ErrInvalidProductPrice = errors.New("product price must be greater than zero")
	ErrProductNotFound     = errors.New("product not found")
	ErrMerchantRequired    = errors.New("merchant is required")
	ErrMerchantNotFound    = errors.New("merchant not found")
	ErrInvalidProductID    = errors.New("product id is required")
)

type Money struct {
	AmountCents int64
	Currency    string
}

func NewMoney(amountCents int64, currency string) (Money, error) {
	if amountCents <= 0 {
		return Money{}, ErrInvalidProductPrice
	}
	c := strings.ToUpper(strings.TrimSpace(currency))
	if c == "" {
		c = "VND"
	}
	return Money{AmountCents: amountCents, Currency: c}, nil
}

type ProductID string

func NewProductID() ProductID {
	return ProductID(uuid.NewString())
}

func ParseProductID(raw string) (ProductID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidProductID
	}
	return ProductID(raw), nil
}

type Product struct {
	ID          ProductID
	MerchantID  string
	Name        string
	Description string
	Price       Money
	Stock       int
	CreatedAt   time.Time
}

func NewProduct(merchantID, name, description string, price Money, stock int) (Product, error) {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return Product{}, ErrMerchantRequired
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Product{}, ErrInvalidProductName
	}
	if stock < 0 {
		stock = 0
	}
	return Product{
		ID:          NewProductID(),
		MerchantID:  merchantID,
		Name:        name,
		Description: strings.TrimSpace(description),
		Price:       price,
		Stock:       stock,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

func (p *Product) Update(merchantID, name, description string, price Money, stock int) error {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return ErrMerchantRequired
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrInvalidProductName
	}
	if stock < 0 {
		stock = 0
	}
	p.MerchantID = merchantID
	p.Name = name
	p.Description = strings.TrimSpace(description)
	p.Price = price
	p.Stock = stock
	return nil
}

// MerchantChecker verifies a merchant account exists for product ownership.
type MerchantChecker interface {
	EnsureExists(merchantID string) error
}

type ProductRepository interface {
	Save(product Product) error
	FindByID(id ProductID) (Product, error)
	List(limit, offset int) ([]Product, error)
	Delete(id ProductID) error
}
