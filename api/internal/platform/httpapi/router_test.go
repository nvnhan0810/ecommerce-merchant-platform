package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/application/commands"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/application/queries"
	cataloginfra "github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/infrastructure"
	catalogpres "github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/presentation"
	healthpres "github.com/nvnhan0810/ecomerce-api/internal/modules/health/presentation"
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
	if err := identityinfra.SeedDemoUsers(userRepo); err != nil {
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
		),
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
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("unexpected body: %v", body)
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
	var body struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Data) < 1 {
		t.Fatal("expected seeded products")
	}
}

func TestListMerchants_should_return_merchant_role_users(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/merchants", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	var body struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Data) != 1 {
		t.Fatalf("expected 1 merchant, got %d", len(body.Data))
	}
}
