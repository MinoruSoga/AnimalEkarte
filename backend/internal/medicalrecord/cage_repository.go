package medicalrecord

// Moved from internal/repository/cage (BE9-2D ⑥ Batch A roll-up・BE8-4 subpackage)。
// generic Repository/New は entity-specific 名へ改名（①⑤先例）— 外部は facade alias 経由で不変。

import (
	"context"
	"fmt"
	"slices"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// Repository is the data access interface for cages.
type CageRepository interface {
	FindAll(ctx context.Context, clinicID uint64, cageType *string) ([]model.Cage, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error)
	// LockByIDForUpdate takes FOR UPDATE under an ambient transaction (SEC-CS-F13).
	LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Cage, error)
	CountUsageByCageID(ctx context.Context, clinicID, id uint64) (int64, error)
	Create(ctx context.Context, cage *model.Cage) error
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateCageInput) (*model.Cage, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type cageRepositoryImpl struct{ db *gorm.DB }

// New constructs a Repository.
func NewCageRepository(db *gorm.DB) CageRepository {
	return &cageRepositoryImpl{db: db}
}

func (r *cageRepositoryImpl) FindAll(ctx context.Context, clinicID uint64, cageType *string) ([]model.Cage, error) {
	cages := make([]model.Cage, 0)
	q := persistence.DBOrTx(ctx, r.db).Model(&model.Cage{}).Scopes(persistence.ClinicScope(clinicID))
	if cageType != nil {
		q = q.Where("cage_type = ?", *cageType)
	}
	if err := q.Order("sort_order ASC, name ASC").Find(&cages).Error; err != nil {
		return nil, apperrors.FromGORM(err, "cage", "")
	}
	return cages, nil
}

// FindByID loads a clinic-scoped cage. When called under an ambient transaction it
// takes FOR SHARE so concurrent soft-delete (FOR UPDATE) waits until the caller
// commits — hospitalization cage FK validation serialization (SEC-CS-F13).
func (r *cageRepositoryImpl) FindByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error) {
	var cage model.Cage
	db := persistence.DBOrTx(ctx, r.db)
	if persistence.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	if err := db.Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&cage).Error; err != nil {
		return nil, apperrors.FromGORM(err, "cage", fmt.Sprintf("%d", id))
	}
	return &cage, nil
}

// LockByIDForUpdate exclusive-locks a cage row for the ambient transaction so usage
// check + soft-delete serialize with hospitalization assignment (SEC-CS-F13).
func (r *cageRepositoryImpl) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Cage, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("cage lock requires ambient transaction")
	}
	var cage model.Cage
	db := persistence.DBOrTx(ctx, r.db).Clauses(clause.Locking{Strength: "UPDATE"})
	if err := db.Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&cage).Error; err != nil {
		return nil, apperrors.FromGORM(err, "cage", fmt.Sprintf("%d", id))
	}
	return &cage, nil
}

func (r *cageRepositoryImpl) Create(ctx context.Context, cage *model.Cage) error {
	db := persistence.DBOrTx(ctx, r.db)
	// Capture intent before Create: gorm default:true omits zero bools from INSERT.
	wantActive := cage.IsActive
	if err := db.Create(cage).Error; err != nil {
		return apperrors.FromGORM(err, "cage", "")
	}
	if !wantActive {
		if err := db.Model(cage).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "cage", fmt.Sprintf("%d", cage.ID))
		}
		cage.IsActive = false
	}
	return nil
}

func (r *cageRepositoryImpl) Update(ctx context.Context, clinicID, id uint64, cmd UpdateCageInput) (*model.Cage, error) {
	if err := r.update(ctx, clinicID, id, buildCageUpdate(&cmd)); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *cageRepositoryImpl) update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return persistence.UpdateScopedByID(ctx, persistence.DBOrTx(ctx, r.db), &model.Cage{}, "cage", clinicID, id, fields)
}

func (r *cageRepositoryImpl) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Where(`NOT EXISTS (
			SELECT 1 FROM hospitalizations
			WHERE hospitalizations.cage_id = cages.id
			  AND hospitalizations.clinic_id = ?
			  AND hospitalizations.deleted_at IS NULL
		)`, clinicID).
		Delete(&model.Cage{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "cage", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return r.normalizeCageDeleteMiss(ctx, clinicID, id)
	}
	return nil
}

