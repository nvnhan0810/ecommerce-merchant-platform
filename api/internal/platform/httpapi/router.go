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
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Webhook-Secret"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", deps.Health.ServeHTTP)
	r.Get("/api/health", deps.Health.ServeHTTP)

	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/products", deps.Catalog.ListProducts)
		api.Get("/products/{id}", deps.Catalog.GetProduct)
		api.Get("/media/*", deps.Catalog.ServeMedia)

		api.Post("/auth/login", deps.Identity.Login)
		api.Post("/auth/merchant/login", deps.Identity.MerchantLogin)
		api.Post("/auth/user/login", deps.Identity.UserLogin)

		api.Post("/webhooks/delivery", deps.Ordering.DeliveryWebhook)

		api.Group(func(user chi.Router) {
			user.Use(BearerAuth(deps.Tokens))
			user.Use(RequireUser)
			user.Get("/auth/user/me", deps.Identity.UserMe)
			user.Put("/auth/user/me", deps.Identity.UpdateProfile)
			user.Get("/me/addresses", deps.Identity.ListUserAddresses)
			user.Post("/me/addresses", deps.Identity.CreateUserAddress)
			user.Get("/me/addresses/{id}", deps.Identity.GetUserAddress)
			user.Put("/me/addresses/{id}", deps.Identity.UpdateUserAddress)
			user.Delete("/me/addresses/{id}", deps.Identity.DeleteUserAddress)
			user.Post("/orders", deps.Ordering.CreateUserOrder)
			user.Get("/me/orders", deps.Ordering.ListUserOrders)
			user.Get("/me/orders/{id}", deps.Ordering.GetUserOrder)
		})

		api.Group(func(merchant chi.Router) {
			merchant.Use(BearerAuth(deps.Tokens))
			merchant.Use(RequireMerchant)
			merchant.Get("/auth/merchant/me", deps.Identity.MerchantMe)
			merchant.Get("/merchant/products", deps.Catalog.ListMerchantProducts)
			merchant.Post("/merchant/products", deps.Catalog.CreateMerchantProduct)
			merchant.Get("/merchant/products/{id}", deps.Catalog.GetMerchantProduct)
			merchant.Put("/merchant/products/{id}", deps.Catalog.UpdateMerchantProduct)
			merchant.Delete("/merchant/products/{id}", deps.Catalog.DeleteMerchantProduct)
			merchant.Post("/merchant/products/{id}/image", deps.Catalog.UploadMerchantProductImage)
			merchant.Delete("/merchant/products/{id}/image", deps.Catalog.DeleteMerchantProductImage)
			merchant.Get("/merchant/orders", deps.Ordering.ListMerchantOrders)
			merchant.Get("/merchant/orders/{id}", deps.Ordering.GetMerchantOrder)
			merchant.Patch("/merchant/orders/{id}/status", deps.Ordering.UpdateMerchantOrderStatus)
		})

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
			admin.Put("/products/{id}", deps.Catalog.UpdateProduct)
			admin.Delete("/products/{id}", deps.Catalog.DeleteProduct)
			admin.Post("/products/{id}/image", deps.Catalog.UploadProductImage)
			admin.Delete("/products/{id}/image", deps.Catalog.DeleteProductImage)
			admin.Get("/orders", deps.Ordering.ListOrders)
			admin.Get("/orders/{id}", deps.Ordering.GetOrder)
			admin.Post("/orders/{id}/delivery-simulate", deps.Ordering.SimulateDelivery)
		})
	})

	return r
}
