package owner

import (
	"fmt"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// buildOwnerUpdate はポインタが非 nil のフィールドのみ map に追加する
func buildOwnerUpdate(input *UpdateOwnerInput) map[string]any {
	fields := make(map[string]any)
	if input.OwnerName != nil {
		fields[colOwnerName] = *input.OwnerName
	}
	if input.OwnerNameKana != nil {
		fields[colOwnerNameKana] = normalizeNameKana(*input.OwnerNameKana)
	}
	if input.BirthDate != nil {
		// *input.BirthDate は nil なら NULL クリア、非 nil なら日付更新。
		fields[colBirthDate] = *input.BirthDate
	}
	if input.Company != nil {
		fields[colCompany] = *input.Company
	}
	if input.PostalCode != nil {
		fields[colPostalCode] = *input.PostalCode
	}
	if input.Address1 != nil {
		fields[colAddress1] = *input.Address1
	}
	if input.Address2 != nil {
		fields[colAddress2] = *input.Address2
	}
	if input.HomePostalCode != nil {
		fields[colHomePostalCode] = *input.HomePostalCode
	}
	if input.HomeAddress1 != nil {
		fields[colHomeAddress1] = *input.HomeAddress1
	}
	if input.HomeAddress2 != nil {
		fields[colHomeAddress2] = *input.HomeAddress2
	}
	if input.Phone != nil {
		fields[colPhone] = *input.Phone
	}
	if input.CompanyPhone != nil {
		fields[colCompanyPhone] = *input.CompanyPhone
	}
	if input.Email != nil {
		fields[colEmail] = *input.Email
	}
	if input.Remarks != nil {
		fields[colRemarks] = *input.Remarks
	}
	if input.IsDangerous != nil {
		fields[colIsDangerous] = *input.IsDangerous
	}
	if input.DiscountRate != nil {
		fields[colDiscountRate] = *input.DiscountRate
	}
	if input.MembershipType != nil {
		fields[colMembershipType] = *input.MembershipType
	}
	if input.DMPreference != nil {
		// *input.DMPreference は *bool: nil = NULL クリア、非nil = 値セット。
		fields[colDMPreference] = *input.DMPreference
	}
	return fields
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
