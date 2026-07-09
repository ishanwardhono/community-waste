package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/ishanwardhono/community-waste/pkg/logger"
)

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		logger.Infof(r.Context(), "%s %s %d %s", r.Method, r.URL.Path, ww.Status(), time.Since(start))
	})
}
