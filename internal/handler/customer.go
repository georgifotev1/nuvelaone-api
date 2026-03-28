package handler

import (
	"net/http"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	svc "github.com/georgifotev1/nuvelaone-api/internal/service"
	"github.com/georgifotev1/nuvelaone-api/pkg/auth"
	"github.com/georgifotev1/nuvelaone-api/pkg/jsonutil"
	"github.com/georgifotev1/nuvelaone-api/pkg/validator"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type CustomerHandler struct {
	svc    svc.CustomerService
	logger *zap.SugaredLogger
}

func NewCustomerHandler(svc svc.CustomerService, logger *zap.SugaredLogger) *CustomerHandler {
	return &CustomerHandler{svc: svc, logger: logger}
}

// List godoc
//
//	@Summary		List customers
//	@Description	Get all customers for the tenant
//	@Tags			customers
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		domain.Customer
//	@Failure		401	{object}	jsonutil.ErrorResponse
//	@Router			/customers [get]
func (h *CustomerHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)

	customers, err := h.svc.ListByTenant(ctx, claims.TenantID)
	if err != nil {
		handleError(w, err, h.logger)
		return
	}

	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(customers))
}

// GetByID godoc
//
//	@Summary		Get customer by ID
//	@Description	Get a specific customer by ID
//	@Tags			customers
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Customer ID"
//	@Success		200	{object}	domain.Customer
//	@Failure		401	{object}	jsonutil.ErrorResponse
//	@Failure		404	{object}	jsonutil.ErrorResponse
//	@Router			/customers/{id} [get]
func (h *CustomerHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)

	id := chi.URLParam(r, "id")
	if id == "" {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	customer, err := h.svc.GetByID(ctx, claims.TenantID, id)
	if err != nil {
		handleError(w, err, h.logger)
		return
	}

	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(customer))
}

// Create godoc
//
//	@Summary		Create customer
//	@Description	Create a new customer (owner/admin only)
//	@Tags			customers
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		domain.CustomerRequest	true	"Create request"
//	@Success		201	{object}	domain.Customer
//	@Failure		400	{object}	jsonutil.ErrorResponse
//	@Failure		401	{object}	jsonutil.ErrorResponse
//	@Failure		403	{object}	jsonutil.ErrorResponse
//	@Router			/customers [post]
func (h *CustomerHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)

	var req domain.CustomerRequest
	if err := jsonutil.Read(r, &req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	customer, err := h.svc.Create(ctx, claims.TenantID, req)
	if err != nil {
		handleError(w, err, h.logger)
		return
	}

	jsonutil.Write(w, http.StatusCreated, jsonutil.NewResponse(customer))
}

// Update godoc
//
//	@Summary		Update customer
//	@Description	Update a customer (owner/admin only)
//	@Tags			customers
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Customer ID"
//	@Param			request	body		domain.CustomerRequest	true	"Update request"
//	@Success		200	{object}	domain.Customer
//	@Failure		400	{object}	jsonutil.ErrorResponse
//	@Failure		401	{object}	jsonutil.ErrorResponse
//	@Failure		403	{object}	jsonutil.ErrorResponse
//	@Failure		404	{object}	jsonutil.ErrorResponse
//	@Router			/customers/{id} [put]
func (h *CustomerHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)
	id := chi.URLParam(r, "id")
	if id == "" {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req domain.CustomerRequest
	if err := jsonutil.Read(r, &req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	customer, err := h.svc.Update(ctx, claims.TenantID, id, req)
	if err != nil {
		handleError(w, err, h.logger)
		return
	}

	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(customer))
}

// Delete godoc
//
//	@Summary		Delete customer
//	@Description	Delete a customer (owner/admin only)
//	@Tags			customers
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Customer ID"
//	@Success		204
//	@Failure		401	{object}	jsonutil.ErrorResponse
//	@Failure		403	{object}	jsonutil.ErrorResponse
//	@Failure		404	{object}	jsonutil.ErrorResponse
//	@Router			/customers/{id} [delete]
func (h *CustomerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)

	id := chi.URLParam(r, "id")
	if id == "" {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	err := h.svc.Delete(ctx, claims.TenantID, id)
	if err != nil {
		handleError(w, err, h.logger)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
