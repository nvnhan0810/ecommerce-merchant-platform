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
	"github.com/nvnhan0810/ecomerce-api/internal/platform/seed"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	productRepo := cataloginfra.NewInMemoryProductRepository()
	users := identityinfra.NewInMemoryAccountRepository()
	merchants := identityinfra.NewInMemoryAccountRepository()
	admins := identityinfra.NewInMemoryAccountRepository()
	hasher := identityinfra.NewBcryptPasswordHasher()
	tokens, err := identityinfra.NewJWTTokenService("test-jwt-secret-16", 2*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if err := seed.RunDemo(users, merchants, admins, productRepo, hasher, "Admin@123456"); err != nil {
		t.Fatal(err)
	}
	checker := cataloginfra.NewAccountMerchantChecker(merchants)
	return httpapi.NewRouter(httpapi.Dependencies{
		Config: config.Config{
			HTTPAddr:    ":0",
			Env:         "test",
			CORSOrigins: []string{"http://localhost:5173"},
		},
		Health: healthpres.NewHealthHandler("test"),
		Catalog: catalogpres.NewCatalogHandler(
			queries.NewListProductsHandler(productRepo),
			queries.NewGetProductHandler(productRepo),
			commands.NewCreateProductHandler(productRepo, checker),
			commands.NewUpdateProductHandler(productRepo, checker),
			commands.NewDeleteProductHandler(productRepo),
		),
		Identity: identitypres.NewIdentityHandler(
			identityqueries.NewListUsersHandler(users),
			identityqueries.NewListMerchantsHandler(merchants),
			identityqueries.NewGetUserHandler(users),
			identityqueries.NewGetMerchantHandler(merchants),
			identitycommands.NewLoginHandler(admins, hasher, tokens),
			identityqueries.NewGetCurrentUserHandler(admins),
			identitycommands.NewCreateUserHandler(users, hasher),
			identitycommands.NewUpdateUserHandler(users, hasher),
			identitycommands.NewDeleteUserHandler(users),
			identitycommands.NewCreateMerchantHandler(merchants, hasher),
			identitycommands.NewUpdateMerchantHandler(merchants, hasher),
			identitycommands.NewDeleteMerchantHandler(merchants),
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

func adminToken(t *testing.T, srv http.Handler) string {
	t.Helper()
	body := bytes.NewBufferString(`{"email":"admin@ecomerce.local","password":"Admin@123456"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	return payload.AccessToken
}

func TestLogin_should_return_token_for_admin(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	token := adminToken(t, srv)
	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+token)
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

func TestUsers_should_require_admin_token(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestMerchantCRUD_should_create_update_delete(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	token := adminToken(t, srv)

	createBody := bytes.NewBufferString(`{"email":"crudshop@ecomerce.local","display_name":"CRUD Shop","password":"Shop@123456"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/merchants", createBody)
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createRec := httptest.NewRecorder()
	srv.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createRec.Code, createRec.Body.String())
	}
	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	updateBody := bytes.NewBufferString(`{"email":"crudshop2@ecomerce.local","display_name":"CRUD Shop 2","password":""}`)
	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/merchants/"+created.Data.ID, updateBody)
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+token)
	updateRec := httptest.NewRecorder()
	srv.ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", updateRec.Code, updateRec.Body.String())
	}

	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/merchants/"+created.Data.ID, nil)
	delReq.Header.Set("Authorization", "Bearer "+token)
	delRec := httptest.NewRecorder()
	srv.ServeHTTP(delRec, delReq)
	if delRec.Code != http.StatusNoContent {
		t.Fatalf("delete status=%d body=%s", delRec.Code, delRec.Body.String())
	}
}

func TestUserCRUD_should_create_update_delete(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	token := adminToken(t, srv)

	createBody := bytes.NewBufferString(`{"email":"cruduser@ecomerce.local","display_name":"CRUD User","password":"Buyer@123456"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/users", createBody)
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createRec := httptest.NewRecorder()
	srv.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createRec.Code, createRec.Body.String())
	}
	var created struct {
		Data struct {
			ID   string `json:"id"`
			Role string `json:"role"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Data.Role != "user" {
		t.Fatalf("expected role user, got %q", created.Data.Role)
	}

	updateBody := bytes.NewBufferString(`{"email":"cruduser2@ecomerce.local","display_name":"CRUD User 2","password":""}`)
	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/users/"+created.Data.ID, updateBody)
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+token)
	updateRec := httptest.NewRecorder()
	srv.ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", updateRec.Code, updateRec.Body.String())
	}

	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+created.Data.ID, nil)
	delReq.Header.Set("Authorization", "Bearer "+token)
	delRec := httptest.NewRecorder()
	srv.ServeHTTP(delRec, delReq)
	if delRec.Code != http.StatusNoContent {
		t.Fatalf("delete status=%d body=%s", delRec.Code, delRec.Body.String())
	}
}

func TestProductCRUD_should_require_merchant_and_admin(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)

	unauth := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewBufferString(`{}`))
	unauth.Header.Set("Content-Type", "application/json")
	unauthRec := httptest.NewRecorder()
	srv.ServeHTTP(unauthRec, unauth)
	if unauthRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for create, got %d", unauthRec.Code)
	}

	token := adminToken(t, srv)
	merchantsReq := httptest.NewRequest(http.MethodGet, "/api/v1/merchants", nil)
	merchantsReq.Header.Set("Authorization", "Bearer "+token)
	merchantsRec := httptest.NewRecorder()
	srv.ServeHTTP(merchantsRec, merchantsReq)
	if merchantsRec.Code != http.StatusOK {
		t.Fatalf("merchants status=%d", merchantsRec.Code)
	}
	var merchantsPayload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(merchantsRec.Body.Bytes(), &merchantsPayload); err != nil {
		t.Fatal(err)
	}
	if len(merchantsPayload.Data) == 0 {
		t.Fatal("expected seeded merchants")
	}
	merchantID := merchantsPayload.Data[0].ID

	createBody := bytes.NewBufferString(`{"merchant_id":"` + merchantID + `","name":"CRUD Product","description":"demo","price_cents":120000,"currency":"VND","stock":7}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/products", createBody)
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createRec := httptest.NewRecorder()
	srv.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createRec.Code, createRec.Body.String())
	}
	var created struct {
		Data struct {
			ID         string `json:"id"`
			MerchantID string `json:"merchant_id"`
			Name       string `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Data.MerchantID != merchantID || created.Data.Name != "CRUD Product" {
		t.Fatalf("unexpected create: %+v", created.Data)
	}

	updateBody := bytes.NewBufferString(`{"merchant_id":"` + merchantID + `","name":"CRUD Product 2","description":"upd","price_cents":150000,"currency":"VND","stock":3}`)
	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/products/"+created.Data.ID, updateBody)
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+token)
	updateRec := httptest.NewRecorder()
	srv.ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", updateRec.Code, updateRec.Body.String())
	}

	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/products/"+created.Data.ID, nil)
	delReq.Header.Set("Authorization", "Bearer "+token)
	delRec := httptest.NewRecorder()
	srv.ServeHTTP(delRec, delReq)
	if delRec.Code != http.StatusNoContent {
		t.Fatalf("delete status=%d body=%s", delRec.Code, delRec.Body.String())
	}
}
