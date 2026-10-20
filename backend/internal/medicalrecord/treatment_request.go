package medicalrecord

import (
	"encoding/json"
	"math/big"
	"strconv"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
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
	// クォート付き文字列（例: "administration_dose_amount":"1.5"）は OpenAPI の number 契約に
	// 反するため、json.Number への Unmarshal に成功しても 400 で拒否する（Codex review P2）。
	if len(data) > 0 && data[0] == '"' {
		return apperrors.WrapInvalidInput(errMsgDoseAmountNotNumber)
	}
	// json.Number 経由で受理し、big.Rat で厳密に解析する（Codex review P2）。
	// pgtype の素の UnmarshalJSON は指数表記を拒否し、ScanScientific は float64 経由で
	// 精度が落ちる（例: 1.00000000000000001e-6 が 0.000001 に丸まり、送信値と異なる臨床量を
	// 保存し得る）。そのため仮数部×10^exp を big.Rat で厳密に計算し、numeric(12,6) 適合
	// （小数6桁以内・絶対値 ≤ 999999.999999）を満たさない入力は 400 で拒否する。
	var num json.Number
	if err := json.Unmarshal(data, &num); err != nil {
		return apperrors.WrapInvalidInput(errMsgDoseAmountNotNumber)
	}
	rat, exact := parseDecimalWithExponent(num.String())
	if !exact {
		return apperrors.WrapInvalidInput(errMsgDoseAmountNotNumber)
	}
	scaled := new(big.Rat).Mul(rat, big.NewRat(1000000, 1))
	if !scaled.IsInt() {
		return apperrors.WrapInvalidInput(errMsgDoseAmountRange)
	}
	limit := big.NewInt(999999999999) // 999999.999999 × 10^6
	if scaled.Num().CmpAbs(limit) > 0 {
		return apperrors.WrapInvalidInput(errMsgDoseAmountRange)
	}
	// numeric(12,6) に正規化して保存（Int = 6桁スケールの整数, Exp = -6）
	f.set = true
	f.value = model.ExactNumeric{pgtype.Numeric{
		Int:   new(big.Int).Set(scaled.Num()),
		Exp:   -6,
		Valid: true,
	}}
	return nil
}

// parseDecimalWithExponent は 10進リテラル（任意の指数部付き）を big.Rat へ厳密に変換する。
// "2.5" / "1e-06" / "-1.5E+3" 等を受理し、形式が不正な場合は ok=false を返す。
func parseDecimalWithExponent(text string) (*big.Rat, bool) {
	mantissa, expPart, hasExp := strings.Cut(text, "e")
	if !hasExp {
		mantissa, expPart, hasExp = strings.Cut(text, "E")
	}
	exp := 0
	if hasExp {
		var err error
		exp, err = strconv.Atoi(expPart)
		// 指数が極端に大きい/小さい入力（例: 1e1000000000）による big.Int 巨大計算の
		// DoS を防ぐため、numeric(12,6) の取り得る範囲を大きく外れる指数は解析前に拒否する
		// （Codex review P1 / CodeRabbit Major）。
		if err != nil || exp > 40 || exp < -40 {
			return nil, false
		}
	}
	mantissaRat, ok := new(big.Rat).SetString(mantissa)
	if !ok {
		return nil, false
	}
	if exp == 0 {
		return mantissaRat, true
	}
	factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(absInt(exp))), nil)
	if exp > 0 {
		return mantissaRat.Mul(mantissaRat, new(big.Rat).SetInt(factor)), true
	}
	return mantissaRat.Quo(mantissaRat, new(big.Rat).SetInt(factor)), true
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
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
