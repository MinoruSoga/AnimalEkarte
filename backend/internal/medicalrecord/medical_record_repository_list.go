package medicalrecord

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (r *medicalRecordRepository) FindAll(ctx context.Context, clinicIDs []uint64, filters MedicalRecordListFilters, page, limit int) ([]model.MedicalRecord, int64, error) {
	records := make([]model.MedicalRecord, 0)
	var total int64

	// フェイルセーフ: 検証バグ等で空スライスが渡っても全件露出させない
	if len(clinicIDs) == 0 {
		return []model.MedicalRecord{}, 0, nil
	}

	// animal_species_id / 列ソート(pet_name・owner_name) は pets / owners への JOIN が必要。
	// search はペア IN + UNION のID集合駆動（applyMedicalRecordSearch）で組み立てるため
	// JOIN 不要 — JOIN+OR は複数テーブルにまたがる OR でインデックスを殺し全行逐行評価化する。
	needsPetJoin := filters.AnimalSpeciesID != nil || filters.Sort == "pet_name"
	needsOwnerJoin := filters.Sort == "owner_name"

	buildBase := func(withIsolation bool) *gorm.DB {
		// clinicScopeIn は "clinic_id" を無修飾で参照するため、pets/owners
		// （いずれも clinic_id 列を持つ）を LEFT JOIN すると曖昧になる。
		// search/animal_species_id フィルタで JOIN が入るケースがあるため、
		// ここでは常に medical_records.clinic_id を明示指定する。
		q := r.db.WithContext(ctx).
			Model(&model.MedicalRecord{}).
			Where("medical_records.clinic_id IN ?", clinicIDs)
		if withIsolation {
			q = q.Scopes(medicalRecordDetailRelationsScope())
		}
		if needsPetJoin {
			q = q.Joins("LEFT JOIN pets ON pets.id = medical_records.pet_id AND pets.clinic_id = medical_records.clinic_id AND pets.deleted_at IS NULL")
		}
		if needsOwnerJoin {
			q = q.Joins("LEFT JOIN owners ON owners.id = medical_records.owner_id AND owners.clinic_id = medical_records.clinic_id AND owners.deleted_at IS NULL")
		}
		if filters.PetID != nil {
			q = q.Where("medical_records.pet_id = ?", *filters.PetID)
		}
		if filters.OwnerID != nil {
			q = q.Where(`
				EXISTS (
					SELECT 1
					FROM pets current_owner_pet
					JOIN owners current_owner
					  ON current_owner.id = current_owner_pet.owner_id
					 AND current_owner.clinic_id = current_owner_pet.clinic_id
					WHERE current_owner_pet.id = medical_records.pet_id
					  AND current_owner_pet.clinic_id = medical_records.clinic_id
					  AND current_owner.id = ?
				)
			`, *filters.OwnerID)
		}
		if filters.StartDate != nil {
			q = q.Where("medical_records.date >= ?", *filters.StartDate)
		}
		if filters.EndDate != nil {
			q = q.Where("medical_records.date <= ?", *filters.EndDate)
		}
		if filters.Status != nil {
			q = q.Where("medical_records.status = ?", *filters.Status)
		}
		if filters.DoctorID != nil {
			q = q.Where("medical_records.doctor_id = ?", *filters.DoctorID)
		}
		if filters.AnimalSpeciesID != nil {
			q = q.Where("pets.animal_species_id = ?", *filters.AnimalSpeciesID)
		}
		if filters.MedicineID != nil {
			q = applyTreatmentItemIDFilter(q, "medicine_id", *filters.MedicineID)
		}
		if filters.ProcedureID != nil {
			q = applyTreatmentItemIDFilter(q, "procedure_id", *filters.ProcedureID)
		}
		if filters.ConsultationID != nil {
			q = applyTreatmentItemIDFilter(q, "consultation_id", *filters.ConsultationID)
		}
		if filters.InventoryID != nil {
			q = applyTreatmentItemIDFilter(q, "inventory_id", *filters.InventoryID)
		}
		if filters.Search != "" {
			q = applyMedicalRecordSearch(q, filters.Search)
		}
		// EMR-245: 表示列（飼主名・ペット名・主訴）の独立部分一致条件は
		// Search の横断 OR とは別に AND で積む。
		if filters.OwnerName != "" {
			q = applyMedicalRecordOwnerNameFilter(q, filters.OwnerName)
		}
		if filters.PetName != "" {
			q = applyMedicalRecordPetNameFilter(q, filters.PetName)
		}
		if filters.ChiefComplaint != "" {
			q = applyMedicalRecordChiefComplaintFilter(q, filters.ChiefComplaint)
		}
		return q
	}

	if err := buildBase(false).Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "medical_record", "")
	}
	if err := buildBase(true).
		Scopes(paginate(page, limit)).Order(medicalRecordOrderClause(filters.Sort, filters.Order)).
		Preload("Owner", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("Pet", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("Pet.AnimalSpecies").
		Preload("Doctor", medicalRecordStaffPreload(clinicIDs, true)).
		Preload("EnteredByStaff", medicalRecordEnteredByStaffPreload()).
		Preload("Inquiry").
		Preload("Billing", medicalRecordBillingPreload(clinicIDs)).
		Find(&records).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "medical_record", "")
	}
	return records, total, nil
}

// applyTreatmentItemIDFilter は「treatments のマスタ列（medicine_id/procedure_id/
// consultation_id/inventory_id）が指定IDのカルテ」を返す AND フィルタ。
// treatments(<列>) の btree インデックスで一致集合を先に計算し、id の半結合で引く。
// treatments には clinic_id 列がないため相関は medical_record_id のみ
// （clinic スコープは外側の medical_records.clinic_id IN が担保）。
// column は呼出側が固定リテラルで渡す内部限定値（ユーザー入力は渡さない）。
func applyTreatmentItemIDFilter(q *gorm.DB, column string, id uint64) *gorm.DB {
	return q.Where(
		`medical_records.id IN (`+
			`SELECT filtered_treatment.medical_record_id`+
			` FROM treatments filtered_treatment`+
			` WHERE filtered_treatment.`+column+` = ?`+
			` AND filtered_treatment.deleted_at IS NULL)`,
		id,
	)
}
