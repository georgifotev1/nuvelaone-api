package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/georgifotev1/nuvelaone-api/internal/service"
	"github.com/georgifotev1/nuvelaone-api/pkg/jsonutil"
)

func handleError(w http.ResponseWriter, err error) {
	fmt.Println("error: ", err)
	switch {
	case errors.Is(err, service.ErrNotFound):
		jsonutil.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrConflict):
		jsonutil.WriteError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrInvalidCredentials):
		jsonutil.WriteError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, service.ErrInvalidToken):
		jsonutil.WriteError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, service.ErrForbidden):
		jsonutil.WriteError(w, http.StatusForbidden, err.Error())
	default:
		jsonutil.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
