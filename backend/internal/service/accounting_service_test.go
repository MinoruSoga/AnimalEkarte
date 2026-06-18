package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// mockAccountingRepository は AccountingRepository のテスト用モック実装
type mockAccountingRepository struct {
	findAllFn           func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error)
	findByIDFn          func(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
	createFn            func(ctx context.Context, clinicID uint64, accounting *model.Billing) error
	updateFieldsFn      func(ctx context.Context, clinicID, billingID uint64, fields map[string]any) (*model.Billing, error)
	savePaymentSplitsFn func(ctx context.Context, splits []model.PaymentSplit) error
	completeApptsFn     func(ctx context.Context, clinicID uint64, medicalRecordID, ownerID, petID *uint64, scheduledDate time.Time) (int64, error)
	getDailySummaryFn   func(ctx context.Context, clinicID uint64, date time.Time) (*repository.DailySummaryResult, error)
	// #120: start_date/end_date 2引数バリアント
	findUnpaidByBillingFn func(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]model.Billing, int64, error)
	findUnpaidByOwnerFn   func(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]repository.UnpaidOwnerAggregate, int64, repository.UnpaidSummary, error)
	// #114: 月次未納繰越集計
	findMonthlyUnpaidCarryoverFn func(ctx context.Context, clinicID uint64, firstDay, lastDay string, page, limit int) ([]repository.MonthlyUnpaidOwnerPet, int64, repository.MonthlyUnpaidSummary, error)
}

func (m *mockAccountingRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error) {
	return m.findAllFn(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
}

func (m *mockAccountingRepository) FindAllForClinics(_ context.Context, _ []uint64, _, _ *uint64, _, _, _ *string, _, _ int) ([]model.Billing, int64, error) {
	return nil, 0, nil
}

func (m *mockAccountingRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockAccountingRepository) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.Billing, error) {
	return nil, nil
}

func (m *mockAccountingRepository) LockAndFindByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockAccountingRepository) Create(ctx context.Context, clinicID uint64, accounting *model.Billing) error {
	return m.createFn(ctx, clinicID, accounting)
}

func (m *mockAccountingRepository) Update(ctx context.Context, clinicID, billingID uint64, fields map[string]any) (*model.Billing, error) {
	return m.updateFieldsFn(ctx, clinicID, billingID, fields)
}

func (m *mockAccountingRepository) SavePayment(_ context.Context, _ *model.Payment) error {
	return nil
}

func (m *mockAccountingRepository) SavePaymentSplits(ctx context.Context, splits []model.PaymentSplit) error {
	if m.savePaymentSplitsFn != nil {
		return m.savePaymentSplitsFn(ctx, splits)
	}
	return nil
}

func (m *mockAccountingRepository) CompleteAccountingAppointments(ctx context.Context, clinicID uint64, medicalRecordID, ownerID, petID *uint64, scheduledDate time.Time) (int64, error) {
	if m.completeApptsFn != nil {
		return m.completeApptsFn(ctx, clinicID, medicalRecordID, ownerID, petID, scheduledDate)
	}
	return 0, nil
}

// #120: 未納者一覧 repository メソッドの mock（start_date/end_date 2引数）
func (m *mockAccountingRepository) FindUnpaidByBilling(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]model.Billing, int64, error) {
	if m.findUnpaidByBillingFn != nil {
		return m.findUnpaidByBillingFn(ctx, clinicID, startDate, endDate, page, limit)
	}
	return nil, 0, nil
}

func (m *mockAccountingRepository) FindUnpaidByOwner(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]repository.UnpaidOwnerAggregate, int64, repository.UnpaidSummary, error) {
	if m.findUnpaidByOwnerFn != nil {
		return m.findUnpaidByOwnerFn(ctx, clinicID, startDate, endDate, page, limit)
	}
	return nil, 0, repository.UnpaidSummary{}, nil
}

func (m *mockAccountingRepository) GetDailySummary(ctx context.Context, clinicID uint64, date time.Time) (*repository.DailySummaryResult, error) {
	if m.getDailySummaryFn != nil {
		return m.getDailySummaryFn(ctx, clinicID, date)
	}
	return &repository.DailySummaryResult{PaymentTotals: []repository.PaymentMethodTotal{}, CategoryTotals: []repository.CategoryTotal{}}, nil
}

// FEAT-368: 集計・締め機能 mock スタブ
func (m *mockAccountingRepository) GetCloseAggregate(_ context.Context, _ repository.GetCloseAggregateInput) (*repository.CloseAggregateResult, error) {
	return &repository.CloseAggregateResult{
		PaymentRows:    []repository.PaymentAggregateRow{},
		CategoryRows:   []repository.CategoryAggregateRow{},
		BillingDetails: []repository.CloseBillingDetail{},
	}, nil
}

func (m *mockAccountingRepository) GetMonthlyReport(_ context.Context, _ uint64, _, _ int) (*repository.MonthlyReportResult, error) {
	return &repository.MonthlyReportResult{}, nil
}

func (m *mockAccountingRepository) SumPaidByOwner(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockAccountingRepository) MaxSingleVisitAmountByOwner(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockAccountingRepository) FindOwnersByAnnualRevenue(_ context.Context, _ uint64) ([]repository.OwnerAnnualRevenue, error) {
	return nil, nil
}

// #114: 月次未納繰越集計 mock
func (m *mockAccountingRepository) FindMonthlyUnpaidCarryover(ctx context.Context, clinicID uint64, firstDay, lastDay string, page, limit int) ([]repository.MonthlyUnpaidOwnerPet, int64, repository.MonthlyUnpaidSummary, error) {
	if m.findMonthlyUnpaidCarryoverFn != nil {
		return m.findMonthlyUnpaidCarryoverFn(ctx, clinicID, firstDay, lastDay, page, limit)
	}
	return nil, 0, repository.MonthlyUnpaidSummary{}, nil
}

func ptrString(v string) *string { return &v }

