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

	// search / animal_species_id / 列ソート(pet_name・owner_name) は pets / owners / inquiries への JOIN が必要。
	// inquiries.medical_record_id と owners/pets の FK は 1 レコードにつき高々1件のため LEFT JOIN による重複行は発生しない。
	needsPetJoin := filters.AnimalSpeciesID != nil || filters.Search != "" || filters.Sort == "pet_name"
	needsOwnerJoin := filters.Search != "" || filters.Sort == "owner_name"
	needsInquiryJoin := filters.Search != ""

	buildBase := func() *gorm.DB {
		// clinicScopeIn は "clinic_id" を無修飾で参照するため、pets/owners
		// （いずれも clinic_id 列を持つ）を LEFT JOIN すると曖昧になる。
		// search/animal_species_id フィルタで JOIN が入るケースがあるため、
		// ここでは常に medical_records.clinic_id を明示指定する。
		q := r.db.WithContext(ctx).
			Model(&model.MedicalRecord{}).
			Where("medical_records.clinic_id IN ?", clinicIDs).
			Scopes(medicalRecordDetailRelationsScope())
		if needsPetJoin {
			q = q.Joins("LEFT JOIN pets ON pets.id = medical_records.pet_id AND pets.clinic_id = medical_records.clinic_id AND pets.deleted_at IS NULL")
		}
		if needsOwnerJoin {
			q = q.Joins("LEFT JOIN owners ON owners.id = medical_records.owner_id AND owners.clinic_id = medical_records.clinic_id AND owners.deleted_at IS NULL")
		}
		if needsInquiryJoin {
			q = q.Joins("LEFT JOIN inquiries ON inquiries.medical_record_id = medical_records.id")
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
		if filters.Search != "" {
			q = applyMedicalRecordSearch(q, filters.Search)
		}
		return q
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "medical_record", "")
	}
	if err := buildBase().
		Scopes(paginate(page, limit)).Order(medicalRecordOrderClause(filters.Sort, filters.Order)).
		Preload("Owner", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("Pet", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("Pet.AnimalSpecies").
		Preload("Doctor", medicalRecordStaffPreload(clinicIDs, true)).
		Preload("EnteredByStaff", medicalRecordStaffPreload(clinicIDs, false)).
		Preload("Inquiry").
		Preload("Billing", medicalRecordBillingPreload(clinicIDs)).
		Find(&records).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "medical_record", "")
	}
	return records, total, nil
}
