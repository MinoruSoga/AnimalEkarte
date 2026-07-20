package sharedkernel

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type stubLocker struct {
	rec *model.MedicalRecord
	err error
}

func (s *stubLocker) LockByIDForUpdate(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
	return s.rec, s.err
}

// TestLockDraftMedicalRecord_NilParentFailsClosed は fail-closed 契約（A-5: nil 親 = NotFound）の
// kernel 直接テスト。service/medicalrecord 両 package の同名テストは delegate 経由で同じ実装を
// 検証し続ける（昇格で検証が薄まらないことの三重化）。
func TestLockDraftMedicalRecord_NilParentFailsClosed(t *testing.T) {
	err := LockDraftMedicalRecord(context.Background(), &stubLocker{rec: nil}, 1, 10, "find failed", "conflict msg")
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "nil parent は NotFound (fail-closed): %v", err)
}

func TestLockDraftMedicalRecord_Branches(t *testing.T) {
	t.Run("finalized parent returns conflict with given message", func(t *testing.T) {
		err := LockDraftMedicalRecord(context.Background(),
			&stubLocker{rec: &model.MedicalRecord{Status: model.MedicalRecordStatusFinalized}}, 1, 10, "find failed", "編集できません")
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Contains(t, err.Error(), "編集できません")
	})
	t.Run("draft parent passes", func(t *testing.T) {
		err := LockDraftMedicalRecord(context.Background(),
			&stubLocker{rec: &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}}, 1, 10, "find failed", "conflict msg")
		assert.NoError(t, err)
	})
	t.Run("lock error is wrapped with findErrMsg", func(t *testing.T) {
		err := LockDraftMedicalRecord(context.Background(),
			&stubLocker{err: errors.New("db down")}, 1, 10, "find failed", "conflict msg")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "find failed")
	})
}

func TestSharedKernelValidators(t *testing.T) {
	t.Run("ValidateRequiredName rejects empty/too-long/control chars", func(t *testing.T) {
		assert.Error(t, ValidateRequiredName("  "))
		assert.Error(t, ValidateRequiredName(string(make([]rune, 0))+""))
		long := make([]rune, MasterNameMaxLength+1)
		for i := range long {
			long[i] = 'あ'
		}
		assert.Error(t, ValidateRequiredName(string(long)))
		assert.Error(t, ValidateRequiredName("bad\u0000name"))
		assert.NoError(t, ValidateRequiredName("正常な名前"))
	})
	t.Run("ValidateNonNegativePrice", func(t *testing.T) {
		neg := int64(-1)
		zero := int64(0)
		assert.NoError(t, ValidateNonNegativePrice(nil))
		assert.NoError(t, ValidateNonNegativePrice(&zero))
		assert.Error(t, ValidateNonNegativePrice(&neg))
	})
	t.Run("ValidateDiscountRate bounds", func(t *testing.T) {
		assert.NoError(t, ValidateDiscountRate(0))
		assert.NoError(t, ValidateDiscountRate(100))
		assert.Error(t, ValidateDiscountRate(-0.1))
		assert.Error(t, ValidateDiscountRate(100.1))
	})
	t.Run("AuditActorTypeFor", func(t *testing.T) {
		id := uint64(3)
		assert.Equal(t, model.AuditActorTypeStaff, AuditActorTypeFor(&id))
		assert.Equal(t, model.AuditActorTypeSystem, AuditActorTypeFor(nil))
	})
}
