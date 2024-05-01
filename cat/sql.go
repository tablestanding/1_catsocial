package cat

import (
	"context"
	"fmt"
	"strings"

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

	SearchCatRepoArgs struct {
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
		Name:        args.Name,
	}
	err := s.pool.QueryRow(ctx, `
		insert into cats(user_id, race, sex, age_in_month, description, image_urls, name)
		values ($1, $2, $3, $4, $5, $6, $7)
		returning id, created_at, has_matched
	`, args.UserID, args.Race, args.Sex, args.AgeInMonth, args.Description, args.ImageURLs, args.Name).
		Scan(&c.ID, &c.CreatedAt, &c.HasMatched)

	return c, err
}

func (s SQL) Search(ctx context.Context, args SearchCatRepoArgs) ([]Cat, error) {
	var (
		cats         []Cat
		query        strings.Builder
		whereQueries []string
		sqlArgs      []any

		arg = 1
	)

	query.WriteString(`
		select 
			id, user_id, race, sex, name, age_in_month,
			description, image_urls, has_matched, created_at
		from cats
	`)

	if args.AgeInMonth != nil {
		whereQueries = append(whereQueries, fmt.Sprintf("age_in_month = %d", arg))
		sqlArgs = append(sqlArgs, *args.AgeInMonth)
		arg += 1
	} else if args.AgeInMonthGreaterThan != nil {
		whereQueries = append(whereQueries, fmt.Sprintf("age_in_month > %d", arg))
		sqlArgs = append(sqlArgs, *args.AgeInMonthGreaterThan)
		arg += 1
	} else if args.AgeInMonthLessThan != nil {
		whereQueries = append(whereQueries, fmt.Sprintf("age_in_month < %d", arg))
		sqlArgs = append(sqlArgs, *args.AgeInMonthLessThan)
		arg += 1
	}

	if args.HasMatched != nil {
		whereQueries = append(whereQueries, fmt.Sprintf("has_matched = %d", arg))
		sqlArgs = append(sqlArgs, *args.HasMatched)
		arg += 1
	}
	if args.ID != nil {
		whereQueries = append(whereQueries, fmt.Sprintf("id = %d", arg))
		sqlArgs = append(sqlArgs, *args.ID)
		arg += 1
	}
	if args.Race != nil {
		whereQueries = append(whereQueries, fmt.Sprintf("race = %d", arg))
		sqlArgs = append(sqlArgs, *args.Race)
		arg += 1
	}
	if args.Sex != nil {
		whereQueries = append(whereQueries, fmt.Sprintf("sex = %d", arg))
		sqlArgs = append(sqlArgs, *args.Sex)
		arg += 1
	}

	if args.UserID != nil {
		whereQueries = append(whereQueries, fmt.Sprintf("user_id = %d", arg))
		sqlArgs = append(sqlArgs, *args.UserID)
		arg += 1
	} else if args.ExcludeUserID != nil {
		whereQueries = append(whereQueries, fmt.Sprintf("user_id != %d", arg))
		sqlArgs = append(sqlArgs, *args.ExcludeUserID)
		arg += 1
	}

	query.WriteString(fmt.Sprintf(`
		where %s
	`, strings.Join(whereQueries, " and ")))

	if args.Limit != nil {
		query.WriteString(fmt.Sprintf(`
			limit %d
		`, *args.Limit))
	}

	if args.Offset != nil {
		query.WriteString(fmt.Sprintf(`
			offset %d
		`, *args.Offset))
	}

	rows, err := s.pool.Query(ctx, query.String(), sqlArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c Cat
		err = rows.Scan(
			&c.ID, &c.UserID, &c.Race, &c.Sex, &c.Name, &c.AgeInMonth,
			&c.Description, &c.ImageURLs, &c.HasMatched, &c.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		cats = append(cats, c)
	}
	if err == nil {
		err = rows.Err()
	}

	return cats, nil
}
