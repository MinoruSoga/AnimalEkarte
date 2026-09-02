package httpapi

import (
	"fmt"
	"slices"
)

// ValidateEnum accepts v when it is one of allowed. Callers wrap the error
// with the domain-specific invalid-input message.
func ValidateEnum[T ~string](v string, allowed ...T) (T, error) {
	if slices.Contains(allowed, T(v)) {
		return T(v), nil
	}
	var zero T
	return zero, fmt.Errorf("invalid value %q", v)
}
