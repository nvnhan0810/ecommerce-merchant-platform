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

func TestNewAccount_requires_email(t *testing.T) {
	t.Parallel()
	if _, err := NewAccount(" ", "Nhan"); err != ErrInvalidEmail {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestAccount_Authenticate_should_accept_valid_password(t *testing.T) {
	t.Parallel()
	a, err := NewAccount("admin@ecomerce.local", "Admin")
	if err != nil {
		t.Fatal(err)
	}
	h := stubHasher{}
	if err := a.SetPassword(h, "Admin@123456"); err != nil {
		t.Fatal(err)
	}
	if err := a.Authenticate(h, "Admin@123456"); err != nil {
		t.Fatalf("expected auth ok, got %v", err)
	}
}

func TestAccount_Authenticate_should_reject_wrong_password(t *testing.T) {
	t.Parallel()
	a, _ := NewAccount("admin@ecomerce.local", "Admin")
	h := stubHasher{}
	_ = a.SetPassword(h, "Admin@123456")
	if err := a.Authenticate(h, "wrong-pass"); err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestParseAccountID_requires_value(t *testing.T) {
	t.Parallel()
	if _, err := ParseAccountID(" "); err != ErrInvalidAccountID {
		t.Fatalf("expected ErrInvalidAccountID, got %v", err)
	}
}
