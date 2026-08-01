package presentation

import (
	"encoding/json"
	"net/http"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/application/queries"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type IdentityHandler struct {
	listByRole *queries.ListUsersByRoleHandler
}

func NewIdentityHandler(listByRole *queries.ListUsersByRoleHandler) *IdentityHandler {
	return &IdentityHandler{listByRole: listByRole}
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
