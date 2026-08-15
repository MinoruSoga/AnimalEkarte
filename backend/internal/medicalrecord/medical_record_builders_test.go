package medicalrecord

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- generateCryptoRandomString ----

func TestGenerateCryptoRandomString(t *testing.T) {
	t.Run("returns requested length", func(t *testing.T) {
		for _, length := range []int{0, 1, 6, 20} {
			got := generateCryptoRandomString(length)
			assert.Len(t, got, length)
		}
	})

	t.Run("only uses the expected charset", func(t *testing.T) {
		got := generateCryptoRandomString(64)
		matched, err := regexp.MatchString(`^[A-Za-z0-9]*$`, got)
		assert.NoError(t, err)
		assert.True(t, matched, "generated string %q contains characters outside the charset", got)
	})

	t.Run("is not deterministic across calls (statistical)", func(t *testing.T) {
		a := generateCryptoRandomString(16)
		b := generateCryptoRandomString(16)
		assert.NotEqual(t, a, b, "two 16-char random strings collided — extremely unlikely unless RNG is broken")
	})
}

// ---- generateRecordNo ----

func TestGenerateRecordNo(t *testing.T) {
	date := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	got := generateRecordNo(date, 7)
	matched, err := regexp.MatchString(`^MR-20260415-7-[A-Za-z0-9]{6}$`, got)
	assert.NoError(t, err)
	assert.True(t, matched, "generateRecordNo output %q does not match expected format", got)
}

// ---- buildMedicalRecordUpdate ----

func TestBuildMedicalRecordUpdate(t *testing.T) {
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	nextVisit := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	ownerID := uint64(1)
	petID := uint64(2)
	doctorID := uint64(3)
	apptID := uint64(4)
	status := model.MedicalRecordStatusFinalized
	visitType := model.VisitType("revisit")

	tests := []struct {
		name  string
		input UpdateMedicalRecordInput
		want  map[string]any
	}{
		{
			name:  "全フィールド nil は空 map",
			input: UpdateMedicalRecordInput{},
			want:  map[string]any{},
		},
		{
			name:  "Date のみ",
			input: UpdateMedicalRecordInput{Date: &date},
			want:  map[string]any{"date": date},
		},
		{
			name:  "OwnerID のみ",
			input: UpdateMedicalRecordInput{OwnerID: &ownerID},
			want:  map[string]any{"owner_id": ownerID},
		},
		{
			name:  "PetID のみ",
			input: UpdateMedicalRecordInput{PetID: &petID},
			want:  map[string]any{"pet_id": petID},
		},
		{
			name:  "DoctorID のみ",
			input: UpdateMedicalRecordInput{DoctorID: &doctorID},
			want:  map[string]any{"doctor_id": doctorID},
		},
		{
			name:  "AppointmentID は作成後immutableのため更新mapへ含めない",
			input: UpdateMedicalRecordInput{AppointmentID: &apptID},
			want:  map[string]any{},
		},
		{
			name:  "Status のみ",
			input: UpdateMedicalRecordInput{Status: &status},
			want:  map[string]any{"status": status},
		},
		{
			name:  "VisitType のみ",
			input: UpdateMedicalRecordInput{VisitType: &visitType},
			want:  map[string]any{"visit_type": visitType},
		},
		{
			name:  "NextVisitRecommendedDate が設定される",
			input: UpdateMedicalRecordInput{NextVisitRecommendedDate: &nextVisit},
			want:  map[string]any{"next_visit_recommended_date": nextVisit},
		},
		{
			name:  "ClearNextVisitRecommendedDate=true は NULL クリア",
			input: UpdateMedicalRecordInput{ClearNextVisitRecommendedDate: true},
			want:  map[string]any{"next_visit_recommended_date": nil},
		},
		{
			name: "ClearNextVisitRecommendedDate が NextVisitRecommendedDate より優先される",
			input: UpdateMedicalRecordInput{
				ClearNextVisitRecommendedDate: true,
				NextVisitRecommendedDate:      &nextVisit,
			},
			want: map[string]any{"next_visit_recommended_date": nil},
		},
		{
			name: "全フィールド指定",
			input: UpdateMedicalRecordInput{
				Date:          &date,
				OwnerID:       &ownerID,
				PetID:         &petID,
				DoctorID:      &doctorID,
				AppointmentID: &apptID,
				Status:        &status,
				VisitType:     &visitType,
			},
			want: map[string]any{
				"date":       date,
				"owner_id":   ownerID,
				"pet_id":     petID,
				"doctor_id":  doctorID,
				"status":     status,
				"visit_type": visitType,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildMedicalRecordUpdate(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---- buildMedicalRecordForCreate ----

func TestBuildMedicalRecordForCreate(t *testing.T) {
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	ownerID := uint64(1)
	petID := uint64(2)
	doctorID := uint64(3)
	apptID := uint64(4)
	nextVisit := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	reason := "revisit"
	enteredBy := uint64(9)

	t.Run("Status 指定時はその値を使う", func(t *testing.T) {
		status := model.MedicalRecordStatusFinalized
		input := &CreateMedicalRecordInput{
			RecordNo:                 "MR-20260401-1-ABC123",
			Date:                     date,
			OwnerID:                  &ownerID,
			PetID:                    &petID,
			DoctorID:                 &doctorID,
			AppointmentID:            &apptID,
			Status:                   &status,
			VisitType:                model.VisitType("revisit"),
			NextVisitRecommendedDate: &nextVisit,
			RecommendationReason:     &reason,
			EnteredBy:                &enteredBy,
		}
		got := buildMedicalRecordForCreate(1, input)
		assert.Equal(t, uint64(1), got.ClinicID)
		assert.Equal(t, input.RecordNo, got.RecordNo)
		assert.Equal(t, date, got.Date)
		assert.Equal(t, &ownerID, got.OwnerID)
		assert.Equal(t, &petID, got.PetID)
		assert.Equal(t, &doctorID, got.DoctorID)
		assert.Equal(t, &apptID, got.AppointmentID)
		assert.Equal(t, status, got.Status)
		if assert.NotNil(t, got.VisitType) {
			assert.Equal(t, model.VisitType("revisit"), *got.VisitType)
		}
		assert.Equal(t, &nextVisit, got.NextVisitRecommendedDate)
		assert.Equal(t, &reason, got.RecommendationReason)
		assert.Equal(t, &enteredBy, got.EnteredBy)
	})

	t.Run("Status nil の場合は model のゼロ値のまま", func(t *testing.T) {
		input := &CreateMedicalRecordInput{
			RecordNo:  "MR-20260401-1-XYZ789",
			Date:      date,
			VisitType: model.VisitType("first_visit"),
		}
		got := buildMedicalRecordForCreate(1, input)
		assert.Equal(t, model.MedicalRecordStatus(""), got.Status)
		if assert.NotNil(t, got.VisitType) {
			assert.Equal(t, model.VisitType("first_visit"), *got.VisitType)
		}
	})
}
