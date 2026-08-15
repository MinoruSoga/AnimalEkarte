package medicalrecord

import (
	"errors"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

type medicalRecordDeleteConflictKind uint8

const (
	medicalRecordDeleteStateConflict medicalRecordDeleteConflictKind = iota + 1
	medicalRecordDeleteDependencyConflict
)

type medicalRecordDeleteConflictCause struct {
	kind medicalRecordDeleteConflictKind
}

func (e *medicalRecordDeleteConflictCause) Error() string {
	return apperrors.ErrConflict.Error()
}

func (e *medicalRecordDeleteConflictCause) Is(target error) bool {
	return target == apperrors.ErrConflict
}

func wrapMedicalRecordDeleteConflict(kind medicalRecordDeleteConflictKind, message string) error {
	return &apperrors.AppError{
		Code:    "CONFLICT",
		Message: message,
		Err:     &medicalRecordDeleteConflictCause{kind: kind},
	}
}

func medicalRecordDeleteConflictKindFromError(err error) (medicalRecordDeleteConflictKind, bool) {
	var cause *medicalRecordDeleteConflictCause
	if !errors.As(err, &cause) {
		return 0, false
	}
	return cause.kind, true
}
