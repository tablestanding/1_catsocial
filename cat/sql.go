package cat

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	CreateCatRepoArgs struct {
		Race        string
		Sex         string
		Name        string
		AgeInMonth  int
		Description string
		ImageURLs   []string
		UserID      string
	}

	SQL struct {
		pool *pgxpool.Pool
	}
)

func NewSQL(pool *pgxpool.Pool) SQL {
	return SQL{pool}
}

func (s SQL) Create(ctx context.Context, args CreateCatRepoArgs) (Cat, error) {
	c := Cat{
		UserID:      args.UserID,
		Race:        args.Race,
		Sex:         args.Sex,
		AgeInMonth:  args.AgeInMonth,
		Description: args.Description,
		ImageURLs:   args.ImageURLs,
	}
	err := s.pool.QueryRow(ctx, `
		insert into cats(user_id, race, sex, age_in_month, description, image_urls)
		values ($1, $2, $3, $4, $5, $6)
		returning id, created_at, has_matched
	`, args.UserID, args.Race, args.Sex, args.AgeInMonth, args.Description, args.ImageURLs).Scan(&c.ID, &c.CreatedAt, &c.HasMatched)

	return c, err
}

// func (s SQL) GetOneByEmail(ctx context.Context, email string) (User, error) {
// 	var u User
// 	err := s.pool.QueryRow(ctx, `
// 		select id, email, hashed_pw, name
// 		from users
// 		where email = $1
// 	`, email).Scan(&u.ID, &u.Email, &u.HashedPassword, &u.Name)
// 	if err != nil && err == pgx.ErrNoRows {
// 		e := err
// 		if err == pgx.ErrNoRows {
// 			e = ErrUserNotFound
// 		}
// 		return u, fmt.Errorf("finding user by email: %w", e)
// 	}

// 	return u, nil
// }
