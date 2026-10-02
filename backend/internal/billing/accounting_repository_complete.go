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

// FindByHospitalizationID は EMR-253: 入院スロット（hospitalization_id, deleted_at IS NULL）を
// 占有する billing を status 非依存で返す。complete の takeover 解決と退院時の
// no-duplicate ガード（medicalrecord 側 consumer interface 経由）が共有する。
// idx_billings_hospitalization_id_unique（部分 UNIQUE・deleted_at IS NULL 限定）と同じ意味論。
// 見つからなければ (nil, nil)。
func (r *accountingRepository) FindByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) (*model.Billing, error) {
	var billing model.Billing
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("hospitalization_id = ?", hospitalizationID).
		Order("id ASC").
		Limit(1).
		Find(&billing).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "billing", fmt.Sprintf("hospitalization_id=%d", hospitalizationID))
	}
	if billing.ID == 0 {
		return nil, nil
	}
	return &billing, nil
}

// SoftDeleteCancelled は EMR-253: cancelled 状態の占有 billing を soft-delete して
// 部分 UNIQUE（deleted_at IS NULL）の入院スロットを同一 tx 内で解放する。
// WHERE status='cancelled' でガードするため waiting/completed 行の誤削除は 0 件ヒットに留まり、
// 0 件（既に解放済み・status 遷移済み）も冪等成功として扱う（呼出し側は直前に行を観測済み）。
func (r *accountingRepository) SoftDeleteCancelled(ctx context.Context, clinicID, id uint64) error {
	if err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ? AND status = ?", id, model.BillingStatusCancelled).
		Delete(&model.Billing{}).Error; err != nil {
		return apperrors.FromGORM(err, "billing", fmt.Sprintf("%d", id))
	}
	return nil
}
