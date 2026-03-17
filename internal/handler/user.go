package handler

import (
	"net/http"
	"strconv"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/georgifotev1/nuvelaone-api/internal/service"
	"github.com/georgifotev1/nuvelaone-api/pkg/jsonutil"
	"github.com/go-chi/chi/v5"
)

// UserHandler handles HTTP requests for the user resource.
type UserHandler struct {
	svc service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Routes registers all user routes on the provided router.
func (h *UserHandler) Routes(r chi.Router) {
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
}

// List godoc
// @Summary  List all users
// @Tags     users
// @Produce  json
// @Success  200 {array} domain.User
// @Router   /users [get]
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.List(r.Context())
	if err != nil {
		jsonutil.WriteError(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	jsonutil.Write(w, http.StatusOK, users)
}

// GetByID godoc
// @Summary  Get user by ID
// @Tags     users
// @Produce  json
// @Param    id path int true "User ID"
// @Success  200 {object} domain.User
// @Router   /users/{id} [get]
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	user, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		jsonutil.WriteError(w, http.StatusNotFound, "user not found")
		return
	}
	jsonutil.Write(w, http.StatusOK, user)
}

// Create godoc
// @Summary  Create a user
// @Tags     users
// @Accept   json
// @Produce  json
// @Param    body body domain.CreateUserRequest true "User payload"
// @Success  201 {object} domain.User
// @Router   /users [post]
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateUserRequest
	if err := jsonutil.Read(r, &req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.svc.Create(r.Context(), req)
	if err != nil {
		if err == service.ErrConflict {
			jsonutil.WriteError(w, http.StatusConflict, "email already in use")
			return
		}
		jsonutil.WriteError(w, http.StatusInternalServerError, "failed to create user")
		return
	}
	jsonutil.Write(w, http.StatusCreated, user)
}

// Update godoc
// @Summary  Update a user
// @Tags     users
// @Accept   json
// @Produce  json
// @Param    id   path int                       true "User ID"
// @Param    body body domain.UpdateUserRequest  true "Update payload"
// @Success  200 {object} domain.User
// @Router   /users/{id} [put]
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req domain.UpdateUserRequest
	if err := jsonutil.Read(r, &req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		if err == service.ErrNotFound {
			jsonutil.WriteError(w, http.StatusNotFound, "user not found")
			return
		}
		jsonutil.WriteError(w, http.StatusInternalServerError, "failed to update user")
		return
	}
	jsonutil.Write(w, http.StatusOK, user)
}

// Delete godoc
// @Summary  Delete a user
// @Tags     users
// @Param    id path int true "User ID"
// @Success  204
// @Router   /users/{id} [delete]
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		jsonutil.WriteError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}
