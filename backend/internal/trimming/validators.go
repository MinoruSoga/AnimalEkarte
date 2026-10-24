package trimming

import (
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

func validateNonNegativePrice(price *int64) error {
	return sharedkernel.ValidateNonNegativePrice(price)
}
