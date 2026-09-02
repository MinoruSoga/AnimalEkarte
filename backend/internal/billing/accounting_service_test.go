package billing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func ptrString(v string) *string { return &v }

// TestAccountingService_GetMonthlyUnpaidCarryover は月次未納繰越集計サービスメソッドのテスト。#114
func TestAccountingService_GetMonthlyUnpaidCarryover(t *testing.T) {
	petID := uint64(3)
	tests := []struct {
		name        string
		year        int
		month       int
		mockFn      func(ctx context.Context, clinicID uint64, firstDay, lastDay string, page, limit int) ([]MonthlyUnpaidOwnerPet, int64, MonthlyUnpaidSummary, error)
		wantSummary MonthlyUnpaidSummary
		wantTotal   int64
		wantLen     int
		wantErr     bool
	}{
		{
			name:  "正常: firstDay/lastDay が正しく計算されデータが返る",
			year:  2026,
			month: 6,
			mockFn: func(_ context.Context, _ uint64, firstDay, lastDay string, _, _ int) ([]MonthlyUnpaidOwnerPet, int64, MonthlyUnpaidSummary, error) {
				if firstDay != "2026-06-01" || lastDay != "2026-06-30" {
					t.Errorf("want firstDay=2026-06-01 lastDay=2026-06-30, got firstDay=%s lastDay=%s", firstDay, lastDay)
				}
				items := []MonthlyUnpaidOwnerPet{
					{OwnerID: 1, OwnerName: "田中", PetID: &petID, PetName: "ポチ", PrevMonthCarryover: 10000, CurrentMonthUnpaid: 5000, NextMonthCarryover: 15000},
					{OwnerID: 2, OwnerName: "鈴木", PetID: nil, PetName: "", PrevMonthCarryover: 0, CurrentMonthUnpaid: 3000, NextMonthCarryover: 3000},
				}
				return items, 2, MonthlyUnpaidSummary{PrevMonthCarryover: 10000, CurrentMonthUnpaid: 8000, NextMonthCarryover: 18000}, nil
			},
			wantSummary: MonthlyUnpaidSummary{PrevMonthCarryover: 10000, CurrentMonthUnpaid: 8000, NextMonthCarryover: 18000},
			wantTotal:   2,
			wantLen:     2,
		},
		{
			name:  "正常: 1月（firstDay=01-01, lastDay=01-31）",
			year:  2026,
			month: 1,
			mockFn: func(_ context.Context, _ uint64, firstDay, lastDay string, _, _ int) ([]MonthlyUnpaidOwnerPet, int64, MonthlyUnpaidSummary, error) {
				if firstDay != "2026-01-01" || lastDay != "2026-01-31" {
					t.Errorf("want firstDay=2026-01-01 lastDay=2026-01-31, got firstDay=%s lastDay=%s", firstDay, lastDay)
				}
				return nil, 0, MonthlyUnpaidSummary{}, nil
			},
			wantSummary: MonthlyUnpaidSummary{},
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
			mockFn: func(_ context.Context, _ uint64, _, _ string, _, _ int) ([]MonthlyUnpaidOwnerPet, int64, MonthlyUnpaidSummary, error) {
				return nil, 0, MonthlyUnpaidSummary{}, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockAccountingRepository{findMonthlyUnpaidCarryoverFn: tt.mockFn}
			svc := NewAccountingService(mock, nil, nil, nil, nil, nil, nil, &mockPaymentMethodMasterRepository{})

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
				findAllFn: func(_ context.Context, _ uint64, _ AccountingListFilters, _, _ int) ([]model.Billing, int64, error) {
					return tt.repoBillings, tt.repoTotal, tt.repoErr
				},
			}
			svc := NewAccountingService(repo, nil, nil, &mockReservationRepository{}, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})

			billings, total, err := svc.List(context.Background(), tt.clinicID, AccountingListFilters{PetID: tt.petID, OwnerID: tt.ownerID, Status: tt.status}, tt.page, tt.limit)

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
			svc := NewAccountingService(repo, nil, nil, &mockReservationRepository{}, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})

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
			svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})

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

func TestAccountingService_Create_RejectsCompletedStatus(t *testing.T) {
	repo := &mockAccountingRepository{
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			t.Fatal("create must not run for completed status")
			return nil
		},
	}
	svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})

	billing, err := svc.Create(context.Background(), &CreateAccountingInput{
		ClinicID:      1,
		Status:        model.BillingStatusCompleted,
		ScheduledDate: time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC),
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.ErrorContains(t, err, "POST /accountings/complete")
	assert.Nil(t, billing)
}

func TestAccountingService_Create_RejectsClientCompletedAt(t *testing.T) {
	completedAt := time.Date(2026, 5, 28, 10, 0, 0, 0, time.UTC)
	repo := &mockAccountingRepository{
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			t.Fatal("create must not run when completed_at is client-supplied")
			return nil
		},
	}
	svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})

	billing, err := svc.Create(context.Background(), &CreateAccountingInput{
		ClinicID:      1,
		Status:        model.BillingStatusWaiting,
		ScheduledDate: time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC),
		CompletedAt:   &completedAt,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.ErrorContains(t, err, "completed_at")
	assert.Nil(t, billing)
}

