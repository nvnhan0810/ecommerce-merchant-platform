package presentation

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/application/commands"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/application/queries"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
	identitydomain "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/authctx"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/storage"
)

type CatalogHandler struct {
	list               *queries.ListProductsHandler
	get                *queries.GetProductHandler
	create             *commands.CreateProductHandler
	update             *commands.UpdateProductHandler
	delete             *commands.DeleteProductHandler
	uploadImage        *commands.UploadProductImageHandler
	deleteImage        *commands.DeleteProductImageHandler
	listCategories     *queries.ListCategoriesHandler
	getCategory        *queries.GetCategoryHandler
	createCategory     *commands.CreateCategoryHandler
	updateCategory     *commands.UpdateCategoryHandler
	updateCategoryStat *commands.UpdateCategoryStatusHandler
	deleteCategory     *commands.DeleteCategoryHandler
	removeProductCat   *commands.RemoveProductCategoryHandler
	store              storage.ObjectStore
}

func NewCatalogHandler(
	list *queries.ListProductsHandler,
	get *queries.GetProductHandler,
	create *commands.CreateProductHandler,
	update *commands.UpdateProductHandler,
	delete *commands.DeleteProductHandler,
	uploadImage *commands.UploadProductImageHandler,
	deleteImage *commands.DeleteProductImageHandler,
	listCategories *queries.ListCategoriesHandler,
	getCategory *queries.GetCategoryHandler,
	createCategory *commands.CreateCategoryHandler,
	updateCategory *commands.UpdateCategoryHandler,
	updateCategoryStat *commands.UpdateCategoryStatusHandler,
	deleteCategory *commands.DeleteCategoryHandler,
	removeProductCat *commands.RemoveProductCategoryHandler,
	store storage.ObjectStore,
) *CatalogHandler {
	return &CatalogHandler{
		list: list, get: get, create: create, update: update, delete: delete,
		uploadImage: uploadImage, deleteImage: deleteImage,
		listCategories: listCategories, getCategory: getCategory,
		createCategory: createCategory, updateCategory: updateCategory,
		updateCategoryStat: updateCategoryStat, deleteCategory: deleteCategory,
		removeProductCat: removeProductCat,
		store: store,
	}
}

func isAdminRequest(r *http.Request) bool {
	claims, ok := authctx.FromContext(r.Context())
	return ok && claims.Role == identitydomain.RoleAdmin
}

