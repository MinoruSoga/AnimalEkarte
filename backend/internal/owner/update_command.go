package owner

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// UpdateCommand is the typed owner write used by UpdateAndFind / OwnerUpdateApplier.
// Nil pointer profile fields are omitted. Set* groups always write their columns
// (including SQL NULL). Column map assembly stays unexported — no public
// arbitrary column-name factory.
// EMR-491: 氏名・住所・連絡先などの基本情報は共通 basic revision の項目で owners の列ではないため、ここには持たない。
type UpdateCommand struct {
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
		Remarks:        input.Remarks,
		IsDangerous:    input.IsDangerous,
		DiscountRate:   input.DiscountRate,
		MembershipType: input.MembershipType,
		DMPreference:   input.DMPreference,
	}
	return cmd
}

func updateCommandFields(cmd UpdateCommand) map[string]any {
	fields := make(map[string]any)
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
