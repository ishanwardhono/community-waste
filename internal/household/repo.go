package household

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"

	"github.com/ishanwardhono/community-waste/pkg/apperr"
	"github.com/ishanwardhono/community-waste/pkg/db"
)

//go:generate go tool mockgen -source=repo.go -destination=../../test/mocks/household/repo.go -package=mockhousehold
type Repository interface {
	Insert(ctx context.Context, h Household) error
	List(ctx context.Context, limit, offset int) ([]Household, int64, error)
	Get(ctx context.Context, id uuid.UUID) (Household, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *db.Database
}

func NewRepository(database *db.Database) Repository {
	return &repository{db: database}
}

func (r *repository) Insert(ctx context.Context, h Household) error {
	_, err := sqlx.NamedExecContext(ctx, r.db.Ext(ctx), insertQuery, h)
	return err
}

func (r *repository) List(ctx context.Context, limit, offset int) ([]Household, int64, error) {
	var total int64
	if err := sqlx.GetContext(ctx, r.db.Ext(ctx), &total, countQuery); err != nil {
		return nil, 0, err
	}
	items := []Household{}
	if err := sqlx.SelectContext(ctx, r.db.Ext(ctx), &items, listQuery, limit, offset); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *repository) Get(ctx context.Context, id uuid.UUID) (Household, error) {
	var h Household
	err := sqlx.GetContext(ctx, r.db.Ext(ctx), &h, getQuery, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Household{}, apperr.New(http.StatusNotFound, "household not found")
	}
	return h, err
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.Ext(ctx).ExecContext(ctx, deleteQuery, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return apperr.New(http.StatusConflict, "household still has pickups or payments")
		}
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return apperr.New(http.StatusNotFound, "household not found")
	}
	return nil
}
