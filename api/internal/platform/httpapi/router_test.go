package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/application/commands"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/application/queries"
	cataloginfra "github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/infrastructure"
	catalogpres "github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/presentation"
	healthpres "github.com/nvnhan0810/ecomerce-api/internal/modules/health/presentation"
	identitycommands "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/application/commands"
	identityqueries "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/application/queries"
	identityinfra "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/infrastructure"
	identitypres "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/presentation"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/config"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/httpapi"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	productRepo := cataloginfra.NewInMemoryProductRepository()
	if err := cataloginfra.SeedDemoProducts(productRepo); err != nil {
		t.Fatal(err)
	}
	userRepo := identityinfra.NewInMemoryUserRepository()
	hasher := identityinfra.NewBcryptPasswordHasher()
	tokens, err := identityinfra.NewJWTTokenService("test-jwt-secret-16", 2*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if err := identityinfra.SeedDemoUsers(userRepo, hasher, "Admin@123456"); err != nil {
		t.Fatal(err)
	}
	return httpapi.NewRouter(httpapi.Dependencies{
		Config: config.Config{
			HTTPAddr:    ":0",
			Env:         "test",
			CORSOrigins: []string{"http://localhost:5173"},
		},
		Health: healthpres.NewHealthHandler("test"),
		Catalog: catalogpres.NewCatalogHandler(
			queries.NewListProductsHandler(productRepo),
			commands.NewCreateProductHandler(productRepo),
		),
		Identity: identitypres.NewIdentityHandler(
			identityqueries.NewListUsersByRoleHandler(userRepo),
			identitycommands.NewLoginHandler(userRepo, hasher, tokens),
			identityqueries.NewGetCurrentUserHandler(userRepo),
		),
		Tokens: tokens,
	})
}

func TestHealthEndpoint_should_return_ok(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestListProducts_should_return_seeded_items(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLogin_should_return_token_for_admin(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	body := bytes.NewBufferString(`{"email":"admin@ecomerce.local","password":"Admin@123456"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.AccessToken == "" {
		t.Fatal("expected access token")
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+payload.AccessToken)
	meRec := httptest.NewRecorder()
	srv.ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("me status=%d body=%s", meRec.Code, meRec.Body.String())
	}
}

func TestMerchants_should_require_admin_token(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/merchants", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
