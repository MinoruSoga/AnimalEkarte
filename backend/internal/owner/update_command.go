package owner

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// UpdateCommand is the typed owner write used by UpdateAndFind / OwnerUpdateApplier.
// Nil pointer profile fields are omitted. Set* groups always write their columns
// (including SQL NULL). Column map assembly stays unexported — no public
// arbitrary column-name factory.
type UpdateCommand struct {
	Name           *string
	NameKana       *string
	BirthDate      **time.Time
	Company        *string
	PostalCode     *string
	Address1       *string
	Address2       *string
	HomePostalCode *string
	HomeAddress1   *string
	HomeAddress2   *string
	Phone          *string
	CompanyPhone   *string
	Email          *string
	Remarks        *string
	IsDangerous    *bool
	DiscountRate   *float64
	MembershipType *model.MembershipType
	DMPreference   **bool

	SetDeliveryExclusion   bool
	DeliveryExcluded       bool
	DeliveryExcludedReason *string
	LstepOptOut            bool
	LstepOptOutAt          *time.Time
	LstepOptOutReason      *string

	SetDeliveryCaution    bool
	DeliveryCaution       bool
	DeliveryCautionReason *string

	SetTransfer   bool
	IsTransferred bool
	TransferAt    *time.Time

	SetLineIDConfirmation bool
	LineIDConfirmedAt     time.Time
	LineIDConfirmedBy     *uint64
}

func updateCommandFromOwnerInput(input *UpdateOwnerInput) UpdateCommand {
	if input == nil {
		return UpdateCommand{}
	}
	cmd := UpdateCommand{
		Name:           input.OwnerName,
		Company:        input.Company,
		PostalCode:     input.PostalCode,
		Address1:       input.Address1,
		Address2:       input.Address2,
		HomePostalCode: input.HomePostalCode,
		HomeAddress1:   input.HomeAddress1,
		HomeAddress2:   input.HomeAddress2,
		Phone:          input.Phone,
		CompanyPhone:   input.CompanyPhone,
		Email:          input.Email,
		Remarks:        input.Remarks,
		IsDangerous:    input.IsDangerous,
		DiscountRate:   input.DiscountRate,
		MembershipType: input.MembershipType,
		BirthDate:      input.BirthDate,
		DMPreference:   input.DMPreference,
	}
	if input.OwnerNameKana != nil {
		normalized := normalizeNameKana(*input.OwnerNameKana)
		cmd.NameKana = &normalized
	}
	return cmd
}

func updateCommandFields(cmd UpdateCommand) map[string]any {
	fields := make(map[string]any)
	if cmd.Name != nil {
		fields[colOwnerName] = *cmd.Name
	}
	if cmd.NameKana != nil {
		fields[colOwnerNameKana] = *cmd.NameKana
	}
	if cmd.BirthDate != nil {
		fields[colBirthDate] = *cmd.BirthDate
	}
	if cmd.Company != nil {
		fields[colCompany] = *cmd.Company
	}
	if cmd.PostalCode != nil {
		fields[colPostalCode] = *cmd.PostalCode
	}
	if cmd.Address1 != nil {
		fields[colAddress1] = *cmd.Address1
	}
	if cmd.Address2 != nil {
		fields[colAddress2] = *cmd.Address2
	}
	if cmd.HomePostalCode != nil {
		fields[colHomePostalCode] = *cmd.HomePostalCode
	}
	if cmd.HomeAddress1 != nil {
		fields[colHomeAddress1] = *cmd.HomeAddress1
	}
	if cmd.HomeAddress2 != nil {
		fields[colHomeAddress2] = *cmd.HomeAddress2
	}
	if cmd.Phone != nil {
		fields[colPhone] = *cmd.Phone
	}
	if cmd.CompanyPhone != nil {
		fields[colCompanyPhone] = *cmd.CompanyPhone
	}
	if cmd.Email != nil {
		fields[colEmail] = *cmd.Email
	}
	if cmd.Remarks != nil {
		fields[colRemarks] = *cmd.Remarks
	}
	if cmd.IsDangerous != nil {
		fields[colIsDangerous] = *cmd.IsDangerous
	}
	if cmd.DiscountRate != nil {
		fields[colDiscountRate] = *cmd.DiscountRate
	}
	if cmd.MembershipType != nil {
		fields[colMembershipType] = *cmd.MembershipType
	}
	if cmd.DMPreference != nil {
		fields[colDMPreference] = *cmd.DMPreference
	}

	if cmd.SetDeliveryExclusion {
		fields[colDeliveryExcluded] = cmd.DeliveryExcluded
		fields[colDeliveryExcludedReason] = cmd.DeliveryExcludedReason
		fields[colLstepOptOut] = cmd.LstepOptOut
		if cmd.LstepOptOutAt != nil {
			fields[colLstepOptOutAt] = *cmd.LstepOptOutAt
		} else {
			fields[colLstepOptOutAt] = nil
		}
		fields[colLstepOptOutReason] = cmd.LstepOptOutReason
	}
	if cmd.SetDeliveryCaution {
		fields[colDeliveryCaution] = cmd.DeliveryCaution
		fields[colDeliveryCautionReason] = cmd.DeliveryCautionReason
	}
	if cmd.SetTransfer {
		fields[colIsTransferred] = cmd.IsTransferred
		if cmd.TransferAt != nil {
			fields[colTransferAt] = *cmd.TransferAt
		} else {
			fields[colTransferAt] = nil
		}
	}
	if cmd.SetLineIDConfirmation {
		fields[colLineIDConfirmedAt] = cmd.LineIDConfirmedAt
		fields[colLineIDConfirmedBy] = cmd.LineIDConfirmedBy
	}
	return fields
}
