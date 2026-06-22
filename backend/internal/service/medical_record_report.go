package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// OwnerMedicationHistoryItem は飼い主横断投薬歴の1行を表すサービス層 DTO（#158 飼主レポート）。
// repository 層の行型をハンドラへ漏らさないため、サービス境界でこの型へ変換する
// （層分離: handler は repository パッケージに依存しない）。
type OwnerMedicationHistoryItem struct {
	TreatmentID  uint64
	RecordDate   time.Time
	PetID        uint64
	PetName      string
	MedicineName string
	AdminRoute   string
	Quantity     float64
	DoctorName   string
}

// GetOwnerMedicationHistory は飼い主の全ペット横断の投薬履歴（treatments item_type=medicine）を
// 日付降順でページング取得する（#158 飼主レポート）。clinic 隔離は repository 層で medical_records 経由。
func (s *medicalRecordService) GetOwnerMedicationHistory(ctx context.Context, clinicID, ownerID uint64, page, limit int) ([]OwnerMedicationHistoryItem, int64, error) {
	// 飼い主が当該医院に所属するか先に検証する（他院/不存在 ID は NotFound を返す）。
	// 同じ #158 の pet 側エンドポイント（GetFirstVisitDate）と挙動を統一し、ID 探索（IDOR）を防ぐ。
	if _, err := s.ownerRepo.FindByID(ctx, clinicID, ownerID); err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to find owner")
	}

	rows, total, err := s.repo.FindOwnerMedicationHistory(ctx, clinicID, ownerID, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find owner medication history", "error", err, "owner_id", ownerID)
		return nil, 0, apperrors.Wrap(err, "failed to find owner medication history")
	}

	items := make([]OwnerMedicationHistoryItem, len(rows))
	for i, r := range rows {
		items[i] = OwnerMedicationHistoryItem{
			TreatmentID:  r.TreatmentID,
			RecordDate:   r.RecordDate,
			PetID:        r.PetID,
			PetName:      r.PetName,
			MedicineName: r.MedicineName,
			AdminRoute:   r.AdminRoute,
			Quantity:     r.Quantity,
			DoctorName:   r.DoctorName,
		}
	}
	return items, total, nil
}
