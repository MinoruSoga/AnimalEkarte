package medicalrecord

import (
	"context"
	"fmt"

	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// ExamReferenceRangeResolver is the examination result writer's narrow read view.
// The concrete examination repository implements it without widening the legacy
// ExaminationRepository interface or changing composition wiring.
type ExamReferenceRangeResolver interface {
	FindAnimalSpeciesID(ctx context.Context, clinicID, examID uint64) (uint64, error)
	ResolveByFieldIDs(
		ctx context.Context,
		clinicID, animalSpeciesID uint64,
		fieldIDs []uint64,
	) (map[uint64]model.ExamReferenceRange, error)
}

// FindAnimalSpeciesID resolves the species from the locked examination's patient.
// The clinic correlation is explicit and intentionally has no pet lifecycle
// predicate so historical deleted/deceased patients remain valid evidence.
func (r *examinationRepository) FindAnimalSpeciesID(
	ctx context.Context,
	clinicID, examID uint64,
) (uint64, error) {
	var row struct {
		AnimalSpeciesID uint64
	}
	db := persistence.DBOrTx(ctx, r.db).
		Table("exams").
		Select("pets.animal_species_id").
		Joins("JOIN pets ON pets.id = exams.pet_id AND pets.clinic_id = exams.clinic_id").
		Where("exams.id = ? AND exams.clinic_id = ? AND exams.deleted_at IS NULL", examID, clinicID)
	if persistence.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	if err := db.Take(&row).Error; err != nil {
		return 0, apperrors.FromGORM(err, "exam", fmt.Sprintf("%d", examID))
	}
	return row.AnimalSpeciesID, nil
}

// ResolveByFieldIDs loads all ranges for one clinic and species in one query.
func (r *examinationRepository) ResolveByFieldIDs(
	ctx context.Context,
	clinicID, animalSpeciesID uint64,
	fieldIDs []uint64,
) (map[uint64]model.ExamReferenceRange, error) {
	resolved := make(map[uint64]model.ExamReferenceRange, len(fieldIDs))
	if len(fieldIDs) == 0 {
		return resolved, nil
	}

	rows := make([]model.ExamReferenceRange, 0, len(fieldIDs))
	db := persistence.DBOrTx(ctx, r.db).
		Table("exam_reference_ranges").
		Select("id, clinic_id, exam_type_field_id, animal_species_id, ref_min, ref_max, qualitative_min, qualitative_max").
		Where(
			"clinic_id = ? AND animal_species_id = ? AND exam_type_field_id IN ?",
			clinicID,
			animalSpeciesID,
			fieldIDs,
		)
	if persistence.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	if err := db.Find(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "exam_reference_range", "")
	}
	for i := range rows {
		resolved[rows[i].ExamTypeFieldID] = rows[i]
	}
	return resolved, nil
}
