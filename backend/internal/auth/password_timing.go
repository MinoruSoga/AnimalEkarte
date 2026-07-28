package auth

import "golang.org/x/crypto/bcrypt"

const (
	// This is a non-secret bcrypt hash for dummy work only. Its cost must remain
	// aligned with config.BcryptCost so unknown-account paths do comparable work.
	dummyPasswordHash      = "$2y$12$pMYecS0CMq3rztzQ8mMuLukZ0UGbYT4/PPHZwZ6CR42PyxfA9z3Aa" //nolint:gosec // public dummy work factor, never authenticates
	dummyPasswordCandidate = "auth-dummy-password-9M!"                                      //nolint:gosec // public dummy input, never authenticates
)

type passwordCompareFunc func(hashedPassword, password []byte) error

func defaultPasswordCompare(
	hashedPassword, password []byte,
) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}

func normalizePasswordComparer(compare passwordCompareFunc) passwordCompareFunc {
	if compare != nil {
		return compare
	}
	return defaultPasswordCompare
}

func performDummyPasswordComparison(
	compare passwordCompareFunc,
	password string,
) {
	_ = normalizePasswordComparer(compare)(
		[]byte(dummyPasswordHash),
		[]byte(password),
	)
}
