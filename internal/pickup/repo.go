package pickup

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/ishanwardhono/community-waste/pkg/db"
)

//go:generate go tool mockgen -source=repo.go -destination=../../test/mocks/pickup/repo.go -package=mockpickup
type Repository interface {
	Insert(ctx context.Context, p Pickup) error
	List(ctx context.Context, f ListFilter) ([]Pickup, int64, error)
}

type repository struct {
	db *db.Database
}

func NewRepository(database *db.Database) Repository {
	return &repository{db: database}
}

func (r *repository) Insert(ctx context.Context, p Pickup) error {
	_, err := sqlx.NamedExecContext(ctx, r.db.Ext(ctx), insertQuery, p)
	return err
}

func (r *repository) List(ctx context.Context, f ListFilter) ([]Pickup, int64, error) {
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

	var total int64
	if err := sqlx.GetContext(ctx, r.db.Ext(ctx), &total, baseCount+where, args...); err != nil {
		return nil, 0, err
	}

	args = append(args, f.Page.Limit, f.Page.Offset())
	query := fmt.Sprintf("%s%s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		baseSelect, where, len(args)-1, len(args))
	items := []Pickup{}
	if err := sqlx.SelectContext(ctx, r.db.Ext(ctx), &items, query, args...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
