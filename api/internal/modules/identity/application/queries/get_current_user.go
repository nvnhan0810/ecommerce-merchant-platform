package queries

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type GetCurrentUserQuery struct {
	UserID domain.UserID
}

type CurrentUserDTO struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type GetCurrentUserHandler struct {
	repo domain.UserRepository
}

func NewGetCurrentUserHandler(repo domain.UserRepository) *GetCurrentUserHandler {
	return &GetCurrentUserHandler{repo: repo}
}

func (h *GetCurrentUserHandler) Handle(_ context.Context, q GetCurrentUserQuery) (CurrentUserDTO, error) {
	user, err := h.repo.FindByID(q.UserID)
	if err != nil {
		return CurrentUserDTO{}, err
	}
	return CurrentUserDTO{
		ID:          string(user.ID),
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        string(user.Role),
	}, nil
}
