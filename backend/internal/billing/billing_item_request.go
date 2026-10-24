package billing

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

type unbilledItemsQuery struct {
	PetID string
}

func newUnbilledItemsQuery(values url.Values) unbilledItemsQuery {
	return unbilledItemsQuery{PetID: values.Get("pet_id")}
}

func (q unbilledItemsQuery) toPetID() (uint64, error) {
	if q.PetID == "" {
		return 0, apperrors.WrapInvalidInput("pet_id は必須です")
	}
	petID, err := strconv.ParseUint(q.PetID, 10, 64)
	if err != nil || petID == 0 {
		return 0, apperrors.WrapInvalidInput("pet_id の形式が不正です")
	}
	return petID, nil
}

// ---- accounting_request/response.go から B③ で分離移動（明細系のみ）----

// ---- 明細の10進 wire（EMR-505 / Q3 W1）----

// billingItemDecimalSpec は明細の10進 wire フィールド 1 つ分の受理条件。unit_price・quantity・discount_rate・tax_rate は
// JSON number の字句を exact に解釈する（float64 を経由しない）。指数表記・NaN・Infinity・符号 +・文字列は 400 で拒否する。
// 小数部の桁数と範囲は旧 DDL の受理範囲を維持する（001_init.sql:1940 unit_price bigint、1941 quantity numeric(10,1)、
// 1951 discount_rate numeric(5,2)、1943 tax_rate numeric(3,2)）。D10 で列が無制約 numeric になった（同 9398-9401）後も、
// wire の受理範囲を広げるには仕様判断が必要なので広げない。null と省略はどちらも「未指定」として扱う。
type billingItemDecimalSpec struct {
	name     string
	maxScale int32
	hasMax   bool
	maxValue int64
}

// billingItemDecimalKind は明細の10進フィールドごとの受理条件を型で束ねる。
type billingItemDecimalKind interface {
	spec() billingItemDecimalSpec
}

// unitPriceKind は unit_price（円。整数で 0 以上）。
type unitPriceKind struct{}

func (unitPriceKind) spec() billingItemDecimalSpec {
	return billingItemDecimalSpec{name: "unit_price", maxScale: 0}
}

// quantityKind は quantity（小数 1 桁、0 以上。正であることは service が検査する）。
type quantityKind struct{}

func (quantityKind) spec() billingItemDecimalSpec {
	return billingItemDecimalSpec{name: "quantity", maxScale: 1}
}

// discountRateKind は discount_rate（百分率。小数 2 桁、0 以上 100 以下）。
type discountRateKind struct{}

func (discountRateKind) spec() billingItemDecimalSpec {
	return billingItemDecimalSpec{name: "discount_rate", maxScale: 2, hasMax: true, maxValue: 100}
}

// taxRateKind は tax_rate（割合。0.10 が 10%。小数 2 桁、0 以上 1 以下）。
type taxRateKind struct{}

func (taxRateKind) spec() billingItemDecimalSpec {
	return billingItemDecimalSpec{name: "tax_rate", maxScale: 2, hasMax: true, maxValue: 1}
}

// billingItemDecimal は明細の10進 wire 値。set は値が指定されたこと（null・省略は false）。
type billingItemDecimal[K billingItemDecimalKind] struct {
	set   bool
	value model.ExactNumeric
}

type (
	unitPriceWire    = billingItemDecimal[unitPriceKind]
	quantityWire     = billingItemDecimal[quantityKind]
	discountRateWire = billingItemDecimal[discountRateKind]
	taxRateWire      = billingItemDecimal[taxRateKind]
)

// UnmarshalJSON は JSON number の字句を exact に解釈する。K ごとの受理条件は spec が定める。
func (d *billingItemDecimal[K]) UnmarshalJSON(data []byte) error {
	var kind K
	value, set, err := decodeBillingItemDecimal(data, kind.spec())
	if err != nil {
		return err
	}
	*d = billingItemDecimal[K]{set: set, value: value}
	return nil
}

