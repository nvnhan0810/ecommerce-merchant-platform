package presentation

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	catalogdomain "github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/application/commands"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/application/queries"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/authctx"
)

type OrderingHandler struct {
	list           *queries.ListOrdersHandler
	get            *queries.GetOrderHandler
	updateStatus   *commands.UpdateOrderStatusHandler
	create         *commands.CreateOrderHandler
	applyDelivery  *commands.ApplyDeliveryEventHandler
	webhookSecret  string
}

func NewOrderingHandler(
	list *queries.ListOrdersHandler,
	get *queries.GetOrderHandler,
	updateStatus *commands.UpdateOrderStatusHandler,
	create *commands.CreateOrderHandler,
	applyDelivery *commands.ApplyDeliveryEventHandler,
	webhookSecret string,
) *OrderingHandler {
	return &OrderingHandler{
		list:          list,
		get:           get,
		updateStatus:  updateStatus,
		create:        create,
		applyDelivery: applyDelivery,
		webhookSecret: strings.TrimSpace(webhookSecret),
	}
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

type updateStatusBody struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
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
		Reason:          strings.TrimSpace(body.Reason),
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
	Note            string `json:"note"`
	ShippingName    string `json:"shipping_name"`
	ShippingPhone   string `json:"shipping_phone"`
	ShippingAddress string `json:"shipping_address"`
	Items           []struct {
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
		UserID:          userID,
		Note:            body.Note,
		ShippingName:    body.ShippingName,
		ShippingPhone:   body.ShippingPhone,
		ShippingAddress: body.ShippingAddress,
		Items:           items,
		Actor:           actorFromRequest(r),
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

type deliveryEventBody struct {
	OrderCode            string `json:"order_code"`
	DeliveryTrackingCode string `json:"delivery_tracking_code"`
	DeliveryCarrier      string `json:"delivery_carrier"`
	Status               string `json:"status"`
	Message              string `json:"message"`
	Reason               string `json:"reason"`
	OccurredAt           string `json:"occurred_at"`
	EventID              string `json:"event_id"`
}

func (h *OrderingHandler) DeliveryWebhook(w http.ResponseWriter, r *http.Request) {
	if h.webhookSecret == "" || r.Header.Get("X-Webhook-Secret") != h.webhookSecret {
		writeError(w, http.StatusUnauthorized, "invalid webhook secret")
		return
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	var body deliveryEventBody
	if err := json.Unmarshal(raw, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	occurredAt, err := parseOccurredAt(body.OccurredAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid occurred_at")
		return
	}
	item, err := h.applyDelivery.Handle(r.Context(), commands.ApplyDeliveryEventCommand{
		OrderCode:            body.OrderCode,
		DeliveryTrackingCode: body.DeliveryTrackingCode,
		DeliveryCarrier:      body.DeliveryCarrier,
		Status:               body.Status,
		Message:              body.Message,
		Reason:               body.Reason,
		OccurredAt:           occurredAt,
		EventID:              body.EventID,
		Source:               "webhook",
		RawPayload:           json.RawMessage(raw),
	})
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *OrderingHandler) SimulateDelivery(w http.ResponseWriter, r *http.Request) {
	id, err := domain.ParseOrderID(chi.URLParam(r, "id"))
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	var body deliveryEventBody
	if err := json.Unmarshal(raw, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	occurredAt, err := parseOccurredAt(body.OccurredAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid occurred_at")
		return
	}
	item, err := h.applyDelivery.Handle(r.Context(), commands.ApplyDeliveryEventCommand{
		OrderID:              id,
		OrderCode:            body.OrderCode,
		DeliveryTrackingCode: body.DeliveryTrackingCode,
		DeliveryCarrier:      body.DeliveryCarrier,
		Status:               body.Status,
		Message:              body.Message,
		Reason:               body.Reason,
		OccurredAt:           occurredAt,
		EventID:              body.EventID,
		Source:               "simulate",
		RawPayload:           json.RawMessage(raw),
	})
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func parseOccurredAt(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Now().UTC(), nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
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
		errors.Is(err, domain.ErrInvalidStatusTransition),
		errors.Is(err, domain.ErrMerchantConfirmOnly),
		errors.Is(err, domain.ErrMerchantCancelReasonRequired),
		errors.Is(err, domain.ErrEmptyOrderItems),
		errors.Is(err, domain.ErrMissingShippingInfo),
		errors.Is(err, domain.ErrInvalidOrderQuantity),
		errors.Is(err, domain.ErrInvalidOrderPrice),
		errors.Is(err, domain.ErrUserRequired),
		errors.Is(err, domain.ErrInvalidDeliveryStatus),
		errors.Is(err, domain.ErrDeliveryLookupRequired),
		errors.Is(err, catalogdomain.ErrInvalidProductID),
		errors.Is(err, catalogdomain.ErrInvalidQuantity):
		status = http.StatusBadRequest
	case errors.Is(err, domain.ErrDeliveryEventDuplicate):
		status = http.StatusConflict
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
