package pickup

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/ishanwardhono/community-waste/internal/household"
	"github.com/ishanwardhono/community-waste/pkg/apperr"
	"github.com/ishanwardhono/community-waste/pkg/logger"
)

type Service interface {
	Create(ctx context.Context, req CreateRequest) (Pickup, error)
	List(ctx context.Context, f ListFilter) ([]Pickup, int64, error)
	Schedule(ctx context.Context, id uuid.UUID, req ScheduleRequest) (Pickup, error)
	Cancel(ctx context.Context, id uuid.UUID) (Pickup, error)
}

type service struct {
	repo       Repository
	households household.Service
}

func NewService(repo Repository, households household.Service) Service {
	return &service{repo: repo, households: households}
}

func (s *service) Create(ctx context.Context, req CreateRequest) (Pickup, error) {
	if _, err := s.households.Get(ctx, req.HouseholdID); err != nil {
		return Pickup{}, err
	}

	now := time.Now()
	p := Pickup{
		ID:          uuid.Must(uuid.NewV7()),
		HouseholdID: req.HouseholdID,
		Type:        req.Type,
		Status:      StatusPending,
		SafetyCheck: req.SafetyCheck,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Insert(ctx, p); err != nil {
		logger.Errorf(ctx, "insert pickup: %v", err)
		return Pickup{}, err
	}
	return p, nil
}

func (s *service) List(ctx context.Context, f ListFilter) ([]Pickup, int64, error) {
	return s.repo.List(ctx, f)
}

func (s *service) Schedule(ctx context.Context, id uuid.UUID, req ScheduleRequest) (Pickup, error) {
	p, err := s.repo.Get(ctx, id)
	if err != nil {
		return Pickup{}, err
	}
	if p.Type == TypeElectronic && (p.SafetyCheck == nil || !*p.SafetyCheck) {
		return Pickup{}, apperr.New(http.StatusUnprocessableEntity,
			"electronic pickup needs a passed safety check before scheduling")
	}
	return s.repo.Schedule(ctx, id, req.PickupDate)
}

func (s *service) Cancel(ctx context.Context, id uuid.UUID) (Pickup, error) {
	if _, err := s.repo.Get(ctx, id); err != nil {
		return Pickup{}, err
	}
	return s.repo.Cancel(ctx, id)
}
