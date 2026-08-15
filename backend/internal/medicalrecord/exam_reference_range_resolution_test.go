package medicalrecord

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type referenceRangeResolverExaminationRepository struct {
	*mockExaminationRepository
	speciesID       uint64
	rangesBySpecies map[uint64]map[uint64]model.ExamReferenceRange
	speciesCalls    int
	resolveCalls    int
	resolvedFields  []uint64
	savedItems      []model.ExamResult
}

func (r *referenceRangeResolverExaminationRepository) FindAnimalSpeciesID(
	_ context.Context,
	_, _ uint64,
) (uint64, error) {
	r.speciesCalls++
	return r.speciesID, nil
}

func (r *referenceRangeResolverExaminationRepository) ResolveByFieldIDs(
	_ context.Context,
	_, animalSpeciesID uint64,
	fieldIDs []uint64,
) (map[uint64]model.ExamReferenceRange, error) {
	r.resolveCalls++
	r.resolvedFields = append([]uint64(nil), fieldIDs...)
	return r.rangesBySpecies[animalSpeciesID], nil
}

func newReferenceRangeResolverService(
	petID uint64,
	fieldIDs []uint64,
	repo *referenceRangeResolverExaminationRepository,
) ExaminationService {
	examTypeItems := make([]model.ExamTypeField, 0, len(fieldIDs))
	for _, fieldID := range fieldIDs {
		examTypeItems = append(examTypeItems, model.ExamTypeField{ID: fieldID})
	}
	examTypeRepo := &mockExamTypeRepository{
		findByIDFn: func(_ context.Context, _, examTypeID uint64) (*model.ExaminationType, error) {
			return &model.ExaminationType{ID: examTypeID, Items: examTypeItems}, nil
		},
	}
	repo.mockExaminationRepository = &mockExaminationRepository{
		lockByIDForUpdateFn: func(_ context.Context, _, examID uint64) (*model.Examination, error) {
			return &model.Examination{
				ID:         examID,
				ClinicID:   1,
				PetID:      &petID,
				ExamTypeID: 99,
				Status:     model.ExaminationStatusPending,
			}, nil
		},
		replaceItemsByExamIDFn: func(
			_ context.Context,
			_, _ uint64,
			items []model.ExamResult,
		) ([]model.ExamResult, int64, error) {
			repo.savedItems = append([]model.ExamResult(nil), items...)
			return items, 0, nil
		},
	}
	return NewExaminationService(
		repo,
		&mockMedicalRecordRepository{},
		examTypeRepo,
		nil,
		&mockCheckupTransactor{},
	)
}

func TestExaminationService_ReplaceItemsUsesMasterRanges(t *testing.T) {
	const (
		petID     = uint64(70)
		speciesID = uint64(7)
		fieldA    = uint64(11)
		fieldB    = uint64(12)
	)
	masterMinA, masterMaxA := 1.0, 10.0
	masterMinB, masterMaxB := 20.0, 30.0
	repo := &referenceRangeResolverExaminationRepository{
		speciesID: speciesID,
		rangesBySpecies: map[uint64]map[uint64]model.ExamReferenceRange{
			speciesID: {
				fieldA: {ExamTypeFieldID: fieldA, RefMin: &masterMinA, RefMax: &masterMaxA},
				fieldB: {ExamTypeFieldID: fieldB, RefMin: &masterMinB, RefMax: &masterMaxB},
			},
		},
	}
	svc := newReferenceRangeResolverService(petID, []uint64{fieldA, fieldB}, repo)

	got, err := svc.ReplaceItems(context.Background(), 1, 50, nil, []UpsertExamItemInput{
		{
			ExamTypeFieldID: &[]uint64{fieldA}[0],
			InspectionValue: "5",
		},
		{
			ExamTypeFieldID: &[]uint64{fieldB}[0],
			InspectionValue: "35",
		},
		{
			ExamTypeFieldID: &[]uint64{fieldA}[0],
			InspectionValue: "0",
		},
	})

	require.NoError(t, err)
	require.Len(t, got, 3)
	assert.Equal(t, 1, repo.speciesCalls)
	assert.Equal(t, 1, repo.resolveCalls, "reference ranges must be resolved once outside the item loop")
	assert.ElementsMatch(t, []uint64{fieldA, fieldB}, repo.resolvedFields, "field IDs must be deduplicated")
	require.NotNil(t, got[0].RefMin)
	require.NotNil(t, got[0].RefMax)
	assert.Equal(t, masterMinA, *got[0].RefMin)
	assert.Equal(t, masterMaxA, *got[0].RefMax)
	assert.Equal(t, model.ExaminationResultStatusNormal, got[0].Status)
	assert.False(t, got[0].IsAbnormal)
	assert.Equal(t, model.ExaminationResultStatusHigh, got[1].Status)
	assert.True(t, got[1].IsAbnormal)
	assert.Equal(t, model.ExaminationResultStatusLow, got[2].Status)
	assert.True(t, got[2].IsAbnormal)
}

