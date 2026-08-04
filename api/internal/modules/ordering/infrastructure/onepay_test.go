package infrastructure_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/infrastructure"
)

func TestOnePaySecureHash_is_deterministic(t *testing.T) {
	t.Parallel()
	secret := "6D0870CDE5F24F34F3915FB0045120DB"
	params := map[string]string{
		"vpc_Amount":     "10000000",
		"vpc_Command":    "pay",
		"vpc_Currency":    "VND",
		"vpc_AccessCode":  "6BEB2546",
		"vpc_Merchant":    "TESTONEPAY",
		"vpc_MerchTxnRef": "abc123",
		"vpc_OrderInfo":   "DH TEST",
		"vpc_ReturnURL":   "https://example.com/return",
		"vpc_Version":     "2",
		"vpc_Locale":      "vn",
		"vpc_TicketNo":    "127.0.0.1",
	}
	h1, err := infrastructure.OnePaySecureHash(secret, params)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := infrastructure.OnePaySecureHash(secret, params)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 || len(h1) != 64 {
		t.Fatalf("unexpected hash %q", h1)
	}
	params["vpc_SecureHash"] = h1
	if err := infrastructure.VerifyOnePaySecureHash(secret, params); err != nil {
		t.Fatal(err)
	}
}

func TestBuildOnePayRedirectURL_requires_config(t *testing.T) {
	t.Parallel()
	_, err := infrastructure.BuildOnePayRedirectURL(infrastructure.OnePayCheckoutInput{
		Gateway:   domain.OnePayGatewaySettings{Enabled: true},
		ReturnURL: "https://example.com/return",
		Method:    domain.PaymentMethodOnePayDomestic,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "configured") && err != domain.ErrOnePayNotConfigured {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestBuildOnePayRedirectURL_demo_domestic_includes_mandatory_fields(t *testing.T) {
	t.Parallel()
	raw, err := infrastructure.BuildOnePayRedirectURL(infrastructure.OnePayCheckoutInput{
		Gateway:     infrastructure.OnePayDomesticDemo,
		ReturnURL:   "https://ecomerce-api.nvnhan0810.com/api/v1/payments/onepay/return",
		AgainLink:   "https://ecomerce.nvnhan0810.com/",
		Title:       "Thanh toan don hang",
		MerchTxnRef: "demo123",
		AmountCents: 10000,
		ClientIP:    "203.0.113.10",
		Method:      domain.PaymentMethodOnePayDomestic,
	})
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if !strings.Contains(raw, "mtf.onepay.vn/onecomm-pay") {
		t.Fatalf("unexpected url: %s", raw)
	}
	if q.Get("Title") == "" || q.Get("AgainLink") == "" {
		t.Fatalf("missing Title/AgainLink: %s", raw)
	}
	if q.Get("vpc_Currency") != "VND" {
		t.Fatalf("domestic must send vpc_Currency, got %q", q.Get("vpc_Currency"))
	}
	if q.Get("vpc_Amount") != "1000000" { // 10000 VND * 100
		t.Fatalf("amount=%s", q.Get("vpc_Amount"))
	}
	if q.Get("vpc_TicketNo") != "203.0.113.10" {
		t.Fatalf("ticket=%s", q.Get("vpc_TicketNo"))
	}
}

func TestBuildOnePayRedirectURL_international_omits_currency(t *testing.T) {
	t.Parallel()
	raw, err := infrastructure.BuildOnePayRedirectURL(infrastructure.OnePayCheckoutInput{
		Gateway:     infrastructure.OnePayInternationalDemo,
		ReturnURL:   "https://ecomerce-api.nvnhan0810.com/api/v1/payments/onepay/return",
		AgainLink:   "https://ecomerce.nvnhan0810.com/",
		MerchTxnRef: "demo456",
		AmountCents: 10000,
		ClientIP:    "2001:db8::1",
		Method:      domain.PaymentMethodOnePayInternational,
	})
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if q.Get("vpc_Currency") != "" {
		t.Fatalf("international must omit vpc_Currency, got %q", q.Get("vpc_Currency"))
	}
	if q.Get("vpc_TicketNo") != "127.0.0.1" {
		t.Fatalf("ipv6 should fallback, got %s", q.Get("vpc_TicketNo"))
	}
	if !strings.Contains(raw, "vpcpay/vpcpay.op") {
		t.Fatalf("unexpected url: %s", raw)
	}
}

func TestSanitizeOnePayTicketNo(t *testing.T) {
	t.Parallel()
	if got := infrastructure.SanitizeOnePayTicketNo("1.2.3.4, 5.6.7.8"); got != "1.2.3.4" {
		t.Fatalf("got %s", got)
	}
}
