package match

import "errors"

var (
	ErrCatHasBeenMatched    = errors.New("cat has been matched")
	ErrCatsFromTheSameOwner = errors.New("cats are from the same owner")
	ErrCatsHaveSameGender   = errors.New("cats have the same gender")
	ErrUserDoesNotOwnCat    = errors.New("user does not own user cat")
)