func TestExaminationService_ReplaceItemsUsesQualitativeMasterRangesOnce(t *testing.T) {
	const (
		petID     = uint64(70)
		speciesID = uint64(7)
		fieldID   = uint64(11)
	)
	qualitativeMin := "(-)"
	qualitativeMax := "(+)"
	repo := &referenceRangeResolverExaminationRepository{
		speciesID: speciesID,
		rangesBySpecies: map[uint64]map[uint64]model.ExamReferenceRange{
			speciesID: {
				fieldID: {
					ExamTypeFieldID: fieldID,
					QualitativeMin:  &qualitativeMin,
					QualitativeMax:  &qualitativeMax,
				},
			},
		},
	}
	svc := newReferenceRangeResolverService(petID, []uint64{fieldID}, repo)

	got, err := svc.ReplaceItems(context.Background(), 1, 50, nil, []UpsertExamItemInput{
		{ExamTypeFieldID: &[]uint64{fieldID}[0], InspectionValue: "(++)"},
	})

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, 1, repo.speciesCalls)
	assert.Equal(t, 1, repo.resolveCalls, "reference ranges must be resolved once outside the item loop")
	require.NotNil(t, got[0].QualitativeMin)
	require.NotNil(t, got[0].QualitativeMax)
	assert.Equal(t, qualitativeMin, *got[0].QualitativeMin)
	assert.Equal(t, qualitativeMax, *got[0].QualitativeMax)
	assert.Equal(t, model.ExaminationResultStatusHigh, got[0].Status)
	assert.True(t, got[0].IsAbnormal)
}

func TestExaminationService_NonnumericInputWithNumericRangeRemainsUnassessed(t *testing.T) {
	const (
		petID     = uint64(70)
		speciesID = uint64(7)
		fieldID   = uint64(11)
	)
	masterMin, masterMax := 1.0, 10.0
	repo := &referenceRangeResolverExaminationRepository{
		speciesID: speciesID,
		rangesBySpecies: map[uint64]map[uint64]model.ExamReferenceRange{
			speciesID: {
				fieldID: {ExamTypeFieldID: fieldID, RefMin: &masterMin, RefMax: &masterMax},
			},
		},
	}
	svc := newReferenceRangeResolverService(petID, []uint64{fieldID}, repo)

	got, err := svc.ReplaceItems(context.Background(), 1, 50, nil, []UpsertExamItemInput{
		{ExamTypeFieldID: &[]uint64{fieldID}[0], InspectionValue: "陰性"},
	})

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, model.ExaminationResultStatusNormal, got[0].Status)
	assert.False(t, got[0].IsAbnormal)
	assert.False(t, toExamResultResponse(&got[0]).IsAssessed)
}

func TestExaminationService_CoexistingRangeFamiliesFailClosed(t *testing.T) {
	const (
		petID     = uint64(70)
		speciesID = uint64(7)
		fieldID   = uint64(11)
	)
	masterMin, masterMax := 1.0, 10.0
	qualitativeMin, qualitativeMax := "(-)", "(+)"
	repo := &referenceRangeResolverExaminationRepository{
		speciesID: speciesID,
		rangesBySpecies: map[uint64]map[uint64]model.ExamReferenceRange{
			speciesID: {
				fieldID: {
					ExamTypeFieldID: fieldID,
					RefMin:          &masterMin,
					RefMax:          &masterMax,
					QualitativeMin:  &qualitativeMin,
					QualitativeMax:  &qualitativeMax,
				},
			},
		},
	}
	svc := newReferenceRangeResolverService(petID, []uint64{fieldID}, repo)

	got, err := svc.ReplaceItems(context.Background(), 1, 50, nil, []UpsertExamItemInput{
		{ExamTypeFieldID: &[]uint64{fieldID}[0], InspectionValue: "5"},
	})

	require.NoError(t, err)
	require.Len(t, got, 1)
	require.NotNil(t, got[0].RefMin)
	require.NotNil(t, got[0].QualitativeMin)
	assert.Equal(t, model.ExaminationResultStatusNormal, got[0].Status)
	assert.False(t, got[0].IsAbnormal)
	assert.False(t, toExamResultResponse(&got[0]).IsAssessed)
}

