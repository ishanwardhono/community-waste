package household_test

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ishanwardhono/community-waste/internal/household"
	mockhousehold "github.com/ishanwardhono/community-waste/test/mocks/household"
)

func TestCreateFillsIDAndTimestamps(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mockhousehold.NewMockRepository(ctrl)
	svc := household.NewService(repo)

	var saved household.Household
	repo.EXPECT().Insert(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, h household.Household) error {
			saved = h
			return nil
		})

	got, err := svc.Create(context.Background(), household.CreateRequest{
		OwnerName: "  Budi ", Address: "Jl. Melati 1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != saved.ID || got.ID.String() == "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("id not generated: %v", got.ID)
	}
	if got.ID.Version() != 7 {
		t.Fatalf("id version = %d, want 7", got.ID.Version())
	}
	if got.OwnerName != "Budi" {
		t.Fatalf("owner not trimmed: %q", got.OwnerName)
	}
	if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
		t.Fatal("timestamps not set")
	}
}

func TestCreateReturnsRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mockhousehold.NewMockRepository(ctrl)
	svc := household.NewService(repo)

	repo.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(errors.New("db down"))

	_, err := svc.Create(context.Background(), household.CreateRequest{OwnerName: "a", Address: "b"})
	if err == nil {
		t.Fatal("expected error")
	}
}
