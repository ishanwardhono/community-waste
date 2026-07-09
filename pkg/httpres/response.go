package httpres

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ishanwardhono/community-waste/pkg/apperr"
)

type response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

type Meta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

func OK(w http.ResponseWriter, data any) {
	write(w, http.StatusOK, response{Code: http.StatusOK, Message: "success", Data: data})
}

func Created(w http.ResponseWriter, data any) {
	write(w, http.StatusCreated, response{Code: http.StatusCreated, Message: "success", Data: data})
}

func List(w http.ResponseWriter, data any, m Meta) {
	write(w, http.StatusOK, response{Code: http.StatusOK, Message: "success", Data: data, Meta: &m})
}

func Error(w http.ResponseWriter, err error) {
	var app *apperr.AppError
	if !errors.As(err, &app) {
		app = &apperr.AppError{Code: http.StatusInternalServerError, Message: "internal server error"}
	}
	write(w, app.Code, response{Code: app.Code, Message: app.Message})
}

func write(w http.ResponseWriter, status int, body response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
