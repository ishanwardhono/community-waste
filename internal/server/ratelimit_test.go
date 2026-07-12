package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIPLimiterBlocksAfterBurst(t *testing.T) {
	l := NewIPLimiter(1, 2)
	h := l.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	statusFor := func(ip string) int {
		req := httptest.NewRequest(http.MethodPost, "/api/pickups", nil)
		req.RemoteAddr = ip + ":51234"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	if got := statusFor("10.0.0.1"); got != http.StatusCreated {
		t.Fatalf("first request = %d", got)
	}
	if got := statusFor("10.0.0.1"); got != http.StatusCreated {
		t.Fatalf("second request = %d", got)
	}
	if got := statusFor("10.0.0.1"); got != http.StatusTooManyRequests {
		t.Fatalf("third request = %d, want 429", got)
	}
	if got := statusFor("10.0.0.2"); got != http.StatusCreated {
		t.Fatalf("other ip = %d, should not be limited", got)
	}
}