func (r *cageRepositoryImpl) normalizeCageDeleteMiss(ctx context.Context, clinicID, id uint64) error {
	if _, err := r.FindByID(ctx, clinicID, id); err != nil {
		return err
	}
	return apperrors.WrapConflict("このケージは入院データで使用中のため削除できません")
}

func (r *cageRepositoryImpl) CountUsageByCageID(ctx context.Context, clinicID, id uint64) (int64, error) {
	// DBOrTx so delete's ambient transaction sees uncommitted hospitalization assignments.
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Hospitalization{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("cage_id = ? AND deleted_at IS NULL", id).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "hospitalization", "")
	}
	return count, nil
}

func (r *cageRepositoryImpl) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.Cage{}, "cage", clinicID, ids, "sort_order")
}

// ---- cage merge (EMR-306) ----

// CageMergeRepository はケージ一時連結が必要とするロック・占有・リンク操作の面。
// CageRepository にはメソッドを追加しない（INTERFACE STABILITY: 非所有モックが
// CageRepository を実装しているため）。*cageRepositoryImpl が実装し、サービスは
// repo.(CageMergeRepository) の型アサーションで取得する。
type CageMergeRepository interface {
	// LockCagesForUpdate は ambient tx 内で医院スコープのケージ行を id 昇順
	// （デッドロック回避の決定的順序）に FOR UPDATE でロックして返す。
	// 存在しない・他医院の id が含まれる場合は NotFound。
	LockCagesForUpdate(ctx context.Context, clinicID uint64, ids []uint64) ([]model.Cage, error)
	// LockCageGroupForUpdate は代表ケージとそれを merged_into_cage_id で指す
	// 全構成ケージを1文の id 昇順 FOR UPDATE で原子的にロックして返す。
	// 構成集合の発見とロックを同一文で行うため、読み取りとロックの間に
	// メンバーが増減する競合窓がない。代表行が存在しない場合は NotFound。
	LockCageGroupForUpdate(ctx context.Context, clinicID, representativeID uint64) ([]model.Cage, error)
	// CountMergedMembers は merged_into_cage_id が representativeIDs に含まれる
	// ケージ数を代表 id ごとに集計して返す（代表がグループを率いているかの判定）。
	CountMergedMembers(ctx context.Context, clinicID uint64, representativeIDs []uint64) (map[uint64]int64, error)
	// CountOccupiedCages は未削除の admitted/reserved 入院をケージ id ごとに集計する。
	// CountUsageByCageID（退院済みも数える）とは別の、連結可否専用の占有判定。
	CountOccupiedCages(ctx context.Context, clinicID uint64, cageIDs []uint64) (map[uint64]int64, error)
	// ListOccupiedCageIDs は未削除の admitted/reserved 入院が1件以上乗っている
	// ケージ id を全件返す（連結ピッカーの占有表示専用。ページングなし）。
	ListOccupiedCageIDs(ctx context.Context, clinicID uint64) ([]uint64, error)
	// CountOccupiedCagesExcludingHospitalizations は CountOccupiedCages と同じ占有集計から
	// excludeHospitalizationIDs の入院行だけを除外する（EMR-461: 同一 tx で移動・解放される
	// 更新対象/交換当事者の占有を、他滞在の占有と誤認しないため）。空なら除外しない。
	CountOccupiedCagesExcludingHospitalizations(ctx context.Context, clinicID uint64, cageIDs, excludeHospitalizationIDs []uint64) (map[uint64]int64, error)
	// ListOccupiedCageIDsExcludingHospitalization は ListOccupiedCageIDs と同じ占有集合から
	// excludeHospitalizationID の入院行だけを除外する（EMR-461 入院フォームの空き候補 read）。
	ListOccupiedCageIDsExcludingHospitalization(ctx context.Context, clinicID uint64, excludeHospitalizationID *uint64) ([]uint64, error)
	// SetMergedInto は memberIDs の merged_into_cage_id に representativeID を設定する。
	SetMergedInto(ctx context.Context, clinicID, representativeID uint64, memberIDs []uint64) error
	// ClearMergedInto は representativeID を代表とする全構成ケージの
	// merged_into_cage_id を NULL に戻す。
	ClearMergedInto(ctx context.Context, clinicID, representativeID uint64) error
}