func TestAccountingService_Update(t *testing.T) {
	now := time.Now()
	memo := "updated"
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
				Memo:          &memo,
			},
			repoRet: &model.Billing{ID: 1, ClinicID: 1, ScheduledDate: now, Status: model.BillingStatusWaiting},
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
			repoRet: &model.Billing{ID: 1, ClinicID: 1, Status: model.BillingStatusWaiting},
			repoErr: nil,
			wantErr: true,
			wantNF:  false,
		},
		{
			name: "returns not found error when billing does not exist",
			input: UpdateAccountingInput{
				ID:       999,
				ClinicID: 1,
				Memo:     &memo,
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
				Memo:     &memo,
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
			svc := NewAccountingService(repo, nil, nil, &mockReservationRepository{}, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})

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

func TestAccountingService_Update_ReloadFailureRollsBackTransaction(t *testing.T) {
	type txMarkerKey struct{}
	reloadErr := errors.New("reload failed")
	memo := "reload"
	findCalls := 0
	reloadUsedTx := false
	var callbackErr error
	repo := &mockAccountingRepository{
		findByIDFn: func(ctx context.Context, _, _ uint64) (*model.Billing, error) {
			findCalls++
			if findCalls == 1 {
				return &model.Billing{ID: 1, ClinicID: 1, Status: model.BillingStatusWaiting}, nil
			}
			reloadUsedTx = ctx.Value(txMarkerKey{}) == true
			return nil, reloadErr
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Billing, error) {
			return &model.Billing{ID: 1, ClinicID: 1, Status: model.BillingStatusWaiting}, nil
		},
	}
	tx := &mockTransactor{
		withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
			callbackErr = fn(context.WithValue(ctx, txMarkerKey{}, true))
			return callbackErr
		},
	}
	svc := NewAccountingService(repo, nil, nil, &mockReservationRepository{}, nil, tx, &mockAuditService{}, &mockPaymentMethodMasterRepository{})

	billing, err := svc.Update(context.Background(), &UpdateAccountingInput{
		ID:       1,
		ClinicID: 1,
		Memo:     &memo,
	})

	assert.ErrorIs(t, err, reloadErr)
	assert.ErrorIs(t, callbackErr, reloadErr, "reload failure must abort the transaction callback")
	assert.True(t, reloadUsedTx, "final reload must use the ambient transaction")
	assert.Nil(t, billing)
}

func TestAccountingService_Update_RejectsCompletedStatusTransition(t *testing.T) {
	status := model.BillingStatusCompleted
	var synced bool
	repo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Billing, error) {
			return &model.Billing{ID: id, ClinicID: clinicID, Status: model.BillingStatusWaiting}, nil
		},
		updateFieldsFn: func(_ context.Context, clinicID, id uint64, _ map[string]any) (*model.Billing, error) {
			t.Fatal("update must not persist completed via generic PATCH")
			return &model.Billing{ID: id, ClinicID: clinicID}, nil
		},
	}
	tagSync := &mockLstepTagSyncService{
		syncCPMStageTagFn: func(_ context.Context, _, _ uint64) error {
			synced = true
			return nil
		},
	}
	svc := NewAccountingService(repo, nil, nil, &mockReservationRepository{}, tagSync, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})

	billing, err := svc.Update(context.Background(), &UpdateAccountingInput{
		ID:       30,
		ClinicID: 1,
		Status:   &status,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, billing)
	assert.False(t, synced)
}

func TestAccountingService_Update_RejectsCancelledStatusTransition(t *testing.T) {
	status := model.BillingStatusCancelled
	repo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Billing, error) {
			return &model.Billing{ID: id, ClinicID: clinicID, Status: model.BillingStatusWaiting}, nil
		},
		updateFieldsFn: func(_ context.Context, clinicID, id uint64, _ map[string]any) (*model.Billing, error) {
			t.Fatal("update must not persist cancelled via generic PATCH")
			return &model.Billing{ID: id, ClinicID: clinicID}, nil
		},
	}
	svc := NewAccountingService(repo, nil, nil, &mockReservationRepository{}, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})

	billing, err := svc.Update(context.Background(), &UpdateAccountingInput{
		ID:       30,
		ClinicID: 1,
		Status:   &status,
	})

	assert.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
	assert.Nil(t, billing)
}

