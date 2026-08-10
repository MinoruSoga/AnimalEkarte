package billing

// billing_item_status_guard_test.go — BUG-463:
// CreateItem/UpdateItem は確定済み・取消済み会計への明細変更を拒否し、
// DeleteItem と同じ status guard を適用する。UpdateBillingTotals も防御する。

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestBillingItemService_CreateItem_StatusGuard(t *testing.T) {
	tests := []struct {
		name         string
		status       model.BillingStatus
		wantConflict bool
	}{
		{name: "waiting billing accepts create", status: model.BillingStatusWaiting},
		{name: "pending billing accepts create", status: model.BillingStatusPending},
		{name: "completed billing rejects create", status: model.BillingStatusCompleted, wantConflict: true},
		{name: "cancelled billing rejects create", status: model.BillingStatusCancelled, wantConflict: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := setupBillingItemReferenceFixture(t)
			require.NoError(t, f.db.Model(&model.Billing{}).
				Where("id = ? AND clinic_id = ?", f.billing.ID, f.clinicID).
				Update("status", tt.status).Error)
			f.billing.Status = tt.status

			svc := newBillingItemReferenceService(f, f.repo)
			input := billingItemReferenceCreateInput(f)
			created, err := svc.CreateItem(context.Background(), input)

			var persistedCount int64
			require.NoError(t, f.db.Model(&model.BillingItem{}).
				Where("billing_id = ?", f.billing.ID).
				Count(&persistedCount).Error)

			if tt.wantConflict {
				require.Error(t, err)
				assert.True(t, apperrors.IsConflict(err), "finalized billing create must return conflict: %v", err)
				assert.Nil(t, created)
				assert.Zero(t, persistedCount, "rejected create must not persist items")
				return
			}

			require.NoError(t, err)
			require.NotNil(t, created)
			assert.Equal(t, int64(1), persistedCount)
		})
	}
}

func TestBillingItemService_UpdateItem_StatusGuard(t *testing.T) {
	tests := []struct {
		name         string
		status       model.BillingStatus
		wantConflict bool
		wantInvalid  bool
	}{
		{name: "waiting billing accepts update", status: model.BillingStatusWaiting},
		{name: "pending billing accepts update", status: model.BillingStatusPending},
		// BUG-009: completed without reason is invalid input (reason required), not silent conflict alone
		{name: "completed billing rejects update without reason", status: model.BillingStatusCompleted, wantInvalid: true},
		{name: "cancelled billing rejects update", status: model.BillingStatusCancelled, wantConflict: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := setupBillingItemReferenceFixture(t)
			svc := newBillingItemReferenceService(f, f.repo)

			// Create item while billing is still waiting.
			created, err := svc.CreateItem(context.Background(), billingItemReferenceCreateInput(f))
			require.NoError(t, err)
			require.NotNil(t, created)

			require.NoError(t, f.db.Model(&model.Billing{}).
				Where("id = ? AND clinic_id = ?", f.billing.ID, f.clinicID).
				Update("status", tt.status).Error)
			f.billing.Status = tt.status

			newPrice := int64(9999)
			updated, err := svc.UpdateItem(context.Background(), f.clinicID, created.ID, &UpdateBillingItemInput{
				UnitPrice: &newPrice,
			})

			var stored model.BillingItem
			require.NoError(t, f.db.First(&stored, created.ID).Error)

			if tt.wantConflict {
				require.Error(t, err)
				assert.True(t, apperrors.IsConflict(err), "finalized billing update must return conflict: %v", err)
				assert.Nil(t, updated)
				assert.Equal(t, created.UnitPrice, stored.UnitPrice, "rejected update must not change unit price")
				return
			}
			if tt.wantInvalid {
				require.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err), "completed update without reason must be invalid input: %v", err)
				assert.Nil(t, updated)
				assert.Equal(t, created.UnitPrice, stored.UnitPrice, "rejected update must not change unit price")
				return
			}

			require.NoError(t, err)
			require.NotNil(t, updated)
			assert.Equal(t, newPrice, stored.UnitPrice)
		})
	}
}

