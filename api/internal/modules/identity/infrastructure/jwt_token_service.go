package infrastructure

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type JWTTokenService struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTTokenService(secret string, ttl time.Duration) (*JWTTokenService, error) {
	if len(secret) < 16 {
		return nil, errors.New("JWT_SECRET must be at least 16 characters")
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &JWTTokenService{secret: []byte(secret), ttl: ttl}, nil
}

type jwtClaims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

func (s *JWTTokenService) Issue(claims domain.TokenClaims) (string, error) {
	now := time.Now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims{
		Email: claims.Email,
		Role:  string(claims.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   string(claims.UserID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	})
	return token.SignedString(s.secret)
}

func (s *JWTTokenService) Parse(raw string) (domain.TokenClaims, error) {
	parsed, err := jwt.ParseWithClaims(raw, &jwtClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return domain.TokenClaims{}, domain.ErrInvalidCredentials
	}
	c, ok := parsed.Claims.(*jwtClaims)
	if !ok || !parsed.Valid {
		return domain.TokenClaims{}, domain.ErrInvalidCredentials
	}
	role, err := domain.ParseRole(c.Role)
	if err != nil {
		return domain.TokenClaims{}, err
	}
	return domain.TokenClaims{
		UserID: domain.UserID(c.Subject),
		Email:  c.Email,
		Role:   role,
	}, nil
}
