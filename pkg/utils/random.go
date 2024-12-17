package utils

import (
	"github.com/oklog/ulid/v2"
)

func GenerateUID() string {
	return ulid.Make().String()
}
