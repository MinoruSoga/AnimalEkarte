package medicalrecord

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// CheckupFilters はクリニック横断一覧のフィルタ条件。
// Moved from internal/repository/checkup (BE8-4 batch24) — BE9-2D roll-up. Renamed from that
// subpackage's generic "Filters" to this entity-specific name to avoid collisions in this
// multi-repository package; every external caller only ever saw it via the internal/repository
// facade (CheckupFilters alias), so no call site changes.
type CheckupFilters struct {
	StartDate     *string
	EndDate       *string
	NextStartDate *string
	NextEndDate   *string
}

// CheckupRepository is the data access interface for checkups (健診記録).
// Moved from internal/repository/checkup (BE8-4 batch24) — BE9-2D roll-up. Renamed from that
// subpackage's generic "Repository" to this entity-specific name only because medicalrecord
// holds multiple repository interfaces in one package; every external caller only ever saw
// this name via the internal/repository facade, so no call site changes.
type CheckupRepository interface {
	FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error)
	FindByClinicID(ctx context.Context, clinicID uint64, filters CheckupFilters, page, limit int) ([]model.Checkup, int64, error)
	// FindByOwnerID は飼い主に紐づく生存健診記録を全件返す（ISSUE-004 タグ再同期用）。
	// medical_records.pet_id から現在の pets.owner_id を解決する。
	FindByOwnerID(ctx context.Context, clinicID, ownerID uint64) ([]model.Checkup, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Checkup, error)
	LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Checkup, error)
	Create(ctx context.Context, checkup *model.Checkup) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type checkupRepository struct {
	db *gorm.DB
}

// NewCheckupRepository は CheckupRepository を初期化して返す。
func NewCheckupRepository(db *gorm.DB) CheckupRepository {
	return &checkupRepository{db: db}
}

func (r *checkupRepository) FindByClinicID(ctx context.Context, clinicID uint64, filters CheckupFilters, page, limit int) ([]model.Checkup, int64, error) {
	buildBase := func() *gorm.DB {
		q := persistence.DBOrTx(ctx, r.db).Model(&model.Checkup{}).
			Scopes(
				persistence.ClinicScope(clinicID),
				checkupPatientRelationsScope(clinicID),
			)
		if filters.StartDate != nil {
			q = q.Where("date >= ?", *filters.StartDate)
		}
		if filters.EndDate != nil {
			q = q.Where("date <= ?", *filters.EndDate)
		}
		if filters.NextStartDate != nil {
			q = q.Where("next_date >= ?", *filters.NextStartDate)
		}
		if filters.NextEndDate != nil {
			q = q.Where("next_date <= ?", *filters.NextEndDate)
		}
		return q
	}

	var total int64
	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "checkup", "")
	}

	checkups := make([]model.Checkup, 0)
	err := buildBase().
		Scopes(checkupReadPreloads(clinicID)).
		Order("date DESC").
		Scopes(paginate(page, limit)).
		Find(&checkups).Error
	if err != nil {
		return nil, 0, apperrors.FromGORM(err, "checkup", "")
	}
	return checkups, total, nil
}

// healthTagOwnerHistoryMax is a safety cap for LSTEP health-tag resync history
// loads (G2F-02). Order is newest-first so lookback windows still see recent rows.
const healthTagOwnerHistoryMax = 500

