package service

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// extractMedicalRecordImportantFields は重要フィールドのみ map に抽出する（nil 安全）。
// 除外フィールド: version, created_at, updated_at, entered_by, record_no
func extractMedicalRecordImportantFields(r *model.MedicalRecord) map[string]any {
	if r == nil {
		return nil
	}
	m := map[string]any{
		"status":    string(r.Status),
		"date":      r.Date,
		"doctor_id": r.DoctorID,
		"owner_id":  r.OwnerID,
		"pet_id":    r.PetID,
	}
	m["appointment_id"] = r.AppointmentID
	m["next_visit_recommended_date"] = r.NextVisitRecommendedDate
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
		"weight_unit":      string(v.WeightUnit),
		"recorded_at":      v.RecordedAt,
		"staff_id":         v.StaffID,
		"notes":            v.Notes,
	}
}

// diffMedicalRecordImportantFields は変更があったフィールドのみ old/new ペアを返す。
// 変更なしの場合は nil/nil を返す（audit をスキップするシグナル）。
func diffMedicalRecordImportantFields(old, new *model.MedicalRecord) (oldDiff, newDiff map[string]any) {
	type comparison struct {
		key    string
		oldVal any
		newVal any
		equal  bool
	}

	comparisons := []comparison{
		{key: "status", oldVal: string(old.Status), newVal: string(new.Status), equal: old.Status == new.Status},
		{key: "date", oldVal: old.Date, newVal: new.Date, equal: old.Date.Equal(new.Date)},
		{key: "doctor_id", oldVal: old.DoctorID, newVal: new.DoctorID, equal: ptrUint64Equal(old.DoctorID, new.DoctorID)},
		{key: "owner_id", oldVal: old.OwnerID, newVal: new.OwnerID, equal: ptrUint64Equal(old.OwnerID, new.OwnerID)},
		{key: "pet_id", oldVal: old.PetID, newVal: new.PetID, equal: ptrUint64Equal(old.PetID, new.PetID)},
		{key: "appointment_id", oldVal: old.AppointmentID, newVal: new.AppointmentID, equal: ptrUint64Equal(old.AppointmentID, new.AppointmentID)},
		{key: "next_visit_recommended_date", oldVal: old.NextVisitRecommendedDate, newVal: new.NextVisitRecommendedDate, equal: ptrTimeEqual(old.NextVisitRecommendedDate, new.NextVisitRecommendedDate)},
		{key: "recommendation_reason", oldVal: old.RecommendationReason, newVal: new.RecommendationReason, equal: ptrStringEqual(old.RecommendationReason, new.RecommendationReason)},
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
func diffVitalImportantFields(old, new *model.VitalRecord) (oldDiff, newDiff map[string]any) {
	type comparison struct {
		key    string
		oldVal any
		newVal any
		equal  bool
	}

	comparisons := []comparison{
		{key: "temperature", oldVal: old.Temperature, newVal: new.Temperature, equal: ptrFloat64Equal(old.Temperature, new.Temperature)},
		{key: "heart_rate", oldVal: old.HeartRate, newVal: new.HeartRate, equal: ptrIntEqual(old.HeartRate, new.HeartRate)},
		{key: "respiration_rate", oldVal: old.RespirationRate, newVal: new.RespirationRate, equal: ptrIntEqual(old.RespirationRate, new.RespirationRate)},
		{key: "weight", oldVal: old.Weight, newVal: new.Weight, equal: ptrFloat64Equal(old.Weight, new.Weight)},
		{key: "weight_unit", oldVal: string(old.WeightUnit), newVal: string(new.WeightUnit), equal: old.WeightUnit == new.WeightUnit},
		{key: "recorded_at", oldVal: old.RecordedAt, newVal: new.RecordedAt, equal: old.RecordedAt.Equal(new.RecordedAt)},
		{key: "staff_id", oldVal: old.StaffID, newVal: new.StaffID, equal: ptrUint64Equal(old.StaffID, new.StaffID)},
		{key: "notes", oldVal: old.Notes, newVal: new.Notes, equal: old.Notes == new.Notes},
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

func ptrTimeEqual(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}
