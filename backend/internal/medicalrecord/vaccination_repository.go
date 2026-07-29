package medicalrecord

// vaccination_repository.go — VaccinationRepository の実装。
// Moved from internal/repository/vaccination_repository.go — BE9-2D roll-up. Type/constructor
// names are unchanged; the internal/repository facade re-exports them as aliases so no caller
// changes. Package-private helpers (updateScopedByID/deleteScopedByID) are swapped for their
// repohelpers equivalents (this package must not import internal/repository — the facade there
// imports medicalrecord, so the reverse is an import cycle). paginate is medicalrecord-local.

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type VaccinationRepository interface {
	FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Vaccination, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error)
	LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error)
	// FindByOwner は飼い主の生存ワクチン記録（ペット経由）を全件返す（ISSUE-004 タグ再同期用）。
	// 飼い主のペットがすべて削除済みの場合は空配列を返す。
	FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Vaccination, error)
	Create(ctx context.Context, vaccination *model.Vaccination) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccination, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	// FindOwnersByVaccineDeadline はワクチン次回接種日（next_date）が targetDate の飼い主IDリストを返す（FEAT-383）。
	FindOwnersByVaccineDeadline(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error)
}

type vaccinationRepository struct {
	db *gorm.DB
}

func NewVaccinationRepository(db *gorm.DB) VaccinationRepository {
	return &vaccinationRepository{db: db}
}

func (r *vaccinationRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Vaccination, int64, error) {
	buildBase := func() *gorm.DB {
		q := persistence.DBOrTx(ctx, r.db).Model(&model.Vaccination{}).
			Where("vaccinations.clinic_id = ?", clinicID).
			Scopes(vaccinationPatientRelationsScope(clinicID))
		if petID != nil || ownerID != nil {
			q = q.Joins("JOIN pets ON pets.id = vaccinations.pet_id AND pets.clinic_id = vaccinations.clinic_id AND pets.deleted_at IS NULL")
		}
		if petID != nil {
			q = q.Where("vaccinations.pet_id = ?", *petID)
		}
		if ownerID != nil {
			q = q.Joins("JOIN owners ON owners.id = pets.owner_id AND owners.clinic_id = vaccinations.clinic_id AND owners.deleted_at IS NULL").
				Where("owners.id = ?", *ownerID)
		}
		if startDate != nil {
			q = q.Where("vaccinations.date >= ?", *startDate)
		}
		if endDate != nil {
			q = q.Where("vaccinations.date <= ?", *endDate)
		}
		return q
	}

	var total int64
	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "vaccination", "")
	}

	vaccinations := make([]model.Vaccination, 0)
	if err := vaccinationReadPreloads(buildBase(), clinicID).
		Scopes(paginate(page, limit)).Order("vaccinations.date DESC, vaccinations.created_at DESC").
		Find(&vaccinations).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "vaccination", "")
	}
	return vaccinations, total, nil
}

// FindByOwner は飼い主の全生存ワクチン記録を返す（ISSUE-004 タグ再同期用）。
// pets テーブルを JOIN し、飼い主に属する生存ペットのワクチンのみ取得する。
// G2F-02: newest-first hard Limit (healthTagOwnerHistoryMax) to avoid unbounded history.
func (r *vaccinationRepository) FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Vaccination, error) {
	vaccinations := make([]model.Vaccination, 0)
	err := persistence.DBOrTx(ctx, r.db).
		Joins("JOIN pets ON pets.id = vaccinations.pet_id AND pets.clinic_id = vaccinations.clinic_id AND pets.deleted_at IS NULL").
		Joins("JOIN owners ON owners.id = pets.owner_id AND owners.clinic_id = vaccinations.clinic_id AND owners.deleted_at IS NULL").
		Where("vaccinations.clinic_id = ? AND owners.id = ?", clinicID, ownerID).
		Scopes(vaccinationPatientRelationsScope(clinicID)).
		Preload("Vaccine", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Order("vaccinations.date DESC, vaccinations.created_at DESC").
		Limit(healthTagOwnerHistoryMax).
		Find(&vaccinations).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "vaccination", fmt.Sprintf("owner=%d", ownerID))
	}
	return vaccinations, nil
}

