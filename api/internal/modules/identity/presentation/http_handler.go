package presentation

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/application/commands"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/application/queries"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/authctx"
)

type IdentityHandler struct {
	listByRole *queries.ListUsersByRoleHandler
	login      *commands.LoginHandler
	me         *queries.GetCurrentUserHandler
}

func NewIdentityHandler(
	listByRole *queries.ListUsersByRoleHandler,
	login *commands.LoginHandler,
	me *queries.GetCurrentUserHandler,
) *IdentityHandler {
	return &IdentityHandler{listByRole: listByRole, login: login, me: me}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *IdentityHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body loginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	res, err := h.login.Handle(r.Context(), commands.LoginCommand{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		status := http.StatusUnauthorized
		switch {
		case errors.Is(err, domain.ErrForbiddenRole):
			status = http.StatusForbidden
		case errors.Is(err, domain.ErrPasswordNotSet), errors.Is(err, domain.ErrInvalidCredentials):
			status = http.StatusUnauthorized
		case errors.Is(err, domain.ErrWeakPassword):
			status = http.StatusBadRequest
		default:
			status = http.StatusInternalServerError
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *IdentityHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := authctx.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	user, err := h.me.Handle(r.Context(), queries.GetCurrentUserQuery{UserID: claims.UserID})
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": user})
}

func (h *IdentityHandler) ListMerchants(w http.ResponseWriter, r *http.Request) {
	items, err := h.listByRole.Handle(r.Context(), queries.ListUsersByRoleQuery{Role: domain.RoleMerchant})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *IdentityHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	items, err := h.listByRole.Handle(r.Context(), queries.ListUsersByRoleQuery{Role: domain.RoleUser})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
