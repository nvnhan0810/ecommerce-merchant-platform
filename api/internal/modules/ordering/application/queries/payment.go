package queries

import (
	"context"
	"strings"
	"time"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/infrastructure"
)

type PaymentDTO struct {
	ID          string `json:"id,omitempty"`
	Method      string `json:"method"`
	MethodLabel string `json:"method_label"`
	Status      string `json:"status"`
	StatusLabel string `json:"status_label"`
	AmountCents int64  `json:"amount_cents"`
	Currency    string `json:"currency"`
	MerchTxnRef string `json:"merch_txn_ref,omitempty"`
	PaymentURL  string `json:"payment_url,omitempty"`
	Message     string `json:"message,omitempty"`
}

func ToPaymentDTO(p domain.Payment, paymentURL string) PaymentDTO {
	return PaymentDTO{
		ID:          string(p.ID),
		Method:      string(p.Method),
		MethodLabel: p.Method.LabelVI(),
		Status:      string(p.Status),
		StatusLabel: p.Status.LabelVI(),
		AmountCents: p.AmountCents,
		Currency:    p.Currency,
		MerchTxnRef: p.MerchTxnRef,
		PaymentURL:  paymentURL,
		Message:     p.Message,
	}
}

type OnePayGatewaySettingsDTO struct {
	Enabled    bool   `json:"enabled"`
	MerchantID string `json:"merchant_id"`
	AccessCode string `json:"access_code"`
	HashSecret string `json:"hash_secret"`
	PaymentURL string `json:"payment_url"`
	Ready      bool   `json:"ready"`
}

type OnePayDemoCredentialsDTO struct {
	MerchantID string `json:"merchant_id"`
	AccessCode string `json:"access_code"`
	HashSecret string `json:"hash_secret"`
	PaymentURL string `json:"payment_url"`
	Note       string `json:"note,omitempty"`
}

type PaymentSettingsDTO struct {
	OnePayReturnURL     string                   `json:"onepay_return_url"`
	OnePayIPNURL        string                   `json:"onepay_ipn_url"`
	OnePayDomestic      OnePayGatewaySettingsDTO `json:"onepay_domestic"`
	OnePayInternational OnePayGatewaySettingsDTO `json:"onepay_international"`
	DemoDomestic        OnePayDemoCredentialsDTO `json:"demo_domestic"`
	DemoInternational   OnePayDemoCredentialsDTO `json:"demo_international"`
	UpdatedAt           time.Time                `json:"updated_at"`
}

func maskSecret(secret string) string {
	if strings.TrimSpace(secret) == "" {
		return ""
	}
	return "********"
}

func toGatewayDTO(g domain.OnePayGatewaySettings, returnURL, defaultURL string) OnePayGatewaySettingsDTO {
	paymentURL := strings.TrimSpace(g.PaymentURL)
	if paymentURL == "" {
		paymentURL = defaultURL
	}
	return OnePayGatewaySettingsDTO{
		Enabled:    g.Enabled,
		MerchantID: g.MerchantID,
		AccessCode: g.AccessCode,
		HashSecret: maskSecret(g.HashSecret),
		PaymentURL: paymentURL,
		Ready:      g.Ready(returnURL),
	}
}

