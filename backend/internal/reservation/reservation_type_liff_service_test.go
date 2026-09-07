package reservation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockReservationTypeLiffRepository は ReservationTypeLiffRepository のテスト用モック実装
type mockReservationTypeLiffRepository struct {
	findAllFn       func(ctx context.Context, clinicID uint64) ([]model.ReservationType, error)
	findByIDFn      func(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
	countChildrenFn func(ctx context.Context, clinicID, parentID uint64) (int64, error)
	createFn        func(ctx context.Context, st *model.ReservationType) error
	updateFieldsFn  func(ctx context.Context, clinicID, id uint64, cmd UpdateReservationTypeLiffInput) (*model.ReservationType, error)
	deleteFn        func(ctx context.Context, clinicID, id uint64) error
	updateSortOrder func(ctx context.Context, clinicID, id uint64, direction string) error
}

func (m *mockReservationTypeLiffRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockReservationTypeLiffRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockReservationTypeLiffRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	if m.countChildrenFn != nil {
		return m.countChildrenFn(ctx, clinicID, parentID)
	}
	return 0, nil
}

func (m *mockReservationTypeLiffRepository) Create(ctx context.Context, st *model.ReservationType) error {
	return m.createFn(ctx, st)
}

func (m *mockReservationTypeLiffRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateReservationTypeLiffInput) (*model.ReservationType, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, cmd)
	}
	return nil, nil
}

func (m *mockReservationTypeLiffRepository) DeleteWithDependencyChecks(ctx context.Context, clinicID, id uint64, usage reservationTypeUsageChecker) error {
	// Mirror production order: lock/find → children → usage → delete.
	if m.findByIDFn != nil {
		if _, err := m.findByIDFn(ctx, clinicID, id); err != nil {
			return err
		}
	}
	if m.countChildrenFn != nil {
		n, err := m.countChildrenFn(ctx, clinicID, id)
		if err != nil {
			return err
		}
		if n > 0 {
			return apperrors.WrapConflict("この予約コースには子予約区分が登録されているため削除できません")
		}
	}
	if usage != nil {
		exists, err := usage.ExistsByReservationTypeID(ctx, clinicID, id)
		if err != nil {
			return err
		}
		if exists {
			return apperrors.WrapConflict("この予約コースは予約データで使用中のため削除できません")
		}
	}
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockReservationTypeLiffRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockReservationTypeLiffRepository) UpdateSortOrder(ctx context.Context, clinicID, id uint64, direction string) error {
	if m.updateSortOrder != nil {
		return m.updateSortOrder(ctx, clinicID, id, direction)
	}
	return nil
}

// mockReservationQueryRepository は ReservationQueryRepository のテスト用モック実装
type mockReservationQueryRepository struct {
	existsByReservationTypeIDFn func(ctx context.Context, clinicID, reservationTypeID uint64) (bool, error)
}

func (m *mockReservationQueryRepository) ExistsByReservationTypeID(ctx context.Context, clinicID, reservationTypeID uint64) (bool, error) {
	if m.existsByReservationTypeIDFn != nil {
		return m.existsByReservationTypeIDFn(ctx, clinicID, reservationTypeID)
	}
	return false, nil
}

func (m *mockReservationQueryRepository) ExistsByStaffID(_ context.Context, _, _ uint64) (bool, error) {
	return false, nil
}

func (m *mockReservationQueryRepository) CountMedicalRecordsByReservationID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockReservationQueryRepository) CountByCustomerAndDateRange(_ context.Context, _, _ uint64, _, _ time.Time) (int64, error) {
	return 0, nil
}

func (m *mockReservationQueryRepository) CountByDateAndSource(_ context.Context, _ uint64, _ time.Time, _ model.ReservationSource) (int64, error) {
	return 0, nil
}

func (m *mockReservationQueryRepository) FindAllByCategory(_ context.Context, _ uint64, _ model.ReservationTypeCategory, _, _ *uint64, _, _ *string, _, _ int) ([]model.Reservation, int64, error) {
	return nil, 0, nil
}

