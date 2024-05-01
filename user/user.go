package user

type (
	User struct {
		ID             int
		Email          string
		Name           string
		HashedPassword string
	}
)
