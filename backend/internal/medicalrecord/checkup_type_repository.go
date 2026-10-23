package medicalrecord

import (
	"context"
	"fmt"
	"sort"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// CheckupTypeRepository is the data access interface for checkup types.
// Moved from internal/repository/checkuptype (BE8-4 batch5) — BE9-2D roll-up. Renamed from that
// subpackage's generic "Repository" to this entity-specific name only because medicalrecord
// holds multiple repository interfaces in one package; every external caller only ever saw
// this name via the internal/repository facade, so no call site changes.
type CheckupTypeRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.CheckupType, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error)
	Create(ctx context.Context, checkupType *model.CheckupType) error
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateCheckupTypeInput) (*model.CheckupType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByCheckupTypeID(ctx context.Context, clinicID, checkupTypeID uint64) (int64, error)
	CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
}

type checkupTypeRepository struct{ db *gorm.DB }

// NewCheckupTypeRepository constructs a CheckupTypeRepository.
func NewCheckupTypeRepository(db *gorm.DB) CheckupTypeRepository {
	return &checkupTypeRepository{db: db}
}

// checkupTypeReadColumns は is_bill_discount_excluded を含む明示列（EMR-447 A.3）。
// SELECT * だけでは列未接続と DB NULL を区別できないため、読取は列名を明示する。
var checkupTypeReadColumns = []string{
	"id", "clinic_id", "name", "price", "is_active", "description",
	"interval", "target_age", "parent_id", "sort_order",
	"import_namespace", "import_key", "is_non_insurance", colIsBillDiscountExcluded,
	"version", "created_at", "updated_at", "deleted_at",
}

func (r *checkupTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.CheckupType, error) {
	checkupTypes := make([]model.CheckupType, 0)
	err := r.db.WithContext(ctx).
		Model(&model.CheckupType{}).
		Select(checkupTypeReadColumns).
		Scopes(persistence.ClinicScope(clinicID)).
		Order("sort_order ASC, name ASC").
		Find(&checkupTypes).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup_type", "")
	}
	return checkupTypes, nil
}

// FindByID は ambient tx に参加する（EMR-456: Update の前値読取と parent 検証を書込と同じ tx で行う）。
func (r *checkupTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error) {
	return persistence.FindByIDScoped[model.CheckupType](ctx, persistence.DBOrTx(ctx, r.db).Select(checkupTypeReadColumns), "checkup_type", clinicID, id)
}

// LockWriteRows locks target and requested parent rows in a global order before
// validation/mutation（EMR-456 F1: 親 soft-delete と子 Create/親変更の直列化）。
// soft-delete は非 key UPDATE のため FOR KEY SHARE では衝突しない。FOR UPDATE を使う。
func (r *checkupTypeRepository) LockWriteRows(ctx context.Context, clinicID uint64, ids []uint64) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("checkup type parent lock requires an ambient transaction")
	}
	ordered := append([]uint64(nil), ids...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	for i, id := range ordered {
		if id == 0 {
			return apperrors.WrapInvalidInput("checkup type ID must be positive")
		}
		if i > 0 && id == ordered[i-1] {
			continue
		}
		db := persistence.DBOrTx(ctx, r.db).Clauses(clause.Locking{Strength: "UPDATE"})
		if _, err := persistence.FindByIDScoped[model.CheckupType](ctx, db, "checkup_type", clinicID, id); err != nil {
			return err
		}
	}
	return nil
}

// Create は ambient tx に参加し、service の保険設定監査と同じ tx に載る（EMR-456）。
func (r *checkupTypeRepository) Create(ctx context.Context, checkupType *model.CheckupType) error {
	db := persistence.DBOrTx(ctx, r.db)
	wantActive := checkupType.IsActive
	if err := db.Create(checkupType).Error; err != nil {
		return apperrors.FromGORM(err, "checkup_type", "")
	}
	if !wantActive {
		if err := db.Model(checkupType).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "checkup_type", fmt.Sprintf("%d", checkupType.ID))
		}
		// この UPDATE は内容変更なので entity_version_guard が DB の version を加算する。
		// 応答・監査が古い版を返さないよう、同じ tx で再読取した行に揃える（EMR-456）。
		stored, err := r.FindByID(ctx, checkupType.ClinicID, checkupType.ID)
		if err != nil {
			return err
		}
		*checkupType = *stored
	}
	return nil
}

// Update は ambient tx に参加し、cmd.expectedVersion が非 nil なら version を WHERE で照合する
// （EMR-456 楽観ロック）。0 行は「存在しない（別医院含む）」と「版不一致」に区別して返す。
// 返却値は同じ tx 内で再読取するため、commit 後の read error で成功を失敗へ反転させない。
func (r *checkupTypeRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateCheckupTypeInput) (*model.CheckupType, error) {
	q := persistence.DBOrTx(ctx, r.db).
		Model(&model.CheckupType{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id)
	if cmd.expectedVersion != nil {
		q = q.Where("version = ?", *cmd.expectedVersion)
	}
	result := q.Updates(buildCheckupTypeUpdate(&cmd))
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "checkup_type", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		if _, err := r.FindByID(ctx, clinicID, id); err != nil {
			return nil, err
		}
		return nil, apperrors.WrapConflict(errMsgCheckupTypeVersionConflict)
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *checkupTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Where(`NOT EXISTS (
			SELECT 1 FROM checkup_types children
			WHERE children.parent_id = checkup_types.id
			  AND children.clinic_id = ?
			  AND children.deleted_at IS NULL
		)`, clinicID).
		Where(`NOT EXISTS (
			SELECT 1 FROM checkups
			WHERE checkups.checkup_type_id = checkup_types.id
			  AND checkups.clinic_id = ?
			  AND checkups.deleted_at IS NULL
		)`, clinicID).
		Delete(&model.CheckupType{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "checkup_type", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return r.normalizeCheckupTypeDeleteMiss(ctx, clinicID, id)
	}
	return nil
}

func (r *checkupTypeRepository) normalizeCheckupTypeDeleteMiss(ctx context.Context, clinicID, id uint64) error {
	if _, err := r.FindByID(ctx, clinicID, id); err != nil {
		return err
	}
	childCount, err := r.CountChildrenByParentID(ctx, clinicID, id)
	if err != nil {
		return err
	}
	if childCount > 0 {
		return apperrors.WrapConflict("この定期健診種別にはサブ種別が登録されているため削除できません")
	}
	return apperrors.WrapConflict("この定期健診種別は健診記録で使用中のため削除できません")
}

func (r *checkupTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.CheckupType{}, "checkup_type", clinicID, ids, "sort_order")
}

// CountUsageByCheckupTypeID は定期健診種別を参照している checkups の件数を返す（BUG-107）
// checkups テーブルは直接 clinic_id を持つためテナント分離を直接適用する。
// ambient tx に参加する（EMR-456 F1: lock 後の新鮮な statement で確認するため）。
func (r *checkupTypeRepository) CountUsageByCheckupTypeID(ctx context.Context, clinicID, checkupTypeID uint64) (int64, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Checkup{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("checkup_type_id = ? AND deleted_at IS NULL", checkupTypeID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "checkup_type", "")
	}
	return count, nil
}

// CountChildrenByParentID は指定した親 ID を持つ子健診種別の件数を返す。
// ambient tx に参加する（EMR-456 F1: lock 待ち後に READ COMMITTED の新 snapshot で再確認するため）。
func (r *checkupTypeRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.CheckupType{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "checkup_type", "")
	}
	return count, nil
}
