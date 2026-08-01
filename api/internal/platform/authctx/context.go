package authctx

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type ctxKey struct{}

func WithClaims(ctx context.Context, claims domain.TokenClaims) context.Context {
	return context.WithValue(ctx, ctxKey{}, claims)
}

func FromContext(ctx context.Context) (domain.TokenClaims, bool) {
	claims, ok := ctx.Value(ctxKey{}).(domain.TokenClaims)
	return claims, ok
}
