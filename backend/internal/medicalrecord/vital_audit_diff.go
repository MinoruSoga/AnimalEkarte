package medicalrecord

import (
	"github.com/animal-ekarte/backend/internal/model"
)

// This file holds the vital-specific audit field extraction/diff helpers moved from
// internal/service/audit_diff.go together with vitalService (BE9-2D sub-batch④a). The
// service-side audit_diff.go keeps its copies because estimate/medical_record diffs (which stay
// in internal/service) still use them; the ptr*Equal helpers below are re-declared here rather
// than shared to avoid medicalrecord importing internal/service (ADR-006). Behavior is
// byte-for-byte identical to the pre-move code so the audit payload does not change.

// extractVitalImportantFields はバイタルの重要フィールドのみ map に抽出する。
func extractVitalImportantFields(v *model.VitalRecord) map[string]any {
	if v == nil {
		return nil
	}
	return map[string]any{
		"temperature":      v.Temperature,
		"heart_rate":       v.HeartRate,
		"respiration_rate": v.RespirationRate,
		"weight":           v.Weight,
		"weight_unit":      string(v.WeightUnit),
		"recorded_at":      v.RecordedAt,
		"staff_id":         v.StaffID,
		"notes":            v.Notes,
	}
}

// diffVitalImportantFields はバイタルの変更があったフィールドのみ old/new ペアを返す。
func diffVitalImportantFields(oldRec, newRec *model.VitalRecord) (oldDiff, newDiff map[string]any) {
	type comparison struct {
		key    string
		oldVal any
		newVal any
		equal  bool
	}

	comparisons := []comparison{
		{key: "temperature", oldVal: oldRec.Temperature, newVal: newRec.Temperature, equal: ptrFloat64Equal(oldRec.Temperature, newRec.Temperature)},
		{key: "heart_rate", oldVal: oldRec.HeartRate, newVal: newRec.HeartRate, equal: ptrIntEqual(oldRec.HeartRate, newRec.HeartRate)},
		{key: "respiration_rate", oldVal: oldRec.RespirationRate, newVal: newRec.RespirationRate, equal: ptrIntEqual(oldRec.RespirationRate, newRec.RespirationRate)},
		{key: "weight", oldVal: oldRec.Weight, newVal: newRec.Weight, equal: ptrFloat64Equal(oldRec.Weight, newRec.Weight)},
		{key: "weight_unit", oldVal: string(oldRec.WeightUnit), newVal: string(newRec.WeightUnit), equal: oldRec.WeightUnit == newRec.WeightUnit},
		{key: "recorded_at", oldVal: oldRec.RecordedAt, newVal: newRec.RecordedAt, equal: oldRec.RecordedAt.Equal(newRec.RecordedAt)},
		{key: "staff_id", oldVal: oldRec.StaffID, newVal: newRec.StaffID, equal: ptrUint64Equal(oldRec.StaffID, newRec.StaffID)},
		{key: "notes", oldVal: oldRec.Notes, newVal: newRec.Notes, equal: oldRec.Notes == newRec.Notes},
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

// --- ポインタ比較ヘルパー ---

func ptrUint64Equal(a, b *uint64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func ptrFloat64Equal(a, b *float64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func ptrIntEqual(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
