package medicalrecord

// Moved from internal/repository (BE9-2D ⑤ Batch A). 旧 package-private helper は repohelpers 同等物へ
// 置換（同一述語/ambient-tx参加）。外部呼び出しは internal/repository の facade alias 経由で不変。

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/commonbasic"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type HospitalizationRepository interface {
	FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, plannedEntryFrom, plannedEntryTo *string, page, limit int) ([]model.Hospitalization, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error)
	// LockByIDForUpdate は FOR UPDATE で入院行をロック取得する（Discharge Q2-C）。
	// Repositories.Transaction（repo-swap）内の txRepos 経由でのみ呼ぶこと。
	LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error)
	Create(ctx context.Context, hospitalization *model.Hospitalization) error
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateHospitalizationInput) (*model.Hospitalization, error)
	// CreateStayCase writes the stay_cases row that a new hospitalization points to (EMR-499 U1).
	// It writes no movement event: effective state stays unknown until the lifecycle writer exists.
	CreateStayCase(ctx context.Context, stay *model.StayCase) error
	// FindStayCasesByIDs returns live (not deleted) stay cases of the clinic, keyed by stay case ID.
	FindStayCasesByIDs(ctx context.Context, clinicID uint64, ids []uint64) (map[uint64]model.StayCase, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	CountCarePlanItemsByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error)
	CountDailyRecordsByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error)
	CountTreatmentPlansByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error)
}

type hospitalizationRepository struct {
	db *gorm.DB
}

func NewHospitalizationRepository(db *gorm.DB) HospitalizationRepository {
	return &hospitalizationRepository{db: db}
}

// hospitalizationListOrder sorts by the planned entry of the stay case (retired start_date), rows without
// a stay case last. A correlated subquery avoids a join that would make clinic_id/pet_id ambiguous.
const hospitalizationListOrder = `(SELECT sc.planned_entry_at FROM stay_cases sc WHERE sc.id = hospitalizations.stay_case_id
	AND sc.clinic_id = hospitalizations.clinic_id AND sc.deleted_at IS NULL) DESC NULLS LAST,
	hospitalizations.created_at DESC, hospitalizations.id DESC`

