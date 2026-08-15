package owner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockOwnerRepository は OwnerRepository のテスト用モック実装
type mockOwnerRepository struct {
	findAllFn                     func(ctx context.Context, clinicIDs []uint64, page, limit int, search string) ([]model.Owner, int64, error)
	findByIDFn                    func(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
	findByIDsFn                   func(ctx context.Context, clinicID uint64, ids []uint64) ([]*model.Owner, error)
	findByEmailFn                 func(ctx context.Context, clinicID uint64, email string) (*model.Owner, error)
	findByPhoneFn                 func(ctx context.Context, clinicID uint64, phone string) (*model.Owner, error)
	createWithPetsFn              func(ctx context.Context, owner *model.Owner, pets []model.Pet) error
	updateFn                      func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	updateAndFindFn               func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Owner, error)
	deleteFn                      func(ctx context.Context, clinicID, id uint64) error
	countPetsByOwnerIDFn          func(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	findByLineUserIDFn            func(ctx context.Context, clinicID uint64, lineUserID string) (*model.Owner, error)
	updateLineUserIDFn            func(ctx context.Context, clinicID, id uint64, lineUserID *string) error
	findAllWithLineUserIDFn       func(ctx context.Context, clinicID uint64) ([]model.Owner, error)
	findAllWithLineUserIDCursorFn func(ctx context.Context, clinicID uint64, afterID uint64, limit int) ([]model.Owner, error)
}

func (m *mockOwnerRepository) FindAll(ctx context.Context, clinicIDs []uint64, page, limit int, search string) ([]model.Owner, int64, error) {
	return m.findAllFn(ctx, clinicIDs, page, limit, search)
}

func (m *mockOwnerRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockOwnerRepository) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.Owner, error) {
	return nil, nil
}

func (m *mockOwnerRepository) FindByEmail(ctx context.Context, clinicID uint64, email string) (*model.Owner, error) {
	if m.findByEmailFn != nil {
		return m.findByEmailFn(ctx, clinicID, email)
	}
	return nil, nil
}

func (m *mockOwnerRepository) FindByPhone(ctx context.Context, clinicID uint64, phone string) (*model.Owner, error) {
	if m.findByPhoneFn != nil {
		return m.findByPhoneFn(ctx, clinicID, phone)
	}
	return nil, nil
}

func (m *mockOwnerRepository) FindByNameAndPhone(_ context.Context, _ uint64, _, _ string) (*model.Owner, error) {
	return nil, nil
}

func (m *mockOwnerRepository) CreateWithPets(ctx context.Context, owner *model.Owner, pets []model.Pet) error {
	if m.createWithPetsFn != nil {
		return m.createWithPetsFn(ctx, owner, pets)
	}
	return nil
}

func (m *mockOwnerRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return m.updateFn(ctx, clinicID, id, fields)
}

func (m *mockOwnerRepository) UpdateAndFind(
	ctx context.Context,
	clinicID, id uint64,
	fields map[string]any,
) (*model.Owner, error) {
	return m.UpdateAndFindApplying(ctx, clinicID, id, func(_ *model.Owner) (map[string]any, error) {
		return fields, nil
	})
}

func (m *mockOwnerRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
	return m.FindByID(ctx, clinicID, id)
}

func (m *mockOwnerRepository) UpdateAndFindApplying(
	ctx context.Context,
	clinicID, id uint64,
	apply OwnerUpdateApplier,
) (*model.Owner, error) {
	if apply == nil {
		return nil, apperrors.WrapInternalServerError("owner update applier is required")
	}
	locked, err := m.LockByIDForUpdate(ctx, clinicID, id)
	if err != nil {
		return nil, err
	}
	fields, err := apply(locked)
	if err != nil {
		return nil, err
	}
	if m.updateAndFindFn != nil {
		return m.updateAndFindFn(ctx, clinicID, id, fields)
	}
	if m.updateFn != nil {
		if err := m.updateFn(ctx, clinicID, id, fields); err != nil {
			return nil, err
		}
	}
	return m.FindByID(ctx, clinicID, id)
}

func (m *mockOwnerRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockOwnerRepository) CountPetsByOwnerID(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	if m.countPetsByOwnerIDFn != nil {
		return m.countPetsByOwnerIDFn(ctx, clinicID, ownerID)
	}
	return 0, nil
}

func (m *mockOwnerRepository) FindByLineUserID(ctx context.Context, clinicID uint64, lineUserID string) (*model.Owner, error) {
	if m.findByLineUserIDFn != nil {
		return m.findByLineUserIDFn(ctx, clinicID, lineUserID)
	}
	return nil, nil
}

