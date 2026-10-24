package billing

// estimate_audit_diff.go — BE9-2D ⑦: audit_diff.go の medicalrecord 移動に伴い、estimate（billing
// 残留 domain）系 diff のみ本 package へ分離（単一実装・複製なし）。移設 = billing domain 移行時。

import (
	"reflect"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// Callers: estimate_service Create/Update/Delete audit best-effort. Schema: none (in-memory map for audit_logs JSON).
func extractEstimateImportantFields(e *model.Estimate) map[string]any {
	if e == nil {
		return nil
	}
	return map[string]any{
		"status":            estimateStatusAuditValue(e.Status),
		"title":             e.Title,
		"valid_until":       e.ValidUntil,
		"comment":           e.Comment,
		"notes":             e.Notes,
		"owner_id":          e.OwnerID,
		"pet_id":            e.PetID,
		"medical_record_id": e.MedicalRecordID,
		"items":             estimateItemsAuditValue(e.Items),
	}
}

// 変更なしの場合は nil/nil を返す（audit をスキップするシグナル）。
func diffEstimateImportantFields(oldEst, newEst *model.Estimate) (oldDiff, newDiff map[string]any) {
	type comparison struct {
		key    string
		oldVal any
		newVal any
		equal  bool
	}

	oldStatus, newStatus := estimateStatusAuditValue(oldEst.Status), estimateStatusAuditValue(newEst.Status)
	oldItems, newItems := estimateItemsAuditValue(oldEst.Items), estimateItemsAuditValue(newEst.Items)
	comparisons := []comparison{
		{key: "status", oldVal: oldStatus, newVal: newStatus, equal: oldStatus == newStatus},
		{key: "title", oldVal: oldEst.Title, newVal: newEst.Title, equal: oldEst.Title == newEst.Title},
		{key: "valid_until", oldVal: oldEst.ValidUntil, newVal: newEst.ValidUntil, equal: ptrTimeEqual(oldEst.ValidUntil, newEst.ValidUntil)},
		{key: "comment", oldVal: oldEst.Comment, newVal: newEst.Comment, equal: oldEst.Comment == newEst.Comment},
		{key: "notes", oldVal: oldEst.Notes, newVal: newEst.Notes, equal: oldEst.Notes == newEst.Notes},
		{key: "owner_id", oldVal: oldEst.OwnerID, newVal: newEst.OwnerID, equal: ptrUint64Equal(oldEst.OwnerID, newEst.OwnerID)},
		{key: "pet_id", oldVal: oldEst.PetID, newVal: newEst.PetID, equal: ptrUint64Equal(oldEst.PetID, newEst.PetID)},
		{key: "medical_record_id", oldVal: oldEst.MedicalRecordID, newVal: newEst.MedicalRecordID, equal: ptrUint64Equal(oldEst.MedicalRecordID, newEst.MedicalRecordID)},
		{key: "items", oldVal: oldItems, newVal: newItems, equal: reflect.DeepEqual(oldItems, newItems)},
	}

	hasChange := false
	for _, c := range comparisons {
		if !c.equal {
			hasChange = true
			break
		}
	}
	if !hasChange {
		return nil, nil
	}

	oldDiff = make(map[string]any)
	newDiff = make(map[string]any)
	for _, c := range comparisons {
		if !c.equal {
			oldDiff[c.key] = c.oldVal
			newDiff[c.key] = c.newVal
		}
	}
	return oldDiff, newDiff
}

// estimateItemsAuditValue は明細の変更を監査できるよう、明細ごとの値を並べる。見積の金額は保存しないため、
// 明細が金額の唯一の根拠になる。NULL（不明）は nil のまま残し、0 や既定値で埋めない。
func estimateItemsAuditValue(items []model.EstimateItem) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for i := range items {
		it := &items[i]
		out = append(out, map[string]any{
			"name":                    it.Name,
			"category":                string(it.Category),
			"unit_price":              exactAuditValue(&it.UnitPrice),
			"quantity":                exactAuditValue(&it.Quantity),
			"tax_type":                string(it.TaxType),
			"tax_rate":                exactAuditValue(it.TaxRate),
			"discount_rate":           exactAuditValue(it.DiscountRate),
			"discount_amount":         derefAuditValue(it.DiscountAmount),
			"is_insurance_applicable": derefAuditValue(it.IsInsuranceApplicable),
			"sort_order":              derefAuditValue(it.SortOrder),
		})
	}
	return out
}

// exactAuditValue は ExactNumeric を末尾0なしの10進文字列にする（float64 を経由しない）。
// NULL・変換できない値は nil（不明）にし、0 で埋めない。
func exactAuditValue(n *model.ExactNumeric) any {
	if n == nil {
		return nil
	}
	text, err := n.CanonicalString()
	if err != nil {
		return nil
	}
	return text
}

// derefAuditValue は NULL 可能な列の値を取り出す。NULL は nil のまま返す。
func derefAuditValue[T any](v *T) any {
	if v == nil {
		return nil
	}
	return *v
}

// estimateStatusAuditValue は status を監査値にする。NULL（不明）は nil とし、空文字や draft に補わない。
func estimateStatusAuditValue(status *model.EstimateStatus) any {
	if status == nil {
		return nil
	}
	return string(*status)
}

// ptr 等価 helper（audit_diff.go 移動に伴う最小残置コピー・estimate diff 専用）。
func ptrUint64Equal(a, b *uint64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func ptrTimeEqual(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}
