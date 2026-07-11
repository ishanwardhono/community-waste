package pickup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/ishanwardhono/community-waste/pkg/apperr"
	"github.com/ishanwardhono/community-waste/pkg/db"
)

//go:generate go tool mockgen -source=repo.go -destination=../../test/mocks/pickup/repo.go -package=mockpickup
type Repository interface {
	Insert(ctx context.Context, p Pickup) error
	List(ctx context.Context, f ListFilter) ([]Pickup, int64, error)
	Get(ctx context.Context, id uuid.UUID) (Pickup, error)
	Schedule(ctx context.Context, id uuid.UUID, date time.Time) (Pickup, error)
	Cancel(ctx context.Context, id uuid.UUID) (Pickup, error)
	Complete(ctx context.Context, id uuid.UUID) (Pickup, error)
	CancelStaleOrganic(ctx context.Context, cutoff time.Time) (int64, error)
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

func (r *repository) Get(ctx context.Context, id uuid.UUID) (Pickup, error) {
	var p Pickup
	err := sqlx.GetContext(ctx, r.db.Ext(ctx), &p, getQuery, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Pickup{}, apperr.New(http.StatusNotFound, "pickup not found")
	}
	return p, err
}

func (r *repository) Schedule(ctx context.Context, id uuid.UUID, date time.Time) (Pickup, error) {
	return r.guardedUpdate(ctx, scheduleQuery, "pickup can only be scheduled from pending status", id, date)
}

func (r *repository) Complete(ctx context.Context, id uuid.UUID) (Pickup, error) {
	return r.guardedUpdate(ctx, completeQuery, "pickup can only be completed from scheduled status", id)
}

func (r *repository) Cancel(ctx context.Context, id uuid.UUID) (Pickup, error) {
	return r.guardedUpdate(ctx, cancelQuery, "pickup can not be canceled from its current status", id)
}

func (r *repository) CancelStaleOrganic(ctx context.Context, cutoff time.Time) (int64, error) {
	res, err := r.db.Ext(ctx).ExecContext(ctx, cancelStaleOrganicQuery, cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// guardedUpdate runs an update that carries its status rule in the where clause,
// so a no row result means the pickup is in the wrong state.
func (r *repository) guardedUpdate(ctx context.Context, query, conflictMsg string, args ...any) (Pickup, error) {
	var p Pickup
	err := sqlx.GetContext(ctx, r.db.Ext(ctx), &p, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return Pickup{}, apperr.New(http.StatusConflict, conflictMsg)
	}
	return p, err
}
