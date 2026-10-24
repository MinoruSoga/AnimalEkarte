package medicalrecord

import (
	"context"
	"fmt"
	"sort"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-495 A: 診療計画の診断は clinical_plans の単一列ではなく clinical_plan_diagnoses の行で保持する
// （001_init.sql: slot は 1/2 の診断グループ、各 slot に複数行可・件数/名称の一意制約なし）。
// 行の version はストレージ（entity_version_guard trigger）が管理するため、書込側は version を設定しない。

const (
	// maxClinicalPlanDiagnosisRows は 1 plan あたりの診断行の入力上限（入力サイズ guard。業務規則ではない）。
	maxClinicalPlanDiagnosisRows = 50
	// 001 の CHECK(diagnosis_slot IN (1,2))。
	minClinicalPlanDiagnosisSlot int16 = 1
	maxClinicalPlanDiagnosisSlot int16 = 2
)

// ClinicalPlanDiagnosisInput は診断行の書込値（契約 §0.3 の closed shape から id/version を除いたもの）。
// マスタ参照とスナップショットは入力のまま保存し、現在のマスタから補完しない。
type ClinicalPlanDiagnosisInput struct {
	Slot                  int16
	DiagnosisTypeID       *uint64
	DiagnosisNameID       *uint64
	DiagnosisNameSnapshot *string
	CategoryNameSnapshot  *string
	SortOrder             int
}

// ClinicalPlanView は診療計画本体と生存中の診断行（slot, sort_order, id 昇順）。
// model.ClinicalPlan は診断行を保持できないため service / handler 間の受け渡し型として用いる。
type ClinicalPlanView struct {
	Plan      *model.ClinicalPlan
	Diagnoses []model.ClinicalPlanDiagnosis
}

// sortedDiagnosisInputs は入力を (slot, sort_order) で安定ソートしたコピーを返す。
// 同一キーは入力順を保つ（挿入順 = id 昇順 = 表示順）。
func sortedDiagnosisInputs(rows []ClinicalPlanDiagnosisInput) []ClinicalPlanDiagnosisInput {
	out := make([]ClinicalPlanDiagnosisInput, len(rows))
	copy(out, rows)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Slot != out[j].Slot {
			return out[i].Slot < out[j].Slot
		}
		return out[i].SortOrder < out[j].SortOrder
	})
	return out
}

func equalOptionalUint64(a, b *uint64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func equalOptionalString(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

// diagnosisRowsEqual は生存中の行（slot, sort_order, id 昇順）と入力が 6 値すべてで一致するかを返す。
func diagnosisRowsEqual(live []model.ClinicalPlanDiagnosis, rows []ClinicalPlanDiagnosisInput) bool {
	if len(live) != len(rows) {
		return false
	}
	sorted := sortedDiagnosisInputs(rows)
	for i := range live {
		l, r := live[i], sorted[i]
		if l.DiagnosisSlot != r.Slot ||
			l.SortOrder != r.SortOrder ||
			!equalOptionalUint64(l.DiagnosisTypeID, r.DiagnosisTypeID) ||
			!equalOptionalUint64(l.DiagnosisNameID, r.DiagnosisNameID) ||
			!equalOptionalString(l.DiagnosisNameSnapshot, r.DiagnosisNameSnapshot) ||
			!equalOptionalString(l.CategoryNameSnapshot, r.CategoryNameSnapshot) {
			return false
		}
	}
	return true
}

// validateDiagnosisRowShape は DB に触れずに入力の形を検査する（件数 guard・slot の値域）。
func validateDiagnosisRowShape(rows []ClinicalPlanDiagnosisInput) error {
	if len(rows) > maxClinicalPlanDiagnosisRows {
		return apperrors.WrapInvalidInput(fmt.Sprintf("diagnoses must be %d rows or fewer", maxClinicalPlanDiagnosisRows))
	}
	for _, row := range rows {
		if row.Slot < minClinicalPlanDiagnosisSlot || row.Slot > maxClinicalPlanDiagnosisSlot {
			return apperrors.WrapInvalidInput("diagnosis_slot must be 1 or 2")
		}
	}
	return nil
}

// validateDiagnosisRows は診断行を検査する: 形（件数・slot）→ マスタ参照の clinic 所有権 →
// type/name が両方ある行の type↔name 整合（行は完全な値なので全行を強制する）。
// スナップショットは検査・補完しない（入力のまま保存する）。
func validateDiagnosisRows(
	ctx context.Context,
	clinicID uint64,
	rows []ClinicalPlanDiagnosisInput,
	diagTypeRepo DiagnosisTypeRepository,
	diagNameRepo DiagnosisNameRepository,
) error {
	if err := validateDiagnosisRowShape(rows); err != nil {
		return err
	}
	typeIDs := make([]*uint64, 0, len(rows))
	nameIDs := make([]*uint64, 0, len(rows))
	for _, row := range rows {
		typeIDs = append(typeIDs, row.DiagnosisTypeID)
		nameIDs = append(nameIDs, row.DiagnosisNameID)
	}
	if err := validateDiagnosisMasterFKs(ctx, clinicID, typeIDs, nameIDs, diagTypeRepo, diagNameRepo); err != nil {
		return err
	}
	for _, row := range rows {
		if err := assertDiagnosisNameBelongsToType(ctx, clinicID, row.DiagnosisTypeID, row.DiagnosisNameID, "diagnosis", diagNameRepo); err != nil {
			return err
		}
	}
	return nil
}
