package domain

import "testing"

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