// FindByOwnerIDs は複数飼い主の生存ワクチン記録を clinic スコープで一括取得し owner_id 別に返す
// （G2F-02 / BE-ACT-LSTEP-HEALTH-BATCH-BULK）。各 owner は newest-first で
// healthTagOwnerHistoryMax 件に cap する。空 ownerIDs は空 map を即返す。
//
// NOTE: VaccinationRepository インタフェースには載せない（consumer は型アサーションで利用）。
// バッチ経路は ambient tx を持たないため r.db.WithContext を使う（dbOrTx inventory 非参加）。
func (r *vaccinationRepository) FindByOwnerIDs(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64][]model.Vaccination, error) {
	result := make(map[uint64][]model.Vaccination, len(ownerIDs))
	if len(ownerIDs) == 0 {
		return result, nil
	}

	type rankedRow struct {
		ID      uint64
		OwnerID uint64
	}
	var ranked []rankedRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT id, owner_id FROM (
			SELECT vaccinations.id AS id,
			       pets.owner_id AS owner_id,
			       ROW_NUMBER() OVER (
			         PARTITION BY pets.owner_id
			         ORDER BY vaccinations.date DESC, vaccinations.created_at DESC
			       ) AS rn
			FROM vaccinations
			INNER JOIN pets
			  ON pets.id = vaccinations.pet_id
			 AND pets.clinic_id = vaccinations.clinic_id
			 AND pets.deleted_at IS NULL
			INNER JOIN owners
			  ON owners.id = pets.owner_id
			 AND owners.clinic_id = vaccinations.clinic_id
			 AND owners.deleted_at IS NULL
			WHERE vaccinations.clinic_id = ?
			  AND vaccinations.deleted_at IS NULL
			  AND pets.owner_id IN ?
		) ranked
		WHERE rn <= ?
	`, clinicID, ownerIDs, healthTagOwnerHistoryMax).Scan(&ranked).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "vaccination", "find_by_owner_ids")
	}
	if len(ranked) == 0 {
		return result, nil
	}

	ids := make([]uint64, len(ranked))
	ownerByVacID := make(map[uint64]uint64, len(ranked))
	for i, row := range ranked {
		ids[i] = row.ID
		ownerByVacID[row.ID] = row.OwnerID
	}

	vaccinations := make([]model.Vaccination, 0, len(ids))
	err = r.db.WithContext(ctx).
		Where("vaccinations.id IN ?", ids).
		Where("vaccinations.clinic_id = ?", clinicID).
		Scopes(vaccinationPatientRelationsScope(clinicID)).
		Preload("Vaccine", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Order("vaccinations.date DESC, vaccinations.created_at DESC").
		Find(&vaccinations).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "vaccination", "find_by_owner_ids")
	}

	for i := range vaccinations {
		ownerID, ok := ownerByVacID[vaccinations[i].ID]
		if !ok {
			continue
		}
		result[ownerID] = append(result[ownerID], vaccinations[i])
	}
	return result, nil
}

func (r *vaccinationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error) {
	var vaccination model.Vaccination
	err := vaccinationReadPreloads(persistence.DBOrTx(ctx, r.db), clinicID).
		Where("vaccinations.id = ? AND vaccinations.clinic_id = ?", id, clinicID).
		Scopes(vaccinationPatientRelationsScope(clinicID)).
		First(&vaccination).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "vaccination", fmt.Sprintf("%d", id))
	}
	return &vaccination, nil
}

func (r *vaccinationRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("vaccination lock requires an ambient transaction")
	}
	var vaccination model.Vaccination
	err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("vaccinations.id = ? AND vaccinations.clinic_id = ?", id, clinicID).
		First(&vaccination).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "vaccination", fmt.Sprintf("%d", id))
	}
	return &vaccination, nil
}

func (r *vaccinationRepository) Create(ctx context.Context, vaccination *model.Vaccination) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(vaccination).Error; err != nil {
		return apperrors.FromGORM(err, "vaccination", "")
	}
	return nil
}

func (r *vaccinationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccination, error) {
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.Vaccination{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "vaccination", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("vaccination", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *vaccinationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return persistence.DeleteScopedByID(ctx, persistence.DBOrTx(ctx, r.db), &model.Vaccination{}, "vaccination", clinicID, id)
}

const vaccinationAssignedDoctorCond = "staffs.deleted_at IS NULL AND staffs.is_active = TRUE AND staffs.staff_type = 'doctor' AND EXISTS (SELECT 1 FROM staff_clinic_assignments sca WHERE sca.staff_id = staffs.id AND sca.clinic_id = ? AND sca.deleted_at IS NULL)"

func vaccinationReadPreloads(db *gorm.DB, clinicID uint64) *gorm.DB {
	return db.
		Preload("Vaccine", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Pet", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Pet.Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Doctor", vaccinationAssignedDoctorCond, clinicID)
}

// vaccinationPatientRelationsScope excludes directly polluted rows without changing the
// nullable Pet/MedicalRecord contract. Non-nil patient relations must resolve entirely inside
// the vaccination clinic, and a linked medical record must describe the same patient.
func vaccinationPatientRelationsScope(clinicID uint64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(`
			EXISTS (
				SELECT 1 FROM vaccines scoped_vaccine
				WHERE scoped_vaccine.id = vaccinations.vaccine_id
				  AND scoped_vaccine.clinic_id = ?
				  AND scoped_vaccine.deleted_at IS NULL
			)
			AND (
				vaccinations.pet_id IS NULL OR EXISTS (
					SELECT 1 FROM pets scoped_pet
					JOIN owners scoped_pet_owner
					  ON scoped_pet_owner.id = scoped_pet.owner_id
					 AND scoped_pet_owner.clinic_id = scoped_pet.clinic_id
					 AND scoped_pet_owner.deleted_at IS NULL
					WHERE scoped_pet.id = vaccinations.pet_id
					  AND scoped_pet.clinic_id = ?
					  AND scoped_pet.deleted_at IS NULL
				)
			)
			AND (
				vaccinations.medical_record_id IS NULL OR EXISTS (
					SELECT 1 FROM medical_records scoped_record
					WHERE scoped_record.id = vaccinations.medical_record_id
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
					  AND (vaccinations.pet_id IS NULL OR scoped_record.pet_id = vaccinations.pet_id)
				)
			)
		`, clinicID, clinicID, clinicID, clinicID, clinicID)
	}
}

// FindOwnersByVaccineDeadline はワクチン次回接種日（next_date）が targetDate の飼い主IDリストを返す（FEAT-383）。
// pets/owners を clinic_id 一致で JOIN し、生存ペット経由で飼い主IDを取得する。
func (r *vaccinationRepository) FindOwnersByVaccineDeadline(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error) {
	target := targetDate.In(config.JST).Format(time.DateOnly)
	type row struct{ OwnerID uint64 }
	var rows []row
	err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Vaccination{}).
		Joins("JOIN pets ON pets.id = vaccinations.pet_id AND pets.clinic_id = vaccinations.clinic_id AND pets.deleted_at IS NULL AND pets.deceased_at IS NULL").
		Joins("JOIN owners ON owners.id = pets.owner_id AND owners.clinic_id = vaccinations.clinic_id AND owners.deleted_at IS NULL").
		Where("vaccinations.clinic_id = ? AND vaccinations.deleted_at IS NULL", clinicID).
		Where("vaccinations.next_date::date = ?::date", target).
		Distinct("pets.owner_id AS owner_id").
		Scan(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "vaccination", fmt.Sprintf("clinic=%d vaccine_deadline=%s", clinicID, target))
	}
	ids := make([]uint64, len(rows))
	for i, r := range rows {
		ids[i] = r.OwnerID
	}
	return ids, nil
}
