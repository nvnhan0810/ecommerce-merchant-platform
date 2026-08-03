package infrastructure_test

import (
	"testing"

	cataloginfra "github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/infrastructure"
	identityinfra "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/infrastructure"
)

func TestSeedDemoProducts_should_attach_to_merchants(t *testing.T) {
	t.Parallel()

	merchants := identityinfra.NewInMemoryAccountRepository()
	hasher := identityinfra.NewBcryptPasswordHasher()
	users := identityinfra.NewInMemoryAccountRepository()
	admins := identityinfra.NewInMemoryAccountRepository()
	if err := identityinfra.SeedDemoAccounts(users, merchants, admins, hasher, "Admin@123456"); err != nil {
		t.Fatal(err)
	}

	products := cataloginfra.NewInMemoryProductRepository()
	if err := cataloginfra.SeedDemoProducts(products, merchants); err != nil {
		t.Fatal(err)
	}
	if err := cataloginfra.SeedDemoProducts(products, merchants); err != nil {
		t.Fatal(err)
	}

	n, err := products.Count()
	if err != nil {
		t.Fatal(err)
	}
	want := len(cataloginfra.DemoProducts())
	if n != want {
		t.Fatalf("products=%d want %d", n, want)
	}

	list, err := products.List(500, 0)
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]struct{}{}
	ms, _ := merchants.List()
	for _, m := range ms {
		ids[string(m.ID)] = struct{}{}
	}
	for _, p := range list {
		if _, ok := ids[p.MerchantID]; !ok {
			t.Fatalf("product %q orphan merchant %q", p.Name, p.MerchantID)
		}
		if p.MerchantID == "" {
			t.Fatal("empty merchant_id")
		}
	}
}
