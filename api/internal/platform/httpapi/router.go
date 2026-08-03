package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	catalogpres "github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/presentation"
	healthpres "github.com/nvnhan0810/ecomerce-api/internal/modules/health/presentation"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
	identitypres "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/presentation"
	orderingpres "github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/presentation"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/config"
)

type Dependencies struct {
	Config   config.Config
	Health   *healthpres.HealthHandler
	Catalog  *catalogpres.CatalogHandler
	Identity *identitypres.IdentityHandler
	Ordering *orderingpres.OrderingHandler
	Tokens   domain.TokenService
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
		api.Get("/media/*", deps.Catalog.ServeMedia)

		api.Post("/auth/login", deps.Identity.Login)

		api.Group(func(admin chi.Router) {
			admin.Use(BearerAuth(deps.Tokens))
			admin.Use(RequireAdmin)
			admin.Get("/auth/me", deps.Identity.Me)
			admin.Get("/merchants", deps.Identity.ListMerchants)
			admin.Post("/merchants", deps.Identity.CreateMerchant)
			admin.Get("/merchants/{id}", deps.Identity.GetMerchant)
			admin.Put("/merchants/{id}", deps.Identity.UpdateMerchant)
			admin.Delete("/merchants/{id}", deps.Identity.DeleteMerchant)
			admin.Get("/users", deps.Identity.ListUsers)
			admin.Post("/users", deps.Identity.CreateUser)
			admin.Get("/users/{id}", deps.Identity.GetUser)
			admin.Put("/users/{id}", deps.Identity.UpdateUser)
			admin.Delete("/users/{id}", deps.Identity.DeleteUser)
			admin.Post("/products", deps.Catalog.CreateProduct)
			admin.Get("/products/{id}", deps.Catalog.GetProduct)
			admin.Put("/products/{id}", deps.Catalog.UpdateProduct)
			admin.Delete("/products/{id}", deps.Catalog.DeleteProduct)
			admin.Post("/products/{id}/image", deps.Catalog.UploadProductImage)
			admin.Delete("/products/{id}/image", deps.Catalog.DeleteProductImage)
			admin.Get("/orders", deps.Ordering.ListOrders)
			admin.Get("/orders/{id}", deps.Ordering.GetOrder)
			admin.Patch("/orders/{id}/status", deps.Ordering.UpdateOrderStatus)
		})
	})

	return r
}