// orZero は create 用。未指定は 0 として返す（NULL を 0 に潰すのではなく、省略時の既存契約が 0）。
func (d *billingItemDecimal[K]) orZero() model.ExactNumeric {
	if d == nil || !d.set {
		return model.ExactNumericFromInt64(0)
	}
	return d.value
}

// optional は未指定を nil で返す。update は nil を「変更なし」、create の tax_rate は nil を service の既定判定に渡す。
func (d *billingItemDecimal[K]) optional() *model.ExactNumeric {
	if d == nil || !d.set {
		return nil
	}
	value := d.value
	return &value
}

// decodeBillingItemDecimal は wire の字句を検査して ExactNumeric にする。null は未指定（set=false）。
func decodeBillingItemDecimal(data []byte, spec billingItemDecimalSpec) (model.ExactNumeric, bool, error) {
	if string(data) == "null" {
		return model.ExactNumeric{}, false, nil
	}
	if len(data) == 0 || data[0] == '"' {
		return model.ExactNumeric{}, false, apperrors.WrapInvalidInput(spec.name + " は JSON の数値で指定してください（文字列は使えません）")
	}
	value, err := model.ParseExactNumeric(string(data))
	if err != nil {
		return model.ExactNumeric{}, false, apperrors.WrapInvalidInput(spec.name + " には有限の10進数を指定してください（指数表記・NaN・Infinity・符号 + は使えません）")
	}
	if _, err := value.ScaledInt64(spec.maxScale); err != nil {
		return model.ExactNumeric{}, false, billingItemScaleError(spec, err)
	}
	if cmp, ok := value.Cmp(model.ExactNumericFromInt64(0)); !ok || cmp < 0 {
		return model.ExactNumeric{}, false, apperrors.WrapInvalidInput(spec.name + " は0以上で指定してください")
	}
	if spec.hasMax {
		if cmp, ok := value.Cmp(model.ExactNumericFromInt64(spec.maxValue)); !ok || cmp > 0 {
			return model.ExactNumeric{}, false, apperrors.WrapInvalidInput(fmt.Sprintf("%s は%d以下で指定してください", spec.name, spec.maxValue))
		}
	}
	return value, true, nil
}

// billingItemScaleError は小数部の桁数・桁あふれの 400 を作る。
func billingItemScaleError(spec billingItemDecimalSpec, err error) error {
	if errors.Is(err, model.ErrExactNumericOutOfRange) {
		return apperrors.WrapInvalidInput(spec.name + " の値が大きすぎます")
	}
	if spec.maxScale == 0 {
		return apperrors.WrapInvalidInput(spec.name + " は整数で指定してください")
	}
	return apperrors.WrapInvalidInput(fmt.Sprintf("%s の小数部は%d桁以内で指定してください", spec.name, spec.maxScale))
}

// createBillingItemRequest は明細作成リクエスト。
type createBillingItemRequest struct {
	BillingID             uint64           `json:"billing_id" binding:"required"`
	Category              string           `json:"category"  binding:"omitempty,oneof=examination test procedure surgery medicine food goods other vaccine trimming hotel training house_call"`
	Name                  string           `json:"name"      binding:"required,max=255"`
	UnitPrice             unitPriceWire    `json:"unit_price"`
	Quantity              quantityWire     `json:"quantity"`
	DiscountRate          discountRateWire `json:"discount_rate"`
	DiscountAmount        int64            `json:"discount_amount" binding:"min=0"`
	TaxType               string           `json:"tax_type"  binding:"omitempty,oneof=included excluded exempt"`
	TaxRate               taxRateWire      `json:"tax_rate"`
	IsInsuranceApplicable bool             `json:"is_insurance_applicable"`
	Source                string           `json:"source"    binding:"omitempty,oneof=medical_record manual hospitalization trimming"`
	OtherReason           *string          `json:"other_reason" binding:"omitempty,max=500"`
	TreatmentID           *uint64          `json:"treatment_id"`
	VaccinationID         *uint64          `json:"vaccination_id"`
	ExamID                *uint64          `json:"exam_id"`
	AppointmentID         *uint64          `json:"appointment_id"`
	TrimmingCourseID      *uint64          `json:"trimming_course_id"`
	TrimmingOptionID      *uint64          `json:"trimming_option_id"`
	MerchandiseItemID     *uint64          `json:"merchandise_item_id"`
	SortOrder             int              `json:"sort_order"`
	// #115 / BUG-463: 締め後編集理由（レジ締め済み期間の明細を変更する場合は必須）
	PostCloseReason *string `json:"post_close_reason" binding:"omitempty,max=500"`
}

