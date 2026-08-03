package domain

import "testing"

type stubHasher struct{}

func (stubHasher) Hash(plain string) (string, error) { return "hash:" + plain, nil }
func (stubHasher) Compare(hash, plain string) error {
	if hash != "hash:"+plain {
		return ErrInvalidCredentials
	}
	return nil
}

func TestParseRole_should_accept_known_roles(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{"user", "merchant", "admin", "ADMIN"} {
		if _, err := ParseRole(raw); err != nil {
			t.Fatalf("ParseRole(%q): %v", raw, err)
		}
	}
}

func TestParseRole_should_reject_unknown(t *testing.T) {
	t.Parallel()
	if _, err := ParseRole("guest"); err != ErrInvalidRole {
		t.Fatalf("expected ErrInvalidRole, got %v", err)
	}
}

func TestNewUser_requires_email(t *testing.T) {
	t.Parallel()
	if _, err := NewUser(" ", "Nhan", RoleUser); err != ErrInvalidEmail {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestUser_Authenticate_should_accept_valid_password(t *testing.T) {
	t.Parallel()
	u, err := NewUser("admin@ecomerce.local", "Admin", RoleAdmin)
	if err != nil {
		t.Fatal(err)
	}
	h := stubHasher{}
	if err := u.SetPassword(h, "Admin@123456"); err != nil {
		t.Fatal(err)
	}
	if err := u.Authenticate(h, "Admin@123456"); err != nil {
		t.Fatalf("expected auth ok, got %v", err)
	}
}

func TestUser_Authenticate_should_reject_wrong_password(t *testing.T) {
	t.Parallel()
	u, err := NewUser("admin@ecomerce.local", "Admin", RoleAdmin)
	if err != nil {
		t.Fatal(err)
	}
	h := stubHasher{}
	_ = u.SetPassword(h, "Admin@123456")
	if err := u.Authenticate(h, "wrong-pass"); err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUser_RequireRole_admin(t *testing.T) {
	t.Parallel()
	u, _ := NewUser("a@b.c", "A", RoleUser)
	if err := u.RequireRole(RoleAdmin); err != ErrForbiddenRole {
		t.Fatalf("expected ErrForbiddenRole, got %v", err)
	}
}
