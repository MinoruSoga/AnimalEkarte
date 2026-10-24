package medicalrecord

import (
	"github.com/animal-ekarte/backend/internal/model"
)

// extractMedicalRecordImportantFields は重要フィールドのみ map に抽出する（nil 安全）。
// 除外フィールド: version, created_at, updated_at, entered_by, record_no
func extractMedicalRecordImportantFields(r *model.MedicalRecord) map[string]any {
	if r == nil {
		return nil
	}
	m := map[string]any{
		"date":      r.Date,
		"doctor_id": r.DoctorID,
		"owner_id":  r.OwnerID,
		"pet_id":    r.PetID,
	}
	m["appointment_id"] = r.AppointmentID
	m["recommendation_reason"] = r.RecommendationReason
	return m
}

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
		"weight_unit":      weightUnitAuditValue(v.WeightUnit),
		"recorded_at":      v.RecordedAt,
		"staff_id":         v.StaffID,
		"notes":            v.Notes,
	}
}

// extractEstimateImportantFields は見積の重要フィールドのみ map に抽出する（nil 安全）。

// diffEstimateImportantFields は変更があったフィールドのみ old/new ペアを返す。

// diffMedicalRecordImportantFields は変更があったフィールドのみ old/new ペアを返す。
// 変更なしの場合は nil/nil を返す（audit をスキップするシグナル）。
func diffMedicalRecordImportantFields(oldRec, newRec *model.MedicalRecord) (oldDiff, newDiff map[string]any) {
	type comparison struct {
		key    string
		oldVal any
		newVal any
		equal  bool
	}

	comparisons := []comparison{
		{key: "date", oldVal: oldRec.Date, newVal: newRec.Date, equal: oldRec.Date.Equal(newRec.Date)},
		{key: "doctor_id", oldVal: oldRec.DoctorID, newVal: newRec.DoctorID, equal: ptrUint64Equal(oldRec.DoctorID, newRec.DoctorID)},
		{key: "owner_id", oldVal: oldRec.OwnerID, newVal: newRec.OwnerID, equal: ptrUint64Equal(oldRec.OwnerID, newRec.OwnerID)},
		{key: "pet_id", oldVal: oldRec.PetID, newVal: newRec.PetID, equal: ptrUint64Equal(oldRec.PetID, newRec.PetID)},
		{key: "appointment_id", oldVal: oldRec.AppointmentID, newVal: newRec.AppointmentID, equal: ptrUint64Equal(oldRec.AppointmentID, newRec.AppointmentID)},
		{key: "recommendation_reason", oldVal: oldRec.RecommendationReason, newVal: newRec.RecommendationReason, equal: ptrStringEqual(oldRec.RecommendationReason, newRec.RecommendationReason)},
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

// diffVitalImportantFields はバイタルの変更があったフィールドのみ old/new ペアを返す。
func diffVitalImportantFields(oldRec, newRec *model.VitalRecord) (oldDiff, newDiff map[string]any) {
	type comparison struct {
		key    string
		oldVal any
		newVal any
		equal  bool
	}

	comparisons := []comparison{
		{key: "temperature", oldVal: oldRec.Temperature, newVal: newRec.Temperature, equal: ptrExactNumericEqual(oldRec.Temperature, newRec.Temperature)},
		{key: "heart_rate", oldVal: oldRec.HeartRate, newVal: newRec.HeartRate, equal: ptrIntEqual(oldRec.HeartRate, newRec.HeartRate)},
		{key: "respiration_rate", oldVal: oldRec.RespirationRate, newVal: newRec.RespirationRate, equal: ptrIntEqual(oldRec.RespirationRate, newRec.RespirationRate)},
		{key: "weight", oldVal: oldRec.Weight, newVal: newRec.Weight, equal: ptrExactNumericEqual(oldRec.Weight, newRec.Weight)},
		{key: "weight_unit", oldVal: weightUnitAuditValue(oldRec.WeightUnit), newVal: weightUnitAuditValue(newRec.WeightUnit), equal: ptrWeightUnitEqual(oldRec.WeightUnit, newRec.WeightUnit)},
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

func ptrStringEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// weightUnitAuditValue は体重単位を監査 diff 用の値へ写す。nil（未記録）は他のポインタ項目と同じく nil（不在）で記録し、
// 既定単位（Kg）や "<nil>" のような文字列で補完しない。
func weightUnitAuditValue(u *model.BodyWeightUnit) any {
	if u == nil {
		return nil
	}
	return string(*u)
}

// ptrWeightUnitEqual は監査 diff の変更検出用。nil 同士のみ等価で、nil と値は不一致（変更あり）とする。
func ptrWeightUnitEqual(a, b *model.BodyWeightUnit) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}
