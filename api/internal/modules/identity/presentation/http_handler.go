package presentation

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/application/commands"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/application/queries"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/authctx"
)

type IdentityHandler struct {
	listUsers      *queries.ListAccountsHandler
	listMerchants  *queries.ListAccountsHandler
	getUser        *queries.GetUserHandler
	getMerchant    *queries.GetMerchantHandler
	login          *commands.LoginHandler
	merchantLogin  *commands.LoginHandler
	me             *queries.GetCurrentUserHandler
	merchantMe     *queries.GetCurrentUserHandler
	createUser     *commands.CreateUserHandler
	updateUser     *commands.UpdateUserHandler
	deleteUser     *commands.DeleteUserHandler
	createMerchant *commands.CreateMerchantHandler
	updateMerchant *commands.UpdateMerchantHandler
	deleteMerchant *commands.DeleteMerchantHandler
}

func NewIdentityHandler(
	listUsers *queries.ListAccountsHandler,
	listMerchants *queries.ListAccountsHandler,
	getUser *queries.GetUserHandler,
	getMerchant *queries.GetMerchantHandler,
	login *commands.LoginHandler,
	merchantLogin *commands.LoginHandler,
	me *queries.GetCurrentUserHandler,
	merchantMe *queries.GetCurrentUserHandler,
	createUser *commands.CreateUserHandler,
	updateUser *commands.UpdateUserHandler,
	deleteUser *commands.DeleteUserHandler,
	createMerchant *commands.CreateMerchantHandler,
	updateMerchant *commands.UpdateMerchantHandler,
	deleteMerchant *commands.DeleteMerchantHandler,
) *IdentityHandler {
	return &IdentityHandler{
		listUsers:      listUsers,
		listMerchants:  listMerchants,
		getUser:        getUser,
		getMerchant:    getMerchant,
		login:          login,
		merchantLogin:  merchantLogin,
		me:             me,
		merchantMe:     merchantMe,
		createUser:     createUser,
		updateUser:     updateUser,
		deleteUser:     deleteUser,
		createMerchant: createMerchant,
		updateMerchant: updateMerchant,
		deleteMerchant: deleteMerchant,
	}
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
		writeIdentityError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *IdentityHandler) MerchantLogin(w http.ResponseWriter, r *http.Request) {
	var body loginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	res, err := h.merchantLogin.Handle(r.Context(), commands.LoginCommand{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		writeIdentityError(w, err)
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
		writeIdentityError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": user})
}

func (h *IdentityHandler) MerchantMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := authctx.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	user, err := h.merchantMe.Handle(r.Context(), queries.GetCurrentUserQuery{UserID: claims.UserID})
	if err != nil {
		writeIdentityError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": user})
}

func (h *IdentityHandler) ListMerchants(w http.ResponseWriter, r *http.Request) {
	items, err := h.listMerchants.Handle(r.Context())
	if err != nil {
		writeIdentityError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

type accountBody struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

func (h *IdentityHandler) CreateMerchant(w http.ResponseWriter, r *http.Request) {
	var body accountBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	res, err := h.createMerchant.Handle(r.Context(), commands.CreateMerchantCommand{
		Email:       body.Email,
		DisplayName: body.DisplayName,
		Password:    body.Password,
	})
	if err != nil {
		writeIdentityError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": res})
}

func (h *IdentityHandler) GetMerchant(w http.ResponseWriter, r *http.Request) {
	id, err := domain.ParseAccountID(chi.URLParam(r, "id"))
	if err != nil {
		writeIdentityError(w, err)
		return
	}
	item, err := h.getMerchant.Handle(r.Context(), queries.GetMerchantQuery{ID: id})
	if err != nil {
		writeIdentityError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *IdentityHandler) UpdateMerchant(w http.ResponseWriter, r *http.Request) {
	id, err := domain.ParseAccountID(chi.URLParam(r, "id"))
	if err != nil {
		writeIdentityError(w, err)
		return
	}
	var body accountBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	res, err := h.updateMerchant.Handle(r.Context(), commands.UpdateMerchantCommand{
		ID:          id,
		Email:       body.Email,
		DisplayName: body.DisplayName,
		Password:    body.Password,
	})
	if err != nil {
		writeIdentityError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": res})
}

func (h *IdentityHandler) DeleteMerchant(w http.ResponseWriter, r *http.Request) {
	id, err := domain.ParseAccountID(chi.URLParam(r, "id"))
	if err != nil {
		writeIdentityError(w, err)
		return
	}
	if err := h.deleteMerchant.Handle(r.Context(), commands.DeleteMerchantCommand{ID: id}); err != nil {
		writeIdentityError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *IdentityHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	items, err := h.listUsers.Handle(r.Context())
	if err != nil {
		writeIdentityError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *IdentityHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var body accountBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	res, err := h.createUser.Handle(r.Context(), commands.CreateUserCommand{
		Email:       body.Email,
		DisplayName: body.DisplayName,
		Password:    body.Password,
	})
	if err != nil {
		writeIdentityError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": res})
}

func (h *IdentityHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := domain.ParseAccountID(chi.URLParam(r, "id"))
	if err != nil {
		writeIdentityError(w, err)
		return
	}
	item, err := h.getUser.Handle(r.Context(), queries.GetUserQuery{ID: id})
	if err != nil {
		writeIdentityError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *IdentityHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := domain.ParseAccountID(chi.URLParam(r, "id"))
	if err != nil {
		writeIdentityError(w, err)
		return
	}
	var body accountBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	res, err := h.updateUser.Handle(r.Context(), commands.UpdateUserCommand{
		ID:          id,
		Email:       body.Email,
		DisplayName: body.DisplayName,
		Password:    body.Password,
	})
	if err != nil {
		writeIdentityError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": res})
}

func (h *IdentityHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := domain.ParseAccountID(chi.URLParam(r, "id"))
	if err != nil {
		writeIdentityError(w, err)
		return
	}
	if err := h.deleteUser.Handle(r.Context(), commands.DeleteUserCommand{ID: id}); err != nil {
		writeIdentityError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeIdentityError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, domain.ErrAccountNotFound):
		status = http.StatusNotFound
	case errors.Is(err, domain.ErrEmailTaken):
		status = http.StatusConflict
	case errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrWeakPassword),
		errors.Is(err, domain.ErrInvalidAccountID):
		status = http.StatusBadRequest
	case errors.Is(err, domain.ErrInvalidCredentials), errors.Is(err, domain.ErrPasswordNotSet):
		status = http.StatusUnauthorized
	}
	writeError(w, status, err.Error())
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
