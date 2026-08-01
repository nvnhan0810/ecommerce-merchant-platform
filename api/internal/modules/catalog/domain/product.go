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
	name = strings.TrimSpace(name)
	if name == "" {
		return Product{}, ErrInvalidProductName
	}
	if stock < 0 {
		stock = 0
	}
	return Product{
		ID:          NewProductID(),
		MerchantID:  strings.TrimSpace(merchantID),
		Name:        name,
		Description: strings.TrimSpace(description),
		Price:       price,
		Stock:       stock,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

type ProductRepository interface {
	Save(product Product) error
	FindByID(id ProductID) (Product, error)
	List(limit, offset int) ([]Product, error)
}
