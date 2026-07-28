package medicalrecord

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestExamReferenceRangeService_ReplaceAtomicClinicSafeAndNoHistoryRewrite(t *testing.T) {
	db := setupExamTypeTestDB(t)
	repo := NewExamTypeRepository(db)
	service := NewExamTypeService(repo, testTransactor{db: db})
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	examType := makeExamTypeMaster(t, db, clinicA, "reference range owner")
	field := &model.ExamTypeField{ClinicID: clinicA, ExamTypeID: examType.ID, Name: "WBC"}
	require.NoError(t, db.Create(field).Error)
	species := &model.AnimalSpecies{Name: "U4 species", IsActive: true}
	require.NoError(t, db.Create(species).Error)
	exam := makeExaminationRec(t, db, &model.Examination{
		ClinicID: clinicA, ExamTypeID: examType.ID, Date: time.Now(),
	})
	oldMin, oldMax := 1.0, 2.0
	result := &model.ExamResult{
		ExamID: exam.ID, ExamTypeItemID: &field.ID, Name: field.Name,
		InspectionValue: "9", RefMin: &oldMin, RefMax: &oldMax, IsAbnormal: true,
		Status: model.ExaminationResultStatusHigh,
	}
	require.NoError(t, db.Create(result).Error)

	newMin, newMax := 3.0, 4.0
	replacement := func(fieldID uint64, ranges []ReferenceRangeInput) *ReplaceReferenceRangesCommand {
		return &ReplaceReferenceRangesCommand{ExamTypeFieldID: fieldID, Ranges: ranges}
	}
	replaced, err := service.ReplaceReferenceRanges(ctx, clinicA, examType.ID, replacement(field.ID, []ReferenceRangeInput{{
		AnimalSpeciesID: species.ID, RefMin: &newMin, RefMax: &newMax,
	}}))
	require.NoError(t, err)
	require.Len(t, replaced.ReferenceRanges, 1)
	assert.Equal(t, field.Name, replaced.Field.Name)
	assert.Equal(t, species.ID, replaced.ReferenceRanges[0].AnimalSpeciesID)
	assert.NotZero(t, replaced.ReferenceRanges[0].ID)

	unit := "10^3/μL"
	updated, err := service.UpdateField(
		ctx,
		clinicA,
		examType.ID,
		field.ID,
		&UpdateExamTypeFieldInput{Unit: &unit},
	)
	require.NoError(t, err)
	assert.Equal(t, field.Name, updated.Field.Name)
	assert.Equal(t, unit, updated.Field.Unit)
	assert.NotZero(t, updated.Field.CreatedAt)
	require.Len(t, updated.ReferenceRanges, 1)
	assert.Equal(t, species.ID, updated.ReferenceRanges[0].AnimalSpeciesID)

	var historical model.ExamResult
	require.NoError(t, db.First(&historical, result.ID).Error)
	assert.Equal(t, oldMin, *historical.RefMin)
	assert.Equal(t, oldMax, *historical.RefMax)
	assert.True(t, historical.IsAbnormal)
	assert.Equal(t, model.ExaminationResultStatusHigh, historical.Status)

	_, err = service.ReplaceReferenceRanges(ctx, clinicB, examType.ID, replacement(field.ID, []ReferenceRangeInput{}))
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	var count int64
	require.NoError(t, db.Model(&model.ExamReferenceRange{}).Where("exam_type_field_id = ?", field.ID).Count(&count).Error)
	assert.EqualValues(t, 1, count, "cross-clinic failure must not clear ranges")

	_, err = service.ReplaceReferenceRanges(ctx, clinicA, examType.ID, replacement(field.ID, []ReferenceRangeInput{{
		AnimalSpeciesID: ^uint64(0), RefMin: &newMin,
	}}))
	require.Error(t, err)
	require.NoError(t, db.Model(&model.ExamReferenceRange{}).Where("exam_type_field_id = ?", field.ID).Count(&count).Error)
	assert.EqualValues(t, 1, count, "invalid species must leave existing ranges intact")

	tooLarge := 1e100
	_, err = service.ReplaceReferenceRanges(ctx, clinicA, examType.ID, replacement(field.ID, []ReferenceRangeInput{{
		AnimalSpeciesID: species.ID, RefMin: &tooLarge,
	}}))
	require.Error(t, err)
	require.NoError(t, db.Model(&model.ExamReferenceRange{}).Where("exam_type_field_id = ?", field.ID).Count(&count).Error)
	assert.EqualValues(t, 1, count, "insert failure after delete must roll back the full replacement")

	cleared, err := service.ReplaceReferenceRanges(
		ctx,
		clinicA,
		examType.ID,
		replacement(field.ID, []ReferenceRangeInput{}),
	)
	require.NoError(t, err)
	assert.Empty(t, cleared.ReferenceRanges)
	assert.Equal(t, field.Name, cleared.Field.Name)
	require.NoError(t, db.Model(&model.ExamReferenceRange{}).Where("exam_type_field_id = ?", field.ID).Count(&count).Error)
	assert.Zero(t, count)
}

