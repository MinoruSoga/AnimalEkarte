package medicalrecord

import (
	"math"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToExaminationResponse(t *testing.T) {
	petID := uint64(20)
	doctorID := uint64(30)
	medicalRecordID := uint64(40)
	revisionVersion := uint64(2)
	date := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	now := time.Now()

	tests := []struct {
		name string
		in   *model.Examination
	}{
		{
			name: "converts examination with relations",
			in: &model.Examination{
				ID:                     1,
				ClinicID:               2,
				MedicalRecordID:        &medicalRecordID,
				PetID:                  &petID,
				ExamTypeID:             3,
				DoctorID:               &doctorID,
				Date:                   date,
				ResultSummary:          "normal",
				Machine:                "machine-a",
				Status:                 model.ExaminationStatusCompleted,
				CurrentRevisionVersion: &revisionVersion,
				CreatedAt:              now,
				UpdatedAt:              now,
				Pet:                    &model.Pet{ID: petID, Name: "Pochi"},
				Doctor:                 &model.Staff{ID: doctorID, Name: "Dr. Smith"},
				ExaminationType:        &model.ExaminationType{ID: 3, Name: "Blood Test"},
			},
		},
		{
			name: "converts examination without relations",
			in: &model.Examination{
				ID:         2,
				ClinicID:   2,
				ExamTypeID: 3,
				Date:       date,
				Status:     model.ExaminationStatusPending,
				CreatedAt:  now,
				UpdatedAt:  now,
			},
		},
		{
			name: "converts zero-value examination",
			in:   &model.Examination{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toExaminationResponse(tt.in)

			if got.ID != tt.in.ID {
				t.Errorf("ID = %d, want %d", got.ID, tt.in.ID)
			}
			if got.ClinicID != tt.in.ClinicID {
				t.Errorf("ClinicID = %d, want %d", got.ClinicID, tt.in.ClinicID)
			}
			if got.ExamTypeID != tt.in.ExamTypeID {
				t.Errorf("ExamTypeID = %d, want %d", got.ExamTypeID, tt.in.ExamTypeID)
			}
			if got.ResultSummary != tt.in.ResultSummary {
				t.Errorf("ResultSummary = %q, want %q", got.ResultSummary, tt.in.ResultSummary)
			}
			if got.Machine != tt.in.Machine {
				t.Errorf("Machine = %q, want %q", got.Machine, tt.in.Machine)
			}
			if got.Status != string(tt.in.Status) {
				t.Errorf("Status = %q, want %q", got.Status, string(tt.in.Status))
			}
			if got.CurrentRevisionVersion != tt.in.CurrentRevisionVersion {
				t.Errorf(
					"CurrentRevisionVersion = %v, want %v",
					got.CurrentRevisionVersion,
					tt.in.CurrentRevisionVersion,
				)
			}

			if tt.in.Pet == nil {
				if got.Pet != nil {
					t.Errorf("Pet = %+v, want nil", got.Pet)
				}
			} else if got.Pet == nil || got.Pet.Name != tt.in.Pet.Name {
				t.Errorf("Pet = %+v, want name %q", got.Pet, tt.in.Pet.Name)
			}

			if tt.in.Doctor == nil {
				if got.Doctor != nil {
					t.Errorf("Doctor = %+v, want nil", got.Doctor)
				}
			} else if got.Doctor == nil || got.Doctor.Name != tt.in.Doctor.Name {
				t.Errorf("Doctor = %+v, want name %q", got.Doctor, tt.in.Doctor.Name)
			}

			if tt.in.ExaminationType == nil {
				if got.ExamType != nil {
					t.Errorf("ExamType = %+v, want nil", got.ExamType)
				}
			} else {
				if got.ExamType == nil {
					t.Fatalf("ExamType = nil, want non-nil")
				}
				if got.ExamType.ID != tt.in.ExaminationType.ID {
					t.Errorf("ExamType.ID = %d, want %d", got.ExamType.ID, tt.in.ExaminationType.ID)
				}
				if got.ExamType.Name != tt.in.ExaminationType.Name {
					t.Errorf("ExamType.Name = %q, want %q", got.ExamType.Name, tt.in.ExaminationType.Name)
				}
			}
		})
	}
}

func TestExamResultResponsesComputeIsAssessedFromComparableInput(t *testing.T) {
	minimum := 1.0
	maximum := 10.0
	qualitativeMinimum := "(-)"
	qualitativeMaximum := "(+)"
	numericNaN := math.NaN()
	invertedMinimum, invertedMaximum := 10.0, 1.0
	tests := []struct {
		name               string
		inspectionValue    string
		refMin             *float64
		refMax             *float64
		qualitativeMinimum *string
		qualitativeMaximum *string
		wantAssessed       bool
	}{
		{name: "numeric input compared with numeric bounds", inspectionValue: "5", refMin: &minimum, refMax: &maximum, wantAssessed: true},
		{name: "BUG-447 nonnumeric input cannot be compared with numeric bounds", inspectionValue: "陰性", refMin: &minimum, refMax: &maximum},
		{name: "qualitative input compared with qualitative bounds", inspectionValue: "(++)", qualitativeMinimum: &qualitativeMinimum, qualitativeMaximum: &qualitativeMaximum, wantAssessed: true},
		{name: "unknown input cannot be compared with qualitative bounds", inspectionValue: "陰性", qualitativeMinimum: &qualitativeMinimum, qualitativeMaximum: &qualitativeMaximum},
		{name: "empty input is not assessed", inspectionValue: "", refMin: &minimum, refMax: &maximum},
		{name: "input without bounds is not assessed", inspectionValue: "5"},
		{name: "NaN numeric boundary is not assessed", inspectionValue: "5", refMin: &numericNaN, refMax: &maximum},
		{name: "inverted numeric boundaries are not assessed", inspectionValue: "5", refMin: &invertedMinimum, refMax: &invertedMaximum},
		{
			name:               "coexisting numeric and qualitative bounds fail closed",
			inspectionValue:    "5",
			refMin:             &minimum,
			refMax:             &maximum,
			qualitativeMinimum: &qualitativeMinimum,
			qualitativeMaximum: &qualitativeMaximum,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			examResponse := toExamResultResponse(&model.ExamResult{
				InspectionValue: tt.inspectionValue,
				RefMin:          tt.refMin,
				RefMax:          tt.refMax,
				QualitativeMin:  tt.qualitativeMinimum,
				QualitativeMax:  tt.qualitativeMaximum,
			})
			if examResponse.IsAssessed != tt.wantAssessed {
				t.Errorf("exam response IsAssessed = %t, want %t", examResponse.IsAssessed, tt.wantAssessed)
			}
			assertOptionalString(t, "exam response QualitativeMin", examResponse.QualitativeMin, tt.qualitativeMinimum)
			assertOptionalString(t, "exam response QualitativeMax", examResponse.QualitativeMax, tt.qualitativeMaximum)

			labResponse := toLabExamResultItemResponse(&model.LabExamResultItem{
				InspectionValue: tt.inspectionValue,
				RefMin:          tt.refMin,
				RefMax:          tt.refMax,
				QualitativeMin:  tt.qualitativeMinimum,
				QualitativeMax:  tt.qualitativeMaximum,
			})
			if labResponse.IsAssessed != tt.wantAssessed {
				t.Errorf("lab response IsAssessed = %t, want %t", labResponse.IsAssessed, tt.wantAssessed)
			}
			assertOptionalString(t, "lab response QualitativeMin", labResponse.QualitativeMin, tt.qualitativeMinimum)
			assertOptionalString(t, "lab response QualitativeMax", labResponse.QualitativeMax, tt.qualitativeMaximum)
		})
	}
}

