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
	list                  *queries.ListOrdersHandler
	get                   *queries.GetOrderHandler
	updateStatus          *commands.UpdateOrderStatusHandler
	create                *commands.CreateOrderHandler
	repay                 *commands.RepayOrderHandler
	applyDelivery         *commands.ApplyDeliveryEventHandler
	getPaymentSettings    *queries.GetPaymentSettingsHandler
	updatePaymentSettings *commands.UpdatePaymentSettingsHandler
	getPaymentMethods     *queries.GetPublicPaymentMethodsHandler
	listPaymentCallbacks  *queries.ListPaymentCallbacksHandler
	getPaymentCallback    *queries.GetPaymentCallbackHandler
	onePayCallback        *commands.HandleOnePayCallbackHandler
	webhookSecret         string
	webBaseURL            string
}

func NewOrderingHandler(
	list *queries.ListOrdersHandler,
	get *queries.GetOrderHandler,
	updateStatus *commands.UpdateOrderStatusHandler,
	create *commands.CreateOrderHandler,
	repay *commands.RepayOrderHandler,
	applyDelivery *commands.ApplyDeliveryEventHandler,
	getPaymentSettings *queries.GetPaymentSettingsHandler,
	updatePaymentSettings *commands.UpdatePaymentSettingsHandler,
	getPaymentMethods *queries.GetPublicPaymentMethodsHandler,
	listPaymentCallbacks *queries.ListPaymentCallbacksHandler,
	getPaymentCallback *queries.GetPaymentCallbackHandler,
	onePayCallback *commands.HandleOnePayCallbackHandler,
	webhookSecret string,
	webBaseURL string,
) *OrderingHandler {
	return &OrderingHandler{
		list:                  list,
		get:                   get,
		updateStatus:          updateStatus,
		create:                create,
		repay:                 repay,
		applyDelivery:         applyDelivery,
		getPaymentSettings:    getPaymentSettings,
		updatePaymentSettings: updatePaymentSettings,
		getPaymentMethods:     getPaymentMethods,
		listPaymentCallbacks:  listPaymentCallbacks,
		getPaymentCallback:    getPaymentCallback,
		onePayCallback:        onePayCallback,
		webhookSecret:         strings.TrimSpace(webhookSecret),
		webBaseURL:            strings.TrimRight(strings.TrimSpace(webBaseURL), "/"),
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
	PaymentMethod   string `json:"payment_method"`
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
		PaymentMethod:   body.PaymentMethod,
		ClientIP:        clientIP(r),
		AgainLink:       h.webBaseURL + "/",
		Items:           items,
		Actor:           actorFromRequest(r),
	})
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	payload := map[string]any{"data": created.Orders}
	if created.Payment != nil {
		payload["payment"] = created.Payment
	}
	writeJSON(w, http.StatusCreated, payload)
}

func (h *OrderingHandler) GetPaymentMethods(w http.ResponseWriter, r *http.Request) {
	item, err := h.getPaymentMethods.Handle(r.Context())
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *OrderingHandler) GetPaymentSettings(w http.ResponseWriter, r *http.Request) {
	item, err := h.getPaymentSettings.Handle(r.Context())
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

type onePayGatewayBody struct {
	Enabled    bool   `json:"enabled"`
	MerchantID string `json:"merchant_id"`
	AccessCode string `json:"access_code"`
	HashSecret string `json:"hash_secret"`
	PaymentURL string `json:"payment_url"`
}

type paymentSettingsBody struct {
	OnePayReturnURL     string             `json:"onepay_return_url"`
	OnePayIPNURL        string             `json:"onepay_ipn_url"`
	OnePayDomestic      onePayGatewayBody  `json:"onepay_domestic"`
	OnePayInternational onePayGatewayBody  `json:"onepay_international"`
}

func (h *OrderingHandler) UpdatePaymentSettings(w http.ResponseWriter, r *http.Request) {
	var body paymentSettingsBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	item, err := h.updatePaymentSettings.Handle(r.Context(), commands.UpdatePaymentSettingsCommand{
		OnePayReturnURL: body.OnePayReturnURL,
		OnePayIPNURL:    body.OnePayIPNURL,
		OnePayDomestic: commands.OnePayGatewayInput{
			Enabled:    body.OnePayDomestic.Enabled,
			MerchantID: body.OnePayDomestic.MerchantID,
			AccessCode: body.OnePayDomestic.AccessCode,
			HashSecret: body.OnePayDomestic.HashSecret,
			PaymentURL: body.OnePayDomestic.PaymentURL,
		},
		OnePayInternational: commands.OnePayGatewayInput{
			Enabled:    body.OnePayInternational.Enabled,
			MerchantID: body.OnePayInternational.MerchantID,
			AccessCode: body.OnePayInternational.AccessCode,
			HashSecret: body.OnePayInternational.HashSecret,
			PaymentURL: body.OnePayInternational.PaymentURL,
		},
	})
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *OrderingHandler) RepayUserOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	result, err := h.repay.Handle(r.Context(), commands.RepayOrderCommand{
		OrderID:   id,
		UserID:    userID,
		ClientIP:  clientIP(r),
		AgainLink: h.webBaseURL + "/",
		Actor:     actorFromRequest(r),
	})
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":    result.Order,
		"payment": result.Payment,
	})
}

