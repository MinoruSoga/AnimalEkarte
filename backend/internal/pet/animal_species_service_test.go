package pet

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// ---- AnimalSpecies モック ----

type mockAnimalSpeciesRepository struct {
	findAllFn      func(ctx context.Context) ([]model.AnimalSpecies, error)
	findByIDFn     func(ctx context.Context, id uint64) (*model.AnimalSpecies, error)
	createFn       func(ctx context.Context, species *model.AnimalSpecies) error
	updateFieldsFn func(ctx context.Context, id uint64, fields map[string]any) (*model.AnimalSpecies, error)
	deleteFn       func(ctx context.Context, id uint64) error
	reorderFn      func(ctx context.Context, ids []uint64) error
}

func (m *mockAnimalSpeciesRepository) FindAll(ctx context.Context) ([]model.AnimalSpecies, error) {
	return m.findAllFn(ctx)
}

func (m *mockAnimalSpeciesRepository) FindByID(ctx context.Context, id uint64) (*model.AnimalSpecies, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockAnimalSpeciesRepository) Create(ctx context.Context, species *model.AnimalSpecies) error {
	if m.createFn != nil {
		return m.createFn(ctx, species)
	}
	return nil
}

func (m *mockAnimalSpeciesRepository) Update(ctx context.Context, id uint64, fields map[string]any) (*model.AnimalSpecies, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, id, fields)
	}
	return &model.AnimalSpecies{}, nil
}

func (m *mockAnimalSpeciesRepository) Delete(ctx context.Context, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockAnimalSpeciesRepository) Reorder(ctx context.Context, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, ids)
	}
	return nil
}

// Note: mockPetRepository is defined in pet_service_test.go (shared across tests in this package)

// ---- Tests ----

func TestAnimalSpeciesService_List(t *testing.T) {
	tests := []struct {
		name     string
		repoData []model.AnimalSpecies
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns all animal species",
			repoData: []model.AnimalSpecies{
				{ID: 1, Name: "犬", IsActive: true},
				{ID: 2, Name: "猫", IsActive: true},
				{ID: 3, Name: "ウサギ", IsActive: true},
				{ID: 4, Name: "ハムスター", IsActive: true},
			},
			repoErr: nil,
			wantLen: 4,
			wantErr: false,
		},
		{
			name:     "returns empty list when no species exist",
			repoData: []model.AnimalSpecies{},
			repoErr:  nil,
			wantLen:  0,
			wantErr:  false,
		},
		{
			name:     "propagates repository error",
			repoData: nil,
			repoErr:  errors.New("db connection error"),
			wantLen:  0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAnimalSpeciesRepository{
				findAllFn: func(_ context.Context) ([]model.AnimalSpecies, error) {
					return tt.repoData, tt.repoErr
				},
			}
			petRepo := &mockPetRepository{}
			svc := NewAnimalSpeciesService(repo, petRepo)

			species, err := svc.List(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, species, tt.wantLen)
			}
		})
	}
}

func TestAnimalSpeciesService_Create(t *testing.T) {
	repo := &mockAnimalSpeciesRepository{
		createFn: func(_ context.Context, s *model.AnimalSpecies) error {
			s.ID = 99
			return nil
		},
	}
	petRepo := &mockPetRepository{}
	svc := NewAnimalSpeciesService(repo, petRepo)

	species, err := svc.Create(context.Background(), &CreateAnimalSpeciesInput{
		Name:      "フェレット",
		IsActive:  true,
		SortOrder: 7,
	}, AnimalSpeciesMutationMeta{})

	assert.NoError(t, err)
	assert.Equal(t, uint64(99), species.ID)
	assert.Equal(t, "フェレット", species.Name)
}

func TestAnimalSpeciesService_Update_EmptyFields(t *testing.T) {
	repo := &mockAnimalSpeciesRepository{}
	petRepo := &mockPetRepository{}
	svc := NewAnimalSpeciesService(repo, petRepo)

	_, err := svc.Update(context.Background(), 1, &UpdateAnimalSpeciesInput{}, AnimalSpeciesMutationMeta{})
	assert.Error(t, err)
}

