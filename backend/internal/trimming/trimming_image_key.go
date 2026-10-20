package trimming

import (
	"context"
	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ImageKey is a canonical, clinic-owned storage reference. Only this package
// constructs it; lifecycle operations cannot accept a foreign clinic or URL.
type ImageKey struct {
	clinicID uint64
	value    string
}

// ImageLifecycleRepository is the narrow capability shared by reference writes
// and deletion reservations. Locks must last until the ambient transaction ends.
type ImageLifecycleRepository interface {
	LockImageKeys(context.Context, []ImageKey) error
	ImageKeyDeleted(context.Context, ImageKey) (bool, error)
	MarkImageKeyDeleted(context.Context, ImageKey, uint64) error
	// ImageKeyInUse includes trimming details, current/soft-deleted course
	// references and retained course image audit history. A replaced original
	// remains protected. exceptAppointmentID excludes one appointment's own
	// eligible detail rows only; it never excludes course image history.
	ImageKeyInUse(context.Context, uint64, string, uint64) (bool, error)
}

func ownedImageKey(clinicID uint64, stored string) (ImageKey, error) {
	key, isObject, err := resolveTrimmingImageStorageKey(stored)
	if err != nil || !isObject || !trimmingImageKeyBelongsToClinic(key, clinicID) {
		return ImageKey{}, apperrors.WrapInvalidInput("invalid trimming image key")
	}
	return ImageKey{clinicID: clinicID, value: key}, nil
}
