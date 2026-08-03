package infrastructure

import "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"

// DemoAccount is one row to ensure in users / merchants / admins.
type DemoAccount struct {
	Email       string
	DisplayName string
	Password    string
}

func DemoUsers() []DemoAccount {
	return []DemoAccount{
		{Email: "buyer@ecomerce.local", DisplayName: "Buyer Demo", Password: "Buyer@123456"},
		{Email: "an@ecomerce.local", DisplayName: "Nguyễn An", Password: "Buyer@123456"},
		{Email: "binh@ecomerce.local", DisplayName: "Trần Bình", Password: "Buyer@123456"},
		{Email: "chi@ecomerce.local", DisplayName: "Lê Chi", Password: "Buyer@123456"},
	}
}

func DemoMerchants() []DemoAccount {
	return []DemoAccount{
		{Email: "shop@ecomerce.local", DisplayName: "Shop Demo", Password: "Shop@123456"},
		{Email: "fashion@ecomerce.local", DisplayName: "Fashion House", Password: "Shop@123456"},
		{Email: "tech@ecomerce.local", DisplayName: "Tech Store", Password: "Shop@123456"},
		{Email: "home@ecomerce.local", DisplayName: "Home Living", Password: "Shop@123456"},
	}
}

func DemoAdmins(adminPassword string) []DemoAccount {
	if adminPassword == "" {
		adminPassword = "Admin@123456"
	}
	return []DemoAccount{
		{Email: "admin@ecomerce.local", DisplayName: "Admin Demo", Password: adminPassword},
		{Email: "ops@ecomerce.local", DisplayName: "Ops Admin", Password: adminPassword},
	}
}

// SeedDemoAccounts ensures demo rows in users, merchants, and admins (idempotent by email).
func SeedDemoAccounts(
	users domain.AccountRepository,
	merchants domain.AccountRepository,
	admins domain.AccountRepository,
	hasher domain.PasswordHasher,
	adminPassword string,
) error {
	for _, a := range DemoUsers() {
		if err := EnsureAccount(users, hasher, a); err != nil {
			return err
		}
	}
	for _, a := range DemoMerchants() {
		if err := EnsureAccount(merchants, hasher, a); err != nil {
			return err
		}
	}
	for _, a := range DemoAdmins(adminPassword) {
		if err := EnsureAccount(admins, hasher, a); err != nil {
			return err
		}
	}
	return nil
}

func EnsureAccount(repo domain.AccountRepository, hasher domain.PasswordHasher, demo DemoAccount) error {
	existing, err := repo.FindByEmail(demo.Email)
	if err != nil && err != domain.ErrAccountNotFound {
		return err
	}
	if err == domain.ErrAccountNotFound {
		account, err := domain.NewAccount(demo.Email, demo.DisplayName)
		if err != nil {
			return err
		}
		if err := account.SetPassword(hasher, demo.Password); err != nil {
			return err
		}
		return repo.Save(account)
	}

	changed := false
	if existing.DisplayName == "" && demo.DisplayName != "" {
		existing.Rename(demo.DisplayName)
		changed = true
	}
	if existing.PasswordHash == "" {
		if err := existing.SetPassword(hasher, demo.Password); err != nil {
			return err
		}
		changed = true
	}
	if changed {
		return repo.Save(existing)
	}
	return nil
}

// EnsureAdminAccount keeps the bootstrap admin helper used by older call sites.
func EnsureAdminAccount(
	repo domain.AccountRepository,
	hasher domain.PasswordHasher,
	email, displayName, password string,
) error {
	return EnsureAccount(repo, hasher, DemoAccount{
		Email:       email,
		DisplayName: displayName,
		Password:    password,
	})
}
