package handler

import (
	"net/url"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

type listAccountingQuery struct {
	PetID     string
	OwnerID   string
	Status    string
	StartDate string
	EndDate   string
}

func newListAccountingQuery(values url.Values) listAccountingQuery {
	return listAccountingQuery{
		PetID:     values.Get("pet_id"),
		OwnerID:   values.Get("owner_id"),
		Status:    values.Get("status"),
		StartDate: values.Get("start_date"),
		EndDate:   values.Get("end_date"),
	}
}

type listAccountingFilters struct {
	PetID     *uint64
	OwnerID   *uint64
	Status    *string
	StartDate *string
	EndDate   *string
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

func (r *createAccountingRequest) toServiceInput(clinicID uint64) *service.CreateAccountingInput {
	return &service.CreateAccountingInput{
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
	Method          string  `json:"method"           binding:"required,oneof=cash credit_card electronic_money"`
	PaymentMethodID *uint64 `json:"payment_method_id"`
	Amount          int64   `json:"amount"           binding:"required,min=1"`
	ReceivedAmount  int64   `json:"received_amount"`
	ChangeAmount    int64   `json:"change_amount"    binding:"min=0"` // #119: required (non-negative)
}

func (r paymentSplitRequest) toServiceInput() service.PaymentSplitInput {
	return service.PaymentSplitInput{
		Method:          model.PaymentMethod(r.Method),
		PaymentMethodID: r.PaymentMethodID,
		Amount:          r.Amount,
		ReceivedAmount:  r.ReceivedAmount,
		ChangeAmount:    r.ChangeAmount,
	}
}

func toPaymentSplitInputs(reqs []paymentSplitRequest) []service.PaymentSplitInput {
	if len(reqs) == 0 {
		return nil
	}
	out := make([]service.PaymentSplitInput, 0, len(reqs))
	for _, r := range reqs {
		out = append(out, r.toServiceInput())
	}
	return out
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
	PaymentMethod   *string  `json:"payment_method"  binding:"omitempty,oneof=cash credit_card electronic_money"`
	InsuranceRatio  *float64 `json:"insurance_ratio"`
	InsuranceName   *string  `json:"insurance_name"`
	InsuranceAmount *int64   `json:"insurance_amount"`
	DiscountAmount  *int64   `json:"discount_amount"`
	BillingAmount   *int64   `json:"billing_amount"`
	ReceivedAmount  *int64   `json:"received_amount"`
	ChangeAmount    *int64   `json:"change_amount"`
	// PaymentSplits: 混在支払い内訳（nil = 従来単一支払い互換）
	PaymentSplits []paymentSplitRequest `json:"payment_splits" binding:"max=50,dive"`
}

func (r *updateAccountingRequest) toServiceInput(id, clinicID, staffID uint64) *service.UpdateAccountingInput {
	return &service.UpdateAccountingInput{
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

// createBillingItemRequest は明細作成リクエスト。
type createBillingItemRequest struct {
	BillingID             uint64  `json:"billing_id" binding:"required"`
	Category              string  `json:"category"  binding:"omitempty,oneof=examination test procedure surgery medicine food goods other vaccine trimming hotel training"`
	Name                  string  `json:"name"      binding:"required"`
	UnitPrice             int64   `json:"unit_price" binding:"min=0"`
	Quantity              float64 `json:"quantity"   binding:"min=0"`
	DiscountRate          float64 `json:"discount_rate" binding:"min=0,max=100"`
	DiscountAmount        int64   `json:"discount_amount" binding:"min=0"`
	TaxType               string  `json:"tax_type"  binding:"omitempty,oneof=included excluded exempt"`
	TaxRate               float64 `json:"tax_rate"`
	IsInsuranceApplicable bool    `json:"is_insurance_applicable"`
	Source                string  `json:"source"    binding:"omitempty,oneof=medical_record manual hospitalization trimming"`
	TreatmentID           *uint64 `json:"treatment_id"`
	AppointmentID         *uint64 `json:"appointment_id"`
	TrimmingCourseID      *uint64 `json:"trimming_course_id"`
	TrimmingOptionID      *uint64 `json:"trimming_option_id"`
	MerchandiseItemID     *uint64 `json:"merchandise_item_id"`
	SortOrder             int     `json:"sort_order"`
}

func (r *createBillingItemRequest) toServiceInput(clinicID uint64) *service.CreateBillingItemInput {
	return &service.CreateBillingItemInput{
		ClinicID:              clinicID,
		BillingID:             r.BillingID,
		Category:              r.Category,
		Name:                  r.Name,
		UnitPrice:             r.UnitPrice,
		Quantity:              r.Quantity,
		DiscountRate:          r.DiscountRate,
		DiscountAmount:        r.DiscountAmount,
		TaxType:               r.TaxType,
		TaxRate:               r.TaxRate,
		IsInsuranceApplicable: r.IsInsuranceApplicable,
		Source:                r.Source,
		TreatmentID:           r.TreatmentID,
		AppointmentID:         r.AppointmentID,
		TrimmingCourseID:      r.TrimmingCourseID,
		TrimmingOptionID:      r.TrimmingOptionID,
		MerchandiseItemID:     r.MerchandiseItemID,
		SortOrder:             r.SortOrder,
	}
}

// updateBillingItemRequest は明細更新リクエスト（nil = 未指定）。
type updateBillingItemRequest struct {
	UnitPrice             *int64   `json:"unit_price"`
	Quantity              *float64 `json:"quantity"`
	DiscountRate          *float64 `json:"discount_rate" binding:"omitempty,min=0,max=100"`
	DiscountAmount        *int64   `json:"discount_amount" binding:"omitempty,min=0"`
	TaxType               *string  `json:"tax_type"  binding:"omitempty,oneof=included excluded exempt"`
	TaxRate               *float64 `json:"tax_rate"`
	IsInsuranceApplicable *bool    `json:"is_insurance_applicable"`
}

func (r updateBillingItemRequest) toServiceInput() (*service.UpdateBillingItemInput, error) {
	var taxType *model.TaxType
	if r.TaxType != nil {
		if err := validateTaxType(*r.TaxType); err != nil {
			return nil, err
		}
		t := model.TaxType(*r.TaxType)
		taxType = &t
	}

	return &service.UpdateBillingItemInput{
		UnitPrice:             r.UnitPrice,
		Quantity:              r.Quantity,
		DiscountRate:          r.DiscountRate,
		DiscountAmount:        r.DiscountAmount,
		TaxType:               taxType,
		TaxRate:               r.TaxRate,
		IsInsuranceApplicable: r.IsInsuranceApplicable,
	}, nil
}

// createRefundRequest は返金登録リクエスト。
type createRefundRequest struct {
	Amount int64  `json:"amount" binding:"required,min=1"`
	Reason string `json:"reason"`
	// PaymentMethod は返金先の支払手段（任意・ENUM・#60）。会計 payment_splits.method と同体系。
	// 混在支払いの方法別返金上限(Phase 2)に使う。
	PaymentMethod *string `json:"payment_method" binding:"omitempty,oneof=cash credit_card electronic_money"`
}