func hospitalizationListQuery(db *gorm.DB, clinicID uint64, filters listHospitalizationFilters) *gorm.DB {
	q := db.
		Model(&model.Hospitalization{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("EXISTS (SELECT 1 FROM pets p WHERE p.id = hospitalizations.pet_id AND p.clinic_id = hospitalizations.clinic_id)")
	if filters.PetID != nil {
		q = q.Where("pet_id = ?", *filters.PetID)
	}
	if filters.OwnerID != nil {
		q = q.Where(`
			EXISTS (
				SELECT 1
				FROM pets current_owner_pet
				JOIN owners current_owner
				  ON current_owner.id = current_owner_pet.owner_id
				 AND current_owner.clinic_id = current_owner_pet.clinic_id
				WHERE current_owner_pet.id = hospitalizations.pet_id
				  AND current_owner_pet.clinic_id = hospitalizations.clinic_id
				  AND current_owner.id = ?
			)
		`, *filters.OwnerID)
	}
	// Planned-entry range (JST calendar days, the half-open day interval of the spec's actual-day convention,
	// docs/spec/domains/cash-register/business-rules.md#closing-discharge-actual-jst-day-20261017).
	// Rows without a live stay case never match. This filters PLANNED entry only: actual admission is derived.
	if filters.PlannedEntryFrom != nil {
		from, err := time.ParseInLocation(time.DateOnly, *filters.PlannedEntryFrom, config.JST)
		if err != nil {
			_ = q.AddError(apperrors.WrapInvalidInput("planned_entry_from は YYYY-MM-DD 形式で入力してください"))
			return q
		}
		q = q.Where(`EXISTS (SELECT 1 FROM stay_cases sc WHERE sc.id = hospitalizations.stay_case_id
			AND sc.clinic_id = hospitalizations.clinic_id AND sc.deleted_at IS NULL AND sc.planned_entry_at >= ?)`, from)
	}
	if filters.PlannedEntryTo != nil {
		to, err := time.ParseInLocation(time.DateOnly, *filters.PlannedEntryTo, config.JST)
		if err != nil {
			_ = q.AddError(apperrors.WrapInvalidInput("planned_entry_to は YYYY-MM-DD 形式で入力してください"))
			return q
		}
		q = q.Where(`EXISTS (SELECT 1 FROM stay_cases sc WHERE sc.id = hospitalizations.stay_case_id
			AND sc.clinic_id = hospitalizations.clinic_id AND sc.deleted_at IS NULL AND sc.planned_entry_at < ?)`, to.AddDate(0, 0, 1))
	}
	return q
}

func hospitalizationListPreloads(q *gorm.DB, clinicID uint64) *gorm.DB {
	return q.
		Preload("Pet", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Cage", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Doctor", "deleted_at IS NULL")
}

func (r *hospitalizationRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, plannedEntryFrom, plannedEntryTo *string, page, limit int) ([]model.Hospitalization, int64, error) {
	hospitalizations := make([]model.Hospitalization, 0)
	var total int64

	filters := listHospitalizationFilters{
		PetID: petID, OwnerID: ownerID, PlannedEntryFrom: plannedEntryFrom, PlannedEntryTo: plannedEntryTo,
	}
	q := hospitalizationListQuery(r.db.WithContext(ctx), clinicID, filters)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "hospitalization", "")
	}
	if err := hospitalizationListPreloads(q, clinicID).
		Scopes(paginate(page, limit)).Order(hospitalizationListOrder).
		Find(&hospitalizations).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "hospitalization", "")
	}
	if err := hydrateHospitalizationBasics(ctx, r.db, []uint64{clinicID}, commonbasic.Ptrs(hospitalizations)); err != nil {
		return nil, 0, err
	}
	return hospitalizations, total, nil
}

func (r *hospitalizationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	var hospitalization model.Hospitalization
	err := persistence.DBOrTx(ctx, r.db).
		Preload("Pet", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Cage", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Doctor", "deleted_at IS NULL").
		Preload("CarePlanItems").
		// MRB-03: clinic-owned child tables must carry clinic scope (BUG-437 / SEC-SWEEP-02).
		Preload("DailyRecords", "clinic_id = ?", clinicID).
		Preload("TreatmentPlans", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&hospitalization).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "hospitalization", fmt.Sprintf("%d", id))
	}
	if err := hydrateHospitalizationBasics(ctx, r.db, []uint64{clinicID}, []*model.Hospitalization{&hospitalization}); err != nil {
		return nil, err
	}
	return &hospitalization, nil
}

// LockByIDForUpdate は FOR UPDATE で入院を行ロック取得する（Discharge Q2-C）。
// OwnerID/PetID など Q2-A 再検証に必要なスカラーはロック取得時の行スナップショットに含まれる。
// DischargeWithBilling は Repositories.Transaction（repo-swap）内の txRepos 経由で呼ぶこと。
// r.db が tx にバインドされていないとロックは SELECT 終了と同時に解放され直列化できない。
// LockByIDForUpdate / UpdateIfNotDischarged は dbOrTx で ambient tx（Transactor.WithTx）に参加する
// （BE9-2D ⑤: hospitalizationService.DischargeWithBilling の repos.Transaction→WithTx 化に伴い、
// FOR UPDATE 直列化と退院status更新をbilling書込と同一 tx に保つ＝二重会計防止の要）。
func (r *hospitalizationRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	// MRB-02: FOR UPDATE is meaningless outside an ambient transaction (autocommit releases
	// the row lock at statement end). Match examinationRepository.LockByIDForUpdate fail-closed.
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("hospitalization lock requires an ambient transaction")
	}
	var hospitalization model.Hospitalization
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).First(&hospitalization).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "hospitalization", fmt.Sprintf("%d", id))
	}
	return &hospitalization, nil
}

