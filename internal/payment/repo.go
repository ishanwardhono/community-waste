package payment

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/ishanwardhono/community-waste/pkg/db"
)

//go:generate go tool mockgen -source=repo.go -destination=../../test/mocks/payment/repo.go -package=mockpayment
type Repository interface {
	Insert(ctx context.Context, p Payment) error
	List(ctx context.Context, f ListFilter) ([]Payment, int64, error)
	HasPending(ctx context.Context, householdID uuid.UUID) (bool, error)
}

type repository struct {
	db *db.Database
}

func NewRepository(database *db.Database) Repository {
	return &repository{db: database}
}

func (r *repository) Insert(ctx context.Context, p Payment) error {
	_, err := sqlx.NamedExecContext(ctx, r.db.Ext(ctx), insertQuery, p)
	return err
}

func (r *repository) HasPending(ctx context.Context, householdID uuid.UUID) (bool, error) {
	var pending bool
	err := sqlx.GetContext(ctx, r.db.Ext(ctx), &pending, hasPendingQuery, householdID)
	return pending, err
}

func (r *repository) List(ctx context.Context, f ListFilter) ([]Payment, int64, error) {
	where := ""
	args := []any{}
	if f.Status != "" {
		args = append(args, f.Status)
		where += fmt.Sprintf(" AND status = $%d", len(args))
	}
	if f.HouseholdID != uuid.Nil {
		args = append(args, f.HouseholdID)
		where += fmt.Sprintf(" AND household_id = $%d", len(args))
	}
	if f.From != nil {
		args = append(args, *f.From)
		where += fmt.Sprintf(" AND created_at >= $%d", len(args))
	}
	if f.To != nil {
		args = append(args, *f.To)
		where += fmt.Sprintf(" AND created_at < $%d", len(args))
	}

	var total int64
	if err := sqlx.GetContext(ctx, r.db.Ext(ctx), &total, baseCount+where, args...); err != nil {
		return nil, 0, err
	}

	args = append(args, f.Page.Limit, f.Page.Offset())
	query := fmt.Sprintf("%s%s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		baseSelect, where, len(args)-1, len(args))
	items := []Payment{}
	if err := sqlx.SelectContext(ctx, r.db.Ext(ctx), &items, query, args...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
