package billing

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

const (
	billingConfirmationJSONBodyMaxBytes            int64 = 8 * 1024
	billingConfirmationReturnReasonMaxLength             = 500
	billingConfirmationMemoMaxLength                     = 1000
	billingConfirmationBodyTooLargeMessage               = "request body exceeds size limit"
	billingConfirmationInvalidBodyMessage                = "invalid billing confirmation request body"
	billingConfirmationUnsupportedMediaTypeMessage       = "Content-Type must be application/json"
)

// confirmBillingConfirmationRequest は会計医師確認のバインド struct
type confirmBillingConfirmationRequest struct {
	Memo string `json:"memo"`
}

func (r confirmBillingConfirmationRequest) normalizeAndValidate() (confirmBillingConfirmationRequest, error) {
	normalized := confirmBillingConfirmationRequest{
		Memo: strings.TrimSpace(r.Memo),
	}
	if utf8.RuneCountInString(normalized.Memo) > billingConfirmationMemoMaxLength {
		return confirmBillingConfirmationRequest{}, apperrors.WrapInvalidInput(
			fmt.Sprintf("memo must be %d characters or fewer", billingConfirmationMemoMaxLength),
		)
	}
	return normalized, nil
}

func (r confirmBillingConfirmationRequest) toServiceInput(authenticatedStaffID uint64) *ConfirmBillingConfirmationInput {
	return &ConfirmBillingConfirmationInput{
		ConfirmedBy: authenticatedStaffID,
		Memo:        r.Memo,
	}
}

// returnBillingConfirmationRequest は会計差し戻しのバインド struct
type returnBillingConfirmationRequest struct {
	ReturnReason string `json:"return_reason"`
	Memo         string `json:"memo"`
}

func (r returnBillingConfirmationRequest) normalizeAndValidate() (returnBillingConfirmationRequest, error) {
	normalized := returnBillingConfirmationRequest{
		ReturnReason: strings.TrimSpace(r.ReturnReason),
		Memo:         strings.TrimSpace(r.Memo),
	}
	if normalized.ReturnReason == "" {
		return returnBillingConfirmationRequest{}, apperrors.WrapInvalidInput("return_reason is required")
	}
	if utf8.RuneCountInString(normalized.ReturnReason) > billingConfirmationReturnReasonMaxLength {
		return returnBillingConfirmationRequest{}, apperrors.WrapInvalidInput(
			fmt.Sprintf(
				"return_reason must be %d characters or fewer",
				billingConfirmationReturnReasonMaxLength,
			),
		)
	}
	if utf8.RuneCountInString(normalized.Memo) > billingConfirmationMemoMaxLength {
		return returnBillingConfirmationRequest{}, apperrors.WrapInvalidInput(
			fmt.Sprintf("memo must be %d characters or fewer", billingConfirmationMemoMaxLength),
		)
	}
	return normalized, nil
}

func (r returnBillingConfirmationRequest) toServiceInput(authenticatedStaffID uint64) *ReturnBillingConfirmationInput {
	return &ReturnBillingConfirmationInput{
		ReturnedBy:   authenticatedStaffID,
		ReturnReason: r.ReturnReason,
		Memo:         r.Memo,
	}
}
