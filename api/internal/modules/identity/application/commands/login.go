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
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	User        UserDTO `json:"user"`
}

type UserDTO struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type LoginHandler struct {
	repo   domain.UserRepository
	hasher domain.PasswordHasher
	tokens domain.TokenService
}

func NewLoginHandler(
	repo domain.UserRepository,
	hasher domain.PasswordHasher,
	tokens domain.TokenService,
) *LoginHandler {
	return &LoginHandler{repo: repo, hasher: hasher, tokens: tokens}
}

func (h *LoginHandler) Handle(_ context.Context, cmd LoginCommand) (LoginResult, error) {
	user, err := h.repo.FindByEmail(cmd.Email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return LoginResult{}, domain.ErrInvalidCredentials
		}
		return LoginResult{}, err
	}
	if err := user.Authenticate(h.hasher, cmd.Password); err != nil {
		return LoginResult{}, err
	}
	if err := user.RequireRole(domain.RoleAdmin); err != nil {
		return LoginResult{}, err
	}
	token, err := h.tokens.Issue(domain.TokenClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
	})
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{
		AccessToken: token,
		TokenType:   "Bearer",
		User: UserDTO{
			ID:          string(user.ID),
			Email:       user.Email,
			DisplayName: user.DisplayName,
			Role:        string(user.Role),
		},
	}, nil
}