func TestAccountingService_Update_RejectsCancelledResurrection(t *testing.T) {
	waiting := model.BillingStatusWaiting
	memo := "resurrect"
	repo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Billing, error) {
			return &model.Billing{ID: id, ClinicID: clinicID, Status: model.BillingStatusCancelled}, nil
		},
		updateFieldsFn: func(_ context.Context, clinicID, id uint64, _ map[string]any) (*model.Billing, error) {
			t.Fatal("cancelled billing must not be updated")
			return &model.Billing{ID: id, ClinicID: clinicID}, nil
		},
	}
	svc := NewAccountingService(repo, nil, nil, &mockReservationRepository{}, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})

	billing, err := svc.Update(context.Background(), &UpdateAccountingInput{
		ID:       30,
		ClinicID: 1,
		Status:   &waiting,
		Memo:     &memo,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.Nil(t, billing)
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
			auditSvc := &mockAuditService{}
			svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, auditSvc, &mockPaymentMethodMasterRepository{})

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
					// BE-refactor.md R1-2: Cancel は ambient tx 内で LogEntryTx を呼ぶよう移行済み（fail-closed）。
					assert.True(t, auditSvc.logEntryTxCalled, "audit log should be called on success")
					assert.Equal(t, model.AuditActionBillingCancel, auditSvc.logEntryTxInput.Action)
					assert.Equal(t, "billing", auditSvc.logEntryTxInput.Resource)
					assert.NotNil(t, auditSvc.logEntryTxInput.OldValue, "cancel audit: old_value に変更前 status が必要")
					assert.NotNil(t, auditSvc.logEntryTxInput.NewValue, "cancel audit: new_value に変更後 status が必要")
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
			// #188: お釣り直接上書きモード。レジ実機の誤差吸収のため change==received-amount 整合を緩和する。
			// received(6000) - amount(5000) = 1000 だが、実機が誤って 500 を返した現実を記録できる。
			name: "現金: お釣り直接上書き（整合不一致でも有効・#188）",
			splits: []PaymentSplitInput{
				{Method: model.PaymentMethodCash, Amount: 5000, ReceivedAmount: 6000, ChangeAmount: 500, ChangeOverride: true},
			},
			billingAmount: ptrInt64(5000),
			wantErr:       false,
		},
		{
			// #188: 上書きでも下限ガード（received >= amount）は維持する。過少入金は弾く。
			name: "現金: お釣り上書きでも預り金不足は弾く（#188）",
			splits: []PaymentSplitInput{
				{Method: model.PaymentMethodCash, Amount: 5000, ReceivedAmount: 4000, ChangeAmount: 0, ChangeOverride: true},
			},
			billingAmount: ptrInt64(5000),
			wantErr:       true,
			wantInvalid:   true,
		},
		{
			// #188: 上書きでも下限ガード（change >= 0）は維持する。負のお釣りは弾く。
			name: "現金: お釣り上書きでも負のお釣りは弾く（#188）",
			splits: []PaymentSplitInput{
				{Method: model.PaymentMethodCash, Amount: 5000, ReceivedAmount: 6000, ChangeAmount: -100, ChangeOverride: true},
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
			name:   "bank_transfer のみ: bank_transfer を返す",
			splits: []PaymentSplitInput{{Method: model.PaymentMethodBankTransfer}},
			want:   model.PaymentMethodBankTransfer,
		},
		{
			name:   "bank_transfer + electronic_money: bank_transfer を返す",
			splits: []PaymentSplitInput{{Method: model.PaymentMethodBankTransfer}, {Method: model.PaymentMethodElectronicMoney}},
			want:   model.PaymentMethodBankTransfer,
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
	var capturedPayment *model.Payment
	callCount := 0
	repo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			callCount++
			if callCount == 2 {
				return reloadedBilling, nil
			}
			return &model.Billing{ID: 1, ClinicID: 1, Status: model.BillingStatusCompleted}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Billing, error) {
			return &model.Billing{ID: 1, ClinicID: 1}, nil
		},
		savePaymentSplitsFn: func(_ context.Context, splits []model.PaymentSplit) error {
			capturedSplits = splits
			return nil
		},
		savePaymentFn: func(_ context.Context, payment *model.Payment) error {
			capturedPayment = payment
			return nil
		},
	}
	svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, seededPayMethodMock())

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

	// #128: 各 split に当該 clinic の master id が解決されて保存されること（NULL のまま現金に倒れない）
	if assert.NotNil(t, capturedSplits[0].PaymentMethodID) {
		assert.Equal(t, uint64(101), *capturedSplits[0].PaymentMethodID, "現金 → 現金 master id")
	}
	if assert.NotNil(t, capturedSplits[1].PaymentMethodID) {
		assert.Equal(t, uint64(102), *capturedSplits[1].PaymentMethodID, "credit_card → クレジットカード master id")
	}
	if assert.NotNil(t, capturedSplits[2].PaymentMethodID) {
		assert.Equal(t, uint64(103), *capturedSplits[2].PaymentMethodID, "electronic_money → 電子マネー master id")
	}

	// 代表支払方法（payments 行）も master id を併設すること（dual maintain）
	if assert.NotNil(t, capturedPayment) && assert.NotNil(t, capturedPayment.PaymentMethodID) {
		assert.Equal(t, uint64(101), *capturedPayment.PaymentMethodID, "代表手段 cash → 現金 master id")
	}

	// リロード後の billing に PaymentSplits が含まれることを確認
	assert.Len(t, result.PaymentSplits, 3)
}

// seededPayMethodMock は標準4支払方法を返す payment_methods マスタモック（#128/#197 解決テスト用）。
// id は固定: 現金=101 / クレジットカード=102 / 電子マネー=103 / 銀行振込=104。
// system_key (#197) をセットし rename 耐性を持つ。
func seededPayMethodMock() *mockPaymentMethodMasterRepository {
	ptr := func(s string) *string { return &s }
	return &mockPaymentMethodMasterRepository{
		findAllFn: func(_ context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error) {
			return []model.PaymentMethodMaster{
				{ID: 101, ClinicID: clinicID, Name: "現金", SystemKey: ptr("cash"), DisplayOrder: 1, IsActive: true},
				{ID: 102, ClinicID: clinicID, Name: "クレジットカード", SystemKey: ptr("credit_card"), DisplayOrder: 2, IsActive: true},
				{ID: 103, ClinicID: clinicID, Name: "電子マネー", SystemKey: ptr("electronic_money"), DisplayOrder: 3, IsActive: true},
				{ID: 104, ClinicID: clinicID, Name: "銀行振込", SystemKey: ptr("bank_transfer"), DisplayOrder: 4, IsActive: true},
			}, nil
		},
	}
}

