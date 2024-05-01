package user

import "time"

type (
	User struct {
		ID             int
		Email          string
		Name           string
		HashedPassword string
		CreatedAt      time.Time
	}
)
