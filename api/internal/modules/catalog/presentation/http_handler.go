package presentation

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/application/commands"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/application/queries"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
)

type CatalogHandler struct {
	list   *queries.ListProductsHandler
	create *commands.CreateProductHandler
}

func NewCatalogHandler(
	list *queries.ListProductsHandler,
	create *commands.CreateProductHandler,
) *CatalogHandler {
	return &CatalogHandler{list: list, create: create}
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

type createProductRequest struct {
	MerchantID  string `json:"merchant_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents"`
	Currency    string `json:"currency"`
	Stock       int    `json:"stock"`
}

func (h *CatalogHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var body createProductRequest
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
		status := http.StatusBadRequest
		if !errors.Is(err, domain.ErrInvalidProductName) && !errors.Is(err, domain.ErrInvalidProductPrice) {
			status = http.StatusInternalServerError
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, res)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
