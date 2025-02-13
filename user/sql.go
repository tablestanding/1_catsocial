package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	SQL struct {
		pool *pgxpool.Pool
	}
)

func NewSQL(pool *pgxpool.Pool) SQL {
	return SQL{pool}
}

type CreateUserRepoArgs struct {
	Email          string
	HashedPassword string
	Name           string
}

func (s SQL) Create(ctx context.Context, args CreateUserRepoArgs) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, `
		insert into users(name, hashed_pw, email)
		values ($1, $2, $3)
		returning id
	`, args.Name, args.HashedPassword, args.Email).Scan(&id)

	var e *pgconn.PgError
	if errors.As(err, &e) && e.Code == "23505" { // unique constraint violation
		return "", fmt.Errorf("sql create user: %w", ErrUniqueEmailViolation)
	}
	if err != nil {
		return "", fmt.Errorf("sql create user: %w", err)
	}

	return id, err
}

func (s SQL) GetOneByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `
		select id, email, hashed_pw, name, created_at
		from users
		where email = $1
	`, email).Scan(&u.ID, &u.Email, &u.HashedPassword, &u.Name, &u.CreatedAt)
	if err != nil {
		e := err
		if err == pgx.ErrNoRows {
			e = ErrUserNotFound
		}
		return u, fmt.Errorf("sql finding user by email: %w", e)
	}

	return u, nil
}
