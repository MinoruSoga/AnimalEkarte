package owner

import (
	"fmt"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// buildOwnerUpdate keeps existing builder-test map assertions while routing
// through the typed UpdateCommand converter (BE-RC-017).
func buildOwnerUpdate(input *UpdateOwnerInput) map[string]any {
	return updateCommandFields(updateCommandFromOwnerInput(input))
}

func normalizeOwnerReason(reason, fieldName string) (*string, error) {
	trimmed := strings.TrimSpace(reason)
	if len([]rune(trimmed)) > ownerDeliveryReasonMaxLength {
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("%s must be %d characters or less", fieldName, ownerDeliveryReasonMaxLength))
	}
	if trimmed == "" {
		return nil, nil
	}
	return &trimmed, nil
}
