package payment

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ishanwardhono/community-waste/external/storage"
	"github.com/ishanwardhono/community-waste/internal/household"
	"github.com/ishanwardhono/community-waste/internal/pickup"
	"github.com/ishanwardhono/community-waste/pkg/apperr"
	"github.com/ishanwardhono/community-waste/pkg/logger"
)

type Service interface {
	Create(ctx context.Context, req CreateRequest) (Payment, error)
	List(ctx context.Context, f ListFilter) ([]Payment, int64, error)
	Confirm(ctx context.Context, id uuid.UUID, proof ProofFile) (Payment, error)
	HasPending(ctx context.Context, householdID uuid.UUID) (bool, error)
	CreateForPickup(ctx context.Context, householdID, pickupID uuid.UUID, wasteType pickup.WasteType) error
}

type service struct {
	repo       Repository
	households household.Service
	pickups    pickup.Repository
	store      storage.FileStorage
}

func NewService(repo Repository, households household.Service, pickups pickup.Repository, store storage.FileStorage) Service {
	return &service{repo: repo, households: households, pickups: pickups, store: store}
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

func (s *service) Confirm(ctx context.Context, id uuid.UUID, proof ProofFile) (Payment, error) {
	pay, err := s.repo.Get(ctx, id)
	if err != nil {
		return Payment{}, err
	}
	if pay.Status != StatusPending {
		return Payment{}, apperr.New(http.StatusConflict, "payment can only be confirmed from pending status")
	}

	key := fmt.Sprintf("proofs/%s/%d%s", id, time.Now().UnixNano(), strings.ToLower(filepath.Ext(proof.Name)))
	url, err := s.store.Upload(ctx, key, proof.Reader, proof.Size, proof.ContentType)
	if err != nil {
		logger.Errorf(ctx, "upload proof for payment %s: %v", id, err)
		return Payment{}, err
	}
	return s.repo.Confirm(ctx, id, url)
}

func (s *service) HasPending(ctx context.Context, householdID uuid.UUID) (bool, error) {
	return s.repo.HasPending(ctx, householdID)
}

func (s *service) CreateForPickup(ctx context.Context, householdID, pickupID uuid.UUID, wasteType pickup.WasteType) error {
	now := time.Now()
	pay := Payment{
		ID:          uuid.Must(uuid.NewV7()),
		HouseholdID: householdID,
		WasteID:     pickupID,
		Amount:      amountFor(wasteType),
		Status:      StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Insert(ctx, pay); err != nil {
		logger.Errorf(ctx, "insert payment for pickup %s: %v", pickupID, err)
		return err
	}
	return nil
}

func amountFor(t pickup.WasteType) decimal.Decimal {
	if t == pickup.TypeElectronic {
		return decimal.NewFromInt(100000)
	}
	return decimal.NewFromInt(50000)
}