var _ CageMergeRepository = (*cageRepositoryImpl)(nil)

func (r *cageRepositoryImpl) LockCagesForUpdate(ctx context.Context, clinicID uint64, ids []uint64) ([]model.Cage, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("cage lock requires ambient transaction")
	}
	sortedIDs := slices.Clone(ids)
	slices.Sort(sortedIDs)
	sortedIDs = slices.Compact(sortedIDs)
	cages := make([]model.Cage, 0, len(sortedIDs))
	if len(sortedIDs) > 0 {
		// ORDER BY id ASC + FOR UPDATE: 行を取り出す順にロックされるため、
		// 全パスで同一の昇順ロック順序を保証する（デッドロック回避）。
		if err := persistence.DBOrTx(ctx, r.db).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Scopes(persistence.ClinicScope(clinicID)).
			Where("id IN ?", sortedIDs).
			Order("id ASC").
			Find(&cages).Error; err != nil {
			return nil, apperrors.FromGORM(err, "cage", "")
		}
	}
	if len(cages) != len(sortedIDs) {
		present := make(map[uint64]struct{}, len(cages))
		for _, cage := range cages {
			present[cage.ID] = struct{}{}
		}
		for _, id := range sortedIDs {
			if _, ok := present[id]; !ok {
				return nil, apperrors.WrapNotFound("cage", fmt.Sprintf("%d", id))
			}
		}
	}
	return cages, nil
}

func (r *cageRepositoryImpl) CountMergedMembers(ctx context.Context, clinicID uint64, representativeIDs []uint64) (map[uint64]int64, error) {
	counts := make(map[uint64]int64, len(representativeIDs))
	if len(representativeIDs) == 0 {
		return counts, nil
	}
	var rows []struct {
		RepresentativeID uint64 `gorm:"column:representative_id"`
		Members          int64  `gorm:"column:members"`
	}
	if err := persistence.DBOrTx(ctx, r.db).Model(&model.Cage{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Select("merged_into_cage_id AS representative_id, COUNT(*) AS members").
		Where("merged_into_cage_id IN ?", representativeIDs).
		Group("merged_into_cage_id").
		Scan(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "cage", "")
	}
	for _, row := range rows {
		counts[row.RepresentativeID] = row.Members
	}
	return counts, nil
}

func (r *cageRepositoryImpl) LockCageGroupForUpdate(ctx context.Context, clinicID, representativeID uint64) ([]model.Cage, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("cage lock requires ambient transaction")
	}
	cages := make([]model.Cage, 0)
	if err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ? OR merged_into_cage_id = ?", representativeID, representativeID).
		Order("id ASC").
		Find(&cages).Error; err != nil {
		return nil, apperrors.FromGORM(err, "cage", fmt.Sprintf("%d", representativeID))
	}
	found := false
	for i := range cages {
		if cages[i].ID == representativeID {
			found = true
			break
		}
	}
	if !found {
		return nil, apperrors.WrapNotFound("cage", fmt.Sprintf("%d", representativeID))
	}
	return cages, nil
}

