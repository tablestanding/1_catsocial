package pgxtrx

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	DB interface {
		Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
		Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
		QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	}

	contextKey string

	PgxTrx struct {
		pool *pgxpool.Pool
	}
)

const (
	trxContextKey contextKey = "//trx"
)

func New(pool *pgxpool.Pool) PgxTrx {
	return PgxTrx{pool}
}

func (p PgxTrx) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("pgxtrx with transaction: begin transaction: %w", err)
	}

	ctx = context.WithValue(ctx, trxContextKey, tx)

	err = fn(ctx)
	if err != nil {
		tx.Rollback(ctx)

		return fmt.Errorf("pgxtrx with transaction: execute statements: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		tx.Rollback(ctx)

		return fmt.Errorf("pgxtrx with transaction: commit transaction: %w", err)
	}

	return nil
}

func (p PgxTrx) FromContext(ctx context.Context) DB {
	tx, ok := ctx.Value(trxContextKey).(pgx.Tx)
	if !ok {
		return p.pool
	}
	return tx
}
