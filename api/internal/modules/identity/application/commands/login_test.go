package commands

import (
	"context"
	"testing"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

func TestLoginHandler_should_issue_token_for_admin(t *testing.T) {
	t.Parallel()
	admins := infrastructureMem()
	h := stubHasher{}
	a, err := domain.NewAccount("admin@ecomerce.local", "Admin")
	if err != nil {
		t.Fatal(err)
	}
	if err := a.SetPassword(h, "Admin@123456"); err != nil {
		t.Fatal(err)
	}
	_ = admins.Save(a)

	handler := NewLoginHandler(admins, h, stubTokens{}, domain.RoleAdmin)
	res, err := handler.Handle(context.Background(), LoginCommand{
		Email:    "admin@ecomerce.local",
		Password: "Admin@123456",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.AccessToken == "" || res.User.Role != "admin" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestLoginHandler_should_reject_unknown_admin(t *testing.T) {
	t.Parallel()
	handler := NewLoginHandler(infrastructureMem(), stubHasher{}, stubTokens{}, domain.RoleAdmin)
	_, err := handler.Handle(context.Background(), LoginCommand{
		Email:    "shop@ecomerce.local",
		Password: "Shop@123456",
	})
	if err != domain.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginHandler_should_issue_token_for_merchant(t *testing.T) {
	t.Parallel()
	merchants := infrastructureMem()
	h := stubHasher{}
	m, err := domain.NewAccount("shop@ecomerce.local", "Shop Demo")
	if err != nil {
		t.Fatal(err)
	}
	if err := m.SetPassword(h, "Shop@123456"); err != nil {
		t.Fatal(err)
	}
	_ = merchants.Save(m)

	handler := NewLoginHandler(merchants, h, stubTokens{}, domain.RoleMerchant)
	res, err := handler.Handle(context.Background(), LoginCommand{
		Email:    "shop@ecomerce.local",
		Password: "Shop@123456",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.AccessToken == "" || res.User.Role != "merchant" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestLoginHandler_merchant_should_reject_admin_email(t *testing.T) {
	t.Parallel()
	handler := NewLoginHandler(infrastructureMem(), stubHasher{}, stubTokens{}, domain.RoleMerchant)
	_, err := handler.Handle(context.Background(), LoginCommand{
		Email:    "admin@ecomerce.local",
		Password: "Admin@123456",
	})
	if err != domain.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func infrastructureMem() *memAccounts {
	return &memAccounts{byID: map[domain.AccountID]domain.Account{}, byEmail: map[string]domain.Account{}}
}

type memAccounts struct {
	byID    map[domain.AccountID]domain.Account
	byEmail map[string]domain.Account
}

func (r *memAccounts) Save(a domain.Account) error {
	if old, ok := r.byID[a.ID]; ok && old.Email != a.Email {
		delete(r.byEmail, old.Email)
	}
	r.byID[a.ID] = a
	r.byEmail[a.Email] = a
	return nil
}
func (r *memAccounts) FindByEmail(email string) (domain.Account, error) {
	a, ok := r.byEmail[email]
	if !ok {
		return domain.Account{}, domain.ErrAccountNotFound
	}
	return a, nil
}
func (r *memAccounts) FindByID(id domain.AccountID) (domain.Account, error) {
	a, ok := r.byID[id]
	if !ok {
		return domain.Account{}, domain.ErrAccountNotFound
	}
	return a, nil
}
func (r *memAccounts) List() ([]domain.Account, error) {
	out := make([]domain.Account, 0, len(r.byID))
	for _, a := range r.byID {
		out = append(out, a)
	}
	return out, nil
}
func (r *memAccounts) Delete(id domain.AccountID) error {
	a, ok := r.byID[id]
	if !ok {
		return domain.ErrAccountNotFound
	}
	delete(r.byID, id)
	delete(r.byEmail, a.Email)
	return nil
}
func (r *memAccounts) Count() (int, error) { return len(r.byID), nil }

type stubHasher struct{}

func (stubHasher) Hash(plain string) (string, error) { return "h:" + plain, nil }
func (stubHasher) Compare(hash, plain string) error {
	if hash != "h:"+plain {
		return domain.ErrInvalidCredentials
	}
	return nil
}

type stubTokens struct{}

func (stubTokens) Issue(c domain.TokenClaims) (string, error) {
	return "tok-" + string(c.UserID), nil
}
func (stubTokens) Parse(token string) (domain.TokenClaims, error) {
	return domain.TokenClaims{}, nil
}