func (h *OrderingHandler) OnePayReturn(w http.ResponseWriter, r *http.Request) {
	params := commands.ParamsFromURLValues(r.URL.Query())
	result, err := h.onePayCallback.Handle(r.Context(), commands.HandleOnePayCallbackCommand{
		Params:     params,
		Channel:    domain.PaymentCallbackChannelReturn,
		HTTPMethod: r.Method,
	})
	status := "failed"
	paymentID := ""
	orderID := ""
	if err == nil {
		paymentID = result.Payment.ID
		orderID = result.OrderID
		if result.Paid {
			status = "paid"
		}
	} else {
		status = "error"
	}
	target := h.webBaseURL + "/orders/payment/result?status=" + status
	if paymentID != "" {
		target += "&payment_id=" + paymentID
	}
	if orderID != "" {
		target += "&order_id=" + orderID
	}
	http.Redirect(w, r, target, http.StatusFound)
}

func (h *OrderingHandler) OnePayIPN(w http.ResponseWriter, r *http.Request) {
	params := commands.ParamsFromURLValues(r.URL.Query())
	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		for k, vs := range r.Form {
			if len(vs) > 0 {
				params[k] = vs[0]
			}
		}
	}
	result, err := h.onePayCallback.Handle(r.Context(), commands.HandleOnePayCallbackCommand{
		Params:     params,
		Channel:    domain.PaymentCallbackChannelIPN,
		HTTPMethod: r.Method,
	})
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":    result.Payment,
		"paid":    result.Paid,
		"message": "response received",
	})
}

func (h *OrderingHandler) ListPaymentCallbacks(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, err := h.listPaymentCallbacks.Handle(r.Context(), queries.ListPaymentCallbacksQuery{
		Provider:    strings.TrimSpace(r.URL.Query().Get("provider")),
		Channel:     strings.TrimSpace(r.URL.Query().Get("channel")),
		MerchTxnRef: strings.TrimSpace(r.URL.Query().Get("merch_txn_ref")),
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *OrderingHandler) GetPaymentCallback(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	item, err := h.getPaymentCallback.Handle(r.Context(), id)
	if err != nil {
		writeOrderingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func clientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if rip := strings.TrimSpace(r.Header.Get("X-Real-IP")); rip != "" {
		return rip
	}
	host := r.RemoteAddr
	if i := strings.LastIndex(host, ":"); i >= 0 {
		host = host[:i]
	}
	return strings.Trim(host, "[]")
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
		errors.Is(err, domain.ErrPaymentNotFound),
		errors.Is(err, domain.ErrPaymentCallbackNotFound),
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
		errors.Is(err, domain.ErrInvalidPaymentMethod),
		errors.Is(err, domain.ErrInvalidPaymentStatus),
		errors.Is(err, domain.ErrInvalidOnePayHash),
		errors.Is(err, domain.ErrOnePayNotConfigured),
		errors.Is(err, domain.ErrOnePayDisabled),
		errors.Is(err, domain.ErrOrderNotRepayable),
		errors.Is(err, catalogdomain.ErrInvalidProductID),
		errors.Is(err, catalogdomain.ErrInvalidQuantity):
		status = http.StatusBadRequest
	case errors.Is(err, domain.ErrDeliveryEventDuplicate),
		errors.Is(err, domain.ErrPaymentAlreadyHandled):
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
