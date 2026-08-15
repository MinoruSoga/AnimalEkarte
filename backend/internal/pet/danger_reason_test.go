package pet

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	ownerdomain "github.com/animal-ekarte/backend/internal/owner"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func stringPointerPointer(value *string) **string {
	return &value
}

// UpdateAndFind keeps the existing package-local mock compatible with the
// repository's typed atomic capability without expanding unrelated test files.
func (m *mockPetRepository) UpdateAndFind(
	ctx context.Context,
	clinicID, id uint64,
	update PetUpdate,
) (*model.Pet, error) {
	current, err := m.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		current = &model.Pet{}
	}

	effectiveLevel := current.DangerLevel
	if update.dangerLevel != nil {
		effectiveLevel = *update.dangerLevel
	}
	effectiveReason := current.DangerReason
	if update.dangerReason != nil {
		effectiveReason = *update.dangerReason
	}
	normalizedReason, err := normalizeDangerReason(effectiveLevel, effectiveReason)
	if err != nil {
		return nil, err
	}

	fields := make(map[string]any, len(update.fields))
	for key, value := range update.fields {
		fields[key] = value
	}
	if update.dangerReason != nil {
		fields["danger_reason"] = normalizedReason
	}
	if m.updateFn != nil {
		if err := m.updateFn(ctx, clinicID, id, fields); err != nil {
			return nil, err
		}
	}
	return m.FindByID(ctx, clinicID, id)
}

func TestDangerReason_DirectCreatePropagation(t *testing.T) {
	reason := "  噛みつき歴あり  "
	request := createPetRequest{
		OwnerID:         7,
		AnimalSpeciesID: 3,
		Name:            "ポチ",
		DangerLevel:     string(model.DangerLevelHigh),
		DangerReason:    &reason,
	}

	input := request.toServiceInput()
	petModel := buildPetModel(11, "", input)
	draft := CreatePetDraftFromModel(*petModel)
	persisted := draft.model(11, 7, "7-1")

	require.NotNil(t, input.DangerReason)
	assert.Equal(t, reason, *input.DangerReason)
	require.NotNil(t, persisted.DangerReason)
	assert.Equal(t, reason, *persisted.DangerReason)
}

func TestDangerReason_UpdateRequestPreservesOmittedNullAndValue(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantSet    bool
		wantReason *string
	}{
		{name: "omitted", body: `{}`},
		{name: "explicit null", body: `{"danger_reason":null}`, wantSet: true},
		{name: "value", body: `{"danger_reason":"  噛む  "}`, wantSet: true, wantReason: ptrString("  噛む  ")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var request updatePetRequest
			require.NoError(t, json.Unmarshal([]byte(tt.body), &request))

			input := request.toServiceInput()
			if !tt.wantSet {
				assert.Nil(t, input.DangerReason)
				return
			}
			require.NotNil(t, input.DangerReason)
			assert.Equal(t, tt.wantReason, *input.DangerReason)
		})
	}
}

func TestDangerReason_DirectCreateValidation(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		reason      *string
		wantErr     bool
		wantPersist *string
	}{
		{
			name:    "high without reason is rejected",
			level:   string(model.DangerLevelHigh),
			wantErr: true,
		},
		{
			name:    "high with whitespace-only reason is rejected",
			level:   string(model.DangerLevelHigh),
			reason:  ptrString(" \t\n "),
			wantErr: true,
		},
		{
			name:    "reason over 500 Unicode runes is rejected",
			level:   string(model.DangerLevelMedium),
			reason:  ptrString(strings.Repeat("犬", 501)),
			wantErr: true,
		},
		{
			name:        "medium without reason remains valid",
			level:       string(model.DangerLevelMedium),
			wantPersist: nil,
		},
		{
			name:        "high reason is trimmed before persistence",
			level:       string(model.DangerLevelHigh),
			reason:      ptrString("  保定時に噛む  "),
			wantPersist: ptrString("保定時に噛む"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var persisted *model.Pet
			repo := &mockPetRepository{
				createFn: func(_ context.Context, pet *model.Pet) error {
					copy := *pet
					persisted = &copy
					return nil
				},
			}
			svc := newPetSvc(
				repo,
				defaultOwnerRepo(),
				defaultInsuranceRepo(11),
				defaultMedicalRecordRepo(),
			)

			_, err := svc.Create(context.Background(), 11, &CreatePetInput{
				OwnerID:         7,
				AnimalSpeciesID: 3,
				Name:            "ポチ",
				DangerLevel:     tt.level,
				DangerReason:    tt.reason,
			})

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, persisted)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, persisted)
			assert.Equal(t, tt.wantPersist, persisted.DangerReason)
		})
	}
}