func TestExamTypeService_DeleteField_RejectsReferencesWithoutWrites(t *testing.T) {
	db := setupExamTypeTestDB(t)
	repo := NewExamTypeRepository(db)
	service := NewExamTypeService(repo, testTransactor{db: db})
	ctx := context.Background()
	const clinicID, foreignClinicID = uint64(1), uint64(2)

	examType := makeExamTypeMaster(t, db, clinicID, "delete guard")
	field := &model.ExamTypeField{ClinicID: clinicID, ExamTypeID: examType.ID, Name: "guarded"}
	require.NoError(t, db.Create(field).Error)
	species := &model.AnimalSpecies{Name: "U4 delete species", IsActive: true}
	require.NoError(t, db.Create(species).Error)
	require.NoError(t, db.Create(&model.ExamReferenceRange{
		ClinicID: clinicID, ExamTypeFieldID: field.ID, AnimalSpeciesID: species.ID,
	}).Error)

	err := service.DeleteField(ctx, foreignClinicID, examType.ID, field.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	var fieldCount, rangeCount int64
	require.NoError(t, db.Model(&model.ExamTypeField{}).Where("id = ?", field.ID).Count(&fieldCount).Error)
	require.NoError(t, db.Model(&model.ExamReferenceRange{}).Where("exam_type_field_id = ?", field.ID).Count(&rangeCount).Error)
	assert.EqualValues(t, 1, fieldCount)
	assert.EqualValues(t, 1, rangeCount)

	err = service.DeleteField(ctx, clinicID, examType.ID, field.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	require.NoError(t, db.Model(&model.ExamTypeField{}).Where("id = ?", field.ID).Count(&fieldCount).Error)
	assert.EqualValues(t, 1, fieldCount)

	require.NoError(t, db.Where("exam_type_field_id = ?", field.ID).Delete(&model.ExamReferenceRange{}).Error)
	exam := makeExaminationRec(t, db, &model.Examination{
		ClinicID: clinicID, ExamTypeID: examType.ID, Date: time.Now(),
	})
	require.NoError(t, db.Create(&model.ExamResult{
		ExamID: exam.ID, ExamTypeItemID: &field.ID, Name: field.Name,
	}).Error)
	err = service.DeleteField(ctx, clinicID, examType.ID, field.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	require.NoError(t, db.Model(&model.ExamTypeField{}).Where("id = ?", field.ID).Count(&fieldCount).Error)
	assert.EqualValues(t, 1, fieldCount)
}

func TestExamTypeService_FieldCRUDAndReorder_ClinicIsolation(t *testing.T) {
	db := setupExamTypeTestDB(t)
	repo := NewExamTypeRepository(db)
	service := NewExamTypeService(repo, testTransactor{db: db})
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	examType := makeExamTypeMaster(t, db, clinicA, "field CRUD")

	first, err := service.CreateField(ctx, clinicA, &CreateExamTypeFieldCommand{
		ExamTypeID: examType.ID,
		Field:      CreateExamTypeFieldInput{Name: "first", SortOrder: 10},
	})
	require.NoError(t, err)
	_, err = service.CreateField(ctx, clinicB, &CreateExamTypeFieldCommand{
		ExamTypeID: examType.ID,
		Field:      CreateExamTypeFieldInput{Name: "foreign"},
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))

	renamed := "renamed"
	_, err = service.UpdateField(ctx, clinicB, examType.ID, first.ID, &UpdateExamTypeFieldInput{Name: &renamed})
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	var stored model.ExamTypeField
	require.NoError(t, db.First(&stored, first.ID).Error)
	assert.Equal(t, "first", stored.Name)

	second, err := service.CreateField(ctx, clinicA, &CreateExamTypeFieldCommand{
		ExamTypeID: examType.ID,
		Field:      CreateExamTypeFieldInput{Name: "second", SortOrder: 20},
	})
	require.NoError(t, err)
	err = service.ReorderFields(ctx, clinicB, examType.ID, []uint64{second.ID, first.ID})
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	var firstBefore, secondBefore model.ExamTypeField
	require.NoError(t, db.First(&firstBefore, first.ID).Error)
	require.NoError(t, db.First(&secondBefore, second.ID).Error)
	assert.Equal(t, 10, firstBefore.SortOrder)
	assert.Equal(t, 20, secondBefore.SortOrder)

	require.NoError(t, service.ReorderFields(ctx, clinicA, examType.ID, []uint64{second.ID, first.ID}))
	require.NoError(t, db.First(&stored, first.ID).Error)
	assert.Equal(t, 2, stored.SortOrder)

	require.NoError(t, service.DeleteField(ctx, clinicA, examType.ID, second.ID))
	var count int64
	require.NoError(t, db.Model(&model.ExamTypeField{}).Where("id = ?", second.ID).Count(&count).Error)
	assert.Zero(t, count)
}