func assertOptionalString(t *testing.T, label string, got, want *string) {
	t.Helper()
	if want == nil {
		if got != nil {
			t.Errorf("%s = %q, want nil", label, *got)
		}
		return
	}
	if got == nil {
		t.Errorf("%s = nil, want %q", label, *want)
		return
	}
	if *got != *want {
		t.Errorf("%s = %q, want %q", label, *got, *want)
	}
}

// BUG-450: 一覧レスポンスが明細を落としていたため、カルテの検査結果一覧が常に空・
// 飼主レポートの異常件数が常に0になっていた。opt-in 時に明細が載ること、既定では
// 載らないこと、明細が is_assessed を持つことを、この3 assertion で固定する。
func TestToExaminationResponseWithItems_CarriesAssessedItems(t *testing.T) {
	refMin := 5.5
	refMax := 8.5
	exam := &model.Examination{
		ID:       1,
		ClinicID: 1,
		Items: []model.ExamResult{
			{ID: 11, ExamID: 1, Name: "RBC", InspectionValue: "9.0", RefMin: &refMin, RefMax: &refMax},
			{ID: 12, ExamID: 1, Name: "WBC", InspectionValue: "10.0"},
		},
	}

	withItems := toExaminationResponseWithItems(exam)
	if withItems.Items == nil {
		t.Fatal("Items = nil, want the exam's item details")
	}
	if got := len(*withItems.Items); got != 2 {
		t.Fatalf("len(Items) = %d, want 2", got)
	}

	// 基準値が解決できた項目は評価済み、無い項目は未評価。FE はこの差で
	// 「基準値内」と「未判定」を描き分けるため、両方が同じ経路で運ばれる必要がある。
	if !(*withItems.Items)[0].IsAssessed {
		t.Error("bounded item IsAssessed = false, want true")
	}
	if (*withItems.Items)[1].IsAssessed {
		t.Error("unbounded item IsAssessed = true, want false (unassessed is not normal)")
	}

	// 既定の一覧応答は従来どおり明細を持たない（後方互換）。
	if lean := toExaminationResponse(exam); lean.Items != nil {
		t.Errorf("default response Items = %v, want nil", lean.Items)
	}
}
