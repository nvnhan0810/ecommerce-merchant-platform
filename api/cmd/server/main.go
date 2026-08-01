package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/application/commands"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/application/queries"
	cataloginfra "github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/infrastructure"
	catalogpres "github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/presentation"
	healthpres "github.com/nvnhan0810/ecomerce-api/internal/modules/health/presentation"
	identityqueries "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/application/queries"
	identityinfra "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/infrastructure"
	identitypres "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/presentation"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/config"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/envfile"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/httpapi"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/postgres"
)

func main() {
	_ = envfile.Load(".env", "/app/.env")

	cfg := config.Load()
	ctx := context.Background()

	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	if cfg.AutoMigrate {
		if err := postgres.Migrate(ctx, pool); err != nil {
			log.Fatalf("migrate: %v", err)
		}
	}

	productRepo := cataloginfra.NewPostgresProductRepository(pool)
	userRepo := identityinfra.NewPostgresUserRepository(pool)

	if cfg.Seed {
		if err := cataloginfra.SeedDemoProducts(productRepo); err != nil {
			log.Fatalf("seed products: %v", err)
		}
		if err := identityinfra.SeedDemoUsers(userRepo); err != nil {
			log.Fatalf("seed users: %v", err)
		}
	}

	listProducts := queries.NewListProductsHandler(productRepo)
	createProduct := commands.NewCreateProductHandler(productRepo)
	listUsers := identityqueries.NewListUsersByRoleHandler(userRepo)

	router := httpapi.NewRouter(httpapi.Dependencies{
		Config:   cfg,
		Health:   healthpres.NewHealthHandler(cfg.Env),
		Catalog:  catalogpres.NewCatalogHandler(listProducts, createProduct),
		Identity: identitypres.NewIdentityHandler(listUsers),
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("ecomerce-api listening on %s (%s)", cfg.HTTPAddr, cfg.Env)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}