func (r *cageRepositoryImpl) CountOccupiedCages(ctx context.Context, clinicID uint64, cageIDs []uint64) (map[uint64]int64, error) {
	counts := make(map[uint64]int64, len(cageIDs))
	if len(cageIDs) == 0 {
		return counts, nil
	}
	var rows []struct {
		CageID uint64 `gorm:"column:cage_id"`
		N      int64  `gorm:"column:n"`
	}
	if err := persistence.DBOrTx(ctx, r.db).Model(&model.Hospitalization{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Select("cage_id, COUNT(*) AS n").
		Where("cage_id IN ?", cageIDs).
		Where("deleted_at IS NULL").
		Where("status IN ?", []string{
			string(model.HospitalizationStatusAdmitted),
			string(model.HospitalizationStatusReserved),
		}).
		Group("cage_id").
		Scan(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "hospitalization", "")
	}
	for _, row := range rows {
		counts[row.CageID] = row.N
	}
	return counts, nil
}

func (r *cageRepositoryImpl) CountOccupiedCagesExcludingHospitalizations(ctx context.Context, clinicID uint64, cageIDs, excludeHospitalizationIDs []uint64) (map[uint64]int64, error) {
	counts := make(map[uint64]int64, len(cageIDs))
	if len(cageIDs) == 0 {
		return counts, nil
	}
	var rows []struct {
		CageID uint64 `gorm:"column:cage_id"`
		N      int64  `gorm:"column:n"`
	}
	q := persistence.DBOrTx(ctx, r.db).Model(&model.Hospitalization{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Select("cage_id, COUNT(*) AS n").
		Where("cage_id IN ?", cageIDs).
		Where("deleted_at IS NULL").
		Where("status IN ?", []string{
			string(model.HospitalizationStatusAdmitted),
			string(model.HospitalizationStatusReserved),
		})
	if len(excludeHospitalizationIDs) > 0 {
		q = q.Where("id NOT IN ?", excludeHospitalizationIDs)
	}
	if err := q.Group("cage_id").Scan(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "hospitalization", "")
	}
	for _, row := range rows {
		counts[row.CageID] = row.N
	}
	return counts, nil
}

func (r *cageRepositoryImpl) ListOccupiedCageIDsExcludingHospitalization(ctx context.Context, clinicID uint64, excludeHospitalizationID *uint64) ([]uint64, error) {
	ids := []uint64{}
	q := persistence.DBOrTx(ctx, r.db).Model(&model.Hospitalization{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Distinct().
		Where("cage_id IS NOT NULL").
		Where("deleted_at IS NULL").
		Where("status IN ?", []string{
			string(model.HospitalizationStatusAdmitted),
			string(model.HospitalizationStatusReserved),
		})
	if excludeHospitalizationID != nil {
		q = q.Where("id <> ?", *excludeHospitalizationID)
	}
	if err := q.Pluck("cage_id", &ids).Error; err != nil {
		return nil, apperrors.FromGORM(err, "hospitalization", "")
	}
	return ids, nil
}

func (r *cageRepositoryImpl) ListOccupiedCageIDs(ctx context.Context, clinicID uint64) ([]uint64, error) {
	var ids []uint64
	if err := persistence.DBOrTx(ctx, r.db).Model(&model.Hospitalization{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Distinct().
		Where("cage_id IS NOT NULL").
		Where("deleted_at IS NULL").
		Where("status IN ?", []string{
			string(model.HospitalizationStatusAdmitted),
			string(model.HospitalizationStatusReserved),
		}).
		Pluck("cage_id", &ids).Error; err != nil {
		return nil, apperrors.FromGORM(err, "hospitalization", "")
	}
	return ids, nil
}

func (r *cageRepositoryImpl) SetMergedInto(ctx context.Context, clinicID, representativeID uint64, memberIDs []uint64) error {
	if len(memberIDs) == 0 {
		return nil
	}
	// 行ロック済みだが、書込み側でも IS NULL 条件で再チェックしておく（二重防御）。
	result := persistence.DBOrTx(ctx, r.db).Model(&model.Cage{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id IN ? AND merged_into_cage_id IS NULL", memberIDs).
		Update("merged_into_cage_id", representativeID)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "cage", fmt.Sprintf("%d", representativeID))
	}
	if result.RowsAffected != int64(len(memberIDs)) {
		return apperrors.WrapConflict("ケージの連結状態が変化したため連結できません")
	}
	return nil
}

func (r *cageRepositoryImpl) ClearMergedInto(ctx context.Context, clinicID, representativeID uint64) error {
	result := persistence.DBOrTx(ctx, r.db).Model(&model.Cage{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("merged_into_cage_id = ?", representativeID).
		Update("merged_into_cage_id", nil)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "cage", fmt.Sprintf("%d", representativeID))
	}
	return nil
}
