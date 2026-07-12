package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/ishanwardhono/community-waste/internal/household"
	"github.com/ishanwardhono/community-waste/internal/payment"
	"github.com/ishanwardhono/community-waste/internal/pickup"
	"github.com/ishanwardhono/community-waste/internal/report"
	"github.com/ishanwardhono/community-waste/pkg/httpres"
)

func NewRouter(households *household.Handler, pickups *pickup.Handler, payments *payment.Handler, reports *report.Handler, pickupLimiter *IPLimiter) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer, RequestLogger)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		httpres.OK(w, map[string]string{"status": "up"})
	})

	r.Route("/api", func(api chi.Router) {
		households.Register(api)
		pickups.Register(api, pickupLimiter.Middleware)
		payments.Register(api)
		reports.Register(api)
	})

	return r
}
