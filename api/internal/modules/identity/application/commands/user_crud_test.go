package commands

import (
	"context"
	"testing"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

func TestCreateUserHandler_should_create_user(t *testing.T) {
	t.Parallel()
	repo := infrastructureMem()
	h := NewCreateUserHandler(repo, stubHasher{})
	res, err := h.Handle(context.Background(), CreateUserCommand{
		Email:       "newbuyer@ecomerce.local",
		DisplayName: "New Buyer",
		Password:    "Buyer@123456",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Role != "user" || res.ID == "" {
		t.Fatalf("unexpected: %+v", res)
	}
}

func TestCreateUserHandler_should_reject_duplicate_email(t *testing.T) {
	t.Parallel()
	repo := infrastructureMem()
	h := NewCreateUserHandler(repo, stubHasher{})
	_, _ = h.Handle(context.Background(), CreateUserCommand{
		Email: "dupuser@ecomerce.local", DisplayName: "A", Password: "Buyer@123456",
	})
	_, err := h.Handle(context.Background(), CreateUserCommand{
		Email: "dupuser@ecomerce.local", DisplayName: "B", Password: "Buyer@123456",
	})
	if err != domain.ErrEmailTaken {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestUpdateAndDeleteUser(t *testing.T) {
	t.Parallel()
	repo := infrastructureMem()
	hasher := stubHasher{}
	create := NewCreateUserHandler(repo, hasher)
	created, err := create.Handle(context.Background(), CreateUserCommand{
		Email: "edituser@ecomerce.local", DisplayName: "Old", Password: "Buyer@123456",
	})
	if err != nil {
		t.Fatal(err)
	}
	id := domain.AccountID(created.ID)

	update := NewUpdateUserHandler(repo, hasher)
	updated, err := update.Handle(context.Background(), UpdateUserCommand{
		ID: id, Email: "editeduser@ecomerce.local", DisplayName: "New Name",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.DisplayName != "New Name" || updated.Email != "editeduser@ecomerce.local" {
		t.Fatalf("unexpected update: %+v", updated)
	}

	del := NewDeleteUserHandler(repo)
	if err := del.Handle(context.Background(), DeleteUserCommand{ID: id}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.FindByID(id); err != domain.ErrAccountNotFound {
		t.Fatalf("expected deleted, got %v", err)
	}
}