// TestAccountingService_GetMonthlyUnpaidCarryover は月次未納繰越集計サービスメソッドのテスト。#114
func TestAccountingService_GetMonthlyUnpaidCarryover(t *testing.T) {
	petID := uint64(3)
	tests := []struct {
		name        string
		year        int
		month       int
		mockFn      func(ctx context.Context, clinicID uint64, firstDay, lastDay string, page, limit int) ([]repository.MonthlyUnpaidOwnerPet, int64, repository.MonthlyUnpaidSummary, error)
		wantSummary repository.MonthlyUnpaidSummary
		wantTotal   int64
		wantLen     int
		wantErr     bool
	}{
		{
			name:  "正常: firstDay/lastDay が正しく計算されデータが返る",
			year:  2026,
			month: 6,
			mockFn: func(_ context.Context, _ uint64, firstDay, lastDay string, _, _ int) ([]repository.MonthlyUnpaidOwnerPet, int64, repository.MonthlyUnpaidSummary, error) {
				if firstDay != "2026-06-01" || lastDay != "2026-06-30" {
					t.Errorf("want firstDay=2026-06-01 lastDay=2026-06-30, got firstDay=%s lastDay=%s", firstDay, lastDay)
				}
				items := []repository.MonthlyUnpaidOwnerPet{
					{OwnerID: 1, OwnerName: "田中", PetID: &petID, PetName: "ポチ", PrevMonthCarryover: 10000, CurrentMonthUnpaid: 5000, NextMonthCarryover: 15000},
					{OwnerID: 2, OwnerName: "鈴木", PetID: nil, PetName: "", PrevMonthCarryover: 0, CurrentMonthUnpaid: 3000, NextMonthCarryover: 3000},
				}
				return items, 2, repository.MonthlyUnpaidSummary{PrevMonthCarryover: 10000, CurrentMonthUnpaid: 8000, NextMonthCarryover: 18000}, nil
			},
			wantSummary: repository.MonthlyUnpaidSummary{PrevMonthCarryover: 10000, CurrentMonthUnpaid: 8000, NextMonthCarryover: 18000},
			wantTotal:   2,
			wantLen:     2,
		},
		{
			name:  "正常: 1月（firstDay=01-01, lastDay=01-31）",
			year:  2026,
			month: 1,
			mockFn: func(_ context.Context, _ uint64, firstDay, lastDay string, _, _ int) ([]repository.MonthlyUnpaidOwnerPet, int64, repository.MonthlyUnpaidSummary, error) {
				if firstDay != "2026-01-01" || lastDay != "2026-01-31" {
					t.Errorf("want firstDay=2026-01-01 lastDay=2026-01-31, got firstDay=%s lastDay=%s", firstDay, lastDay)
				}
				return nil, 0, repository.MonthlyUnpaidSummary{}, nil
			},
			wantSummary: repository.MonthlyUnpaidSummary{},
			wantTotal:   0,
			wantLen:     0,
		},
		{
			name:    "エラー: month=0 は ErrInvalidInput",
			year:    2026,
			month:   0,
			wantErr: true,
		},
		{
			name:    "エラー: month=13 は ErrInvalidInput",
			year:    2026,
			month:   13,
			wantErr: true,
		},
		{
			name:  "エラー: リポジトリエラーを伝播する",
			year:  2026,
			month: 6,
			mockFn: func(_ context.Context, _ uint64, _, _ string, _, _ int) ([]repository.MonthlyUnpaidOwnerPet, int64, repository.MonthlyUnpaidSummary, error) {
				return nil, 0, repository.MonthlyUnpaidSummary{}, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockAccountingRepository{findMonthlyUnpaidCarryoverFn: tt.mockFn}
			svc := NewAccountingService(mock, nil, nil, nil)

			items, total, summary, err := svc.GetMonthlyUnpaidCarryover(context.Background(), 1, tt.year, tt.month, 1, 20)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantTotal, total)
			assert.Len(t, items, tt.wantLen)
			assert.Equal(t, tt.wantSummary, summary)
		})
	}
}

func TestAccountingService_List(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name         string
		clinicID     uint64
		petID        *uint64
		ownerID      *uint64
		status       *string
		page         int
		limit        int
		repoBillings []model.Billing
		repoTotal    int64
		repoErr      error
		wantLen      int
		wantTotal    int64
		wantErr      bool
	}{
		{
			name:     "returns all billings for clinic",
			clinicID: 1,
			petID:    nil,
			ownerID:  nil,
			status:   nil,
			page:     1,
			limit:    20,
			repoBillings: []model.Billing{
				{ID: 1, ClinicID: 1, ScheduledDate: now, Status: model.BillingStatusWaiting},
				{ID: 2, ClinicID: 1, ScheduledDate: now, Status: model.BillingStatusCompleted},
			},
			repoTotal: 2,
			repoErr:   nil,
			wantLen:   2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:     "filters by pet_id",
			clinicID: 1,
			petID:    ptrUint64(10),
			ownerID:  nil,
			status:   nil,
			page:     1,
			limit:    20,
			repoBillings: []model.Billing{
				{ID: 1, ClinicID: 1, ScheduledDate: now, Status: model.BillingStatusWaiting},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:     "filters by owner_id",
			clinicID: 1,
			petID:    nil,
			ownerID:  ptrUint64(5),
			status:   nil,
			page:     1,
			limit:    20,
			repoBillings: []model.Billing{
				{ID: 2, ClinicID: 1, ScheduledDate: now, Status: model.BillingStatusCompleted},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:     "filters by status",
			clinicID: 1,
			petID:    nil,
			ownerID:  nil,
			status:   ptrString("waiting"),
			page:     1,
			limit:    20,
			repoBillings: []model.Billing{
				{ID: 1, ClinicID: 1, ScheduledDate: now, Status: model.BillingStatusWaiting},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:         "returns empty list when no billings exist",
			clinicID:     1,
			petID:        nil,
			ownerID:      nil,
			status:       nil,
			page:         1,
			limit:        20,
			repoBillings: []model.Billing{},
			repoTotal:    0,
			repoErr:      nil,
			wantLen:      0,
			wantTotal:    0,
			wantErr:      false,
		},
		{
			name:         "propagates repository error",
			clinicID:     1,
			petID:        nil,
			ownerID:      nil,
			status:       nil,
			page:         1,
			limit:        20,
			repoBillings: nil,
			repoTotal:    0,
			repoErr:      errors.New("db connection error"),
			wantLen:      0,
			wantTotal:    0,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAccountingRepository{
				findAllFn: func(_ context.Context, _ uint64, _ *uint64, _ *uint64, _, _, _ *string, _, _ int) ([]model.Billing, int64, error) {
					return tt.repoBillings, tt.repoTotal, tt.repoErr
				},
			}
			svc := NewAccountingService(repo, nil, &mockTransactor{}, &mockAccountingAuditService{})

			billings, total, err := svc.List(context.Background(), tt.clinicID, tt.petID, tt.ownerID, tt.status, nil, nil, tt.page, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, billings, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
			}
		})
	}
}

func TestAccountingService_GetByID(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		clinicID    uint64
		id          uint64
		repoBilling *model.Billing
		repoErr     error
		wantBilling *model.Billing
		wantErr     error
	}{
		{
			name:        "returns billing when found",
			clinicID:    1,
			id:          10,
			repoBilling: &model.Billing{ID: 10, ClinicID: 1, ScheduledDate: now, Status: model.BillingStatusWaiting},
			repoErr:     nil,
			wantBilling: &model.Billing{ID: 10, ClinicID: 1, ScheduledDate: now, Status: model.BillingStatusWaiting},
			wantErr:     nil,
		},
		{
			name:        "returns not found error when billing does not exist",
			clinicID:    1,
			id:          999,
			repoBilling: nil,
			repoErr:     apperrors.WrapNotFound("billing", "999"),
			wantBilling: nil,
			wantErr:     apperrors.ErrNotFound,
		},
		{
			name:        "returns error on repository failure",
			clinicID:    1,
			id:          10,
			repoBilling: nil,
			repoErr:     errors.New("db error"),
			wantBilling: nil,
			wantErr:     errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAccountingRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
					return tt.repoBilling, tt.repoErr
				},
			}
			svc := NewAccountingService(repo, nil, &mockTransactor{}, &mockAccountingAuditService{})

			billing, err := svc.GetByID(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr != nil {
				assert.Error(t, err)
				if errors.Is(tt.wantErr, apperrors.ErrNotFound) {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantBilling, billing)
			}
		})
	}
}

