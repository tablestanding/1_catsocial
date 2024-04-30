package env

import (
	"fmt"
	"os"
)

// MustLoad panics when the requested env var is not found
func MustLoad(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("could not find env var: %s", key))
	}

	return val
}
