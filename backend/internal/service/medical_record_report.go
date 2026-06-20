package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/repository"
)

// GetOwnerMedicationHistory は飼い主の全ペット横断の投薬履歴を日付降順でページング取得する（#158 飼主レポート）。
// 投薬の実体は treatments(item_type=medicine)。clinic 隔離は repository 層で medical_records 経由に行うため、
// ここでは X-Clinic-ID から確定した単一 clinicID をそのまま渡す（別医院の投薬は混入しない）。
func (s *medicalRecordService) GetOwnerMedicationHistory(ctx context.Context, clinicID, ownerID uint64, page, limit int) ([]repository.OwnerMedicationHistoryRow, int64, error) {
	rows, total, err := s.repo.FindOwnerMedicationHistory(ctx, clinicID, ownerID, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find owner medication history", "error", err, "owner_id", ownerID)
		return nil, 0, apperrors.Wrap(err, "failed to find owner medication history")
	}
	return rows, total, nil
}
