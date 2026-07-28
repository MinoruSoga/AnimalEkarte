package medicalrecord

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type treatmentResponse struct {
	ID              string  `json:"id"`
	MedicalRecordID string  `json:"medical_record_id"`
	ItemType        string  `json:"item_type"`
	ConsultationID  *string `json:"consultation_id,omitempty"`
	ProcedureID     *string `json:"procedure_id,omitempty"`
	MedicineID      *string `json:"medicine_id,omitempty"`
	InventoryID     *string `json:"inventory_id,omitempty"`
	UnitPrice       int64   `json:"unit_price"`
	Quantity        float64 `json:"quantity"`
	Selected        bool    `json:"is_selected"`
	Status          string  `json:"status"`
	Content         string  `json:"content"`
	Memo            string  `json:"memo"`
	AdminRoute      string  `json:"admin_route"`
	Insurance       bool    `json:"is_insurance"`
	DiscountRate    float64 `json:"discount_rate"`
	DiscountAmount  int64   `json:"discount_amount"`
	SortOrder       int     `json:"sort_order"`
	// #201 投与量自動計算の保存時スナップショット（medicine 明細・per_weight 計算時のみ値あり）
	DoseWeightKg      *float64        `json:"dose_weight_kg,omitempty"`
	DoseWeightSource  *string         `json:"dose_weight_source,omitempty"`
	DoseAmountMg      *float64        `json:"dose_amount_mg,omitempty"`
	DoseAmountUnit    *string         `json:"dose_amount_unit,omitempty"`
	DoseParamSnapshot json.RawMessage `json:"dose_param_snapshot,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

func toTreatmentResponse(t *model.Treatment) treatmentResponse {
	masterIDs := treatmentMasterIDsForResponse(t)
	return treatmentResponse{
		ID:                strconv.FormatUint(t.ID, 10),
		MedicalRecordID:   strconv.FormatUint(t.MedicalRecordID, 10),
		ItemType:          string(t.ItemType),
		ConsultationID:    masterIDs.consultation,
		ProcedureID:       masterIDs.procedure,
		MedicineID:        masterIDs.medicine,
		InventoryID:       masterIDs.inventory,
		UnitPrice:         t.UnitPrice,
		Quantity:          t.Quantity,
		Selected:          t.IsSelected,
		Status:            string(t.Status),
		Content:           t.Content,
		Memo:              t.Memo,
		AdminRoute:        t.AdminRoute,
		Insurance:         t.IsInsurance,
		DiscountRate:      t.DiscountRate,
		DiscountAmount:    t.DiscountAmount,
		SortOrder:         t.SortOrder,
		DoseWeightKg:      t.DoseWeightKg,
		DoseWeightSource:  t.DoseWeightSource,
		DoseAmountMg:      t.DoseAmountMg,
		DoseAmountUnit:    t.DoseAmountUnit,
		DoseParamSnapshot: t.DoseParamSnapshot,
		CreatedAt:         httpapi.LocalTime(t.CreatedAt),
		UpdatedAt:         httpapi.LocalTime(t.UpdatedAt),
	}
}

// petTreatmentHistoryResponse は #158 飼主レポート用のペット横断治療履歴 1 行。
// 診療日 (medical_records.date) と薬剤/処置の名称・麻酔種別を含めて表示用に整形する。
type petTreatmentHistoryResponse struct {
	ID              string     `json:"id"`
	MedicalRecordID string     `json:"medical_record_id"`
	Date            *time.Time `json:"date"`
	ItemType        string     `json:"item_type"`
	Content         string     `json:"content"`
	Memo            string     `json:"memo"`
	AdminRoute      string     `json:"admin_route"`
	Quantity        float64    `json:"quantity"`
	UnitPrice       int64      `json:"unit_price"`
	Status          string     `json:"status"`
	MedicineID      *string    `json:"medicine_id,omitempty"`
	MedicineName    *string    `json:"medicine_name,omitempty"`
	ProcedureID     *string    `json:"procedure_id,omitempty"`
	ProcedureName   *string    `json:"procedure_name,omitempty"`
	Anesthesia      *string    `json:"anesthesia,omitempty"`
	IsSurgery       *bool      `json:"is_surgery,omitempty"`
}

func toPetTreatmentHistoryResponse(t *model.Treatment) petTreatmentHistoryResponse {
	masterIDs := treatmentMasterIDsForResponse(t)
	resp := petTreatmentHistoryResponse{
		ID:              strconv.FormatUint(t.ID, 10),
		MedicalRecordID: strconv.FormatUint(t.MedicalRecordID, 10),
		ItemType:        string(t.ItemType),
		Content:         t.Content,
		Memo:            t.Memo,
		AdminRoute:      t.AdminRoute,
		Quantity:        t.Quantity,
		UnitPrice:       t.UnitPrice,
		Status:          string(t.Status),
		MedicineID:      masterIDs.medicine,
		ProcedureID:     masterIDs.procedure,
	}
	// 診療日は medical_records.date 由来（treatments.created_at は入力時刻なので使わない）。
	if t.MedicalRecord != nil {
		resp.Date = httpapi.LocalTimePtr(&t.MedicalRecord.Date)
	}
	if masterIDs.medicine != nil && t.Medicine != nil {
		name := t.Medicine.Name
		resp.MedicineName = &name
	}
	if masterIDs.procedure != nil && t.Procedure != nil {
		name := t.Procedure.Name
		resp.ProcedureName = &name
		anesthesia := string(t.Procedure.Anesthesia)
		resp.Anesthesia = &anesthesia
		isSurgery := t.Procedure.IsSurgery
		resp.IsSurgery = &isSurgery
	}
	return resp
}

type treatmentResponseMasterIDs struct {
	consultation *string
	procedure    *string
	medicine     *string
	inventory    *string
}

func treatmentMasterIDsForResponse(treatment *model.Treatment) treatmentResponseMasterIDs {
	if treatment == nil {
		return treatmentResponseMasterIDs{}
	}

	var consultationID, consultationClinicID uint64
	if treatment.Consultation != nil {
		consultationID, consultationClinicID = treatment.Consultation.ID, treatment.Consultation.ClinicID
	}
	var procedureID, procedureClinicID uint64
	if treatment.Procedure != nil {
		procedureID, procedureClinicID = treatment.Procedure.ID, treatment.Procedure.ClinicID
	}
	var medicineID, medicineClinicID uint64
	if treatment.Medicine != nil {
		medicineID, medicineClinicID = treatment.Medicine.ID, treatment.Medicine.ClinicID
	}
	var inventoryID, inventoryClinicID uint64
	if treatment.Inventory != nil {
		inventoryID, inventoryClinicID = treatment.Inventory.ID, treatment.Inventory.ClinicID
	}

	return treatmentResponseMasterIDs{
		consultation: treatmentMasterIDForResponse(
			treatment.ConsultationID, consultationID, consultationClinicID, treatment.MedicalRecord,
		),
		procedure: treatmentMasterIDForResponse(
			treatment.ProcedureID, procedureID, procedureClinicID, treatment.MedicalRecord,
		),
		medicine: treatmentMasterIDForResponse(
			treatment.MedicineID, medicineID, medicineClinicID, treatment.MedicalRecord,
		),
		inventory: treatmentMasterIDForResponse(
			treatment.InventoryID, inventoryID, inventoryClinicID, treatment.MedicalRecord,
		),
	}
}

func treatmentMasterIDForResponse(
	referenceID *uint64,
	loadedID, loadedClinicID uint64,
	parent *model.MedicalRecord,
) *string {
	if referenceID == nil {
		return nil
	}
	// Freshly-created writes have no preloaded parent/masters; the service has
	// already validated those request-derived IDs transactionally. Repository
	// reads always preload and sanitize the parent/master graph.
	if parent == nil && loadedID == 0 {
		return uint64PtrToStringPtr(referenceID)
	}
	if parent == nil ||
		loadedID != *referenceID ||
		loadedClinicID != parent.ClinicID {
		return nil
	}
	return uint64PtrToStringPtr(referenceID)
}

// uint64PtrToStringPtr は *uint64 を *string に変換するヘルパー。
// nilの場合はnilを返す。
func uint64PtrToStringPtr(v *uint64) *string {
	if v == nil {
		return nil
	}
	s := strconv.FormatUint(*v, 10)
	return &s
}
