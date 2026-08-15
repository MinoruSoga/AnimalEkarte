package medicalrecord

import (
	"encoding/json"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// updateClinicalPlanRequest は診療計画 PATCH のバインド struct。
// Diagnosis2TypeID / Diagnosis2NameID は nullableUint64RequestField:
// JSON未送信=未更新, null=NULLクリア, 数値=セット。
type updateClinicalPlanRequest struct {
	PhysicalExam     *string                    `json:"physical_exam"`
	DiagnosisTypeID  *uint64                    `json:"diagnosis_type_id"`
	DiagnosisNameID  *uint64                    `json:"diagnosis_name_id"`
	DiagnosisDetails *string                    `json:"diagnosis_details"`
	TreatmentPolicy  *string                    `json:"treatment_policy"`
	Diagnosis2TypeID nullableUint64RequestField `json:"diagnosis_2_type_id"`
	Diagnosis2NameID nullableUint64RequestField `json:"diagnosis_2_name_id"`
	Version          *int                       `json:"version"` // 楽観的ロック用
}

func (r updateClinicalPlanRequest) toServiceInput() *UpdateClinicalPlanInput {
	return &UpdateClinicalPlanInput{
		PhysicalExam:     r.PhysicalExam,
		DiagnosisTypeID:  r.DiagnosisTypeID,
		DiagnosisNameID:  r.DiagnosisNameID,
		DiagnosisDetails: r.DiagnosisDetails,
		TreatmentPolicy:  r.TreatmentPolicy,
		Diagnosis2TypeID: r.Diagnosis2TypeID.toServiceInput(),
		Diagnosis2NameID: r.Diagnosis2NameID.toServiceInput(),
		Version:          r.Version,
	}
}

// nullableUint64RequestField は PATCH DTO で「未送信」と「JSON null」を区別する。
// Go 標準の **uint64 だけでは encoding/json が null と未送信の両方を nil にするため、
// set フラグでフィールド存在を保持し、service 層の **uint64（nil / &nil / &&value）へ変換する。
//
// BE9-2D sub-batch④a: internal/handler.nullableUint64RequestField（owner_request.go 由来）から
// clinical_plan の消費分をローカル移植したもの。owner 側は handler package に残る（別々に存在してよい）。
type nullableUint64RequestField struct {
	set   bool
	value *uint64
}

func (f *nullableUint64RequestField) UnmarshalJSON(data []byte) error {
	f.set = true
	if string(data) == "null" {
		f.value = nil
		return nil
	}
	var value uint64
	if err := json.Unmarshal(data, &value); err != nil {
		return apperrors.WrapInvalidInput("数値IDの形式が正しくありません")
	}
	f.value = &value
	return nil
}

func (f nullableUint64RequestField) toServiceInput() **uint64 {
	if !f.set {
		return nil
	}
	return &f.value
}
