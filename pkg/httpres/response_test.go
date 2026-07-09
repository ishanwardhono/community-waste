package httpres

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ishanwardhono/community-waste/pkg/apperr"
)

func TestErrorWritesAppError(t *testing.T) {
	rec := httptest.NewRecorder()
	Error(rec, apperr.New(http.StatusNotFound, "household not found"))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	var body map[string]any
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["message"] != "household not found" {
		t.Fatalf("message = %v", body["message"])
	}
}

func TestErrorHidesUnknownErrors(t *testing.T) {
	rec := httptest.NewRecorder()
	Error(rec, sqlErrSentinel{})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	var body map[string]any
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["message"] != "internal server error" {
		t.Fatalf("unknown error leaked: %v", body["message"])
	}
}

type sqlErrSentinel struct{}

func (sqlErrSentinel) Error() string { return "pq: secret table detail" }