func TestAnimalSpeciesService_Reorder_EmptyIDs(t *testing.T) {
	repo := &mockAnimalSpeciesRepository{}
	petRepo := &mockPetRepository{}
	svc := NewAnimalSpeciesService(repo, petRepo)

	err := svc.Reorder(context.Background(), []uint64{}, AnimalSpeciesMutationMeta{})
	assert.Error(t, err)
}

// ---- buildAnimalSpeciesUpdate (pure function) ----

func TestBuildAnimalSpeciesUpdate(t *testing.T) {
	name := "フェレット"
	isActive := true
	sortOrder := 3

	tests := []struct {
		name    string
		input   *UpdateAnimalSpeciesInput
		wantLen int
	}{
		{
			name:    "no fields set returns empty map",
			input:   &UpdateAnimalSpeciesInput{},
			wantLen: 0,
		},
		{
			name: "all fields set returns all keys",
			input: &UpdateAnimalSpeciesInput{
				Name:      &name,
				IsActive:  &isActive,
				SortOrder: &sortOrder,
			},
			wantLen: 3,
		},
		{
			name: "single field set returns single key",
			input: &UpdateAnimalSpeciesInput{
				Name: &name,
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := buildAnimalSpeciesUpdate(tt.input)
			assert.Len(t, fields, tt.wantLen)
		})
	}
}

// ---- GetByID ----

