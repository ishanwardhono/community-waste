package report

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/ishanwardhono/community-waste/pkg/apperr"
	"github.com/ishanwardhono/community-waste/pkg/db"
)

//go:generate go tool mockgen -source=repo.go -destination=../../test/mocks/report/repo.go -package=mockreport
type Repository interface {
	WasteSummary(ctx context.Context) ([]WasteSummaryRow, error)
	PaymentSummary(ctx context.Context) ([]PaymentStatusRow, error)
	HouseholdInfo(ctx context.Context, id uuid.UUID) (HouseholdInfo, error)
	PickupsByHousehold(ctx context.Context, id uuid.UUID) ([]PickupRow, error)
	PaymentsByHousehold(ctx context.Context, id uuid.UUID) ([]PaymentRow, error)
}

type repository struct {
	db *db.Database
}

func NewRepository(database *db.Database) Repository {
	return &repository{db: database}
}

func (r *repository) WasteSummary(ctx context.Context) ([]WasteSummaryRow, error) {
	rows := []WasteSummaryRow{}
	err := sqlx.SelectContext(ctx, r.db.Ext(ctx), &rows, wasteSummaryQuery)
	return rows, err
}

func (r *repository) PaymentSummary(ctx context.Context) ([]PaymentStatusRow, error) {
	rows := []PaymentStatusRow{}
	err := sqlx.SelectContext(ctx, r.db.Ext(ctx), &rows, paymentSummaryQuery)
	return rows, err
}

func (r *repository) HouseholdInfo(ctx context.Context, id uuid.UUID) (HouseholdInfo, error) {
	var info HouseholdInfo
	err := sqlx.GetContext(ctx, r.db.Ext(ctx), &info, householdInfoQuery, id)
	if errors.Is(err, sql.ErrNoRows) {
		return HouseholdInfo{}, apperr.New(http.StatusNotFound, "household not found")
	}
	return info, err
}

func (r *repository) PickupsByHousehold(ctx context.Context, id uuid.UUID) ([]PickupRow, error) {
	rows := []PickupRow{}
	err := sqlx.SelectContext(ctx, r.db.Ext(ctx), &rows, pickupsByHouseholdQuery, id)
	return rows, err
}

func (r *repository) PaymentsByHousehold(ctx context.Context, id uuid.UUID) ([]PaymentRow, error) {
	rows := []PaymentRow{}
	err := sqlx.SelectContext(ctx, r.db.Ext(ctx), &rows, paymentsByHouseholdQuery, id)
	return rows, err
}