func TestExaminationService_ReplaceItemsMissingRangeLeavesUnassessedSnapshot(t *testing.T) {
	const (
		petID     = uint64(70)
		speciesID = uint64(7)
		bounded   = uint64(11)
		missing   = uint64(12)
	)
	masterMin, masterMax := 1.0, 10.0
	repo := &referenceRangeResolverExaminationRepository{
		speciesID: speciesID,
		rangesBySpecies: map[uint64]map[uint64]model.ExamReferenceRange{
			speciesID: {
				bounded: {ExamTypeFieldID: bounded, RefMin: &masterMin, RefMax: &masterMax},
			},
		},
	}
	svc := newReferenceRangeResolverService(petID, []uint64{bounded, missing}, repo)

	got, err := svc.ReplaceItems(context.Background(), 1, 50, nil, []UpsertExamItemInput{
		{ExamTypeFieldID: &[]uint64{bounded}[0], InspectionValue: "5"},
		{ExamTypeFieldID: &[]uint64{missing}[0], InspectionValue: "5"},
	})

	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, model.ExaminationResultStatusNormal, got[0].Status)
	require.NotNil(t, got[0].RefMin, "bounded normal result must carry its assessed range")
	require.NotNil(t, got[0].RefMax, "bounded normal result must carry its assessed range")
	assert.Equal(t, model.ExaminationResultStatusNormal, got[1].Status)
	assert.Nil(t, got[1].RefMin, "missing master range must remain distinguishable from assessed normal")
	assert.Nil(t, got[1].RefMax, "missing master range must remain distinguishable from assessed normal")
	assert.False(t, got[1].IsAbnormal)
}

func TestExaminationService_ReplaceItemsUsesSingleMasterRangeForHighLowAndLeavesMissingUnassessed(t *testing.T) {
	const (
		petID        = uint64(70)
		speciesID    = uint64(7)
		boundedField = uint64(11)
		missingField = uint64(12)
	)
	masterMin, masterMax := 1.0, 10.0
	repo := &referenceRangeResolverExaminationRepository{
		speciesID: speciesID,
		rangesBySpecies: map[uint64]map[uint64]model.ExamReferenceRange{
			speciesID: {
				boundedField: {
					ExamTypeFieldID: boundedField,
					RefMin:          &masterMin,
					RefMax:          &masterMax,
				},
			},
		},
	}
	svc := newReferenceRangeResolverService(petID, []uint64{boundedField, missingField}, repo)

	got, err := svc.ReplaceItems(context.Background(), 1, 50, nil, []UpsertExamItemInput{
		{ExamTypeFieldID: &[]uint64{boundedField}[0], InspectionValue: "11"},
		{ExamTypeFieldID: &[]uint64{boundedField}[0], InspectionValue: "0"},
		{ExamTypeFieldID: &[]uint64{missingField}[0], InspectionValue: "5"},
	})

	require.NoError(t, err)
	require.Len(t, got, 3)
	assert.Equal(t, 1, repo.resolveCalls, "reference ranges must be resolved once outside the item loop")
	assert.ElementsMatch(t, []uint64{boundedField, missingField}, repo.resolvedFields)

	tests := []struct {
		name          string
		index         int
		wantStatus    model.ExaminationResultStatus
		wantAbnormal  bool
		wantAssessed  bool
		wantSnapshots bool
	}{
		{
			name:          "value above master maximum is high and assessed",
			index:         0,
			wantStatus:    model.ExaminationResultStatusHigh,
			wantAbnormal:  true,
			wantAssessed:  true,
			wantSnapshots: true,
		},
		{
			name:          "value below master minimum is low and assessed",
			index:         1,
			wantStatus:    model.ExaminationResultStatusLow,
			wantAbnormal:  true,
			wantAssessed:  true,
			wantSnapshots: true,
		},
		{
			name:         "field without master range is normal but unassessed",
			index:        2,
			wantStatus:   model.ExaminationResultStatusNormal,
			wantAssessed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := got[tt.index]
			response := toExamResultResponse(&item)
			assert.Equal(t, tt.wantStatus, item.Status)
			assert.Equal(t, tt.wantAbnormal, item.IsAbnormal)
			assert.Equal(t, tt.wantAssessed, response.IsAssessed)
			if tt.wantSnapshots {
				require.NotNil(t, item.RefMin)
				require.NotNil(t, item.RefMax)
				assert.Equal(t, masterMin, *item.RefMin)
				assert.Equal(t, masterMax, *item.RefMax)
				return
			}
			assert.Nil(t, item.RefMin)
			assert.Nil(t, item.RefMax)
		})
	}
}

