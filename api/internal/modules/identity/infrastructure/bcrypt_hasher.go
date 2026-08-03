package infrastructure

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type BcryptPasswordHasher struct {
	cost int
}

func NewBcryptPasswordHasher() *BcryptPasswordHasher {
	return &BcryptPasswordHasher{cost: bcrypt.DefaultCost}
}

func (h *BcryptPasswordHasher) Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (h *BcryptPasswordHasher) Compare(hash, plain string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)); err != nil {
		return domain.ErrInvalidCredentials
	}
	return nil
}