func (m *mockReservationQueryRepository) FindNoShowCandidates(_ context.Context, _ uint64) ([]model.Reservation, error) {
	return nil, nil
}

func (m *mockReservationQueryRepository) AssertOwnerInClinic(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockReservationQueryRepository) FindPetOwnerInClinic(_ context.Context, _, _ uint64) (uint64, error) {
	return 0, nil
}

func (m *mockReservationQueryRepository) FindPetByIDInClinic(_ context.Context, _, petID uint64) (*model.Pet, error) {
	return &model.Pet{ID: petID, Status: model.PetStatusAlive}, nil
}

func (m *mockReservationQueryRepository) AssertLineCustomerInClinic(_ context.Context, _, _ uint64) error {
	return nil
}

func newTestReservationTypeLiffService(
	repo *mockReservationTypeLiffRepository,
	resRepo *mockReservationQueryRepository,
) ReservationTypeLiffService {
	return NewReservationTypeLiffService(repo, resRepo)
}

func TestReservationTypeLiffService_List_ReturnsTypes(t *testing.T) {
	tests := []struct {
		name      string
		repoTypes []model.ReservationType
		repoErr   error
		wantLen   int
		wantErr   bool
	}{
		{
			name: "returns reservation type list",
			repoTypes: []model.ReservationType{
				{ID: 1, ClinicID: 1, Name: "一般診察", IsActive: true},
				{ID: 2, ClinicID: 1, Name: "トリミング", IsActive: true},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:      "returns empty list when no types exist",
			repoTypes: []model.ReservationType{},
			repoErr:   nil,
			wantLen:   0,
			wantErr:   false,
		},
		{
			name:      "propagates repository error",
			repoTypes: nil,
			repoErr:   errors.New("db connection error"),
			wantLen:   0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationTypeLiffRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.ReservationType, error) {
					return tt.repoTypes, tt.repoErr
				},
			}
			resRepo := &mockReservationQueryRepository{}
			svc := newTestReservationTypeLiffService(repo, resRepo)

			types, err := svc.List(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, types, tt.wantLen)
			}
		})
	}
}

func TestReservationTypeLiffService_GetByID_Found(t *testing.T) {
	// Delete + PatchStatus 以外で GetByID に相当する公開メソッドはないため、
	// Delete を経由してリポジトリの FindByID が呼ばれることを検証する代わりに、
	// Update を通じて間接的に FindByID を呼び出す動作を確認する。
	expected := &model.ReservationType{
		ID:       1,
		ClinicID: 1,
		Name:     "一般診察",
		IsActive: true,
	}
	name := "一般診察（更新）"
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return expected, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateReservationTypeLiffInput) (*model.ReservationType, error) {
			return &model.ReservationType{ID: 1, ClinicID: 1, Name: name, IsActive: true}, nil
		},
	}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	updated, err := svc.Update(context.Background(), 1, 1, &UpdateReservationTypeLiffInput{Name: &name})

	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, name, updated.Name)
}

func TestReservationTypeLiffService_GetByID_NotFound(t *testing.T) {
	// Update を通じて FindByID の not found エラー伝播を確認する
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return nil, apperrors.WrapNotFound("reservation_type_liff", "999")
		},
	}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	name := "なんでもいい"
	result, err := svc.Update(context.Background(), 1, 999, &UpdateReservationTypeLiffInput{Name: &name})

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, result)
}

func TestReservationTypeLiffService_Create_Success(t *testing.T) {
	input := &CreateReservationTypeLiffInput{
		Name:            "新コース",
		Color:           "#FF5733",
		Description:     "テスト説明",
		SortOrder:       1,
		DurationMinutes: 30,
	}
	created := &model.ReservationType{
		ID:       10,
		ClinicID: 1,
		Name:     input.Name,
		Color:    input.Color,
		IsActive: true,
	}
	repo := &mockReservationTypeLiffRepository{
		createFn: func(_ context.Context, st *model.ReservationType) error {
			st.ID = 10
			return nil
		},
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return created, nil
		},
	}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	result, err := svc.Create(context.Background(), 1, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "新コース", result.Name)
}

