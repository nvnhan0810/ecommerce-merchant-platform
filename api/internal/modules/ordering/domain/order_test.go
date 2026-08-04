package domain

import (
	"strings"
	"testing"
)

func TestParseOrderStatus(t *testing.T) {
	t.Parallel()
	cases := []OrderStatus{
		StatusNew, StatusPaid, StatusConfirmed, StatusShipping, StatusSucceeded,
		StatusReturning, StatusReturned, StatusFailed, StatusCancelled,
	}
	for _, want := range cases {
		got, err := ParseOrderStatus(string(want))
		if err != nil || got != want {
			t.Fatalf("%s: got=%q err=%v", want, got, err)
		}
	}
	if _, err := ParseOrderStatus("unknown"); err != ErrInvalidOrderStatus {
		t.Fatalf("expected ErrInvalidOrderStatus, got %v", err)
	}
}

func TestNewOrder_requiresUserMerchantAndItems(t *testing.T) {
	t.Parallel()
	_, err := NewOrder("", "m1", "VND", "", "", "", "", nil)
	if err != ErrUserRequired {
		t.Fatalf("expected ErrUserRequired, got %v", err)
	}
	_, err = NewOrder("u1", "", "VND", "", "", "", "", []OrderLineInput{{ProductID: "p", MerchantID: "m1", UnitPriceCents: 1, Quantity: 1}})
	if err != ErrMerchantRequired {
		t.Fatalf("expected ErrMerchantRequired, got %v", err)
	}
	_, err = NewOrder("u1", "m1", "VND", "", "", "", "", nil)
	if err != ErrEmptyOrderItems {
		t.Fatalf("expected ErrEmptyOrderItems, got %v", err)
	}
}

func TestNewOrder_requiresShippingInfo(t *testing.T) {
	t.Parallel()
	_, err := NewOrder("u1", "m1", "VND", "", "", "", "", []OrderLineInput{
		{ProductID: "p1", ProductName: "A", MerchantID: "m1", UnitPriceCents: 1000, Quantity: 1},
	})
	if err != ErrMissingShippingInfo {
		t.Fatalf("expected ErrMissingShippingInfo, got %v", err)
	}
}

func TestNewOrder_rejectsProductFromOtherMerchant(t *testing.T) {
	t.Parallel()
	_, err := NewOrder("u1", "m1", "VND", "", "Nguyen A", "0901111222", "12 Nguyen Hue, HCM", []OrderLineInput{
		{ProductID: "p1", ProductName: "A", MerchantID: "m2", UnitPriceCents: 1000, Quantity: 1},
	})
	if err != ErrProductMerchantMismatch {
		t.Fatalf("expected ErrProductMerchantMismatch, got %v", err)
	}
}

func TestNewOrderCode_and_ParseOrderCode(t *testing.T) {
	t.Parallel()
	code, err := NewOrderCode()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseOrderCode(code)
	if err != nil || got != code {
		t.Fatalf("got=%q err=%v", got, err)
	}
	if _, err := ParseOrderCode("short"); err != ErrInvalidOrderCode {
		t.Fatalf("expected ErrInvalidOrderCode, got %v", err)
	}
	if _, err := ParseOrderCode("ABCDEFGHI!"); err != ErrInvalidOrderCode {
		t.Fatalf("expected ErrInvalidOrderCode for symbol, got %v", err)
	}
}

