package infrastructure

import (
	"fmt"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
	identitydomain "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type DemoProduct struct {
	MerchantEmail string
	Name          string
	Description   string
	PriceCents    int64
	Stock         int
}

// DemoProducts is the catalog dataset seeded for each demo merchant (by email).
func DemoProducts() []DemoProduct {
	return []DemoProduct{
		{MerchantEmail: "shop@ecomerce.local", Name: "Áo thun basic", Description: "Cotton 100%, form regular", PriceCents: 149000, Stock: 50},
		{MerchantEmail: "shop@ecomerce.local", Name: "Giày sneaker", Description: "Unisex, đế cao su", PriceCents: 899000, Stock: 20},
		{MerchantEmail: "shop@ecomerce.local", Name: "Balo laptop 15\"", Description: "Chống nước, ngăn laptop", PriceCents: 459000, Stock: 15},
		{MerchantEmail: "shop@ecomerce.local", Name: "Mũ lưỡi trai", Description: "Thêu logo, freesize", PriceCents: 129000, Stock: 80},
		{MerchantEmail: "shop@ecomerce.local", Name: "Tất thể thao (3 đôi)", Description: "Thoáng khí", PriceCents: 89000, Stock: 120},

		{MerchantEmail: "fashion@ecomerce.local", Name: "Đầm hoa mùa hè", Description: "Voan nhẹ, form suông", PriceCents: 529000, Stock: 25},
		{MerchantEmail: "fashion@ecomerce.local", Name: "Quần jean slim", Description: "Denim stretch", PriceCents: 399000, Stock: 40},
		{MerchantEmail: "fashion@ecomerce.local", Name: "Áo khoác denim", Description: "Wash nhẹ", PriceCents: 649000, Stock: 18},
		{MerchantEmail: "fashion@ecomerce.local", Name: "Chân váy midi", Description: "Xếp ly, lưng thun", PriceCents: 359000, Stock: 30},
		{MerchantEmail: "fashion@ecomerce.local", Name: "Áo sơ mi linen", Description: "Mát, dễ ủi", PriceCents: 289000, Stock: 35},

		{MerchantEmail: "tech@ecomerce.local", Name: "Tai nghe Bluetooth", Description: "Noise cancel cơ bản", PriceCents: 790000, Stock: 30},
		{MerchantEmail: "tech@ecomerce.local", Name: "Chuột không dây", Description: "Silent click", PriceCents: 259000, Stock: 60},
		{MerchantEmail: "tech@ecomerce.local", Name: "Ốp lưng iPhone", Description: "Trong suốt chống sốc", PriceCents: 99000, Stock: 100},
		{MerchantEmail: "tech@ecomerce.local", Name: "Sạc nhanh 30W", Description: "USB-C PD", PriceCents: 349000, Stock: 45},
		{MerchantEmail: "tech@ecomerce.local", Name: "Hub USB-C 6-in-1", Description: "HDMI + USB 3.0", PriceCents: 520000, Stock: 22},

		{MerchantEmail: "home@ecomerce.local", Name: "Đèn bàn LED", Description: "3 chế độ sáng", PriceCents: 320000, Stock: 22},
		{MerchantEmail: "home@ecomerce.local", Name: "Bình giữ nhiệt 500ml", Description: "Inox 304", PriceCents: 189000, Stock: 45},
		{MerchantEmail: "home@ecomerce.local", Name: "Thảm trải sàn", Description: "Anti-slip", PriceCents: 275000, Stock: 12},
		{MerchantEmail: "home@ecomerce.local", Name: "Gối memory foam", Description: "Vỏ tháo giặt", PriceCents: 410000, Stock: 28},
		{MerchantEmail: "home@ecomerce.local", Name: "Bộ ly thủy tinh (4 cái)", Description: "Trong suốt", PriceCents: 159000, Stock: 50},
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

	seen, err := existingProductKeys(products)
	if err != nil {
		return err
	}

	created := 0
	for _, s := range DemoProducts() {
		merchantID, ok := byEmail[s.MerchantEmail]
		if !ok || merchantID == "" {
			return fmt.Errorf("seed product %q: merchant %q not found (seed accounts first)", s.Name, s.MerchantEmail)
		}
		key := merchantID + "|" + s.Name
		if _, exists := seen[key]; exists {
			continue
		}
		price, err := domain.NewMoney(s.PriceCents, "VND")
		if err != nil {
			return err
		}
		p, err := domain.NewProduct(merchantID, s.Name, s.Description, price, s.Stock)
		if err != nil {
			return err
		}
		if err := products.Save(p); err != nil {
			return err
		}
		seen[key] = struct{}{}
		created++
	}
	_ = created
	return nil
}

func existingProductKeys(products domain.ProductRepository) (map[string]struct{}, error) {
	seen := make(map[string]struct{})
	const page = 200
	offset := 0
	for {
		batch, err := products.List(page, offset)
		if err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			break
		}
		for _, p := range batch {
			seen[p.MerchantID+"|"+p.Name] = struct{}{}
		}
		if len(batch) < page {
			break
		}
		offset += page
	}
	return seen, nil
}
