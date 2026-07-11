package pickup

import (
	"context"
	"time"

	"github.com/ishanwardhono/community-waste/pkg/logger"
)

// Worker cancels organic pickups that were not picked up in time.
type Worker struct {
	repo     Repository
	interval time.Duration
	maxAge   time.Duration
}

func NewWorker(repo Repository, interval, maxAge time.Duration) *Worker {
	return &Worker{repo: repo, interval: interval, maxAge: maxAge}
}

func (w *Worker) Run(ctx context.Context) {
	logger.Infof(ctx, "auto cancel worker started, interval %s, max age %s", w.interval, w.maxAge)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		w.sweep(ctx)
		select {
		case <-ctx.Done():
			logger.Infof(ctx, "auto cancel worker stopped")
			return
		case <-ticker.C:
		}
	}
}

func (w *Worker) sweep(ctx context.Context) {
	cutoff := time.Now().Add(-w.maxAge)
	n, err := w.repo.CancelStaleOrganic(ctx, cutoff)
	if err != nil {
		logger.Errorf(ctx, "auto cancel sweep: %v", err)
		return
	}
	if n > 0 {
		logger.Infof(ctx, "auto canceled %d stale organic pickups", n)
	}
}