func TestReservationTypeLiffService_Delete_InUse_Returns409(t *testing.T) {
	existing := &model.ReservationType{ID: 1, ClinicID: 1, Name: "使用中コース", IsActive: true}
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return existing, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			return nil
		},
	}
	resRepo := &mockReservationQueryRepository{
		existsByReservationTypeIDFn: func(_ context.Context, _, _ uint64) (bool, error) {
			// 予約データが存在する（使用中）
			return true, nil
		},
	}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	err := svc.Delete(context.Background(), 1, 1)

	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
}

func TestReservationTypeLiffService_Delete_WithChildren_Returns409(t *testing.T) {
	existing := &model.ReservationType{ID: 1, ClinicID: 1, Name: "親コース", IsActive: true}
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return existing, nil
		},
		countChildrenFn: func(_ context.Context, _, _ uint64) (int64, error) {
			return 2, nil
		},
	}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	err := svc.Delete(context.Background(), 1, 1)

	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
}

func TestReservationTypeLiffService_Delete_NotFound(t *testing.T) {
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return nil, apperrors.WrapNotFound("reservation_type_liff", "999")
		},
	}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	err := svc.Delete(context.Background(), 1, 999)

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

// ---- buildReservationTypeLiffUpdate (pure function) ----

func TestBuildReservationTypeLiffUpdate(t *testing.T) {
	name := "コース"
	color := "#FFFFFF"
	desc := "説明"
	sortOrder := 2
	duration := 30
	maxConcurrent := 3
	shortName := "コ"
	showShortName := true
	visible := true
	comment := "コメント"
	dayOption := "weekday"
	isInternal := true

	tests := []struct {
		name  string
		input *UpdateReservationTypeLiffInput
		want  map[string]any
	}{
		{
			name: "all fields set",
			input: &UpdateReservationTypeLiffInput{
				Name: &name, Color: &color, Description: &desc, SortOrder: &sortOrder,
				DurationMinutes: &duration, MaxConcurrent: &maxConcurrent,
				ShortName: &shortName, ShowShortName: &showShortName,
				ReservationVisible: &visible, ReservationComment: &comment,
				ReservationDayOption: &dayOption, IsInternal: &isInternal,
			},
			want: map[string]any{
				colReservationTypeLiffName:                 name,
				colReservationTypeLiffColor:                color,
				colReservationTypeLiffDescription:          desc,
				colReservationTypeLiffSortOrder:            sortOrder,
				colReservationTypeLiffDurationMinutes:      duration,
				colReservationTypeLiffMaxConcurrent:        maxConcurrent,
				colReservationTypeLiffShortName:            shortName,
				colReservationTypeLiffShowShortName:        showShortName,
				colReservationTypeLiffReservationVisible:   visible,
				colReservationTypeLiffReservationComment:   comment,
				colReservationTypeLiffReservationDayOption: dayOption,
				colReservationTypeLiffIsInternal:           isInternal,
			},
		},
		{
			name:  "all nil returns empty map",
			input: &UpdateReservationTypeLiffInput{},
			want:  map[string]any{},
		},
		{
			name:  "ClearMaxConcurrent takes precedence over MaxConcurrent",
			input: &UpdateReservationTypeLiffInput{ClearMaxConcurrent: true, MaxConcurrent: &maxConcurrent},
			want: map[string]any{
				colReservationTypeLiffMaxConcurrent: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildReservationTypeLiffUpdate(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---- Create: remaining branches ----

func TestReservationTypeLiffService_Create_ValidationError(t *testing.T) {
	repo := &mockReservationTypeLiffRepository{}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	badMax := 0
	input := &CreateReservationTypeLiffInput{Name: "コース", MaxConcurrent: &badMax}
	result, err := svc.Create(context.Background(), 1, input)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestReservationTypeLiffService_Create_RepoCreateError(t *testing.T) {
	repo := &mockReservationTypeLiffRepository{
		createFn: func(_ context.Context, _ *model.ReservationType) error {
			return errors.New("db error")
		},
	}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	input := &CreateReservationTypeLiffInput{Name: "コース"}
	result, err := svc.Create(context.Background(), 1, input)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestReservationTypeLiffService_Create_ReturnsWriteResultWithoutReload(t *testing.T) {
	// RSV-03: Create returns the write result; post-commit FindByID is no longer used.
	repo := &mockReservationTypeLiffRepository{
		createFn: func(_ context.Context, st *model.ReservationType) error {
			st.ID = 10
			return nil
		},
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return nil, errors.New("db error")
		},
	}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	input := &CreateReservationTypeLiffInput{Name: "コース", ReservationDayOption: "weekday"}
	result, err := svc.Create(context.Background(), 1, input)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, uint64(10), result.ID)
	assert.Equal(t, "コース", result.Name)
}

// ---- Update: remaining branches ----

func TestReservationTypeLiffService_Update_Success(t *testing.T) {
	existing := &model.ReservationType{ID: 1, ClinicID: 1, Name: "元コース"}
	updated := &model.ReservationType{ID: 1, ClinicID: 1, Name: "更新後コース"}
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return existing, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, cmd UpdateReservationTypeLiffInput) (*model.ReservationType, error) {
			require.NotNil(t, cmd.Name)
			assert.Equal(t, "更新後コース", *cmd.Name)
			return updated, nil
		},
	}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	name := "更新後コース"
	result, err := svc.Update(context.Background(), 1, 1, &UpdateReservationTypeLiffInput{Name: &name})

	assert.NoError(t, err)
	assert.Equal(t, updated, result)
}

func TestReservationTypeLiffService_Update_ValidationError(t *testing.T) {
	existing := &model.ReservationType{ID: 1, ClinicID: 1}
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return existing, nil
		},
	}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	badMax := -1
	result, err := svc.Update(context.Background(), 1, 1, &UpdateReservationTypeLiffInput{MaxConcurrent: &badMax})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestReservationTypeLiffService_Update_NoFields(t *testing.T) {
	existing := &model.ReservationType{ID: 1, ClinicID: 1, Name: "現状維持"}
	callCount := 0
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			callCount++
			return existing, nil
		},
	}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	result, err := svc.Update(context.Background(), 1, 1, &UpdateReservationTypeLiffInput{})

	assert.NoError(t, err)
	assert.Equal(t, existing, result)
	assert.Equal(t, 2, callCount)
}