// renamedPayMethodMock は標準4支払方法の name を変更したモック（#197 rename 耐性テスト用）。
// system_key は正しいまま name のみ改名 → resolvePaymentMethodMasterID が成功することを検証する。
func renamedPayMethodMock() *mockPaymentMethodMasterRepository {
	ptr := func(s string) *string { return &s }
	return &mockPaymentMethodMasterRepository{
		findAllFn: func(_ context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error) {
			return []model.PaymentMethodMaster{
				{ID: 101, ClinicID: clinicID, Name: "お金", SystemKey: ptr("cash"), DisplayOrder: 1, IsActive: true},
				{ID: 102, ClinicID: clinicID, Name: "カード", SystemKey: ptr("credit_card"), DisplayOrder: 2, IsActive: true},
				{ID: 103, ClinicID: clinicID, Name: "電子決済", SystemKey: ptr("electronic_money"), DisplayOrder: 3, IsActive: true},
				{ID: 104, ClinicID: clinicID, Name: "振込", SystemKey: ptr("bank_transfer"), DisplayOrder: 4, IsActive: true},
			}, nil
		},
	}
}

// TestAccountingService_Update_ResolvesPaymentMethodID は #128 hotfix の中核を検証する:
// 会計完了の書込み経路で、各 method(ENUM) が当該 clinic の payment_methods master id に解決され、
// payment_splits.payment_method_id が NULL のまま保存されない（集計で現金に倒れない）こと。
func TestAccountingService_Update_ResolvesPaymentMethodID(t *testing.T) {
	billingAmount := int64(5000)

	newRepo := func(captured *[]model.PaymentSplit) *mockAccountingRepository {
		return &mockAccountingRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
				return &model.Billing{ID: 1, ClinicID: 1, Status: model.BillingStatusCompleted}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Billing, error) {
				return &model.Billing{ID: 1, ClinicID: 1, Status: model.BillingStatusCompleted}, nil
			},
			savePaymentSplitsFn: func(_ context.Context, splits []model.PaymentSplit) error {
				*captured = splits
				return nil
			},
		}
	}

	t.Run("全 method が当該 clinic の master id に解決される（cash/credit_card/electronic_money/bank_transfer）", func(t *testing.T) {
		cases := []struct {
			method model.PaymentMethod
			wantID uint64
		}{
			{model.PaymentMethodCash, 101},
			{model.PaymentMethodCreditCard, 102},
			{model.PaymentMethodElectronicMoney, 103},
			{model.PaymentMethodBankTransfer, 104},
		}
		for _, c := range cases {
			t.Run(string(c.method), func(t *testing.T) {
				var captured []model.PaymentSplit
				svc := NewAccountingService(newRepo(&captured), nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, seededPayMethodMock())

				received, change := int64(0), int64(0)
				if c.method == model.PaymentMethodCash {
					received = 5000
				}
				_, err := svc.Update(context.Background(), &UpdateAccountingInput{
					ID:            1,
					ClinicID:      1,
					BillingAmount: &billingAmount,
					PaymentSplits: []PaymentSplitInput{
						{Method: c.method, Amount: 5000, ReceivedAmount: received, ChangeAmount: change},
					},
				})

				assert.NoError(t, err)
				assert.Len(t, captured, 1)
				if assert.NotNil(t, captured[0].PaymentMethodID, "payment_method_id は NULL にならない") {
					assert.Equal(t, c.wantID, *captured[0].PaymentMethodID)
				}
			})
		}
	})

	t.Run("明示供給 id が当該 clinic の当該 method マスタと一致する場合は許可する", func(t *testing.T) {
		var captured []model.PaymentSplit
		svc := NewAccountingService(newRepo(&captured), nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, seededPayMethodMock())

		matching := uint64(102) // credit_card = クレジットカード = 102
		_, err := svc.Update(context.Background(), &UpdateAccountingInput{
			ID:            1,
			ClinicID:      1,
			BillingAmount: &billingAmount,
			PaymentSplits: []PaymentSplitInput{
				{Method: model.PaymentMethodCreditCard, PaymentMethodID: &matching, Amount: 5000},
			},
		})

		assert.NoError(t, err)
		assert.Len(t, captured, 1)
		if assert.NotNil(t, captured[0].PaymentMethodID) {
			assert.Equal(t, uint64(102), *captured[0].PaymentMethodID)
		}
	})

	t.Run("明示供給 id が method/clinic と矛盾する場合は拒否する（他クリニック id 混入防止）", func(t *testing.T) {
		var captured []model.PaymentSplit
		svc := NewAccountingService(newRepo(&captured), nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, seededPayMethodMock())

		foreign := uint64(999) // 当該 clinic の credit_card master(102) と不一致
		_, err := svc.Update(context.Background(), &UpdateAccountingInput{
			ID:            1,
			ClinicID:      1,
			BillingAmount: &billingAmount,
			PaymentSplits: []PaymentSplitInput{
				{Method: model.PaymentMethodCreditCard, PaymentMethodID: &foreign, Amount: 5000},
			},
		})

		assert.Error(t, err, "method と矛盾する id は拒否")
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Nil(t, captured, "拒否時は保存しない")
	})

	t.Run("master が欠落している場合は明示エラーになり、現金に倒さず保存しない", func(t *testing.T) {
		var captured []model.PaymentSplit
		emptyMaster := &mockPaymentMethodMasterRepository{
			findAllFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return []model.PaymentMethodMaster{}, nil // 現金マスタが存在しない設定不整合
			},
		}
		svc := NewAccountingService(newRepo(&captured), nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, emptyMaster)

		_, err := svc.Update(context.Background(), &UpdateAccountingInput{
			ID:            1,
			ClinicID:      1,
			BillingAmount: &billingAmount,
			PaymentSplits: []PaymentSplitInput{
				{Method: model.PaymentMethodCreditCard, Amount: 5000},
			},
		})

		assert.Error(t, err, "master 欠落時は明示エラー")
		assert.Nil(t, captured, "解決失敗時は SavePaymentSplits を呼ばない（NULL/現金で保存しない）")
	})
}

