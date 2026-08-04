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

// LoginHandler authenticates against a single account portal (admin or merchant).
type LoginHandler struct {
	accounts domain.AccountRepository
	hasher   domain.PasswordHasher
	tokens   domain.TokenService
	role     domain.Role
}

func NewLoginHandler(
	accounts domain.AccountRepository,
	hasher domain.PasswordHasher,
	tokens domain.TokenService,
	role domain.Role,
) *LoginHandler {
	return &LoginHandler{accounts: accounts, hasher: hasher, tokens: tokens, role: role}
}

func (h *LoginHandler) Handle(_ context.Context, cmd LoginCommand) (LoginResult, error) {
	account, err := h.accounts.FindByEmail(cmd.Email)
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
		Role:   h.role,
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
			Role:        string(h.role),
		},
	}, nil
}