func TestExaminationService_ReplaceItemsRequiresPetForMappedFields(t *testing.T) {
	const fieldID = uint64(11)
	replaced := false
	repo := &referenceRangeResolverExaminationRepository{
		speciesID:       7,
		rangesBySpecies: map[uint64]map[uint64]model.ExamReferenceRange{},
	}
	repo.mockExaminationRepository = &mockExaminationRepository{
		lockByIDForUpdateFn: func(_ context.Context, _, examID uint64) (*model.Examination, error) {
			return &model.Examination{
				ID:         examID,
				ClinicID:   1,
				ExamTypeID: 99,
				Status:     model.ExaminationStatusPending,
			}, nil
		},
		replaceItemsByExamIDFn: func(
			_ context.Context,
			_, _ uint64,
			items []model.ExamResult,
		) ([]model.ExamResult, int64, error) {
			replaced = true
			return items, 0, nil
		},
	}
	examTypeRepo := &mockExamTypeRepository{
		findByIDFn: func(_ context.Context, _, examTypeID uint64) (*model.ExaminationType, error) {
			return &model.ExaminationType{
				ID:    examTypeID,
				Items: []model.ExamTypeField{{ID: fieldID}},
			}, nil
		},
	}
	svc := NewExaminationService(
		repo,
		&mockMedicalRecordRepository{},
		examTypeRepo,
		nil,
		&mockCheckupTransactor{},
	)

	_, err := svc.ReplaceItems(context.Background(), 1, 50, nil, []UpsertExamItemInput{
		{ExamTypeFieldID: &[]uint64{fieldID}[0], InspectionValue: "5"},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ペット")
	assert.False(t, replaced)
	assert.Zero(t, repo.speciesCalls)
	assert.Zero(t, repo.resolveCalls)
}

func TestExaminationService_ReplaceItemsUsesSpeciesSpecificRanges(t *testing.T) {
	const (
		fieldID = uint64(11)
		dogID   = uint64(1)
		catID   = uint64(2)
	)
	dogMin, dogMax := 1.0, 10.0
	catMin, catMax := 20.0, 30.0
	ranges := map[uint64]map[uint64]model.ExamReferenceRange{
		dogID: {
			fieldID: {ExamTypeFieldID: fieldID, AnimalSpeciesID: dogID, RefMin: &dogMin, RefMax: &dogMax},
		},
		catID: {
			fieldID: {ExamTypeFieldID: fieldID, AnimalSpeciesID: catID, RefMin: &catMin, RefMax: &catMax},
		},
	}

	tests := []struct {
		name       string
		speciesID  uint64
		wantMin    float64
		wantMax    float64
		wantStatus model.ExaminationResultStatus
	}{
		{name: "dog range", speciesID: dogID, wantMin: dogMin, wantMax: dogMax, wantStatus: model.ExaminationResultStatusHigh},
		{name: "cat range", speciesID: catID, wantMin: catMin, wantMax: catMax, wantStatus: model.ExaminationResultStatusNormal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &referenceRangeResolverExaminationRepository{
				speciesID:       tt.speciesID,
				rangesBySpecies: ranges,
			}
			svc := newReferenceRangeResolverService(70, []uint64{fieldID}, repo)

			got, err := svc.ReplaceItems(context.Background(), 1, 50, nil, []UpsertExamItemInput{
				{ExamTypeFieldID: &[]uint64{fieldID}[0], InspectionValue: "25"},
			})

			require.NoError(t, err)
			require.Len(t, got, 1)
			require.NotNil(t, got[0].RefMin)
			require.NotNil(t, got[0].RefMax)
			assert.Equal(t, tt.wantMin, *got[0].RefMin)
			assert.Equal(t, tt.wantMax, *got[0].RefMax)
			assert.Equal(t, tt.wantStatus, got[0].Status)
		})
	}
}

func setupExamReferenceRangeResolutionDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupExaminationTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ExamReferenceRange{}))
	require.NoError(t, db.Exec("TRUNCATE TABLE exam_reference_ranges CASCADE").Error)
	return db
}