func TestAnimalSpeciesService_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockAnimalSpeciesRepository
		wantErr bool
	}{
		{
			name: "returns animal species successfully",
			repo: &mockAnimalSpeciesRepository{
				findByIDFn: func(_ context.Context, id uint64) (*model.AnimalSpecies, error) {
					return &model.AnimalSpecies{ID: id, Name: "犬"}, nil
				},
			},
			wantErr: false,
		},
		{
			name: "propagates repository error",
			repo: &mockAnimalSpeciesRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.AnimalSpecies, error) {
					return nil, errors.New("not found")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			petRepo := &mockPetRepository{}
			svc := NewAnimalSpeciesService(tt.repo, petRepo)

			result, err := svc.GetByID(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// ---- Create: additional validation and error branches ----

func TestAnimalSpeciesService_Create_Validation(t *testing.T) {
	tests := []struct {
		name      string
		input     *CreateAnimalSpeciesInput
		createErr error
		wantErr   bool
	}{
		{
			name:    "returns error for empty name",
			input:   &CreateAnimalSpeciesInput{Name: "", IsActive: true},
			wantErr: true,
		},
		{
			name:      "propagates repository create error",
			input:     &CreateAnimalSpeciesInput{Name: "フェレット", IsActive: true},
			createErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAnimalSpeciesRepository{
				createFn: func(_ context.Context, _ *model.AnimalSpecies) error {
					return tt.createErr
				},
			}
			petRepo := &mockPetRepository{}
			svc := NewAnimalSpeciesService(repo, petRepo)

			result, err := svc.Create(context.Background(), tt.input, AnimalSpeciesMutationMeta{})

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- Update ----

func TestAnimalSpeciesService_Update(t *testing.T) {
	name := "更新後の名前"
	invalidName := ""

	tests := []struct {
		name      string
		id        uint64
		input     *UpdateAnimalSpeciesInput
		findErr   error
		updateErr error
		wantErr   bool
	}{
		{
			name:    "returns error when input is nil",
			id:      1,
			input:   nil,
			wantErr: true,
		},
		{
			name:    "propagates FindByID error",
			id:      1,
			input:   &UpdateAnimalSpeciesInput{Name: &name},
			findErr: errors.New("not found"),
			wantErr: true,
		},
		{
			name:    "returns error for invalid name",
			id:      1,
			input:   &UpdateAnimalSpeciesInput{Name: &invalidName},
			wantErr: true,
		},
		{
			name:    "returns error when no fields provided",
			id:      1,
			input:   &UpdateAnimalSpeciesInput{},
			wantErr: true,
		},
		{
			name:    "updates successfully",
			id:      1,
			input:   &UpdateAnimalSpeciesInput{Name: &name},
			wantErr: false,
		},
		{
			name:      "propagates repository update error",
			id:        1,
			input:     &UpdateAnimalSpeciesInput{Name: &name},
			updateErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAnimalSpeciesRepository{
				findByIDFn: func(_ context.Context, id uint64) (*model.AnimalSpecies, error) {
					if tt.findErr != nil {
						return nil, tt.findErr
					}
					return &model.AnimalSpecies{ID: id}, nil
				},
				updateFieldsFn: func(_ context.Context, id uint64, _ map[string]any) (*model.AnimalSpecies, error) {
					if tt.updateErr != nil {
						return nil, tt.updateErr
					}
					return &model.AnimalSpecies{ID: id, Name: name}, nil
				},
			}
			petRepo := &mockPetRepository{}
			svc := NewAnimalSpeciesService(repo, petRepo)

			result, err := svc.Update(context.Background(), tt.id, tt.input, AnimalSpeciesMutationMeta{})

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// ---- Delete: additional error branches ----

func TestAnimalSpeciesService_Delete_ErrorBranches(t *testing.T) {
	tests := []struct {
		name        string
		findErr     error
		petCountErr error
		deleteErr   error
		wantErr     bool
	}{
		{
			name:    "propagates FindByID error",
			findErr: errors.New("not found"),
			wantErr: true,
		},
		{
			name:        "propagates CountUsageByAnimalSpeciesID error",
			petCountErr: errors.New("db error"),
			wantErr:     true,
		},
		{
			name:      "propagates repository delete error",
			deleteErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAnimalSpeciesRepository{
				findByIDFn: func(_ context.Context, id uint64) (*model.AnimalSpecies, error) {
					if tt.findErr != nil {
						return nil, tt.findErr
					}
					return &model.AnimalSpecies{ID: id}, nil
				},
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.deleteErr
				},
			}
			petRepo := &mockPetRepository{
				countUsageByAnimalSpeciesIDFn: func(_ context.Context, _ uint64) (int64, error) {
					if tt.petCountErr != nil {
						return 0, tt.petCountErr
					}
					return 0, nil
				},
			}
			svc := NewAnimalSpeciesService(repo, petRepo)

			err := svc.Delete(context.Background(), 1, AnimalSpeciesMutationMeta{})

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- Reorder ----

func TestAnimalSpeciesService_Reorder(t *testing.T) {
	tests := []struct {
		name       string
		ids        []uint64
		reorderErr error
		wantErr    bool
	}{
		{
			name:    "reorders successfully",
			ids:     []uint64{3, 1, 2},
			wantErr: false,
		},
		{
			name:       "propagates repository error",
			ids:        []uint64{3, 1, 2},
			reorderErr: errors.New("db error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAnimalSpeciesRepository{
				reorderFn: func(_ context.Context, _ []uint64) error {
					return tt.reorderErr
				},
			}
			petRepo := &mockPetRepository{}
			svc := NewAnimalSpeciesService(repo, petRepo)

			err := svc.Reorder(context.Background(), tt.ids, AnimalSpeciesMutationMeta{})

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAnimalSpeciesService_Delete_WithPetReference(t *testing.T) {
	tests := []struct {
		name      string
		petCount  int64
		wantError bool
		wantCode  string
	}{
		{
			name:      "returns 409 when animal species is referenced by pets",
			petCount:  2,
			wantError: true,
			wantCode:  "CONFLICT",
		},
		{
			name:      "succeeds when animal species has no pet references",
			petCount:  0,
			wantError: false,
			wantCode:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAnimalSpeciesRepository{
				deleteFn: func(_ context.Context, _ uint64) error {
					return nil
				},
			}
			petRepo := &mockPetRepository{
				countUsageByAnimalSpeciesIDFn: func(_ context.Context, _ uint64) (int64, error) {
					return tt.petCount, nil
				},
			}
			svc := NewAnimalSpeciesService(repo, petRepo)

			err := svc.Delete(context.Background(), 1, AnimalSpeciesMutationMeta{})

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- Audit (POC-07 / U-X03-PET-SPECIES-AUDIT) ----

type mockSpeciesAudit struct {
	entries []*audit.Entry
	err     error
}

func (m *mockSpeciesAudit) LogEntryTx(_ context.Context, input *audit.Entry) error {
	if m.err != nil {
		return m.err
	}
	// copy pointer fields for assertions
	cp := *input
	m.entries = append(m.entries, &cp)
	return nil
}

type mockSpeciesTx struct{}

func (mockSpeciesTx) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func testSpeciesMeta() AnimalSpeciesMutationMeta {
	actor := uint64(7)
	return AnimalSpeciesMutationMeta{
		ClinicID:  1,
		ActorID:   &actor,
		ActorType: "staff",
		IPAddress: "127.0.0.1",
		UserAgent: "test",
	}
}

func TestAnimalSpeciesService_Create_WritesAudit(t *testing.T) {
	repo := &mockAnimalSpeciesRepository{
		createFn: func(_ context.Context, species *model.AnimalSpecies) error {
			species.ID = 11
			return nil
		},
	}
	aud := &mockSpeciesAudit{}
	svc := NewAnimalSpeciesServiceWithAudit(repo, &mockPetRepository{}, aud, mockSpeciesTx{})
	got, err := svc.Create(context.Background(), &CreateAnimalSpeciesInput{Name: "鳥", IsActive: true}, testSpeciesMeta())
	assert.NoError(t, err)
	assert.Equal(t, uint64(11), got.ID)
	require.Len(t, aud.entries, 1)
	assert.Equal(t, auditActionAnimalSpeciesCreate, aud.entries[0].Action)
	assert.Equal(t, auditResourceAnimalSpecies, aud.entries[0].Resource)
	assert.Equal(t, uint64(1), *aud.entries[0].ClinicID)
}

func TestAnimalSpeciesService_Update_AuditFailureRollsBackPath(t *testing.T) {
	repo := &mockAnimalSpeciesRepository{
		findByIDFn: func(_ context.Context, id uint64) (*model.AnimalSpecies, error) {
			return &model.AnimalSpecies{ID: id, Name: "旧"}, nil
		},
		updateFieldsFn: func(_ context.Context, id uint64, _ map[string]any) (*model.AnimalSpecies, error) {
			return &model.AnimalSpecies{ID: id, Name: "新"}, nil
		},
	}
	aud := &mockSpeciesAudit{err: errors.New("audit unavailable")}
	svc := NewAnimalSpeciesServiceWithAudit(repo, &mockPetRepository{}, aud, mockSpeciesTx{})
	name := "新"
	_, err := svc.Update(context.Background(), 1, &UpdateAnimalSpeciesInput{Name: &name}, testSpeciesMeta())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audit")
}

func TestAnimalSpeciesService_Delete_WritesAudit(t *testing.T) {
	repo := &mockAnimalSpeciesRepository{
		findByIDFn: func(_ context.Context, id uint64) (*model.AnimalSpecies, error) {
			return &model.AnimalSpecies{ID: id, Name: "削除対象"}, nil
		},
		deleteFn: func(_ context.Context, _ uint64) error { return nil },
	}
	petRepo := &mockPetRepository{
		countUsageByAnimalSpeciesIDFn: func(_ context.Context, _ uint64) (int64, error) { return 0, nil },
	}
	aud := &mockSpeciesAudit{}
	svc := NewAnimalSpeciesServiceWithAudit(repo, petRepo, aud, mockSpeciesTx{})
	err := svc.Delete(context.Background(), 3, testSpeciesMeta())
	assert.NoError(t, err)
	require.Len(t, aud.entries, 1)
	assert.Equal(t, auditActionAnimalSpeciesDelete, aud.entries[0].Action)
}

// ---- DB-backed audit fail-closed proofs (BE-ACT-ANIMAL-SPECIES-DBORTX / POC-07) ----
//
// mockSpeciesTx は fn を呼ぶだけで rollback しないため、監査失敗時に Create/Update/Delete/
// Reorder が DB に残らないことは実 DB + persistence.NewTransactor + 実 repository で証明する。
// リポジトリが r.db.WithContext のままだと ambient tx に参加せず部分コミットになる（RED）。

func newAnimalSpeciesAuditFailService(t *testing.T) (
	AnimalSpeciesService,
	AnimalSpeciesRepository,
	*gorm.DB,
) {
	t.Helper()
	db := setupAnimalSpeciesTestDB(t)
	repo := NewAnimalSpeciesRepository(db)
	aud := &mockSpeciesAudit{err: errors.New("audit unavailable")}
	svc := NewAnimalSpeciesServiceWithAudit(
		repo,
		&mockPetRepository{},
		aud,
		persistence.NewTransactor(db),
	)
	return svc, repo, db
}

func TestAnimalSpeciesService_Create_AuditFailureRollsBack(t *testing.T) {
	svc, _, db := newAnimalSpeciesAuditFailService(t)
	ctx := context.Background()

	got, err := svc.Create(ctx, &CreateAnimalSpeciesInput{
		Name:      "audit-fail-create",
		IsActive:  true,
		SortOrder: 1,
	}, testSpeciesMeta())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "audit")
	assert.Nil(t, got)

	var count int64
	require.NoError(t, db.WithContext(ctx).
		Model(&model.AnimalSpecies{}).
		Where("name = ?", "audit-fail-create").
		Count(&count).Error)
	assert.Equal(t, int64(0), count, "audit failure must roll back Create")
}

func TestAnimalSpeciesService_Update_AuditFailureRollsBack(t *testing.T) {
	svc, repo, _ := newAnimalSpeciesAuditFailService(t)
	ctx := context.Background()

	species := &model.AnimalSpecies{Name: "audit-fail-update-before", SortOrder: 3}
	require.NoError(t, repo.Create(ctx, species))

	name := "audit-fail-update-after"
	got, err := svc.Update(ctx, species.ID, &UpdateAnimalSpeciesInput{Name: &name}, testSpeciesMeta())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "audit")
	assert.Nil(t, got)

	reloaded, err := repo.FindByID(ctx, species.ID)
	require.NoError(t, err)
	assert.Equal(t, "audit-fail-update-before", reloaded.Name, "audit failure must roll back Update")
}

func TestAnimalSpeciesService_Delete_AuditFailureRollsBack(t *testing.T) {
	svc, repo, _ := newAnimalSpeciesAuditFailService(t)
	ctx := context.Background()

	species := &model.AnimalSpecies{Name: "audit-fail-delete-target"}
	require.NoError(t, repo.Create(ctx, species))

	err := svc.Delete(ctx, species.ID, testSpeciesMeta())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "audit")

	reloaded, err := repo.FindByID(ctx, species.ID)
	require.NoError(t, err)
	assert.Equal(t, "audit-fail-delete-target", reloaded.Name, "audit failure must roll back Delete")
}

func TestAnimalSpeciesService_Reorder_AuditFailureRollsBack(t *testing.T) {
	svc, repo, _ := newAnimalSpeciesAuditFailService(t)
	ctx := context.Background()

	first := &model.AnimalSpecies{Name: "audit-fail-reorder-A", SortOrder: 10}
	second := &model.AnimalSpecies{Name: "audit-fail-reorder-B", SortOrder: 20}
	require.NoError(t, repo.Create(ctx, first))
	require.NoError(t, repo.Create(ctx, second))

	err := svc.Reorder(ctx, []uint64{second.ID, first.ID}, testSpeciesMeta())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "audit")

	gotFirst, err := repo.FindByID(ctx, first.ID)
	require.NoError(t, err)
	assert.Equal(t, 10, gotFirst.SortOrder, "audit failure must roll back Reorder")
	gotSecond, err := repo.FindByID(ctx, second.ID)
	require.NoError(t, err)
	assert.Equal(t, 20, gotSecond.SortOrder, "audit failure must roll back Reorder")
}
