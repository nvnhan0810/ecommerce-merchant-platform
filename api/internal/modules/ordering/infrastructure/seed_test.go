package infrastructure_test

import (
	"testing"

	cataloginfra "github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/infrastructure"
	identityinfra "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/infrastructure"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
	orderinginfra "github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/infrastructure"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/seed"
)

func TestSeedDemoOrders_should_create_one_per_status(t *testing.T) {
	t.Parallel()

	users := identityinfra.NewInMemoryAccountRepository()
	merchants := identityinfra.NewInMemoryAccountRepository()
	admins := identityinfra.NewInMemoryAccountRepository()
	products := cataloginfra.NewInMemoryProductRepository()
	orders := orderinginfra.NewInMemoryOrderRepository()
	hasher := identityinfra.NewBcryptPasswordHasher()

	if err := seed.RunDemo(users, merchants, admins, products, orders, hasher, "Admin@123456"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := seed.RunDemo(users, merchants, admins, products, orders, hasher, "Admin@123456"); err != nil {
		t.Fatalf("reseed: %v", err)
	}

	n, err := orders.Count()
	if err != nil {
		t.Fatal(err)
	}
	if n != 7 {
		t.Fatalf("orders=%d want 7", n)
	}

	list, err := orders.List(20, 0)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[domain.OrderStatus]int{}
	codes := map[string]struct{}{}
	for _, o := range list {
		seen[o.Status]++
		if o.UserID == "" || o.MerchantID == "" {
			t.Fatalf("order missing user/merchant: %+v", o)
		}
		if _, err := domain.ParseOrderCode(o.Code); err != nil {
			t.Fatalf("invalid order code %q: %v", o.Code, err)
		}
		if _, dup := codes[o.Code]; dup {
			t.Fatalf("duplicate order code %q", o.Code)
		}
		codes[o.Code] = struct{}{}
		if len(o.Items) == 0 {
			t.Fatal("order has no items")
		}
		for _, item := range o.Items {
			// product merchant ownership already enforced at NewOrder time
			if item.ProductID == "" {
				t.Fatal("empty product id")
			}
		}
	}
	for _, st := range []domain.OrderStatus{
		domain.StatusNew, domain.StatusPaid, domain.StatusConfirmed, domain.StatusShipping,
		domain.StatusSucceeded, domain.StatusFailed, domain.StatusCancelled,
	} {
		if seen[st] != 1 {
			t.Fatalf("status %s count=%d", st, seen[st])
		}
	}
}