func (m *mockOwnerRepository) FindAllWithLineUserID(ctx context.Context, clinicID uint64) ([]model.Owner, error) {
	if m.findAllWithLineUserIDFn != nil {
		return m.findAllWithLineUserIDFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockOwnerRepository) FindAllWithLineUserIDCursor(ctx context.Context, clinicID, afterID uint64, limit int) ([]model.Owner, error) {
	if m.findAllWithLineUserIDCursorFn != nil {
		return m.findAllWithLineUserIDCursorFn(ctx, clinicID, afterID, limit)
	}
	return nil, nil
}

func (m *mockOwnerRepository) FindByIDs(ctx context.Context, clinicID uint64, ids []uint64) ([]*model.Owner, error) {
	if m.findByIDsFn != nil {
		return m.findByIDsFn(ctx, clinicID, ids)
	}
	return nil, nil
}

func (m *mockOwnerRepository) UpdateLineUserID(ctx context.Context, clinicID, id uint64, lineUserID *string) error {
	if m.updateLineUserIDFn != nil {
		return m.updateLineUserIDFn(ctx, clinicID, id, lineUserID)
	}
	return nil
}

func (m *mockOwnerRepository) UpdateLineFollowedAt(
	_ context.Context,
	_, _ uint64,
	_ string,
	_ time.Time,
) (bool, error) {
	return true, nil
}

func (m *mockOwnerRepository) UpdateLineBlockedAt(
	_ context.Context,
	_, _ uint64,
	_ string,
	_ time.Time,
) (bool, error) {
	return true, nil
}

// ポインタヘルパー関数（ptrString は accounting_service_test.go で定義済み）
func ptrBool(b bool) *bool          { return &b }
func ptrFloat64(f float64) *float64 { return &f }

func TestOwnerService_List(t *testing.T) {
	tests := []struct {
		name       string
		clinicID   uint64
		page       int
		limit      int
		search     string
		repoOwners []model.Owner
		repoTotal  int64
		repoErr    error
		wantLen    int
		wantTotal  int64
		wantErr    bool
	}{
		{
			name:     "returns owner list with total count",
			clinicID: 1,
			page:     1,
			limit:    20,
			search:   "",
			repoOwners: []model.Owner{
				{ID: 1, ClinicID: 1, Name: "山田 太郎"},
				{ID: 2, ClinicID: 1, Name: "鈴木 花子"},
			},
			repoTotal: 2,
			repoErr:   nil,
			wantLen:   2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:       "returns empty list when no owners exist",
			clinicID:   1,
			page:       1,
			limit:      20,
			search:     "",
			repoOwners: []model.Owner{},
			repoTotal:  0,
			repoErr:    nil,
			wantLen:    0,
			wantTotal:  0,
			wantErr:    false,
		},
		{
			name:       "filters by search keyword",
			clinicID:   1,
			page:       1,
			limit:      20,
			search:     "山田",
			repoOwners: []model.Owner{{ID: 1, ClinicID: 1, Name: "山田 太郎"}},
			repoTotal:  1,
			repoErr:    nil,
			wantLen:    1,
			wantTotal:  1,
			wantErr:    false,
		},
		{
			name:       "propagates repository error",
			clinicID:   1,
			page:       1,
			limit:      20,
			search:     "",
			repoOwners: nil,
			repoTotal:  0,
			repoErr:    errors.New("db connection error"),
			wantLen:    0,
			wantTotal:  0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOwnerRepository{
				findAllFn: func(_ context.Context, _ []uint64, _, _ int, _ string) ([]model.Owner, int64, error) {
					return tt.repoOwners, tt.repoTotal, tt.repoErr
				},
			}
			svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)

			owners, total, err := svc.List(context.Background(), []uint64{tt.clinicID}, tt.page, tt.limit, tt.search)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, owners, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
			}
		})
	}
}

