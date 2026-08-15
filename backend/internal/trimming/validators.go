package trimming

import (
	"context"

	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

const (
	ErrMsgAtLeastOneField = sharedkernel.ErrMsgAtLeastOneField
	ErrMsgIDsNotEmpty     = sharedkernel.ErrMsgIDsNotEmpty
	ErrMsgInputNotNil     = sharedkernel.ErrMsgInputNotNil
)

func validateRequiredName(name string) error {
	return sharedkernel.ValidateRequiredName(name)
}

func validateOptionalName(name *string) error {
	return sharedkernel.ValidateOptionalName(name)
}

func validateOwnedMasterFK(
	ctx context.Context,
	entity string,
	clinicID uint64,
	id *uint64,
	find func(context.Context, uint64, uint64) error,
) error {
	return sharedkernel.ValidateOwnedMasterFK(ctx, entity, clinicID, id, find)
}

func validateOwnedMasterFKs(
	ctx context.Context,
	entity string,
	clinicID uint64,
	ids []uint64,
	find func(context.Context, uint64, uint64) error,
) error {
	return sharedkernel.ValidateOwnedMasterFKs(ctx, entity, clinicID, ids, find)
}