func TestBillingItemService_UpdateItem_CompletedWithReason_EmitsAudit(t *testing.T) {
	f := setupBillingItemReferenceFixture(t)
	baseSvc := newBillingItemReferenceService(f, f.repo)
	created, err := baseSvc.CreateItem(context.Background(), billingItemReferenceCreateInput(f))
	require.NoError(t, err)

	require.NoError(t, f.db.Model(&model.Billing{}).
		Where("id = ? AND clinic_id = ?", f.billing.ID, f.clinicID).
		Update("status", model.BillingStatusCompleted).Error)
	f.billing.Status = model.BillingStatusCompleted

	audit := &mockAuditService{}
	billingRepo := defaultMockBillingRepo()
	billingRepo.findByIDFn = func(_ context.Context, clinicID, billingID uint64) (*model.Billing, error) {
		if clinicID != f.clinicID || billingID != f.billing.ID {
			return nil, apperrors.WrapNotFound("billing", "test")
		}
		return f.billing, nil
	}
	svc := NewBillingItemServiceWithCampaign(
		f.repo,
		billingRepo,
		defaultMockTreatmentRepo(),
		testNewTransactor(f.db),
		okTrimmingCourseRepo(),
		okTrimmingOptionRepo(),
		nil,
		nil,
		WithBillingItemAuditTx(audit),
	)

	reason := "割引額の入力誤り"
	staffID := uint64(11)
	newPrice := int64(8888)
	updated, err := svc.UpdateItem(context.Background(), f.clinicID, created.ID, &UpdateBillingItemInput{
		UnitPrice:       &newPrice,
		PostCloseReason: &reason,
		StaffID:         &staffID,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, newPrice, updated.UnitPrice)
	require.True(t, audit.logEntryTxCalled)
	require.NotNil(t, audit.logEntryTxInput)
	assert.Equal(t, model.AuditActionBillingPostCloseEdit, audit.logEntryTxInput.Action)
	meta, ok := audit.logEntryTxInput.Metadata.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, reason, meta["reason"])
	assert.Equal(t, "update", meta["operation"])

	var storedBilling model.Billing
	require.NoError(t, f.db.First(&storedBilling, f.billing.ID).Error)
	// totals が completed でも再計算されていること
	assert.NotEqual(t, int64(0), storedBilling.TotalAmount)
}

