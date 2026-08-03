package commands

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type CreateUserCommand struct {
	Email       string
	DisplayName string
	Password    string
}

type UserResult struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type CreateUserHandler struct {
	users  domain.AccountRepository
	hasher domain.PasswordHasher
}

func NewCreateUserHandler(users domain.AccountRepository, hasher domain.PasswordHasher) *CreateUserHandler {
	return &CreateUserHandler{users: users, hasher: hasher}
}

func (h *CreateUserHandler) Handle(_ context.Context, cmd CreateUserCommand) (UserResult, error) {
	if _, err := h.users.FindByEmail(cmd.Email); err == nil {
		return UserResult{}, domain.ErrEmailTaken
	} else if err != domain.ErrAccountNotFound {
		return UserResult{}, err
	}

	account, err := domain.NewAccount(cmd.Email, cmd.DisplayName)
	if err != nil {
		return UserResult{}, err
	}
	if err := account.SetPassword(h.hasher, cmd.Password); err != nil {
		return UserResult{}, err
	}
	if err := h.users.Save(account); err != nil {
		return UserResult{}, err
	}
	return toUserResult(account), nil
}

func toUserResult(account domain.Account) UserResult {
	return UserResult{
		ID:          string(account.ID),
		Email:       account.Email,
		DisplayName: account.DisplayName,
		Role:        string(domain.RoleUser),
	}
}
