package user

import (
	"catsocial/pkg/jwt"
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type (
	repo interface {
		Create(ctx context.Context, args CreateUserRepoArgs) error
		GetOneByEmail(ctx context.Context, email string) (User, error)
	}

	Service struct {
		r         repo
		saltCount int
		jwtSecret string
	}
)

func NewService(r repo, saltCount int, jwtSecret string) Service {
	return Service{r: r, saltCount: saltCount, jwtSecret: jwtSecret}
}

type RegisterArgs struct {
	Email    string
	Name     string
	Password string
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

type LoginArgs struct {
	Email    string
	Password string
}

func (s Service) Login(ctx context.Context, args LoginArgs) (User, error) {
	u, err := s.r.GetOneByEmail(ctx, args.Email)
	if err != nil {
		return u, fmt.Errorf("login: %w", err)
	}

	pwErr := bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(args.Password))
	if pwErr != nil {
		return u, ErrInvalidPassword
	}

	return u, nil
}

func (s Service) GetAccessToken(ctx context.Context) (string, error) {
	return jwt.GenerateToken(8*time.Hour, s.jwtSecret)
}
