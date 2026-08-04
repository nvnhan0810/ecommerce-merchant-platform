package infrastructure

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
)

const (
	DefaultOnePayDomesticPaymentURL      = "https://mtf.onepay.vn/onecomm-pay/vpc.op"
	DefaultOnePayInternationalPaymentURL = "https://mtf.onepay.vn/vpcpay/vpcpay.op"
)

// Deprecated alias kept for older call sites during transition.
const DefaultOnePayPaymentURL = DefaultOnePayDomesticPaymentURL

// Public OnePay sandbox credentials (mtf.onepay.vn).
// Domestic classic merchant "ONEPAY" is outdated on many sandboxes and often fails
// after card entry ("Hết thời gian thanh toán / giao dịch không hợp lệ").
// Current public sandbox uses TESTONEPAY for both ATM (nội địa) and Visa/Master.
var OnePayDomesticDemo = domain.OnePayGatewaySettings{
	Enabled:    true,
	MerchantID: "TESTONEPAY",
	AccessCode: "6BEB2546",
	HashSecret: "6D0870CDE5F24F34F3915FB0045120DB",
	PaymentURL: DefaultOnePayDomesticPaymentURL,
}

var OnePayInternationalDemo = domain.OnePayGatewaySettings{
	Enabled:    true,
	MerchantID: "TESTONEPAY",
	AccessCode: "6BEB2546",
	HashSecret: "6D0870CDE5F24F34F3915FB0045120DB",
	PaymentURL: DefaultOnePayInternationalPaymentURL,
}

func DefaultPaymentURLFor(method domain.PaymentMethod) string {
	if method == domain.PaymentMethodOnePayInternational {
		return DefaultOnePayInternationalPaymentURL
	}
	return DefaultOnePayDomesticPaymentURL
}

// OnePaySecureHash builds the VPC HMAC-SHA256 secure hash (hex secret packed as bytes).
func OnePaySecureHash(hashSecret string, params map[string]string) (string, error) {
	secretHex := strings.TrimSpace(hashSecret)
	if secretHex == "" {
		return "", domain.ErrOnePayNotConfigured
	}
	key, err := hex.DecodeString(secretHex)
	if err != nil {
		key = []byte(secretHex)
	}

	keys := make([]string, 0, len(params))
	for k, v := range params {
		if v == "" {
			continue
		}
		if strings.HasPrefix(k, "vpc_") || strings.HasPrefix(k, "user_") {
			if k == "vpc_SecureHash" || k == "vpc_SecureHashType" {
				continue
			}
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	payload := strings.Join(parts, "&")

	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(payload))
	return strings.ToUpper(hex.EncodeToString(mac.Sum(nil))), nil
}

func VerifyOnePaySecureHash(hashSecret string, params map[string]string) error {
	received := strings.TrimSpace(params["vpc_SecureHash"])
	if received == "" {
		return domain.ErrInvalidOnePayHash
	}
	expected, err := OnePaySecureHash(hashSecret, params)
	if err != nil {
		return err
	}
	if !hmac.Equal([]byte(strings.ToUpper(received)), []byte(expected)) {
		return domain.ErrInvalidOnePayHash
	}
	return nil
}

// SanitizeOnePayTicketNo keeps only IPv4 (max 15 chars) as required by OnePay.
func SanitizeOnePayTicketNo(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "127.0.0.1"
	}
	// X-Forwarded-For may include port or multiple IPs.
	if i := strings.Index(raw, ","); i >= 0 {
		raw = strings.TrimSpace(raw[:i])
	}
	if host, _, err := net.SplitHostPort(raw); err == nil {
		raw = host
	}
	raw = strings.Trim(raw, "[]")
	ip := net.ParseIP(raw)
	if ip == nil {
		return "127.0.0.1"
	}
	if v4 := ip.To4(); v4 != nil {
		return v4.String()
	}
	// OnePay TicketNo is IPv4-only (max 15).
	return "127.0.0.1"
}

