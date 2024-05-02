package cat

import (
	"context"
	"fmt"
	"strconv"
)

type (
	repo interface {
		Create(ctx context.Context, args CreateRepoArgs) (Cat, error)
		Search(ctx context.Context, args SearchRepoArgs) ([]Cat, error)
		GetOneByID(ctx context.Context, id int) (Cat, error)
		GetByIDs(ctx context.Context, ids []int) ([]Cat, error)
		Update(ctx context.Context, args UpdateRepoArgs) error
	}

	Service struct {
		r repo
	}
)

func NewService(r repo) Service {
	return Service{r}
}

type CreateArgs struct {
	Race        string
	Sex         string
	Name        string
	AgeInMonth  int
	Description string
	ImageURLs   []string
	UserID      string
}

func (s Service) Create(ctx context.Context, args CreateArgs) (Cat, error) {
	return s.r.Create(ctx, CreateRepoArgs{
		Race:        args.Race,
		Sex:         args.Sex,
		Name:        args.Name,
		AgeInMonth:  args.AgeInMonth,
		Description: args.Description,
		ImageURLs:   args.ImageURLs,
		UserID:      args.UserID,
	})
}

type SearchArgs struct {
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
	NameQuery             *string
}

func (s Service) Search(ctx context.Context, args SearchArgs) ([]Cat, error) {
	return s.r.Search(ctx, SearchRepoArgs{
		ID:                    args.ID,
		Limit:                 args.Limit,
		Offset:                args.Offset,
		Race:                  args.Race,
		Sex:                   args.Sex,
		HasMatched:            args.HasMatched,
		AgeInMonthGreaterThan: args.AgeInMonthGreaterThan,
		AgeInMonthLessThan:    args.AgeInMonthLessThan,
		AgeInMonth:            args.AgeInMonth,
		UserID:                args.UserID,
		NameQuery:             args.NameQuery,
	})
}

func (s Service) GetOneByID(ctx context.Context, id string) (Cat, error) {
	i, err := strconv.Atoi(id)
	if err != nil {
		return Cat{}, fmt.Errorf("get cat by id: %w", err)
	}

	return s.r.GetOneByID(ctx, i)
}

func (s Service) GetByIDs(ctx context.Context, ids ...string) ([]Cat, error) {
	var intIds []int
	for _, id := range ids {
		i, err := strconv.Atoi(id)
		if err != nil {
			return nil, fmt.Errorf("get cat by ids: %w", err)
		}

		intIds = append(intIds, i)
	}

	return s.r.GetByIDs(ctx, intIds)
}

type UpdateArgs struct {
	IDs         []int
	HasMatched  *bool
	Name        *string
	Race        *string
	Sex         *string
	AgeInMonth  *int
	Description *string
	ImageURLs   []string
}

func (s Service) Update(ctx context.Context, args UpdateArgs) error {
	err := s.r.Update(ctx, UpdateRepoArgs{
		IDs:         args.IDs,
		HasMatched:  args.HasMatched,
		Name:        args.Name,
		Race:        args.Race,
		Sex:         args.Sex,
		AgeInMonth:  args.AgeInMonth,
		Description: args.Description,
		ImageURLs:   args.ImageURLs,
	})
	if err != nil {
		return fmt.Errorf("update cats by ids: %w", err)
	}

	return nil
}
