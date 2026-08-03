package commands

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type LoginCommand struct {
	Email    string
	Password string
}

type LoginResult struct {
	AccessToken string  `json:"access_token"`
	TokenType   string  `json:"token_type"`
	User        UserDTO `json:"user"`
}

type UserDTO struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type LoginHandler struct {
	admins domain.AccountRepository
	hasher domain.PasswordHasher
	tokens domain.TokenService
}

func NewLoginHandler(
	admins domain.AccountRepository,
	hasher domain.PasswordHasher,
	tokens domain.TokenService,
) *LoginHandler {
	return &LoginHandler{admins: admins, hasher: hasher, tokens: tokens}
}

func (h *LoginHandler) Handle(_ context.Context, cmd LoginCommand) (LoginResult, error) {
	account, err := h.admins.FindByEmail(cmd.Email)
	if err != nil {
		if err == domain.ErrAccountNotFound {
			return LoginResult{}, domain.ErrInvalidCredentials
		}
		return LoginResult{}, err
	}
	if err := account.Authenticate(h.hasher, cmd.Password); err != nil {
		return LoginResult{}, err
	}
	token, err := h.tokens.Issue(domain.TokenClaims{
		UserID: account.ID,
		Email:  account.Email,
		Role:   domain.RoleAdmin,
	})
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{
		AccessToken: token,
		TokenType:   "Bearer",
		User: UserDTO{
			ID:          string(account.ID),
			Email:       account.Email,
			DisplayName: account.DisplayName,
			Role:        string(domain.RoleAdmin),
		},
	}, nil
}
