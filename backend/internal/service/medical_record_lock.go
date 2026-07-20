package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// medicalRecordLocker は LockByIDForUpdate 1メソッドの narrow interface（BE-refactor.md E-5）。
// 実装正本は sharedkernel.MedicalRecordLocker（共有カーネル昇格batch）。既存呼び出し面互換の
// ローカル別名として維持する（削除=各 domain 移行時）。
type medicalRecordLocker = sharedkernel.MedicalRecordLocker

// lockDraftMedicalRecord は sharedkernel.LockDraftMedicalRecord への既存呼び出し面互換 delegate
// （X-11 行ロック+finalized ガード。実装正本は sharedkernel — 複製ドリフト排除のため本 package に
// ロジックを持たない。呼び出し側の直参照切替=各 domain 移行時）。
func lockDraftMedicalRecord(ctx context.Context, repo medicalRecordLocker, clinicID, recordID uint64, findErrMsg, conflictMsg string) error {
	return sharedkernel.LockDraftMedicalRecord(ctx, repo, clinicID, recordID, findErrMsg, conflictMsg)
}
