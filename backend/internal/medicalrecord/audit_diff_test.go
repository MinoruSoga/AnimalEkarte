package medicalrecord

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- extractMedicalRecordImportantFields ----

func TestExtractMedicalRecordImportantFields(t *testing.T) {
	t.Run("returns nil for nil input", func(t *testing.T) {
		got := extractMedicalRecordImportantFields(nil)
		assert.Nil(t, got)
	})

	t.Run("extracts important fields with all pointers set", func(t *testing.T) {
		doctorID := uint64(1)
		ownerID := uint64(2)
		petID := uint64(3)
		appointmentID := uint64(4)
		nextVisit := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		reason := "follow up"
		date := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)

		r := &model.MedicalRecord{
			Status:                   model.MedicalRecordStatusFinalized,
			Date:                     date,
			DoctorID:                 &doctorID,
			OwnerID:                  &ownerID,
			PetID:                    &petID,
			AppointmentID:            &appointmentID,
			NextVisitRecommendedDate: &nextVisit,
			RecommendationReason:     &reason,
		}

		got := extractMedicalRecordImportantFields(r)

		require.NotNil(t, got)
		assert.Equal(t, "finalized", got["status"])
		assert.Equal(t, date, got["date"])
		assert.Equal(t, &doctorID, got["doctor_id"])
		assert.Equal(t, &ownerID, got["owner_id"])
		assert.Equal(t, &petID, got["pet_id"])
		assert.Equal(t, &appointmentID, got["appointment_id"])
		assert.Equal(t, &nextVisit, got["next_visit_recommended_date"])
		assert.Equal(t, &reason, got["recommendation_reason"])
	})

	t.Run("extracts important fields with all pointers nil", func(t *testing.T) {
		date := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
		r := &model.MedicalRecord{
			Status: model.MedicalRecordStatusDraft,
			Date:   date,
		}

		got := extractMedicalRecordImportantFields(r)

		require.NotNil(t, got)
		assert.Equal(t, "draft", got["status"])
		assert.Nil(t, got["doctor_id"])
		assert.Nil(t, got["owner_id"])
		assert.Nil(t, got["pet_id"])
		assert.Nil(t, got["appointment_id"])
		assert.Nil(t, got["next_visit_recommended_date"])
		assert.Nil(t, got["recommendation_reason"])
	})
}

// ---- extractVitalImportantFields ----

func TestExtractVitalImportantFields(t *testing.T) {
	t.Run("returns nil for nil input", func(t *testing.T) {
		got := extractVitalImportantFields(nil)
		assert.Nil(t, got)
	})

	t.Run("extracts important fields with all pointers set", func(t *testing.T) {
		temp := 38.5
		heartRate := 90
		respRate := 20
		weight := 4.2
		staffID := uint64(7)
		recordedAt := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)

		v := &model.VitalRecord{
			Temperature:     &temp,
			HeartRate:       &heartRate,
			RespirationRate: &respRate,
			Weight:          &weight,
			WeightUnit:      model.BodyWeightUnitKg,
			RecordedAt:      recordedAt,
			StaffID:         &staffID,
			Notes:           "note",
		}

		got := extractVitalImportantFields(v)

		require.NotNil(t, got)
		assert.Equal(t, &temp, got["temperature"])
		assert.Equal(t, &heartRate, got["heart_rate"])
		assert.Equal(t, &respRate, got["respiration_rate"])
		assert.Equal(t, &weight, got["weight"])
		assert.Equal(t, "Kg", got["weight_unit"])
		assert.Equal(t, recordedAt, got["recorded_at"])
		assert.Equal(t, &staffID, got["staff_id"])
		assert.Equal(t, "note", got["notes"])
	})

	t.Run("extracts important fields with all pointers nil", func(t *testing.T) {
		recordedAt := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
		v := &model.VitalRecord{
			WeightUnit: model.BodyWeightUnitG,
			RecordedAt: recordedAt,
		}

		got := extractVitalImportantFields(v)

		require.NotNil(t, got)
		assert.Nil(t, got["temperature"])
		assert.Nil(t, got["heart_rate"])
		assert.Nil(t, got["respiration_rate"])
		assert.Nil(t, got["weight"])
		assert.Equal(t, "g", got["weight_unit"])
		assert.Nil(t, got["staff_id"])
		assert.Equal(t, "", got["notes"])
	})
}

// ---- ptrUint64Equal ----

func TestPtrUint64Equal(t *testing.T) {
	a := uint64(1)
	b := uint64(1)
	c := uint64(2)

	tests := []struct {
		name string
		a, b *uint64
		want bool
	}{
		{name: "both nil", a: nil, b: nil, want: true},
		{name: "a nil b set", a: nil, b: &b, want: false},
		{name: "a set b nil", a: &a, b: nil, want: false},
		{name: "both set equal", a: &a, b: &b, want: true},
		{name: "both set different", a: &a, b: &c, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ptrUint64Equal(tt.a, tt.b))
		})
	}
}

// ---- ptrFloat64Equal ----

func TestPtrFloat64Equal(t *testing.T) {
	a := 1.5
	b := 1.5
	c := 2.5

	tests := []struct {
		name string
		a, b *float64
		want bool
	}{
		{name: "both nil", a: nil, b: nil, want: true},
		{name: "a nil b set", a: nil, b: &b, want: false},
		{name: "a set b nil", a: &a, b: nil, want: false},
		{name: "both set equal", a: &a, b: &b, want: true},
		{name: "both set different", a: &a, b: &c, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ptrFloat64Equal(tt.a, tt.b))
		})
	}
}

// ---- ptrIntEqual ----

func TestPtrIntEqual(t *testing.T) {
	a := 5
	b := 5
	c := 10

	tests := []struct {
		name string
		a, b *int
		want bool
	}{
		{name: "both nil", a: nil, b: nil, want: true},
		{name: "a nil b set", a: nil, b: &b, want: false},
		{name: "a set b nil", a: &a, b: nil, want: false},
		{name: "both set equal", a: &a, b: &b, want: true},
		{name: "both set different", a: &a, b: &c, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ptrIntEqual(tt.a, tt.b))
		})
	}
}

// ---- ptrStringEqual ----

func TestPtrStringEqual(t *testing.T) {
	a := "foo"
	b := "foo"
	c := "bar"

	tests := []struct {
		name string
		a, b *string
		want bool
	}{
		{name: "both nil", a: nil, b: nil, want: true},
		{name: "a nil b set", a: nil, b: &b, want: false},
		{name: "a set b nil", a: &a, b: nil, want: false},
		{name: "both set equal", a: &a, b: &b, want: true},
		{name: "both set different", a: &a, b: &c, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ptrStringEqual(tt.a, tt.b))
		})
	}
}

// ---- ptrTimeEqual ----

func TestPtrTimeEqual(t *testing.T) {
	a := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	b := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	c := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		a, b *time.Time
		want bool
	}{
		{name: "both nil", a: nil, b: nil, want: true},
		{name: "a nil b set", a: nil, b: &b, want: false},
		{name: "a set b nil", a: &a, b: nil, want: false},
		{name: "both set equal", a: &a, b: &b, want: true},
		{name: "both set different", a: &a, b: &c, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ptrTimeEqual(tt.a, tt.b))
		})
	}
}
