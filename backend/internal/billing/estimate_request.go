package billing

import (
	"net/url"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type listEstimateQuery struct {
	OwnerID         string
	MedicalRecordID string
	Status          string
}

func newListEstimateQuery(values url.Values) listEstimateQuery {
	return listEstimateQuery{
		OwnerID:         values.Get("owner_id"),
		MedicalRecordID: values.Get("medical_record_id"),
		Status:          values.Get("status"),
	}
}

type listEstimateFilters struct {
	OwnerID         *uint64
	MedicalRecordID *uint64
	Status          *string
}

func (q listEstimateQuery) toServiceFilters() (listEstimateFilters, error) {
	ownerID, err := parseOptionalUintQueryFilter(q.OwnerID, "owner_id")
	if err != nil {
		return listEstimateFilters{}, err
	}
	medicalRecordID, err := parseOptionalUintQueryFilter(q.MedicalRecordID, "medical_record_id")
	if err != nil {
		return listEstimateFilters{}, err
	}
	return listEstimateFilters{
		OwnerID:         ownerID,
		MedicalRecordID: medicalRecordID,
		Status:          optionalStringQueryFilter(q.Status),
	}, nil
}

// createEstimateRequest は見積書作成リクエスト
type createEstimateRequest struct {
	MedicalRecordID *uint64                     `json:"medical_record_id"`
	Title           string                      `json:"title" binding:"required,min=1,max=255"`
	OwnerID         *uint64                     `json:"owner_id"`
	PetID           *uint64                     `json:"pet_id"`
	Status          string                      `json:"status"  binding:"omitempty,oneof=draft sent"`
	Subtotal        int64                       `json:"subtotal"      binding:"min=0"`
	TaxTotal        int64                       `json:"tax_total"     binding:"min=0"`
	TotalAmount     int64                       `json:"total_amount"  binding:"min=0"`
	InsuranceAmount int64                       `json:"insurance_amount"`
	DiscountAmount  int64                       `json:"discount_amount"`
	ValidUntil      *time.Time                  `json:"valid_until"`
	Comment         string                      `json:"comment"`
	Notes           string                      `json:"notes"`
	Items           []createEstimateItemRequest `json:"items" binding:"omitempty,dive"`
}

type createEstimateItemRequest struct {
	Name                  string  `json:"name" binding:"required,min=1,max=255"`
	Category              string  `json:"category"`
	UnitPrice             int64   `json:"unit_price" binding:"min=0"`
	Quantity              float64 `json:"quantity"`
	DiscountRate          float64 `json:"discount_rate"`
	DiscountAmount        int64   `json:"discount_amount"`
	IsInsuranceApplicable bool    `json:"is_insurance_applicable"`
	SortOrder             int     `json:"sort_order"`
}

func estimateItemInputsFromRequest(items []createEstimateItemRequest) []EstimateItemInput {
	out := make([]EstimateItemInput, 0, len(items))
	for _, item := range items {
		out = append(out, EstimateItemInput{
			Name:                  item.Name,
			Category:              model.ItemCategory(item.Category),
			UnitPrice:             item.UnitPrice,
			Quantity:              item.Quantity,
			DiscountRate:          item.DiscountRate,
			DiscountAmount:        item.DiscountAmount,
			IsInsuranceApplicable: item.IsInsuranceApplicable,
			SortOrder:             item.SortOrder,
		})
	}
	return out
}

// toServiceInput は認証済み staffID を created_by に設定する（body の created_by は受け取らない・AUD-005）。
func (r *createEstimateRequest) toServiceInput(staffID uint64) *CreateEstimateInput {
	input := &CreateEstimateInput{
		MedicalRecordID: r.MedicalRecordID,
		Title:           r.Title,
		OwnerID:         r.OwnerID,
		PetID:           r.PetID,
		Subtotal:        r.Subtotal,
		TaxTotal:        r.TaxTotal,
		TotalAmount:     r.TotalAmount,
		InsuranceAmount: r.InsuranceAmount,
		DiscountAmount:  r.DiscountAmount,
		ValidUntil:      r.ValidUntil,
		Comment:         r.Comment,
		Notes:           r.Notes,
		CreatedBy:       &staffID,
		Items:           estimateItemInputsFromRequest(r.Items),
	}
	if r.Status != "" {
		input.Status = model.EstimateStatus(r.Status)
	}
	return input
}

// updateEstimateRequest は見積書更新リクエスト（PATCH: nil = 未送信）
type updateEstimateRequest struct {
	Title           *string                      `json:"title"`
	Status          *string                      `json:"status"  binding:"omitempty,oneof=draft sent approved rejected"`
	Subtotal        *int64                       `json:"subtotal"`
	TaxTotal        *int64                       `json:"tax_total"`
	TotalAmount     *int64                       `json:"total_amount"`
	InsuranceAmount *int64                       `json:"insurance_amount"`
	DiscountAmount  *int64                       `json:"discount_amount"`
	ValidUntil      *time.Time                   `json:"valid_until"`
	ClearValidUntil bool                         `json:"clear_valid_until"`
	Comment         *string                      `json:"comment"`
	Notes           *string                      `json:"notes"`
	Items           *[]createEstimateItemRequest `json:"items" binding:"omitempty,dive"`
}

func (r *updateEstimateRequest) toServiceInput(actorID uint64) *UpdateEstimateInput {
	input := &UpdateEstimateInput{
		Title:           r.Title,
		Subtotal:        r.Subtotal,
		TaxTotal:        r.TaxTotal,
		TotalAmount:     r.TotalAmount,
		InsuranceAmount: r.InsuranceAmount,
		DiscountAmount:  r.DiscountAmount,
		ValidUntil:      r.ValidUntil,
		ClearValidUntil: r.ClearValidUntil,
		Comment:         r.Comment,
		Notes:           r.Notes,
		ActorID:         &actorID,
	}
	if r.Items != nil {
		items := estimateItemInputsFromRequest(*r.Items)
		input.Items = &items
	}
	if r.Status != nil {
		status := model.EstimateStatus(*r.Status)
		input.Status = &status
	}
	return input
}

// createEstimateSuccessorRequest は確定見積の後継ドラフト作成リクエスト（TASK-012 FINAL B）。
// reason 必須。title/comment/notes は任意上書き。
type createEstimateSuccessorRequest struct {
	Reason  string  `json:"reason" binding:"required,min=1,max=500"`
	Title   *string `json:"title" binding:"omitempty,min=1,max=255"`
	Comment *string `json:"comment"`
	Notes   *string `json:"notes"`
}

func (r *createEstimateSuccessorRequest) toServiceInput(actorID uint64) *CreateSuccessorInput {
	return &CreateSuccessorInput{
		Reason:  r.Reason,
		Title:   r.Title,
		Comment: r.Comment,
		Notes:   r.Notes,
		ActorID: actorID,
	}
}