// TestAccountingService_Update_ResolvesPaymentMethodID_RenameResilient は #197 system_key 導入後の
// rename 耐性を検証する: 支払方法 name を改名しても system_key ベースで master id に正しく解決される。
func TestAccountingService_Update_ResolvesPaymentMethodID_RenameResilient(t *testing.T) {
	billingAmount := int64(5000)

	newRepo := func(captured *[]model.PaymentSplit) *mockAccountingRepository {
		return &mockAccountingRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
				return &model.Billing{ID: 1, ClinicID: 1, Status: model.BillingStatusCompleted}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Billing, error) {
				return &model.Billing{ID: 1, ClinicID: 1, Status: model.BillingStatusCompleted}, nil
			},
			savePaymentSplitsFn: func(_ context.Context, splits []model.PaymentSplit) error {
				*captured = splits
				return nil
			},
		}
	}

	cases := []struct {
		method model.PaymentMethod
		wantID uint64
	}{
		{model.PaymentMethodCash, 101},
		{model.PaymentMethodCreditCard, 102},
		{model.PaymentMethodElectronicMoney, 103},
		{model.PaymentMethodBankTransfer, 104},
	}
	for _, c := range cases {
		t.Run("name改名後も system_key で解決: "+string(c.method), func(t *testing.T) {
			var captured []model.PaymentSplit
			svc := NewAccountingService(newRepo(&captured), nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, renamedPayMethodMock())

			received, change := int64(0), int64(0)
			if c.method == model.PaymentMethodCash {
				received = 5000
			}
			_, err := svc.Update(context.Background(), &UpdateAccountingInput{
				ID:            1,
				ClinicID:      1,
				BillingAmount: &billingAmount,
				PaymentSplits: []PaymentSplitInput{
					{Method: c.method, Amount: 5000, ReceivedAmount: received, ChangeAmount: change},
				},
			})

			assert.NoError(t, err)
			assert.Len(t, captured, 1)
			if assert.NotNil(t, captured[0].PaymentMethodID, "name 改名後も payment_method_id は NULL にならない") {
				assert.Equal(t, c.wantID, *captured[0].PaymentMethodID)
			}
		})
	}
}

// TestAccountingService_GetDailySummary は日次集計取得ロジックを検証する。
func TestAccountingService_GetDailySummary(t *testing.T) {
	tests := []struct {
		name              string
		dateStr           string
		getDailySummaryFn func(ctx context.Context, clinicID uint64, date time.Time) (*DailySummaryResult, error)
		wantErr           bool
		wantErrIs         error
		checkResult       func(t *testing.T, got *DailySummaryResult)
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
			getDailySummaryFn: func(_ context.Context, _ uint64, _ time.Time) (*DailySummaryResult, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:    "正常: 空文字列 → today をデフォルト使用、エラーなし",
			dateStr: "",
			getDailySummaryFn: func(_ context.Context, _ uint64, _ time.Time) (*DailySummaryResult, error) {
				return &DailySummaryResult{
					PaymentTotals:  []PaymentMethodTotal{},
					CategoryTotals: []CategoryTotal{},
					BillingCount:   0,
					GrandTotal:     0,
				}, nil
			},
			checkResult: func(t *testing.T, got *DailySummaryResult) {
				assert.NotNil(t, got)
				assert.Equal(t, int64(0), got.GrandTotal)
			},
		},
		{
			name:    "正常: 3種混在支払い → PaymentTotals が支払方法別に正しく返される",
			dateStr: "2026-05-01",
			getDailySummaryFn: func(_ context.Context, _ uint64, _ time.Time) (*DailySummaryResult, error) {
				return &DailySummaryResult{
					PaymentTotals: []PaymentMethodTotal{
						{Method: "現金", Total: 5000},
						{Method: "クレジットカード", Total: 3000},
						{Method: "電子マネー", Total: 2000},
					},
					CategoryTotals: []CategoryTotal{
						{Category: string(model.ItemCategoryExamination), Total: 10000},
					},
					BillingCount: 3,
					GrandTotal:   10000,
				}, nil
			},
			checkResult: func(t *testing.T, got *DailySummaryResult) {
				assert.Len(t, got.PaymentTotals, 3)
				assert.Equal(t, "現金", got.PaymentTotals[0].Method)
				assert.Equal(t, int64(5000), got.PaymentTotals[0].Total)
				assert.Equal(t, "クレジットカード", got.PaymentTotals[1].Method)
				assert.Equal(t, int64(3000), got.PaymentTotals[1].Total)
				assert.Equal(t, "電子マネー", got.PaymentTotals[2].Method)
				assert.Equal(t, int64(2000), got.PaymentTotals[2].Total)
				assert.Len(t, got.CategoryTotals, 1)
				assert.Equal(t, string(model.ItemCategoryExamination), got.CategoryTotals[0].Category)
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
			svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})

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
			svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})

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
		repoAggs  []UnpaidOwnerAggregate
		repoTotal int64
		repoErr   error
		wantErr   bool
	}{
		{
			name:      "正常: start_date/end_date でリポジトリに渡される",
			startDate: "2026-01-01",
			endDate:   "2026-01-31",
			repoAggs: []UnpaidOwnerAggregate{
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
				findUnpaidByOwnerFn: func(_ context.Context, _ uint64, start, end string, _, _ int) ([]UnpaidOwnerAggregate, int64, UnpaidSummary, error) {
					capturedStart = start
					capturedEnd = end
					return tt.repoAggs, tt.repoTotal, UnpaidSummary{}, tt.repoErr
				},
			}
			svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})

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