type OnePayCheckoutInput struct {
	Gateway     domain.OnePayGatewaySettings
	ReturnURL   string
	AgainLink   string
	Title       string
	MerchTxnRef string
	AmountCents int64
	OrderInfo   string
	Locale      string
	ClientIP    string
	Method      domain.PaymentMethod
}

func BuildOnePayRedirectURL(in OnePayCheckoutInput) (string, error) {
	returnURL := strings.TrimSpace(in.ReturnURL)
	if !in.Gateway.Ready(returnURL) {
		if !in.Gateway.Enabled {
			return "", domain.ErrOnePayDisabled
		}
		return "", domain.ErrOnePayNotConfigured
	}
	paymentURL := strings.TrimSpace(in.Gateway.PaymentURL)
	if paymentURL == "" {
		paymentURL = DefaultPaymentURLFor(in.Method)
	}
	locale := strings.TrimSpace(in.Locale)
	if locale == "" {
		locale = "vn"
	}
	orderInfo := strings.TrimSpace(in.OrderInfo)
	if orderInfo == "" {
		orderInfo = "Thanh toan don hang"
	}
	// Keep ASCII-safe and within OnePay limit.
	orderInfo = asciiSafe(orderInfo)
	if len(orderInfo) > 34 {
		orderInfo = orderInfo[:34]
	}

	title := strings.TrimSpace(in.Title)
	if title == "" {
		title = "Thanh toan don hang"
	}
	title = asciiSafe(title)
	if len(title) > 64 {
		title = title[:64]
	}

	againLink := strings.TrimSpace(in.AgainLink)
	if againLink == "" {
		// Fallback: origin of return URL (still better than empty — AgainLink is mandatory).
		if u, err := url.Parse(returnURL); err == nil && u.Scheme != "" && u.Host != "" {
			againLink = u.Scheme + "://" + u.Host + "/"
		} else {
			againLink = returnURL
		}
	}
	if len(againLink) > 64 {
		againLink = againLink[:64]
	}

	merchRef := strings.TrimSpace(in.MerchTxnRef)
	if merchRef == "" {
		merchRef = fmt.Sprintf("T%d", time.Now().UnixNano())
	}
	if len(merchRef) > 40 {
		merchRef = merchRef[:40]
	}

	ticketNo := SanitizeOnePayTicketNo(in.ClientIP)

	// AmountCents stores VND integer; OnePay requires amount * 100.
	amount := strconv.FormatInt(in.AmountCents*100, 10)

	// Hashable vpc_/user_ params only. Title + AgainLink are mandatory query fields
	// but are NOT included in the secure hash (official OnePay PHP samples).
	params := map[string]string{
		"vpc_Version":     "2",
		"vpc_Command":     "pay",
		"vpc_AccessCode":  strings.TrimSpace(in.Gateway.AccessCode),
		"vpc_Merchant":    strings.TrimSpace(in.Gateway.MerchantID),
		"vpc_Locale":      locale,
		"vpc_TicketNo":    ticketNo,
		"vpc_ReturnURL":   returnURL,
		"vpc_MerchTxnRef": merchRef,
		"vpc_OrderInfo":   orderInfo,
		"vpc_Amount":      amount,
	}

	// Domestic requires vpc_Currency. International gateway must omit it from both
	// the request checksum set and (practically) the hashed fields — matching OnePay samples.
	if in.Method != domain.PaymentMethodOnePayInternational {
		params["vpc_Currency"] = "VND"
	}

	hash, err := OnePaySecureHash(in.Gateway.HashSecret, params)
	if err != nil {
		return "", err
	}
	params["vpc_SecureHash"] = hash

	u, err := url.Parse(paymentURL)
	if err != nil {
		return "", fmt.Errorf("invalid onepay payment url: %w", err)
	}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	q.Set("Title", title)
	q.Set("AgainLink", againLink)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func asciiSafe(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= 32 && r <= 126 {
			b.WriteRune(r)
		} else if r == ' ' {
			b.WriteByte(' ')
		} else {
			b.WriteByte(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func OnePayResponseSuccess(code string) bool {
	return strings.TrimSpace(code) == "0"
}
