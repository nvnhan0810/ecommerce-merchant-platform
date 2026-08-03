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
	"github.com/nvnhan0810/ecomerce-api/internal/platform/storage"
)

type CatalogHandler struct {
	list        *queries.ListProductsHandler
	get         *queries.GetProductHandler
	create      *commands.CreateProductHandler
	update      *commands.UpdateProductHandler
	delete      *commands.DeleteProductHandler
	uploadImage *commands.UploadProductImageHandler
	deleteImage *commands.DeleteProductImageHandler
	store       storage.ObjectStore
}

func NewCatalogHandler(
	list *queries.ListProductsHandler,
	get *queries.GetProductHandler,
	create *commands.CreateProductHandler,
	update *commands.UpdateProductHandler,
	delete *commands.DeleteProductHandler,
	uploadImage *commands.UploadProductImageHandler,
	deleteImage *commands.DeleteProductImageHandler,
	store storage.ObjectStore,
) *CatalogHandler {
	return &CatalogHandler{
		list: list, get: get, create: create, update: update, delete: delete,
		uploadImage: uploadImage, deleteImage: deleteImage, store: store,
	}
}

func (h *CatalogHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, err := h.list.Handle(r.Context(), queries.ListProductsQuery{Limit: limit, Offset: offset})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

type productBody struct {
	MerchantID  string `json:"merchant_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents"`
	Currency    string `json:"currency"`
	Stock       int    `json:"stock"`
}

func (h *CatalogHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var body productBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	res, err := h.create.Handle(r.Context(), commands.CreateProductCommand{
		MerchantID:  body.MerchantID,
		Name:        body.Name,
		Description: body.Description,
		PriceCents:  body.PriceCents,
		Currency:    body.Currency,
		Stock:       body.Stock,
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
	item, err := h.get.Handle(r.Context(), queries.GetProductQuery{ID: id})
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
	res, err := h.update.Handle(r.Context(), commands.UpdateProductCommand{
		ID:          id,
		MerchantID:  body.MerchantID,
		Name:        body.Name,
		Description: body.Description,
		PriceCents:  body.PriceCents,
		Currency:    body.Currency,
		Stock:       body.Stock,
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
	res, err := h.deleteImage.Handle(r.Context(), id)
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": res})
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
	case errors.Is(err, domain.ErrProductNotFound), errors.Is(err, domain.ErrMerchantNotFound):
		status = http.StatusNotFound
	case errors.Is(err, domain.ErrInvalidProductName),
		errors.Is(err, domain.ErrInvalidProductPrice),
		errors.Is(err, domain.ErrMerchantRequired),
		errors.Is(err, domain.ErrInvalidProductID),
		errors.Is(err, domain.ErrInvalidImage):
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