func TestAccountingService_Create(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		input   CreateAccountingInput
		repoErr error
		wantErr bool
	}{
		{
			name: "creates billing successfully",
			input: CreateAccountingInput{
				ClinicID:      1,
				ScheduledDate: now,
				Status:        model.BillingStatusWaiting,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns validation error when scheduled_date is zero",
			input: CreateAccountingInput{
				ClinicID: 1,
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: CreateAccountingInput{
				ClinicID:      1,
				ScheduledDate: now,
			},
			repoErr: errors.New("db connection error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAccountingRepository{
				createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
					return tt.repoErr
				},
			}
			svc := NewAccountingService(repo, nil, &mockTransactor{}, &mockAccountingAuditService{})

			billing, err := svc.Create(context.Background(), &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, billing)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, billing)
			}
		})
	}
}

func TestAccountingService_Create_SyncsCPMStageTagBestEffortWhenCompleted(t *testing.T) {
	ownerID := uint64(10)
	var syncedClinicID, syncedOwnerID uint64

	repo := &mockAccountingRepository{
		createFn: func(_ context.Context, _ uint64, accounting *model.Billing) error {
			accounting.ID = 30
			return nil
		},
	}
	tagSync := &mockLstepTagSyncService{
		syncCPMStageTagFn: func(_ context.Context, clinicID, ownerID uint64) error {
			syncedClinicID = clinicID
			syncedOwnerID = ownerID
			return errors.New("sync failed")
		},
	}
	svc := NewAccountingService(repo, tagSync, &mockTransactor{}, &mockAccountingAuditService{})

	billing, err := svc.Create(context.Background(), &CreateAccountingInput{
		ClinicID:      1,
		OwnerID:       &ownerID,
		Status:        model.BillingStatusCompleted,
		ScheduledDate: time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC),
	})

	assert.NoError(t, err)
	assert.NotNil(t, billing)
	assert.Equal(t, uint64(1), syncedClinicID)
	assert.Equal(t, ownerID, syncedOwnerID)
}

func TestAccountingService_Create_CompletesSameDayAccountingAppointments(t *testing.T) {
	ownerID := uint64(10)
	petID := uint64(20)
	scheduledDate := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	var completedClinicID uint64
	var completedOwnerID uint64
	var completedPetID uint64
	var completedDate time.Time

	repo := &mockAccountingRepository{
		createFn: func(_ context.Context, _ uint64, accounting *model.Billing) error {
			accounting.ID = 30
			return nil
		},
		completeApptsFn: func(_ context.Context, clinicID uint64, _, ownerID, petID *uint64, scheduledDate time.Time) (int64, error) {
			completedClinicID = clinicID
			completedOwnerID = *ownerID
			completedPetID = *petID
			completedDate = scheduledDate
			return 2, nil
		},
	}
	svc := NewAccountingService(repo, nil, &mockTransactor{}, &mockAccountingAuditService{})

	billing, err := svc.Create(context.Background(), &CreateAccountingInput{
		ClinicID:      1,
		OwnerID:       &ownerID,
		PetID:         &petID,
		Status:        model.BillingStatusCompleted,
		ScheduledDate: scheduledDate,
	})

	assert.NoError(t, err)
	assert.NotNil(t, billing)
	assert.Equal(t, uint64(1), completedClinicID)
	assert.Equal(t, ownerID, completedOwnerID)
	assert.Equal(t, petID, completedPetID)
	assert.Equal(t, scheduledDate, completedDate)
}

