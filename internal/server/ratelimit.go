package server

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/ishanwardhono/community-waste/pkg/apperr"
	"github.com/ishanwardhono/community-waste/pkg/httpres"
)

type IPLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rps      rate.Limit
	burst    int
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewIPLimiter(rps float64, burst int) *IPLimiter {
	return &IPLimiter{visitors: map[string]*visitor{}, rps: rate.Limit(rps), burst: burst}
}

func (l *IPLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.allow(clientIP(r)) {
			httpres.Error(w, apperr.New(http.StatusTooManyRequests, "too many pickup requests, try again shortly"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (l *IPLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	v, ok := l.visitors[ip]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(l.rps, l.burst)}
		l.visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter.Allow()
}

// CleanupLoop drops idle entries so the visitors map does not grow forever.
func (l *IPLimiter) CleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			l.mu.Lock()
			for ip, v := range l.visitors {
				if time.Since(v.lastSeen) > 3*time.Minute {
					delete(l.visitors, ip)
				}
			}
			l.mu.Unlock()
		}
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
