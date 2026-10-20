package medicalrecord

import (
	"encoding/json"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-488: 薬剤3値の wire 契約（null 契約の統一・checkupTypeInsuranceField と同一方針）。
// encoding/json は *T/**T に null を落とすと省略と区別できないため、専有のフィールド型で
// 「未送信」「明示 null」「値」を区別する。
// create: 省略 → NULL 保存（未入力は正当。推測補完しない）/ 明示 null → 400 / 値 → 保存。
// update(PATCH): 省略 → 変更なし / 明示 null → 400（明示クリア経路はない）/ 値 → 保存。
// NaN は臨床量として無意味なため wire で 400（PostgreSQL numeric は NaN を格納できてしまう）。

const (
	errMsgDoseAmountNull      = "薬剤3値に null は指定できません。省略するか数値を指定してください"
	errMsgDoseAmountNotNumber = "薬剤3値には数値を指定してください"
	errMsgDoseAmountNaN       = "薬剤3値に NaN は指定できません"
	errMsgDoseTextNull        = "薬剤3値の単位・用法に null は指定できません。省略するか文字列を指定してください"
	errMsgDoseTextNotString   = "薬剤3値の単位・用法には文字列を指定してください"
)

// doseAmountField は 量（administration_dose_amount / prescription_total_amount）の wire 型。
type doseAmountField struct {
	set   bool
	value model.ExactNumeric
}

func (f *doseAmountField) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return apperrors.WrapInvalidInput(errMsgDoseAmountNull)
	}
	if err := json.Unmarshal(data, &f.value); err != nil {
		return apperrors.WrapInvalidInput(errMsgDoseAmountNotNumber)
	}
	if f.value.NaN {
		return apperrors.WrapInvalidInput(errMsgDoseAmountNaN)
	}
	f.set = true
	return nil
}

// toServiceInput は未送信を nil に変換する（create=NULL 保存 / PATCH=変更なし）。
func (f doseAmountField) toServiceInput() *model.ExactNumeric {
	if !f.set {
		return nil
	}
	value := f.value
	return &value
}

// doseTextField は 単位・用法（administration_dose_unit / administration_instructions /
// prescription_total_unit）の wire 型。
type doseTextField struct {
	set   bool
	value string
}

func (f *doseTextField) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return apperrors.WrapInvalidInput(errMsgDoseTextNull)
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return apperrors.WrapInvalidInput(errMsgDoseTextNotString)
	}
	f.set = true
	f.value = value
	return nil
}

func (f doseTextField) toServiceInput() *string {
	if !f.set {
		return nil
	}
	value := f.value
	return &value
}

type createTreatmentRequest struct {
	ItemType       string  `json:"item_type"       binding:"required,oneof=consultation procedure medicine other"`
	ConsultationID *uint64 `json:"consultation_id"`
	ProcedureID    *uint64 `json:"procedure_id"`
	MedicineID     *uint64 `json:"medicine_id"`
	InventoryID    *uint64 `json:"inventory_id"`
	UnitPrice      int64   `json:"unit_price"`
	Quantity       float64 `json:"quantity"`
	IsSelected     bool    `json:"is_selected"`
	Status         string  `json:"status"          binding:"omitempty,oneof=pending completed not_applicable"`
	Content        string  `json:"content"`
	Memo           string  `json:"memo"`
	AdminRoute     string  `json:"admin_route"`
	IsInsurance    bool    `json:"is_insurance"`
	DiscountRate   float64 `json:"discount_rate"`
	DiscountAmount int64   `json:"discount_amount"`
	SortOrder      int     `json:"sort_order"`
	// DoseDeviationReason は TASK-377: 上限内の著しい乖離/下限割れで必須。service が最終権限で再評価する。
	DoseDeviationReason string `json:"dose_deviation_reason"`
	// EMR-488: 薬剤3値（1回製剤量+単位・用法・処方総量+単位）。省略は NULL 保存（推測補完なし）。
	// 0 は明示 0。明示 null は 400（doseAmountField / doseTextField）。
	AdministrationDoseAmount   doseAmountField `json:"administration_dose_amount"`
	AdministrationDoseUnit     doseTextField   `json:"administration_dose_unit"`
	AdministrationInstructions doseTextField   `json:"administration_instructions"`
	PrescriptionTotalAmount    doseAmountField `json:"prescription_total_amount"`
	PrescriptionTotalUnit      doseTextField   `json:"prescription_total_unit"`
}