func TestDangerReason_UpdateUsesEffectiveState(t *testing.T) {
	tests := []struct {
		name          string
		initialLevel  model.DangerLevel
		initialReason *string
		input         *UpdatePetInput
		wantErr       bool
		wantLevel     model.DangerLevel
		wantReason    *string
	}{
		{
			name:         "raising level to high without saved reason is rejected",
			initialLevel: model.DangerLevelMedium,
			input: &UpdatePetInput{
				DangerLevel: ptrString(string(model.DangerLevelHigh)),
			},
			wantErr:   true,
			wantLevel: model.DangerLevelMedium,
		},
		{
			name:          "clearing reason while level remains high is rejected",
			initialLevel:  model.DangerLevelHigh,
			initialReason: ptrString("噛む"),
			input: &UpdatePetInput{
				DangerReason: stringPointerPointer(nil),
			},
			wantErr:    true,
			wantLevel:  model.DangerLevelHigh,
			wantReason: ptrString("噛む"),
		},
		{
			name:          "low can clear reason",
			initialLevel:  model.DangerLevelLow,
			initialReason: ptrString("旧理由"),
			input: &UpdatePetInput{
				DangerReason: stringPointerPointer(nil),
			},
			wantLevel: model.DangerLevelLow,
		},
		{
			name:         "high accepts and trims a supplied reason",
			initialLevel: model.DangerLevelMedium,
			input: &UpdatePetInput{
				DangerLevel:  ptrString(string(model.DangerLevelHigh)),
				DangerReason: stringPointerPointer(ptrString("  噛む  ")),
			},
			wantLevel:  model.DangerLevelHigh,
			wantReason: ptrString("噛む"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := model.Pet{
				ID:           5,
				ClinicID:     11,
				OwnerID:      7,
				DangerLevel:  tt.initialLevel,
				DangerReason: tt.initialReason,
			}
			repo := &mockPetRepository{
				findByIDFn: func(context.Context, uint64, uint64) (*model.Pet, error) {
					copy := current
					return &copy, nil
				},
				updateFn: func(_ context.Context, _, _ uint64, fields map[string]any) error {
					if value, ok := fields["danger_level"]; ok {
						current.DangerLevel = model.DangerLevel(value.(string))
					}
					if value, ok := fields["danger_reason"]; ok {
						current.DangerReason, _ = value.(*string)
					}
					return nil
				},
			}
			svc := newPetSvc(
				repo,
				defaultOwnerRepo(),
				defaultInsuranceRepo(11),
				defaultMedicalRecordRepo(),
			)

			_, err := svc.Update(context.Background(), 11, 5, tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantLevel, current.DangerLevel)
			assert.Equal(t, tt.wantReason, current.DangerReason)
		})
	}
}

type firstReadBarrierRepository struct {
	ServiceRepository
	mu        sync.Mutex
	readCount int
	ready     chan struct{}
}

func (r *firstReadBarrierRepository) FindByID(
	ctx context.Context,
	clinicID, id uint64,
) (*model.Pet, error) {
	r.mu.Lock()
	r.readCount++
	readNumber := r.readCount
	if readNumber == 2 {
		close(r.ready)
	}
	r.mu.Unlock()
	if readNumber <= 2 {
		<-r.ready
	}
	return r.ServiceRepository.FindByID(ctx, clinicID, id)
}

