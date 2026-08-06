package billing

import (
	"net/url"
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type listAccountingQuery struct {
	PetID     string
	OwnerID   string
	Status    string
	StartDate string
	EndDate   string
	Search    string
}

func newListAccountingQuery(values url.Values) listAccountingQuery {
	return listAccountingQuery{
		PetID:     values.Get("pet_id"),
		OwnerID:   values.Get("owner_id"),
		Status:    values.Get("status"),
		StartDate: values.Get("start_date"),
		EndDate:   values.Get("end_date"),
		Search:    values.Get("search"),
	}
}

type listAccountingFilters struct {
	PetID     *uint64
	OwnerID   *uint64
	Status    *string
	StartDate *string
	EndDate   *string
	Search    string
}

func (q *listAccountingQuery) toServiceFilters() (listAccountingFilters, error) {
	petID, err := parseOptionalUintQueryFilter(q.PetID, "pet_id")
	if err != nil {
		return listAccountingFilters{}, err
	}
	ownerID, err := parseOptionalUintQueryFilter(q.OwnerID, "owner_id")
	if err != nil {
		return listAccountingFilters{}, err
	}
	startDate, err := parseOptionalDateQueryFilter(q.StartDate, "start_date")
	if err != nil {
		return listAccountingFilters{}, err
	}
	endDate, err := parseOptionalDateQueryFilter(q.EndDate, "end_date")
	if err != nil {
		return listAccountingFilters{}, err
	}
	return listAccountingFilters{
		PetID:     petID,
		OwnerID:   ownerID,
		Status:    optionalStringQueryFilter(q.Status),
		StartDate: startDate,
		EndDate:   endDate,
		Search:    q.Search,
	}, nil
}

type listUnpaidBillingsQuery struct {
	StartDate string
	EndDate   string
	GroupBy   string
}

func newListUnpaidBillingsQuery(values url.Values) listUnpaidBillingsQuery {
	return listUnpaidBillingsQuery{
		StartDate: values.Get("start_date"),
		EndDate:   values.Get("end_date"),
		GroupBy:   values.Get("group_by"),
	}
}

type listUnpaidBillingsFilters struct {
	StartDate string
	EndDate   string
	GroupBy   string
}

type dailySummaryQuery struct {
	Date string
}

func newDailySummaryQuery(values url.Values) dailySummaryQuery {
	return dailySummaryQuery{Date: values.Get("date")}
}

// #114: 月次未納繰越集計クエリ
type monthlyUnpaidQuery struct {
	Year  string
	Month string
}

func newMonthlyUnpaidQuery(values url.Values) monthlyUnpaidQuery {
	return monthlyUnpaidQuery{
		Year:  values.Get("year"),
		Month: values.Get("month"),
	}
}

// parse は year/month を検証して int に変換する。
func (q monthlyUnpaidQuery) parse() (year, month int, err error) {
	if q.Year == "" {
		return 0, 0, apperrors.WrapInvalidInput("year is required")
	}
	if q.Month == "" {
		return 0, 0, apperrors.WrapInvalidInput("month is required")
	}
	year, err = strconv.Atoi(q.Year)
	if err != nil || year < 2000 || year > 2100 {
		return 0, 0, apperrors.WrapInvalidInput("year must be a valid year (2000-2100)")
	}
	month, err = strconv.Atoi(q.Month)
	if err != nil || month < 1 || month > 12 {
		return 0, 0, apperrors.WrapInvalidInput("month must be between 1 and 12")
	}
	return year, month, nil
}

// toServiceFilters は #120: start_date/end_date を必須パラメータとして検証する。
// どちらか欠けた場合は ErrInvalidInput を返す（フォールバックなし）。
func (q listUnpaidBillingsQuery) toServiceFilters() (listUnpaidBillingsFilters, error) {
	if q.StartDate == "" {
		return listUnpaidBillingsFilters{}, apperrors.WrapInvalidInput("start_date is required")
	}
	if q.EndDate == "" {
		return listUnpaidBillingsFilters{}, apperrors.WrapInvalidInput("end_date is required")
	}
	if _, err := parseOptionalDateQueryFilter(q.StartDate, "start_date"); err != nil {
		return listUnpaidBillingsFilters{}, err
	}
	if _, err := parseOptionalDateQueryFilter(q.EndDate, "end_date"); err != nil {
		return listUnpaidBillingsFilters{}, err
	}

	groupBy := q.GroupBy
	if groupBy == "" {
		groupBy = "owner"
	}

	return listUnpaidBillingsFilters{
		StartDate: q.StartDate,
		EndDate:   q.EndDate,
		GroupBy:   groupBy,
	}, nil
}

// createAccountingRequest は会計作成リクエスト。
type createAccountingRequest struct {
	MedicalRecordID   *uint64    `json:"medical_record_id"`
	HospitalizationID *uint64    `json:"hospitalization_id"`
	OwnerID           *uint64    `json:"owner_id"`
	PetID             *uint64    `json:"pet_id"`
	Subtotal          int64      `json:"subtotal"      binding:"min=0"`
	TaxTotal          int64      `json:"tax_total"     binding:"min=0"`
	TotalAmount       int64      `json:"total_amount"  binding:"min=0"`
	HasInsurance      bool       `json:"has_insurance"`
	Status            string     `json:"status"        binding:"omitempty,oneof=waiting pending completed cancelled"`
	ScheduledDate     time.Time  `json:"scheduled_date" binding:"required"`
	CompletedAt       *time.Time `json:"completed_at"`
	Memo              string     `json:"memo"`
}

// completeAccountingItemRequest は BUG-018 complete command の明細1行。
type completeAccountingItemRequest struct {
	Category              string  `json:"category"  binding:"omitempty,oneof=examination test procedure surgery medicine food goods other vaccine trimming hotel training"`
	Name                  string  `json:"name"      binding:"required"`
	UnitPrice             int64   `json:"unit_price" binding:"min=0"`
	Quantity              float64 `json:"quantity"   binding:"required,gt=0"`
	DiscountRate          float64 `json:"discount_rate" binding:"min=0,max=100"`
	DiscountAmount        int64   `json:"discount_amount" binding:"min=0"`
	TaxType               string  `json:"tax_type"  binding:"omitempty,oneof=included excluded exempt"`
	TaxRate               float64 `json:"tax_rate"`
	IsInsuranceApplicable bool    `json:"is_insurance_applicable"`
	Source                string  `json:"source"    binding:"omitempty,oneof=medical_record manual hospitalization trimming"`
	OtherReason           *string `json:"other_reason"`
	MerchandiseItemID     *uint64 `json:"merchandise_item_id"`
	TreatmentID           *uint64 `json:"treatment_id"`
	VaccinationID         *uint64 `json:"vaccination_id"`
	AppointmentID         *uint64 `json:"appointment_id"`
	TrimmingCourseID      *uint64 `json:"trimming_course_id"`
	TrimmingOptionID      *uint64 `json:"trimming_option_id"`
	SortOrder             int     `json:"sort_order"`
}

// completeAccountingRequest は BUG-018 POST /accountings/complete の body。
// client total は受け取らず server が items から再計算する。
type completeAccountingRequest struct {
	MedicalRecordID   *uint64                         `json:"medical_record_id"`
	HospitalizationID *uint64                         `json:"hospitalization_id"`
	OwnerID           *uint64                         `json:"owner_id"`
	PetID             *uint64                         `json:"pet_id"`
	ScheduledDate     time.Time                       `json:"scheduled_date" binding:"required"`
	Memo              string                          `json:"memo"`
	HasInsurance      bool                            `json:"has_insurance"`
	InsuranceRatio    *float64                        `json:"insurance_ratio"`
	InsuranceName     *string                         `json:"insurance_name"`
	InsuranceAmount   *int64                          `json:"insurance_amount"`
	DiscountAmount    *int64                          `json:"discount_amount"`
	Items             []completeAccountingItemRequest `json:"items" binding:"required,min=1,dive"`
	PaymentSplits     []paymentSplitRequest           `json:"payment_splits" binding:"max=50,dive"`
	PostCloseReason   *string                         `json:"post_close_reason"`
}

func (r *completeAccountingRequest) toServiceInput(clinicID, staffID uint64, idempotencyKey string) *CompleteAccountingInput {
	items := make([]CompleteAccountingItemInput, 0, len(r.Items))
	for _, it := range r.Items {
		items = append(items, CompleteAccountingItemInput{
			Category:              it.Category,
			Name:                  it.Name,
			UnitPrice:             it.UnitPrice,
			Quantity:              it.Quantity,
			DiscountRate:          it.DiscountRate,
			DiscountAmount:        it.DiscountAmount,
			TaxType:               it.TaxType,
			TaxRate:               it.TaxRate,
			IsInsuranceApplicable: it.IsInsuranceApplicable,
			Source:                it.Source,
			OtherReason:           it.OtherReason,
			MerchandiseItemID:     it.MerchandiseItemID,
			TreatmentID:           it.TreatmentID,
			VaccinationID:         it.VaccinationID,
			AppointmentID:         it.AppointmentID,
			TrimmingCourseID:      it.TrimmingCourseID,
			TrimmingOptionID:      it.TrimmingOptionID,
			SortOrder:             it.SortOrder,
		})
	}
	return &CompleteAccountingInput{
		ClinicID:          clinicID,
		StaffID:           &staffID,
		IdempotencyKey:    idempotencyKey,
		MedicalRecordID:   r.MedicalRecordID,
		HospitalizationID: r.HospitalizationID,
		OwnerID:           r.OwnerID,
		PetID:             r.PetID,
		ScheduledDate:     r.ScheduledDate,
		Memo:              r.Memo,
		HasInsurance:      r.HasInsurance,
		InsuranceRatio:    r.InsuranceRatio,
		InsuranceName:     r.InsuranceName,
		InsuranceAmount:   r.InsuranceAmount,
		DiscountAmount:    r.DiscountAmount,
		Items:             items,
		PaymentSplits:     toPaymentSplitInputs(r.PaymentSplits),
		PostCloseReason:   r.PostCloseReason,
	}
}

func (r *createAccountingRequest) toServiceInput(clinicID uint64) *CreateAccountingInput {
	return &CreateAccountingInput{
		ClinicID:          clinicID,
		MedicalRecordID:   r.MedicalRecordID,
		HospitalizationID: r.HospitalizationID,
		OwnerID:           r.OwnerID,
		PetID:             r.PetID,
		Subtotal:          r.Subtotal,
		TaxTotal:          r.TaxTotal,
		TotalAmount:       r.TotalAmount,
		HasInsurance:      r.HasInsurance,
		Status:            model.BillingStatus(r.Status),
		ScheduledDate:     r.ScheduledDate,
		CompletedAt:       r.CompletedAt,
		Memo:              r.Memo,
	}
}

// paymentSplitRequest は混在会計の支払い内訳1行リクエスト。
type paymentSplitRequest struct {
	Method          string  `json:"method"           binding:"required,oneof=cash credit_card electronic_money bank_transfer"`
	PaymentMethodID *uint64 `json:"payment_method_id"`
	Amount          int64   `json:"amount"           binding:"required,min=1"`
	ReceivedAmount  int64   `json:"received_amount"`
	ChangeAmount    int64   `json:"change_amount"    binding:"min=0"` // #119: 非負のみ強制（必須ではない。お釣り整合性は service の validatePaymentSplits が検証）
	// ChangeOverride は #188: お釣り直接上書きモード。true の場合、レジ実機の誤差吸収のため
	// service の validatePaymentSplits が change == received - amount 整合検証を緩和する（下限ガードは維持）。
	ChangeOverride bool `json:"change_override"`
}

func (r paymentSplitRequest) toServiceInput() PaymentSplitInput {
	return PaymentSplitInput{
		Method:          model.PaymentMethod(r.Method),
		PaymentMethodID: r.PaymentMethodID,
		Amount:          r.Amount,
		ReceivedAmount:  r.ReceivedAmount,
		ChangeAmount:    r.ChangeAmount,
		ChangeOverride:  r.ChangeOverride,
	}
}

func toPaymentSplitInputs(reqs []paymentSplitRequest) []PaymentSplitInput {
	if len(reqs) == 0 {
		return nil
	}
	out := make([]PaymentSplitInput, 0, len(reqs))
	for _, r := range reqs {
		out = append(out, r.toServiceInput())
	}
	return out
}

// correctCreditPaymentRequest は確定済み会計のクレジット（カード）金額の確定後訂正リクエスト（#189）。
// 訂正対象はカード系手段1件。理由は必須。受領額・お釣りはカード内訳では持たないため受け取らない。
type correctCreditPaymentRequest struct {
	Method string `json:"method" binding:"required,oneof=credit_card electronic_money"`
	Amount int64  `json:"amount" binding:"required,min=1"`
	Reason string `json:"reason" binding:"required"`
	Memo   string `json:"memo"`
}

func (r *correctCreditPaymentRequest) toServiceInput(id, clinicID, staffID uint64) *CorrectCreditPaymentInput {
	return &CorrectCreditPaymentInput{
		ClinicID:  clinicID,
		BillingID: id,
		StaffID:   &staffID,
		Method:    model.PaymentMethod(r.Method),
		Amount:    r.Amount,
		Reason:    r.Reason,
		Memo:      r.Memo,
	}
}

// updateAccountingRequest は会計更新リクエスト。
// nil フィールドは更新しない（GORM ゼロ値スキップ問題を回避するためポインタ型を使用）。
// Payment フィールドが含まれている場合、会計完了時に Payment を同時 upsert する。
type updateAccountingRequest struct {
	MedicalRecordID   *uint64    `json:"medical_record_id"`
	HospitalizationID *uint64    `json:"hospitalization_id"`
	OwnerID           *uint64    `json:"owner_id"`
	PetID             *uint64    `json:"pet_id"`
	Subtotal          *int64     `json:"subtotal"`
	TaxTotal          *int64     `json:"tax_total"`
	TotalAmount       *int64     `json:"total_amount"`
	HasInsurance      *bool      `json:"has_insurance"`
	Status            *string    `json:"status"        binding:"omitempty,oneof=waiting pending completed cancelled"`
	ScheduledDate     *time.Time `json:"scheduled_date"`
	CompletedAt       *time.Time `json:"completed_at"`
	Memo              *string    `json:"memo"`
	// Payment フィールド（会計完了時に同時送信される）
	PaymentMethod   *string  `json:"payment_method"  binding:"omitempty,oneof=cash credit_card electronic_money bank_transfer"`
	InsuranceRatio  *float64 `json:"insurance_ratio"`
	InsuranceName   *string  `json:"insurance_name"`
	InsuranceAmount *int64   `json:"insurance_amount"`
	DiscountAmount  *int64   `json:"discount_amount"`
	BillingAmount   *int64   `json:"billing_amount"`
	ReceivedAmount  *int64   `json:"received_amount"`
	ChangeAmount    *int64   `json:"change_amount"`
	// PaymentSplits: 混在支払い内訳（nil = 従来単一支払い互換）
	PaymentSplits []paymentSplitRequest `json:"payment_splits" binding:"max=50,dive"`
	// #115: 締め後編集理由（レジ締め済み期間の会計を編集する場合は必須）
	PostCloseReason *string `json:"post_close_reason"`
}

func (r *updateAccountingRequest) toServiceInput(id, clinicID, staffID uint64) *UpdateAccountingInput {
	return &UpdateAccountingInput{
		ID:                id,
		ClinicID:          clinicID,
		StaffID:           &staffID,
		MedicalRecordID:   r.MedicalRecordID,
		HospitalizationID: r.HospitalizationID,
		OwnerID:           r.OwnerID,
		PetID:             r.PetID,
		Subtotal:          r.Subtotal,
		TaxTotal:          r.TaxTotal,
		TotalAmount:       r.TotalAmount,
		HasInsurance:      r.HasInsurance,
		ScheduledDate:     r.ScheduledDate,
		CompletedAt:       r.CompletedAt,
		Memo:              r.Memo,
		InsuranceRatio:    r.InsuranceRatio,
		InsuranceName:     r.InsuranceName,
		InsuranceAmount:   r.InsuranceAmount,
		DiscountAmount:    r.DiscountAmount,
		BillingAmount:     r.BillingAmount,
		ReceivedAmount:    r.ReceivedAmount,
		ChangeAmount:      r.ChangeAmount,
		Status:            billingStatusPtr(r.Status),
		PaymentMethod:     paymentMethodPtr(r.PaymentMethod),
		PaymentSplits:     toPaymentSplitInputs(r.PaymentSplits),
		PostCloseReason:   r.PostCloseReason,
	}
}

// billingStatusPtr は *string を *model.BillingStatus に変換する。nil の場合は nil を返す。
func billingStatusPtr(s *string) *model.BillingStatus {
	if s == nil {
		return nil
	}
	v := model.BillingStatus(*s)
	return &v
}

// paymentMethodPtr は *string を *model.PaymentMethod に変換する。nil の場合は nil を返す。
func paymentMethodPtr(s *string) *model.PaymentMethod {
	if s == nil {
		return nil
	}
	v := model.PaymentMethod(*s)
	return &v
}