// #77: 会計完了時、診察カードの orphan 残留を防ぐため billing.medical_record_id が
// CompleteAccountingAppointments に渡ることを検証する。
func TestAccountingService_Create_PassesMedicalRecordIDToCompleteAppointments(t *testing.T) {
	medicalRecordID := uint64(42)
	var gotMedicalRecordID *uint64
	called := false

	repo := &mockAccountingRepository{
		createFn: func(_ context.Context, _ uint64, accounting *model.Billing) error {
			accounting.ID = 30
			return nil
		},
		completeApptsFn: func(_ context.Context, _ uint64, mrID, _, _ *uint64, _ time.Time) (int64, error) {
			called = true
			gotMedicalRecordID = mrID
			return 1, nil
		},
	}
	svc := NewAccountingService(repo, nil, &mockTransactor{}, &mockAccountingAuditService{})

	_, err := svc.Create(context.Background(), &CreateAccountingInput{
		ClinicID:        1,
		MedicalRecordID: &medicalRecordID,
		Status:          model.BillingStatusCompleted,
		ScheduledDate:   time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC),
	})

	assert.NoError(t, err)
	assert.True(t, called, "CompleteAccountingAppointments が呼ばれること")
	if assert.NotNil(t, gotMedicalRecordID) {
		assert.Equal(t, medicalRecordID, *gotMedicalRecordID)
	}
}

// #77: トリミングのみ会計（medical_record_id なし）では nil が渡り、診察補完がスキップされること。
func TestAccountingService_Create_NilMedicalRecordIDForTrimmingOnly(t *testing.T) {
	ownerID := uint64(10)
	petID := uint64(20)
	gotNonNil := true

	repo := &mockAccountingRepository{
		createFn: func(_ context.Context, _ uint64, accounting *model.Billing) error {
			accounting.ID = 30
			return nil
		},
		completeApptsFn: func(_ context.Context, _ uint64, mrID, _, _ *uint64, _ time.Time) (int64, error) {
			gotNonNil = mrID != nil
			return 1, nil
		},
	}
	svc := NewAccountingService(repo, nil, &mockTransactor{}, &mockAccountingAuditService{})

	_, err := svc.Create(context.Background(), &CreateAccountingInput{
		ClinicID:      1,
		OwnerID:       &ownerID,
		PetID:         &petID,
		Status:        model.BillingStatusCompleted,
		ScheduledDate: time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC),
	})

	assert.NoError(t, err)
	assert.False(t, gotNonNil, "medical_record_id 未設定なら nil が渡る")
}

func TestAccountingService_Update(t *testing.T) {
	now := time.Now()
	status := model.BillingStatusCompleted
	tests := []struct {
		name    string
		input   UpdateAccountingInput
		repoRet *model.Billing
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name: "updates billing successfully",
			input: UpdateAccountingInput{
				ID:            1,
				ClinicID:      1,
				ScheduledDate: &now,
				Status:        &status,
			},
			repoRet: &model.Billing{ID: 1, ClinicID: 1, ScheduledDate: now, Status: status},
			repoErr: nil,
			wantErr: false,
			wantNF:  false,
		},
		{
			name: "returns error when no fields to update",
			input: UpdateAccountingInput{
				ID:       1,
				ClinicID: 1,
				// 全フィールドが nil
			},
			repoRet: nil,
			repoErr: nil,
			wantErr: true,
			wantNF:  false,
		},
		{
			name: "returns not found error when billing does not exist",
			input: UpdateAccountingInput{
				ID:       999,
				ClinicID: 1,
				Status:   &status,
			},
			repoRet: nil,
			repoErr: apperrors.WrapNotFound("billing", "999"),
			wantErr: true,
			wantNF:  true,
		},
		{
			name: "returns error on repository failure",
			input: UpdateAccountingInput{
				ID:       1,
				ClinicID: 1,
				Status:   &status,
			},
			repoRet: nil,
			repoErr: errors.New("db error"),
			wantErr: true,
			wantNF:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAccountingRepository{
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Billing, error) {
					return tt.repoRet, tt.repoErr
				},
				findByIDFn: func(_ context.Context, _ uint64, _ uint64) (*model.Billing, error) {
					return tt.repoRet, tt.repoErr
				},
			}
			svc := NewAccountingService(repo, nil, &mockTransactor{}, &mockAccountingAuditService{})

			billing, err := svc.Update(context.Background(), &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, billing)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, billing)
			}
		})
	}
}

func TestAccountingService_Update_SyncsCPMStageTagBestEffortWhenCompleted(t *testing.T) {
	ownerID := uint64(10)
	status := model.BillingStatusCompleted
	var syncedClinicID, syncedOwnerID uint64

	repo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Billing, error) {
			return &model.Billing{ID: id, ClinicID: clinicID, OwnerID: &ownerID, Status: status}, nil
		},
		updateFieldsFn: func(_ context.Context, clinicID, id uint64, _ map[string]any) (*model.Billing, error) {
			return &model.Billing{ID: id, ClinicID: clinicID, OwnerID: &ownerID, Status: status}, nil
		},
	}
	tagSync := &mockLstepTagSyncService{
		syncCPMStageTagFn: func(_ context.Context, clinicID, ownerID uint64) error {
			syncedClinicID = clinicID
			syncedOwnerID = ownerID
			return errors.New("sync failed")
		},
	}
	svc := NewAccountingService(repo, tagSync, &mockTransactor{}, &mockAccountingAuditService{})

	billing, err := svc.Update(context.Background(), &UpdateAccountingInput{
		ID:       30,
		ClinicID: 1,
		Status:   &status,
	})

	assert.NoError(t, err)
	assert.NotNil(t, billing)
	assert.Equal(t, uint64(1), syncedClinicID)
	assert.Equal(t, ownerID, syncedOwnerID)
}

