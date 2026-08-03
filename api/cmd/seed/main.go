package main

import (
	"context"
	"fmt"
	"log"
	"os"

	cataloginfra "github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/infrastructure"
	identityinfra "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/infrastructure"
	orderinginfra "github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/infrastructure"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/config"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/envfile"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/postgres"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/seed"
)

func main() {
	_ = envfile.Load(".env", "/app/.env")
	cfg := config.Load()

	cmd := "all"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "help", "-h", "--help":
		printUsage()
		return
	case "all", "products", "orders":
		// handled below
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(2)
	}

	ctx := context.Background()
	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	users := identityinfra.NewPostgresUserRepository(pool)
	merchants := identityinfra.NewPostgresMerchantRepository(pool)
	admins := identityinfra.NewPostgresAdminRepository(pool)
	products := cataloginfra.NewPostgresProductRepository(pool)
	orders := orderinginfra.NewPostgresOrderRepository(pool)
	hasher := identityinfra.NewBcryptPasswordHasher()

	switch cmd {
	case "products":
		if err := seed.RunProducts(products, merchants); err != nil {
			log.Fatalf("seed: %v", err)
		}
	case "orders":
		if err := seed.RunOrders(orders, users, merchants, products); err != nil {
			log.Fatalf("seed: %v", err)
		}
	default:
		if err := seed.RunDemo(users, merchants, admins, products, orders, hasher, cfg.AdminBootstrapPassword); err != nil {
			log.Fatalf("seed: %v", err)
		}
	}

	userN, _ := users.Count()
	merchantN, _ := merchants.Count()
	adminN, _ := admins.Count()
	productN, _ := products.Count()
	orderN, _ := orders.Count()
	log.Printf("counts: users=%d merchants=%d admins=%d products=%d orders=%d", userN, merchantN, adminN, productN, orderN)
	seed.LogSummary(cfg.AdminBootstrapPassword)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: seed [command]

Commands:
  all        Seed demo accounts + products + orders (default)
  products   Seed demo products only (merchants must already exist)
  orders     Seed demo orders only (users/merchants/products must exist)

Env: same as API (.env / DB_* / ADMIN_BOOTSTRAP_PASSWORD)

Host examples:
  DB_HOST=127.0.0.1 DB_SSLMODE=require go run ./cmd/seed
  DB_HOST=127.0.0.1 DB_SSLMODE=require go run ./cmd/seed products
  DB_HOST=127.0.0.1 DB_SSLMODE=require go run ./cmd/seed orders
`)
}
