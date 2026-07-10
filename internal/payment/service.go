package payment

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/ishanwardhono/community-waste/internal/household"
	"github.com/ishanwardhono/community-waste/internal/pickup"
	"github.com/ishanwardhono/community-waste/pkg/apperr"
	"github.com/ishanwardhono/community-waste/pkg/logger"
)

type Service interface {
	Create(ctx context.Context, req CreateRequest) (Payment, error)
	List(ctx context.Context, f ListFilter) ([]Payment, int64, error)
}

type service struct {
	repo       Repository
	households household.Service
	pickups    pickup.Repository
}

func NewService(repo Repository, households household.Service, pickups pickup.Repository) Service {
	return &service{repo: repo, households: households, pickups: pickups}
}

func (s *service) Create(ctx context.Context, req CreateRequest) (Payment, error) {
	if _, err := s.households.Get(ctx, req.HouseholdID); err != nil {
		return Payment{}, err
	}
	p, err := s.pickups.Get(ctx, req.WasteID)
	if err != nil {
		return Payment{}, err
	}
	if p.HouseholdID != req.HouseholdID {
		return Payment{}, apperr.New(http.StatusUnprocessableEntity, "pickup does not belong to this household")
	}

	now := time.Now()
	pay := Payment{
		ID:          uuid.Must(uuid.NewV7()),
		HouseholdID: req.HouseholdID,
		WasteID:     req.WasteID,
		Amount:      req.Amount,
		Status:      StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Insert(ctx, pay); err != nil {
		logger.Errorf(ctx, "insert payment: %v", err)
		return Payment{}, err
	}
	return pay, nil
}

func (s *service) List(ctx context.Context, f ListFilter) ([]Payment, int64, error) {
	return s.repo.List(ctx, f)
}