// TestGetOwnerUnpaidBalance は #182 の未納残高取得サービスを検証する。
func TestGetOwnerUnpaidBalance(t *testing.T) {
	t.Run("owner_id=0 は 400（リポジトリを呼ばない）", func(t *testing.T) {
		called := false
		repo := &mockAccountingRepository{
			sumUnpaidByOwnerFn: func(_ context.Context, _, _ uint64) (OwnerUnpaidBalance, error) {
				called = true
				return OwnerUnpaidBalance{}, nil
			},
		}
		svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})
		_, err := svc.GetOwnerUnpaidBalance(context.Background(), 1, 0)
		assert.Error(t, err)
		assert.False(t, called, "owner_id=0 ではリポジトリを呼ばない")
	})

	t.Run("リポジトリ結果をそのまま返す", func(t *testing.T) {
		var gotClinic, gotOwner uint64
		repo := &mockAccountingRepository{
			sumUnpaidByOwnerFn: func(_ context.Context, clinicID, ownerID uint64) (OwnerUnpaidBalance, error) {
				gotClinic, gotOwner = clinicID, ownerID
				return OwnerUnpaidBalance{TotalAmount: 2100, Count: 2}, nil
			},
		}
		svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})
		got, err := svc.GetOwnerUnpaidBalance(context.Background(), 7, 42)
		assert.NoError(t, err)
		assert.Equal(t, int64(2100), got.TotalAmount)
		assert.Equal(t, int64(2), got.Count)
		assert.Equal(t, uint64(7), gotClinic)
		assert.Equal(t, uint64(42), gotOwner)
	})

	t.Run("リポジトリエラーはラップして返す", func(t *testing.T) {
		repo := &mockAccountingRepository{
			sumUnpaidByOwnerFn: func(_ context.Context, _, _ uint64) (OwnerUnpaidBalance, error) {
				return OwnerUnpaidBalance{}, errors.New("db error")
			},
		}
		svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})
		_, err := svc.GetOwnerUnpaidBalance(context.Background(), 1, 42)
		assert.Error(t, err)
	})
}

// TestAccountingService_Update_PostCloseReasonRequired は #115 / B4 の service 層 enforcement を検証する。
// handler を経由せず Update を直接呼んだ場合でも、締め後編集（IsPostClose=true）で
// post_close_reason を欠くと拒否される（サイレント通過しない）ことを示す。
func TestAccountingService_Update_PostCloseReasonRequired(t *testing.T) {
	emptyReason := ""
	validReason := "訂正のため"

	tests := []struct {
		name        string
		input       *UpdateAccountingInput
		wantErrText string
	}{
		{
			name:        "post-close edit without reason (nil) is rejected at service boundary",
			input:       &UpdateAccountingInput{ID: 1, ClinicID: 1, IsPostClose: true, PostCloseReason: nil},
			wantErrText: "レジ締め済み期間の会計編集には post_close_reason の入力が必要です",
		},
		{
			name:        "post-close edit with empty reason is rejected at service boundary",
			input:       &UpdateAccountingInput{ID: 1, ClinicID: 1, IsPostClose: true, PostCloseReason: &emptyReason},
			wantErrText: "レジ締め済み期間の会計編集には post_close_reason の入力が必要です",
		},
		{
			// 締め後＋理由ありはガードを通過する。本ケースは他フィールド未指定のため
			// 後続の "no fields to update" 検証で停止する＝理由ガードを越えたことを示す。
			name:        "post-close edit with valid reason passes the reason guard",
			input:       &UpdateAccountingInput{ID: 1, ClinicID: 1, IsPostClose: true, PostCloseReason: &validReason},
			wantErrText: "no fields to update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAccountingRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
					return &model.Billing{ID: 1, ClinicID: 1}, nil
				},
			}
			svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})

			_, err := svc.Update(context.Background(), tt.input)

			assert.Error(t, err)
			assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
			assert.ErrorContains(t, err, tt.wantErrText)
		})
	}
}

