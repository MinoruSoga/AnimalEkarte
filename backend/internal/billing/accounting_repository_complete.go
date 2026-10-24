package billing

import (
	"context"
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

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
// WHERE claim_state='cancelled' でガードするため draft/finalized 行の誤削除は 0 件ヒットに留まり、
// 0 件（既に解放済み・状態遷移済み）も冪等成功として扱う（呼出し側は直前に行を観測済み）。billings.status は廃止され claim_state に置換された。
func (r *accountingRepository) SoftDeleteCancelled(ctx context.Context, clinicID, id uint64) error {
	if err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ? AND claim_state = ?", id, billingClaimStateCancelled).
		Delete(&model.Billing{}).Error; err != nil {
		return apperrors.FromGORM(err, "billing", fmt.Sprintf("%d", id))
	}
	return nil
}
