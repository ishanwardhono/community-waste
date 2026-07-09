package household

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
	r.Route("/households", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Get("/{id}", h.Get)
		r.Delete("/{id}", h.Delete)
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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	page := httpres.ParsePage(r)
	items, total, err := h.svc.List(r.Context(), page)
	if err != nil {
		httpres.Error(w, err)
		return
	}
	httpres.List(w, items, page.Meta(total))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpres.Error(w, apperr.New(http.StatusBadRequest, "invalid household id"))
		return
	}
	item, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httpres.Error(w, err)
		return
	}
	httpres.OK(w, item)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpres.Error(w, apperr.New(http.StatusBadRequest, "invalid household id"))
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		httpres.Error(w, err)
		return
	}
	httpres.OK(w, nil)
}