// updateBillingItemRequest は明細更新リクエスト（nil = 未指定）。
type updateBillingItemRequest struct {
	// EMR-229: 社販処理の目安として項目名（年月追記等）を会計画面から更新できるようにする。
	Name                  *string           `json:"name" binding:"omitempty,max=255"`
	UnitPrice             *unitPriceWire    `json:"unit_price"`
	Quantity              *quantityWire     `json:"quantity"`
	DiscountRate          *discountRateWire `json:"discount_rate"`
	DiscountAmount        *int64            `json:"discount_amount" binding:"omitempty,min=0"`
	TaxType               *string           `json:"tax_type"  binding:"omitempty,oneof=included excluded exempt"`
	TaxRate               *taxRateWire      `json:"tax_rate"`
	IsInsuranceApplicable *bool             `json:"is_insurance_applicable"`
	// #115 / BUG-463: 締め後編集理由（レジ締め済み期間の明細を変更する場合は必須）
	PostCloseReason *string `json:"post_close_reason" binding:"omitempty,max=500"`
}

// deleteBillingItemRequest は明細削除の任意 body。
// 締め後削除時のみ post_close_reason を送る（BUG-463 residual）。
type deleteBillingItemRequest struct {
	PostCloseReason *string `json:"post_close_reason" binding:"omitempty,max=500"`
}

func (r *createBillingItemRequest) toServiceInput(clinicID uint64) *CreateBillingItemInput {
	return &CreateBillingItemInput{
		ClinicID:              clinicID,
		BillingID:             r.BillingID,
		Category:              r.Category,
		Name:                  r.Name,
		UnitPrice:             r.UnitPrice.orZero(),
		Quantity:              r.Quantity.orZero(),
		DiscountRate:          r.DiscountRate.orZero(),
		DiscountAmount:        r.DiscountAmount,
		TaxType:               r.TaxType,
		TaxRate:               r.TaxRate.optional(),
		IsInsuranceApplicable: r.IsInsuranceApplicable,
		Source:                r.Source,
		OtherReason:           r.OtherReason,
		TreatmentID:           r.TreatmentID,
		VaccinationID:         r.VaccinationID,
		ExamID:                r.ExamID,
		AppointmentID:         r.AppointmentID,
		TrimmingCourseID:      r.TrimmingCourseID,
		TrimmingOptionID:      r.TrimmingOptionID,
		MerchandiseItemID:     r.MerchandiseItemID,
		SortOrder:             r.SortOrder,
		PostCloseReason:       r.PostCloseReason,
	}
}

func (r updateBillingItemRequest) toServiceInput() (*UpdateBillingItemInput, error) {
	var taxType *model.TaxType
	if r.TaxType != nil {
		if err := sharedkernel.ValidateTaxType(*r.TaxType); err != nil {
			return nil, err
		}
		t := model.TaxType(*r.TaxType)
		taxType = &t
	}

	return &UpdateBillingItemInput{
		Name:                  r.Name,
		UnitPrice:             r.UnitPrice.optional(),
		Quantity:              r.Quantity.optional(),
		DiscountRate:          r.DiscountRate.optional(),
		DiscountAmount:        r.DiscountAmount,
		TaxType:               taxType,
		TaxRate:               r.TaxRate.optional(),
		IsInsuranceApplicable: r.IsInsuranceApplicable,
		PostCloseReason:       r.PostCloseReason,
	}, nil
}