func (h *CatalogHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, err := h.list.Handle(r.Context(), queries.ListProductsQuery{
		Limit:                limit,
		Offset:               offset,
		MerchantID:           strings.TrimSpace(r.URL.Query().Get("merchant_id")),
		CategoryID:           strings.TrimSpace(r.URL.Query().Get("category_id")),
		IncludeAllCategories: isAdminRequest(r),
	})
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

type productBody struct {
	MerchantID   string    `json:"merchant_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	PriceCents   int64     `json:"price_cents"`
	Currency     string    `json:"currency"`
	Stock        int       `json:"stock"`
	CategoryIDs  *[]string `json:"category_ids"`
}

func categoryIDsFromBody(ids *[]string) ([]string, bool) {
	if ids == nil {
		return nil, false
	}
	return *ids, true
}

func (h *CatalogHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var body productBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	cats, _ := categoryIDsFromBody(body.CategoryIDs)
	res, err := h.create.Handle(r.Context(), commands.CreateProductCommand{
		MerchantID:  body.MerchantID,
		Name:        body.Name,
		Description: body.Description,
		PriceCents:  body.PriceCents,
		Currency:    body.Currency,
		Stock:       body.Stock,
		CategoryIDs: cats,
	})
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": res})
}

func (h *CatalogHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id, err := domain.ParseProductID(chi.URLParam(r, "id"))
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	item, err := h.get.Handle(r.Context(), queries.GetProductQuery{
		ID:                   id,
		IncludeAllCategories: isAdminRequest(r),
	})
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *CatalogHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := domain.ParseProductID(chi.URLParam(r, "id"))
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	var body productBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	cats, setCats := categoryIDsFromBody(body.CategoryIDs)
	res, err := h.update.Handle(r.Context(), commands.UpdateProductCommand{
		ID:            id,
		MerchantID:    body.MerchantID,
		Name:          body.Name,
		Description:   body.Description,
		PriceCents:    body.PriceCents,
		Currency:      body.Currency,
		Stock:         body.Stock,
		CategoryIDs:   cats,
		SetCategories: setCats,
	})
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": res})
}

func (h *CatalogHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := domain.ParseProductID(chi.URLParam(r, "id"))
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	if err := h.delete.Handle(r.Context(), commands.DeleteProductCommand{ID: id}); err != nil {
		writeCatalogError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type merchantProductBody struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PriceCents  int64     `json:"price_cents"`
	Currency    string    `json:"currency"`
	Stock       int       `json:"stock"`
	CategoryIDs *[]string `json:"category_ids"`
}

func merchantIDFromClaims(r *http.Request) (string, bool) {
	claims, ok := authctx.FromContext(r.Context())
	if !ok || claims.UserID == "" {
		return "", false
	}
	return string(claims.UserID), true
}

func (h *CatalogHandler) ListMerchantProducts(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := merchantIDFromClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, err := h.list.Handle(r.Context(), queries.ListProductsQuery{
		Limit:                limit,
		Offset:               offset,
		MerchantID:           merchantID,
		IncludeOrderFlags:    true,
		IncludeAllCategories: true,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *CatalogHandler) CreateMerchantProduct(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := merchantIDFromClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body merchantProductBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	cats, _ := categoryIDsFromBody(body.CategoryIDs)
	res, err := h.create.Handle(r.Context(), commands.CreateProductCommand{
		MerchantID:      merchantID,
		OwnerMerchantID: merchantID,
		Name:            body.Name,
		Description:     body.Description,
		PriceCents:      body.PriceCents,
		Currency:        body.Currency,
		Stock:           body.Stock,
		CategoryIDs:     cats,
	})
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": res})
}

func (h *CatalogHandler) GetMerchantProduct(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := merchantIDFromClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := domain.ParseProductID(chi.URLParam(r, "id"))
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	item, err := h.get.Handle(r.Context(), queries.GetProductQuery{
		ID:                   id,
		OwnerMerchantID:      merchantID,
		IncludeOrderFlags:    true,
		IncludeAllCategories: true,
	})
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *CatalogHandler) UpdateMerchantProduct(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := merchantIDFromClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := domain.ParseProductID(chi.URLParam(r, "id"))
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	var body merchantProductBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	cats, setCats := categoryIDsFromBody(body.CategoryIDs)
	res, err := h.update.Handle(r.Context(), commands.UpdateProductCommand{
		ID:              id,
		OwnerMerchantID: merchantID,
		Name:            body.Name,
		Description:     body.Description,
		PriceCents:      body.PriceCents,
		Currency:        body.Currency,
		Stock:           body.Stock,
		CategoryIDs:     cats,
		SetCategories:   setCats,
	})
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": res})
}

func (h *CatalogHandler) DeleteMerchantProduct(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := merchantIDFromClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := domain.ParseProductID(chi.URLParam(r, "id"))
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	if err := h.delete.Handle(r.Context(), commands.DeleteProductCommand{
		ID:              id,
		OwnerMerchantID: merchantID,
	}); err != nil {
		writeCatalogError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CatalogHandler) UploadMerchantProductImage(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := merchantIDFromClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := domain.ParseProductID(chi.URLParam(r, "id"))
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	if err := r.ParseMultipartForm(6 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	limited := http.MaxBytesReader(w, file, 5<<20)
	data, err := io.ReadAll(limited)
	if err != nil {
		writeError(w, http.StatusBadRequest, "file too large or unreadable (max 5MB)")
		return
	}
	if len(data) == 0 {
		writeCatalogError(w, domain.ErrInvalidImage)
		return
	}
	ct := http.DetectContentType(data)

	res, err := h.uploadImage.Handle(r.Context(), commands.UploadProductImageCommand{
		ID:              id,
		OwnerMerchantID: merchantID,
		Filename:        header.Filename,
		ContentType:     ct,
		Size:            int64(len(data)),
		Body:            bytes.NewReader(data),
	})
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": res})
}

func (h *CatalogHandler) DeleteMerchantProductImage(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := merchantIDFromClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := domain.ParseProductID(chi.URLParam(r, "id"))
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	res, err := h.deleteImage.Handle(r.Context(), commands.DeleteProductImageCommand{
		ID:              id,
		OwnerMerchantID: merchantID,
	})
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": res})
}

func (h *CatalogHandler) UploadProductImage(w http.ResponseWriter, r *http.Request) {
	id, err := domain.ParseProductID(chi.URLParam(r, "id"))
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	if err := r.ParseMultipartForm(6 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	limited := http.MaxBytesReader(w, file, 5<<20)
	data, err := io.ReadAll(limited)
	if err != nil {
		writeError(w, http.StatusBadRequest, "file too large or unreadable (max 5MB)")
		return
	}
	if len(data) == 0 {
		writeCatalogError(w, domain.ErrInvalidImage)
		return
	}
	ct := http.DetectContentType(data)

	res, err := h.uploadImage.Handle(r.Context(), commands.UploadProductImageCommand{
		ID:          id,
		Filename:    header.Filename,
		ContentType: ct,
		Size:        int64(len(data)),
		Body:        bytes.NewReader(data),
	})
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": res})
}

func (h *CatalogHandler) DeleteProductImage(w http.ResponseWriter, r *http.Request) {
	id, err := domain.ParseProductID(chi.URLParam(r, "id"))
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	res, err := h.deleteImage.Handle(r.Context(), commands.DeleteProductImageCommand{ID: id})
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": res})
}

// --- Categories ---

func (h *CatalogHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	q := queries.ListCategoriesQuery{}
	if isAdminRequest(r) {
		q.Status = strings.TrimSpace(r.URL.Query().Get("status"))
	} else {
		q.ApprovedOnly = true
	}
	items, err := h.listCategories.Handle(r.Context(), q)
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *CatalogHandler) ListMerchantCategories(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := merchantIDFromClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	items, err := h.listCategories.Handle(r.Context(), queries.ListCategoriesQuery{
		MerchantAssignable: true,
		MerchantViewerID:   merchantID,
	})
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *CatalogHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	id, err := domain.ParseCategoryID(chi.URLParam(r, "id"))
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	item, err := h.getCategory.Handle(r.Context(), queries.GetCategoryQuery{ID: id})
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

type categoryBody struct {
	Name string `json:"name"`
}

func (h *CatalogHandler) CreateAdminCategory(w http.ResponseWriter, r *http.Request) {
	var body categoryBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	res, err := h.createCategory.Handle(r.Context(), commands.CreateCategoryCommand{
		Name:    body.Name,
		AsAdmin: true,
	})
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": res})
}

func (h *CatalogHandler) CreateMerchantCategory(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := merchantIDFromClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body categoryBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	res, err := h.createCategory.Handle(r.Context(), commands.CreateCategoryCommand{
		Name:                body.Name,
		CreatedByMerchantID: merchantID,
	})
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": res})
}

func (h *CatalogHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := domain.ParseCategoryID(chi.URLParam(r, "id"))
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	var body categoryBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	res, err := h.updateCategory.Handle(r.Context(), commands.UpdateCategoryCommand{
		ID:   id,
		Name: body.Name,
	})
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": res})
}

type categoryStatusBody struct {
	Status string `json:"status"`
}

func (h *CatalogHandler) UpdateCategoryStatus(w http.ResponseWriter, r *http.Request) {
	id, err := domain.ParseCategoryID(chi.URLParam(r, "id"))
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	var body categoryStatusBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	status, err := domain.ParseCategoryStatus(body.Status)
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	res, err := h.updateCategoryStat.Handle(r.Context(), commands.UpdateCategoryStatusCommand{
		ID:     id,
		Status: status,
	})
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": res})
}

func (h *CatalogHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := domain.ParseCategoryID(chi.URLParam(r, "id"))
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	if err := h.deleteCategory.Handle(r.Context(), commands.DeleteCategoryCommand{ID: id}); err != nil {
		writeCatalogError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CatalogHandler) RemoveProductCategory(w http.ResponseWriter, r *http.Request) {
	productID, err := domain.ParseProductID(chi.URLParam(r, "id"))
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	categoryID, err := domain.ParseCategoryID(chi.URLParam(r, "categoryId"))
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	if err := h.removeProductCat.Handle(r.Context(), commands.RemoveProductCategoryCommand{
		ProductID:  productID,
		CategoryID: categoryID,
	}); err != nil {
		writeCatalogError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CatalogHandler) RemoveMerchantProductCategory(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := merchantIDFromClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	productID, err := domain.ParseProductID(chi.URLParam(r, "id"))
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	categoryID, err := domain.ParseCategoryID(chi.URLParam(r, "categoryId"))
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	if err := h.removeProductCat.Handle(r.Context(), commands.RemoveProductCategoryCommand{
		ProductID:       productID,
		CategoryID:      categoryID,
		OwnerMerchantID: merchantID,
	}); err != nil {
		writeCatalogError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CatalogHandler) GetAdminProduct(w http.ResponseWriter, r *http.Request) {
	id, err := domain.ParseProductID(chi.URLParam(r, "id"))
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	item, err := h.get.Handle(r.Context(), queries.GetProductQuery{
		ID:                   id,
		IncludeAllCategories: true,
	})
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *CatalogHandler) ListAdminProducts(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, err := h.list.Handle(r.Context(), queries.ListProductsQuery{
		Limit:                limit,
		Offset:               offset,
		MerchantID:           strings.TrimSpace(r.URL.Query().Get("merchant_id")),
		CategoryID:           strings.TrimSpace(r.URL.Query().Get("category_id")),
		IncludeAllCategories: true,
	})
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *CatalogHandler) ServeMedia(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(chi.URLParam(r, "*"), "/")
	if key == "" {
		writeError(w, http.StatusBadRequest, "object key required")
		return
	}
	if h.store == nil || !h.store.Enabled() {
		writeError(w, http.StatusServiceUnavailable, storage.ErrObjectStoreDisabled.Error())
		return
	}
	body, contentType, err := h.store.Download(r.Context(), key)
	if err != nil {
		writeError(w, http.StatusNotFound, "object not found")
		return
	}
	defer body.Close()
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, body)
}

func writeCatalogError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, domain.ErrProductNotFound),
		errors.Is(err, domain.ErrMerchantNotFound),
		errors.Is(err, domain.ErrCategoryNotFound):
		status = http.StatusNotFound
	case errors.Is(err, domain.ErrProductHasOrders),
		errors.Is(err, domain.ErrCategoryExists):
		status = http.StatusConflict
	case errors.Is(err, domain.ErrInvalidProductName),
		errors.Is(err, domain.ErrInvalidProductPrice),
		errors.Is(err, domain.ErrMerchantRequired),
		errors.Is(err, domain.ErrInvalidProductID),
		errors.Is(err, domain.ErrInvalidImage),
		errors.Is(err, domain.ErrInvalidCategoryName),
		errors.Is(err, domain.ErrInvalidCategoryID),
		errors.Is(err, domain.ErrInvalidCategoryStatus),
		errors.Is(err, domain.ErrCategoryNotAssignable):
		status = http.StatusBadRequest
	case errors.Is(err, storage.ErrObjectStoreDisabled):
		status = http.StatusServiceUnavailable
	}
	writeError(w, status, err.Error())
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