// postCloseCloseRepoForTests は締め後編集テスト用の close repo モックを返す。
// 既定で ScheduledDate の am 区分 close を返し、CreateAdjustment を成功させる。
func postCloseCloseRepoForTests(billingDate time.Time) *mockCashRegisterCloseRepository {
	closeRec := &model.CashRegisterClose{
		ID:        99,
		ClinicID:  1,
		CloseDate: time.Date(billingDate.Year(), billingDate.Month(), billingDate.Day(), 0, 0, 0, 0, time.UTC),
		Period:    "am",
	}
	return &mockCashRegisterCloseRepository{
		hasCloseOnDateFn: func(_ context.Context, _ uint64, date time.Time) (bool, error) {
			return date.Format(time.DateOnly) == closeRec.CloseDate.Format(time.DateOnly), nil
		},
		findByDateAndPeriodFn: func(_ context.Context, _ uint64, date time.Time, period string) (*model.CashRegisterClose, error) {
			if period == "am" && date.Format(time.DateOnly) == closeRec.CloseDate.Format(time.DateOnly) {
				return closeRec, nil
			}
			return nil, nil
		},
	}
}

// TestAccountingService_Update_PostCloseReevaluatedInTx は handler が IsPostClose=false でも
// write 時に HasCloseOnDate=true なら post-close 経路へ昇格し、理由欠落は fail-closed に拒否する（W-013 HIGH-1）。
func TestAccountingService_Update_PostCloseReevaluatedInTx(t *testing.T) {
	memo := "メモのみ"
	staffID := uint64(3)
	scheduled := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	billing := &model.Billing{ID: 7, ClinicID: 1, ScheduledDate: scheduled, TotalAmount: 10000}
	closeRepo := postCloseCloseRepoForTests(scheduled)
	repo := &mockAccountingRepository{
		findByIDFn:     func(_ context.Context, _, _ uint64) (*model.Billing, error) { return billing, nil },
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Billing, error) { return billing, nil },
	}
	audit := &mockAuditService{}
	svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, audit, &mockPaymentMethodMasterRepository{},
		WithCashRegisterCloseRepository(closeRepo))

	// 理由なし・handler は非締め → write 再評価で締め済み → 拒否
	_, err := svc.Update(context.Background(), &UpdateAccountingInput{
		ID: 7, ClinicID: 1, StaffID: &staffID, Memo: &memo, IsPostClose: false,
	})
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "expected invalid input for missing reason after re-eval: %v", err)
	assert.False(t, audit.logEntryTxCalled)

	// 理由あり → 昇格して adjustment + audit
	reason := "締め後に判明した金額誤り"
	var captured *model.CashRegisterCloseAdjustment
	closeRepo.createAdjustmentFn = func(_ context.Context, adj *model.CashRegisterCloseAdjustment) error {
		captured = adj
		return nil
	}
	_, err = svc.Update(context.Background(), &UpdateAccountingInput{
		ID: 7, ClinicID: 1, StaffID: &staffID, Memo: &memo, IsPostClose: false, PostCloseReason: &reason,
	})
	assert.NoError(t, err)
	assert.True(t, audit.logEntryTxCalled)
	assert.NotNil(t, captured)
	assert.Equal(t, reason, captured.Reason)
}

// TestAccountingService_Update_PostCloseEmitsAudit は #115 / B4 の成功経路で
// 締め後編集監査ログ（AuditActionBillingPostCloseEdit・reason 付き）が必ず記録されることを固定する。
// reason ガード通過後に監査が欠落しないことの回帰防止網（review follow-up）。
// BE-refactor.md R1-2: Update の締め後編集監査を fail-closed 化（LogEntryTx・ambient tx 参加）した後の
// 挙動保存を確認する（refund/CorrectCreditPayment と同型）。
// W-013: 同一経路で cash_register_close_adjustments も追記される。
func TestAccountingService_Update_PostCloseEmitsAudit(t *testing.T) {
	reason := "金額訂正のため"
	memo := "訂正済み"
	staffID := uint64(3)
	scheduled := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	billing := &model.Billing{ID: 7, ClinicID: 1, ScheduledDate: scheduled, TotalAmount: 10000}

	var capturedAdj *model.CashRegisterCloseAdjustment
	closeRepo := postCloseCloseRepoForTests(scheduled)
	closeRepo.createAdjustmentFn = func(_ context.Context, adj *model.CashRegisterCloseAdjustment) error {
		capturedAdj = adj
		return nil
	}

	repo := &mockAccountingRepository{
		findByIDFn:     func(_ context.Context, _, _ uint64) (*model.Billing, error) { return billing, nil },
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Billing, error) { return billing, nil },
	}
	audit := &mockAuditService{}
	svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, audit, &mockPaymentMethodMasterRepository{},
		WithCashRegisterCloseRepository(closeRepo))

	newTotal := int64(9500)
	_, err := svc.Update(context.Background(), &UpdateAccountingInput{
		ID:              7,
		ClinicID:        1,
		StaffID:         &staffID,
		Memo:            &memo,
		TotalAmount:     &newTotal,
		IsPostClose:     true,
		PostCloseReason: &reason,
	})

	assert.NoError(t, err)
	assert.True(t, audit.logEntryTxCalled, "post-close edit must emit an audit log entry")
	if assert.NotNil(t, audit.logEntryTxInput) {
		assert.Equal(t, model.AuditActionBillingPostCloseEdit, audit.logEntryTxInput.Action)
		meta, ok := audit.logEntryTxInput.Metadata.(map[string]any)
		if assert.True(t, ok, "audit metadata should be map[string]any") {
			assert.Equal(t, reason, meta["reason"])
		}
	}
	if assert.NotNil(t, capturedAdj, "post-close edit must write an append-only adjustment") {
		assert.Equal(t, uint64(99), capturedAdj.CloseID)
		assert.Equal(t, uint64(7), capturedAdj.BillingID)
		assert.Equal(t, int64(-500), capturedAdj.AccountingDelta)
		assert.Equal(t, int64(0), capturedAdj.CashMovementAmount)
		assert.Equal(t, reason, capturedAdj.Reason)
	}
}