func TestExamReferenceRangeRepository_ResolveByFieldIDs_ClinicAndSpeciesIsolation(t *testing.T) {
	db := setupExamReferenceRangeResolutionDB(t)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	typeA := makeExamTypeMaster(t, db, clinicA, "clinic A range type")
	typeB := makeExamTypeMaster(t, db, clinicB, "clinic B range type")
	fieldA := &model.ExamTypeField{ExamTypeID: typeA.ID, ClinicID: clinicA, Name: "WBC A"}
	fieldB := &model.ExamTypeField{ExamTypeID: typeB.ID, ClinicID: clinicB, Name: "WBC B"}
	require.NoError(t, db.Create(fieldA).Error)
	require.NoError(t, db.Create(fieldB).Error)
	dog := &model.AnimalSpecies{Name: "reference range dog"}
	cat := &model.AnimalSpecies{Name: "reference range cat"}
	require.NoError(t, db.Create(dog).Error)
	require.NoError(t, db.Create(cat).Error)
	aDogMin, aDogMax := 1.0, 10.0
	aCatMin, aCatMax := 20.0, 30.0
	bDogMin, bDogMax := 100.0, 200.0
	aDogQualMin, aDogQualMax := "(-)", "(+)"
	aCatQualMin, aCatQualMax := "(±)", "(++)"
	bDogQualMin, bDogQualMax := "(++)", "(+++)"
	require.NoError(t, db.Create(&[]model.ExamReferenceRange{
		{
			ClinicID: clinicA, ExamTypeFieldID: fieldA.ID, AnimalSpeciesID: dog.ID,
			RefMin: &aDogMin, RefMax: &aDogMax, QualitativeMin: &aDogQualMin, QualitativeMax: &aDogQualMax,
		},
		{
			ClinicID: clinicA, ExamTypeFieldID: fieldA.ID, AnimalSpeciesID: cat.ID,
			RefMin: &aCatMin, RefMax: &aCatMax, QualitativeMin: &aCatQualMin, QualitativeMax: &aCatQualMax,
		},
		{
			ClinicID: clinicB, ExamTypeFieldID: fieldB.ID, AnimalSpeciesID: dog.ID,
			RefMin: &bDogMin, RefMax: &bDogMax, QualitativeMin: &bDogQualMin, QualitativeMax: &bDogQualMax,
		},
	}).Error)
	resolver, ok := NewExaminationRepository(db).(ExamReferenceRangeResolver)
	require.True(t, ok)

	dogRanges, err := resolver.ResolveByFieldIDs(
		ctx,
		clinicA,
		dog.ID,
		[]uint64{fieldA.ID, fieldB.ID},
	)
	require.NoError(t, err)
	require.Len(t, dogRanges, 1)
	require.Contains(t, dogRanges, fieldA.ID)
	assert.NotContains(t, dogRanges, fieldB.ID, "clinic B range must not cross into clinic A resolution")
	require.NotNil(t, dogRanges[fieldA.ID].RefMin)
	assert.Equal(t, aDogMin, *dogRanges[fieldA.ID].RefMin)
	require.NotNil(t, dogRanges[fieldA.ID].QualitativeMin)
	require.NotNil(t, dogRanges[fieldA.ID].QualitativeMax)
	assert.Equal(t, aDogQualMin, *dogRanges[fieldA.ID].QualitativeMin)
	assert.Equal(t, aDogQualMax, *dogRanges[fieldA.ID].QualitativeMax)

	catRanges, err := resolver.ResolveByFieldIDs(ctx, clinicA, cat.ID, []uint64{fieldA.ID})
	require.NoError(t, err)
	require.Len(t, catRanges, 1)
	require.NotNil(t, catRanges[fieldA.ID].RefMin)
	assert.Equal(t, aCatMin, *catRanges[fieldA.ID].RefMin)
	require.NotNil(t, catRanges[fieldA.ID].QualitativeMin)
	assert.Equal(t, aCatQualMin, *catRanges[fieldA.ID].QualitativeMin)
}

