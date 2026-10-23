package medicalrecord

import (
	"github.com/animal-ekarte/backend/internal/apperrors"
)

// clinicalPlanDiagnosisRequest は診断行の入力（EMR-495 A, 契約 §0.3 の closed shape から id/version を除いたもの）。
// diagnosis_slot は必須（1 または 2）。マスタ参照とスナップショットは nullable で、送信値をそのまま保存する
// （現在のマスタから補完しない）。sort_order 省略は 0。
type clinicalPlanDiagnosisRequest struct {
	DiagnosisSlot         *int16  `json:"diagnosis_slot"`
	DiagnosisTypeID       *uint64 `json:"diagnosis_type_id"`
	DiagnosisNameID       *uint64 `json:"diagnosis_name_id"`
	DiagnosisNameSnapshot *string `json:"diagnosis_name_snapshot"`
	CategoryNameSnapshot  *string `json:"category_name_snapshot"`
	SortOrder             *int    `json:"sort_order"`
}

// toDiagnosisInputs は診断行の request を service 入力へ変換する。slot の欠落は 400。
func toDiagnosisInputs(rows []clinicalPlanDiagnosisRequest) ([]ClinicalPlanDiagnosisInput, error) {
	out := make([]ClinicalPlanDiagnosisInput, 0, len(rows))
	for _, row := range rows {
		if row.DiagnosisSlot == nil {
			return nil, apperrors.WrapInvalidInput("diagnosis_slot is required")
		}
		sortOrder := 0
		if row.SortOrder != nil {
			sortOrder = *row.SortOrder
		}
		out = append(out, ClinicalPlanDiagnosisInput{
			Slot:                  *row.DiagnosisSlot,
			DiagnosisTypeID:       row.DiagnosisTypeID,
			DiagnosisNameID:       row.DiagnosisNameID,
			DiagnosisNameSnapshot: row.DiagnosisNameSnapshot,
			CategoryNameSnapshot:  row.CategoryNameSnapshot,
			SortOrder:             sortOrder,
		})
	}
	return out, nil
}

// updateClinicalPlanRequest は診療計画 PATCH のバインド struct。
// diagnoses: 未送信 / null = 診断行を変更しない、[] = 全行クリア、配列 = その集合へ置換。
type updateClinicalPlanRequest struct {
	PhysicalExam     *string                         `json:"physical_exam"`
	DiagnosisDetails *string                         `json:"diagnosis_details"`
	TreatmentPolicy  *string                         `json:"treatment_policy"`
	Diagnoses        *[]clinicalPlanDiagnosisRequest `json:"diagnoses"`
	Version          *int                            `json:"version"` // 楽観的ロック用
}

func (r updateClinicalPlanRequest) toServiceInput() (*UpdateClinicalPlanInput, error) {
	input := &UpdateClinicalPlanInput{
		PhysicalExam:     r.PhysicalExam,
		DiagnosisDetails: r.DiagnosisDetails,
		TreatmentPolicy:  r.TreatmentPolicy,
		Version:          r.Version,
	}
	if r.Diagnoses != nil {
		rows, err := toDiagnosisInputs(*r.Diagnoses)
		if err != nil {
			return nil, err
		}
		input.Diagnoses = &rows
	}
	return input, nil
}