// FindByOwnerID は現在飼主のペットに紐づく生存健診記録を返す（ISSUE-004 タグ再同期用）。
// medical_records.pet_id から pets.owner_id を解決し、checkup と medical_record の両方が生存しているレコードのみ返す。
// G2F-02: newest-first hard Limit to avoid unbounded owner history materialization.
func (r *checkupRepository) FindByOwnerID(ctx context.Context, clinicID, ownerID uint64) ([]model.Checkup, error) {
	checkups := make([]model.Checkup, 0)
	err := persistence.DBOrTx(ctx, r.db).
		Joins("JOIN medical_records ON medical_records.id = checkups.medical_record_id"+
			" AND medical_records.clinic_id = ?"+
			" AND medical_records.deleted_at IS NULL", clinicID).
		Where(`
			checkups.clinic_id = ?
			AND EXISTS (
				SELECT 1
				FROM pets current_owner_pet
				JOIN owners current_owner
				  ON current_owner.id = current_owner_pet.owner_id
				 AND current_owner.clinic_id = current_owner_pet.clinic_id
				WHERE current_owner_pet.id = medical_records.pet_id
				  AND current_owner_pet.clinic_id = medical_records.clinic_id
				  AND current_owner.id = ?
			)
		`, clinicID, ownerID).
		Scopes(checkupPatientRelationsScope(clinicID)).
		Preload("CheckupType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Order("checkups.date DESC").
		Limit(healthTagOwnerHistoryMax).
		Find(&checkups).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup", fmt.Sprintf("owner=%d", ownerID))
	}
	return checkups, nil
}

// FindByOwnerIDs は複数飼い主の生存健診記録を clinic スコープで一括取得し owner_id 別に返す
// （G2F-02 / BE-ACT-LSTEP-HEALTH-BATCH-BULK）。各 owner は newest-first で
// healthTagOwnerHistoryMax 件に cap する。空 ownerIDs は空 map を即返す。
//
// NOTE: CheckupRepository インタフェースには載せない（consumer は型アサーションで利用）。
// バッチ経路は ambient tx を持たないため r.db.WithContext を使う（dbOrTx inventory 非参加）。
func (r *checkupRepository) FindByOwnerIDs(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64][]model.Checkup, error) {
	result := make(map[uint64][]model.Checkup, len(ownerIDs))
	if len(ownerIDs) == 0 {
		return result, nil
	}

	type rankedRow struct {
		ID      uint64
		OwnerID uint64
	}
	var ranked []rankedRow
	// Per-owner newest-first cap via window function so page bulk stays memory-bounded.
	err := r.db.WithContext(ctx).Raw(`
		SELECT id, owner_id FROM (
			SELECT checkups.id AS id,
			       current_owner_pet.owner_id AS owner_id,
			       ROW_NUMBER() OVER (
			         PARTITION BY current_owner_pet.owner_id
			         ORDER BY checkups.date DESC
			       ) AS rn
			FROM checkups
			INNER JOIN medical_records
			  ON medical_records.id = checkups.medical_record_id
			 AND medical_records.clinic_id = ?
			 AND medical_records.deleted_at IS NULL
			INNER JOIN pets current_owner_pet
			  ON current_owner_pet.id = medical_records.pet_id
			 AND current_owner_pet.clinic_id = medical_records.clinic_id
			WHERE checkups.clinic_id = ?
			  AND checkups.deleted_at IS NULL
			  AND current_owner_pet.owner_id IN ?
		) ranked
		WHERE rn <= ?
	`, clinicID, clinicID, ownerIDs, healthTagOwnerHistoryMax).Scan(&ranked).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup", "find_by_owner_ids")
	}
	if len(ranked) == 0 {
		return result, nil
	}

	ids := make([]uint64, len(ranked))
	ownerByCheckupID := make(map[uint64]uint64, len(ranked))
	for i, row := range ranked {
		ids[i] = row.ID
		ownerByCheckupID[row.ID] = row.OwnerID
	}

	checkups := make([]model.Checkup, 0, len(ids))
	err = r.db.WithContext(ctx).
		Where("checkups.id IN ?", ids).
		Where("checkups.clinic_id = ?", clinicID).
		Scopes(checkupPatientRelationsScope(clinicID)).
		Preload("CheckupType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Order("checkups.date DESC").
		Find(&checkups).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup", "find_by_owner_ids")
	}

	for i := range checkups {
		ownerID, ok := ownerByCheckupID[checkups[i].ID]
		if !ok {
			continue
		}
		result[ownerID] = append(result[ownerID], checkups[i])
	}
	return result, nil
}