func TestExamReferenceRangeRepository_FindAnimalSpeciesIDPreservesHistoricalPetsAndClinicCorrelation(t *testing.T) {
	db := setupExamReferenceRangeResolutionDB(t)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	typeA := makeExamTypeMaster(t, db, clinicA, "species lookup type")
	ownerA := makeTestOwner(t, db, clinicA, "species lookup owner A")
	ownerB := makeTestOwner(t, db, clinicB, "species lookup owner B")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "species lookup pet A")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "species lookup pet B")
	exam := makeExaminationRec(t, db, &model.Examination{
		ClinicID:   clinicA,
		ExamTypeID: typeA.ID,
		PetID:      &petA.ID,
		Date:       time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC),
	})
	resolver, ok := NewExaminationRepository(db).(ExamReferenceRangeResolver)
	require.True(t, ok)

	speciesID, err := resolver.FindAnimalSpeciesID(ctx, clinicA, exam.ID)
	require.NoError(t, err)
	assert.Equal(t, petA.AnimalSpeciesID, speciesID)

	deceasedAt := time.Now()
	require.NoError(t, db.Model(petA).Update("deceased_at", deceasedAt).Error)
	require.NoError(t, db.Delete(petA).Error)
	speciesID, err = resolver.FindAnimalSpeciesID(ctx, clinicA, exam.ID)
	require.NoError(t, err)
	assert.Equal(t, petA.AnimalSpeciesID, speciesID, "historical deleted/deceased pets must remain resolvable")

	require.NoError(t, db.Model(&model.Examination{}).
		Where("id = ?", exam.ID).
		Update("pet_id", petB.ID).Error)
	_, err = resolver.FindAnimalSpeciesID(ctx, clinicA, exam.ID)
	require.Error(t, err, "cross-clinic pet correlation must fail closed")
}

func TestExaminationService_CreateUsesMasterRanges(t *testing.T) {
	db := setupExamReferenceRangeResolutionDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)
	examType := makeExamTypeMaster(t, db, clinicID, "create snapshot type")
	field := &model.ExamTypeField{ExamTypeID: examType.ID, ClinicID: clinicID, Name: "create snapshot field"}
	require.NoError(t, db.Create(field).Error)
	owner := makeTestOwner(t, db, clinicID, "create snapshot owner")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "create snapshot pet")
	masterMin, masterMax := 1.0, 10.0
	require.NoError(t, db.Create(&model.ExamReferenceRange{
		ClinicID: clinicID, ExamTypeFieldID: field.ID, AnimalSpeciesID: pet.AnimalSpeciesID,
		RefMin: &masterMin, RefMax: &masterMax,
	}).Error)
	repo := NewExaminationRepository(db)
	svc := NewExaminationService(
		repo,
		&mockMedicalRecordRepository{},
		NewExamTypeRepository(db),
		&mockAuditTxLogger{},
		testTransactor{db: db},
	)
	items := []UpsertExamItemInput{{
		ExamTypeFieldID: &field.ID,
		InspectionValue: "5",
	}}

	exam, err := svc.Create(ctx, clinicID, &CreateExaminationInput{
		PetID:      &pet.ID,
		ExamTypeID: examType.ID,
		Date:       time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC),
		Items:      &items,
		ActorID:    ptrUint64(1),
	})
	require.NoError(t, err)

	saved, err := repo.FindAllItemsByExamID(ctx, clinicID, exam.ID)
	require.NoError(t, err)
	require.Len(t, saved, 1)
	require.NotNil(t, saved[0].RefMin)
	require.NotNil(t, saved[0].RefMax)
	assert.Equal(t, masterMin, *saved[0].RefMin)
	assert.Equal(t, masterMax, *saved[0].RefMax)
	assert.Equal(t, model.ExaminationResultStatusNormal, saved[0].Status)
	assert.False(t, saved[0].IsAbnormal)
}