func TestAccountingService_Update_CompletesSameDayAccountingAppointments(t *testing.T) {
	ownerID := uint64(10)
	petID := uint64(20)
	scheduledDate := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	status := model.BillingStatusCompleted
	var completedClinicID uint64
	var completedOwnerID uint64
	var completedPetID uint64
	var completedDate time.Time

	repo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Billing, error) {
			return &model.Billing{
				ID:            id,
				ClinicID:      clinicID,
				OwnerID:       &ownerID,
				PetID:         &petID,
				ScheduledDate: scheduledDate,
				Status:        status,
			}, nil
		},
		updateFieldsFn: func(_ context.Context, clinicID, id uint64, _ map[string]any) (*model.Billing, error) {
			return &model.Billing{
				ID:            id,
				ClinicID:      clinicID,
				OwnerID:       &ownerID,
				PetID:         &petID,
				ScheduledDate: scheduledDate,
				Status:        status,
			}, nil
		},
		completeApptsFn: func(_ context.Context, clinicID uint64, _, ownerID, petID *uint64, scheduledDate time.Time) (int64, error) {
			completedClinicID = clinicID
			completedOwnerID = *ownerID
			completedPetID = *petID
			completedDate = scheduledDate
			return 2, nil
		},
	}
	svc := NewAccountingService(repo, nil, &mockTransactor{}, &mockAccountingAuditService{})

	billing, err := svc.Update(context.Background(), &UpdateAccountingInput{
		ID:       30,
		ClinicID: 1,
		Status:   &status,
	})

	assert.NoError(t, err)
	assert.NotNil(t, billing)
	assert.Equal(t, uint64(1), completedClinicID)
	assert.Equal(t, ownerID, completedOwnerID)
	assert.Equal(t, petID, completedPetID)
	assert.Equal(t, scheduledDate, completedDate)
}

// mockAccountingAuditService は #118 用 AuditService テストモック
type mockAccountingAuditService struct {
	logEntryCalled bool
	logEntryInput  *AuditLogInput
}

func (m *mockAccountingAuditService) Log(_ context.Context, _ *model.AuditLog) error { return nil }
func (m *mockAccountingAuditService) LogEntry(_ context.Context, input *AuditLogInput) error {
	m.logEntryCalled = true
	m.logEntryInput = input
	return nil
}
func (m *mockAccountingAuditService) LogAuthLogin(_ context.Context, _, _ *uint64, _, _, _ string) error {
	return nil
}
func (m *mockAccountingAuditService) LogLstepOperation(_ context.Context, _ uint64, _ *uint64, _, _ string, _ *uint64) error {
	return nil
}
func (m *mockAccountingAuditService) LogLstepOperationWithMetadata(_ context.Context, _ uint64, _ *uint64, _, _ string, _ *uint64, _ any) error {
	return nil
}
func (m *mockAccountingAuditService) LogMedicalRecordChange(_ context.Context, _ uint64, _ *uint64, _ string, _ uint64, _, _ map[string]any) error {
	return nil
}
func (m *mockAccountingAuditService) LogVitalChange(_ context.Context, _ uint64, _ *uint64, _ string, _, _ uint64, _, _ map[string]any) error {
	return nil
}
func (m *mockAccountingAuditService) LogAddendumCreate(_ context.Context, _ uint64, _ *uint64, _, _ uint64, _ *model.MedicalRecordAddendum) error {
	return nil
}
func (m *mockAccountingAuditService) LogClinicSwitch(_ context.Context, _ *uint64, _, _ uint64, _, _ string) error {
	return nil
}

// TestAccountingService_Cancel は BUG-371: 論理削除 (status=cancelled) + #118: audit ログ記録の挙動を検証する。
func TestAccountingService_Cancel(t *testing.T) {
	actorID := uint64(42)
	tests := []struct {
		name            string
		clinicID        uint64
		id              uint64
		actorID         *uint64
		findByIDResult  *model.Billing
		findByIDErr     error
		updateErr       error
		wantErr         bool
		wantConflict    bool
		wantNF          bool
		wantAuditLogged bool
	}{
		{
			name:            "正常: waiting → cancelled に遷移する",
			clinicID:        1,
			id:              10,
			actorID:         &actorID,
			findByIDResult:  &model.Billing{ID: 10, ClinicID: 1, Status: model.BillingStatusWaiting},
			wantErr:         false,
			wantAuditLogged: true,
		},
		{
			name:            "正常: completed → cancelled に遷移する",
			clinicID:        1,
			id:              10,
			actorID:         &actorID,
			findByIDResult:  &model.Billing{ID: 10, ClinicID: 1, Status: model.BillingStatusCompleted},
			wantErr:         false,
			wantAuditLogged: true,
		},
		{
			name:            "正常: actorID nil でも audit ログが記録される",
			clinicID:        1,
			id:              10,
			actorID:         nil,
			findByIDResult:  &model.Billing{ID: 10, ClinicID: 1, Status: model.BillingStatusWaiting},
			wantErr:         false,
			wantAuditLogged: true,
		},
		{
			name:           "異常: 既に cancelled の場合は ErrConflict",
			clinicID:       1,
			id:             10,
			actorID:        &actorID,
			findByIDResult: &model.Billing{ID: 10, ClinicID: 1, Status: model.BillingStatusCancelled},
			wantErr:        true,
			wantConflict:   true,
		},
		{
			name:        "異常: 存在しない場合は ErrNotFound 経由で error",
			clinicID:    1,
			id:          999,
			actorID:     &actorID,
			findByIDErr: apperrors.WrapNotFound("billing", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name:           "異常: Update 失敗時はエラー伝播",
			clinicID:       1,
			id:             10,
			actorID:        &actorID,
			findByIDResult: &model.Billing{ID: 10, ClinicID: 1, Status: model.BillingStatusWaiting},
			updateErr:      errors.New("db error"),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAccountingRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return tt.findByIDResult, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Billing, error) {
					if tt.updateErr != nil {
						return nil, tt.updateErr
					}
					return tt.findByIDResult, nil
				},
			}
			auditSvc := &mockAccountingAuditService{}
			svc := NewAccountingService(repo, nil, &mockTransactor{}, auditSvc)

			err := svc.Cancel(context.Background(), tt.clinicID, tt.id, tt.actorID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
			} else {
				assert.NoError(t, err)
				if tt.wantAuditLogged {
					assert.True(t, auditSvc.logEntryCalled, "audit log should be called on success")
					assert.Equal(t, model.AuditActionBillingCancel, auditSvc.logEntryInput.Action)
					assert.Equal(t, "billing", auditSvc.logEntryInput.Resource)
					assert.NotNil(t, auditSvc.logEntryInput.OldValue, "cancel audit: old_value に変更前 status が必要")
					assert.NotNil(t, auditSvc.logEntryInput.NewValue, "cancel audit: new_value に変更後 status が必要")
				}
			}
		})
	}
}

