package main

import (
	"context"
	"fmt"
	"log"
	"os"

	cataloginfra "github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/infrastructure"
	identityinfra "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/infrastructure"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/config"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/envfile"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/postgres"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/seed"
)

func main() {
	_ = envfile.Load(".env", "/app/.env")
	cfg := config.Load()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "help", "-h", "--help":
			printUsage()
			return
		default:
			fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
			printUsage()
			os.Exit(2)
		}
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
	hasher := identityinfra.NewBcryptPasswordHasher()

	if err := seed.RunDemo(users, merchants, admins, products, hasher, cfg.AdminBootstrapPassword); err != nil {
		log.Fatalf("seed: %v", err)
	}

	userN, _ := users.Count()
	merchantN, _ := merchants.Count()
	adminN, _ := admins.Count()
	productN, _ := products.Count()
	log.Printf("counts: users=%d merchants=%d admins=%d products=%d", userN, merchantN, adminN, productN)
	seed.LogSummary(cfg.AdminBootstrapPassword)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: seed

Inserts / ensures demo accounts and products (idempotent).

Env: same as API (.env / DB_* / ADMIN_BOOTSTRAP_PASSWORD)

Host example:
  DB_HOST=127.0.0.1 DB_SSLMODE=require go run ./cmd/seed
`)
}
