package billing

import (
	"context"
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// FindCompleteConflict は EMR-66: complete が占めるスロット（medical_record_id /
// hospitalization_id）の UNIQUE インデックスと同じ意味論で衝突する既存 billing を返す。
// INSERT の UNIQUE 違反を待たずに衝突を検出し 409 + 既存会計で返すために使う。
//   - medical_record_id: idx_billings_medical_record_id_unique は deleted_at を述語に
//     持たないため、soft-deleted 済み billing も挿入をブロックする → Unscoped で検索。
//   - hospitalization_id: idx_billings_hospitalization_id_unique は deleted_at IS NULL
//     限定（soft-delete 後の再作成を許す意図的非対称）→ 既定スコープで検索。
//
// 見つからなければ (nil, nil)。
func (r *accountingRepository) FindCompleteConflict(ctx context.Context, clinicID uint64, medicalRecordID, hospitalizationID *uint64) (*model.Billing, error) {
	if medicalRecordID != nil {
		var billing model.Billing
		err := persistence.DBOrTx(ctx, r.db).
			Unscoped().
			Scopes(persistence.ClinicScope(clinicID)).
			Where("medical_record_id = ?", *medicalRecordID).
			// 親表 medical_records の clinic 相関（lint 必須）。UNIQUE スロット占有者の
			// 検出に影響しない: 整合データでは billing.clinic_id = mr.clinic_id が常に成立する。
			// mr.deleted_at は見ない（soft-deleted な MR 参照の行もスロットを占有しうる）。
			Where(`EXISTS (
				SELECT 1
				FROM medical_records AS mr
				WHERE mr.id = billings.medical_record_id
				  AND mr.clinic_id = billings.clinic_id
			)`).
			Order("id ASC").
			Limit(1).
			Find(&billing).Error
		if err != nil {
			return nil, apperrors.FromGORM(err, "billing", fmt.Sprintf("medical_record_id=%d", *medicalRecordID))
		}
		if billing.ID != 0 {
			return &billing, nil
		}
	}
	if hospitalizationID != nil {
		var billing model.Billing
		err := persistence.DBOrTx(ctx, r.db).
			Scopes(persistence.ClinicScope(clinicID)).
			Where("hospitalization_id = ?", *hospitalizationID).
			Order("id ASC").
			Limit(1).
			Find(&billing).Error
		if err != nil {
			return nil, apperrors.FromGORM(err, "billing", fmt.Sprintf("hospitalization_id=%d", *hospitalizationID))
		}
		if billing.ID != 0 {
			return &billing, nil
		}
	}
	return nil, nil
}
