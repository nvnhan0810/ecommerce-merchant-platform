package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidCategoryName   = errors.New("category name is required")
	ErrInvalidCategoryID     = errors.New("category id is required")
	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryExists        = errors.New("category already exists")
	ErrInvalidCategoryStatus = errors.New("invalid category status")
	ErrCategoryNotAssignable = errors.New("category cannot be assigned")
)

type CategoryStatus string

const (
	CategoryStatusPending  CategoryStatus = "pending"
	CategoryStatusApproved CategoryStatus = "approved"
	CategoryStatusRejected CategoryStatus = "rejected"
)

func ParseCategoryStatus(raw string) (CategoryStatus, error) {
	s := CategoryStatus(strings.ToLower(strings.TrimSpace(raw)))
	switch s {
	case CategoryStatusPending, CategoryStatusApproved, CategoryStatusRejected:
		return s, nil
	default:
		return "", ErrInvalidCategoryStatus
	}
}

func (s CategoryStatus) LabelVI() string {
	switch s {
	case CategoryStatusPending:
		return "Chờ duyệt"
	case CategoryStatusApproved:
		return "Đã duyệt"
	case CategoryStatusRejected:
		return "Từ chối"
	default:
		return string(s)
	}
}

type CategoryID string

func NewCategoryID() CategoryID {
	return CategoryID(uuid.NewString())
}

func ParseCategoryID(raw string) (CategoryID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidCategoryID
	}
	return CategoryID(raw), nil
}

type Category struct {
	ID                   CategoryID
	Name                 string
	Status               CategoryStatus
	CreatedByMerchantID  string
	CreatedAt            time.Time
}

func NewCategory(name, createdByMerchantID string, status CategoryStatus) (Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Category{}, ErrInvalidCategoryName
	}
	if status == "" {
		status = CategoryStatusPending
	}
	if _, err := ParseCategoryStatus(string(status)); err != nil {
		return Category{}, err
	}
	return Category{
		ID:                  NewCategoryID(),
		Name:                name,
		Status:              status,
		CreatedByMerchantID: strings.TrimSpace(createdByMerchantID),
		CreatedAt:           time.Now().UTC(),
	}, nil
}

func (c *Category) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrInvalidCategoryName
	}
	c.Name = name
	return nil
}

func (c *Category) SetStatus(status CategoryStatus) error {
	if _, err := ParseCategoryStatus(string(status)); err != nil {
		return err
	}
	c.Status = status
	return nil
}

// AssignableBy reports whether a merchant (or admin when merchantID="") may link this category.
func (c Category) AssignableBy(merchantID string) bool {
	if c.Status == CategoryStatusApproved {
		return true
	}
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		// admin
		return true
	}
	return c.Status == CategoryStatusPending && c.CreatedByMerchantID == merchantID
}

type CategoryListFilter struct {
	Status              CategoryStatus // empty = any
	CreatedByMerchantID string         // empty = any
	IncludeApproved     bool           // when listing for merchant: approved OR own
	MerchantViewerID    string
}

type CategoryRepository interface {
	Save(category Category) error
	FindByID(id CategoryID) (Category, error)
	FindByName(name string) (Category, error)
	List(filter CategoryListFilter) ([]Category, error)
	Delete(id CategoryID) error
	ListByProductIDs(productIDs []ProductID) (map[ProductID][]Category, error)
	SetProductCategories(productID ProductID, categoryIDs []CategoryID) error
	ListProductIDsByCategory(categoryID CategoryID, approvedOnly bool, limit, offset int) ([]ProductID, error)
}
