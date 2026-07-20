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
// medical_record_lock.go は internal/service の同名ヘルパーと byte-identical だが、移設先の
// medicalrecord package ではこの fail-closed 不変条件が未カバー（カバレッジ0）だったため
// BE9-2D sub-batch④a で新設する。service 側の同 test は残留 consumer のため据え置く。
func TestLockDraftMedicalRecord_NilParentFailsClosed(t *testing.T) {
	repo := &mockMedicalRecordRepository{} // findByIDFn 未設定 = (nil, nil) を返す

	err := lockDraftMedicalRecord(context.Background(), repo, 1, 999, "failed to get medical record", "確定済みの診療記録です")

	assert.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}
