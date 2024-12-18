package utils

import (
	"github.com/oklog/ulid/v2"
)

func GenerateUID() ulid.ULID {
	return ulid.Make()
}

func GenerateID() string {
	return GenerateUID().String()
}