// FindOwnerVisitSummariesByOwnerIDs は複数飼い主の診療集計を clinic スコープで一括取得する
// （G2F-02 health-prevention page bulk）。MedicalRecordRepository インタフェース外の
// concrete メソッド。結果に含まれない owner は来院 0 件として扱う。
//
// Co-located in this file under health-tag bulk ownership (visit summary repo is reference-only).
func (r *medicalRecordRepository) FindOwnerVisitSummariesByOwnerIDs(
	ctx context.Context,
	clinicID uint64,
	ownerIDs []uint64,
) (map[uint64]*OwnerVisitSummary, error) {
	result := make(map[uint64]*OwnerVisitSummary, len(ownerIDs))
	if len(ownerIDs) == 0 {
		return result, nil
	}
	type row struct {
		OwnerID      uint64
		FirstVisitAt *time.Time
		LastVisitAt  *time.Time
		TotalCount   int64
		AnnualCount  int64
	}
	var rows []row
	oneYearAgo := time.Now().In(time.Local).AddDate(-1, 0, 0)
	err := r.db.WithContext(ctx).
		Table("medical_records AS mr").
		Joins("JOIN pets AS current_owner_pet ON current_owner_pet.id = mr.pet_id AND current_owner_pet.clinic_id = mr.clinic_id").
		Joins("JOIN owners AS o ON o.id = current_owner_pet.owner_id AND o.clinic_id = mr.clinic_id AND o.deleted_at IS NULL").
		Where("mr.clinic_id = ? AND mr.deleted_at IS NULL AND current_owner_pet.owner_id IN ?", clinicID, ownerIDs).
		Select(`current_owner_pet.owner_id AS owner_id,
			MIN(mr.date) AS first_visit_at,
			MAX(mr.date) AS last_visit_at,
			COUNT(*) AS total_count,
			COUNT(CASE WHEN mr.date >= ? THEN 1 END) AS annual_count`, oneYearAgo).
		Group("current_owner_pet.owner_id").
		Scan(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medical_record", "find_owner_visit_summaries_by_owner_ids")
	}
	for i := range rows {
		result[rows[i].OwnerID] = &OwnerVisitSummary{
			FirstVisitAt: rows[i].FirstVisitAt,
			LastVisitAt:  rows[i].LastVisitAt,
			TotalCount:   rows[i].TotalCount,
			AnnualCount:  rows[i].AnnualCount,
		}
	}
	return result, nil
}

func (r *checkupRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error) {
	checkups := make([]model.Checkup, 0)
	err := persistence.DBOrTx(ctx, r.db).
		Joins("JOIN medical_records ON medical_records.id = checkups.medical_record_id"+
			" AND medical_records.clinic_id = ?"+
			" AND medical_records.deleted_at IS NULL", clinicID).
		Where("checkups.clinic_id = ? AND checkups.medical_record_id = ?", clinicID, medicalRecordID).
		Scopes(checkupPatientRelationsScope(clinicID)).
		Preload("CheckupType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Doctor", "deleted_at IS NULL AND is_active = TRUE AND staff_type = ?", model.StaffTypeDoctor).
		Order("checkups.date ASC").
		Find(&checkups).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup", "")
	}
	return checkups, nil
}

func (r *checkupRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Checkup, error) {
	var checkup model.Checkup
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(
			persistence.ClinicScope(clinicID),
			checkupPatientRelationsScope(clinicID),
			checkupReadPreloads(clinicID),
		).
		Where("id = ?", id).
		First(&checkup).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup", fmt.Sprintf("%d", id))
	}
	return &checkup, nil
}

// LockByIDForUpdate serializes checkup deletion with other writes under the
// caller's medical-record-first transaction lock order.
func (r *checkupRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Checkup, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("checkup lock requires an ambient transaction")
	}
	var checkup model.Checkup
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		First(&checkup).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup", fmt.Sprintf("%d", id))
	}
	return &checkup, nil
}

func (r *checkupRepository) Create(ctx context.Context, checkup *model.Checkup) error {
	err := persistence.DBOrTx(ctx, r.db).Create(checkup).Error
	if err != nil {
		return apperrors.FromGORM(err, "checkup", "")
	}
	return nil
}

