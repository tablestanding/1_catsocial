package cat

import (
	"catsocial/pkg/pgxtrx"
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type (
	SQL struct {
		pgxTrx pgxtrx.PgxTrx
	}
)

func NewSQL(pgxTrx pgxtrx.PgxTrx) SQL {
	return SQL{pgxTrx}
}

type createRepoArgs struct {
	Race        string
	Sex         string
	Name        string
	AgeInMonth  int
	Description string
	ImageURLs   []string
	UserID      string
}

func (s SQL) Create(ctx context.Context, args createRepoArgs) (Cat, error) {
	db := s.pgxTrx.FromContext(ctx)

	c := Cat{
		UserID:      args.UserID,
		Race:        args.Race,
		Sex:         args.Sex,
		AgeInMonth:  args.AgeInMonth,
		Description: args.Description,
		ImageURLs:   args.ImageURLs,
		Name:        args.Name,
	}
	err := db.QueryRow(ctx, `
		insert into cats(user_id, race, sex, age_in_month, description, image_urls, name, name_normalized)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
		returning id, created_at, has_matched
	`, args.UserID, args.Race, args.Sex, args.AgeInMonth, args.Description, args.ImageURLs, args.Name, strings.ToLower(args.Name)).
		Scan(&c.ID, &c.CreatedAt, &c.HasMatched)
	if err != nil {
		return c, fmt.Errorf("sql create cat: %w", err)
	}

	return c, nil
}

type searchRepoArgs struct {
	ID                    *string
	Limit                 *int
	Offset                *int
	Race                  *string
	Sex                   *string
	HasMatched            *bool
	AgeInMonthGreaterThan *int
	AgeInMonthLessThan    *int
	AgeInMonth            *int
	UserID                *string
	ExcludeUserID         *string
	NameQuery             *string
	IncludeDeleted        bool
}

func (s SQL) Search(ctx context.Context, args searchRepoArgs) ([]Cat, error) {
	var (
		cats         []Cat
		query        strings.Builder
		whereQueries []string
		sqlArgs      []any

		arg = 1
	)

	query.WriteString(`
		select 
			id, user_id, race, sex, name, age_in_month, match_count,
			description, image_urls, has_matched, created_at
		from cats
	`)

	if !args.IncludeDeleted {
		whereQueries = append(whereQueries, fmt.Sprintf("is_deleted = $%d", arg))
		sqlArgs = append(sqlArgs, false)
		arg += 1
	}

	if args.AgeInMonth != nil {
		whereQueries = append(whereQueries, fmt.Sprintf("age_in_month = $%d", arg))
		sqlArgs = append(sqlArgs, *args.AgeInMonth)
		arg += 1
	} else if args.AgeInMonthGreaterThan != nil {
		whereQueries = append(whereQueries, fmt.Sprintf("age_in_month > $%d", arg))
		sqlArgs = append(sqlArgs, *args.AgeInMonthGreaterThan)
		arg += 1
	} else if args.AgeInMonthLessThan != nil {
		whereQueries = append(whereQueries, fmt.Sprintf("age_in_month < $%d", arg))
		sqlArgs = append(sqlArgs, *args.AgeInMonthLessThan)
		arg += 1
	}

	if args.HasMatched != nil {
		whereQueries = append(whereQueries, fmt.Sprintf("has_matched = $%d", arg))
		sqlArgs = append(sqlArgs, *args.HasMatched)
		arg += 1
	}
	if args.ID != nil {
		whereQueries = append(whereQueries, fmt.Sprintf("id = $%d", arg))
		sqlArgs = append(sqlArgs, *args.ID)
		arg += 1
	}
	if args.Race != nil {
		whereQueries = append(whereQueries, fmt.Sprintf("race = $%d", arg))
		sqlArgs = append(sqlArgs, *args.Race)
		arg += 1
	}
	if args.Sex != nil {
		whereQueries = append(whereQueries, fmt.Sprintf("sex = $%d", arg))
		sqlArgs = append(sqlArgs, *args.Sex)
		arg += 1
	}
	if args.NameQuery != nil {
		whereQueries = append(whereQueries, fmt.Sprintf("name_normalized like $%d", arg))
		sqlArgs = append(sqlArgs, fmt.Sprintf("%%%s%%", strings.ToLower(*args.NameQuery)))
		arg += 1
	}

	if args.UserID != nil {
		whereQueries = append(whereQueries, fmt.Sprintf("user_id = $%d", arg))
		sqlArgs = append(sqlArgs, *args.UserID)
		arg += 1
	} else if args.ExcludeUserID != nil {
		whereQueries = append(whereQueries, fmt.Sprintf("user_id != $%d", arg))
		sqlArgs = append(sqlArgs, *args.ExcludeUserID)
		arg += 1
	}

	if len(whereQueries) > 0 {
		query.WriteString(fmt.Sprintf(`
			where %s
		`, strings.Join(whereQueries, " and ")))
	}

	query.WriteString(`
		order by id desc
	`)

	if args.Limit != nil {
		query.WriteString(fmt.Sprintf(`
			limit $%d
		`, arg))
		sqlArgs = append(sqlArgs, *args.Limit)
		arg += 1
	}

	if args.Offset != nil {
		query.WriteString(fmt.Sprintf(`
			offset $%d
		`, arg))
		sqlArgs = append(sqlArgs, *args.Offset)
		arg += 1
	}

	db := s.pgxTrx.FromContext(ctx)
	rows, err := db.Query(ctx, query.String(), sqlArgs...)
	if err != nil {
		return nil, fmt.Errorf("sql search cat: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var c Cat
		err = rows.Scan(
			&c.ID, &c.UserID, &c.Race, &c.Sex, &c.Name, &c.AgeInMonth, &c.MatchCount,
			&c.Description, &c.ImageURLs, &c.HasMatched, &c.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("sql search cat: %w", err)
		}

		cats = append(cats, c)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("sql search cat: %w", rows.Err())
	}

	return cats, nil
}

type getOneByIDRepoArgs struct {
	ID        int
	ForUpdate bool
}

func (s SQL) GetOneByID(ctx context.Context, args getOneByIDRepoArgs) (Cat, error) {
	db := s.pgxTrx.FromContext(ctx)

	forUpdate := ""
	if args.ForUpdate {
		forUpdate = "for update"
	}

	var c Cat
	err := db.QueryRow(ctx, fmt.Sprintf(`
		select
			id, user_id, race, sex, name, age_in_month, match_count,
			description, image_urls, has_matched, created_at
		from cats
		where id = $1
		and is_deleted = false %s
	`, forUpdate), args.ID).Scan(&c.ID, &c.UserID, &c.Race, &c.Sex, &c.Name, &c.AgeInMonth, &c.MatchCount,
		&c.Description, &c.ImageURLs, &c.HasMatched, &c.CreatedAt)
	if err != nil {
		e := err
		if err == pgx.ErrNoRows {
			e = ErrCatNotFound
		}
		return c, fmt.Errorf("sql finding cat by id: %w", e)
	}

	return c, nil
}

type getByIDsRepoArgs struct {
	IDs       []int
	ForUpdate bool
}

func (s SQL) GetByIDs(ctx context.Context, args getByIDsRepoArgs) ([]Cat, error) {
	db := s.pgxTrx.FromContext(ctx)

	forUpdate := ""
	if args.ForUpdate {
		forUpdate = "for update"
	}

	var cats []Cat
	rows, err := db.Query(ctx, fmt.Sprintf(`
		select
			id, user_id, race, sex, name, age_in_month, match_count,
			description, image_urls, has_matched, created_at
		from cats
		where id = any($1)
		and is_deleted = false %s
	`, forUpdate), args.IDs)
	if err != nil {
		return nil, fmt.Errorf("sql get cats by ids: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var c Cat
		err = rows.Scan(
			&c.ID, &c.UserID, &c.Race, &c.Sex, &c.Name, &c.AgeInMonth, &c.MatchCount,
			&c.Description, &c.ImageURLs, &c.HasMatched, &c.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("sql get cats by ids: %w", err)
		}

		cats = append(cats, c)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("sql get cats by ids: %w", rows.Err())
	}

	return cats, nil
}

type UpdateRepoArgs struct {
	IDs           []int
	HasMatched    *bool
	Name          *string
	Race          *string
	Sex           *string
	AgeInMonth    *int
	Description   *string
	ImageURLs     []string
	IsDeleted     *bool
	IncMatchCount *int
	MatchCount    *int
}

func (s SQL) Update(ctx context.Context, args UpdateRepoArgs) error {
	var (
		query         strings.Builder
		sqlArgs       []any
		updateQueries []string

		arg = 1
	)
	query.WriteString("update cats")

	if args.HasMatched != nil {
		updateQueries = append(updateQueries, fmt.Sprintf(`
			has_matched = $%d
		`, arg))
		sqlArgs = append(sqlArgs, *args.HasMatched)
		arg += 1
	}

	if args.Name != nil {
		updateQueries = append(updateQueries, fmt.Sprintf(`
			name = $%d
		`, arg))
		sqlArgs = append(sqlArgs, *args.Name)
		arg += 1
	}

	if args.Race != nil {
		updateQueries = append(updateQueries, fmt.Sprintf(`
			race = $%d
		`, arg))
		sqlArgs = append(sqlArgs, *args.Race)
		arg += 1
	}

	if args.Sex != nil {
		updateQueries = append(updateQueries, fmt.Sprintf(`
			sex = $%d
		`, arg))
		sqlArgs = append(sqlArgs, *args.Sex)
		arg += 1
	}

	if args.AgeInMonth != nil {
		updateQueries = append(updateQueries, fmt.Sprintf(`
			age_in_month = $%d
		`, arg))
		sqlArgs = append(sqlArgs, *args.AgeInMonth)
		arg += 1
	}

	if args.Description != nil {
		updateQueries = append(updateQueries, fmt.Sprintf(`
			description = $%d
		`, arg))
		sqlArgs = append(sqlArgs, *args.Description)
		arg += 1
	}

	if args.IsDeleted != nil {
		updateQueries = append(updateQueries, fmt.Sprintf(`
			is_deleted = $%d
		`, arg))
		sqlArgs = append(sqlArgs, *args.IsDeleted)
		arg += 1
	}

	if args.IncMatchCount != nil {
		updateQueries = append(updateQueries, fmt.Sprintf(`
			match_count = match_count + $%d
		`, arg))
		sqlArgs = append(sqlArgs, *args.IncMatchCount)
		arg += 1
	}

	if args.MatchCount != nil {
		updateQueries = append(updateQueries, fmt.Sprintf(`
			match_count = $%d
		`, arg))
		sqlArgs = append(sqlArgs, *args.MatchCount)
		arg += 1
	}

	if args.ImageURLs != nil && len(args.ImageURLs) > 0 {
		updateQueries = append(updateQueries, fmt.Sprintf(`
			image_urls = $%d
		`, arg))
		sqlArgs = append(sqlArgs, args.ImageURLs)
		arg += 1
	}

	if len(updateQueries) > 0 {
		query.WriteString(fmt.Sprintf(`
			set %s
		`, strings.Join(updateQueries, ", ")))
	}

	query.WriteString(fmt.Sprintf(`
		where id = any($%d)
	`, arg))
	sqlArgs = append(sqlArgs, args.IDs)

	db := s.pgxTrx.FromContext(ctx)
	_, err := db.Exec(ctx, query.String(), sqlArgs...)
	if err != nil {
		return fmt.Errorf("sql update cats by ids: %w", err)
	}

	return nil
}
