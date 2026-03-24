package handler

import (
	"net/http"
	"time"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/georgifotev1/nuvelaone-api/internal/service"
	"github.com/georgifotev1/nuvelaone-api/pkg/jsonutil"
	"github.com/georgifotev1/nuvelaone-api/pkg/validator"
	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	svc             service.AuthService
	refreshTokenTTL time.Duration
}

func NewAuthHandler(svc service.AuthService, refreshTokenTTL time.Duration) *AuthHandler {
	return &AuthHandler{svc: svc, refreshTokenTTL: refreshTokenTTL}
}

func (h *AuthHandler) Routes(r chi.Router) {
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)
	r.Post("/logout", h.Logout)
}

// Register godoc
//
//	@Summary		Register a new user
//	@Description	Register a new user with email, password and name
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		domain.RegisterRequest	true	"Register request"
//	@Success		201		{object}	domain.User
//	@Failure		400		{object}	jsonutil.ErrorResponse
//	@Failure		409		{object}	jsonutil.ErrorResponse
//	@Router			/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	if err := jsonutil.Read(r, &req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	user, err := h.svc.Register(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}
	jsonutil.Write(w, http.StatusCreated, jsonutil.NewResponse(user))
}

// Login godoc
//
//	@Summary		Login user
//	@Description	Authenticate user with email and password, returns access token and sets refresh token cookie
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		domain.LoginRequest	true	"Login request"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	jsonutil.ErrorResponse
//	@Failure		401		{object}	jsonutil.ErrorResponse
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := jsonutil.Read(r, &req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	pair, err := h.svc.Login(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}

	h.setRefreshCookie(w, pair.RefreshToken, int(h.refreshTokenTTL.Seconds()))
	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(map[string]string{
		"access_token": pair.AccessToken,
	}))
}

// Refresh godoc
//
//	@Summary		Refresh access token
//	@Description	Use refresh token cookie to get a new access token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		RefreshToken
//	@Success		200		{object}	map[string]string
//	@Failure		401		{object}	jsonutil.ErrorResponse
//	@Router			/auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		jsonutil.WriteError(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	pair, err := h.svc.Refresh(r.Context(), cookie.Value)
	if err != nil {
		handleError(w, err)
		return
	}

	h.setRefreshCookie(w, pair.RefreshToken, int(h.refreshTokenTTL.Seconds()))

	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(map[string]string{
		"access_token": pair.AccessToken,
	}))
}

// Logout godoc
//
//	@Summary		Logout user
//	@Description	Invalidate refresh token and clear cookie
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		RefreshToken
//	@Success		204		"No Content"
//	@Failure		400		{object}	jsonutil.ErrorResponse
//	@Failure		401		{object}	jsonutil.ErrorResponse
//	@Router			/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "missing refresh token")
		return
	}

	if err := h.svc.Logout(r.Context(), cookie.Value); err != nil {
		handleError(w, err)
		return
	}

	h.setRefreshCookie(w, "", -1)

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) setRefreshCookie(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    value,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/v1/auth",
		MaxAge:   maxAge,
	})
}
