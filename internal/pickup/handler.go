package pickup

import (
	"encoding/json"
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
	r.Route("/pickups", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Put("/{id}/schedule", h.Schedule)
		r.Put("/{id}/cancel", h.Cancel)
		r.Put("/{id}/complete", h.Complete)
	})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpres.Error(w, apperr.New(http.StatusBadRequest, "invalid request body"))
		return
	}
	defer r.Body.Close()

	if err := req.Validate(); err != nil {
		httpres.Error(w, err)
		return
	}
	created, err := h.svc.Create(r.Context(), req)
	if err != nil {
		httpres.Error(w, err)
		return
	}
	httpres.Created(w, created)
}

func (h *Handler) Schedule(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpres.Error(w, apperr.New(http.StatusBadRequest, "invalid pickup id"))
		return
	}
	var req ScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpres.Error(w, apperr.New(http.StatusBadRequest, "invalid request body"))
		return
	}
	defer r.Body.Close()

	if err := req.Validate(); err != nil {
		httpres.Error(w, err)
		return
	}
	updated, err := h.svc.Schedule(r.Context(), id, req)
	if err != nil {
		httpres.Error(w, err)
		return
	}
	httpres.OK(w, updated)
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpres.Error(w, apperr.New(http.StatusBadRequest, "invalid pickup id"))
		return
	}
	updated, err := h.svc.Cancel(r.Context(), id)
	if err != nil {
		httpres.Error(w, err)
		return
	}
	httpres.OK(w, updated)
}

func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpres.Error(w, apperr.New(http.StatusBadRequest, "invalid pickup id"))
		return
	}
	updated, err := h.svc.Complete(r.Context(), id)
	if err != nil {
		httpres.Error(w, err)
		return
	}
	httpres.OK(w, updated)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	f, err := ParseListFilter(r)
	if err != nil {
		httpres.Error(w, err)
		return
	}
	items, total, err := h.svc.List(r.Context(), f)
	if err != nil {
		httpres.Error(w, err)
		return
	}
	httpres.List(w, items, f.Page.Meta(total))
}
