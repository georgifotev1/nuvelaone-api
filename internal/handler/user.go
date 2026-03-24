package handler

import (
	"net/http"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/georgifotev1/nuvelaone-api/internal/service"
	"github.com/georgifotev1/nuvelaone-api/pkg/auth"
	"github.com/georgifotev1/nuvelaone-api/pkg/jsonutil"
	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) Routes(r chi.Router) {
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
}

func (h *UserHandler) MeRoutes(r chi.Router) {
	r.Get("/", h.GetMe)
}

// GetMe godoc
//
//	@Summary		Get current user
//	@Description	Get the authenticated user's profile
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200		{object}	domain.User
//	@Failure		401		{object}	jsonutil.ErrorResponse
//	@Failure		404		{object}	jsonutil.ErrorResponse
//	@Router			/me [get]
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)

	user, err := h.svc.GetByID(ctx, claims.UserID)
	if err != nil {
		handleError(w, err)
		return
	}

	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(user))
}

// List godoc
//
//	@Summary		List all users
//	@Description	Get all users (admin/owner only)
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200		{array}		domain.User
//	@Failure		401		{object}	jsonutil.ErrorResponse
//	@Failure		403		{object}	jsonutil.ErrorResponse
//	@Router			/users [get]
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.List(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}
	jsonutil.Write(w, http.StatusOK, users)
}

// GetByID godoc
//
//	@Summary		Get user by ID
//	@Description	Get a specific user by ID (admin/owner only)
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	domain.User
//	@Failure		400	{object}	jsonutil.ErrorResponse
//	@Failure		401	{object}	jsonutil.ErrorResponse
//	@Failure		403	{object}	jsonutil.ErrorResponse
//	@Failure		404	{object}	jsonutil.ErrorResponse
//	@Router			/users/{id} [get]
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	user, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(user))
}

// Update godoc
//
//	@Summary		Update user
//	@Description	Update a user's information (admin/owner only)
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"User ID"
//	@Param			request	body		domain.UpdateUserRequest	true	"Update request"
//	@Success		200		{object}	domain.User
//	@Failure		400		{object}	jsonutil.ErrorResponse
//	@Failure		401		{object}	jsonutil.ErrorResponse
//	@Failure		403		{object}	jsonutil.ErrorResponse
//	@Failure		404		{object}	jsonutil.ErrorResponse
//	@Router			/users/{id} [put]
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
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
		handleError(w, err)
		return
	}
	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(user))
}

// Delete godoc
//
//	@Summary		Delete user
//	@Description	Delete a user by ID (admin/owner only)
//	@Tags			users
//	@Security		BearerAuth
//	@Param			id	path		string	true	"User ID"
//	@Success		204	"No Content"
//	@Failure		400	{object}	jsonutil.ErrorResponse
//	@Failure		401	{object}	jsonutil.ErrorResponse
//	@Failure		403	{object}	jsonutil.ErrorResponse
//	@Failure		404	{object}	jsonutil.ErrorResponse
//	@Router			/users/{id} [delete]
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
