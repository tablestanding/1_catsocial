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
		Create(ctx context.Context, args CreateRepoArgs) error
		Get(ctx context.Context, args GetRepoArgs) ([]Match, error)
		GetByID(ctx context.Context, id int) (MatchRaw, error)
		GetByCatID(ctx context.Context, catID int) (MatchRaw, error)
		Update(ctx context.Context, id int, args UpdateRepoArgs) error
		Delete(ctx context.Context, args DeleteRepoArgs) error
	}

	catSvc interface {
		GetByIDs(ctx context.Context, ids ...string) ([]cat.Cat, error)
		Update(ctx context.Context, args cat.UpdateArgs) error
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
		return fmt.Errorf("create match: get cat by ids: %w", err)
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

	err = s.catSvc.Update(ctx, cat.UpdateArgs{
		IDs:           []int{userCat.ID, matchCat.ID},
		IncMatchCount: pointer.Pointer(1),
	})
	if err != nil {
		return fmt.Errorf("create match: increment cats match count: %w", err)
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

	matchRaw, err := s.matchRepo.GetByID(ctx, intID)
	if err != nil {
		return fmt.Errorf("approve match: get match by id: %w", err)
	}
	if matchRaw.HasBeenApprovedOrRejected {
		return fmt.Errorf("approve match: %w", ErrMatchNotValid)
	}

	err = s.matchRepo.Update(ctx, intID, UpdateRepoArgs{
		HasBeenApprovedOrRejected: pointer.Pointer(true),
	})
	if err != nil {
		return fmt.Errorf("approve match: update match: %w", err)
	}

	err = s.matchRepo.Delete(ctx, DeleteRepoArgs{
		CatIDs:         []int{matchRaw.IssuerCatID, matchRaw.ReceiverCatID},
		ExcludeMatchID: pointer.Pointer(matchRaw.ID),
	})
	if err != nil {
		return fmt.Errorf("approve match: delete other matches: %w", err)
	}

	err = s.catSvc.Update(ctx, cat.UpdateArgs{
		IDs:        []int{matchRaw.IssuerCatID, matchRaw.ReceiverCatID},
		HasMatched: pointer.Pointer(true),
		MatchCount: pointer.Pointer(1),
	})
	if err != nil {
		return fmt.Errorf("approve match: update cats: %w", err)
	}

	return nil
}

func (s Service) Reject(ctx context.Context, matchID string) error {
	intID, err := strconv.Atoi(matchID)
	if err != nil {
		return fmt.Errorf("reject match: match id is not valid: %w", err)
	}

	matchRaw, err := s.matchRepo.GetByID(ctx, intID)
	if err != nil {
		return fmt.Errorf("reject match: get match by id: %w", err)
	}
	if matchRaw.HasBeenApprovedOrRejected {
		return fmt.Errorf("reject match: %w", ErrMatchNotValid)
	}

	err = s.matchRepo.Update(ctx, intID, UpdateRepoArgs{
		HasBeenApprovedOrRejected: pointer.Pointer(true),
	})
	if err != nil {
		return fmt.Errorf("reject match: update matches: %w", err)
	}

	return nil
}

type DeleteArgs struct {
	MatchID int
	UserID  string
}

func (s Service) Delete(ctx context.Context, args DeleteArgs) error {
	matchRaw, err := s.matchRepo.GetByID(ctx, args.MatchID)
	if err != nil {
		return fmt.Errorf("delete match: get match by id: %w", err)
	}
	if matchRaw.HasBeenApprovedOrRejected {
		return fmt.Errorf("delete match: %w", ErrMatchNotValid)
	}
	if strconv.Itoa(matchRaw.IssuerUserID) != args.UserID {
		return fmt.Errorf("delete match: %w", ErrMatchNotFound)
	}

	err = s.matchRepo.Delete(ctx, DeleteRepoArgs{
		MatchID: &args.MatchID,
	})
	if err != nil {
		return fmt.Errorf("delete match: %w", err)
	}

	err = s.catSvc.Update(ctx, cat.UpdateArgs{
		IDs:           []int{matchRaw.IssuerCatID, matchRaw.ReceiverCatID},
		IncMatchCount: pointer.Pointer(-1),
	})
	if err != nil {
		return fmt.Errorf("delete match: decrement cats match count: %w", err)
	}

	return nil
}
