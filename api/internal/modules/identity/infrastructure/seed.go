package infrastructure

import "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"

func SeedDemoUsers(repo domain.UserRepository, hasher domain.PasswordHasher, adminPassword string) error {
	type counter interface {
		Count() (int, error)
	}
	if c, ok := repo.(counter); ok {
		n, err := c.Count()
		if err != nil {
			return err
		}
		if n == 0 {
			samples := []struct {
				email, name, password string
				role                  domain.Role
			}{
				{"buyer@ecomerce.local", "Buyer Demo", "Buyer@123456", domain.RoleUser},
				{"shop@ecomerce.local", "Shop Demo", "Shop@123456", domain.RoleMerchant},
				{"admin@ecomerce.local", "Admin Demo", adminPassword, domain.RoleAdmin},
			}
			for _, s := range samples {
				u, err := domain.NewUser(s.email, s.name, s.role)
				if err != nil {
					return err
				}
				if err := u.SetPassword(hasher, s.password); err != nil {
					return err
				}
				if err := repo.Save(u); err != nil {
					return err
				}
			}
		}
	}

	return EnsureAdminUser(repo, hasher, "admin@ecomerce.local", "Admin Demo", adminPassword)
}

// EnsureAdminUser creates or updates the bootstrap admin password.
func EnsureAdminUser(
	repo domain.UserRepository,
	hasher domain.PasswordHasher,
	email, displayName, password string,
) error {
	existing, err := repo.FindByEmail(email)
	if err != nil && err != domain.ErrUserNotFound {
		return err
	}
	if err == domain.ErrUserNotFound {
		u, err := domain.NewUser(email, displayName, domain.RoleAdmin)
		if err != nil {
			return err
		}
		if err := u.SetPassword(hasher, password); err != nil {
			return err
		}
		return repo.Save(u)
	}
	if existing.PasswordHash == "" || existing.Role != domain.RoleAdmin {
		existing.Role = domain.RoleAdmin
		if err := existing.SetPassword(hasher, password); err != nil {
			return err
		}
		return repo.Save(existing)
	}
	return nil
}
