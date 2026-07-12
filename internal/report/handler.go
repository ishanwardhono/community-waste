package report

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ishanwardhono/community-waste/pkg/apperr"
	"github.com/ishanwardhono/community-waste/pkg/httpres"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r chi.Router) {
	r.Route("/reports", func(r chi.Router) {
		r.Get("/waste-summary", h.WasteSummary)
		r.Get("/payment-summary", h.PaymentSummary)
		r.Get("/households/{id}/history", h.HouseholdHistory)
	})
}

func (h *Handler) WasteSummary(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.WasteSummary(r.Context())
	if err != nil {
		httpres.Error(w, err)
		return
	}
	httpres.OK(w, rows)
}

func (h *Handler) PaymentSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.svc.PaymentSummary(r.Context())
	if err != nil {
		httpres.Error(w, err)
		return
	}
	httpres.OK(w, summary)
}

func (h *Handler) HouseholdHistory(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpres.Error(w, apperr.New(http.StatusBadRequest, "invalid household id"))
		return
	}
	history, err := h.svc.HouseholdHistory(r.Context(), id)
	if err != nil {
		httpres.Error(w, err)
		return
	}
	httpres.OK(w, history)
}
