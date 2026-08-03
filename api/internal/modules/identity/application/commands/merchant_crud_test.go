package commands

import (
	"context"
	"testing"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

func TestCreateMerchantHandler_should_create_merchant(t *testing.T) {
	t.Parallel()
	repo := infrastructureMem()
	h := NewCreateMerchantHandler(repo, stubHasher{})
	res, err := h.Handle(context.Background(), CreateMerchantCommand{
		Email:       "newshop@ecomerce.local",
		DisplayName: "New Shop",
		Password:    "Shop@123456",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Role != "merchant" || res.ID == "" {
		t.Fatalf("unexpected: %+v", res)
	}
}

func TestCreateMerchantHandler_should_reject_duplicate_email(t *testing.T) {
	t.Parallel()
	repo := infrastructureMem()
	h := NewCreateMerchantHandler(repo, stubHasher{})
	_, _ = h.Handle(context.Background(), CreateMerchantCommand{
		Email: "dup@ecomerce.local", DisplayName: "A", Password: "Shop@123456",
	})
	_, err := h.Handle(context.Background(), CreateMerchantCommand{
		Email: "dup@ecomerce.local", DisplayName: "B", Password: "Shop@123456",
	})
	if err != domain.ErrEmailTaken {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestUpdateAndDeleteMerchant(t *testing.T) {
	t.Parallel()
	repo := infrastructureMem()
	hasher := stubHasher{}
	create := NewCreateMerchantHandler(repo, hasher)
	created, err := create.Handle(context.Background(), CreateMerchantCommand{
		Email: "edit@ecomerce.local", DisplayName: "Old", Password: "Shop@123456",
	})
	if err != nil {
		t.Fatal(err)
	}
	id := domain.AccountID(created.ID)

	update := NewUpdateMerchantHandler(repo, hasher)
	updated, err := update.Handle(context.Background(), UpdateMerchantCommand{
		ID: id, Email: "edited@ecomerce.local", DisplayName: "New Name",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.DisplayName != "New Name" || updated.Email != "edited@ecomerce.local" {
		t.Fatalf("unexpected update: %+v", updated)
	}

	del := NewDeleteMerchantHandler(repo)
	if err := del.Handle(context.Background(), DeleteMerchantCommand{ID: id}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.FindByID(id); err != domain.ErrAccountNotFound {
		t.Fatalf("expected deleted, got %v", err)
	}
}
