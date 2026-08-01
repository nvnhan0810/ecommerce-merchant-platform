package infrastructure

import (
	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
)

func SeedDemoProducts(repo domain.ProductRepository) error {
	type counter interface {
		Count() (int, error)
	}
	if c, ok := repo.(counter); ok {
		n, err := c.Count()
		if err != nil {
			return err
		}
		if n > 0 {
			return nil
		}
	}

	samples := []struct {
		merchant, name, desc string
		price                int64
		stock                int
	}{
		{"merchant-demo", "Áo thun basic", "Cotton 100%", 149000, 50},
		{"merchant-demo", "Giày sneaker", "Unisex", 899000, 20},
		{"merchant-demo", "Balo laptop 15\"", "Chống nước", 459000, 15},
	}
	for _, s := range samples {
		price, err := domain.NewMoney(s.price, "VND")
		if err != nil {
			return err
		}
		p, err := domain.NewProduct(s.merchant, s.name, s.desc, price, s.stock)
		if err != nil {
			return err
		}
		if err := repo.Save(p); err != nil {
			return err
		}
	}
	return nil
}