func TestDangerReason_ConcurrentSetHighAndClearReasonNeverCommitsInvalidState(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(81)
	owner := makeTestOwner(t, db, clinicID, "並行更新飼主")
	speciesID := makeSyncSpeciesID(t, db)
	reason := "保定時注意"
	petModel := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         owner.ID,
		AnimalSpeciesID: speciesID,
		Name:            "並行更新ペット",
		DangerLevel:     model.DangerLevelMedium,
		DangerReason:    &reason,
	}
	require.NoError(t, db.WithContext(ctx).Create(petModel).Error)

	baseRepo := NewRepository(db)
	repo := &firstReadBarrierRepository{
		ServiceRepository: baseRepo,
		ready:             make(chan struct{}),
	}
	svc := NewService(
		repo,
		defaultOwnerRepo(),
		defaultInsuranceRepo(clinicID),
		defaultMedicalRecordRepo(),
		nil,
	)

	start := make(chan struct{})
	results := make(chan error, 2)
	go func() {
		<-start
		_, err := svc.Update(ctx, clinicID, petModel.ID, &UpdatePetInput{
			DangerLevel: ptrString(string(model.DangerLevelHigh)),
		})
		results <- err
	}()
	go func() {
		<-start
		_, err := svc.Update(ctx, clinicID, petModel.ID, &UpdatePetInput{
			DangerReason: stringPointerPointer(nil),
		})
		results <- err
	}()
	close(start)

	errs := []error{<-results, <-results}
	errorCount := 0
	for _, err := range errs {
		if err != nil {
			errorCount++
		}
	}
	assert.Equal(t, 1, errorCount, "exactly one conflicting patch must be rejected")

	var loaded model.Pet
	require.NoError(t, db.WithContext(ctx).First(&loaded, petModel.ID).Error)
	assert.False(t,
		loaded.DangerLevel == model.DangerLevelHigh &&
			(loaded.DangerReason == nil || strings.TrimSpace(*loaded.DangerReason) == ""),
		"high with an empty danger reason must never commit",
	)
}

func TestDangerReason_GenericUpdateCannotBypassInvariant(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(82)
	owner := makeTestOwner(t, db, clinicID, "generic update guard")
	validReason := "handles poorly"
	tests := []struct {
		name          string
		key           string
		value         any
		initialLevel  model.DangerLevel
		initialReason *string
	}{
		{
			name:         "column danger level",
			key:          "danger_level",
			value:        model.DangerLevelHigh,
			initialLevel: model.DangerLevelLow,
		},
		{
			name:         "Go field danger level",
			key:          "DangerLevel",
			value:        model.DangerLevelHigh,
			initialLevel: model.DangerLevelLow,
		},
		{
			name:          "column danger reason",
			key:           "danger_reason",
			value:         nil,
			initialLevel:  model.DangerLevelHigh,
			initialReason: &validReason,
		},
		{
			name:          "Go field danger reason",
			key:           "DangerReason",
			value:         nil,
			initialLevel:  model.DangerLevelHigh,
			initialReason: &validReason,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			petModel := makeSpeciesAndPet(t, db, clinicID, owner.ID, tt.name)
			require.NoError(t, db.Model(&model.Pet{}).
				Where("id = ?", petModel.ID).
				Updates(map[string]any{
					"danger_level":  tt.initialLevel,
					"danger_reason": tt.initialReason,
				}).Error)

			err := NewRepository(db).Update(ctx, clinicID, petModel.ID, map[string]any{
				tt.key: tt.value,
			})

			require.Error(t, err)
			var loaded model.Pet
			require.NoError(t, db.First(&loaded, petModel.ID).Error)
			assert.Equal(t, tt.initialLevel, loaded.DangerLevel)
			assert.Equal(t, tt.initialReason, loaded.DangerReason)
		})
	}
}

func TestDangerReason_GenericUpdateRejectsStructuralAliases(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicA, clinicB = uint64(85), uint64(86)
	ownerA := makeTestOwner(t, db, clinicA, "structural guard owner A")
	ownerB := makeTestOwner(t, db, clinicB, "structural guard owner B")
	insuranceA := &model.Insurance{ClinicID: clinicA, Name: "insurance A"}
	insuranceB := &model.Insurance{ClinicID: clinicB, Name: "insurance B"}
	require.NoError(t, db.Create(insuranceA).Error)
	require.NoError(t, db.Create(insuranceB).Error)
	petModel := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "structural guard pet")
	require.NoError(t, db.Model(&model.Pet{}).
		Where("id = ?", petModel.ID).
		Update("insurance_id", insuranceA.ID).Error)

	tests := []struct {
		key   string
		value any
	}{
		{key: "clinic_id", value: clinicB},
		{key: "ClinicID", value: clinicB},
		{key: "owner_id", value: ownerB.ID},
		{key: "OwnerID", value: ownerB.ID},
		{key: "insurance_id", value: insuranceB.ID},
		{key: "InsuranceID", value: insuranceB.ID},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			err := NewRepository(db).Update(ctx, clinicA, petModel.ID, map[string]any{
				tt.key: tt.value,
			})

			require.Error(t, err)
			var loaded model.Pet
			require.NoError(t, db.First(&loaded, petModel.ID).Error)
			assert.Equal(t, clinicA, loaded.ClinicID)
			assert.Equal(t, ownerA.ID, loaded.OwnerID)
			assert.Equal(t, &insuranceA.ID, loaded.InsuranceID)
		})
	}
}

