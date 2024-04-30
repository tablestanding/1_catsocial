package user

import (
	"context"

	"golang.org/x/crypto/bcrypt"
)

type (
	RegisterArgs struct {
		Email    string
		Name     string
		Password string
	}

	Repo interface {
		Create(context.Context, CreateUserRepoArgs) error
	}

	Service struct {
		r         Repo
		saltCount int
	}
)

func NewService(r Repo, saltCount int) Service {
	return Service{r: r, saltCount: saltCount}
}

func (s Service) Register(ctx context.Context, args RegisterArgs) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(args.Password), s.saltCount)
	if err != nil {
		return err
	}

	err = s.r.Create(ctx, CreateUserRepoArgs{
		Email:          args.Email,
		HashedPassword: string(hashedPassword),
		Name:           args.Name,
	})
	if err != nil {
		return err
	}

	return nil
}
