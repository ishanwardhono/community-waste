package db

import (
	"context"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/ishanwardhono/community-waste/pkg/config"
	"github.com/ishanwardhono/community-waste/pkg/logger"
)

type Database struct {
	pool *sqlx.DB
}

func New(cfg config.DBConfig) (*Database, error) {
	pool, err := sqlx.Connect("pgx", cfg.DSN())
	if err != nil {
		return nil, err
	}
	return &Database{pool: pool}, nil
}

func (d *Database) Close() error { return d.pool.Close() }

type txKey struct{}

// WithTx runs fn inside a transaction carried on the context.
// Nested calls reuse the outer transaction.
func (d *Database) WithTx(ctx context.Context, fn func(context.Context) error) error {
	if txFrom(ctx) != nil {
		return fn(ctx)
	}
	tx, err := d.pool.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Errorf(ctx, "rollback failed: %v", rbErr)
		}
		return err
	}
	return tx.Commit()
}

// Ext returns the transaction from ctx when present, otherwise the pool.
func (d *Database) Ext(ctx context.Context) sqlx.ExtContext {
	if tx := txFrom(ctx); tx != nil {
		return tx
	}
	return d.pool
}

func txFrom(ctx context.Context) *sqlx.Tx {
	tx, _ := ctx.Value(txKey{}).(*sqlx.Tx)
	return tx
}