func TestDangerReason_AtomicUpdateJoinsAmbientTransaction(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(83)
	owner := makeTestOwner(t, db, clinicID, "ambient transaction")
	reason := "保定注意"
	petModel := makeSpeciesAndPet(t, db, clinicID, owner.ID, "before rollback")
	require.NoError(t, db.Model(&model.Pet{}).
		Where("id = ?", petModel.ID).
		Updates(map[string]any{
			"danger_level":  model.DangerLevelMedium,
			"danger_reason": &reason,
		}).Error)

	rollbackErr := errors.New("force outer rollback")
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		level := model.DangerLevelHigh
		updated, updateErr := NewRepository(db).UpdateAndFind(
			persistence.WithTxValue(ctx, tx),
			clinicID,
			petModel.ID,
			PetUpdate{
				fields:      map[string]any{"name": "must roll back"},
				dangerLevel: &level,
			},
		)
		require.NoError(t, updateErr)
		assert.Equal(t, model.DangerLevelHigh, updated.DangerLevel)
		assert.Equal(t, "must roll back", updated.Name)
		return rollbackErr
	})
	require.ErrorIs(t, err, rollbackErr)

	var loaded model.Pet
	require.NoError(t, db.First(&loaded, petModel.ID).Error)
	assert.Equal(t, model.DangerLevelMedium, loaded.DangerLevel)
	assert.Equal(t, "before rollback", loaded.Name)
}

func TestDangerReason_AtomicUpdateRejectsCrossClinicTarget(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicA, clinicB = uint64(87), uint64(88)
	ownerA := makeTestOwner(t, db, clinicA, "atomic clinic owner A")
	petModel := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "clinic A pet")

	_, err := NewRepository(db).UpdateAndFind(
		ctx,
		clinicB,
		petModel.ID,
		PetUpdate{fields: map[string]any{"name": "cross-clinic mutation"}},
	)

	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "cross-clinic target must be indistinguishable from missing: %v", err)
	var loaded model.Pet
	require.NoError(t, db.First(&loaded, petModel.ID).Error)
	assert.Equal(t, clinicA, loaded.ClinicID)
	assert.Equal(t, "clinic A pet", loaded.Name)
}

func TestDangerReason_TypedUpdateFieldsCannotSmuggleDangerAliases(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(84)
	owner := makeTestOwner(t, db, clinicID, "typed update guard")
	validReason := "保定注意"
	tests := []struct {
		name          string
		key           string
		value         any
		initialLevel  model.DangerLevel
		initialReason *string
	}{
		{
			name:         "Go field danger level",
			key:          "DangerLevel",
			value:        model.DangerLevelHigh,
			initialLevel: model.DangerLevelLow,
		},
		{
			name:          "Go field danger reason",
			key:           "DangerReason",
			value:         nil,
			initialLevel:  model.DangerLevelHigh,
			initialReason: &validReason,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			petModel := makeSpeciesAndPet(t, db, clinicID, owner.ID, tt.name)
			require.NoError(t, db.Model(&model.Pet{}).
				Where("id = ?", petModel.ID).
				Updates(map[string]any{
					"danger_level":  tt.initialLevel,
					"danger_reason": tt.initialReason,
				}).Error)

			updated, err := NewRepository(db).UpdateAndFind(
				ctx,
				clinicID,
				petModel.ID,
				PetUpdate{fields: map[string]any{tt.key: tt.value}},
			)

			require.NoError(t, err)
			assert.Equal(t, tt.initialLevel, updated.DangerLevel)
			assert.Equal(t, tt.initialReason, updated.DangerReason)
		})
	}
}

