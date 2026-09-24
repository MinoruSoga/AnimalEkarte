package medicalrecord

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestCreateVitalRequest_ToServiceInput(t *testing.T) {
	recordedAt := time.Date(2026, 5, 28, 10, 30, 0, 0, time.UTC)
	staffID := uint64(11)
	temperature := 38.5
	heartRate := 120
	respirationRate := 30
	weight := 4.2
	weightUnit := "kg"
	notes := "stable"

	req := createVitalRequest{
		RecordedAt:      recordedAt,
		StaffID:         &staffID,
		Temperature:     &temperature,
		HeartRate:       &heartRate,
		RespirationRate: &respirationRate,
		Weight:          &weight,
		WeightUnit:      &weightUnit,
		Notes:           &notes,
	}

	input := req.toServiceInput(1, 2)

	if input.ClinicID != 1 {
		t.Fatalf("ClinicID = %d, want 1", input.ClinicID)
	}
	if input.PetID != 2 {
		t.Fatalf("PetID = %d, want 2", input.PetID)
	}
	if !input.RecordedAt.Equal(recordedAt) {
		t.Fatalf("RecordedAt = %v, want %v", input.RecordedAt, recordedAt)
	}
	if input.StaffID != &staffID || input.Temperature != &temperature || input.HeartRate != &heartRate ||
		input.RespirationRate != &respirationRate || input.Weight != &weight {
		t.Fatalf("measurement pointers were not preserved")
	}
	if input.WeightUnit == nil || string(*input.WeightUnit) != weightUnit {
		t.Fatalf("WeightUnit = %v, want %q", input.WeightUnit, weightUnit)
	}
	if input.Notes != notes {
		t.Fatalf("Notes = %q, want %q", input.Notes, notes)
	}
}

// TestCreateVitalRequest_JSONBinding は POST body のワイヤ契約を直接検証する:
// canonical key は `notes`（legacy `note` は無視）、optional fields は null 許容の
// ポインタバインドで、Notes nil は toServiceInput で DB 安全な "" に正規化される。
func TestCreateVitalRequest_JSONBinding(t *testing.T) {
	cases := []struct {
		name      string
		body      string
		wantNotes string
		wantNil   bool
	}{
		{
			name:      "notes string is preserved",
			body:      `{"recorded_at":"2026-05-28T10:30:00Z","temperature":38.5,"notes":"食欲やや低下"}`,
			wantNotes: "食欲やや低下",
		},
		{
			name:      "notes null binds nil and normalizes to empty",
			body:      `{"recorded_at":"2026-05-28T10:30:00Z","temperature":38.5,"notes":null}`,
			wantNotes: "",
			wantNil:   true,
		},
		{
			name:      "notes absent binds nil and normalizes to empty",
			body:      `{"recorded_at":"2026-05-28T10:30:00Z","temperature":38.5}`,
			wantNotes: "",
			wantNil:   true,
		},
		{
			name:      "legacy note key is ignored",
			body:      `{"recorded_at":"2026-05-28T10:30:00Z","temperature":38.5,"note":"legacy memo"}`,
			wantNotes: "",
			wantNil:   true,
		},
		{
			name:      "all optional fields null bind as nil pointers",
			body:      `{"recorded_at":"2026-05-28T10:30:00Z","staff_id":null,"temperature":null,"heart_rate":null,"respiration_rate":null,"weight":null,"weight_unit":null,"notes":null}`,
			wantNotes: "",
			wantNil:   true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var req createVitalRequest
			if err := json.Unmarshal([]byte(tt.body), &req); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if tt.wantNil && req.Notes != nil {
				t.Fatalf("Notes = %q, want nil", *req.Notes)
			}
			if !tt.wantNil && (req.Notes == nil || *req.Notes != tt.wantNotes) {
				t.Fatalf("Notes = %v, want %q", req.Notes, tt.wantNotes)
			}
			input := req.toServiceInput(1, 2)
			if input.Notes != tt.wantNotes {
				t.Fatalf("service input Notes = %q, want %q", input.Notes, tt.wantNotes)
			}
		})
	}

	t.Run("null optional fields bind all-nil pointers", func(t *testing.T) {
		var req createVitalRequest
		body := `{"recorded_at":"2026-05-28T10:30:00Z","staff_id":null,"temperature":null,"heart_rate":null,"respiration_rate":null,"weight":null,"weight_unit":null,"notes":null}`
		if err := json.Unmarshal([]byte(body), &req); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if req.StaffID != nil || req.Temperature != nil || req.HeartRate != nil ||
			req.RespirationRate != nil || req.Weight != nil || req.WeightUnit != nil || req.Notes != nil {
			t.Fatalf("req = %+v, want all optional fields nil", req)
		}
	})
}

