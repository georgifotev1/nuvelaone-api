package handler

import (
	"net/http"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/georgifotev1/nuvelaone-api/internal/service"
	"github.com/georgifotev1/nuvelaone-api/pkg/jsonutil"
	"github.com/georgifotev1/nuvelaone-api/pkg/validator"
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
	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.List(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}
	jsonutil.Write(w, http.StatusOK, users)
}

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

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateUserRequest
	if err := jsonutil.Read(r, &req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	user, err := h.svc.Create(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}
	jsonutil.Write(w, http.StatusCreated, jsonutil.NewResponse(user))
}

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
