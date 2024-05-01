package cat

import "time"

var (
	races = []string{
		"Persian",
		"Maine Coon",
		"Siamese",
		"Ragdoll",
		"Bengal",
		"Sphynx",
		"British Shorthair",
		"Abyssinian",
		"Scottish Fold",
		"Birman",
	}
)

type (
	Cat struct {
		ID          int
		UserID      string
		Race        string
		Sex         string
		AgeInMonth  int
		Description string
		ImageURLs   []string
		HasMatched  bool
		Name        string
		CreatedAt   time.Time
	}
)
