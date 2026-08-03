package seed

import (
	"fmt"
	"log"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
	cataloginfra "github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/infrastructure"
	identitydomain "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
	identityinfra "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/infrastructure"
	orderingdomain "github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
	orderinginfra "github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/infrastructure"
)

// RunDemo seeds identity + catalog + orders demo data. Safe to re-run (idempotent).
func RunDemo(
	users, merchants, admins identitydomain.AccountRepository,
	products domain.ProductRepository,
	orders orderingdomain.OrderRepository,
	hasher identitydomain.PasswordHasher,
	adminPassword string,
) error {
	if err := identityinfra.SeedDemoAccounts(users, merchants, admins, hasher, adminPassword); err != nil {
		return fmt.Errorf("accounts: %w", err)
	}
	if err := RunProducts(products, merchants); err != nil {
		return err
	}
	if err := RunOrders(orders, users, merchants, products); err != nil {
		return err
	}
	return nil
}

// RunProducts seeds demo products linked to existing merchants.
func RunProducts(products domain.ProductRepository, merchants identitydomain.AccountRepository) error {
	if err := cataloginfra.SeedDemoProducts(products, merchants); err != nil {
		return fmt.Errorf("products: %w", err)
	}
	return nil
}

// RunOrders seeds demo orders (user + merchant + products of that merchant).
func RunOrders(
	orders orderingdomain.OrderRepository,
	users, merchants identitydomain.AccountRepository,
	products domain.ProductRepository,
) error {
	if err := orderinginfra.SeedDemoOrders(orders, users, merchants, products); err != nil {
		return fmt.Errorf("orders: %w", err)
	}
	return nil
}

// LogSummary prints demo credentials for local/dev use.
func LogSummary(adminPassword string) {
	if adminPassword == "" {
		adminPassword = "Admin@123456"
	}
	log.Println("demo seed ready:")
	log.Printf("  admin     admin@ecomerce.local / %s", adminPassword)
	log.Printf("  admin     ops@ecomerce.local / %s", adminPassword)
	log.Println("  merchant  shop@ecomerce.local / Shop@123456 (+ fashion/tech/home)")
	log.Println("  user      buyer@ecomerce.local / Buyer@123456 (+ an/binh/chi)")
	log.Printf("  products  %d demo SKUs across shop/fashion/tech/home", len(cataloginfra.DemoProducts()))
	log.Println("  orders    7 demo orders with tracking codes (new/paid/confirmed/shipping/succeeded/failed/cancelled)")
}
