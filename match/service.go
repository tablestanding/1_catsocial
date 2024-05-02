package match

import (
	"catsocial/cat"
	"catsocial/pkg/pointer"
	"context"
	"fmt"
	"strconv"
)

type (
	matchRepo interface {
		Create(ctx context.Context, args createRepoArgs) error
		Get(ctx context.Context, args getRepoArgs) ([]Match, error)
		GetByID(ctx context.Context, args getByIDRepoArgs) (MatchRaw, error)
		GetByCatID(ctx context.Context, catID int) (MatchRaw, error)
		Update(ctx context.Context, args updateRepoArgs) error
		Delete(ctx context.Context, args deleteRepoArgs) error
	}

	catSvc interface {
		GetByIDs(ctx context.Context, args cat.GetByIDsArgs) ([]cat.Cat, error)
		Update(ctx context.Context, args cat.UpdateArgs) error
	}

	trx interface {
		WithTransaction(ctx context.Context, fn func(context.Context) error) error
	}

	Service struct {
		matchRepo matchRepo
		catSvc    catSvc
		trx       trx
	}
)

func NewService(matchRepo matchRepo, catSvc catSvc, trx trx) Service {
	return Service{matchRepo: matchRepo, catSvc: catSvc, trx: trx}
}

type CreateArgs struct {
	MatchCatID string
	UserCatID  string
	UserID     string
	Msg        string
}

func (s Service) Create(ctx context.Context, args CreateArgs) error {
	err := s.trx.WithTransaction(ctx, func(ctx context.Context) error {
		cats, err := s.catSvc.GetByIDs(ctx, cat.GetByIDsArgs{
			IDs:       []string{args.MatchCatID, args.UserCatID},
			ForUpdate: true,
		})
		if err != nil {
			return fmt.Errorf("get cat by ids: %w", err)
		}

		// must be 2 valid cats
		if len(cats) != 2 {
			return cat.ErrCatNotFound
		}

		// both cats must have not been matched
		if cats[0].HasMatched || cats[1].HasMatched {
			return ErrCatHasBeenMatched
		}

		// the cats are from different owner
		if cats[0].UserID == cats[1].UserID {
			return ErrCatsFromTheSameOwner
		}

		// the cats must have different gender
		if cats[0].Sex == cats[1].Sex {
			return ErrCatsHaveSameGender
		}

		// the current user must own the user cat
		userCat := cats[0]
		matchCat := cats[1]
		if strconv.Itoa(cats[1].ID) == args.UserCatID {
			userCat = cats[1]
			matchCat = cats[0]
		}
		if userCat.UserID != args.UserID {
			return ErrUserDoesNotOwnCat
		}

		err = s.matchRepo.Create(ctx, createRepoArgs{
			IssuerUserID:   userCat.UserID,
			ReceiverUserID: matchCat.UserID,
			IssuerCatID:    strconv.Itoa(userCat.ID),
			ReceiverCatID:  strconv.Itoa(matchCat.ID),
			Msg:            args.Msg,
		})
		if err != nil {
			return err
		}

		err = s.catSvc.Update(ctx, cat.UpdateArgs{
			IDs:           []int{userCat.ID, matchCat.ID},
			IncMatchCount: pointer.Pointer(1),
		})
		if err != nil {
			return fmt.Errorf("increment cats match count: %w", err)
		}

		return nil
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
	matches, err := s.matchRepo.Get(ctx, getRepoArgs{UserID: args.UserID})
	if err != nil {
		return nil, fmt.Errorf("get match: %w", err)
	}

	return matches, nil
}

func (s Service) GetByCatID(ctx context.Context, catID int) (MatchRaw, error) {
	matches, err := s.matchRepo.GetByCatID(ctx, catID)
	if err != nil {
		return MatchRaw{}, fmt.Errorf("get match by cat id: %w", err)
	}

	return matches, nil
}

func (s Service) Approve(ctx context.Context, matchID string) error {
	intID, err := strconv.Atoi(matchID)
	if err != nil {
		return fmt.Errorf("approve match: match id is not valid: %w", err)
	}

	err = s.trx.WithTransaction(ctx, func(ctx context.Context) error {
		matchRaw, err := s.matchRepo.GetByID(ctx, getByIDRepoArgs{
			ID:            intID,
			ForUpdateCats: true,
		})
		if err != nil {
			return fmt.Errorf("get match by id: %w", err)
		}
		if matchRaw.HasBeenApprovedOrRejected {
			return ErrMatchNotValid
		}

		err = s.matchRepo.Update(ctx, updateRepoArgs{
			ID:                        intID,
			HasBeenApprovedOrRejected: pointer.Pointer(true),
		})
		if err != nil {
			return fmt.Errorf("update match: %w", err)
		}

		err = s.matchRepo.Delete(ctx, deleteRepoArgs{
			CatIDs:         []int{matchRaw.IssuerCatID, matchRaw.ReceiverCatID},
			ExcludeMatchID: pointer.Pointer(matchRaw.ID),
		})
		if err != nil {
			return fmt.Errorf("delete other matches: %w", err)
		}

		err = s.catSvc.Update(ctx, cat.UpdateArgs{
			IDs:        []int{matchRaw.IssuerCatID, matchRaw.ReceiverCatID},
			HasMatched: pointer.Pointer(true),
			MatchCount: pointer.Pointer(1),
		})
		if err != nil {
			return fmt.Errorf("update cats: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("approve match: %w", err)
	}

	return nil
}

func (s Service) Reject(ctx context.Context, matchID string) error {
	intID, err := strconv.Atoi(matchID)
	if err != nil {
		return fmt.Errorf("reject match: match id is not valid: %w", err)
	}

	err = s.trx.WithTransaction(ctx, func(ctx context.Context) error {
		matchRaw, err := s.matchRepo.GetByID(ctx, getByIDRepoArgs{
			ID:            intID,
			ForUpdateCats: true,
		})
		if err != nil {
			return fmt.Errorf("get match by id: %w", err)
		}
		if matchRaw.HasBeenApprovedOrRejected {
			return ErrMatchNotValid
		}

		err = s.matchRepo.Update(ctx, updateRepoArgs{
			ID:                        intID,
			HasBeenApprovedOrRejected: pointer.Pointer(true),
		})
		if err != nil {
			return fmt.Errorf("update matches: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("reject match: %w", err)
	}

	return nil
}

type DeleteArgs struct {
	MatchID int
	UserID  string
}

func (s Service) Delete(ctx context.Context, args DeleteArgs) error {
	err := s.trx.WithTransaction(ctx, func(ctx context.Context) error {
		matchRaw, err := s.matchRepo.GetByID(ctx, getByIDRepoArgs{
			ID:            args.MatchID,
			ForUpdateCats: true,
		})
		if err != nil {
			return fmt.Errorf("get match by id: %w", err)
		}
		if matchRaw.HasBeenApprovedOrRejected {
			return ErrMatchNotValid
		}
		if strconv.Itoa(matchRaw.IssuerUserID) != args.UserID {
			return ErrMatchNotFound
		}

		err = s.matchRepo.Delete(ctx, deleteRepoArgs{
			MatchID: &args.MatchID,
		})
		if err != nil {
			return err
		}

		err = s.catSvc.Update(ctx, cat.UpdateArgs{
			IDs:           []int{matchRaw.IssuerCatID, matchRaw.ReceiverCatID},
			IncMatchCount: pointer.Pointer(-1),
		})
		if err != nil {
			return fmt.Errorf("decrement cats match count: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("delete match: %w", err)
	}

	return nil
}
