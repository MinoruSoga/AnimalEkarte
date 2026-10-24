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
// u3: 明細の出所 id（廃止列）を外し、数値を正規の10進文字列にし、NULL を値と区別する（EMR-505）。
const unbilledRevisionAlgorithm = "u3"

// unbilledRevisionItem は revision に含める候補行の請求意味集合。
// クライアントに表示される請求意味（種別・名称・金額・税・保険）を網羅し、
// 追加・削除・変更のいずれも検出する。ID は候補行の ID（出所の列は廃止され、請求元は billing_source_bindings へ移った）。
// 数値は正規の10進文字列、NULL は null で表し、未設定と 0 を区別する。
type unbilledRevisionItem struct {
	ID                    uint64  `json:"id"`
	Category              string  `json:"category"`
	Name                  string  `json:"name"`
	UnitPrice             string  `json:"unit_price"`
	Quantity              string  `json:"quantity"`
	DiscountRate          string  `json:"discount_rate"`
	DiscountAmount        int64   `json:"discount_amount"`
	TaxType               string  `json:"tax_type"`
	TaxRate               *string `json:"tax_rate"`
	IsInsuranceApplicable *bool   `json:"is_insurance_applicable"`
	Source                *string `json:"source"`
	OtherReason           string  `json:"other_reason"`
	MerchandiseItemID     uint64  `json:"merchandise_item_id"`
	MedicalRecordID       uint64  `json:"medical_record_id"`
	SortOrder             *int    `json:"sort_order"`
}

func toUnbilledRevisionItem(item *model.BillingItem) unbilledRevisionItem {
	ri := unbilledRevisionItem{
		ID:                    item.ID,
		Category:              string(item.Category),
		Name:                  item.Name,
		UnitPrice:             revisionDecimal(item.UnitPrice),
		Quantity:              revisionDecimal(item.Quantity),
		DiscountRate:          revisionDecimal(item.DiscountRate),
		DiscountAmount:        item.DiscountAmount,
		TaxType:               string(item.TaxType),
		TaxRate:               revisionOptionalDecimal(item.TaxRate),
		IsInsuranceApplicable: item.IsInsuranceApplicable,
		Source:                revisionOptionalSource(item.Source),
		SortOrder:             item.SortOrder,
	}
	if item.OtherReason != nil {
		ri.OtherReason = *item.OtherReason
	}
	if item.MerchandiseItemID != nil {
		ri.MerchandiseItemID = *item.MerchandiseItemID
	}
	if item.MedicalRecordID != nil {
		ri.MedicalRecordID = *item.MedicalRecordID
	}
	return ri
}

// revisionDecimal は fingerprint 用に ExactNumeric を正規の10進文字列にする（10.0 と 10 は同じ値）。
// 表せない値（NaN・非有限など）は固定の印にし、有効な数値とは区別する。
func revisionDecimal(value model.ExactNumeric) string {
	text, err := value.CanonicalString()
	if err != nil {
		return "invalid"
	}
	return text
}

// revisionOptionalDecimal は NULL を null のまま（未設定と 0 を区別して）fingerprint に載せる。
func revisionOptionalDecimal(value *model.ExactNumeric) *string {
	if value == nil {
		return nil
	}
	text := revisionDecimal(*value)
	return &text
}

// revisionOptionalSource は NULL の source を manual などに補わず null のまま載せる。
func revisionOptionalSource(source *model.ItemSource) *string {
	if source == nil {
		return nil
	}
	text := string(*source)
	return &text
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