func TestExaminationService_UpdatePetBeforeFirstConfirmReassessesExistingItems(t *testing.T) {
	db := setupExamReferenceRangeResolutionDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)
	examType := makeExamTypeMaster(t, db, clinicID, "patient-change reassessment type")
	field := &model.ExamTypeField{ExamTypeID: examType.ID, ClinicID: clinicID, Name: "patient-change field"}
	require.NoError(t, db.Create(field).Error)
	owner := makeTestOwner(t, db, clinicID, "patient-change owner")
	oldSpecies := &model.AnimalSpecies{Name: "patient-change old species"}
	newSpecies := &model.AnimalSpecies{Name: "patient-change new species"}
	require.NoError(t, db.Create(oldSpecies).Error)
	require.NoError(t, db.Create(newSpecies).Error)
	oldPet := &model.Pet{ClinicID: clinicID, OwnerID: owner.ID, AnimalSpeciesID: oldSpecies.ID, Name: "old patient"}
	newPet := &model.Pet{ClinicID: clinicID, OwnerID: owner.ID, AnimalSpeciesID: newSpecies.ID, Name: "new patient"}
	require.NoError(t, db.Create(oldPet).Error)
	require.NoError(t, db.Create(newPet).Error)
	oldMin, oldMax := 1.0, 10.0
	newMin, newMax := 10.0, 20.0
	require.NoError(t, db.Create(&[]model.ExamReferenceRange{
		{ClinicID: clinicID, ExamTypeFieldID: field.ID, AnimalSpeciesID: oldSpecies.ID, RefMin: &oldMin, RefMax: &oldMax},
		{ClinicID: clinicID, ExamTypeFieldID: field.ID, AnimalSpeciesID: newSpecies.ID, RefMin: &newMin, RefMax: &newMax},
	}).Error)
	repo := NewExaminationRepository(db)
	relations := &mockMedicalRecordRepository{
		findPetOwnerInClinicFn: func(_ context.Context, gotClinicID, petID uint64) (uint64, error) {
			assert.Equal(t, clinicID, gotClinicID)
			if petID != oldPet.ID && petID != newPet.ID {
				return 0, apperrors.WrapNotFound("pet", "scoped")
			}
			return owner.ID, nil
		},
	}
	svc := NewExaminationService(
		repo,
		relations,
		NewExamTypeRepository(db),
		&mockAuditTxLogger{},
		testTransactor{db: db},
	)
	items := []UpsertExamItemInput{{ExamTypeFieldID: &field.ID, InspectionValue: "5"}}
	exam, err := svc.Create(ctx, clinicID, &CreateExaminationInput{
		PetID: ptrUint64(oldPet.ID), ExamTypeID: examType.ID,
		Date: time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC), Items: &items, ActorID: ptrUint64(1),
	})
	require.NoError(t, err)

	updated, err := svc.Update(ctx, clinicID, exam.ID, UpdateExaminationInput{
		PetID: ptrUint64(newPet.ID), ActorID: ptrUint64(1),
	})
	require.NoError(t, err)
	require.NotNil(t, updated.PetID)
	assert.Equal(t, newPet.ID, *updated.PetID)
	saved, err := repo.FindAllItemsByExamID(ctx, clinicID, exam.ID)
	require.NoError(t, err)
	require.Len(t, saved, 1)
	require.NotNil(t, saved[0].RefMin)
	require.NotNil(t, saved[0].RefMax)
	assert.Equal(t, newMin, *saved[0].RefMin)
	assert.Equal(t, newMax, *saved[0].RefMax)
	assert.Equal(t, model.ExaminationResultStatusLow, saved[0].Status)
	assert.True(t, saved[0].IsAbnormal)
}

