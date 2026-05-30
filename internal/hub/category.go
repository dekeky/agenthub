package hub

import (
	"errors"
	"fmt"
	"strings"

	"github.com/agenthub/internal/archive"
)

const (
	CategoryPicoClaw = "picoclaw"
	CategoryOpenClaw = "openclaw"
)

// KnownCategories lists commonly used agent runtime categories.
var KnownCategories = []string{CategoryPicoClaw, CategoryOpenClaw}

var ErrInvalidCategory = errors.New("invalid category")

// NormalizeCategory validates and normalizes an agent category name.
// Empty input defaults to picoclaw.
func NormalizeCategory(category string) (string, error) {
	c := strings.ToLower(strings.TrimSpace(category))
	if c == "" {
		return CategoryPicoClaw, nil
	}
	if err := archive.ValidateIdentifier(c); err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidCategory, err)
	}
	return c, nil
}

// IsInvalidCategory reports whether err was caused by a bad category value.
func IsInvalidCategory(err error) bool {
	return errors.Is(err, ErrInvalidCategory)
}
