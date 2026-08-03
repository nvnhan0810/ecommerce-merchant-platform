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
	identitycommands "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/application/commands"
	identityqueries "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/application/queries"
	identityinfra "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/infrastructure"
	identitypres "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/presentation"
	orderinginfra "github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/infrastructure"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/config"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/envfile"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/httpapi"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/postgres"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/seed"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/storage"
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
		if err := postgres.Migrate(ctx, cfg.DatabaseURL); err != nil {
			log.Fatalf("migrate: %v", err)
		}
	}

	productRepo := cataloginfra.NewPostgresProductRepository(pool)
	orderRepo := orderinginfra.NewPostgresOrderRepository(pool)
	userRepo := identityinfra.NewPostgresUserRepository(pool)
	merchantRepo := identityinfra.NewPostgresMerchantRepository(pool)
	adminRepo := identityinfra.NewPostgresAdminRepository(pool)
	hasher := identityinfra.NewBcryptPasswordHasher()
	tokens, err := identityinfra.NewJWTTokenService(cfg.JWTSecret, cfg.JWTTTL)
	if err != nil {
		log.Fatalf("jwt: %v", err)
	}

	if cfg.Seed {
		if err := seed.RunDemo(userRepo, merchantRepo, adminRepo, productRepo, orderRepo, hasher, cfg.AdminBootstrapPassword); err != nil {
			log.Fatalf("seed: %v", err)
		}
	}

	listProducts := queries.NewListProductsHandler(productRepo, cfg.PublicAPIBase)
	merchantChecker := cataloginfra.NewAccountMerchantChecker(merchantRepo)
	getProduct := queries.NewGetProductHandler(productRepo, cfg.PublicAPIBase)
	createProduct := commands.NewCreateProductHandler(productRepo, merchantChecker, cfg.PublicAPIBase)
	updateProduct := commands.NewUpdateProductHandler(productRepo, merchantChecker, cfg.PublicAPIBase)

	var objectStore storage.ObjectStore = storage.NopStore{}
	if s3, err := storage.NewS3Store(cfg.S3); err != nil {
		log.Printf("object storage disabled: %v", err)
	} else {
		objectStore = s3
		log.Printf("object storage ready bucket=%s endpoint=%s", cfg.S3.Bucket, cfg.S3.Endpoint)
	}
	deleteProduct := commands.NewDeleteProductHandler(productRepo, objectStore)
	uploadImage := commands.NewUploadProductImageHandler(productRepo, objectStore, cfg.PublicAPIBase)
	deleteImage := commands.NewDeleteProductImageHandler(productRepo, objectStore, cfg.PublicAPIBase)

	router := httpapi.NewRouter(httpapi.Dependencies{
		Config: cfg,
		Health: healthpres.NewHealthHandler(cfg.Env),
		Catalog: catalogpres.NewCatalogHandler(
			listProducts, getProduct, createProduct, updateProduct, deleteProduct,
			uploadImage, deleteImage, objectStore,
		),
		Identity: identitypres.NewIdentityHandler(
			identityqueries.NewListUsersHandler(userRepo),
			identityqueries.NewListMerchantsHandler(merchantRepo),
			identityqueries.NewGetUserHandler(userRepo),
			identityqueries.NewGetMerchantHandler(merchantRepo),
			identitycommands.NewLoginHandler(adminRepo, hasher, tokens),
			identityqueries.NewGetCurrentUserHandler(adminRepo),
			identitycommands.NewCreateUserHandler(userRepo, hasher),
			identitycommands.NewUpdateUserHandler(userRepo, hasher),
			identitycommands.NewDeleteUserHandler(userRepo),
			identitycommands.NewCreateMerchantHandler(merchantRepo, hasher),
			identitycommands.NewUpdateMerchantHandler(merchantRepo, hasher),
			identitycommands.NewDeleteMerchantHandler(merchantRepo),
		),
		Tokens: tokens,
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
