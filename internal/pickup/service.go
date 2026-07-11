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
	Complete(ctx context.Context, id uuid.UUID) (Pickup, error)
}

//go:generate go tool mockgen -destination=../../test/mocks/pickup/deps.go -package=mockpickup github.com/ishanwardhono/community-waste/internal/pickup PaymentService,TxRunner

// PaymentService is what pickup needs from the payment module.
// Declared here so imports keep pointing down the module tree.
type PaymentService interface {
	HasPending(ctx context.Context, householdID uuid.UUID) (bool, error)
	CreateForPickup(ctx context.Context, householdID, pickupID uuid.UUID, wasteType WasteType) error
}

type TxRunner interface {
	WithTx(ctx context.Context, fn func(context.Context) error) error
}

type service struct {
	repo       Repository
	households household.Service
	payments   PaymentService
	tx         TxRunner
}

func NewService(repo Repository, households household.Service, payments PaymentService, tx TxRunner) Service {
	return &service{repo: repo, households: households, payments: payments, tx: tx}
}

func (s *service) Create(ctx context.Context, req CreateRequest) (Pickup, error) {
	if _, err := s.households.Get(ctx, req.HouseholdID); err != nil {
		return Pickup{}, err
	}
	pending, err := s.payments.HasPending(ctx, req.HouseholdID)
	if err != nil {
		return Pickup{}, err
	}
	if pending {
		return Pickup{}, apperr.New(http.StatusUnprocessableEntity, "household has a pending payment")
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

func (s *service) Complete(ctx context.Context, id uuid.UUID) (Pickup, error) {
	p, err := s.repo.Get(ctx, id)
	if err != nil {
		return Pickup{}, err
	}

	var completed Pickup
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		var err error
		completed, err = s.repo.Complete(ctx, id)
		if err != nil {
			return err
		}
		return s.payments.CreateForPickup(ctx, p.HouseholdID, p.ID, p.Type)
	})
	return completed, err
}
