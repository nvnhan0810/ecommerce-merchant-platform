package seed_test

import (
	"testing"

	cataloginfra "github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/infrastructure"
	identityinfra "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/infrastructure"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/seed"
)

func TestSeedDemo_should_be_idempotent_and_link_products_to_merchants(t *testing.T) {
	t.Parallel()

	users := identityinfra.NewInMemoryAccountRepository()
	merchants := identityinfra.NewInMemoryAccountRepository()
	admins := identityinfra.NewInMemoryAccountRepository()
	products := cataloginfra.NewInMemoryProductRepository()
	hasher := identityinfra.NewBcryptPasswordHasher()

	if err := seed.RunDemo(users, merchants, admins, products, hasher, "Admin@123456"); err != nil {
		t.Fatalf("first seed: %v", err)
	}
	if err := seed.RunDemo(users, merchants, admins, products, hasher, "Admin@123456"); err != nil {
		t.Fatalf("second seed: %v", err)
	}

	userN, err := users.Count()
	if err != nil {
		t.Fatal(err)
	}
	merchantN, err := merchants.Count()
	if err != nil {
		t.Fatal(err)
	}
	adminN, err := admins.Count()
	if err != nil {
		t.Fatal(err)
	}
	productN, err := products.Count()
	if err != nil {
		t.Fatal(err)
	}

	if userN != len(identityinfra.DemoUsers()) {
		t.Fatalf("users=%d want %d", userN, len(identityinfra.DemoUsers()))
	}
	if merchantN != len(identityinfra.DemoMerchants()) {
		t.Fatalf("merchants=%d want %d", merchantN, len(identityinfra.DemoMerchants()))
	}
	if adminN != len(identityinfra.DemoAdmins("Admin@123456")) {
		t.Fatalf("admins=%d want %d", adminN, len(identityinfra.DemoAdmins("x")))
	}
	if productN != 12 {
		t.Fatalf("products=%d want 12", productN)
	}

	list, err := products.List(100, 0)
	if err != nil {
		t.Fatal(err)
	}
	merchantIDs := map[string]struct{}{}
	ms, err := merchants.List()
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range ms {
		merchantIDs[string(m.ID)] = struct{}{}
	}
	for _, p := range list {
		if _, ok := merchantIDs[p.MerchantID]; !ok {
			t.Fatalf("product %q has unknown merchant_id %q", p.Name, p.MerchantID)
		}
	}
}