// TestOwnerService_List_KanaSearch はひらがな・カタカナ検索語をリポジトリにそのまま渡すことを保証する。
// 正規化はリポジトリ層 (NormalizeKana + translate SQL) が担うため、サービスは変換しない。
func TestOwnerService_List_KanaSearch(t *testing.T) {
	tests := []struct {
		name          string
		searchInput   string
		wantSearchArg string // repository に渡るべき search 引数
		repoOwners    []model.Owner
	}{
		{
			name:          "ひらがな検索語をそのまま repository に渡す",
			searchInput:   "ぴ",
			wantSearchArg: "ぴ",
			repoOwners:    []model.Owner{{ID: 1, ClinicID: 1, Name: "田中", NameKana: "ピーター"}},
		},
		{
			name:          "カタカナ検索語をそのまま repository に渡す",
			searchInput:   "ピーター",
			wantSearchArg: "ピーター",
			repoOwners:    []model.Owner{{ID: 1, ClinicID: 1, Name: "田中", NameKana: "ピーター"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedSearch string
			repo := &mockOwnerRepository{
				findAllFn: func(_ context.Context, _ []uint64, _, _ int, search string) ([]model.Owner, int64, error) {
					capturedSearch = search
					return tt.repoOwners, int64(len(tt.repoOwners)), nil
				},
			}
			svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)

			owners, _, err := svc.List(context.Background(), []uint64{1}, 1, 20, tt.searchInput)

			assert.NoError(t, err)
			assert.Len(t, owners, len(tt.repoOwners))
			assert.Equal(t, tt.wantSearchArg, capturedSearch, "service はカナ変換せずそのまま repository に渡す")
		})
	}
}

func TestOwnerService_GetByID(t *testing.T) {
	tests := []struct {
		name      string
		clinicID  uint64
		id        uint64
		repoOwner *model.Owner
		repoErr   error
		wantOwner *model.Owner
		wantErr   error
	}{
		{
			name:      "returns owner when found",
			clinicID:  1,
			id:        10,
			repoOwner: &model.Owner{ID: 10, ClinicID: 1, Name: "山田 太郎"},
			repoErr:   nil,
			wantOwner: &model.Owner{ID: 10, ClinicID: 1, Name: "山田 太郎"},
			wantErr:   nil,
		},
		{
			name:      "returns not found error when owner does not exist",
			clinicID:  1,
			id:        999,
			repoOwner: nil,
			repoErr:   apperrors.WrapNotFound("owner", "999"),
			wantOwner: nil,
			wantErr:   apperrors.ErrNotFound,
		},
		{
			name:      "returns error on repository failure",
			clinicID:  1,
			id:        10,
			repoOwner: nil,
			repoErr:   errors.New("db error"),
			wantOwner: nil,
			wantErr:   errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return tt.repoOwner, tt.repoErr
				},
			}
			svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)

			owner, err := svc.GetByID(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr != nil {
				assert.Error(t, err)
				if errors.Is(tt.wantErr, apperrors.ErrNotFound) {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantOwner, owner)
			}
		})
	}
}

func TestOwnerService_GetByID_NotFound(t *testing.T) {
	repo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return nil, apperrors.WrapNotFound("owner", "999")
		},
	}
	svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)

	owner, err := svc.GetByID(context.Background(), 1, 999)

	assert.Nil(t, owner)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestOwnerService_CreateWithPets(t *testing.T) {
	tests := []struct {
		name      string
		clinicID  uint64
		input     CreateOwnerInput
		repoErr   error
		wantErr   bool
		wantOwner bool // ownerが返されるか
	}{
		{
			name:     "creates owner with pets atomically",
			clinicID: 1,
			input: CreateOwnerInput{
				OwnerName: "林 文昭",
				Email:     "hayashi@example.com",
				Pets: []CreatePetForOwnerInput{
					{Name: "ポチ", AnimalSpeciesID: 1, Gender: "male"},
					{Name: "タマ", AnimalSpeciesID: 2, Gender: "female"},
				},
			},
			repoErr:   nil,
			wantErr:   false,
			wantOwner: true,
		},
		{
			name:     "creates owner without pets",
			clinicID: 1,
			input: CreateOwnerInput{
				OwnerName: "鈴木 次郎",
				Pets:      []CreatePetForOwnerInput{},
			},
			repoErr:   nil,
			wantErr:   false,
			wantOwner: true,
		},
		{
			name:     "rejects invalid discount_rate",
			clinicID: 1,
			input: CreateOwnerInput{
				OwnerName:    "バリデーション",
				DiscountRate: 150,
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:     "rejects invalid membership_type",
			clinicID: 1,
			input: CreateOwnerInput{
				OwnerName:      "バリデーション",
				MembershipType: "invalid_type",
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:     "rejects invalid pet gender",
			clinicID: 1,
			input: CreateOwnerInput{
				OwnerName: "バリデーション",
				Pets: []CreatePetForOwnerInput{
					{Name: "ポチ", AnimalSpeciesID: 1, Gender: "invalid"},
				},
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:     "propagates repository error",
			clinicID: 1,
			input: CreateOwnerInput{
				OwnerName: "エラー 飼主",
				Pets:      []CreatePetForOwnerInput{{Name: "ペット", AnimalSpeciesID: 1}},
			},
			repoErr: errors.New("transaction failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOwnerRepository{
				createWithPetsFn: func(_ context.Context, _ *model.Owner, _ []model.Pet) error {
					return tt.repoErr
				},
			}
			svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)

			owner, err := svc.CreateWithPets(context.Background(), tt.clinicID, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, owner)
			} else {
				assert.NoError(t, err)
				if tt.wantOwner {
					assert.NotNil(t, owner)
					assert.Equal(t, tt.clinicID, owner.ClinicID)
					assert.Equal(t, tt.input.OwnerName, owner.Name)
				}
			}
		})
	}
}

