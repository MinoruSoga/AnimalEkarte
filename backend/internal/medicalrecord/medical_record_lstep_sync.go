package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/model"
)

func (s *medicalRecordService) syncVisitCompletionTags(ctx context.Context, clinicID uint64, record *model.MedicalRecord) {
	if s.tagSyncSvc == nil || record == nil || record.OwnerID == nil {
		return
	}
	ownerID := *record.OwnerID
	if err := s.tagSyncSvc.SyncVisitCompletionTags(ctx, clinicID, ownerID); err != nil {
		slog.ErrorContext(ctx, "failed to sync visit completion tags", "error", err, "clinic_id", clinicID, "owner_id", ownerID, "record_id", record.ID)
	}
}

func (s *medicalRecordService) syncNextVisitTag(ctx context.Context, clinicID uint64, record *model.MedicalRecord) {
	if s.tagSyncSvc == nil || record == nil || record.OwnerID == nil {
		return
	}
	ownerID := *record.OwnerID
	if err := s.tagSyncSvc.SyncNextVisitTag(ctx, clinicID, ownerID); err != nil {
		slog.ErrorContext(ctx, "failed to sync next visit tag", "error", err, "clinic_id", clinicID, "owner_id", ownerID, "record_id", record.ID)
	}
}

// CreateSubRecords はカルテ作成と同時に inquiry / clinical_plan を best-effort で作成する。
// 失敗しても呼び出し元のカルテ作成は完了済みのためエラーは握りつぶし slog.Warn のみ出力する。
