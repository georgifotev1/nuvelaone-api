package handler

import (
	"net/http"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	svc "github.com/georgifotev1/nuvelaone-api/internal/service"
	"github.com/georgifotev1/nuvelaone-api/pkg/auth"
	"github.com/georgifotev1/nuvelaone-api/pkg/jsonutil"
	"github.com/georgifotev1/nuvelaone-api/pkg/validator"
	"github.com/go-chi/chi/v5"
)

type ServiceHandler struct {
	svc svc.ServiceService
}

func NewServiceHandler(svc svc.ServiceService) *ServiceHandler {
	return &ServiceHandler{svc: svc}
}

// List godoc
//
//	@Summary		List services
//	@Description	Get all services for the tenant
//	@Tags			services
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		domain.Service
//	@Failure		401	{object}	jsonutil.ErrorResponse
//	@Router			/services [get]
func (h *ServiceHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)

	services, err := h.svc.ListByTenant(ctx, claims.TenantID)
	if err != nil {
		writeError(w, err)
		return
	}

	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(services))
}

// ListPublic godoc
//
//	@Summary		List visible services
//	@Description	Get all visible services for the tenant that have at least one provider (public endpoint)
//	@Tags			public
//	@Produce		json
//	@Param			slug	path		string	true	"Tenant slug"
//	@Success		200	{array}		domain.Service
//	@Router			/p/{slug}/services [get]
func (h *ServiceHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := domain.TenantFromContext(ctx)
	if tenant == nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "tenant not found")
		return
	}

	services, err := h.svc.ListVisible(ctx, tenant.ID)
	if err != nil {
		writeError(w, err)
		return
	}

	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(services))
}

// GetProviders godoc
//
//	@Summary		Get providers for a service
//	@Description	Returns all users who can perform this service
//	@Tags			public
//	@Produce		json
//	@Param			slug	path		string	true	"Tenant slug"
//	@Param			id		path		string	true	"Service ID"
//	@Success		200	{array}		domain.User
//	@Router			/p/{slug}/services/{id}/providers [get]
func (h *ServiceHandler) GetProviders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := domain.TenantFromContext(ctx)
	if tenant == nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "tenant not found")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	providers, err := h.svc.GetProviders(ctx, tenant.ID, id)
	if err != nil {
		writeError(w, err)
		return
	}

	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(providers))
}

// GetByID godoc
//
//	@Summary		Get service by ID
//	@Description	Get a specific service by ID
//	@Tags			services
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Service ID"
//	@Success		200	{object}	domain.Service
//	@Failure		401	{object}	jsonutil.ErrorResponse
//	@Failure		404	{object}	jsonutil.ErrorResponse
//	@Router			/services/{id} [get]
func (h *ServiceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)

	id := chi.URLParam(r, "id")
	if id == "" {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	service, err := h.svc.GetByID(ctx, claims.TenantID, id)
	if err != nil {
		writeError(w, err)
		return
	}

	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(service))
}

// Create godoc
//
//	@Summary		Create service
//	@Description	Create a new service (owner/admin only)
//	@Tags			services
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		domain.ServiceRequest	true	"Create request"
//	@Success		201	{object}	domain.Service
//	@Failure		400	{object}	jsonutil.ErrorResponse
//	@Failure		401	{object}	jsonutil.ErrorResponse
//	@Failure		403	{object}	jsonutil.ErrorResponse
//	@Router			/services [post]
func (h *ServiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)

	var req domain.ServiceRequest
	if err := jsonutil.Read(r, &req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	service, err := h.svc.Create(ctx, claims.TenantID, req)
	if err != nil {
		writeError(w, err)
		return
	}

	jsonutil.Write(w, http.StatusCreated, jsonutil.NewResponse(service))
}

// Update godoc
//
//	@Summary		Update service
//	@Description	Update a service (owner/admin only)
//	@Tags			services
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Service ID"
//	@Param			request	body		domain.ServiceRequest	true	"Update request"
//	@Success		200	{object}	domain.Service
//	@Failure		400	{object}	jsonutil.ErrorResponse
//	@Failure		401	{object}	jsonutil.ErrorResponse
//	@Failure		403	{object}	jsonutil.ErrorResponse
//	@Failure		404	{object}	jsonutil.ErrorResponse
//	@Router			/services/{id} [put]
func (h *ServiceHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)

	id := chi.URLParam(r, "id")
	if id == "" {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req domain.ServiceRequest
	if err := jsonutil.Read(r, &req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	service, err := h.svc.Update(ctx, claims.TenantID, id, req)
	if err != nil {
		writeError(w, err)
		return
	}

	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(service))
}

// Delete godoc
//
//	@Summary		Delete service
//	@Description	Delete a service (owner/admin only)
//	@Tags			services
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Service ID"
//	@Success		204
//	@Failure		401	{object}	jsonutil.ErrorResponse
//	@Failure		403	{object}	jsonutil.ErrorResponse
//	@Failure		404	{object}	jsonutil.ErrorResponse
//	@Router			/services/{id} [delete]
func (h *ServiceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)

	id := chi.URLParam(r, "id")
	if id == "" {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	err := h.svc.Delete(ctx, claims.TenantID, id)
	if err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
