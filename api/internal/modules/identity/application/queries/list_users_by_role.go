package queries

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type ListUsersByRoleQuery struct {
	Role domain.Role
}

type UserDTO struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type ListUsersByRoleHandler struct {
	repo domain.UserRepository
}

func NewListUsersByRoleHandler(repo domain.UserRepository) *ListUsersByRoleHandler {
	return &ListUsersByRoleHandler{repo: repo}
}

func (h *ListUsersByRoleHandler) Handle(_ context.Context, q ListUsersByRoleQuery) ([]UserDTO, error) {
	users, err := h.repo.ListByRole(q.Role)
	if err != nil {
		return nil, err
	}
	out := make([]UserDTO, 0, len(users))
	for _, u := range users {
		out = append(out, UserDTO{
			ID:          string(u.ID),
			Email:       u.Email,
			DisplayName: u.DisplayName,
			Role:        string(u.Role),
		})
	}
	return out, nil
}