func TestBillingItemRepository_ValidateCreateReferences_StatusGuard(t *testing.T) {
	tests := []struct {
		name         string
		status       model.BillingStatus
		wantConflict bool
	}{
		{name: "waiting accepts", status: model.BillingStatusWaiting},
		{name: "completed rejects", status: model.BillingStatusCompleted, wantConflict: true},
		{name: "cancelled rejects", status: model.BillingStatusCancelled, wantConflict: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := setupBillingItemReferenceFixture(t)
			require.NoError(t, f.db.Model(&model.Billing{}).
				Where("id = ? AND clinic_id = ?", f.billing.ID, f.clinicID).
				Update("status", tt.status).Error)

			_, err := f.validate(t, nil, nil, nil, nil, nil)
			if tt.wantConflict {
				require.Error(t, err)
				assert.True(t, apperrors.IsConflict(err), "ValidateCreateReferences must conflict on finalized: %v", err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestBillingItemRepository_UpdateBillingTotals_StatusGuard(t *testing.T) {
	f := setupBillingItemReferenceFixture(t)
	repo := f.repo

	// Open status: totals update succeeds.
	require.NoError(t, repo.UpdateBillingTotals(context.Background(), f.clinicID, f.billing.ID, 1000, 100, 1100))
	var open model.Billing
	require.NoError(t, f.db.First(&open, f.billing.ID).Error)
	assert.EqualValues(t, 1100, open.TotalAmount)

	// Completed: totals rewrite is Conflict and values stay.
	require.NoError(t, f.db.Model(&model.Billing{}).
		Where("id = ?", f.billing.ID).
		Update("status", model.BillingStatusCompleted).Error)

	err := repo.UpdateBillingTotals(context.Background(), f.clinicID, f.billing.ID, 9999, 999, 10998)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "completed totals rewrite must conflict: %v", err)

	var completed model.Billing
	require.NoError(t, f.db.First(&completed, f.billing.ID).Error)
	assert.EqualValues(t, 1100, completed.TotalAmount, "completed billing totals must remain unchanged")

	// Cancelled: same guard.
	require.NoError(t, f.db.Model(&model.Billing{}).
		Where("id = ?", f.billing.ID).
		Updates(map[string]any{"status": model.BillingStatusCancelled, "total_amount": 1100}).Error)

	err = repo.UpdateBillingTotals(context.Background(), f.clinicID, f.billing.ID, 1, 1, 2)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "cancelled totals rewrite must conflict: %v", err)
}

func TestBillingItemService_CreateItem_PostCloseReasonRequired(t *testing.T) {
	f := setupBillingItemReferenceFixture(t)
	svc := newBillingItemReferenceService(f, f.repo)
	input := billingItemReferenceCreateInput(f)
	input.IsPostClose = true
	input.PostCloseReason = nil

	item, err := svc.CreateItem(context.Background(), input)
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "missing post_close_reason must be invalid input: %v", err)
	assert.Nil(t, item)

	empty := ""
	input.PostCloseReason = &empty
	item, err = svc.CreateItem(context.Background(), input)
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, item)
}

func TestBillingItemService_UpdateItem_PostCloseReasonRequired(t *testing.T) {
	f := setupBillingItemReferenceFixture(t)
	svc := newBillingItemReferenceService(f, f.repo)
	created, err := svc.CreateItem(context.Background(), billingItemReferenceCreateInput(f))
	require.NoError(t, err)

	newPrice := int64(2000)
	_, err = svc.UpdateItem(context.Background(), f.clinicID, created.ID, &UpdateBillingItemInput{
		UnitPrice:   &newPrice,
		IsPostClose: true,
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "missing post_close_reason must be invalid input: %v", err)
}

func TestBillingItemService_CreateItem_PostCloseAuditFailureRollsBack(t *testing.T) {
	f := setupBillingItemReferenceFixture(t)
	audit := &mockAuditService{logEntryTxErr: errors.New("audit write failed")}
	billingRepo := defaultMockBillingRepo()
	billingRepo.findByIDFn = func(_ context.Context, clinicID, billingID uint64) (*model.Billing, error) {
		if clinicID != f.clinicID || billingID != f.billing.ID {
			return nil, apperrors.WrapNotFound("billing", "test")
		}
		return f.billing, nil
	}
	svc := NewBillingItemServiceWithCampaign(
		f.repo,
		billingRepo,
		defaultMockTreatmentRepo(),
		testNewTransactor(f.db),
		okTrimmingCourseRepo(),
		okTrimmingOptionRepo(),
		nil,
		nil,
		WithBillingItemAuditTx(audit),
		WithBillingItemCloseRepository(postCloseCloseRepoForTests(f.billing.ScheduledDate)),
	)

	reason := "締め後修正"
	staffID := uint64(7)
	input := billingItemReferenceCreateInput(f)
	input.IsPostClose = true
	input.PostCloseReason = &reason
	input.StaffID = &staffID

	item, err := svc.CreateItem(context.Background(), input)
	require.Error(t, err)
	assert.Nil(t, item)
	assert.True(t, audit.logEntryTxCalled, "post-close create must attempt audit")

	var count int64
	require.NoError(t, f.db.Model(&model.BillingItem{}).
		Where("billing_id = ?", f.billing.ID).
		Count(&count).Error)
	assert.Zero(t, count, "audit failure must roll back item create")
}

func TestBillingItemService_CreateItem_PostCloseEmitsAudit(t *testing.T) {
	f := setupBillingItemReferenceFixture(t)
	audit := &mockAuditService{}
	billingRepo := defaultMockBillingRepo()
	billingRepo.findByIDFn = func(_ context.Context, clinicID, billingID uint64) (*model.Billing, error) {
		if clinicID != f.clinicID || billingID != f.billing.ID {
			return nil, apperrors.WrapNotFound("billing", "test")
		}
		return f.billing, nil
	}
	svc := NewBillingItemServiceWithCampaign(
		f.repo,
		billingRepo,
		defaultMockTreatmentRepo(),
		testNewTransactor(f.db),
		okTrimmingCourseRepo(),
		okTrimmingOptionRepo(),
		nil,
		nil,
		WithBillingItemAuditTx(audit),
		WithBillingItemCloseRepository(postCloseCloseRepoForTests(f.billing.ScheduledDate)),
	)

	reason := "締め後明細追加"
	staffID := uint64(9)
	input := billingItemReferenceCreateInput(f)
	input.IsPostClose = true
	input.PostCloseReason = &reason
	input.StaffID = &staffID

	item, err := svc.CreateItem(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, item)
	require.True(t, audit.logEntryTxCalled)
	require.NotNil(t, audit.logEntryTxInput)
	assert.Equal(t, model.AuditActionBillingPostCloseEdit, audit.logEntryTxInput.Action)
	assert.Equal(t, "billing_item", audit.logEntryTxInput.Resource)
	meta, ok := audit.logEntryTxInput.Metadata.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, reason, meta["reason"])
	assert.Equal(t, f.billing.ID, meta["billing_id"])
	assert.Equal(t, "create", meta["operation"])
}

func TestBillingItemService_UpdateItem_PostCloseAuditFailureRollsBack(t *testing.T) {
	f := setupBillingItemReferenceFixture(t)
	// First create without post-close.
	baseSvc := newBillingItemReferenceService(f, f.repo)
	created, err := baseSvc.CreateItem(context.Background(), billingItemReferenceCreateInput(f))
	require.NoError(t, err)
	originalPrice := created.UnitPrice

	audit := &mockAuditService{logEntryTxErr: errors.New("audit write failed")}
	billingRepo := defaultMockBillingRepo()
	billingRepo.findByIDFn = func(_ context.Context, clinicID, billingID uint64) (*model.Billing, error) {
		if clinicID != f.clinicID || billingID != f.billing.ID {
			return nil, apperrors.WrapNotFound("billing", "test")
		}
		return f.billing, nil
	}
	svc := NewBillingItemServiceWithCampaign(
		f.repo,
		billingRepo,
		defaultMockTreatmentRepo(),
		testNewTransactor(f.db),
		okTrimmingCourseRepo(),
		okTrimmingOptionRepo(),
		nil,
		nil,
		WithBillingItemAuditTx(audit),
		WithBillingItemCloseRepository(postCloseCloseRepoForTests(f.billing.ScheduledDate)),
	)

	reason := "締め後単価修正"
	staffID := uint64(3)
	newPrice := int64(7777)
	_, err = svc.UpdateItem(context.Background(), f.clinicID, created.ID, &UpdateBillingItemInput{
		UnitPrice:       &newPrice,
		IsPostClose:     true,
		PostCloseReason: &reason,
		StaffID:         &staffID,
	})
	require.Error(t, err)
	assert.True(t, audit.logEntryTxCalled)

	var stored model.BillingItem
	require.NoError(t, f.db.First(&stored, created.ID).Error)
	assert.Equal(t, originalPrice, stored.UnitPrice, "audit failure must roll back item update")
}

// ---- DeleteItem post-close gate (BUG-463 residual) ----

func TestBillingItemService_DeleteItem_PostCloseReasonRequired(t *testing.T) {
	f := setupBillingItemReferenceFixture(t)
	svc := newBillingItemReferenceService(f, f.repo)
	created, err := svc.CreateItem(context.Background(), billingItemReferenceCreateInput(f))
	require.NoError(t, err)

	err = svc.DeleteItem(context.Background(), f.clinicID, created.ID, &DeleteBillingItemInput{
		IsPostClose: true,
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "missing post_close_reason must be invalid input: %v", err)

	empty := ""
	err = svc.DeleteItem(context.Background(), f.clinicID, created.ID, &DeleteBillingItemInput{
		IsPostClose:     true,
		PostCloseReason: &empty,
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))

	// Item must still exist (reason rejected outside tx).
	var count int64
	require.NoError(t, f.db.Model(&model.BillingItem{}).
		Where("id = ? AND deleted_at IS NULL", created.ID).
		Count(&count).Error)
	assert.Equal(t, int64(1), count, "invalid post-close delete must not soft-delete the item")
}

func TestBillingItemService_DeleteItem_PostCloseEmitsAudit(t *testing.T) {
	f := setupBillingItemReferenceFixture(t)
	baseSvc := newBillingItemReferenceService(f, f.repo)
	created, err := baseSvc.CreateItem(context.Background(), billingItemReferenceCreateInput(f))
	require.NoError(t, err)

	audit := &mockAuditService{}
	billingRepo := defaultMockBillingRepo()
	billingRepo.findByIDFn = func(_ context.Context, clinicID, billingID uint64) (*model.Billing, error) {
		if clinicID != f.clinicID || billingID != f.billing.ID {
			return nil, apperrors.WrapNotFound("billing", "test")
		}
		return f.billing, nil
	}
	svc := NewBillingItemServiceWithCampaign(
		f.repo,
		billingRepo,
		defaultMockTreatmentRepo(),
		testNewTransactor(f.db),
		okTrimmingCourseRepo(),
		okTrimmingOptionRepo(),
		nil,
		nil,
		WithBillingItemAuditTx(audit),
		WithBillingItemCloseRepository(postCloseCloseRepoForTests(f.billing.ScheduledDate)),
	)

	reason := "締め後明細削除"
	staffID := uint64(15)
	err = svc.DeleteItem(context.Background(), f.clinicID, created.ID, &DeleteBillingItemInput{
		StaffID:         &staffID,
		IsPostClose:     true,
		PostCloseReason: &reason,
	})
	require.NoError(t, err)
	require.True(t, audit.logEntryTxCalled)
	require.NotNil(t, audit.logEntryTxInput)
	assert.Equal(t, model.AuditActionBillingPostCloseEdit, audit.logEntryTxInput.Action)
	assert.Equal(t, "billing_item", audit.logEntryTxInput.Resource)
	meta, ok := audit.logEntryTxInput.Metadata.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, reason, meta["reason"])
	assert.Equal(t, f.billing.ID, meta["billing_id"])
	assert.Equal(t, "delete", meta["operation"])
}

func TestBillingItemService_DeleteItem_PostCloseAuditFailureRollsBack(t *testing.T) {
	f := setupBillingItemReferenceFixture(t)
	baseSvc := newBillingItemReferenceService(f, f.repo)
	created, err := baseSvc.CreateItem(context.Background(), billingItemReferenceCreateInput(f))
	require.NoError(t, err)

	audit := &mockAuditService{logEntryTxErr: errors.New("audit write failed")}
	billingRepo := defaultMockBillingRepo()
	billingRepo.findByIDFn = func(_ context.Context, clinicID, billingID uint64) (*model.Billing, error) {
		if clinicID != f.clinicID || billingID != f.billing.ID {
			return nil, apperrors.WrapNotFound("billing", "test")
		}
		return f.billing, nil
	}
	svc := NewBillingItemServiceWithCampaign(
		f.repo,
		billingRepo,
		defaultMockTreatmentRepo(),
		testNewTransactor(f.db),
		okTrimmingCourseRepo(),
		okTrimmingOptionRepo(),
		nil,
		nil,
		WithBillingItemAuditTx(audit),
		WithBillingItemCloseRepository(postCloseCloseRepoForTests(f.billing.ScheduledDate)),
	)

	reason := "締め後削除失敗"
	staffID := uint64(16)
	err = svc.DeleteItem(context.Background(), f.clinicID, created.ID, &DeleteBillingItemInput{
		StaffID:         &staffID,
		IsPostClose:     true,
		PostCloseReason: &reason,
	})
	require.Error(t, err)
	assert.True(t, audit.logEntryTxCalled, "post-close delete must attempt audit")

	var stored struct {
		DeletedAt *time.Time
	}
	require.NoError(t, f.db.Unscoped().
		Table("billing_items").
		Select("deleted_at").
		Where("id = ?", created.ID).
		Take(&stored).Error)
	assert.Nil(t, stored.DeletedAt, "audit failure must roll back soft-delete")
}
