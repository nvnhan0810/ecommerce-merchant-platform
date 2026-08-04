package presentation

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	catalogdomain "github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/application/commands"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/application/queries"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/authctx"
)

type OrderingHandler struct {
	list         *queries.ListOrdersHandler
	get          *queries.GetOrderHandler
	updateStatus *commands.UpdateOrderStatusHandler
	create       *commands.CreateOrderHandler
}

func NewOrderingHandler(
	list *queries.ListOrdersHandler,
	get *queries.GetOrderHandler,
	updateStatus *commands.UpdateOrderStatusHandler,
	create *commands.CreateOrderHandler,
) *OrderingHandler {
	return &OrderingHandler{list: list, get: get, updateStatus: updateStatus, create: create}
}

func merchantIDFromClaims(r *http.Request) (string, bool) {
	claims, ok := authctx.FromContext(r.Context())
	if !ok || claims.UserID == "" {
		return "", false
	}
	return string(claims.UserID), true
}

func userIDFromClaims(r *http.Request) (string, bool) {
	return merchantIDFromClaims(r)
}

func actorFromRequest(r *http.Request) domain.Actor {
	actor := domain.SystemActor()
	if claims, ok := authctx.FromContext(r.Context()); ok {
		actor = domain.Actor{
			ID:          string(claims.UserID),
			Email:       claims.Email,
			Role:        string(claims.Role),
			DisplayName: claims.Email,
		}
	}
	return actor
}

func (h *OrderingHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, err := h.list.Handle(r.Context(), queries.ListOrdersQuery{
		Limit:  limit,
		Offset: offset,
		Code:   r.URL.Query().Get("code"),
		Status: r.URL.Query().Get("status"),
	})
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *OrderingHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id, err := domain.ParseOrderID(chi.URLParam(r, "id"))
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	item, err := h.get.Handle(r.Context(), queries.GetOrderQuery{ID: id})
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

type updateStatusBody struct {
	Status string `json:"status"`
}

func (h *OrderingHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id, err := domain.ParseOrderID(chi.URLParam(r, "id"))
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	var body updateStatusBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	item, err := h.updateStatus.Handle(r.Context(), commands.UpdateOrderStatusCommand{
		ID:     id,
		Status: strings.TrimSpace(body.Status),
		Actor:  actorFromRequest(r),
	})
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *OrderingHandler) ListMerchantOrders(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := merchantIDFromClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, err := h.list.Handle(r.Context(), queries.ListOrdersQuery{
		Limit:      limit,
		Offset:     offset,
		Code:       r.URL.Query().Get("code"),
		Status:     r.URL.Query().Get("status"),
		MerchantID: merchantID,
	})
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *OrderingHandler) GetMerchantOrder(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := merchantIDFromClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := domain.ParseOrderID(chi.URLParam(r, "id"))
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	item, err := h.get.Handle(r.Context(), queries.GetOrderQuery{
		ID:              id,
		OwnerMerchantID: merchantID,
	})
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *OrderingHandler) UpdateMerchantOrderStatus(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := merchantIDFromClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := domain.ParseOrderID(chi.URLParam(r, "id"))
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	var body updateStatusBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	item, err := h.updateStatus.Handle(r.Context(), commands.UpdateOrderStatusCommand{
		ID:              id,
		Status:          strings.TrimSpace(body.Status),
		Actor:           actorFromRequest(r),
		OwnerMerchantID: merchantID,
	})
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

type createOrderBody struct {
	Note  string `json:"note"`
	Items []struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	} `json:"items"`
}

func (h *OrderingHandler) CreateUserOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body createOrderBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	items := make([]commands.CreateOrderItemInput, 0, len(body.Items))
	for _, item := range body.Items {
		items = append(items, commands.CreateOrderItemInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}
	created, err := h.create.Handle(r.Context(), commands.CreateOrderCommand{
		UserID: userID,
		Note:   body.Note,
		Items:  items,
		Actor:  actorFromRequest(r),
	})
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": created})
}

func (h *OrderingHandler) ListUserOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, err := h.list.Handle(r.Context(), queries.ListOrdersQuery{
		Limit:  limit,
		Offset: offset,
		Code:   r.URL.Query().Get("code"),
		Status: r.URL.Query().Get("status"),
		UserID: userID,
	})
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *OrderingHandler) GetUserOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := domain.ParseOrderID(chi.URLParam(r, "id"))
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	item, err := h.get.Handle(r.Context(), queries.GetOrderQuery{
		ID:          id,
		OwnerUserID: userID,
	})
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func writeOrderingError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, domain.ErrOrderNotFound),
		errors.Is(err, catalogdomain.ErrProductNotFound):
		status = http.StatusNotFound
	case errors.Is(err, domain.ErrInvalidOrderID),
		errors.Is(err, domain.ErrInvalidOrderCode),
		errors.Is(err, domain.ErrInvalidOrderStatus),
		errors.Is(err, domain.ErrEmptyOrderItems),
		errors.Is(err, domain.ErrInvalidOrderQuantity),
		errors.Is(err, domain.ErrInvalidOrderPrice),
		errors.Is(err, domain.ErrUserRequired),
		errors.Is(err, catalogdomain.ErrInvalidProductID),
		errors.Is(err, catalogdomain.ErrInvalidQuantity):
		status = http.StatusBadRequest
	case errors.Is(err, catalogdomain.ErrInsufficientStock):
		status = http.StatusConflict
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