func TestExaminationService_ReferenceRangeSnapshotDoesNotChangeWhenMasterChanges(t *testing.T) {
	db := setupExamReferenceRangeResolutionDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)
	examType := makeExamTypeMaster(t, db, clinicID, "snapshot type")
	field := &model.ExamTypeField{ExamTypeID: examType.ID, ClinicID: clinicID, Name: "snapshot field"}
	require.NoError(t, db.Create(field).Error)
	owner := makeTestOwner(t, db, clinicID, "snapshot owner")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "snapshot pet")
	exam := makeExaminationRec(t, db, &model.Examination{
		ClinicID:   clinicID,
		ExamTypeID: examType.ID,
		PetID:      &pet.ID,
		Date:       time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC),
		Status:     model.ExaminationStatusPending,
	})
	originalMin, originalMax := 1.0, 10.0
	const wantSnapshotMin, wantSnapshotMax = 1.0, 10.0
	referenceRange := &model.ExamReferenceRange{
		ClinicID: clinicID, ExamTypeFieldID: field.ID, AnimalSpeciesID: pet.AnimalSpeciesID,
		RefMin: &originalMin, RefMax: &originalMax,
	}
	require.NoError(t, db.Create(referenceRange).Error)
	repo := NewExaminationRepository(db)
	svc := NewExaminationService(
		repo,
		&mockMedicalRecordRepository{},
		NewExamTypeRepository(db),
		nil,
		testTransactor{db: db},
	)
	_, err := svc.ReplaceItems(ctx, clinicID, exam.ID, nil, []UpsertExamItemInput{{
		ExamTypeFieldID: &field.ID,
		InspectionValue: "5",
	}})
	require.NoError(t, err)

	updatedMin, updatedMax := 6.0, 7.0
	require.NoError(t, db.Model(referenceRange).Updates(map[string]any{
		"ref_min": updatedMin,
		"ref_max": updatedMax,
	}).Error)
	saved, err := repo.FindAllItemsByExamID(ctx, clinicID, exam.ID)
	require.NoError(t, err)
	require.Len(t, saved, 1)
	require.NotNil(t, saved[0].RefMin)
	require.NotNil(t, saved[0].RefMax)
	assert.Equal(t, wantSnapshotMin, *saved[0].RefMin)
	assert.Equal(t, wantSnapshotMax, *saved[0].RefMax)
	assert.Equal(t, model.ExaminationResultStatusNormal, saved[0].Status)
}

func TestExaminationService_QualitativeReferenceRangeSnapshotDoesNotChangeWhenMasterChanges(t *testing.T) {
	db := setupExamReferenceRangeResolutionDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)
	examType := makeExamTypeMaster(t, db, clinicID, "qualitative snapshot type")
	field := &model.ExamTypeField{ExamTypeID: examType.ID, ClinicID: clinicID, Name: "qualitative snapshot field"}
	require.NoError(t, db.Create(field).Error)
	owner := makeTestOwner(t, db, clinicID, "qualitative snapshot owner")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "qualitative snapshot pet")
	exam := makeExaminationRec(t, db, &model.Examination{
		ClinicID:   clinicID,
		ExamTypeID: examType.ID,
		PetID:      &pet.ID,
		Date:       time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC),
		Status:     model.ExaminationStatusPending,
	})
	originalMin, originalMax := "(-)", "(+)"
	const wantSnapshotMin, wantSnapshotMax = "(-)", "(+)"
	referenceRange := &model.ExamReferenceRange{
		ClinicID: clinicID, ExamTypeFieldID: field.ID, AnimalSpeciesID: pet.AnimalSpeciesID,
		QualitativeMin: &originalMin, QualitativeMax: &originalMax,
	}
	require.NoError(t, db.Create(referenceRange).Error)
	repo := NewExaminationRepository(db)
	svc := NewExaminationService(
		repo,
		&mockMedicalRecordRepository{},
		NewExamTypeRepository(db),
		nil,
		testTransactor{db: db},
	)

	_, err := svc.ReplaceItems(ctx, clinicID, exam.ID, nil, []UpsertExamItemInput{{
		ExamTypeFieldID: &field.ID,
		InspectionValue: "(++)",
	}})
	require.NoError(t, err)

	updatedMin, updatedMax := "(++)", "(+++)"
	require.NoError(t, db.Model(referenceRange).Updates(map[string]any{
		"qualitative_min": updatedMin,
		"qualitative_max": updatedMax,
	}).Error)
	saved, err := repo.FindAllItemsByExamID(ctx, clinicID, exam.ID)
	require.NoError(t, err)
	require.Len(t, saved, 1)
	require.NotNil(t, saved[0].QualitativeMin)
	require.NotNil(t, saved[0].QualitativeMax)
	assert.Equal(t, wantSnapshotMin, *saved[0].QualitativeMin)
	assert.Equal(t, wantSnapshotMax, *saved[0].QualitativeMax)
	assert.Equal(t, model.ExaminationResultStatusHigh, saved[0].Status)
	assert.True(t, saved[0].IsAbnormal)
}
