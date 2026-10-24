package billing

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type estimateItemResponse struct {
	ID                    uint64              `json:"id"`
	EstimateID            uint64              `json:"estimate_id"`
	Name                  string              `json:"name"`
	Category              string              `json:"category"`
	UnitPrice             model.ExactNumeric  `json:"unit_price"`
	Quantity              model.ExactNumeric  `json:"quantity"`
	Subtotal              *int64              `json:"subtotal"`
	TaxType               string              `json:"tax_type"`
	TaxRate               *model.ExactNumeric `json:"tax_rate"`
	TaxAmount             *int64              `json:"tax_amount"`
	DiscountRate          *model.ExactNumeric `json:"discount_rate"`
	DiscountAmount        *int64              `json:"discount_amount"`
	IsInsuranceApplicable *bool               `json:"is_insurance_applicable"`
	SortOrder             int                 `json:"sort_order"`
	CreatedAt             time.Time           `json:"created_at"`
}

type estimateResponse struct {
	ID                   uint64                 `json:"id"`
	ClinicID             uint64                 `json:"clinic_id"`
	EstimateNo           string                 `json:"estimate_no"`
	MedicalRecordID      *uint64                `json:"medical_record_id,omitempty"`
	Title                string                 `json:"title"`
	OwnerID              *uint64                `json:"owner_id,omitempty"`
	Owner                *ownerSummaryResponse  `json:"owner,omitempty"`
	PetID                *uint64                `json:"pet_id,omitempty"`
	Pet                  *petSummaryResponse    `json:"pet,omitempty"`
	Status               *string                `json:"status"`
	Subtotal             *int64                 `json:"subtotal"`
	TaxTotal             *int64                 `json:"tax_total"`
	TotalAmount          *int64                 `json:"total_amount"`
	InsuranceAmount      int64                  `json:"insurance_amount"`
	DiscountAmount       int64                  `json:"discount_amount"`
	ValidUntil           *time.Time             `json:"valid_until,omitempty"`
	Comment              string                 `json:"comment"`
	Notes                string                 `json:"notes"`
	CreatedBy            *uint64                `json:"created_by,omitempty"`
	SupersedesEstimateID *uint64                `json:"supersedes_estimate_id,omitempty"`
	Items                []estimateItemResponse `json:"items,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

func toEstimateItemResponse(item *model.EstimateItem) estimateItemResponse {
	// subtotal は明細値引き後の税抜額（単価 × 数量の 1 円未満四捨五入 − 明細の値引き）、tax_amount は表示用の
	// 行別税額で請求書の税額の正本ではない。算定できない行は null にし、0 で埋めない（EMR-508 / Q12）。
	var subtotal, taxAmount *int64
	if line, lineErr := estimateItemCalcLine(item, 0); lineErr == nil {
		if amounts, calcErr := CalculateLineDisplayAmounts(line); calcErr == nil {
			subtotal = &amounts.Subtotal
			taxAmount = amounts.TaxAmount
		}
	}
	// sort_order は表示順であり業務値ではないため、NULL は 0 として返す（Q12。変換はこの mapper に限る）。
	sortOrder := 0
	if item.SortOrder != nil {
		sortOrder = *item.SortOrder
	}
	return estimateItemResponse{
		ID:                    item.ID,
		EstimateID:            item.EstimateID,
		Name:                  item.Name,
		Category:              string(item.Category),
		UnitPrice:             item.UnitPrice,
		Quantity:              item.Quantity,
		Subtotal:              subtotal,
		TaxType:               string(item.TaxType),
		TaxRate:               item.TaxRate,
		TaxAmount:             taxAmount,
		DiscountRate:          item.DiscountRate,
		DiscountAmount:        item.DiscountAmount,
		IsInsuranceApplicable: item.IsInsuranceApplicable,
		SortOrder:             sortOrder,
		CreatedAt:             httpapi.LocalTime(item.CreatedAt),
	}
}

func toEstimateResponse(e *model.Estimate) estimateResponse {
	items := make([]estimateItemResponse, 0, len(e.Items))
	for i := range e.Items {
		items = append(items, toEstimateItemResponse(&e.Items[i]))
	}
	// 合計は保存せず明細から求める。算定できないときは 3 つとも null にし、0 で埋めない（EMR-508）。
	var subtotal, taxTotal, totalAmount *int64
	if totals, err := calculateEstimateTotals(e.Items); err == nil {
		subtotal, taxTotal, totalAmount = &totals.Subtotal, &totals.TaxTotal, &totals.TotalAmount
	}
	// status が NULL（不明）なら null で返し、draft などに補わない（Q12）。
	var status *string
	if e.Status != nil {
		value := string(*e.Status)
		status = &value
	}
	// insurance_amount / discount_amount は保存先がなく、受け付けるのは 0 のみ（Q6）なので応答は常に 0。
	// supersedes_estimate_id は estimate_successor_links の reader が無い（一覧で 1 件ずつ引く必要がある）ため返さない。
	return estimateResponse{
		ID:                   e.ID,
		ClinicID:             e.ClinicID,
		EstimateNo:           e.EstimateNo,
		MedicalRecordID:      e.MedicalRecordID,
		Title:                e.Title,
		OwnerID:              e.OwnerID,
		Owner:                toOwnerSummary(e.Owner),
		PetID:                e.PetID,
		Pet:                  toPetSummary(e.Pet),
		Status:               status,
		Subtotal:             subtotal,
		TaxTotal:             taxTotal,
		TotalAmount:          totalAmount,
		InsuranceAmount:      0,
		DiscountAmount:       0,
		ValidUntil:           e.ValidUntil,
		Comment:              e.Comment,
		Notes:                e.Notes,
		CreatedBy:            e.CreatedBy,
		SupersedesEstimateID: nil,
		Items:                items,
		CreatedAt:            httpapi.LocalTime(e.CreatedAt),
		UpdatedAt:            httpapi.LocalTime(e.UpdatedAt),
	}
}
