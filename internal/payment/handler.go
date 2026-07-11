package payment

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
	r.Route("/payments", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Put("/{id}/confirm", h.Confirm)
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

func (h *Handler) Confirm(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpres.Error(w, apperr.New(http.StatusBadRequest, "invalid payment id"))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxProofSize+4096)
	file, header, err := r.FormFile("proof")
	if err != nil {
		httpres.Error(w, apperr.New(http.StatusBadRequest, "proof file is required"))
		return
	}
	defer file.Close()

	proof := ProofFile{
		Name:        header.Filename,
		Size:        header.Size,
		ContentType: header.Header.Get("Content-Type"),
		Reader:      file,
	}
	if err := proof.Validate(); err != nil {
		httpres.Error(w, err)
		return
	}
	confirmed, err := h.svc.Confirm(r.Context(), id, proof)
	if err != nil {
		httpres.Error(w, err)
		return
	}
	httpres.OK(w, confirmed)
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