type BillingItemResponse struct {
	ID                    uint64              `json:"id"`
	BillingID             uint64              `json:"billing_id"`
	Category              string              `json:"category"`
	Name                  string              `json:"name"`
	UnitPrice             model.ExactNumeric  `json:"unit_price"`
	Quantity              model.ExactNumeric  `json:"quantity"`
	DiscountRate          model.ExactNumeric  `json:"discount_rate"`
	DiscountAmount        int64               `json:"discount_amount"`
	Subtotal              *int64              `json:"subtotal"`
	TaxType               string              `json:"tax_type"`
	TaxRate               *model.ExactNumeric `json:"tax_rate"`
	TaxAmount             *int64              `json:"tax_amount"`
	IsInsuranceApplicable *bool               `json:"is_insurance_applicable"`
	Source                *string             `json:"source"`
	OtherReason           *string             `json:"other_reason,omitempty" binding:"omitempty,max=500"`
	// MedicalRecordID は未請求候補など、treatment 由来の親カルテ（DB 列ではない仮想値）。
	MedicalRecordID *uint64 `json:"medical_record_id,omitempty"`
	// ServiceDate は未請求候補の発生日（DB 列ではない仮想値。EMR-246 当日/過去区分用）。
	ServiceDate       *time.Time `json:"service_date,omitempty"`
	MerchandiseItemID *uint64    `json:"merchandise_item_id,omitempty"`
	SortOrder         int        `json:"sort_order"`
	CreatedAt         time.Time  `json:"created_at"`
}

func ToBillingItemResponse(item *model.BillingItem) BillingItemResponse {
	// subtotal は明細値引き後の税抜額（単価 × 数量の 1 円未満四捨五入 − 明細の値引き）で、会計全体の値引きは含まない。
	// 算定できない行（NULL・非有限の値など）は subtotal・tax_amount を null にし、0 で埋めない。
	// tax_amount は表示用の行別税額で、請求書の税額の正本ではない（EMR-505 / D5）。
	var subtotal, taxAmount *int64
	if amounts, err := CalculateLineDisplayAmounts(CalcLineFromBillingItem(item)); err == nil {
		subtotal = &amounts.Subtotal
		taxAmount = amounts.TaxAmount
	}
	// NULL の source は manual などに補わず null で返す（EMR-505 / Q12）。
	var source *string
	if item.Source != nil {
		value := string(*item.Source)
		source = &value
	}
	// sort_order は表示順であり業務値ではないため、NULL は 0 として返す（Q12。変換はこの mapper に限る）。
	sortOrder := 0
	if item.SortOrder != nil {
		sortOrder = *item.SortOrder
	}
	return BillingItemResponse{
		ID:                    item.ID,
		BillingID:             item.BillingID,
		Category:              string(item.Category),
		Name:                  item.Name,
		UnitPrice:             item.UnitPrice,
		Quantity:              item.Quantity,
		DiscountRate:          item.DiscountRate,
		DiscountAmount:        item.DiscountAmount,
		Subtotal:              subtotal,
		TaxType:               string(item.TaxType),
		TaxRate:               item.TaxRate,
		TaxAmount:             taxAmount,
		IsInsuranceApplicable: item.IsInsuranceApplicable,
		Source:                source,
		OtherReason:           item.OtherReason,
		MedicalRecordID:       item.MedicalRecordID,
		ServiceDate:           item.ServiceDate,
		MerchandiseItemID:     item.MerchandiseItemID,
		SortOrder:             sortOrder,
		CreatedAt:             httpapi.LocalTime(item.CreatedAt),
	}
}
