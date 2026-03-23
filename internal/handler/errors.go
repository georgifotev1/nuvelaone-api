package handler

import (
	"errors"
	"net/http"

	"github.com/georgifotev1/nuvelaone-api/internal/service"
	"github.com/georgifotev1/nuvelaone-api/pkg/jsonutil"
)

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		jsonutil.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrConflict):
		jsonutil.WriteError(w, http.StatusConflict, err.Error())
	default:
		jsonutil.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
