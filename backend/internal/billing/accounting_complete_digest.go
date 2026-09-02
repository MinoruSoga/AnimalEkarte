package billing

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

type completeDigestItem struct {
	Category              string  `json:"category"`
	Name                  string  `json:"name"`
	UnitPrice             int64   `json:"unit_price"`
	Quantity              float64 `json:"quantity"`
	DiscountRate          float64 `json:"discount_rate"`
	DiscountAmount        int64   `json:"discount_amount"`
	TaxType               string  `json:"tax_type"`
	TaxRate               float64 `json:"tax_rate"`
	IsInsuranceApplicable bool    `json:"is_insurance_applicable"`
	Source                string  `json:"source"`
	OtherReason           string  `json:"other_reason"`
	MerchandiseItemID     uint64  `json:"merchandise_item_id"`
	TreatmentID           uint64  `json:"treatment_id"`
	VaccinationID         uint64  `json:"vaccination_id"`
	ExamID                uint64  `json:"exam_id"`
	AppointmentID         uint64  `json:"appointment_id"`
	TrimmingCourseID      uint64  `json:"trimming_course_id"`
	TrimmingOptionID      uint64  `json:"trimming_option_id"`
	SortOrder             int     `json:"sort_order"`
}

type completeDigestSplit struct {
	Method         string `json:"method"`
	Amount         int64  `json:"amount"`
	ReceivedAmount int64  `json:"received_amount"`
	ChangeAmount   int64  `json:"change_amount"`
	ChangeOverride bool   `json:"change_override"`
}

type completeDigestRoot struct {
	MedicalRecordID   uint64                `json:"medical_record_id"`
	HospitalizationID uint64                `json:"hospitalization_id"`
	OwnerID           uint64                `json:"owner_id"`
	PetID             uint64                `json:"pet_id"`
	ScheduledDate     string                `json:"scheduled_date"`
	Memo              string                `json:"memo"`
	HasInsurance      bool                  `json:"has_insurance"`
	InsuranceRatio    float64               `json:"insurance_ratio"`
	InsuranceName     string                `json:"insurance_name"`
	InsuranceAmount   int64                 `json:"insurance_amount"`
	DiscountAmount    int64                 `json:"discount_amount"`
	PostCloseReason   string                `json:"post_close_reason"`
	Items             []completeDigestItem  `json:"items"`
	PaymentSplits     []completeDigestSplit `json:"payment_splits"`
}

// ComputeCompleteAccountingDigest は complete request の正規化 digest（hex）を返す。
// 同一 payload は常に同一 digest になる決定的方式（JSON marshal of normalized struct）。
func ComputeCompleteAccountingDigest(input *CompleteAccountingInput) (string, error) {
	if input == nil {
		return "", apperrors.WrapInvalidInput("complete input is required")
	}
	raw, err := json.Marshal(buildCompleteDigestRoot(input))
	if err != nil {
		return "", apperrors.Wrap(err, "failed to marshal complete digest payload")
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func buildCompleteDigestRoot(input *CompleteAccountingInput) completeDigestRoot {
	root := completeDigestRoot{
		Memo:          strings.TrimSpace(input.Memo),
		HasInsurance:  input.HasInsurance,
		ScheduledDate: input.ScheduledDate.UTC().Format(time.RFC3339),
		Items:         make([]completeDigestItem, 0, len(input.Items)),
		PaymentSplits: make([]completeDigestSplit, 0, len(input.PaymentSplits)),
	}
	if input.MedicalRecordID != nil {
		root.MedicalRecordID = *input.MedicalRecordID
	}
	if input.HospitalizationID != nil {
		root.HospitalizationID = *input.HospitalizationID
	}
	if input.OwnerID != nil {
		root.OwnerID = *input.OwnerID
	}
	if input.PetID != nil {
		root.PetID = *input.PetID
	}
	if input.InsuranceRatio != nil {
		root.InsuranceRatio = *input.InsuranceRatio
	}
	if input.InsuranceName != nil {
		root.InsuranceName = *input.InsuranceName
	}
	if input.InsuranceAmount != nil {
		root.InsuranceAmount = *input.InsuranceAmount
	}
	if input.DiscountAmount != nil {
		root.DiscountAmount = *input.DiscountAmount
	}
	if input.PostCloseReason != nil {
		root.PostCloseReason = strings.TrimSpace(*input.PostCloseReason)
	}
	for _, it := range input.Items {
		di := completeDigestItem{
			Category:              it.Category,
			Name:                  it.Name,
			UnitPrice:             it.UnitPrice,
			Quantity:              it.Quantity,
			DiscountRate:          it.DiscountRate,
			DiscountAmount:        it.DiscountAmount,
			TaxType:               it.TaxType,
			TaxRate:               it.TaxRate,
			IsInsuranceApplicable: it.IsInsuranceApplicable,
			Source:                it.Source,
			SortOrder:             it.SortOrder,
		}
		if it.OtherReason != nil {
			di.OtherReason = *it.OtherReason
		}
		if it.MerchandiseItemID != nil {
			di.MerchandiseItemID = *it.MerchandiseItemID
		}
		if it.TreatmentID != nil {
			di.TreatmentID = *it.TreatmentID
		}
		if it.VaccinationID != nil {
			di.VaccinationID = *it.VaccinationID
		}
		if it.ExamID != nil {
			di.ExamID = *it.ExamID
		}
		if it.AppointmentID != nil {
			di.AppointmentID = *it.AppointmentID
		}
		if it.TrimmingCourseID != nil {
			di.TrimmingCourseID = *it.TrimmingCourseID
		}
		if it.TrimmingOptionID != nil {
			di.TrimmingOptionID = *it.TrimmingOptionID
		}
		root.Items = append(root.Items, di)
	}
	for _, sp := range input.PaymentSplits {
		root.PaymentSplits = append(root.PaymentSplits, completeDigestSplit{
			Method:         string(sp.Method),
			Amount:         sp.Amount,
			ReceivedAmount: sp.ReceivedAmount,
			ChangeAmount:   sp.ChangeAmount,
			ChangeOverride: sp.ChangeOverride,
		})
	}
	return root
}