func ptrInt64(v int64) *int64 { return &v }

// TestValidatePaymentSplits は validatePaymentSplits の全バリデーションブランチを検証する。
func TestValidatePaymentSplits(t *testing.T) {
	tests := []struct {
		name          string
		splits        []PaymentSplitInput
		billingAmount *int64
		wantErr       bool
		wantInvalid   bool
	}{
		{
			name:          "splits が空: バリデーションをスキップ",
			splits:        nil,
			billingAmount: ptrInt64(5000),
			wantErr:       false,
		},
		{
			name: "現金1種のみ: 有効",
			splits: []PaymentSplitInput{
				{Method: model.PaymentMethodCash, Amount: 5000, ReceivedAmount: 5000, ChangeAmount: 0},
			},
			billingAmount: ptrInt64(5000),
			wantErr:       false,
		},
		{
			name: "3種混在 (cash + credit_card + electronic_money): 有効",
			splits: []PaymentSplitInput{
				{Method: model.PaymentMethodCash, Amount: 2000, ReceivedAmount: 3000, ChangeAmount: 1000},
				{Method: model.PaymentMethodCreditCard, Amount: 1500},
				{Method: model.PaymentMethodElectronicMoney, Amount: 1500},
			},
			billingAmount: ptrInt64(5000),
			wantErr:       false,
		},
		{
			name: "支払い手段の重複: 無効",
			splits: []PaymentSplitInput{
				{Method: model.PaymentMethodCash, Amount: 2000, ReceivedAmount: 2000, ChangeAmount: 0},
				{Method: model.PaymentMethodCash, Amount: 3000, ReceivedAmount: 3000, ChangeAmount: 0},
			},
			billingAmount: ptrInt64(5000),
			wantErr:       true,
			wantInvalid:   true,
		},
		{
			name: "金額ゼロ: 無効",
			splits: []PaymentSplitInput{
				{Method: model.PaymentMethodCreditCard, Amount: 0},
			},
			billingAmount: ptrInt64(0),
			wantErr:       true,
			wantInvalid:   true,
		},
		{
			name: "合計不一致: 無効",
			splits: []PaymentSplitInput{
				{Method: model.PaymentMethodCash, Amount: 3000, ReceivedAmount: 3000, ChangeAmount: 0},
			},
			billingAmount: ptrInt64(5000),
			wantErr:       true,
			wantInvalid:   true,
		},
		{
			name: "現金: 預り金不足",
			splits: []PaymentSplitInput{
				{Method: model.PaymentMethodCash, Amount: 5000, ReceivedAmount: 4000, ChangeAmount: 0},
			},
			billingAmount: ptrInt64(5000),
			wantErr:       true,
			wantInvalid:   true,
		},
		{
			name: "現金: お釣り計算不正",
			splits: []PaymentSplitInput{
				{Method: model.PaymentMethodCash, Amount: 5000, ReceivedAmount: 6000, ChangeAmount: 500},
			},
			billingAmount: ptrInt64(5000),
			wantErr:       true,
			wantInvalid:   true,
		},
		{
			name: "billingAmount が nil: 合計チェックをスキップ",
			splits: []PaymentSplitInput{
				{Method: model.PaymentMethodCreditCard, Amount: 9999},
			},
			billingAmount: nil,
			wantErr:       false,
		},
		{
			// #127: 銀行振込は現金以外の手段として扱われ、預り金/お釣りの制約を受けず保存可能。
			name: "銀行振込1種のみ (#127): 有効",
			splits: []PaymentSplitInput{
				{Method: model.PaymentMethodBankTransfer, Amount: 5000},
			},
			billingAmount: ptrInt64(5000),
			wantErr:       false,
		},
		{
			name: "銀行振込 + 現金 混在 (#127): 有効",
			splits: []PaymentSplitInput{
				{Method: model.PaymentMethodCash, Amount: 2000, ReceivedAmount: 2000, ChangeAmount: 0},
				{Method: model.PaymentMethodBankTransfer, Amount: 3000},
			},
			billingAmount: ptrInt64(5000),
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePaymentSplits(tt.splits, tt.billingAmount)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantInvalid {
					assert.True(t, apperrors.IsInvalidInput(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRepresentativeMethod は splits から代表支払い手段を選択するロジックを検証する。
func TestRepresentativeMethod(t *testing.T) {
	tests := []struct {
		name   string
		splits []PaymentSplitInput
		want   model.PaymentMethod
	}{
		{
			name:   "cash を含む場合: cash を返す",
			splits: []PaymentSplitInput{{Method: model.PaymentMethodCash}, {Method: model.PaymentMethodCreditCard}},
			want:   model.PaymentMethodCash,
		},
		{
			name:   "cash なし credit_card あり: credit_card を返す",
			splits: []PaymentSplitInput{{Method: model.PaymentMethodCreditCard}, {Method: model.PaymentMethodElectronicMoney}},
			want:   model.PaymentMethodCreditCard,
		},
		{
			name:   "electronic_money のみ: electronic_money を返す",
			splits: []PaymentSplitInput{{Method: model.PaymentMethodElectronicMoney}},
			want:   model.PaymentMethodElectronicMoney,
		},
		{
			// #127 既知の制約 / #128 スコープ:
			// representativeMethod の優先順位は cash > credit_card > (else) electronic_money で
			// PO判断B (2026-05-25) として確定済み。bank_transfer は明示分岐を持たず else に落ちるため、
			// 銀行振込のみの会計は代表手段が electronic_money になる（一覧の代表表示はこの値を使う）。
			// payment_splits.method は bank_transfer を正しく保持する（=一次情報は保全）。
			// 代表手段の優先順位への bank_transfer 組み込みは #128 の設計判断であり本 Issue 対象外。
			// この振る舞いをここに pin し、#128 で意図的に更新されることを保証する。
			name:   "bank_transfer のみ (#128 既知制約): 代表手段は electronic_money に落ちる",
			splits: []PaymentSplitInput{{Method: model.PaymentMethodBankTransfer}},
			want:   model.PaymentMethodElectronicMoney,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, representativeMethod(tt.splits))
		})
	}
}

// TestBuildPaymentSplits は splits 変換および単一支払い backward compat を検証する。
func TestBuildPaymentSplits(t *testing.T) {
	t.Run("PaymentSplits が設定済み: そのまま変換", func(t *testing.T) {
		input := &UpdateAccountingInput{
			ID:       10,
			ClinicID: 1,
			PaymentSplits: []PaymentSplitInput{
				{Method: model.PaymentMethodCash, Amount: 2000, ReceivedAmount: 3000, ChangeAmount: 1000},
				{Method: model.PaymentMethodCreditCard, Amount: 3000},
			},
		}
		result := buildPaymentSplits(input)
		assert.Len(t, result, 2)
		assert.Equal(t, uint64(1), result[0].ClinicID)
		assert.Equal(t, uint64(10), result[0].BillingID)
		assert.Equal(t, model.PaymentMethodCash, result[0].Method)
		assert.Equal(t, int64(2000), result[0].Amount)
		assert.Equal(t, int64(3000), result[0].ReceivedAmount)
		assert.Equal(t, int64(1000), result[0].ChangeAmount)
		assert.Equal(t, model.PaymentMethodCreditCard, result[1].Method)
	})

	t.Run("PaymentSplits 空 + BillingAmount あり: 単一 split を生成 (backward compat)", func(t *testing.T) {
		input := &UpdateAccountingInput{
			ID:             10,
			ClinicID:       1,
			PaymentMethod:  func() *model.PaymentMethod { m := model.PaymentMethodCreditCard; return &m }(),
			BillingAmount:  ptrInt64(5000),
			ReceivedAmount: ptrInt64(5000),
		}
		result := buildPaymentSplits(input)
		assert.Len(t, result, 1)
		assert.Equal(t, model.PaymentMethodCreditCard, result[0].Method)
		assert.Equal(t, int64(5000), result[0].Amount)
	})

	t.Run("PaymentSplits 空 + BillingAmount nil: nil を返す", func(t *testing.T) {
		input := &UpdateAccountingInput{ID: 10, ClinicID: 1}
		result := buildPaymentSplits(input)
		assert.Nil(t, result)
	})
}

// TestAccountingService_Update_MixedPayment は3種混在支払いの全フローを検証する。
// SavePaymentSplits に渡った splits の内容と、リロード後 billing の PaymentSplits を確認する。
func TestAccountingService_Update_MixedPayment(t *testing.T) {
	billingAmount := int64(5000)
	reloadedBilling := &model.Billing{
		ID:       1,
		ClinicID: 1,
		PaymentSplits: []model.PaymentSplit{
			{ClinicID: 1, BillingID: 1, Method: model.PaymentMethodCash, Amount: 2000, ReceivedAmount: 3000, ChangeAmount: 1000},
			{ClinicID: 1, BillingID: 1, Method: model.PaymentMethodCreditCard, Amount: 1500},
			{ClinicID: 1, BillingID: 1, Method: model.PaymentMethodElectronicMoney, Amount: 1500},
		},
	}

	var capturedSplits []model.PaymentSplit
	callCount := 0
	repo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			callCount++
			if callCount == 2 {
				return reloadedBilling, nil
			}
			return &model.Billing{ID: 1, ClinicID: 1}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Billing, error) {
			return &model.Billing{ID: 1, ClinicID: 1}, nil
		},
		savePaymentSplitsFn: func(_ context.Context, splits []model.PaymentSplit) error {
			capturedSplits = splits
			return nil
		},
	}
	svc := NewAccountingService(repo, nil, &mockTransactor{}, &mockAccountingAuditService{})

	input := &UpdateAccountingInput{
		ID:            1,
		ClinicID:      1,
		BillingAmount: &billingAmount,
		PaymentSplits: []PaymentSplitInput{
			{Method: model.PaymentMethodCash, Amount: 2000, ReceivedAmount: 3000, ChangeAmount: 1000},
			{Method: model.PaymentMethodCreditCard, Amount: 1500},
			{Method: model.PaymentMethodElectronicMoney, Amount: 1500},
		},
	}

	result, err := svc.Update(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// SavePaymentSplits に3件渡されたことを確認
	assert.Len(t, capturedSplits, 3)
	assert.Equal(t, model.PaymentMethodCash, capturedSplits[0].Method)
	assert.Equal(t, int64(2000), capturedSplits[0].Amount)
	assert.Equal(t, int64(1000), capturedSplits[0].ChangeAmount)
	assert.Equal(t, model.PaymentMethodCreditCard, capturedSplits[1].Method)
	assert.Equal(t, model.PaymentMethodElectronicMoney, capturedSplits[2].Method)

	// リロード後の billing に PaymentSplits が含まれることを確認
	assert.Len(t, result.PaymentSplits, 3)
}

// TestAccountingService_GetDailySummary は日次集計取得ロジックを検証する。
func TestAccountingService_GetDailySummary(t *testing.T) {
	tests := []struct {
		name              string
		dateStr           string
		getDailySummaryFn func(ctx context.Context, clinicID uint64, date time.Time) (*repository.DailySummaryResult, error)
		wantErr           bool
		wantErrIs         error
		checkResult       func(t *testing.T, got *repository.DailySummaryResult)
	}{
		{
			name:      "エラー: 不正な日付文字列 → ErrInvalidInput",
			dateStr:   "not-a-date",
			wantErr:   true,
			wantErrIs: apperrors.ErrInvalidInput,
		},
		{
			name:    "エラー: repo がエラーを返す → ラップされたエラー",
			dateStr: "2026-05-01",
			getDailySummaryFn: func(_ context.Context, _ uint64, _ time.Time) (*repository.DailySummaryResult, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:    "正常: 空文字列 → today をデフォルト使用、エラーなし",
			dateStr: "",
			getDailySummaryFn: func(_ context.Context, _ uint64, _ time.Time) (*repository.DailySummaryResult, error) {
				return &repository.DailySummaryResult{
					PaymentTotals:  []repository.PaymentMethodTotal{},
					CategoryTotals: []repository.CategoryTotal{},
					BillingCount:   0,
					GrandTotal:     0,
				}, nil
			},
			checkResult: func(t *testing.T, got *repository.DailySummaryResult) {
				assert.NotNil(t, got)
				assert.Equal(t, int64(0), got.GrandTotal)
			},
		},
		{
			name:    "正常: 3種混在支払い → PaymentTotals が支払方法別に正しく返される",
			dateStr: "2026-05-01",
			getDailySummaryFn: func(_ context.Context, _ uint64, _ time.Time) (*repository.DailySummaryResult, error) {
				return &repository.DailySummaryResult{
					PaymentTotals: []repository.PaymentMethodTotal{
						{Method: "現金", Total: 5000},
						{Method: "クレジットカード", Total: 3000},
						{Method: "電子マネー", Total: 2000},
					},
					CategoryTotals: []repository.CategoryTotal{
						{Category: "診察", Total: 10000},
					},
					BillingCount: 3,
					GrandTotal:   10000,
				}, nil
			},
			checkResult: func(t *testing.T, got *repository.DailySummaryResult) {
				assert.Len(t, got.PaymentTotals, 3)
				assert.Equal(t, "現金", got.PaymentTotals[0].Method)
				assert.Equal(t, int64(5000), got.PaymentTotals[0].Total)
				assert.Equal(t, "クレジットカード", got.PaymentTotals[1].Method)
				assert.Equal(t, int64(3000), got.PaymentTotals[1].Total)
				assert.Equal(t, "電子マネー", got.PaymentTotals[2].Method)
				assert.Equal(t, int64(2000), got.PaymentTotals[2].Total)
				assert.Len(t, got.CategoryTotals, 1)
				assert.Equal(t, "診察", got.CategoryTotals[0].Category)
				assert.Equal(t, int64(10000), got.CategoryTotals[0].Total)
				assert.Equal(t, int64(3), got.BillingCount)
				assert.Equal(t, int64(10000), got.GrandTotal)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAccountingRepository{
				getDailySummaryFn: tt.getDailySummaryFn,
			}
			svc := NewAccountingService(repo, nil, &mockTransactor{}, &mockAccountingAuditService{})

			got, err := svc.GetDailySummary(context.Background(), 1, tt.dateStr)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrIs != nil {
					assert.True(t, errors.Is(err, tt.wantErrIs), "want errors.Is(%v), got %v", tt.wantErrIs, err)
				}
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, got)
			if tt.checkResult != nil {
				tt.checkResult(t, got)
			}
		})
	}
}

// TestAccountingService_ListUnpaidByBilling は #120: start_date/end_date 2引数仕様を検証する。
func TestAccountingService_ListUnpaidByBilling(t *testing.T) {
	tests := []struct {
		name        string
		startDate   string
		endDate     string
		repoResults []model.Billing
		repoTotal   int64
		repoErr     error
		wantErr     bool
	}{
		{
			name:      "正常: start_date/end_date でリポジトリに渡される",
			startDate: "2026-01-01",
			endDate:   "2026-01-31",
			repoResults: []model.Billing{
				{ID: 1, ClinicID: 1, Status: model.BillingStatusWaiting},
			},
			repoTotal: 1,
		},
		{
			name:      "正常: 結果0件でも正常",
			startDate: "2026-02-01",
			endDate:   "2026-02-28",
		},
		{
			name:      "異常: repo エラーが伝播する",
			startDate: "2026-01-01",
			endDate:   "2026-01-31",
			repoErr:   errors.New("db error"),
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedStart, capturedEnd string
			repo := &mockAccountingRepository{
				findUnpaidByBillingFn: func(_ context.Context, _ uint64, start, end string, _, _ int) ([]model.Billing, int64, error) {
					capturedStart = start
					capturedEnd = end
					return tt.repoResults, tt.repoTotal, tt.repoErr
				},
			}
			svc := NewAccountingService(repo, nil, &mockTransactor{}, &mockAccountingAuditService{})

			result, total, err := svc.ListUnpaidByBilling(context.Background(), 1, tt.startDate, tt.endDate, 1, 20)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.startDate, capturedStart)
				assert.Equal(t, tt.endDate, capturedEnd)
				assert.Len(t, result, len(tt.repoResults))
				assert.Equal(t, tt.repoTotal, total)
			}
		})
	}
}

