package cat

import (
	"context"
)

type (
	repo interface {
		Create(ctx context.Context, args CreateCatRepoArgs) (Cat, error)
		Search(ctx context.Context, args SearchCatRepoArgs) ([]Cat, error)
	}

	Service struct {
		r repo
	}
)

func NewService(r repo) Service {
	return Service{r}
}

type CreateCatArgs struct {
	Race        string
	Sex         string
	Name        string
	AgeInMonth  int
	Description string
	ImageURLs   []string
	UserID      string
}

func (s Service) Create(ctx context.Context, args CreateCatArgs) (Cat, error) {
	return s.r.Create(ctx, CreateCatRepoArgs{
		Race:        args.Race,
		Sex:         args.Sex,
		Name:        args.Name,
		AgeInMonth:  args.AgeInMonth,
		Description: args.Description,
		ImageURLs:   args.ImageURLs,
		UserID:      args.UserID,
	})
}

type SearchCatArgs struct {
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

func (s Service) Search(ctx context.Context, args SearchCatArgs) ([]Cat, error) {
	return s.r.Search(ctx, SearchCatRepoArgs{
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