// TestAccountingService_Update_PostCloseAdjustmentFailureRollsBack は W-013 の fail-closed 契約:
// adjustment 書込失敗時は Update 全体がエラーになり、監査も本体も残さない。
func TestAccountingService_Update_PostCloseAdjustmentFailureRollsBack(t *testing.T) {
	reason := "金額訂正のため"
	memo := "訂正済み"
	staffID := uint64(3)
	scheduled := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	billing := &model.Billing{ID: 7, ClinicID: 1, ScheduledDate: scheduled}

	closeRepo := postCloseCloseRepoForTests(scheduled)
	closeRepo.createAdjustmentFn = func(_ context.Context, _ *model.CashRegisterCloseAdjustment) error {
		return errors.New("adjustment write failed")
	}
	repo := &mockAccountingRepository{
		findByIDFn:     func(_ context.Context, _, _ uint64) (*model.Billing, error) { return billing, nil },
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Billing, error) { return billing, nil },
	}
	audit := &mockAuditService{}
	svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, audit, &mockPaymentMethodMasterRepository{},
		WithCashRegisterCloseRepository(closeRepo))

	_, err := svc.Update(context.Background(), &UpdateAccountingInput{
		ID:              7,
		ClinicID:        1,
		StaffID:         &staffID,
		Memo:            &memo,
		IsPostClose:     true,
		PostCloseReason: &reason,
	})

	assert.Error(t, err, "adjustment failure must fail the whole update (fail-closed)")
	assert.False(t, audit.logEntryTxCalled, "audit must not run after adjustment failure")
}

// TestAccountingService_Update_PostCloseMissingCloseRepoRollsBack は close repo 未配線時に
// 締め後編集が fail-closed で拒否されることを固定する（W-013）。
func TestAccountingService_Update_PostCloseMissingCloseRepoRollsBack(t *testing.T) {
	reason := "金額訂正のため"
	memo := "訂正済み"
	billing := &model.Billing{ID: 7, ClinicID: 1, ScheduledDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)}

	repo := &mockAccountingRepository{
		findByIDFn:     func(_ context.Context, _, _ uint64) (*model.Billing, error) { return billing, nil },
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Billing, error) { return billing, nil },
	}
	audit := &mockAuditService{}
	svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, audit, &mockPaymentMethodMasterRepository{})

	_, err := svc.Update(context.Background(), &UpdateAccountingInput{
		ID:              7,
		ClinicID:        1,
		Memo:            &memo,
		IsPostClose:     true,
		PostCloseReason: &reason,
	})

	assert.Error(t, err)
	assert.ErrorContains(t, err, "cash register close repository")
	assert.False(t, audit.logEntryTxCalled)
}

// TestAccountingService_Update_PostCloseAuditFailureRollsBack は BE-refactor.md R1-2 の
// fail-closed 契約を固定する: 締め後編集監査（LogEntryTx）が失敗すると Update 全体がエラーを返し
// tx がロールバックされる（本体の fields 更新のみが残る部分コミットを許さない）。
// #211/refund_service で確立した「監査失敗注入で本体も失敗すること」の検証パターンを踏襲する。
func TestAccountingService_Update_PostCloseAuditFailureRollsBack(t *testing.T) {
	reason := "金額訂正のため"
	memo := "訂正済み"
	staffID := uint64(3)
	scheduled := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	billing := &model.Billing{ID: 7, ClinicID: 1, ScheduledDate: scheduled}

	repo := &mockAccountingRepository{
		findByIDFn:     func(_ context.Context, _, _ uint64) (*model.Billing, error) { return billing, nil },
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Billing, error) { return billing, nil },
	}
	audit := &mockAuditService{logEntryTxErr: errors.New("audit write failed")}
	svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, audit, &mockPaymentMethodMasterRepository{},
		WithCashRegisterCloseRepository(postCloseCloseRepoForTests(scheduled)))

	_, err := svc.Update(context.Background(), &UpdateAccountingInput{
		ID:              7,
		ClinicID:        1,
		StaffID:         &staffID,
		Memo:            &memo,
		IsPostClose:     true,
		PostCloseReason: &reason,
	})

	assert.Error(t, err, "audit failure must fail the whole update (fail-closed)")
	assert.True(t, audit.logEntryTxCalled, "audit write must have been attempted")
}
