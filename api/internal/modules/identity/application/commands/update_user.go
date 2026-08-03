package commands

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type UpdateUserCommand struct {
	ID          domain.AccountID
	Email       string
	DisplayName string
	Password    string
}

type UpdateUserHandler struct {
	users  domain.AccountRepository
	hasher domain.PasswordHasher
}

func NewUpdateUserHandler(users domain.AccountRepository, hasher domain.PasswordHasher) *UpdateUserHandler {
	return &UpdateUserHandler{users: users, hasher: hasher}
}

func (h *UpdateUserHandler) Handle(_ context.Context, cmd UpdateUserCommand) (UserResult, error) {
	account, err := h.users.FindByID(cmd.ID)
	if err != nil {
		return UserResult{}, err
	}

	if err := account.ChangeEmail(cmd.Email); err != nil {
		return UserResult{}, err
	}
	account.Rename(cmd.DisplayName)

	existing, err := h.users.FindByEmail(account.Email)
	if err == nil && existing.ID != account.ID {
		return UserResult{}, domain.ErrEmailTaken
	}
	if err != nil && err != domain.ErrAccountNotFound {
		return UserResult{}, err
	}

	if cmd.Password != "" {
		if err := account.SetPassword(h.hasher, cmd.Password); err != nil {
			return UserResult{}, err
		}
	}

	if err := h.users.Save(account); err != nil {
		return UserResult{}, err
	}
	return toUserResult(account), nil
}