func ToPaymentSettingsDTO(s domain.PaymentSettings, publicAPIBase string) PaymentSettingsDTO {
	publicAPIBase = strings.TrimRight(strings.TrimSpace(publicAPIBase), "/")
	returnURL := strings.TrimSpace(s.OnePayReturnURL)
	if returnURL == "" && publicAPIBase != "" {
		returnURL = publicAPIBase + "/api/v1/payments/onepay/return"
	}
	ipnURL := strings.TrimSpace(s.OnePayIPNURL)
	if ipnURL == "" && publicAPIBase != "" {
		ipnURL = publicAPIBase + "/api/v1/payments/onepay/ipn"
	}
	return PaymentSettingsDTO{
		OnePayReturnURL:     returnURL,
		OnePayIPNURL:        ipnURL,
		OnePayDomestic:      toGatewayDTO(s.OnePayDomestic, returnURL, infrastructure.DefaultOnePayDomesticPaymentURL),
		OnePayInternational: toGatewayDTO(s.OnePayInternational, returnURL, infrastructure.DefaultOnePayInternationalPaymentURL),
		DemoDomestic: OnePayDemoCredentialsDTO{
			MerchantID: infrastructure.OnePayDomesticDemo.MerchantID,
			AccessCode: infrastructure.OnePayDomesticDemo.AccessCode,
			HashSecret: infrastructure.OnePayDomesticDemo.HashSecret,
			PaymentURL: infrastructure.OnePayDomesticDemo.PaymentURL,
			Note:       "Sandbox nội địa hiện dùng TESTONEPAY (bộ ONEPAY/D67342C2 cũ thường lỗi sau khi nhập thẻ).",
		},
		DemoInternational: OnePayDemoCredentialsDTO{
			MerchantID: infrastructure.OnePayInternationalDemo.MerchantID,
			AccessCode: infrastructure.OnePayInternationalDemo.AccessCode,
			HashSecret: infrastructure.OnePayInternationalDemo.HashSecret,
			PaymentURL: infrastructure.OnePayInternationalDemo.PaymentURL,
			Note:       "Sandbox quốc tế TESTONEPAY trên vpcpay/vpcpay.op.",
		},
		UpdatedAt: s.UpdatedAt,
	}
}

type GetPaymentSettingsHandler struct {
	payments      domain.PaymentRepository
	publicAPIBase string
}

func NewGetPaymentSettingsHandler(payments domain.PaymentRepository, publicAPIBase string) *GetPaymentSettingsHandler {
	return &GetPaymentSettingsHandler{
		payments:      payments,
		publicAPIBase: strings.TrimRight(strings.TrimSpace(publicAPIBase), "/"),
	}
}

func (h *GetPaymentSettingsHandler) Handle(_ context.Context) (PaymentSettingsDTO, error) {
	s, err := h.payments.GetSettings()
	if err != nil {
		return PaymentSettingsDTO{}, err
	}
	return ToPaymentSettingsDTO(s, h.publicAPIBase), nil
}

type PublicPaymentMethodsDTO struct {
	Methods []PaymentMethodOptionDTO `json:"methods"`
}

type PaymentMethodOptionDTO struct {
	Method      string `json:"method"`
	Label       string `json:"label"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description,omitempty"`
}

type GetPublicPaymentMethodsHandler struct {
	payments domain.PaymentRepository
}

func NewGetPublicPaymentMethodsHandler(payments domain.PaymentRepository) *GetPublicPaymentMethodsHandler {
	return &GetPublicPaymentMethodsHandler{payments: payments}
}

func (h *GetPublicPaymentMethodsHandler) Handle(_ context.Context) (PublicPaymentMethodsDTO, error) {
	settings, err := h.payments.GetSettings()
	if err != nil {
		return PublicPaymentMethodsDTO{}, err
	}
	returnURL := strings.TrimSpace(settings.OnePayReturnURL)
	if returnURL == "" {
		// Ready() still needs non-empty return URL; treat configured credentials as enough for listing.
		returnURL = "configured"
	}
	return PublicPaymentMethodsDTO{
		Methods: []PaymentMethodOptionDTO{
			{
				Method:      string(domain.PaymentMethodCOD),
				Label:       domain.PaymentMethodCOD.LabelVI(),
				Enabled:     true,
				Description: "Thanh toán bằng tiền mặt khi nhận hàng",
			},
			{
				Method:      string(domain.PaymentMethodOnePayDomestic),
				Label:       domain.PaymentMethodOnePayDomestic.LabelVI(),
				Enabled:     settings.OnePayDomestic.Ready(returnURL),
				Description: "Thẻ ATM nội địa / tài khoản ngân hàng",
			},
			{
				Method:      string(domain.PaymentMethodOnePayInternational),
				Label:       domain.PaymentMethodOnePayInternational.LabelVI(),
				Enabled:     settings.OnePayInternational.Ready(returnURL),
				Description: "Thẻ Visa / Mastercard / JCB",
			},
		},
	}, nil
}
