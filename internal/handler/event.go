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

type EventHandler struct {
	svc svc.EventService
}

func NewEventHandler(svc svc.EventService) *EventHandler {
	return &EventHandler{svc: svc}
}

// Create godoc
//
//	@Summary		Create event
//	@Description	Create a new event
//	@Tags			events
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		domain.EventRequest	true	"Create request"
//	@Success		201	{object}	domain.Event
//	@Failure		400	{object}	jsonutil.ErrorResponse
//	@Failure		401	{object}	jsonutil.ErrorResponse
//	@Failure		409	{object}	jsonutil.ErrorResponse
//	@Router			/events [post]
func (h *EventHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)

	var req domain.EventRequest
	if err := jsonutil.Read(r, &req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	event, err := h.svc.Create(ctx, claims.TenantID, req)
	if err != nil {
		writeError(w, err)
		return
	}

	jsonutil.Write(w, http.StatusCreated, jsonutil.NewResponse(event))
}

// Update godoc
//
//	@Summary		Update event
//	@Description	Update an existing event
//	@Tags			events
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Event ID"
//	@Param			request	body		domain.EventUpdateRequest	true	"Update request"
//	@Success		200	{object}	domain.Event
//	@Failure		400	{object}	jsonutil.ErrorResponse
//	@Failure		401	{object}	jsonutil.ErrorResponse
//	@Failure		409	{object}	jsonutil.ErrorResponse
//	@Router			/events/{id} [put]
func (h *EventHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)

	id := chi.URLParam(r, "id")
	if id == "" {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req domain.EventUpdateRequest
	if err := jsonutil.Read(r, &req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	event, err := h.svc.Update(ctx, claims.TenantID, id, req)
	if err != nil {
		writeError(w, err)
		return
	}

	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(event))
}

// List godoc
//
//	@Summary		List events
//	@Description	List events with date filtering
//	@Tags			events
//	@Produce		json
//	@Security		BearerAuth
//	@Param			startDate	query		string	true	"Start date (YYYY-MM-DD)"
//	@Param			endDate		query		string	true	"End date (YYYY-MM-DD)"
//	@Success		200		{object}	jsonutil.Response[[]domain.Event]
//	@Failure		400		{object}	jsonutil.ErrorResponse
//	@Failure		401		{object}	jsonutil.ErrorResponse
//	@Router			/events [get]
func (h *EventHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if startDate == "" || endDate == "" {
		jsonutil.WriteError(w, http.StatusBadRequest, "start_date and end_date are required")
		return
	}

	filter := domain.EventListFilter{
		StartDate: startDate,
		EndDate:   endDate,
	}

	events, err := h.svc.List(ctx, claims.TenantID, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(events))
}

// ListMyBookings godoc
//
//	@Summary		List my bookings
//	@Description	List bookings for the authenticated customer
//	@Tags			public
//	@Produce		json
//	@Security		CustomerBearerAuth
//	@Param			slug	path		string	true	"Tenant slug"
//	@Success		200	{array}	domain.Event
//	@Router			/p/{slug}/bookings [get]
func (h *EventHandler) ListMyBookings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.CustomerClaimsFromContext(ctx)
	if claims == nil {
		jsonutil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	events, err := h.svc.ListByCustomer(ctx, claims.TenantID, claims.CustomerID)
	if err != nil {
		writeError(w, err)
		return
	}

	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(events))
}

// CreateCustomerBooking godoc
//
//	@Summary		Create booking
//	@Description	Create a new booking for the authenticated customer
//	@Tags			public
//	@Accept			json
//	@Produce		json
//	@Security		CustomerBearerAuth
//	@Param			slug		path		string				true	"Tenant slug"
//	@Param			request	body		domain.EventRequest	true	"Create request"
//	@Success		201	{object}	domain.Event
//	@Router			/p/{slug}/bookings [post]
func (h *EventHandler) CreateCustomerBooking(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.CustomerClaimsFromContext(ctx)
	if claims == nil {
		jsonutil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req domain.EventRequest
	if err := jsonutil.Read(r, &req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	event, err := h.svc.CreateForCustomer(ctx, claims.TenantID, claims.CustomerID, req)
	if err != nil {
		writeError(w, err)
		return
	}

	jsonutil.Write(w, http.StatusCreated, jsonutil.NewResponse(event))
}

// GetTimeslots godoc
//
//	@Summary		Get available timeslots
//	@Description	Get available timeslots for a service, user, and date
//	@Tags			public
//	@Produce		json
//	@Param			slug		path		string	true	"Tenant slug"
//	@Param			service_id	query		string	true	"Service ID"
//	@Param			user_id		query		string	true	"User ID"
//	@Param			date		query		string	true	"Date (YYYY-MM-DD)"
//	@Success		200	{object}	domain.TimeslotResponse
//	@Router			/p/{slug}/timeslots [get]
func (h *EventHandler) GetTimeslots(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := domain.TenantIDFromContext(ctx)
	if tenantID == "" {
		jsonutil.WriteError(w, http.StatusBadRequest, "tenant not found")
		return
	}

	serviceID := r.URL.Query().Get("service_id")
	userID := r.URL.Query().Get("user_id")
	date := r.URL.Query().Get("date")

	if serviceID == "" || userID == "" || date == "" {
		jsonutil.WriteError(w, http.StatusBadRequest, "service_id, user_id, and date are required")
		return
	}

	req := domain.TimeslotRequest{
		ServiceID: serviceID,
		UserID:    userID,
		Date:      date,
	}

	timeslots, err := h.svc.GetTimeslots(ctx, tenantID, req)
	if err != nil {
		writeError(w, err)
		return
	}

	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(domain.TimeslotResponse{Timeslots: timeslots}))
}
