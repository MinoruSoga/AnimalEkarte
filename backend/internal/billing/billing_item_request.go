package billing

import (
	"net/url"
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

type unbilledItemsQuery struct {
	PetID string
}

func newUnbilledItemsQuery(values url.Values) unbilledItemsQuery {
	return unbilledItemsQuery{PetID: values.Get("pet_id")}
}

func (q unbilledItemsQuery) toPetID() (uint64, error) {
	if q.PetID == "" {
		return 0, apperrors.WrapInvalidInput("pet_id は必須です")
	}
	petID, err := strconv.ParseUint(q.PetID, 10, 64)
	if err != nil || petID == 0 {
		return 0, apperrors.WrapInvalidInput("pet_id の形式が不正です")
	}
	return petID, nil
}

// ---- accounting_request/response.go から B③ で分離移動（明細系のみ）----

// createBillingItemRequest は明細作成リクエスト。
type createBillingItemRequest struct {
	BillingID             uint64  `json:"billing_id" binding:"required"`
	Category              string  `json:"category"  binding:"omitempty,oneof=examination test procedure surgery medicine food goods other vaccine trimming hotel training"`
	Name                  string  `json:"name"      binding:"required"`
	UnitPrice             int64   `json:"unit_price" binding:"min=0"`
	Quantity              float64 `json:"quantity"   binding:"min=0"`
	DiscountRate          float64 `json:"discount_rate" binding:"min=0,max=100"`
	DiscountAmount        int64   `json:"discount_amount" binding:"min=0"`
	TaxType               string  `json:"tax_type"  binding:"omitempty,oneof=included excluded exempt"`
	TaxRate               float64 `json:"tax_rate"`
	IsInsuranceApplicable bool    `json:"is_insurance_applicable"`
	Source                string  `json:"source"    binding:"omitempty,oneof=medical_record manual hospitalization trimming"`
	OtherReason           *string `json:"other_reason"`
	TreatmentID           *uint64 `json:"treatment_id"`
	VaccinationID         *uint64 `json:"vaccination_id"`
	AppointmentID         *uint64 `json:"appointment_id"`
	TrimmingCourseID      *uint64 `json:"trimming_course_id"`
	TrimmingOptionID      *uint64 `json:"trimming_option_id"`
	MerchandiseItemID     *uint64 `json:"merchandise_item_id"`
	SortOrder             int     `json:"sort_order"`
	// #115 / BUG-463: 締め後編集理由（レジ締め済み期間の明細を変更する場合は必須）
	PostCloseReason *string `json:"post_close_reason"`
}

// updateBillingItemRequest は明細更新リクエスト（nil = 未指定）。
type updateBillingItemRequest struct {
	UnitPrice             *int64   `json:"unit_price"`
	Quantity              *float64 `json:"quantity"`
	DiscountRate          *float64 `json:"discount_rate" binding:"omitempty,min=0,max=100"`
	DiscountAmount        *int64   `json:"discount_amount" binding:"omitempty,min=0"`
	TaxType               *string  `json:"tax_type"  binding:"omitempty,oneof=included excluded exempt"`
	TaxRate               *float64 `json:"tax_rate"`
	IsInsuranceApplicable *bool    `json:"is_insurance_applicable"`
	// #115 / BUG-463: 締め後編集理由（レジ締め済み期間の明細を変更する場合は必須）
	PostCloseReason *string `json:"post_close_reason"`
}

// deleteBillingItemRequest は明細削除の任意 body。
// 締め後削除時のみ post_close_reason を送る（BUG-463 residual）。
type deleteBillingItemRequest struct {
	PostCloseReason *string `json:"post_close_reason"`
}

func (r *createBillingItemRequest) toServiceInput(clinicID uint64) *CreateBillingItemInput {
	return &CreateBillingItemInput{
		ClinicID:              clinicID,
		BillingID:             r.BillingID,
		Category:              r.Category,
		Name:                  r.Name,
		UnitPrice:             r.UnitPrice,
		Quantity:              r.Quantity,
		DiscountRate:          r.DiscountRate,
		DiscountAmount:        r.DiscountAmount,
		TaxType:               r.TaxType,
		TaxRate:               r.TaxRate,
		IsInsuranceApplicable: r.IsInsuranceApplicable,
		Source:                r.Source,
		OtherReason:           r.OtherReason,
		TreatmentID:           r.TreatmentID,
		VaccinationID:         r.VaccinationID,
		AppointmentID:         r.AppointmentID,
		TrimmingCourseID:      r.TrimmingCourseID,
		TrimmingOptionID:      r.TrimmingOptionID,
		MerchandiseItemID:     r.MerchandiseItemID,
		SortOrder:             r.SortOrder,
		PostCloseReason:       r.PostCloseReason,
	}
}

func (r updateBillingItemRequest) toServiceInput() (*UpdateBillingItemInput, error) {
	var taxType *model.TaxType
	if r.TaxType != nil {
		if err := sharedkernel.ValidateTaxType(*r.TaxType); err != nil {
			return nil, err
		}
		t := model.TaxType(*r.TaxType)
		taxType = &t
	}

	return &UpdateBillingItemInput{
		UnitPrice:             r.UnitPrice,
		Quantity:              r.Quantity,
		DiscountRate:          r.DiscountRate,
		DiscountAmount:        r.DiscountAmount,
		TaxType:               taxType,
		TaxRate:               r.TaxRate,
		IsInsuranceApplicable: r.IsInsuranceApplicable,
		PostCloseReason:       r.PostCloseReason,
	}, nil
}

type BillingItemResponse struct {
	ID                    uint64    `json:"id"`
	BillingID             uint64    `json:"billing_id"`
	Category              string    `json:"category"`
	Name                  string    `json:"name"`
	UnitPrice             int64     `json:"unit_price"`
	Quantity              float64   `json:"quantity"`
	DiscountRate          float64   `json:"discount_rate"`
	DiscountAmount        int64     `json:"discount_amount"`
	Subtotal              int64     `json:"subtotal"`
	TaxType               string    `json:"tax_type"`
	TaxRate               float64   `json:"tax_rate"`
	TaxAmount             int64     `json:"tax_amount"`
	IsInsuranceApplicable bool      `json:"is_insurance_applicable"`
	Source                string    `json:"source"`
	OtherReason           *string   `json:"other_reason,omitempty"`
	TreatmentID           *uint64   `json:"treatment_id,omitempty"`
	// MedicalRecordID は未請求候補など、treatment 由来の親カルテ（DB 列ではない仮想値）。
	MedicalRecordID       *uint64   `json:"medical_record_id,omitempty"`
	VaccinationID         *uint64   `json:"vaccination_id,omitempty"`
	AppointmentID         *uint64   `json:"appointment_id,omitempty"`
	TrimmingCourseID      *uint64   `json:"trimming_course_id,omitempty"`
	TrimmingOptionID      *uint64   `json:"trimming_option_id,omitempty"`
	MerchandiseItemID     *uint64   `json:"merchandise_item_id,omitempty"`
	SortOrder             int       `json:"sort_order"`
	CreatedAt             time.Time `json:"created_at"`
}

func ToBillingItemResponse(item *model.BillingItem) BillingItemResponse {
	// #85: subtotal は割引後（単価×数量 − 割引額）。負値は 0 クランプ。
	subtotal := max(int64(float64(item.UnitPrice)*item.Quantity)-item.DiscountAmount, 0)
	taxAmount := item.CalculateTaxAmount()
	return BillingItemResponse{
		ID:                    item.ID,
		BillingID:             item.BillingID,
		Category:              string(item.Category),
		Name:                  item.Name,
		UnitPrice:             item.UnitPrice,
		Quantity:              item.Quantity,
		DiscountRate:          item.DiscountRate,
		DiscountAmount:        item.DiscountAmount,
		Subtotal:              subtotal,
		TaxType:               string(item.TaxType),
		TaxRate:               item.TaxRate,
		TaxAmount:             taxAmount,
		IsInsuranceApplicable: item.IsInsuranceApplicable,
		Source:                string(item.Source),
		OtherReason:           item.OtherReason,
		TreatmentID:           item.TreatmentID,
		MedicalRecordID:       item.MedicalRecordID,
		VaccinationID:         item.VaccinationID,
		AppointmentID:         item.AppointmentID,
		TrimmingCourseID:      item.TrimmingCourseID,
		TrimmingOptionID:      item.TrimmingOptionID,
		MerchandiseItemID:     item.MerchandiseItemID,
		SortOrder:             item.SortOrder,
		CreatedAt:             httpapi.LocalTime(item.CreatedAt),
	}
}
