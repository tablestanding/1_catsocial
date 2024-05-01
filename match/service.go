package match

import (
	"catsocial/cat"
	"context"
	"fmt"
	"strconv"
)

type (
	matchRepo interface {
		Create(ctx context.Context, args CreateRepoArgs) error
		Get(ctx context.Context, args GetRepoArgs) ([]Match, error)
	}

	catSvc interface {
		GetByIDs(ctx context.Context, ids ...string) ([]cat.Cat, error)
	}

	Service struct {
		matchRepo matchRepo
		catSvc    catSvc
	}
)

func NewService(matchRepo matchRepo, catSvc catSvc) Service {
	return Service{matchRepo: matchRepo, catSvc: catSvc}
}

type CreateArgs struct {
	MatchCatID string
	UserCatID  string
	UserID     string
	Msg        string
}

func (s Service) Create(ctx context.Context, args CreateArgs) error {
	cats, err := s.catSvc.GetByIDs(ctx, args.MatchCatID, args.UserCatID)
	if err != nil {
		return fmt.Errorf("create match: %w", err)
	}

	// must be 2 valid cats
	if len(cats) != 2 {
		return fmt.Errorf("create match: %w", cat.ErrCatNotFound)
	}

	// both cats must have not been matched
	if cats[0].HasMatched || cats[1].HasMatched {
		return fmt.Errorf("create match: %w", ErrCatHasBeenMatched)
	}

	// the cats are from different owner
	if cats[0].UserID == cats[1].UserID {
		return fmt.Errorf("create match: %w", ErrCatsFromTheSameOwner)
	}

	// the cats must have different gender
	if cats[0].Sex == cats[1].Sex {
		return fmt.Errorf("create match: %w", ErrCatsHaveSameGender)
	}

	// the current user must own the user cat
	userCat := cats[0]
	matchCat := cats[1]
	if strconv.Itoa(cats[1].ID) == args.UserCatID {
		userCat = cats[1]
		matchCat = cats[0]
	}
	if userCat.UserID != args.UserID {
		return fmt.Errorf("create match: %w", ErrUserDoesNotOwnCat)
	}

	err = s.matchRepo.Create(ctx, CreateRepoArgs{
		IssuerUserID:   userCat.UserID,
		ReceiverUserID: matchCat.UserID,
		IssuerCatID:    strconv.Itoa(userCat.ID),
		ReceiverCatID:  strconv.Itoa(matchCat.ID),
		Msg:            args.Msg,
	})
	if err != nil {
		return fmt.Errorf("create match: %w", err)
	}

	return nil
}

type GetArgs struct {
	UserID string
}

func (s Service) Get(ctx context.Context, args GetArgs) ([]Match, error) {
	matches, err := s.matchRepo.Get(ctx, GetRepoArgs{UserID: args.UserID})
	if err != nil {
		return nil, fmt.Errorf("get match: %w", err)
	}

	return matches, nil
}
