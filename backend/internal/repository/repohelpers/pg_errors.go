package repohelpers

import (
	"github.com/animal-ekarte/backend/internal/persistence"
)

// IsUniqueConstraintErr delegates to persistence.
func IsUniqueConstraintErr(err error) bool {
	return persistence.IsUniqueConstraintErr(err)
}

// IsFKConstraintErr delegates to persistence.
func IsFKConstraintErr(err error) bool {
	return persistence.IsFKConstraintErr(err)
}
