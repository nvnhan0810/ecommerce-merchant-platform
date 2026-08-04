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
	identitydomain "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
	identityinfra "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/infrastructure"
	identitypres "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/presentation"
	orderingcommands "github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/application/commands"
	orderingqueries "github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/application/queries"
	orderinginfra "github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/infrastructure"
	orderingpres "github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/presentation"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/config"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/httpapi"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/seed"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/storage"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	productRepo := cataloginfra.NewInMemoryProductRepository()
	orderRepo := orderinginfra.NewInMemoryOrderRepository()
	users := identityinfra.NewInMemoryAccountRepository()
	merchants := identityinfra.NewInMemoryAccountRepository()
	admins := identityinfra.NewInMemoryAccountRepository()
	hasher := identityinfra.NewBcryptPasswordHasher()
	tokens, err := identityinfra.NewJWTTokenService("test-jwt-secret-16", 2*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if err := seed.RunDemo(users, merchants, admins, productRepo, orderRepo, hasher, "Admin@123456"); err != nil {
		t.Fatal(err)
	}
	checker := cataloginfra.NewAccountMerchantChecker(merchants)
	nop := storage.NopStore{}
	base := "https://ecomerce-api.nvnhan0810.com"
	return httpapi.NewRouter(httpapi.Dependencies{
		Config: config.Config{
			HTTPAddr:    ":0",
			Env:         "test",
			CORSOrigins: []string{"http://localhost:5173"},
		},
		Health: healthpres.NewHealthHandler("test"),
		Catalog: catalogpres.NewCatalogHandler(
			queries.NewListProductsHandler(productRepo, base),
			queries.NewGetProductHandler(productRepo, base),
			commands.NewCreateProductHandler(productRepo, checker, base),
			commands.NewUpdateProductHandler(productRepo, checker, base),
			commands.NewDeleteProductHandler(productRepo, nop),
			commands.NewUploadProductImageHandler(productRepo, nop, base),
			commands.NewDeleteProductImageHandler(productRepo, nop, base),
			nop,
		),
		Identity: identitypres.NewIdentityHandler(
			identityqueries.NewListUsersHandler(users),
			identityqueries.NewListMerchantsHandler(merchants),
			identityqueries.NewGetUserHandler(users),
			identityqueries.NewGetMerchantHandler(merchants),
			identitycommands.NewLoginHandler(admins, hasher, tokens, identitydomain.RoleAdmin),
			identitycommands.NewLoginHandler(merchants, hasher, tokens, identitydomain.RoleMerchant),
			identityqueries.NewGetCurrentUserHandler(admins, identitydomain.RoleAdmin),
			identityqueries.NewGetCurrentUserHandler(merchants, identitydomain.RoleMerchant),
			identitycommands.NewCreateUserHandler(users, hasher),
			identitycommands.NewUpdateUserHandler(users, hasher),
			identitycommands.NewDeleteUserHandler(users),
			identitycommands.NewCreateMerchantHandler(merchants, hasher),
			identitycommands.NewUpdateMerchantHandler(merchants, hasher),
			identitycommands.NewDeleteMerchantHandler(merchants),
		),
		Ordering: orderingpres.NewOrderingHandler(
			orderingqueries.NewListOrdersHandler(orderRepo),
			orderingqueries.NewGetOrderHandler(orderRepo),
			orderingcommands.NewUpdateOrderStatusHandler(orderRepo),
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

func merchantToken(t *testing.T, srv http.Handler) string {
	t.Helper()
	body := bytes.NewBufferString(`{"email":"shop@ecomerce.local","password":"Shop@123456"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/merchant/login", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("merchant login status=%d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		AccessToken string `json:"access_token"`
		User        struct {
			Role string `json:"role"`
		} `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.User.Role != "merchant" {
		t.Fatalf("role=%s want merchant", payload.User.Role)
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

func TestOrders_admin_list_get_update_status(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)

	unauth := httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	unauthRec := httptest.NewRecorder()
	srv.ServeHTTP(unauthRec, unauth)
	if unauthRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", unauthRec.Code)
	}

	token := adminToken(t, srv)
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/orders?limit=50", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listRec := httptest.NewRecorder()
	srv.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listRec.Code, listRec.Body.String())
	}
	var listPayload struct {
		Data []struct {
			ID     string `json:"id"`
			Code   string `json:"code"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listPayload); err != nil {
		t.Fatal(err)
	}
	if len(listPayload.Data) != 7 {
		t.Fatalf("orders=%d want 7", len(listPayload.Data))
	}
	first := listPayload.Data[0]

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/orders/"+first.ID, nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getRec := httptest.NewRecorder()
	srv.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", getRec.Code, getRec.Body.String())
	}
	var getPayload struct {
		Data struct {
			History []struct {
				EventType  string `json:"event_type"`
				ActorRole  string `json:"actor_role"`
				ActorEmail string `json:"actor_email"`
			} `json:"history"`
		} `json:"data"`
	}
	if err := json.Unmarshal(getRec.Body.Bytes(), &getPayload); err != nil {
		t.Fatal(err)
	}
	if len(getPayload.Data.History) == 0 {
		t.Fatal("expected order history")
	}
	if getPayload.Data.History[0].EventType != "created" {
		t.Fatalf("first event=%s", getPayload.Data.History[0].EventType)
	}

	codeReq := httptest.NewRequest(http.MethodGet, "/api/v1/orders?code="+first.Code, nil)
	codeReq.Header.Set("Authorization", "Bearer "+token)
	codeRec := httptest.NewRecorder()
	srv.ServeHTTP(codeRec, codeReq)
	if codeRec.Code != http.StatusOK {
		t.Fatalf("code search status=%d body=%s", codeRec.Code, codeRec.Body.String())
	}

	patchBody := bytes.NewBufferString(`{"status":"confirmed"}`)
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/orders/"+first.ID+"/status", patchBody)
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.Header.Set("Authorization", "Bearer "+token)
	patchRec := httptest.NewRecorder()
	srv.ServeHTTP(patchRec, patchReq)
	if patchRec.Code != http.StatusOK {
		t.Fatalf("patch status=%d body=%s", patchRec.Code, patchRec.Body.String())
	}
	var patched struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(patchRec.Body.Bytes(), &patched); err != nil {
		t.Fatal(err)
	}
	if patched.Data.Status != "confirmed" {
		t.Fatalf("status=%s want confirmed", patched.Data.Status)
	}
}

func TestMerchantLogin_should_return_token_and_me(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)

	adminTok := adminToken(t, srv)
	forbid := httptest.NewRequest(http.MethodGet, "/api/v1/auth/merchant/me", nil)
	forbid.Header.Set("Authorization", "Bearer "+adminTok)
	forbidRec := httptest.NewRecorder()
	srv.ServeHTTP(forbidRec, forbid)
	if forbidRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for admin on merchant me, got %d", forbidRec.Code)
	}

	token := merchantToken(t, srv)
	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/merchant/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+token)
	meRec := httptest.NewRecorder()
	srv.ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("me status=%d body=%s", meRec.Code, meRec.Body.String())
	}
	var me struct {
		Data struct {
			Email string `json:"email"`
			Role  string `json:"role"`
		} `json:"data"`
	}
	if err := json.Unmarshal(meRec.Body.Bytes(), &me); err != nil {
		t.Fatal(err)
	}
	if me.Data.Role != "merchant" || me.Data.Email != "shop@ecomerce.local" {
		t.Fatalf("unexpected me: %+v", me.Data)
	}

	adminMe := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	adminMe.Header.Set("Authorization", "Bearer "+token)
	adminMeRec := httptest.NewRecorder()
	srv.ServeHTTP(adminMeRec, adminMe)
	if adminMeRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for merchant on admin me, got %d", adminMeRec.Code)
	}
}

func TestMerchantProductCRUD_should_scope_and_block_delete_with_orders(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)

	unauth := httptest.NewRequest(http.MethodGet, "/api/v1/merchant/products", nil)
	unauthRec := httptest.NewRecorder()
	srv.ServeHTTP(unauthRec, unauth)
	if unauthRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", unauthRec.Code)
	}

	adminTok := adminToken(t, srv)
	forbid := httptest.NewRequest(http.MethodGet, "/api/v1/merchant/products", nil)
	forbid.Header.Set("Authorization", "Bearer "+adminTok)
	forbidRec := httptest.NewRecorder()
	srv.ServeHTTP(forbidRec, forbid)
	if forbidRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for admin, got %d", forbidRec.Code)
	}

	token := merchantToken(t, srv)
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/merchant/products?limit=100", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listRec := httptest.NewRecorder()
	srv.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listRec.Code, listRec.Body.String())
	}
	var listPayload struct {
		Data []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			CanDelete *bool  `json:"can_delete"`
			HasOrders *bool  `json:"has_orders"`
		} `json:"data"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listPayload); err != nil {
		t.Fatal(err)
	}
	if len(listPayload.Data) == 0 {
		t.Fatal("expected merchant products")
	}

	var orderedID string
	for _, p := range listPayload.Data {
		if p.CanDelete == nil || p.HasOrders == nil {
			t.Fatalf("expected order flags on product %+v", p)
		}
		if *p.HasOrders && !*p.CanDelete && orderedID == "" {
			orderedID = p.ID
		}
	}
	if orderedID == "" {
		t.Fatal("expected at least one product with orders")
	}

	blocked := httptest.NewRequest(http.MethodDelete, "/api/v1/merchant/products/"+orderedID, nil)
	blocked.Header.Set("Authorization", "Bearer "+token)
	blockedRec := httptest.NewRecorder()
	srv.ServeHTTP(blockedRec, blocked)
	if blockedRec.Code != http.StatusConflict {
		t.Fatalf("expected 409 delete ordered product, got %d body=%s", blockedRec.Code, blockedRec.Body.String())
	}

	createBody := bytes.NewBufferString(`{"name":"Merchant CRUD SKU","description":"own","price_cents":99000,"currency":"VND","stock":4}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/merchant/products", createBody)
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
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Data.Name != "Merchant CRUD SKU" {
		t.Fatalf("unexpected create: %+v", created.Data)
	}

	updateBody := bytes.NewBufferString(`{"name":"Merchant CRUD SKU 2","description":"upd","price_cents":110000,"currency":"VND","stock":2}`)
	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/merchant/products/"+created.Data.ID, updateBody)
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+token)
	updateRec := httptest.NewRecorder()
	srv.ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", updateRec.Code, updateRec.Body.String())
	}

	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/merchant/products/"+created.Data.ID, nil)
	delReq.Header.Set("Authorization", "Bearer "+token)
	delRec := httptest.NewRecorder()
	srv.ServeHTTP(delRec, delReq)
	if delRec.Code != http.StatusNoContent {
		t.Fatalf("delete status=%d body=%s", delRec.Code, delRec.Body.String())
	}
}

func TestMerchantOrders_should_list_get_and_update_own_only(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)

	unauth := httptest.NewRequest(http.MethodGet, "/api/v1/merchant/orders", nil)
	unauthRec := httptest.NewRecorder()
	srv.ServeHTTP(unauthRec, unauth)
	if unauthRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", unauthRec.Code)
	}

	adminTok := adminToken(t, srv)
	forbid := httptest.NewRequest(http.MethodGet, "/api/v1/merchant/orders", nil)
	forbid.Header.Set("Authorization", "Bearer "+adminTok)
	forbidRec := httptest.NewRecorder()
	srv.ServeHTTP(forbidRec, forbid)
	if forbidRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for admin, got %d", forbidRec.Code)
	}

	token := merchantToken(t, srv)
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/merchant/orders?limit=50", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listRec := httptest.NewRecorder()
	srv.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listRec.Code, listRec.Body.String())
	}
	var listPayload struct {
		Data []struct {
			ID         string `json:"id"`
			Code       string `json:"code"`
			MerchantID string `json:"merchant_id"`
			Status     string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listPayload); err != nil {
		t.Fatal(err)
	}
	if len(listPayload.Data) == 0 {
		t.Fatal("expected merchant orders")
	}
	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/merchant/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+token)
	meRec := httptest.NewRecorder()
	srv.ServeHTTP(meRec, meReq)
	var me struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(meRec.Body.Bytes(), &me); err != nil {
		t.Fatal(err)
	}
	for _, o := range listPayload.Data {
		if o.MerchantID != me.Data.ID {
			t.Fatalf("order %s merchant=%s want %s", o.ID, o.MerchantID, me.Data.ID)
		}
	}

	first := listPayload.Data[0]
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/merchant/orders/"+first.ID, nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getRec := httptest.NewRecorder()
	srv.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", getRec.Code, getRec.Body.String())
	}

	adminOrders := httptest.NewRequest(http.MethodGet, "/api/v1/orders?limit=50", nil)
	adminOrders.Header.Set("Authorization", "Bearer "+adminTok)
	adminOrdersRec := httptest.NewRecorder()
	srv.ServeHTTP(adminOrdersRec, adminOrders)
	var allOrders struct {
		Data []struct {
			ID         string `json:"id"`
			MerchantID string `json:"merchant_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminOrdersRec.Body.Bytes(), &allOrders); err != nil {
		t.Fatal(err)
	}
	foreignID := ""
	for _, o := range allOrders.Data {
		if o.MerchantID != me.Data.ID {
			foreignID = o.ID
			break
		}
	}
	if foreignID == "" {
		t.Fatal("expected another merchant order in seed")
	}
	foreignGet := httptest.NewRequest(http.MethodGet, "/api/v1/merchant/orders/"+foreignID, nil)
	foreignGet.Header.Set("Authorization", "Bearer "+token)
	foreignGetRec := httptest.NewRecorder()
	srv.ServeHTTP(foreignGetRec, foreignGet)
	if foreignGetRec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for foreign order, got %d", foreignGetRec.Code)
	}

	patchBody := bytes.NewBufferString(`{"status":"confirmed"}`)
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/merchant/orders/"+first.ID+"/status", patchBody)
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.Header.Set("Authorization", "Bearer "+token)
	patchRec := httptest.NewRecorder()
	srv.ServeHTTP(patchRec, patchReq)
	if patchRec.Code != http.StatusOK {
		t.Fatalf("patch status=%d body=%s", patchRec.Code, patchRec.Body.String())
	}
	var patched struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(patchRec.Body.Bytes(), &patched); err != nil {
		t.Fatal(err)
	}
	if patched.Data.Status != "confirmed" {
		t.Fatalf("status=%s want confirmed", patched.Data.Status)
	}
}