func TestUpdateVitalRequest_ToServiceInput(t *testing.T) {
	recordedAt := time.Date(2026, 5, 28, 10, 30, 0, 0, time.UTC)
	staffID := uint64(0)
	temperature := 0.0
	heartRate := 0
	respirationRate := 0
	weight := 0.0
	weightUnit := "g"
	notes := ""
	version := 4

	req := updateVitalRequest{
		RecordedAt:      &recordedAt,
		StaffID:         &staffID,
		Temperature:     &temperature,
		HeartRate:       &heartRate,
		RespirationRate: &respirationRate,
		Weight:          &weight,
		WeightUnit:      &weightUnit,
		Notes:           &notes,
		Version:         &version,
	}

	input := req.toServiceInput(9)

	if input.RecordedAt != &recordedAt || input.StaffID != &staffID || input.Temperature != &temperature ||
		input.HeartRate != &heartRate || input.RespirationRate != &respirationRate || input.Weight != &weight ||
		input.Notes != &notes {
		t.Fatalf("update pointers were not preserved")
	}
	if input.WeightUnit == nil || string(*input.WeightUnit) != weightUnit {
		t.Fatalf("WeightUnit = %v, want %q", input.WeightUnit, weightUnit)
	}
	if input.ActorID == nil || *input.ActorID != 9 {
		t.Fatalf("ActorID = %v, want 9", input.ActorID)
	}
	if input.Version == nil || *input.Version != version {
		t.Fatalf("Version = %v, want %d", input.Version, version)
	}
}

func TestUpdateVitalRequest_ToServiceInput_NilFields(t *testing.T) {
	input := (&updateVitalRequest{}).toServiceInput(9)

	if input.RecordedAt != nil || input.StaffID != nil || input.Temperature != nil || input.HeartRate != nil ||
		input.RespirationRate != nil || input.Weight != nil || input.WeightUnit != nil || input.Notes != nil ||
		input.Version != nil {
		t.Fatalf("input = %+v, want request fields nil", input)
	}
	if input.ActorID == nil || *input.ActorID != 9 {
		t.Fatalf("ActorID = %v, want 9", input.ActorID)
	}
}

func TestValidateVitalWeightFields(t *testing.T) {
	nan := math.NaN()
	inf := math.Inf(1)
	zero := 0.0
	neg := -0.1
	ok := 8.5
	unitKg := "Kg"
	unitG := "g"
	unitBad := "lb"
	unitLower := "kg"

	cases := []struct {
		name    string
		weight  *float64
		unit    *string
		wantErr bool
	}{
		{name: "nil weight and unit ok", weight: nil, unit: nil, wantErr: false},
		{name: "ok weight with Kg", weight: &ok, unit: &unitKg, wantErr: false},
		{name: "ok weight with g", weight: &ok, unit: &unitG, wantErr: false},
		{name: "unit only ok", weight: nil, unit: &unitKg, wantErr: false},
		{name: "NaN rejected", weight: &nan, unit: &unitKg, wantErr: true},
		{name: "Inf rejected", weight: &inf, unit: nil, wantErr: true},
		{name: "zero rejected", weight: &zero, unit: &unitKg, wantErr: true},
		{name: "negative rejected", weight: &neg, unit: &unitG, wantErr: true},
		{name: "invalid unit rejected", weight: &ok, unit: &unitBad, wantErr: true},
		{name: "lowercase kg rejected", weight: &ok, unit: &unitLower, wantErr: true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVitalWeightFields(tt.weight, tt.unit)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !apperrors.IsInvalidInput(err) {
					t.Fatalf("want invalid input, got %v", err)
				}
				// エラー文言に SQL / 内部詳細を含めない（安定 400 契約）
				msg := err.Error()
				for _, leak := range []string{"SQL", "sqlstate", "gorm", "clinic_id="} {
					if strings.Contains(msg, leak) {
						t.Fatalf("error message leaked internal detail %q: %q", leak, msg)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
