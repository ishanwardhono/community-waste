package report

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Service interface {
	WasteSummary(ctx context.Context) ([]WasteSummaryRow, error)
	PaymentSummary(ctx context.Context) (PaymentSummary, error)
	HouseholdHistory(ctx context.Context, id uuid.UUID) (HouseholdHistory, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) WasteSummary(ctx context.Context) ([]WasteSummaryRow, error) {
	return s.repo.WasteSummary(ctx)
}

func (s *service) PaymentSummary(ctx context.Context) (PaymentSummary, error) {
	rows, err := s.repo.PaymentSummary(ctx)
	if err != nil {
		return PaymentSummary{}, err
	}
	revenue := decimal.Zero
	for _, row := range rows {
		if row.Status == "paid" {
			revenue = revenue.Add(row.TotalAmount)
		}
	}
	return PaymentSummary{ByStatus: rows, TotalRevenue: revenue}, nil
}

func (s *service) HouseholdHistory(ctx context.Context, id uuid.UUID) (HouseholdHistory, error) {
	info, err := s.repo.HouseholdInfo(ctx, id)
	if err != nil {
		return HouseholdHistory{}, err
	}
	pickups, err := s.repo.PickupsByHousehold(ctx, id)
	if err != nil {
		return HouseholdHistory{}, err
	}
	payments, err := s.repo.PaymentsByHousehold(ctx, id)
	if err != nil {
		return HouseholdHistory{}, err
	}
	return HouseholdHistory{Household: info, Pickups: pickups, Payments: payments}, nil
}
