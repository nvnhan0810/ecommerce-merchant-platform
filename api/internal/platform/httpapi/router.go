package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	catalogpres "github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/presentation"
	healthpres "github.com/nvnhan0810/ecomerce-api/internal/modules/health/presentation"
	identitypres "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/presentation"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/config"
)

type Dependencies struct {
	Config   config.Config
	Health   *healthpres.HealthHandler
	Catalog  *catalogpres.CatalogHandler
	Identity *identitypres.IdentityHandler
}

func NewRouter(deps Dependencies) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   deps.Config.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", deps.Health.ServeHTTP)
	r.Get("/api/health", deps.Health.ServeHTTP)

	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/products", deps.Catalog.ListProducts)
		api.Post("/products", deps.Catalog.CreateProduct)
		api.Get("/merchants", deps.Identity.ListMerchants)
		api.Get("/users", deps.Identity.ListUsers)
	})

	return r
}
