package billing

// estimate_audit_diff.go — BE9-2D ⑦: audit_diff.go の medicalrecord 移動に伴い、estimate（billing
// 残留 domain）系 diff のみ本 package へ分離（単一実装・複製なし）。移設 = billing domain 移行時。

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// Callers: estimate_service Create/Update/Delete audit best-effort. Schema: none (in-memory map for audit_logs JSON).
func extractEstimateImportantFields(e *model.Estimate) map[string]any {
	if e == nil {
		return nil
	}
	return map[string]any{
		"status":            string(e.Status),
		"title":             e.Title,
		"subtotal":          e.Subtotal,
		"tax_total":         e.TaxTotal,
		"total_amount":      e.TotalAmount,
		"insurance_amount":  e.InsuranceAmount,
		"discount_amount":   e.DiscountAmount,
		"valid_until":       e.ValidUntil,
		"comment":           e.Comment,
		"notes":             e.Notes,
		"owner_id":          e.OwnerID,
		"medical_record_id": e.MedicalRecordID,
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

	comparisons := []comparison{
		{key: "status", oldVal: string(oldEst.Status), newVal: string(newEst.Status), equal: oldEst.Status == newEst.Status},
		{key: "title", oldVal: oldEst.Title, newVal: newEst.Title, equal: oldEst.Title == newEst.Title},
		{key: "subtotal", oldVal: oldEst.Subtotal, newVal: newEst.Subtotal, equal: oldEst.Subtotal == newEst.Subtotal},
		{key: "tax_total", oldVal: oldEst.TaxTotal, newVal: newEst.TaxTotal, equal: oldEst.TaxTotal == newEst.TaxTotal},
		{key: "total_amount", oldVal: oldEst.TotalAmount, newVal: newEst.TotalAmount, equal: oldEst.TotalAmount == newEst.TotalAmount},
		{key: "insurance_amount", oldVal: oldEst.InsuranceAmount, newVal: newEst.InsuranceAmount, equal: oldEst.InsuranceAmount == newEst.InsuranceAmount},
		{key: "discount_amount", oldVal: oldEst.DiscountAmount, newVal: newEst.DiscountAmount, equal: oldEst.DiscountAmount == newEst.DiscountAmount},
		{key: "valid_until", oldVal: oldEst.ValidUntil, newVal: newEst.ValidUntil, equal: ptrTimeEqual(oldEst.ValidUntil, newEst.ValidUntil)},
		{key: "comment", oldVal: oldEst.Comment, newVal: newEst.Comment, equal: oldEst.Comment == newEst.Comment},
		{key: "notes", oldVal: oldEst.Notes, newVal: newEst.Notes, equal: oldEst.Notes == newEst.Notes},
		{key: "owner_id", oldVal: oldEst.OwnerID, newVal: newEst.OwnerID, equal: ptrUint64Equal(oldEst.OwnerID, newEst.OwnerID)},
		{key: "medical_record_id", oldVal: oldEst.MedicalRecordID, newVal: newEst.MedicalRecordID, equal: ptrUint64Equal(oldEst.MedicalRecordID, newEst.MedicalRecordID)},
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