func TestOwnerService_CreateWithPets_SyncAnimalClassificationTagsBestEffort(t *testing.T) {
	syncCalled := false
	repo := &mockOwnerRepository{
		createWithPetsFn: func(_ context.Context, owner *model.Owner, _ []model.Pet) error {
			owner.ID = 123
			return nil
		},
	}
	tagSync := &mockLstepTagSyncService{
		syncOwnerAnimalClassificationTagFn: func(_ context.Context, clinicID, ownerID uint64) error {
			syncCalled = true
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(123), ownerID)
			return errors.New("lstep unavailable")
		},
	}
	svc := NewService(repo, nil, tagSync, nil)

	owner, err := svc.CreateWithPets(context.Background(), 1, &CreateOwnerInput{
		OwnerName: "同期 太郎",
		Pets: []CreatePetForOwnerInput{
			{Name: "ポチ", AnimalSpeciesID: 1},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, owner)
	assert.True(t, syncCalled)
}

func TestOwnerService_Update(t *testing.T) {
	updatedOwner := &model.Owner{ID: 1, ClinicID: 1, Name: "更新後 氏名"}

	tests := []struct {
		name       string
		clinicID   uint64
		id         uint64
		input      UpdateOwnerInput
		updateErr  error
		findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
		wantErr    bool
		wantOwner  bool
	}{
		{
			name:     "updates owner successfully",
			clinicID: 1,
			id:       1,
			input: UpdateOwnerInput{
				OwnerName:           ptrString("更新後 氏名"),
				DiscountRate:        ptrFloat64(10),
				DiscountEditAllowed: true, // SEC-CS-F15: discount change requires edit flag
			},
			updateErr: nil,
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return updatedOwner, nil
			},
			wantErr:   false,
			wantOwner: true,
		},
		{
			name:     "rejects invalid discount_rate via pointer",
			clinicID: 1,
			id:       1,
			input: UpdateOwnerInput{
				DiscountRate: ptrFloat64(-5),
			},
			updateErr: nil,
			wantErr:   true,
		},
		{
			name:     "rejects invalid membership_type via pointer",
			clinicID: 1,
			id:       1,
			input: UpdateOwnerInput{
				MembershipType: func() *model.MembershipType {
					v := model.MembershipType("unknown_type")
					return &v
				}(),
			},
			updateErr: nil,
			wantErr:   true,
		},
		{
			name:      "returns error when no fields provided",
			clinicID:  1,
			id:        1,
			input:     UpdateOwnerInput{}, // 全 nil
			updateErr: nil,
			wantErr:   true,
		},
		{
			name:     "returns not found error when owner does not exist",
			clinicID: 1,
			id:       999,
			input: UpdateOwnerInput{
				OwnerName: ptrString("存在しない 飼主"),
			},
			updateErr: apperrors.WrapNotFound("owner", "999"),
			wantErr:   true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       1,
			input: UpdateOwnerInput{
				OwnerName: ptrString("エラー ケース"),
			},
			updateErr: errors.New("db error"),
			wantErr:   true,
		},
		{
			name:     "updates is_dangerous to false (zero value)",
			clinicID: 1,
			id:       1,
			input: UpdateOwnerInput{
				IsDangerous: ptrBool(false),
			},
			updateErr: nil,
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return &model.Owner{ID: 1, ClinicID: 1, IsDangerous: false}, nil
			},
			wantErr:   false,
			wantOwner: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findByIDFn := tt.findByIDFn
			if findByIDFn == nil {
				findByIDFn = func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return nil, errors.New("findByID should not be called")
				}
			}
			repo := &mockOwnerRepository{
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
					return tt.updateErr
				},
				findByIDFn: findByIDFn,
			}
			svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)

			owner, err := svc.Update(context.Background(), tt.clinicID, tt.id, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, owner)
			} else {
				assert.NoError(t, err)
				if tt.wantOwner {
					assert.NotNil(t, owner)
				}
			}
		})
	}
}
