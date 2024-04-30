package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrUniqueEmailViolation = errors.New("unique email constraint violation")
)

type (
	CreateUserRepoArgs struct {
		Email          string
		HashedPassword string
		Name           string
	}

	SQL struct {
		pool *pgx.ConnPool
	}
)

func NewSQL(pool *pgx.ConnPool) SQL {
	return SQL{pool}
}

func (s SQL) Create(ctx context.Context, args CreateUserRepoArgs) error {
	_, err := s.pool.ExecEx(ctx, `
		insert into users(name, hashed_pw, email)
		values ($1, $2, $3)
	`, nil, args.Name, args.HashedPassword, args.Email)

	var e *pgconn.PgError
	if errors.As(err, &e) && e.Code == "23505" { // unique constraint violation
		return fmt.Errorf("sql create user: %w", ErrUniqueEmailViolation)
	}

	return err
}
