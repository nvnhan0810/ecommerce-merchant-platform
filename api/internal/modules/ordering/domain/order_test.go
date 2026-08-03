package domain

import "testing"

func TestParseOrderStatus(t *testing.T) {
	t.Parallel()
	cases := []OrderStatus{
		StatusNew, StatusPaid, StatusConfirmed, StatusShipping, StatusSucceeded, StatusFailed, StatusCancelled,
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
	_, err := NewOrder("", "m1", "VND", "", nil)
	if err != ErrUserRequired {
		t.Fatalf("expected ErrUserRequired, got %v", err)
	}
	_, err = NewOrder("u1", "", "VND", "", []OrderLineInput{{ProductID: "p", MerchantID: "m1", UnitPriceCents: 1, Quantity: 1}})
	if err != ErrMerchantRequired {
		t.Fatalf("expected ErrMerchantRequired, got %v", err)
	}
	_, err = NewOrder("u1", "m1", "VND", "", nil)
	if err != ErrEmptyOrderItems {
		t.Fatalf("expected ErrEmptyOrderItems, got %v", err)
	}
}

func TestNewOrder_rejectsProductFromOtherMerchant(t *testing.T) {
	t.Parallel()
	_, err := NewOrder("u1", "m1", "VND", "", []OrderLineInput{
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
	o, err := NewOrder("u1", "m1", "vnd", "note", []OrderLineInput{
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
