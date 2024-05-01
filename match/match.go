package match

import (
	"catsocial/cat"
	"catsocial/user"
	"time"
)

type (
	Match struct {
		ID                        int
		IssuerUser                user.User
		ReceiverUser              user.User
		IssuerCat                 cat.Cat
		ReceiverCat               cat.Cat
		HasBeenApprovedOrRejected bool
		CreatedAt                 time.Time
		Msg                       string
	}
)