func (r *checkupRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return persistence.UpdateScopedByID(ctx, persistence.DBOrTx(ctx, r.db), &model.Checkup{}, "checkup", clinicID, id, fields)
}

func (r *checkupRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return persistence.DeleteScopedByID(ctx, persistence.DBOrTx(ctx, r.db), &model.Checkup{}, "checkup", clinicID, id)
}

func checkupReadPreloads(clinicID uint64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.
			Preload("CheckupType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
			Preload("Doctor", "deleted_at IS NULL AND is_active = TRUE AND staff_type = ?", model.StaffTypeDoctor).
			Preload("MedicalRecord", "clinic_id = ? AND deleted_at IS NULL", clinicID).
			Preload("MedicalRecord.Pet", "clinic_id = ? AND deleted_at IS NULL", clinicID).
			Preload("MedicalRecord.Pet.Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID)
	}
}

// checkupPatientRelationsScope excludes polluted rows before their raw foreign IDs can
// reach an HTTP response. The required medical record and checkup type, plus every
// non-nil patient relation, must resolve inside the checkup clinic.
func checkupPatientRelationsScope(clinicID uint64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(`
			EXISTS (
				SELECT 1 FROM checkup_types scoped_checkup_type
				WHERE scoped_checkup_type.id = checkups.checkup_type_id
				  AND scoped_checkup_type.clinic_id = ?
				  AND scoped_checkup_type.deleted_at IS NULL
			)
			AND EXISTS (
				SELECT 1 FROM medical_records scoped_record
				WHERE scoped_record.id = checkups.medical_record_id
				  AND scoped_record.clinic_id = ?
				  AND scoped_record.deleted_at IS NULL
				  AND (
					scoped_record.owner_id IS NULL OR EXISTS (
						SELECT 1 FROM owners scoped_record_owner
						WHERE scoped_record_owner.id = scoped_record.owner_id
						  AND scoped_record_owner.clinic_id = ?
						  AND scoped_record_owner.deleted_at IS NULL
					)
				  )
				  AND (
					scoped_record.pet_id IS NULL OR EXISTS (
						SELECT 1 FROM pets scoped_record_pet
						JOIN owners scoped_record_pet_owner
						  ON scoped_record_pet_owner.id = scoped_record_pet.owner_id
						 AND scoped_record_pet_owner.clinic_id = scoped_record_pet.clinic_id
						 AND scoped_record_pet_owner.deleted_at IS NULL
						WHERE scoped_record_pet.id = scoped_record.pet_id
						  AND scoped_record_pet.clinic_id = ?
						  AND scoped_record_pet.deleted_at IS NULL
					)
				  )
				  AND (
					checkups.pet_id IS NULL OR
					scoped_record.pet_id = checkups.pet_id
				  )
			)
				AND (
					checkups.pet_id IS NULL OR EXISTS (
					SELECT 1 FROM pets scoped_pet
					JOIN owners scoped_pet_owner
					  ON scoped_pet_owner.id = scoped_pet.owner_id
					 AND scoped_pet_owner.clinic_id = scoped_pet.clinic_id
					 AND scoped_pet_owner.deleted_at IS NULL
					WHERE scoped_pet.id = checkups.pet_id
					  AND scoped_pet.clinic_id = ?
					  AND scoped_pet.deleted_at IS NULL
					)
				)
				AND (
					checkups.doctor_id IS NULL OR EXISTS (
						SELECT 1 FROM staff_clinic_assignments scoped_doctor_assignment
						JOIN staffs scoped_doctor
						  ON scoped_doctor.id = scoped_doctor_assignment.staff_id
						 AND scoped_doctor.deleted_at IS NULL
						 AND scoped_doctor.is_active = TRUE
						 AND scoped_doctor.staff_type = ?
						WHERE scoped_doctor_assignment.staff_id = checkups.doctor_id
						  AND scoped_doctor_assignment.clinic_id = ?
						  AND scoped_doctor_assignment.deleted_at IS NULL
					)
				)
			`, clinicID, clinicID, clinicID, clinicID, clinicID, model.StaffTypeDoctor, clinicID)
	}
}
