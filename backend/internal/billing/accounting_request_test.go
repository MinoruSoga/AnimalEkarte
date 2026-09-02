package billing

import (
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- newListAccountingQuery ----

func TestNewListAccountingQuery(t *testing.T) {
	tests := []struct {
		name   string
		values url.Values
		want   listAccountingQuery
	}{
		{
			name: "normal: all params present",
			values: url.Values{
				"pet_id":            {"10"},
				"owner_id":          {"20"},
				"status":            {"completed"},
				"status_op":         {"is_not"},
				"start_date":        {"2026-05-01"},
				"end_date":          {"2026-05-31"},
				"search":            {"べるす"},
				"payment_method":    {"cash"},
				"payment_method_op": {"is"},
			},
			want: listAccountingQuery{
				PetID: "10", OwnerID: "20", Status: "completed", StatusOp: "is_not",
				StartDate: "2026-05-01", EndDate: "2026-05-31",
				Search: "べるす", PaymentMethod: "cash", PaymentMethodOp: "is",
			},
		},
		{
			name:   "zero value: empty url.Values",
			values: url.Values{},
			want:   listAccountingQuery{},
		},
		{
			name:   "nil url.Values",
			values: nil,
			want:   listAccountingQuery{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newListAccountingQuery(tt.values)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---- newListUnpaidBillingsQuery ----

func TestNewListUnpaidBillingsQuery(t *testing.T) {
	tests := []struct {
		name   string
		values url.Values
		want   listUnpaidBillingsQuery
	}{
		{
			name: "normal: all params present",
			values: url.Values{
				"start_date": {"2026-01-01"},
				"end_date":   {"2026-01-31"},
				"group_by":   {"billing"},
			},
			want: listUnpaidBillingsQuery{StartDate: "2026-01-01", EndDate: "2026-01-31", GroupBy: "billing"},
		},
		{
			name:   "zero value: empty url.Values",
			values: url.Values{},
			want:   listUnpaidBillingsQuery{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newListUnpaidBillingsQuery(tt.values)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---- newDailySummaryQuery ----

func TestNewDailySummaryQuery(t *testing.T) {
	tests := []struct {
		name   string
		values url.Values
		want   dailySummaryQuery
	}{
		{
			name:   "normal: date present",
			values: url.Values{"date": {"2026-06-01"}},
			want:   dailySummaryQuery{Date: "2026-06-01"},
		},
		{
			name:   "zero value: empty url.Values",
			values: url.Values{},
			want:   dailySummaryQuery{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newDailySummaryQuery(tt.values)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---- newMonthlyUnpaidQuery ----

func TestNewMonthlyUnpaidQuery(t *testing.T) {
	tests := []struct {
		name   string
		values url.Values
		want   monthlyUnpaidQuery
	}{
		{
			name:   "normal: year/month present",
			values: url.Values{"year": {"2026"}, "month": {"6"}},
			want:   monthlyUnpaidQuery{Year: "2026", Month: "6"},
		},
		{
			name:   "zero value: empty url.Values",
			values: url.Values{},
			want:   monthlyUnpaidQuery{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newMonthlyUnpaidQuery(tt.values)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestListAccountingQuery_ToServiceFilters(t *testing.T) {
	filters, err := (&listAccountingQuery{
		PetID:     "10",
		OwnerID:   "20",
		Status:    "completed",
		StartDate: "2026-05-01",
		EndDate:   "2026-05-31",
	}).toServiceFilters()
	if err != nil {
		t.Fatalf("toServiceFilters returned error: %v", err)
	}

	if filters.PetID == nil || *filters.PetID != 10 {
		t.Fatalf("PetID = %v, want 10", filters.PetID)
	}
	if filters.OwnerID == nil || *filters.OwnerID != 20 {
		t.Fatalf("OwnerID = %v, want 20", filters.OwnerID)
	}
	if filters.Status == nil || *filters.Status != "completed" {
		t.Fatalf("Status = %v, want completed", filters.Status)
	}
	if filters.StartDate == nil || *filters.StartDate != "2026-05-01" {
		t.Fatalf("StartDate = %v, want 2026-05-01", filters.StartDate)
	}
	if filters.EndDate == nil || *filters.EndDate != "2026-05-31" {
		t.Fatalf("EndDate = %v, want 2026-05-31", filters.EndDate)
	}
}

func TestListAccountingQuery_ToServiceFilters_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		query listAccountingQuery
	}{
		{name: "pet_id", query: listAccountingQuery{PetID: "abc"}},
		{name: "owner_id", query: listAccountingQuery{OwnerID: "abc"}},
		{name: "start_date", query: listAccountingQuery{StartDate: "2026/05/01"}},
		{name: "end_date", query: listAccountingQuery{EndDate: "2026/05/31"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters, err := tt.query.toServiceFilters()
			if err == nil {
				t.Fatal("toServiceFilters returned nil error")
			}
			if filters != (AccountingListFilters{}) {
				t.Fatalf("filters = %#v, want zero value", filters)
			}
			if !apperrors.IsInvalidInput(err) {
				t.Fatalf("error = %v, want invalid input", err)
			}
		})
	}
}

// #120: start_date/end_date 必須テスト
func TestListUnpaidBillingsQuery_ToServiceFilters(t *testing.T) {
	filters, err := (&listUnpaidBillingsQuery{
		StartDate: "2026-01-01",
		EndDate:   "2026-01-31",
		GroupBy:   "billing",
	}).toServiceFilters()
	if err != nil {
		t.Fatalf("toServiceFilters returned error: %v", err)
	}

	if filters.StartDate != "2026-01-01" {
		t.Fatalf("StartDate = %q, want 2026-01-01", filters.StartDate)
	}
	if filters.EndDate != "2026-01-31" {
		t.Fatalf("EndDate = %q, want 2026-01-31", filters.EndDate)
	}
	if filters.GroupBy != "billing" {
		t.Fatalf("GroupBy = %q, want billing", filters.GroupBy)
	}
}

func TestListUnpaidBillingsQuery_ToServiceFilters_DefaultGroupBy(t *testing.T) {
	filters, err := (&listUnpaidBillingsQuery{
		StartDate: "2026-01-01",
		EndDate:   "2026-01-31",
	}).toServiceFilters()
	if err != nil {
		t.Fatalf("toServiceFilters returned error: %v", err)
	}

	if filters.GroupBy != "owner" {
		t.Fatalf("GroupBy = %q, want owner", filters.GroupBy)
	}
}

func TestListUnpaidBillingsQuery_ToServiceFilters_MissingStartDate(t *testing.T) {
	_, err := (&listUnpaidBillingsQuery{EndDate: "2026-01-31"}).toServiceFilters()
	if err == nil {
		t.Fatal("toServiceFilters returned nil error for missing start_date")
	}
	if !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
}

func TestListUnpaidBillingsQuery_ToServiceFilters_MissingEndDate(t *testing.T) {
	_, err := (&listUnpaidBillingsQuery{StartDate: "2026-01-01"}).toServiceFilters()
	if err == nil {
		t.Fatal("toServiceFilters returned nil error for missing end_date")
	}
	if !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
}

func TestListUnpaidBillingsQuery_ToServiceFilters_InvalidDate(t *testing.T) {
	filters, err := (&listUnpaidBillingsQuery{StartDate: "2026/01/01", EndDate: "2026-01-31"}).toServiceFilters()
	if err == nil {
		t.Fatal("toServiceFilters returned nil error for invalid date format")
	}
	if filters != (listUnpaidBillingsFilters{}) {
		t.Fatalf("filters = %#v, want zero value", filters)
	}
	if !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
}

func TestCreateAccountingRequest_ToServiceInput(t *testing.T) {
	scheduledDate := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, 5, 28, 10, 30, 0, 0, time.UTC)
	medicalRecordID := uint64(10)
	hospitalizationID := uint64(20)
	ownerID := uint64(30)
	petID := uint64(40)
	req := createAccountingRequest{
		MedicalRecordID:   &medicalRecordID,
		HospitalizationID: &hospitalizationID,
		OwnerID:           &ownerID,
		PetID:             &petID,
		Subtotal:          1000,
		TaxTotal:          100,
		TotalAmount:       1100,
		HasInsurance:      true,
		Status:            string(model.BillingStatusWaiting),
		ScheduledDate:     scheduledDate,
		CompletedAt:       &completedAt,
		Memo:              "memo",
	}

	input := req.toServiceInput(1)

	if input.ClinicID != 1 {
		t.Fatalf("ClinicID = %d, want 1", input.ClinicID)
	}
	if input.MedicalRecordID == nil || *input.MedicalRecordID != medicalRecordID {
		t.Fatalf("MedicalRecordID = %v, want %d", input.MedicalRecordID, medicalRecordID)
	}
	if input.HospitalizationID == nil || *input.HospitalizationID != hospitalizationID {
		t.Fatalf("HospitalizationID = %v, want %d", input.HospitalizationID, hospitalizationID)
	}
	if input.OwnerID == nil || *input.OwnerID != ownerID {
		t.Fatalf("OwnerID = %v, want %d", input.OwnerID, ownerID)
	}
	if input.PetID == nil || *input.PetID != petID {
		t.Fatalf("PetID = %v, want %d", input.PetID, petID)
	}
	if input.Subtotal != req.Subtotal || input.TaxTotal != req.TaxTotal || input.TotalAmount != req.TotalAmount {
		t.Fatalf("amounts = (%d, %d, %d), want (%d, %d, %d)", input.Subtotal, input.TaxTotal, input.TotalAmount, req.Subtotal, req.TaxTotal, req.TotalAmount)
	}
	if !input.HasInsurance {
		t.Fatal("HasInsurance = false, want true")
	}
	if input.Status != model.BillingStatusWaiting {
		t.Fatalf("Status = %q, want %q", input.Status, model.BillingStatusWaiting)
	}
	if !input.ScheduledDate.Equal(scheduledDate) {
		t.Fatalf("ScheduledDate = %v, want %v", input.ScheduledDate, scheduledDate)
	}
	if input.CompletedAt == nil || !input.CompletedAt.Equal(completedAt) {
		t.Fatalf("CompletedAt = %v, want %v", input.CompletedAt, completedAt)
	}
	if input.Memo != req.Memo {
		t.Fatalf("Memo = %q, want %q", input.Memo, req.Memo)
	}
}

func TestUpdateAccountingRequest_ToServiceInput(t *testing.T) {
	status := string(model.BillingStatusCompleted)
	paymentMethod := string(model.PaymentMethodCash)
	memo := ""
	subtotal := int64(1000)
	taxTotal := int64(100)
	totalAmount := int64(1100)
	hasInsurance := false
	insuranceRatio := 50.0
	insuranceName := "insurance"
	insuranceAmount := int64(500)
	discountAmount := int64(0)
	billingAmount := int64(600)
	receivedAmount := int64(1000)
	changeAmount := int64(400)
	paymentMethodID := uint64(9)
	req := updateAccountingRequest{
		Subtotal:        &subtotal,
		TaxTotal:        &taxTotal,
		TotalAmount:     &totalAmount,
		HasInsurance:    &hasInsurance,
		Status:          &status,
		Memo:            &memo,
		PaymentMethod:   &paymentMethod,
		InsuranceRatio:  &insuranceRatio,
		InsuranceName:   &insuranceName,
		InsuranceAmount: &insuranceAmount,
		DiscountAmount:  &discountAmount,
		BillingAmount:   &billingAmount,
		ReceivedAmount:  &receivedAmount,
		ChangeAmount:    &changeAmount,
		PaymentSplits: []paymentSplitRequest{
			{
				Method:          string(model.PaymentMethodCash),
				PaymentMethodID: &paymentMethodID,
				Amount:          600,
				ReceivedAmount:  1000,
				ChangeAmount:    400,
			},
		},
	}

	input := req.toServiceInput(1, 2, 3)

	if input.ID != 1 || input.ClinicID != 2 {
		t.Fatalf("ID/ClinicID = %d/%d, want 1/2", input.ID, input.ClinicID)
	}
	if input.StaffID == nil || *input.StaffID != 3 {
		t.Fatalf("StaffID = %v, want 3", input.StaffID)
	}
	if input.Subtotal == nil || *input.Subtotal != subtotal {
		t.Fatalf("Subtotal = %v, want %d", input.Subtotal, subtotal)
	}
	if input.HasInsurance == nil || *input.HasInsurance != hasInsurance {
		t.Fatalf("HasInsurance = %v, want false", input.HasInsurance)
	}
	if input.Status == nil || *input.Status != model.BillingStatusCompleted {
		t.Fatalf("Status = %v, want %q", input.Status, model.BillingStatusCompleted)
	}
	if input.Memo == nil || *input.Memo != memo {
		t.Fatalf("Memo = %v, want explicit empty string", input.Memo)
	}
	if input.PaymentMethod == nil || *input.PaymentMethod != model.PaymentMethodCash {
		t.Fatalf("PaymentMethod = %v, want %q", input.PaymentMethod, model.PaymentMethodCash)
	}
	if input.DiscountAmount == nil || *input.DiscountAmount != 0 {
		t.Fatalf("DiscountAmount = %v, want explicit zero", input.DiscountAmount)
	}
	if len(input.PaymentSplits) != 1 {
		t.Fatalf("PaymentSplits length = %d, want 1", len(input.PaymentSplits))
	}
	split := input.PaymentSplits[0]
	if split.Method != model.PaymentMethodCash {
		t.Fatalf("split.Method = %q, want %q", split.Method, model.PaymentMethodCash)
	}
	if split.PaymentMethodID == nil || *split.PaymentMethodID != paymentMethodID {
		t.Fatalf("split.PaymentMethodID = %v, want %d", split.PaymentMethodID, paymentMethodID)
	}
	if split.Amount != 600 || split.ReceivedAmount != 1000 || split.ChangeAmount != 400 {
		t.Fatalf("split amounts = (%d, %d, %d), want (600, 1000, 400)", split.Amount, split.ReceivedAmount, split.ChangeAmount)
	}
}

func TestUpdateAccountingRequest_ToServiceInput_NilPaymentSplits(t *testing.T) {
	req := updateAccountingRequest{}

	input := req.toServiceInput(1, 2, 3)

	if input.PaymentSplits != nil {
		t.Fatalf("PaymentSplits = %#v, want nil", input.PaymentSplits)
	}
	if input.Status != nil {
		t.Fatalf("Status = %v, want nil", input.Status)
	}
	if input.PaymentMethod != nil {
		t.Fatalf("PaymentMethod = %v, want nil", input.PaymentMethod)
	}
}

// TestUpdateAccountingRequest_PostCloseReason verifies PostCloseReason propagates to service input (#115)
func TestUpdateAccountingRequest_PostCloseReason(t *testing.T) {
	reason := "締め後修正: 入力誤りのため"
	req := updateAccountingRequest{
		PostCloseReason: &reason,
	}

	input := req.toServiceInput(1, 2, 3)

	if input.PostCloseReason == nil {
		t.Fatal("PostCloseReason = nil, want non-nil pointer")
	}
	if *input.PostCloseReason != reason {
		t.Fatalf("PostCloseReason = %q, want %q", *input.PostCloseReason, reason)
	}
}

// TestUpdateAccountingRequest_PostCloseReasonNil verifies nil PostCloseReason stays nil
func TestUpdateAccountingRequest_PostCloseReasonNil(t *testing.T) {
	req := updateAccountingRequest{}

	input := req.toServiceInput(1, 2, 3)

	if input.PostCloseReason != nil {
		t.Fatalf("PostCloseReason = %v, want nil", input.PostCloseReason)
	}
}

// #114: 月次未納繰越クエリのバリデーション
func TestMonthlyUnpaidQuery_Parse(t *testing.T) {
	tests := []struct {
		name      string
		query     monthlyUnpaidQuery
		wantYear  int
		wantMonth int
		wantErr   bool
	}{
		{name: "正常: 2026年6月", query: monthlyUnpaidQuery{Year: "2026", Month: "6"}, wantYear: 2026, wantMonth: 6},
		{name: "正常: 2000年1月(下限)", query: monthlyUnpaidQuery{Year: "2000", Month: "1"}, wantYear: 2000, wantMonth: 1},
		{name: "正常: 2100年12月(上限)", query: monthlyUnpaidQuery{Year: "2100", Month: "12"}, wantYear: 2100, wantMonth: 12},
		{name: "エラー: year 欠損", query: monthlyUnpaidQuery{Month: "6"}, wantErr: true},
		{name: "エラー: month 欠損", query: monthlyUnpaidQuery{Year: "2026"}, wantErr: true},
		{name: "エラー: year 非数値", query: monthlyUnpaidQuery{Year: "abc", Month: "6"}, wantErr: true},
		{name: "エラー: year=1999 (下限未満)", query: monthlyUnpaidQuery{Year: "1999", Month: "6"}, wantErr: true},
		{name: "エラー: year=2101 (上限超)", query: monthlyUnpaidQuery{Year: "2101", Month: "6"}, wantErr: true},
		{name: "エラー: month=0", query: monthlyUnpaidQuery{Year: "2026", Month: "0"}, wantErr: true},
		{name: "エラー: month=13", query: monthlyUnpaidQuery{Year: "2026", Month: "13"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			year, month, err := tt.query.parse()
			if tt.wantErr {
				if err == nil {
					t.Fatal("parse() returned nil error, want error")
				}
				if !apperrors.IsInvalidInput(err) {
					t.Fatalf("error = %v, want invalid input", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parse() returned error: %v", err)
			}
			if year != tt.wantYear {
				t.Fatalf("year = %d, want %d", year, tt.wantYear)
			}
			if month != tt.wantMonth {
				t.Fatalf("month = %d, want %d", month, tt.wantMonth)
			}
		})
	}
}

// #127: payment_method の oneof バインド検証。
// oneof 制約は gin の binding 層（ShouldBindJSON 時）で評価されるため、
// gin と同じ tag 名 "binding" を使う binding.Validator.ValidateStruct で直接検証する。
// toServiceInput は変換のみで制約を評価しないため、ここでないと oneof を回帰できない。
func TestPaymentSplitRequest_Binding_MethodOneof(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		wantErr bool
	}{
		{name: "cash", method: string(model.PaymentMethodCash), wantErr: false},
		{name: "credit_card", method: string(model.PaymentMethodCreditCard), wantErr: false},
		{name: "electronic_money", method: string(model.PaymentMethodElectronicMoney), wantErr: false},
		{name: "bank_transfer (#127)", method: string(model.PaymentMethodBankTransfer), wantErr: false},
		{name: "unknown method rejected", method: "paypay", wantErr: true},
		{name: "empty method rejected (required)", method: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Amount は required,min=1 のため Method 単独を検証できるよう常に有効値を入れる。
			req := paymentSplitRequest{Method: tt.method, Amount: 1}
			err := binding.Validator.ValidateStruct(&req)
			if tt.wantErr && err == nil {
				t.Fatalf("ValidateStruct(method=%q) = nil, want validation error", tt.method)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("ValidateStruct(method=%q) = %v, want nil", tt.method, err)
			}
		})
	}
}

// #189: クレジット訂正リクエストの binding 検証。
// method は credit_card / electronic_money のみ許可（cash/bank_transfer/未知は拒否）、
// amount は required,min=1、reason は required。oneof/required は gin binding 層で評価される。
func TestCorrectCreditPaymentRequest_Binding(t *testing.T) {
	tests := []struct {
		name    string
		req     correctCreditPaymentRequest
		wantErr bool
	}{
		{name: "正常: credit_card", req: correctCreditPaymentRequest{Method: "credit_card", Amount: 12000, Reason: "打ち間違い"}, wantErr: false},
		{name: "正常: electronic_money", req: correctCreditPaymentRequest{Method: "electronic_money", Amount: 500, Reason: "訂正"}, wantErr: false},
		{name: "拒否: cash は対象外", req: correctCreditPaymentRequest{Method: "cash", Amount: 100, Reason: "x"}, wantErr: true},
		{name: "拒否: bank_transfer は対象外", req: correctCreditPaymentRequest{Method: "bank_transfer", Amount: 100, Reason: "x"}, wantErr: true},
		{name: "拒否: 未知の手段", req: correctCreditPaymentRequest{Method: "paypay", Amount: 100, Reason: "x"}, wantErr: true},
		{name: "拒否: reason 空", req: correctCreditPaymentRequest{Method: "credit_card", Amount: 100, Reason: ""}, wantErr: true},
		{name: "拒否: amount=0", req: correctCreditPaymentRequest{Method: "credit_card", Amount: 0, Reason: "x"}, wantErr: true},
		{name: "拒否: amount 負", req: correctCreditPaymentRequest{Method: "credit_card", Amount: -100, Reason: "x"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := binding.Validator.ValidateStruct(&tt.req)
			if tt.wantErr && err == nil {
				t.Fatalf("ValidateStruct(%+v) = nil, want validation error", tt.req)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("ValidateStruct(%+v) = %v, want nil", tt.req, err)
			}
		})
	}
}

// #189: クレジット訂正リクエストの service 入力変換が全フィールドを伝播することを検証する。
func TestCorrectCreditPaymentRequest_ToServiceInput(t *testing.T) {
	req := correctCreditPaymentRequest{Method: "credit_card", Amount: 8800, Reason: "端末打ち間違い", Memo: "確認済み"}
	in := req.toServiceInput(10, 1, 7)
	if in.BillingID != 10 || in.ClinicID != 1 {
		t.Fatalf("BillingID/ClinicID = %d/%d, want 10/1", in.BillingID, in.ClinicID)
	}
	if in.StaffID == nil || *in.StaffID != 7 {
		t.Fatalf("StaffID = %v, want 7", in.StaffID)
	}
	if in.Method != model.PaymentMethodCreditCard {
		t.Fatalf("Method = %q, want credit_card", in.Method)
	}
	if in.Amount != 8800 {
		t.Fatalf("Amount = %d, want 8800", in.Amount)
	}
	if in.Reason != "端末打ち間違い" || in.Memo != "確認済み" {
		t.Fatalf("Reason/Memo = %q/%q, want 端末打ち間違い/確認済み", in.Reason, in.Memo)
	}
}

func TestPaymentSplitRequest_ToServiceInput(t *testing.T) {
	paymentMethodID := uint64(8)
	req := paymentSplitRequest{
		Method:          string(model.PaymentMethodCreditCard),
		PaymentMethodID: &paymentMethodID,
		Amount:          1200,
		ReceivedAmount:  0,
		ChangeAmount:    0,
	}

	input := req.toServiceInput()

	if input.Method != model.PaymentMethodCreditCard {
		t.Fatalf("Method = %q, want %q", input.Method, model.PaymentMethodCreditCard)
	}
	if input.PaymentMethodID == nil || *input.PaymentMethodID != paymentMethodID {
		t.Fatalf("PaymentMethodID = %v, want %d", input.PaymentMethodID, paymentMethodID)
	}
	if input.Amount != req.Amount {
		t.Fatalf("Amount = %d, want %d", input.Amount, req.Amount)
	}
}

// TestPaymentMethodPointer_Binding_Oneof は updateAccountingRequest の oneof binding を検証する
// （返金分は internal/billing 側へ分離・B③）。
func TestPaymentMethodPointer_Binding_Oneof(t *testing.T) {
	bankTransfer := string(model.PaymentMethodBankTransfer)
	invalid := "paypay"
	t.Run("updateAccountingRequest accepts bank_transfer", func(t *testing.T) {
		req := updateAccountingRequest{PaymentMethod: &bankTransfer}
		if err := binding.Validator.ValidateStruct(&req); err != nil {
			t.Fatalf("ValidateStruct = %v, want nil for bank_transfer", err)
		}
	})
	t.Run("updateAccountingRequest accepts nil (omitempty)", func(t *testing.T) {
		req := updateAccountingRequest{}
		if err := binding.Validator.ValidateStruct(&req); err != nil {
			t.Fatalf("ValidateStruct = %v, want nil for nil payment_method", err)
		}
	})
	t.Run("updateAccountingRequest rejects unknown method", func(t *testing.T) {
		req := updateAccountingRequest{PaymentMethod: &invalid}
		if err := binding.Validator.ValidateStruct(&req); err == nil {
			t.Fatal("ValidateStruct = nil, want validation error for unknown payment_method")
		}
	})
}
