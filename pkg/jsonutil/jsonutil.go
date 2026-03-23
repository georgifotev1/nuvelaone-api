package jsonutil

import (
	"encoding/json"
	"net/http"
)

type Response[T any] struct {
	Data T `json:"data"`
}

type PaginatedResponse[T any] struct {
	Data T    `json:"data"`
	Meta Meta `json:"meta"`
}

type Meta struct {
	Total   int `json:"total"`
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewResponse[T any](data T) Response[T] {
	return Response[T]{Data: data}
}

func NewPaginatedResponse[T any](data T, total, page, perPage int) PaginatedResponse[T] {
	return PaginatedResponse[T]{
		Data: data,
		Meta: Meta{
			Total:   total,
			Page:    page,
			PerPage: perPage,
		},
	}
}

func Write(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, message string) {
	Write(w, status, ErrorResponse{Error: message})
}

func Read(r *http.Request, data any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(data)
}
