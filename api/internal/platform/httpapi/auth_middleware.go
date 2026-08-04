package httpapi

import (
	"net/http"
	"strings"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/authctx"
)

func BearerAuth(tokens domain.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(strings.ToLower(header), "bearer ") {
				writeAuthError(w, http.StatusUnauthorized, "missing bearer token")
				return
			}
			raw := strings.TrimSpace(header[7:])
			claims, err := tokens.Parse(raw)
			if err != nil {
				writeAuthError(w, http.StatusUnauthorized, "invalid token")
				return
			}
			ctx := authctx.WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := authctx.FromContext(r.Context())
		if !ok || claims.Role != domain.RoleAdmin {
			writeAuthError(w, http.StatusForbidden, "admin role required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireMerchant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := authctx.FromContext(r.Context())
		if !ok || claims.Role != domain.RoleMerchant {
			writeAuthError(w, http.StatusForbidden, "merchant role required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := authctx.FromContext(r.Context())
		if !ok || claims.Role != domain.RoleUser {
			writeAuthError(w, http.StatusForbidden, "user role required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + message + `"}`))
}
