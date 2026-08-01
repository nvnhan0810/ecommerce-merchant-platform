package infrastructure

import "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"

func SeedDemoUsers(repo domain.UserRepository) error {
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
		email, name string
		role        domain.Role
	}{
		{"buyer@ecomerce.local", "Buyer Demo", domain.RoleUser},
		{"shop@ecomerce.local", "Shop Demo", domain.RoleMerchant},
		{"admin@ecomerce.local", "Admin Demo", domain.RoleAdmin},
	}
	for _, s := range samples {
		u, err := domain.NewUser(s.email, s.name, s.role)
		if err != nil {
			return err
		}
		if err := repo.Save(u); err != nil {
			return err
		}
	}
	return nil
}
