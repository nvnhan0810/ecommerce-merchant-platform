package infrastructure

import (
	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
	identitydomain "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type demoProduct struct {
	merchantEmail string
	name          string
	desc          string
	priceCents    int64
	stock         int
}

func demoProductCatalog() []demoProduct {
	return []demoProduct{
		{"shop@ecomerce.local", "Áo thun basic", "Cotton 100%, form regular", 149000, 50},
		{"shop@ecomerce.local", "Giày sneaker", "Unisex, đế cao su", 899000, 20},
		{"shop@ecomerce.local", "Balo laptop 15\"", "Chống nước, ngăn laptop", 459000, 15},

		{"fashion@ecomerce.local", "Đầm hoa mùa hè", "Voan nhẹ, form suông", 529000, 25},
		{"fashion@ecomerce.local", "Quần jean slim", "Denim stretch", 399000, 40},
		{"fashion@ecomerce.local", "Áo khoác denim", "Wash nhẹ", 649000, 18},

		{"tech@ecomerce.local", "Tai nghe Bluetooth", "Noise cancel cơ bản", 790000, 30},
		{"tech@ecomerce.local", "Chuột không dây", "Silent click", 259000, 60},
		{"tech@ecomerce.local", "Ốp lưng iPhone", "Trong suốt chống sốc", 99000, 100},

		{"home@ecomerce.local", "Đèn bàn LED", "3 chế độ sáng", 320000, 22},
		{"home@ecomerce.local", "Bình giữ nhiệt 500ml", "Inox 304", 189000, 45},
		{"home@ecomerce.local", "Thảm trải sàn", "Anti-slip", 275000, 12},
	}
}

// SeedDemoProducts attaches demo catalog rows to real merchant IDs (idempotent by merchant+name).
func SeedDemoProducts(products domain.ProductRepository, merchants identitydomain.AccountRepository) error {
	merchantAccounts, err := merchants.List()
	if err != nil {
		return err
	}
	byEmail := make(map[string]string, len(merchantAccounts))
	for _, m := range merchantAccounts {
		byEmail[m.Email] = string(m.ID)
	}

	existing, err := products.List(1000, 0)
	if err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(existing))
	for _, p := range existing {
		seen[p.MerchantID+"|"+p.Name] = struct{}{}
	}

	for _, s := range demoProductCatalog() {
		merchantID, ok := byEmail[s.merchantEmail]
		if !ok || merchantID == "" {
			continue
		}
		key := merchantID + "|" + s.name
		if _, exists := seen[key]; exists {
			continue
		}
		price, err := domain.NewMoney(s.priceCents, "VND")
		if err != nil {
			return err
		}
		p, err := domain.NewProduct(merchantID, s.name, s.desc, price, s.stock)
		if err != nil {
			return err
		}
		if err := products.Save(p); err != nil {
			return err
		}
		seen[key] = struct{}{}
	}
	return nil
}
