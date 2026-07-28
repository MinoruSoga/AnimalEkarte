package medicalrecord

import (
	"context"

	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// lockDraftMedicalRecord は sharedkernel.LockDraftMedicalRecord への既存呼び出し面互換 delegate。
// ④a までは internal/service との documented duplicate だったが、共有カーネル昇格batch で
// 実装正本を sharedkernel へ一本化した（X-11 臨床安全ガードの複製ドリフト排除）。
// medicalRecordLocker（service_deps.go）は structural typing で sharedkernel.MedicalRecordLocker を満たす。
func lockDraftMedicalRecord(ctx context.Context, repo medicalRecordLocker, clinicID, recordID uint64, findErrMsg, conflictMsg string) error {
	return sharedkernel.LockDraftMedicalRecord(ctx, repo, clinicID, recordID, findErrMsg, conflictMsg)
}
