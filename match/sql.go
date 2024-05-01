package match

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	CreateRepoArgs struct {
		IssuerUserID   string
		ReceiverUserID string
		IssuerCatID    string
		ReceiverCatID  string
		Msg            string
	}

	GetRepoArgs struct {
		UserID string
	}

	SQL struct {
		pool *pgxpool.Pool
	}
)

func NewSQL(pool *pgxpool.Pool) SQL {
	return SQL{pool}
}

func (s SQL) Create(ctx context.Context, args CreateRepoArgs) error {
	_, err := s.pool.Exec(ctx, `
		insert into matches(issuer_user_id, receiver_user_id, issuer_cat_id, receiver_cat_id, msg)
		values ($1, $2, $3, $4, $5)
	`, args.IssuerUserID, args.ReceiverUserID, args.IssuerCatID, args.ReceiverCatID, args.Msg)
	if err != nil {
		return fmt.Errorf("sql create match: %w", err)
	}

	return err
}

func (s SQL) Get(ctx context.Context, args GetRepoArgs) ([]Match, error) {
	var matches []Match
	rows, err := s.pool.Query(ctx, `
		select
			m.id,
			m.msg,
			m.created_at,

			issuer_user.name,
			issuer_user.email,
			issuer_user.created_at,

			issuer_cat.id,
			issuer_cat.name, 
			issuer_cat.race, 
			issuer_cat.sex, 
			issuer_cat.description,
			issuer_cat.age_in_month, 
			issuer_cat.image_urls,
			issuer_cat.has_matched,
			issuer_cat.created_at,

			receiver_cat.id,
			receiver_cat.name, 
			receiver_cat.race, 
			receiver_cat.sex, 
			receiver_cat.description,
			receiver_cat.age_in_month, 
			receiver_cat.image_urls,
			receiver_cat.has_matched,
			receiver_cat.created_at
		from matches m
			inner join users issuer_user
				on m.issuer_user_id = issuer_user.id
			inner join cats issuer_cat
				on m.issuer_cat_id = issuer_cat.id
			inner join cats receiver_cat
				on m.receiver_cat_id = receiver_cat.id
		where m.issuer_user_id = $1
		or m.receiver_user_id = $1
		order by m.id desc
	`, args.UserID)
	if err != nil {
		return nil, fmt.Errorf("sql get matches: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m Match
		err = rows.Scan(&m.ID, &m.Msg, &m.CreatedAt,
			// issuer user
			&m.IssuerUser.Name, &m.IssuerUser.Email, &m.IssuerUser.CreatedAt,
			// issuer cat
			&m.IssuerCat.ID, &m.IssuerCat.Name, &m.IssuerCat.Race, &m.IssuerCat.Sex,
			&m.IssuerCat.Description, &m.IssuerCat.AgeInMonth, &m.IssuerCat.ImageURLs,
			&m.IssuerCat.HasMatched, &m.IssuerCat.CreatedAt,
			//receiver cat
			&m.ReceiverCat.ID, &m.ReceiverCat.Name, &m.ReceiverCat.Race, &m.ReceiverCat.Sex,
			&m.ReceiverCat.Description, &m.ReceiverCat.AgeInMonth, &m.ReceiverCat.ImageURLs,
			&m.ReceiverCat.HasMatched, &m.ReceiverCat.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("sql get matches: %w", err)
		}

		matches = append(matches, m)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("sql get matches: %w", err)
	}

	return matches, nil
}