func TestDangerReason_OwnerRegistrationWriterValidation(t *testing.T) {
	tests := []struct {
		name    string
		level   model.DangerLevel
		reason  *string
		wantErr bool
	}{
		{
			name:    "nested high without reason is rejected",
			level:   model.DangerLevelHigh,
			wantErr: true,
		},
		{
			name:    "nested reason over 500 runes is rejected",
			level:   model.DangerLevelMedium,
			reason:  ptrString(strings.Repeat("犬", 501)),
			wantErr: true,
		},
		{
			name:    "nested medium without reason remains valid",
			level:   model.DangerLevelMedium,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupPetRepositoryTestDB(t)
			ctx := context.Background()
			const clinicID = uint64(83)
			owner := makeTestOwner(t, db, clinicID, "nested create owner")
			speciesID := makeSyncSpeciesID(t, db)
			writer := NewOwnerRegistrationWriter()

			err := persistence.NewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
				_, err := writer.CreateForOwnerRegistration(txCtx, OwnerRegistrationIntent{
					ClinicID: clinicID,
					OwnerID:  owner.ID,
					Pets: []CreatePetDraft{{
						AnimalSpeciesID: speciesID,
						Name:            "nested pet",
						DangerLevel:     tt.level,
						DangerReason:    tt.reason,
					}},
				})
				return err
			})

			if tt.wantErr {
				require.Error(t, err)
				var appErr *apperrors.AppError
				require.True(t, errors.As(err, &appErr))
				assert.Equal(t, "INVALID_INPUT", appErr.Code)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestDangerReason_OwnerRepositoryAdapterWriterIntegration(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(89)
	species := &model.AnimalSpecies{Name: "owner danger integration species"}
	require.NoError(t, db.Create(species).Error)
	repo := ownerdomain.NewRepository(db, NewOwnerRegistrationAdapter(NewWriter(db)))

	t.Run("high reason is trimmed and committed through the pet write owner", func(t *testing.T) {
		reason := "  診察台で噛む  "
		ownerModel := &model.Owner{
			ClinicID: clinicID,
			Name:     "owner danger integration success",
		}

		err := repo.CreateWithPets(ctx, ownerModel, []model.Pet{{
			AnimalSpeciesID: species.ID,
			Name:            "nested danger pet",
			DangerLevel:     model.DangerLevelHigh,
			DangerReason:    &reason,
		}})

		require.NoError(t, err)
		require.NotZero(t, ownerModel.ID)
		var persisted model.Pet
		require.NoError(t, db.
			Where("clinic_id = ? AND owner_id = ? AND name = ?", clinicID, ownerModel.ID, "nested danger pet").
			First(&persisted).Error)
		assert.Equal(t, model.DangerLevelHigh, persisted.DangerLevel)
		require.NotNil(t, persisted.DangerReason)
		assert.Equal(t, "診察台で噛む", *persisted.DangerReason)
	})

	t.Run("high without reason rolls back owner and pet", func(t *testing.T) {
		const ownerName = "owner danger integration rollback"
		const petName = "invalid nested danger pet"
		ownerModel := &model.Owner{
			ClinicID: clinicID,
			Name:     ownerName,
		}

		err := repo.CreateWithPets(ctx, ownerModel, []model.Pet{{
			AnimalSpeciesID: species.ID,
			Name:            petName,
			DangerLevel:     model.DangerLevelHigh,
		}})

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "missing high reason must remain invalid input: %v", err)
		var ownerCount int64
		require.NoError(t, db.Model(&model.Owner{}).
			Where("clinic_id = ? AND name = ?", clinicID, ownerName).
			Count(&ownerCount).Error)
		assert.Zero(t, ownerCount)
		var petCount int64
		require.NoError(t, db.Model(&model.Pet{}).
			Where("clinic_id = ? AND name = ?", clinicID, petName).
			Count(&petCount).Error)
		assert.Zero(t, petCount)
	})
}

func TestDangerReason_StaffPetResponsesIncludeField(t *testing.T) {
	reason := "保定時注意"
	petModel := &model.Pet{
		ID:           9,
		DangerLevel:  model.DangerLevelHigh,
		DangerReason: &reason,
	}

	for name, response := range map[string]any{
		"detail": toPetResponse(petModel),
		"list":   toPetListResponse(petModel),
	} {
		t.Run(name, func(t *testing.T) {
			body, err := json.Marshal(response)
			require.NoError(t, err)
			var payload map[string]any
			require.NoError(t, json.Unmarshal(body, &payload))
			assert.Equal(t, reason, payload["danger_reason"])
		})
	}
}

func TestDangerReason_OwnerReportResponseOmitsField(t *testing.T) {
	reason := "staff only"
	body, err := json.Marshal(toOwnerReportPetResponse(&model.Pet{
		ID:           9,
		DangerLevel:  model.DangerLevelHigh,
		DangerReason: &reason,
	}))
	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(body, &payload))
	_, exposed := payload["danger_reason"]
	assert.False(t, exposed, "Owner Report must physically omit danger_reason: %s", body)
}
