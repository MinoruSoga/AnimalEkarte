package medicalrecord

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// TestLockDraftMedicalRecord_NilParentFailsClosed は、LockByIDForUpdate が (nil, nil) を
// 返すモック locker に対して lockDraftMedicalRecord が NotFound を返すことを検証する
// （BE-refactor.md A-5: 第5期回帰で nil を draft とみなして続行する fail-open だった）。
func TestLockDraftMedicalRecord_NilParentFailsClosed(t *testing.T) {
	repo := &mockMedicalRecordRepository{} // findByIDFn 未設定 = (nil, nil) を返す

	err := lockDraftMedicalRecord(context.Background(), repo, 1, 999, "failed to get medical record", "確定済みの診療記録です")

	assert.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}