func (r *hospitalizationRepository) Create(ctx context.Context, hospitalization *model.Hospitalization) error {
	err := persistence.DBOrTx(ctx, r.db).Create(hospitalization).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("hospitalization", fmt.Sprintf("stay_case=%v", hospitalization.StayCaseID))
		}
		return apperrors.FromGORM(err, "hospitalization", "")
	}
	return nil
}

func (r *hospitalizationRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateHospitalizationInput) (*model.Hospitalization, error) {
	if err := r.update(ctx, clinicID, id, buildHospitalizationUpdate(&cmd)); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *hospitalizationRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.Hospitalization{}).
		Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "hospitalization", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("hospitalization", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *hospitalizationRepository) CreateStayCase(ctx context.Context, stay *model.StayCase) error {
	if stay == nil {
		return apperrors.WrapInvalidInput("stay case must not be nil")
	}
	if err := persistence.DBOrTx(ctx, r.db).Create(stay).Error; err != nil {
		return apperrors.FromGORM(err, "stay_case", "")
	}
	return nil
}

func (r *hospitalizationRepository) FindStayCasesByIDs(ctx context.Context, clinicID uint64, ids []uint64) (map[uint64]model.StayCase, error) {
	result := make(map[uint64]model.StayCase, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	var rows []model.StayCase
	if err := persistence.DBOrTx(ctx, r.db).
		Where("clinic_id = ? AND id IN ? AND deleted_at IS NULL", clinicID, ids).
		Find(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "stay_case", "")
	}
	for _, row := range rows {
		result[row.ID] = row
	}
	return result, nil
}

func (r *hospitalizationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	// Ambient-tx participation so Delete + fail-closed audit share one transaction (MRB-05).
	result := persistence.DBOrTx(ctx, r.db).Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).Delete(&model.Hospitalization{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "hospitalization", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("hospitalization", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *hospitalizationRepository) CountCarePlanItemsByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.CarePlanItem{}).
		Joins("JOIN hospitalizations ON care_plan_items.hospitalization_id = hospitalizations.id AND hospitalizations.deleted_at IS NULL").
		Where("hospitalizations.clinic_id = ? AND care_plan_items.hospitalization_id = ?", clinicID, hospitalizationID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "care_plan_item", fmt.Sprintf("hospitalization_id=%d", hospitalizationID))
	}
	return count, nil
}

func (r *hospitalizationRepository) CountDailyRecordsByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.DailyRecord{}).
		Joins("JOIN hospitalizations ON daily_records.hospitalization_id = hospitalizations.id AND daily_records.clinic_id = hospitalizations.clinic_id AND hospitalizations.deleted_at IS NULL").
		Where("hospitalizations.clinic_id = ? AND daily_records.clinic_id = ? AND daily_records.hospitalization_id = ?", clinicID, clinicID, hospitalizationID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "daily_record", fmt.Sprintf("hospitalization_id=%d", hospitalizationID))
	}
	return count, nil
}

func (r *hospitalizationRepository) CountTreatmentPlansByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.TreatmentPlan{}).
		Joins("JOIN hospitalizations ON hospitalizations.id = treatment_plans.hospitalization_id AND hospitalizations.clinic_id = treatment_plans.clinic_id").
		Where("treatment_plans.clinic_id = ? AND treatment_plans.hospitalization_id = ? AND treatment_plans.deleted_at IS NULL", clinicID, hospitalizationID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "treatment_plan", fmt.Sprintf("hospitalization_id=%d", hospitalizationID))
	}
	return count, nil
}

// isUniqueConstraintErr は internal/repository/db.go の同名ヘルパーの最小複製
// （BE9-2D ⑤ Batch A: 原本は旧 package の他 repo が引き続き使うため移動不可。PostgreSQL 23505 判定の純関数）。
func isUniqueConstraintErr(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