// TestAccountingService_ListUnpaidByOwner は #120: start_date/end_date 2引数仕様を検証する。
func TestAccountingService_ListUnpaidByOwner(t *testing.T) {
	tests := []struct {
		name      string
		startDate string
		endDate   string
		repoAggs  []repository.UnpaidOwnerAggregate
		repoTotal int64
		repoErr   error
		wantErr   bool
	}{
		{
			name:      "正常: start_date/end_date でリポジトリに渡される",
			startDate: "2026-01-01",
			endDate:   "2026-01-31",
			repoAggs: []repository.UnpaidOwnerAggregate{
				{OwnerID: 1, OwnerName: "田中太郎", Count: 2, TotalAmount: 5000},
			},
			repoTotal: 1,
		},
		{
			name:      "異常: repo エラーが伝播する",
			startDate: "2026-01-01",
			endDate:   "2026-01-31",
			repoErr:   errors.New("db error"),
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedStart, capturedEnd string
			repo := &mockAccountingRepository{
				findUnpaidByOwnerFn: func(_ context.Context, _ uint64, start, end string, _, _ int) ([]repository.UnpaidOwnerAggregate, int64, repository.UnpaidSummary, error) {
					capturedStart = start
					capturedEnd = end
					return tt.repoAggs, tt.repoTotal, repository.UnpaidSummary{}, tt.repoErr
				},
			}
			svc := NewAccountingService(repo, nil, &mockTransactor{}, &mockAccountingAuditService{})

			result, total, _, err := svc.ListUnpaidByOwner(context.Background(), 1, tt.startDate, tt.endDate, 1, 20)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.startDate, capturedStart)
				assert.Equal(t, tt.endDate, capturedEnd)
				assert.Len(t, result, len(tt.repoAggs))
				assert.Equal(t, tt.repoTotal, total)
			}
		})
	}
}
