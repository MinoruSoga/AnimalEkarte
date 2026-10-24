package medicalrecord

import (
	"encoding/json"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

type updateInquiryRequest struct {
	ChiefComplaint *string `json:"chief_complaint"`
	// ChiefComplaintTypeID は PATCH DTO で「未送信」(保持) と「JSON null」(解除) を区別する。
	// nullableUint64RequestField 契約（未送信 / null / 値 を区別する）。
	ChiefComplaintTypeID nullableUint64RequestField `json:"chief_complaint_type_id"`
	Notes                *string                    `json:"notes"`
}

func (r updateInquiryRequest) toServiceInput(clinicID, medicalRecordID uint64) UpsertInquiryInput {
	return UpsertInquiryInput{
		ClinicID:             clinicID,
		MedicalRecordID:      medicalRecordID,
		ChiefComplaintTypeID: r.ChiefComplaintTypeID.toServiceInput(),
		ChiefComplaint:       r.ChiefComplaint,
		Notes:                r.Notes,
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
