package utils

import (
	"errors"
	"strconv"
)

// PositiveInt parses a positive integer or returns fallback when raw is empty.
func PositiveInt(raw string, fallback int) (int, error) {
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, errors.New("invalid positive integer")
	}
	return value, nil
}
