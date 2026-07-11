package pickup_test

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/ishanwardhono/community-waste/internal/pickup"
	mockpickup "github.com/ishanwardhono/community-waste/test/mocks/pickup"
)

func TestWorkerSweepsAndStops(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mockpickup.NewMockRepository(ctrl)

	maxAge := 72 * time.Hour
	var gotCutoff time.Time
	repo.EXPECT().CancelStaleOrganic(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, cutoff time.Time) (int64, error) {
			gotCutoff = cutoff
			return 2, nil
		}).MinTimes(1)

	w := pickup.NewWorker(repo, time.Millisecond, maxAge)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		w.Run(ctx)
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after cancel")
	}

	wantCutoff := time.Now().Add(-maxAge)
	if gotCutoff.Before(wantCutoff.Add(-time.Minute)) || gotCutoff.After(wantCutoff.Add(time.Minute)) {
		t.Fatalf("cutoff = %v, want about %v", gotCutoff, wantCutoff)
	}
}
