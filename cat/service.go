package cat

import (
	"context"
)

type (
	repo interface {
		Create(ctx context.Context, args CreateCatRepoArgs) (Cat, error)
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