func TestReservationTypeLiffService_Update_NoFields_SecondFindByIDError(t *testing.T) {
	existing := &model.ReservationType{ID: 1, ClinicID: 1}
	callCount := 0
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			callCount++
			if callCount == 1 {
				return existing, nil
			}
			return nil, errors.New("db error")
		},
	}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	result, err := svc.Update(context.Background(), 1, 1, &UpdateReservationTypeLiffInput{})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestReservationTypeLiffService_Update_RepoError(t *testing.T) {
	existing := &model.ReservationType{ID: 1, ClinicID: 1}
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return existing, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateReservationTypeLiffInput) (*model.ReservationType, error) {
			return nil, errors.New("db error")
		},
	}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	name := "更新"
	result, err := svc.Update(context.Background(), 1, 1, &UpdateReservationTypeLiffInput{Name: &name})

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ---- Delete: remaining branches ----

func TestReservationTypeLiffService_Delete_Success(t *testing.T) {
	existing := &model.ReservationType{ID: 1, ClinicID: 1}
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return existing, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error { return nil },
	}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	err := svc.Delete(context.Background(), 1, 1)

	assert.NoError(t, err)
}

func TestReservationTypeLiffService_Delete_CountChildrenError(t *testing.T) {
	existing := &model.ReservationType{ID: 1, ClinicID: 1}
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return existing, nil
		},
		countChildrenFn: func(_ context.Context, _, _ uint64) (int64, error) {
			return 0, errors.New("db error")
		},
	}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	err := svc.Delete(context.Background(), 1, 1)

	assert.Error(t, err)
}

