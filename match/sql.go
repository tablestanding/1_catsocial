package match

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
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

type createRepoArgs struct {
	IssuerUserID   string
	ReceiverUserID string
	IssuerCatID    string
	ReceiverCatID  string
	Msg            string
}

func (s SQL) Create(ctx context.Context, args createRepoArgs) error {
	_, err := s.pool.Exec(ctx, `
		insert into matches(issuer_user_id, receiver_user_id, issuer_cat_id, receiver_cat_id, msg)
		values ($1, $2, $3, $4, $5)
	`, args.IssuerUserID, args.ReceiverUserID, args.IssuerCatID, args.ReceiverCatID, args.Msg)
	if err != nil {
		return fmt.Errorf("sql create match: %w", err)
	}

	return err
}

type getRepoArgs struct {
	UserID string
}

func (s SQL) Get(ctx context.Context, args getRepoArgs) ([]Match, error) {
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

func (s SQL) IsExist(ctx context.Context, id int) (bool, error) {
	err := s.pool.QueryRow(ctx, "select 1 from matches where id = $1", id).Scan()
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("sql finding match by id: %w", err)
	}

	return true, nil
}

func (s SQL) GetHasBeenApprovedOrRejectedById(ctx context.Context, id int) (bool, error) {
	var b bool
	err := s.pool.QueryRow(ctx, "select has_been_approved_or_rejected from matches where id = $1", id).Scan(&b)
	if err != nil {
		e := err
		if err == pgx.ErrNoRows {
			e = ErrMatchNotFound
		}
		return b, fmt.Errorf("sql finding match by id: %w", e)
	}

	return b, nil
}

func (s SQL) GetByID(ctx context.Context, id int) (MatchRaw, error) {
	var m MatchRaw
	err := s.pool.QueryRow(ctx, `
		select 
			id, issuer_user_id, receiver_user_id, issuer_cat_id, receiver_cat_id,
			has_been_approved_or_rejected, created_at, msg
		from matches 
		where id = $1
	`, id).Scan(&m.ID, &m.IssuerUserID, &m.ReceiverUserID, &m.IssuerCatID, &m.ReceiverCatID,
		&m.HasBeenApprovedOrRejected, &m.CreatedAt, &m.Msg)
	if err != nil {
		e := err
		if err == pgx.ErrNoRows {
			e = ErrMatchNotFound
		}
		return MatchRaw{}, fmt.Errorf("sql finding match by id: %w", e)
	}

	return m, nil
}

func (s SQL) GetByCatID(ctx context.Context, catID int) (MatchRaw, error) {
	var m MatchRaw
	err := s.pool.QueryRow(ctx, `
		select 
			id, issuer_user_id, receiver_user_id, issuer_cat_id, receiver_cat_id,
			has_been_approved_or_rejected, created_at, msg
		from matches 
		where issuer_cat_id = $1
		or receiver_cat_id = $1
	`, catID).Scan(&m.ID, &m.IssuerUserID, &m.ReceiverUserID, &m.IssuerCatID, &m.ReceiverCatID,
		&m.HasBeenApprovedOrRejected, &m.CreatedAt, &m.Msg)
	if err != nil {
		e := err
		if err == pgx.ErrNoRows {
			e = ErrMatchNotFound
		}
		return MatchRaw{}, fmt.Errorf("sql finding match by id: %w", e)
	}

	return m, nil
}

type updateRepoArgs struct {
	ID                        int
	HasBeenApprovedOrRejected *bool
}

func (s SQL) Update(ctx context.Context, args updateRepoArgs) error {
	var (
		query   strings.Builder
		sqlArgs []any

		arg = 1
	)
	query.WriteString("update matches set ")

	if args.HasBeenApprovedOrRejected != nil {
		query.WriteString(fmt.Sprintf(`
			has_been_approved_or_rejected = $%d
		`, arg))
		sqlArgs = append(sqlArgs, *args.HasBeenApprovedOrRejected)
		arg += 1
	}

	query.WriteString(fmt.Sprintf(`
		where id = $%d
	`, arg))
	sqlArgs = append(sqlArgs, args.ID)

	_, err := s.pool.Exec(ctx, query.String(), sqlArgs...)
	if err != nil {
		return fmt.Errorf("sql update match by id: %w", err)
	}

	return nil
}

type deleteRepoArgs struct {
	CatIDs         []int
	ExcludeMatchID *int
	MatchID        *int
}

func (s SQL) Delete(ctx context.Context, args deleteRepoArgs) error {
	var (
		query        strings.Builder
		whereQueries []string
		sqlArgs      []any

		arg = 1
	)
	query.WriteString("delete from matches ")

	if args.CatIDs != nil && len(args.CatIDs) > 0 {
		whereQueries = append(whereQueries, fmt.Sprintf(`
			(issuer_cat_id = any($%d) or receiver_cat_id = any($%d))
		`, arg, arg))
		sqlArgs = append(sqlArgs, args.CatIDs)
		arg += 1
	}

	if args.ExcludeMatchID != nil {
		whereQueries = append(whereQueries, (fmt.Sprintf(`
			id != $%d
		`, arg)))
		sqlArgs = append(sqlArgs, *args.ExcludeMatchID)
		arg += 1
	}

	if args.MatchID != nil {
		whereQueries = append(whereQueries, (fmt.Sprintf(`
			id = $%d
		`, arg)))
		sqlArgs = append(sqlArgs, *args.MatchID)
		arg += 1
	}

	if len(whereQueries) > 0 {
		query.WriteString(fmt.Sprintf(`
			where %s
		`, strings.Join(whereQueries, " and ")))
	}

	_, err := s.pool.Exec(ctx, query.String(), sqlArgs...)
	if err != nil {
		return fmt.Errorf("sql delete matches: %w", err)
	}

	return nil
}
