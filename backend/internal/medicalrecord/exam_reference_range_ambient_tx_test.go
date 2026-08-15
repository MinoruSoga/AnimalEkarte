package medicalrecord

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func TestExamReferenceRangeRepository_FindAnimalSpeciesID_HoldsExamShareLockUntilAmbientTransactionCommits(
	t *testing.T,
) {
	db := setupExamReferenceRangeResolutionDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	const clinicID = uint64(1)

	examType := makeExamTypeMaster(t, db, clinicID, "ambient species lock type")
	owner := makeTestOwner(t, db, clinicID, "ambient species lock owner")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "ambient species lock pet")
	exam := makeExaminationRec(t, db, &model.Examination{
		ClinicID:   clinicID,
		ExamTypeID: examType.ID,
		PetID:      &pet.ID,
		Date:       time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC),
	})
	resolver, ok := NewExaminationRepository(db).(ExamReferenceRangeResolver)
	require.True(t, ok)

	ambientTx := db.WithContext(ctx).Begin()
	require.NoError(t, ambientTx.Error)
	ambientCommitted := false
	defer func() {
		if !ambientCommitted {
			_ = ambientTx.Rollback().Error
		}
	}()

	speciesID, err := resolver.FindAnimalSpeciesID(
		persistence.WithTxValue(ctx, ambientTx),
		clinicID,
		exam.ID,
	)
	require.NoError(t, err)
	require.Equal(t, pet.AnimalSpeciesID, speciesID)

	competingTx := db.WithContext(ctx).Begin()
	require.NoError(t, competingTx.Error)
	require.NoError(t, competingTx.Exec("SET LOCAL lock_timeout = '200ms'").Error)
	err = competingTx.Exec(
		"UPDATE exams SET machine = ? WHERE id = ? AND clinic_id = ?",
		"blocked before ambient commit",
		exam.ID,
		clinicID,
	).Error
	require.ErrorContains(
		t,
		err,
		"lock timeout",
		"exclusive update to the exact exam row must block while the ambient transaction holds its SHARE lock",
	)
	require.NoError(t, competingTx.Rollback().Error)

	require.NoError(t, ambientTx.Commit().Error)
	ambientCommitted = true

	afterCommitTx := db.WithContext(ctx).Begin()
	require.NoError(t, afterCommitTx.Error)
	defer afterCommitTx.Rollback()
	require.NoError(t, afterCommitTx.Exec("SET LOCAL lock_timeout = '200ms'").Error)
	result := afterCommitTx.Exec(
		"UPDATE exams SET machine = ? WHERE id = ? AND clinic_id = ?",
		"succeeds after ambient commit",
		exam.ID,
		clinicID,
	)
	require.NoError(t, result.Error)
	require.EqualValues(t, 1, result.RowsAffected)
	require.NoError(t, afterCommitTx.Commit().Error)
}

func TestExamReferenceRangeRepository_ResolveByFieldIDs_HoldsReferenceRangeShareLockUntilAmbientTransactionCommits(
	t *testing.T,
) {
	db := setupExamReferenceRangeResolutionDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	const clinicID = uint64(1)

	examType := makeExamTypeMaster(t, db, clinicID, "ambient reference range lock type")
	field := &model.ExamTypeField{
		ExamTypeID: examType.ID,
		ClinicID:   clinicID,
		Name:       "ambient reference range lock field",
	}
	require.NoError(t, db.WithContext(ctx).Create(field).Error)
	species := &model.AnimalSpecies{Name: "ambient reference range lock species"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)
	refMin, refMax := 1.0, 10.0
	referenceRange := &model.ExamReferenceRange{
		ClinicID:        clinicID,
		ExamTypeFieldID: field.ID,
		AnimalSpeciesID: species.ID,
		RefMin:          &refMin,
		RefMax:          &refMax,
	}
	require.NoError(t, db.WithContext(ctx).Create(referenceRange).Error)
	resolver, ok := NewExaminationRepository(db).(ExamReferenceRangeResolver)
	require.True(t, ok)

	ambientTx := db.WithContext(ctx).Begin()
	require.NoError(t, ambientTx.Error)
	ambientCommitted := false
	defer func() {
		if !ambientCommitted {
			_ = ambientTx.Rollback().Error
		}
	}()

	resolved, err := resolver.ResolveByFieldIDs(
		persistence.WithTxValue(ctx, ambientTx),
		clinicID,
		species.ID,
		[]uint64{field.ID},
	)
	require.NoError(t, err)
	require.Contains(t, resolved, field.ID)

	competingTx := db.WithContext(ctx).Begin()
	require.NoError(t, competingTx.Error)
	require.NoError(t, competingTx.Exec("SET LOCAL lock_timeout = '200ms'").Error)
	err = competingTx.Exec(
		`UPDATE exam_reference_ranges
		 SET ref_min = ?
		 WHERE id = ? AND clinic_id = ? AND animal_species_id = ? AND exam_type_field_id = ?`,
		2.0,
		referenceRange.ID,
		clinicID,
		species.ID,
		field.ID,
	).Error
	require.ErrorContains(
		t,
		err,
		"lock timeout",
		"exclusive update to the exact reference-range row must block while the ambient transaction holds its SHARE lock",
	)
	require.NoError(t, competingTx.Rollback().Error)

	require.NoError(t, ambientTx.Commit().Error)
	ambientCommitted = true

	afterCommitTx := db.WithContext(ctx).Begin()
	require.NoError(t, afterCommitTx.Error)
	defer afterCommitTx.Rollback()
	require.NoError(t, afterCommitTx.Exec("SET LOCAL lock_timeout = '200ms'").Error)
	result := afterCommitTx.Exec(
		`UPDATE exam_reference_ranges
		 SET ref_min = ?
		 WHERE id = ? AND clinic_id = ? AND animal_species_id = ? AND exam_type_field_id = ?`,
		2.0,
		referenceRange.ID,
		clinicID,
		species.ID,
		field.ID,
	)
	require.NoError(t, result.Error)
	require.EqualValues(t, 1, result.RowsAffected)
	require.NoError(t, afterCommitTx.Commit().Error)
}
