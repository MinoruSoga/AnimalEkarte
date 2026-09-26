package billing

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// AccountingCodeUnbilledConflict は EMR-196② の stale unbilled 409 で返す安定エラーコード。
// complete 確定時にクライアントが表示した未請求集約と現在の集約が一致しない場合に対応する。
const AccountingCodeUnbilledConflict = "UNBILLED_ITEMS_CHANGED"

// unbilledRevisionAlgorithm は集約フィンガープリントの方式版。
// 正規化対象フィールドやエンコードを変える場合は接頭辞を上げる。
const unbilledRevisionAlgorithm = "u1"

// unbilledRevisionItem は revision に含める候補行の請求意味集合。
// クライアントに表示される請求意味（種別・名称・金額・税・保険・出所リンク）を網羅し、
// 追加・削除・変更のいずれも検出する。ID は treatment 由来なら treatment ID、
// billing_items 由来なら行 ID（FE の provenance key と同じ識別粒度）。
type unbilledRevisionItem struct {
	ID                    uint64  `json:"id"`
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
	MedicalRecordID       uint64  `json:"medical_record_id"`
	VaccinationID         uint64  `json:"vaccination_id"`
	ExamID                uint64  `json:"exam_id"`
	AppointmentID         uint64  `json:"appointment_id"`
	TrimmingCourseID      uint64  `json:"trimming_course_id"`
	TrimmingOptionID      uint64  `json:"trimming_option_id"`
	SortOrder             int     `json:"sort_order"`
}

func toUnbilledRevisionItem(item *model.BillingItem) unbilledRevisionItem {
	ri := unbilledRevisionItem{
		ID:                    item.ID,
		Category:              string(item.Category),
		Name:                  item.Name,
		UnitPrice:             item.UnitPrice,
		Quantity:              item.Quantity,
		DiscountRate:          item.DiscountRate,
		DiscountAmount:        item.DiscountAmount,
		TaxType:               string(item.TaxType),
		TaxRate:               item.TaxRate,
		IsInsuranceApplicable: item.IsInsuranceApplicable,
		Source:                string(item.Source),
		SortOrder:             item.SortOrder,
	}
	if item.OtherReason != nil {
		ri.OtherReason = *item.OtherReason
	}
	if item.MerchandiseItemID != nil {
		ri.MerchandiseItemID = *item.MerchandiseItemID
	}
	if item.TreatmentID != nil {
		ri.TreatmentID = *item.TreatmentID
	}
	if item.MedicalRecordID != nil {
		ri.MedicalRecordID = *item.MedicalRecordID
	}
	if item.VaccinationID != nil {
		ri.VaccinationID = *item.VaccinationID
	}
	if item.ExamID != nil {
		ri.ExamID = *item.ExamID
	}
	if item.AppointmentID != nil {
		ri.AppointmentID = *item.AppointmentID
	}
	if item.TrimmingCourseID != nil {
		ri.TrimmingCourseID = *item.TrimmingCourseID
	}
	if item.TrimmingOptionID != nil {
		ri.TrimmingOptionID = *item.TrimmingOptionID
	}
	return ri
}

// computeUnbilledRevision は unbilled 集約の決定的フィンガープリントを返す（EMR-196②）。
// クライアントは unbilled-details 応答の revision を complete の expected_unbilled_revision
// として返送し、service が complete tx 内の再集計と照合する optimistic concurrency token。
// 集約順序に依存しないよう各行を正規化 JSON 化して sort してから SHA-256 に掛ける。
func computeUnbilledRevision(items []model.BillingItem, warnings []UnbilledWarning) string {
	itemParts := make([]string, 0, len(items))
	for i := range items {
		// 固定 shape の struct marshal は失敗しない。
		raw, _ := json.Marshal(toUnbilledRevisionItem(&items[i]))
		itemParts = append(itemParts, string(raw))
	}
	sort.Strings(itemParts)

	warningParts := make([]string, 0, len(warnings))
	for i := range warnings {
		raw, _ := json.Marshal(warnings[i])
		warningParts = append(warningParts, string(raw))
	}
	sort.Strings(warningParts)

	h := sha256.New()
	for _, part := range itemParts {
		h.Write([]byte(part))
		h.Write([]byte{0x1f})
	}
	// items/warnings の区切り（部分集合が偶発一致しないよう domain separation）。
	h.Write([]byte{0x1e})
	for _, part := range warningParts {
		h.Write([]byte(part))
		h.Write([]byte{0x1f})
	}
	return unbilledRevisionAlgorithm + ":" + hex.EncodeToString(h.Sum(nil))
}

// unbilledRevisionConflictError は表示済み集約と現在集約の版不一致を表す 409（EMR-196②）。
// CurrentRevision は handler が response body の unbilled_revision として返す。
type unbilledRevisionConflictError struct {
	appErr          *apperrors.AppError
	CurrentRevision string
}

func (e *unbilledRevisionConflictError) Error() string { return e.appErr.Error() }
func (e *unbilledRevisionConflictError) Unwrap() error { return e.appErr }

func newUnbilledRevisionConflictError(currentRevision string) *unbilledRevisionConflictError {
	return &unbilledRevisionConflictError{
		appErr: &apperrors.AppError{
			Code:    AccountingCodeUnbilledConflict,
			Message: "未請求明細が更新されました。最新の内容を確認してから再度確定してください",
			Err:     apperrors.ErrConflict,
		},
		CurrentRevision: currentRevision,
	}
}