func (r *createTreatmentRequest) toServiceInput() *CreateTreatmentInput {
	return &CreateTreatmentInput{
		ItemType:            model.TreatmentItemType(r.ItemType),
		ConsultationID:      r.ConsultationID,
		ProcedureID:         r.ProcedureID,
		MedicineID:          r.MedicineID,
		InventoryID:         r.InventoryID,
		UnitPrice:           r.UnitPrice,
		Quantity:            r.Quantity,
		IsSelected:          r.IsSelected,
		Status:              r.Status,
		Content:             r.Content,
		Memo:                r.Memo,
		AdminRoute:          r.AdminRoute,
		IsInsurance:         r.IsInsurance,
		DiscountRate:        r.DiscountRate,
		DiscountAmount:      r.DiscountAmount,
		SortOrder:           r.SortOrder,
		DoseDeviationReason: r.DoseDeviationReason,

		AdministrationDoseAmount:   r.AdministrationDoseAmount.toServiceInput(),
		AdministrationDoseUnit:     r.AdministrationDoseUnit.toServiceInput(),
		AdministrationInstructions: r.AdministrationInstructions.toServiceInput(),
		PrescriptionTotalAmount:    r.PrescriptionTotalAmount.toServiceInput(),
		PrescriptionTotalUnit:      r.PrescriptionTotalUnit.toServiceInput(),
	}
}

type updateTreatmentRequest struct {
	ItemType       *string  `json:"item_type"       binding:"omitempty,oneof=consultation procedure medicine other"`
	ConsultationID *uint64  `json:"consultation_id"`
	ProcedureID    *uint64  `json:"procedure_id"`
	MedicineID     *uint64  `json:"medicine_id"`
	InventoryID    *uint64  `json:"inventory_id"`
	UnitPrice      *int64   `json:"unit_price"`
	Quantity       *float64 `json:"quantity"`
	IsSelected     *bool    `json:"is_selected"`
	Status         *string  `json:"status"          binding:"omitempty,oneof=pending completed not_applicable"`
	Content        *string  `json:"content"`
	Memo           *string  `json:"memo"`
	AdminRoute     *string  `json:"admin_route"`
	IsInsurance    *bool    `json:"is_insurance"`
	DiscountRate   *float64 `json:"discount_rate"`
	DiscountAmount *int64   `json:"discount_amount"`
	SortOrder      *int     `json:"sort_order"`
	// DoseDeviationReason は TASK-377: reason-required 時に必須。
	DoseDeviationReason *string `json:"dose_deviation_reason"`
	Version             *int    `json:"version"` // 楽観的ロック用
	// EMR-488: 薬剤3値。省略は変更なし（既存 PATCH 方式）。明示 null は 400（doseAmountField / doseTextField）。
	AdministrationDoseAmount   doseAmountField `json:"administration_dose_amount"`
	AdministrationDoseUnit     doseTextField   `json:"administration_dose_unit"`
	AdministrationInstructions doseTextField   `json:"administration_instructions"`
	PrescriptionTotalAmount    doseAmountField `json:"prescription_total_amount"`
	PrescriptionTotalUnit      doseTextField   `json:"prescription_total_unit"`
}

func (r *updateTreatmentRequest) toServiceInput() *UpdateTreatmentInput {
	input := &UpdateTreatmentInput{
		ConsultationID:      r.ConsultationID,
		ProcedureID:         r.ProcedureID,
		MedicineID:          r.MedicineID,
		InventoryID:         r.InventoryID,
		UnitPrice:           r.UnitPrice,
		Quantity:            r.Quantity,
		IsSelected:          r.IsSelected,
		Status:              r.Status,
		Content:             r.Content,
		Memo:                r.Memo,
		AdminRoute:          r.AdminRoute,
		IsInsurance:         r.IsInsurance,
		DiscountRate:        r.DiscountRate,
		DiscountAmount:      r.DiscountAmount,
		SortOrder:           r.SortOrder,
		DoseDeviationReason: r.DoseDeviationReason,
		Version:             r.Version,

		AdministrationDoseAmount:   r.AdministrationDoseAmount.toServiceInput(),
		AdministrationDoseUnit:     r.AdministrationDoseUnit.toServiceInput(),
		AdministrationInstructions: r.AdministrationInstructions.toServiceInput(),
		PrescriptionTotalAmount:    r.PrescriptionTotalAmount.toServiceInput(),
		PrescriptionTotalUnit:      r.PrescriptionTotalUnit.toServiceInput(),
	}
	if r.ItemType != nil {
		itemType := model.TreatmentItemType(*r.ItemType)
		input.ItemType = &itemType
	}
	return input
}

type bulkUpdateTreatmentsRequest struct {
	Treatments []bulkTreatmentItem `json:"treatments" binding:"required"`
}

func (r bulkUpdateTreatmentsRequest) toServiceInput() *BulkUpdateTreatmentsInput {
	items := make([]BulkTreatmentItem, 0, len(r.Treatments))
	for _, treatment := range r.Treatments {
		items = append(items, BulkTreatmentItem(treatment))
	}
	return &BulkUpdateTreatmentsInput{Treatments: items}
}

type bulkTreatmentItem struct {
	ID        uint64 `json:"id"         binding:"required"`
	SortOrder int    `json:"sort_order"`
}
