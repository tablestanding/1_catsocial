package cat

import (
	"catsocial/pkg/pointer"
	"context"
	"errors"
	"fmt"
	"strconv"
)

type (
	repo interface {
		Create(ctx context.Context, args createRepoArgs) (Cat, error)
		Search(ctx context.Context, args searchRepoArgs) ([]Cat, error)
		GetOneByID(ctx context.Context, args getOneByIDRepoArgs) (Cat, error)
		GetByIDs(ctx context.Context, args getByIDsRepoArgs) ([]Cat, error)
		Update(ctx context.Context, args UpdateRepoArgs) error
	}

	trx interface {
		WithTransaction(ctx context.Context, fn func(context.Context) error) error
	}

	Service struct {
		r   repo
		trx trx
	}
)

func NewService(r repo, trx trx) Service {
	return Service{r: r, trx: trx}
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
	return s.r.Create(ctx, createRepoArgs{
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
	ExcludeUserID         *string
	NameQuery             *string
}

func (s Service) Search(ctx context.Context, args SearchArgs) ([]Cat, error) {
	return s.r.Search(ctx, searchRepoArgs{
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
		ExcludeUserID:         args.ExcludeUserID,
		NameQuery:             args.NameQuery,
	})
}

type GetOneByIDArgs struct {
	ID        string
	ForUpdate bool
}

func (s Service) GetOneByID(ctx context.Context, args GetOneByIDArgs) (Cat, error) {
	i, err := strconv.Atoi(args.ID)
	if err != nil {
		return Cat{}, fmt.Errorf("get cat by id: %w", err)
	}

	return s.r.GetOneByID(ctx, getOneByIDRepoArgs{
		ID:        i,
		ForUpdate: args.ForUpdate,
	})
}

type GetByIDsArgs struct {
	IDs       []string
	ForUpdate bool
}

func (s Service) GetByIDs(ctx context.Context, args GetByIDsArgs) ([]Cat, error) {
	var intIds []int
	for _, id := range args.IDs {
		i, err := strconv.Atoi(id)
		if err != nil {
			return nil, fmt.Errorf("get cat by ids: invalid id: %w", ErrCatNotFound)
		}

		intIds = append(intIds, i)
	}

	return s.r.GetByIDs(ctx, getByIDsRepoArgs{
		IDs:       intIds,
		ForUpdate: args.ForUpdate,
	})
}

type UpdateArgs struct {
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

func (s Service) Update(ctx context.Context, args UpdateArgs) error {
	err := s.trx.WithTransaction(ctx, func(ctx context.Context) error {
		cats, err := s.r.GetByIDs(ctx, getByIDsRepoArgs{
			IDs:       args.IDs,
			ForUpdate: true,
		})
		if errors.Is(err, ErrCatNotFound) {
			return fmt.Errorf("get cat by ids: %w", ErrCatNotFound)
		}
		if err != nil {
			return fmt.Errorf("get cat by ids: %w", err)
		}

		// cat sex could not be edited after match has been requested
		if args.Sex != nil {
			for _, cat := range cats {
				if cat.MatchCount > 0 && *args.Sex != cat.Sex {
					return ErrCatSexEditedAfterMatchRequested
				}
			}
		}

		err = s.r.Update(ctx, UpdateRepoArgs{
			IDs:         args.IDs,
			Race:        args.Race,
			Sex:         args.Sex,
			Name:        args.Name,
			AgeInMonth:  args.AgeInMonth,
			Description: args.Description,
			ImageURLs:   args.ImageURLs,
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("update cats: %w", err)
	}

	return nil
}

func (s Service) Delete(ctx context.Context, id int) error {
	err := s.trx.WithTransaction(ctx, func(ctx context.Context) error {
		_, err := s.r.GetOneByID(ctx, getOneByIDRepoArgs{
			ID:        id,
			ForUpdate: true,
		})
		if errors.Is(err, ErrCatNotFound) {
			return fmt.Errorf("get cat by id: %w", ErrCatNotFound)
		}
		if err != nil {
			return fmt.Errorf("get cat by id: %w", err)
		}

		err = s.r.Update(ctx, UpdateRepoArgs{
			IDs:       []int{id},
			IsDeleted: pointer.Pointer(true),
		})
		if err != nil {
			return fmt.Errorf("update cat: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("delete cat: %w", err)
	}

	return nil
}