func TestNewOrder_ok(t *testing.T) {
	t.Parallel()
	o, err := NewOrder("u1", "m1", "vnd", "note", "Name", "Phone", "Addr", []OrderLineInput{
		{ProductID: "p1", ProductName: "Áo", MerchantID: "m1", UnitPriceCents: 10000, Quantity: 2},
		{ProductID: "p2", ProductName: "Quần", MerchantID: "m1", UnitPriceCents: 20000, Quantity: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if o.Status != StatusNew {
		t.Fatalf("status=%s", o.Status)
	}
	if o.Currency != "VND" {
		t.Fatalf("currency=%s", o.Currency)
	}
	if o.TotalCents != 40000 {
		t.Fatalf("total=%d", o.TotalCents)
	}
	if _, err := ParseOrderCode(o.Code); err != nil {
		t.Fatalf("code=%q err=%v", o.Code, err)
	}
	if len(o.Items) != 2 {
		t.Fatalf("items=%d", len(o.Items))
	}
	o.RecordCreated(Actor{ID: "u1", Email: "buyer@x", Role: "user", DisplayName: "Buyer"})
	if err := o.ChangeStatus(StatusPaid, Actor{ID: "a1", Email: "admin@x", Role: "admin", DisplayName: "Admin"}); err != nil {
		t.Fatal(err)
	}
	if o.Status != StatusPaid {
		t.Fatalf("status after change=%s", o.Status)
	}
	if len(o.PendingEvents()) != 2 {
		t.Fatalf("pending events=%d", len(o.PendingEvents()))
	}
	if len(o.History) != 2 {
		t.Fatalf("history=%d", len(o.History))
	}
}

func TestMapDeliveryToOrderStatus(t *testing.T) {
	t.Parallel()
	cases := []struct {
		code   DeliveryStatusCode
		status OrderStatus
		change bool
	}{
		{DeliveryAccepted, StatusShipping, true},
		{DeliveryDelivering, StatusShipping, false},
		{DeliveryInTransit, StatusShipping, false},
		{DeliveryFail, StatusShipping, false},
		{DeliveryDelivered, StatusSucceeded, true},
		{DeliveryReturning, StatusReturning, true},
		{DeliveryReturned, StatusReturned, true},
		{DeliveryReturnFail, StatusFailed, true},
		{DeliveryLost, StatusFailed, true},
		{DeliveryDamage, StatusFailed, true},
	}
	for _, tc := range cases {
		got, change := MapDeliveryToOrderStatus(tc.code)
		if got != tc.status || change != tc.change {
			t.Fatalf("%s: got=%s change=%v want=%s/%v", tc.code, got, change, tc.status, tc.change)
		}
	}
}

func TestApplyDeliveryEvent_setsTrackingAndStatus(t *testing.T) {
	t.Parallel()
	o, err := NewOrder("u1", "m1", "VND", "", "Nguyen A", "0901111222", "12 Nguyen Hue, HCM", []OrderLineInput{
		{ProductID: "p1", ProductName: "Áo", MerchantID: "m1", UnitPriceCents: 1000, Quantity: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := o.ChangeStatus(StatusConfirmed, SystemActor()); err != nil {
		t.Fatal(err)
	}
	ev, err := o.ApplyDeliveryEvent(ApplyDeliveryInput{
		DeliveryTrackingCode: "GHN123",
		DeliveryCarrier:      "ghn",
		StatusCode:           DeliveryAccepted,
		Message:              "Đã tiếp nhận",
		Source:               "simulate",
	})
	if err != nil {
		t.Fatal(err)
	}
	if o.DeliveryTrackingCode != "GHN123" {
		t.Fatalf("tracking=%q", o.DeliveryTrackingCode)
	}
	if o.DeliveryCarrier != "ghn" {
		t.Fatalf("carrier=%q", o.DeliveryCarrier)
	}
	if o.Status != StatusShipping {
		t.Fatalf("status=%s", o.Status)
	}
	if ev.StatusCode != DeliveryAccepted {
		t.Fatalf("event status=%s", ev.StatusCode)
	}
	if len(o.PendingDeliveryEvents()) != 1 {
		t.Fatalf("pending delivery=%d", len(o.PendingDeliveryEvents()))
	}

	_, err = o.ApplyDeliveryEvent(ApplyDeliveryInput{
		StatusCode: DeliveryFail,
		Message:    "Không liên lạc được",
		Source:     "webhook",
	})
	if err != nil {
		t.Fatal(err)
	}
	if o.Status != StatusShipping {
		t.Fatalf("status after fail=%s", o.Status)
	}

	_, err = o.ApplyDeliveryEvent(ApplyDeliveryInput{
		StatusCode: DeliveryDelivered,
		Source:     "webhook",
	})
	if err != nil {
		t.Fatal(err)
	}
	if o.Status != StatusSucceeded {
		t.Fatalf("status after delivered=%s", o.Status)
	}
}

func TestMerchantConfirm(t *testing.T) {
	t.Parallel()
	o, err := NewOrder("u1", "m1", "VND", "", "Nguyen A", "0901111222", "12 Nguyen Hue, HCM", []OrderLineInput{
		{ProductID: "p1", ProductName: "Áo", MerchantID: "m1", UnitPriceCents: 1000, Quantity: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := o.MerchantConfirm(Actor{Role: "merchant", Email: "m@x"}); err != nil {
		t.Fatal(err)
	}
	if o.Status != StatusShipping {
		t.Fatalf("status=%s want shipping (auto-dispatch)", o.Status)
	}
	if o.DeliveryTrackingCode == "" || !strings.HasPrefix(o.DeliveryTrackingCode, "DELI-") {
		t.Fatalf("tracking=%q", o.DeliveryTrackingCode)
	}
	if o.DeliveryCarrier != DefaultDeliveryCarrier {
		t.Fatalf("carrier=%q", o.DeliveryCarrier)
	}
	if len(o.PendingDeliveryEvents()) != 1 {
		t.Fatalf("delivery events=%d", len(o.PendingDeliveryEvents()))
	}
	if o.PendingDeliveryEvents()[0].StatusCode != DeliveryAccepted {
		t.Fatalf("delivery status=%s", o.PendingDeliveryEvents()[0].StatusCode)
	}
	if err := o.MerchantConfirm(Actor{Role: "merchant"}); err != ErrMerchantConfirmOnly {
		t.Fatalf("expected ErrMerchantConfirmOnly, got %v", err)
	}
}

func TestMerchantCancel_requiresReason(t *testing.T) {
	t.Parallel()
	o, err := NewOrder("u1", "m1", "VND", "", "Nguyen A", "0901111222", "12 Nguyen Hue, HCM", []OrderLineInput{
		{ProductID: "p1", ProductName: "Áo", MerchantID: "m1", UnitPriceCents: 1000, Quantity: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := o.MerchantCancel(Actor{Role: "merchant"}, "  "); err != ErrMerchantCancelReasonRequired {
		t.Fatalf("expected ErrMerchantCancelReasonRequired, got %v", err)
	}
	if err := o.MerchantCancel(Actor{Role: "merchant"}, "Hết hàng"); err != nil {
		t.Fatal(err)
	}
	if o.Status != StatusCancelled {
		t.Fatalf("status=%s", o.Status)
	}
	evs := o.PendingEvents()
	if len(evs) == 0 || !strings.Contains(evs[len(evs)-1].Message, "Hết hàng") {
		t.Fatalf("message=%v", evs)
	}
}

func TestMerchantCancel_fromConfirmed(t *testing.T) {
	t.Parallel()
	o, err := NewOrder("u1", "m1", "VND", "", "Nguyen A", "0901111222", "12 Nguyen Hue, HCM", []OrderLineInput{
		{ProductID: "p1", ProductName: "Áo", MerchantID: "m1", UnitPriceCents: 1000, Quantity: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	// Confirmed-without-dispatch path (manual status) can still cancel.
	if err := o.ChangeStatus(StatusConfirmed, Actor{Role: "admin"}); err != nil {
		t.Fatal(err)
	}
	if err := o.MerchantCancel(Actor{Role: "merchant"}, "Khách đổi ý"); err != nil {
		t.Fatal(err)
	}
	if o.Status != StatusCancelled {
		t.Fatalf("status=%s", o.Status)
	}

	shipping, err := NewOrder("u1", "m1", "VND", "", "Nguyen A", "0901111222", "12 Nguyen Hue, HCM", []OrderLineInput{
		{ProductID: "p1", ProductName: "Áo", MerchantID: "m1", UnitPriceCents: 1000, Quantity: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = shipping.MerchantConfirm(Actor{Role: "merchant"})
	if shipping.Status != StatusShipping {
		t.Fatalf("status=%s", shipping.Status)
	}
	if err := shipping.MerchantCancel(Actor{Role: "merchant"}, "late"); err != ErrMerchantConfirmOnly {
		t.Fatalf("expected ErrMerchantConfirmOnly after dispatch, got %v", err)
	}
}
