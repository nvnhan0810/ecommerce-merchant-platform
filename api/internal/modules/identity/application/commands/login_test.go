package commands

import (
	"context"
	"testing"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type memRepo struct {
	byEmail map[string]domain.User
	byID    map[domain.UserID]domain.User
}

func newMemRepo() *memRepo {
	return &memRepo{
		byEmail: map[string]domain.User{},
		byID:    map[domain.UserID]domain.User{},
	}
}

func (r *memRepo) Save(u domain.User) error {
	r.byEmail[u.Email] = u
	r.byID[u.ID] = u
	return nil
}
func (r *memRepo) FindByEmail(email string) (domain.User, error) {
	u, ok := r.byEmail[email]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return u, nil
}
func (r *memRepo) FindByID(id domain.UserID) (domain.User, error) {
	u, ok := r.byID[id]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return u, nil
}
func (r *memRepo) ListByRole(role domain.Role) ([]domain.User, error) {
	out := []domain.User{}
	for _, u := range r.byEmail {
		if u.Role == role {
			out = append(out, u)
		}
	}
	return out, nil
}

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

func TestLoginHandler_should_issue_token_for_admin(t *testing.T) {
	t.Parallel()
	repo := newMemRepo()
	h := stubHasher{}
	u, err := domain.NewUser("admin@ecomerce.local", "Admin", domain.RoleAdmin)
	if err != nil {
		t.Fatal(err)
	}
	if err := u.SetPassword(h, "Admin@123456"); err != nil {
		t.Fatal(err)
	}
	_ = repo.Save(u)

	handler := NewLoginHandler(repo, h, stubTokens{})
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

func TestLoginHandler_should_reject_merchant(t *testing.T) {
	t.Parallel()
	repo := newMemRepo()
	h := stubHasher{}
	u, _ := domain.NewUser("shop@ecomerce.local", "Shop", domain.RoleMerchant)
	_ = u.SetPassword(h, "Shop@123456")
	_ = repo.Save(u)

	handler := NewLoginHandler(repo, h, stubTokens{})
	_, err := handler.Handle(context.Background(), LoginCommand{
		Email:    "shop@ecomerce.local",
		Password: "Shop@123456",
	})
	if err != domain.ErrForbiddenRole {
		t.Fatalf("expected ErrForbiddenRole, got %v", err)
	}
}
