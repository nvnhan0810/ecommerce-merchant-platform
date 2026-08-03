package commands

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type DeleteUserCommand struct {
	ID domain.AccountID
}

type DeleteUserHandler struct {
	users domain.AccountRepository
}

func NewDeleteUserHandler(users domain.AccountRepository) *DeleteUserHandler {
	return &DeleteUserHandler{users: users}
}

func (h *DeleteUserHandler) Handle(_ context.Context, cmd DeleteUserCommand) error {
	if _, err := h.users.FindByID(cmd.ID); err != nil {
		return err
	}
	return h.users.Delete(cmd.ID)
}