func TestReservationTypeLiffService_Delete_ExistsError(t *testing.T) {
	existing := &model.ReservationType{ID: 1, ClinicID: 1}
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return existing, nil
		},
	}
	resRepo := &mockReservationQueryRepository{
		existsByReservationTypeIDFn: func(_ context.Context, _, _ uint64) (bool, error) {
			return false, errors.New("db error")
		},
	}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	err := svc.Delete(context.Background(), 1, 1)

	assert.Error(t, err)
}

func TestReservationTypeLiffService_Delete_RepoDeleteError(t *testing.T) {
	existing := &model.ReservationType{ID: 1, ClinicID: 1}
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return existing, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			return errors.New("db error")
		},
	}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	err := svc.Delete(context.Background(), 1, 1)

	assert.Error(t, err)
}

// ---- PatchStatus ----

func TestReservationTypeLiffService_PatchStatus_Success(t *testing.T) {
	existing := &model.ReservationType{ID: 1, ClinicID: 1, IsActive: true}
	updated := &model.ReservationType{ID: 1, ClinicID: 1, IsActive: false}
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return existing, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, cmd UpdateReservationTypeLiffInput) (*model.ReservationType, error) {
			require.NotNil(t, cmd.IsActive)
			assert.Equal(t, false, *cmd.IsActive)
			return updated, nil
		},
	}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	result, err := svc.PatchStatus(context.Background(), 1, 1, false)

	assert.NoError(t, err)
	assert.False(t, result.IsActive)
}

func TestReservationTypeLiffService_PatchStatus_NotFound(t *testing.T) {
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return nil, apperrors.WrapNotFound("reservation_type_liff", "1")
		},
	}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	result, err := svc.PatchStatus(context.Background(), 1, 1, true)

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, result)
}

func TestReservationTypeLiffService_PatchStatus_RepoError(t *testing.T) {
	existing := &model.ReservationType{ID: 1, ClinicID: 1}
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return existing, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateReservationTypeLiffInput) (*model.ReservationType, error) {
			return nil, errors.New("db error")
		},
	}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	result, err := svc.PatchStatus(context.Background(), 1, 1, true)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ---- PatchSortOrder ----

func TestReservationTypeLiffService_PatchSortOrder_Success(t *testing.T) {
	existing := &model.ReservationType{ID: 1, ClinicID: 1}
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return existing, nil
		},
		updateSortOrder: func(_ context.Context, _, _ uint64, direction string) error {
			assert.Equal(t, "up", direction)
			return nil
		},
	}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	err := svc.PatchSortOrder(context.Background(), 1, 1, "up")

	assert.NoError(t, err)
}

func TestReservationTypeLiffService_PatchSortOrder_InvalidDirection(t *testing.T) {
	repo := &mockReservationTypeLiffRepository{}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	err := svc.PatchSortOrder(context.Background(), 1, 1, "sideways")

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestReservationTypeLiffService_PatchSortOrder_NotFound(t *testing.T) {
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return nil, apperrors.WrapNotFound("reservation_type_liff", "1")
		},
	}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	err := svc.PatchSortOrder(context.Background(), 1, 1, "down")

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestReservationTypeLiffService_PatchSortOrder_RepoError(t *testing.T) {
	existing := &model.ReservationType{ID: 1, ClinicID: 1}
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return existing, nil
		},
		updateSortOrder: func(_ context.Context, _, _ uint64, _ string) error {
			return errors.New("db error")
		},
	}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resRepo)

	err := svc.PatchSortOrder(context.Background(), 1, 1, "up")

	assert.Error(t, err)
}
