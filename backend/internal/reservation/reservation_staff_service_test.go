package reservation

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockReservationStaffRepository は ReservationStaffRepository のテスト用モック実装
type mockReservationStaffRepository struct {
	findAllFn                              func(ctx context.Context, clinicID uint64) ([]model.Staff, error)
	findByIDFn                             func(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	lockForMutationFn                      func(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	createFn                               func(ctx context.Context, staff *model.Staff, clinicID uint64) error
	updateFieldsFn                         func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn                               func(ctx context.Context, clinicID, id uint64) error
	countUsageByStaffIDFn                  func() (int64, error)
	swapSortOrderFn                        func(ctx context.Context, clinicID, id uint64, direction string) error
	findExcludedReservationTypesFn         func(ctx context.Context, clinicID, staffID uint64) ([]model.StaffReservationExclusion, error)
	findExcludedReservationTypesByStaffIDs func(ctx context.Context, clinicID uint64, staffIDs []uint64) ([]model.StaffReservationExclusion, error)
	replaceExcludedReservationTypesFn      func(ctx context.Context, clinicID, staffID uint64, courseIDs []uint64) error
	findCapabilitiesFn                     func(ctx context.Context, clinicID, staffID uint64) ([]model.StaffReservationCapability, error)
	findCapabilitiesByStaffIDsFn           func(ctx context.Context, clinicID uint64, staffIDs []uint64) ([]model.StaffReservationCapability, error)
	replaceCapabilitiesFn                  func(ctx context.Context, clinicID, staffID uint64, typeIDs []uint64) error
	supportsReservationTypeFn              func(ctx context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error)
}

func (m *mockReservationStaffRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Staff, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockReservationStaffRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockReservationStaffRepository) LockForMutation(ctx context.Context, clinicID, id uint64) (*model.Staff, error) {
	if m.lockForMutationFn != nil {
		return m.lockForMutationFn(ctx, clinicID, id)
	}
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.Staff{ID: id, ClinicID: clinicID}, nil
}

func (m *mockReservationStaffRepository) Create(ctx context.Context, staff *model.Staff, clinicID uint64) error {
	return m.createFn(ctx, staff, clinicID)
}

func (m *mockReservationStaffRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return nil
}

func (m *mockReservationStaffRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockReservationStaffRepository) CountUsageByStaffID(_ context.Context, _, _ uint64) (int64, error) {
	if m.countUsageByStaffIDFn != nil {
		return m.countUsageByStaffIDFn()
	}
	return 0, nil
}

func (m *mockReservationStaffRepository) UpdateSortOrder(ctx context.Context, clinicID, id uint64, direction string) error {
	if m.swapSortOrderFn != nil {
		return m.swapSortOrderFn(ctx, clinicID, id, direction)
	}
	return nil
}

func (m *mockReservationStaffRepository) FindAllExcludedReservationTypes(ctx context.Context, clinicID, staffID uint64) ([]model.StaffReservationExclusion, error) {
	if m.findExcludedReservationTypesFn != nil {
		return m.findExcludedReservationTypesFn(ctx, clinicID, staffID)
	}
	return []model.StaffReservationExclusion{}, nil
}

func (m *mockReservationStaffRepository) FindAllExcludedReservationTypesByStaffIDs(ctx context.Context, clinicID uint64, staffIDs []uint64) ([]model.StaffReservationExclusion, error) {
	if m.findExcludedReservationTypesByStaffIDs != nil {
		return m.findExcludedReservationTypesByStaffIDs(ctx, clinicID, staffIDs)
	}
	return []model.StaffReservationExclusion{}, nil
}

func (m *mockReservationStaffRepository) UpdateExcludedReservationTypes(ctx context.Context, clinicID, staffID uint64, courseIDs []uint64) error {
	if m.replaceExcludedReservationTypesFn != nil {
		return m.replaceExcludedReservationTypesFn(ctx, clinicID, staffID, courseIDs)
	}
	return nil
}

func (m *mockReservationStaffRepository) FindAllReservationCapabilities(ctx context.Context, clinicID, staffID uint64) ([]model.StaffReservationCapability, error) {
	if m.findCapabilitiesFn != nil {
		return m.findCapabilitiesFn(ctx, clinicID, staffID)
	}
	return []model.StaffReservationCapability{}, nil
}

func (m *mockReservationStaffRepository) FindAllReservationCapabilitiesByStaffIDs(ctx context.Context, clinicID uint64, staffIDs []uint64) ([]model.StaffReservationCapability, error) {
	if m.findCapabilitiesByStaffIDsFn != nil {
		return m.findCapabilitiesByStaffIDsFn(ctx, clinicID, staffIDs)
	}
	return []model.StaffReservationCapability{}, nil
}

func (m *mockReservationStaffRepository) UpdateReservationCapabilities(ctx context.Context, clinicID, staffID uint64, typeIDs []uint64) error {
	if m.replaceCapabilitiesFn != nil {
		return m.replaceCapabilitiesFn(ctx, clinicID, staffID, typeIDs)
	}
	return nil
}

func (m *mockReservationStaffRepository) SupportsReservationType(ctx context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error) {
	if m.supportsReservationTypeFn != nil {
		return m.supportsReservationTypeFn(ctx, clinicID, staffID, reservationTypeID)
	}
	return true, nil
}

// mockTransactor は trimming_service_test.go で定義済み

func newTestReservationStaffService(repo *mockReservationStaffRepository, transactor *mockTransactor) ReservationStaffService {
	return NewReservationStaffService(repo, transactor, &reservationStaffDeleteRecorder{})
}

func TestReservationStaffService_List_ReturnsStaffs(t *testing.T) {
	tests := []struct {
		name       string
		repoStaffs []model.Staff
		repoErr    error
		wantLen    int
		wantErr    bool
	}{
		{
			name: "returns staff list",
			repoStaffs: []model.Staff{
				{ID: 1, Name: "田中医師", IsActive: true, StaffType: model.StaffTypeDoctor},
				{ID: 2, Name: "鈴木スタッフ", IsActive: true, StaffType: model.StaffTypeNurse},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:       "returns empty list when no staffs exist",
			repoStaffs: []model.Staff{},
			repoErr:    nil,
			wantLen:    0,
			wantErr:    false,
		},
		{
			name:       "propagates repository error",
			repoStaffs: nil,
			repoErr:    errors.New("db connection error"),
			wantLen:    0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationStaffRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
					return tt.repoStaffs, tt.repoErr
				},
			}
			transactor := &mockTransactor{}
			svc := newTestReservationStaffService(repo, transactor)

			staffs, err := svc.List(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, staffs, tt.wantLen)
			}
		})
	}
}

func TestReservationStaffService_GetByID_Found(t *testing.T) {
	expected := &model.Staff{
		ID:        1,
		Name:      "田中医師",
		IsActive:  true,
		StaffType: model.StaffTypeDoctor,
	}
	repo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Staff, error) {
			return expected, nil
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	staff, err := svc.GetByID(context.Background(), 1, 1)

	assert.NoError(t, err)
	assert.Equal(t, expected, staff)
}

func TestReservationStaffService_GetByID_NotFound(t *testing.T) {
	repo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Staff, error) {
			return nil, apperrors.WrapNotFound("reservation_staff", "999")
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	staff, err := svc.GetByID(context.Background(), 1, 999)

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, staff)
}

func TestReservationStaffService_Create_Success(t *testing.T) {
	input := &CreateReservationStaffInput{
		Name:               "新スタッフ",
		StaffType:          string(model.StaffTypeDoctor),
		ReservationVisible: true,
		ReservationComment: "担当コメント",
		SortOrder:          1,
	}
	repo := &mockReservationStaffRepository{
		createFn: func(_ context.Context, staff *model.Staff, clinicID uint64) error {
			assert.Equal(t, clinicID, staff.ClinicID)
			staff.ID = 10
			return nil
		},
		findExcludedReservationTypesFn: func(_ context.Context, _, _ uint64) ([]model.StaffReservationExclusion, error) {
			return []model.StaffReservationExclusion{}, nil
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	staff, excluded, err := svc.Create(context.Background(), 1, input)

	assert.NoError(t, err)
	assert.NotNil(t, staff)
	assert.NotNil(t, excluded)
	assert.Equal(t, "新スタッフ", staff.Name)
}

func TestReservationStaffService_Create_DefaultsStaffTypeWhenEmpty(t *testing.T) {
	input := &CreateReservationStaffInput{
		Name:      "空タイプスタッフ",
		StaffType: "",
		SortOrder: 2,
	}
	var createdStaff *model.Staff
	repo := &mockReservationStaffRepository{
		createFn: func(_ context.Context, staff *model.Staff, _ uint64) error {
			staff.ID = 11
			createdStaff = staff
			return nil
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	staff, excluded, err := svc.Create(context.Background(), 1, input)

	assert.NoError(t, err)
	assert.NotNil(t, staff)
	assert.NotNil(t, excluded)
	assert.Equal(t, model.StaffTypeDoctor, createdStaff.StaffType)
}

func TestReservationStaffService_Create_AlwaysSeedsFullUniverseViaEmptyExclusion(t *testing.T) {
	// TASK-021 UNIT-021-A: request excluded_type_ids removed; Create always seeds via empty inverse.
	var replacedIDs []uint64
	repo := &mockReservationStaffRepository{
		createFn: func(_ context.Context, staff *model.Staff, _ uint64) error {
			staff.ID = 12
			return nil
		},
		replaceExcludedReservationTypesFn: func(_ context.Context, _, _ uint64, courseIDs []uint64) error {
			replacedIDs = courseIDs
			return nil
		},
		findExcludedReservationTypesFn: func(_ context.Context, _, _ uint64) ([]model.StaffReservationExclusion, error) {
			return []model.StaffReservationExclusion{}, nil
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	staff, excluded, err := svc.Create(context.Background(), 1, &CreateReservationStaffInput{
		Name:      "除外なし",
		StaffType: string(model.StaffTypeNurse),
	})

	assert.NoError(t, err)
	assert.NotNil(t, staff)
	assert.NotNil(t, excluded)
	assert.Equal(t, []uint64{}, replacedIDs)
}
func TestReservationStaffService_Create_TxCreateError(t *testing.T) {
	input := &CreateReservationStaffInput{Name: "エラースタッフ"}
	repo := &mockReservationStaffRepository{
		createFn: func(_ context.Context, _ *model.Staff, _ uint64) error {
			return errors.New("db error")
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	staff, excluded, err := svc.Create(context.Background(), 1, input)

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.Nil(t, excluded)
}

func TestReservationStaffService_Create_TxUpdateExcludedError(t *testing.T) {
	// Create always seeds universe; repo failure on that seed path is surfaced.
	input := &CreateReservationStaffInput{
		Name: "除外エラー",
	}
	repo := &mockReservationStaffRepository{
		createFn: func(_ context.Context, staff *model.Staff, _ uint64) error {
			staff.ID = 13
			return nil
		},
		replaceExcludedReservationTypesFn: func(_ context.Context, _, _ uint64, _ []uint64) error {
			return errors.New("db error")
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	staff, excluded, err := svc.Create(context.Background(), 1, input)

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.Nil(t, excluded)
}

func TestReservationStaffService_Create_TransactorError(t *testing.T) {
	input := &CreateReservationStaffInput{Name: "トランザクションエラー"}
	repo := &mockReservationStaffRepository{}
	transactor := &mockTransactor{withTxErr: errors.New("tx error")}
	svc := newTestReservationStaffService(repo, transactor)

	staff, excluded, err := svc.Create(context.Background(), 1, input)

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.Nil(t, excluded)
}

func TestReservationStaffService_Create_FindExcludedAfterCreateError(t *testing.T) {
	input := &CreateReservationStaffInput{Name: "取得エラー"}
	repo := &mockReservationStaffRepository{
		createFn: func(_ context.Context, staff *model.Staff, _ uint64) error {
			staff.ID = 14
			return nil
		},
		findExcludedReservationTypesFn: func(_ context.Context, _, _ uint64) ([]model.StaffReservationExclusion, error) {
			return nil, errors.New("db error")
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	staff, excluded, err := svc.Create(context.Background(), 1, input)

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.Nil(t, excluded)
}

func TestBuildReservationStaffUpdate(t *testing.T) {
	name := "新しい名前"
	staffType := string(model.StaffTypeNurse)
	visible := false
	comment := "コメント"
	sortOrder := 5

	tests := []struct {
		name       string
		input      *UpdateReservationStaffInput
		wantFields map[string]any
	}{
		{
			name:       "all fields nil produces empty map",
			input:      &UpdateReservationStaffInput{},
			wantFields: map[string]any{},
		},
		{
			name: "all fields set",
			input: &UpdateReservationStaffInput{
				Name:               &name,
				StaffType:          &staffType,
				ReservationVisible: &visible,
				ReservationComment: &comment,
				SortOrder:          &sortOrder,
			},
			wantFields: map[string]any{
				colReservationStaffName:               name,
				colReservationStaffStaffType:          staffType,
				colReservationStaffReservationVisible: visible,
				colReservationStaffReservationComment: comment,
				colReservationStaffSortOrder:          sortOrder,
			},
		},
		{
			name: "only name set",
			input: &UpdateReservationStaffInput{
				Name: &name,
			},
			wantFields: map[string]any{
				colReservationStaffName: name,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildReservationStaffUpdate(tt.input)
			assert.Equal(t, tt.wantFields, got)
		})
	}
}

func TestReservationStaffService_Update_Success(t *testing.T) {
	name := "更新後の名前"
	repo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id, Name: "更新後の名前"}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			return nil
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	staff, excluded, err := svc.Update(context.Background(), 1, 1, &UpdateReservationStaffInput{Name: &name})

	assert.NoError(t, err)
	assert.NotNil(t, staff)
	assert.NotNil(t, excluded)
}

func TestReservationStaffService_Update_DoesNotReplaceCapabilities(t *testing.T) {
	// TASK-021 UNIT-021-A: Update has no excluded_type_ids; inverse replace must not run.
	name := "更新後の名前"
	replaceCalled := false
	repo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id, Name: name}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			return nil
		},
		replaceExcludedReservationTypesFn: func(_ context.Context, _, _ uint64, _ []uint64) error {
			replaceCalled = true
			return nil
		},
	}
	svc := newTestReservationStaffService(repo, &mockTransactor{})

	staff, excluded, err := svc.Update(context.Background(), 1, 1, &UpdateReservationStaffInput{Name: &name})

	assert.NoError(t, err)
	assert.NotNil(t, staff)
	assert.NotNil(t, excluded)
	assert.False(t, replaceCalled)
}
func TestReservationStaffService_Update_NotFound(t *testing.T) {
	name := "名前"
	repo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Staff, error) {
			return nil, apperrors.WrapNotFound("reservation_staff", "999")
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	staff, excluded, err := svc.Update(context.Background(), 1, 999, &UpdateReservationStaffInput{Name: &name})

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.Nil(t, excluded)
}

func TestReservationStaffService_Update_RepoUpdateError(t *testing.T) {
	name := "名前"
	repo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			return errors.New("db error")
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	staff, excluded, err := svc.Update(context.Background(), 1, 1, &UpdateReservationStaffInput{Name: &name})

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.Nil(t, excluded)
}

// Update has no request excluded_type_ids (UNIT-021-A); capability changes use capable-reservation-types.
func TestReservationStaffService_Update_FindByIDAfterUpdateError(t *testing.T) {
	name := "名前"
	callCount := 0
	repo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Staff, error) {
			callCount++
			if callCount == 1 {
				return &model.Staff{ID: id}, nil
			}
			return nil, errors.New("db error")
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			return nil
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	staff, excluded, err := svc.Update(context.Background(), 1, 1, &UpdateReservationStaffInput{Name: &name})

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.Nil(t, excluded)
}

func TestReservationStaffService_Update_FindExcludedAfterUpdateError(t *testing.T) {
	name := "名前"
	repo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			return nil
		},
		findExcludedReservationTypesFn: func(_ context.Context, _, _ uint64) ([]model.StaffReservationExclusion, error) {
			return nil, errors.New("db error")
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	staff, excluded, err := svc.Update(context.Background(), 1, 1, &UpdateReservationStaffInput{Name: &name})

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.Nil(t, excluded)
}

// TestReservationStaffService_Update_TransactorError は BE-refactor.md X-8 の judgment call
// （Update を WithTx で括る）の配線を検証する。修正前は Update が s.transactor.WithTx を
// 一切呼ばなかったため、withTxErr を設定しても無視されて Update は成功していた（この場合 RED）。
func TestReservationStaffService_Update_TransactorError(t *testing.T) {
	name := "トランザクションエラー"
	repo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			t.Fatal("transactor.WithTx がエラーを返す場合、repo.Update は呼ばれてはならない")
			return nil
		},
	}
	transactor := &mockTransactor{withTxErr: errors.New("tx error")}
	svc := newTestReservationStaffService(repo, transactor)

	staff, excluded, err := svc.Update(context.Background(), 1, 1, &UpdateReservationStaffInput{Name: &name})

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.Nil(t, excluded)
}

func TestReservationStaffService_PatchStatus_Success(t *testing.T) {
	repo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id, IsActive: true}, nil
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	staff, excluded, err := svc.PatchStatus(context.Background(), 1, 1, false)

	assert.NoError(t, err)
	assert.NotNil(t, staff)
	assert.NotNil(t, excluded)
}

func TestReservationStaffService_PatchStatus_NotFound(t *testing.T) {
	repo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Staff, error) {
			return nil, apperrors.WrapNotFound("reservation_staff", "999")
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	staff, excluded, err := svc.PatchStatus(context.Background(), 1, 999, false)

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.Nil(t, excluded)
}

func TestReservationStaffService_PatchStatus_UpdateError(t *testing.T) {
	repo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			return errors.New("db error")
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	staff, excluded, err := svc.PatchStatus(context.Background(), 1, 1, false)

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.Nil(t, excluded)
}

func TestReservationStaffService_PatchStatus_FindByIDAfterPatchError(t *testing.T) {
	callCount := 0
	repo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Staff, error) {
			callCount++
			if callCount == 1 {
				return &model.Staff{ID: id}, nil
			}
			return nil, errors.New("db error")
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	staff, excluded, err := svc.PatchStatus(context.Background(), 1, 1, false)

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.Nil(t, excluded)
}

func TestReservationStaffService_PatchStatus_FindExcludedAfterPatchError(t *testing.T) {
	repo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id}, nil
		},
		findExcludedReservationTypesFn: func(_ context.Context, _, _ uint64) ([]model.StaffReservationExclusion, error) {
			return nil, errors.New("db error")
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	staff, excluded, err := svc.PatchStatus(context.Background(), 1, 1, false)

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.Nil(t, excluded)
}

func TestReservationStaffService_MutationsAcquireExclusiveOwnershipFirst(t *testing.T) {
	name := "更新名"
	tests := []struct {
		name       string
		mutate     func(svc ReservationStaffService) error
		lockErr    error
		wantEvents []string
		wantErr    bool
	}{
		{
			name: "update",
			mutate: func(svc ReservationStaffService) error {
				_, _, err := svc.Update(
					context.Background(),
					1,
					2,
					&UpdateReservationStaffInput{Name: &name},
				)
				return err
			},
			wantEvents: []string{"lock", "update", "read", "exclusions"},
		},
		{
			name: "patch status",
			mutate: func(svc ReservationStaffService) error {
				_, _, err := svc.PatchStatus(context.Background(), 1, 2, false)
				return err
			},
			wantEvents: []string{"lock", "update", "read", "exclusions"},
		},
		{
			name: "ownership lock failure stops update",
			mutate: func(svc ReservationStaffService) error {
				_, _, err := svc.PatchStatus(context.Background(), 1, 2, false)
				return err
			},
			lockErr:    apperrors.WrapNotFound("reservation_staff", "2"),
			wantEvents: []string{"lock"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := make([]string, 0, 4)
			repo := &mockReservationStaffRepository{
				lockForMutationFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
					events = append(events, "lock")
					if tt.lockErr != nil {
						return nil, tt.lockErr
					}
					return &model.Staff{ID: id, ClinicID: clinicID}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
					events = append(events, "update")
					return nil
				},
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
					events = append(events, "read")
					return &model.Staff{ID: id, ClinicID: clinicID}, nil
				},
				findExcludedReservationTypesFn: func(_ context.Context, _, _ uint64) ([]model.StaffReservationExclusion, error) {
					events = append(events, "exclusions")
					return []model.StaffReservationExclusion{}, nil
				},
			}
			svc := newTestReservationStaffService(repo, &mockTransactor{})

			err := tt.mutate(svc)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantEvents, events)
		})
	}
}

func TestReservationStaffService_PatchSortOrder_Success(t *testing.T) {
	var capturedDirection string
	repo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id}, nil
		},
		swapSortOrderFn: func(_ context.Context, _, _ uint64, direction string) error {
			capturedDirection = direction
			return nil
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	err := svc.PatchSortOrder(context.Background(), 1, 1, "up")

	assert.NoError(t, err)
	assert.Equal(t, "up", capturedDirection)
}

func TestReservationStaffService_PatchSortOrder_InvalidDirection(t *testing.T) {
	repo := &mockReservationStaffRepository{}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	err := svc.PatchSortOrder(context.Background(), 1, 1, "sideways")

	assert.Error(t, err)
}

func TestReservationStaffService_PatchSortOrder_NotFound(t *testing.T) {
	repo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Staff, error) {
			return nil, apperrors.WrapNotFound("reservation_staff", "999")
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	err := svc.PatchSortOrder(context.Background(), 1, 999, "up")

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestReservationStaffService_PatchSortOrder_RepoError(t *testing.T) {
	repo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id}, nil
		},
		swapSortOrderFn: func(_ context.Context, _, _ uint64, _ string) error {
			return errors.New("db error")
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	err := svc.PatchSortOrder(context.Background(), 1, 1, "down")

	assert.Error(t, err)
}

func TestReservationStaffService_ListExcludedByStaffIDs_Success(t *testing.T) {
	repo := &mockReservationStaffRepository{
		findExcludedReservationTypesByStaffIDs: func(_ context.Context, clinicID uint64, staffIDs []uint64) ([]model.StaffReservationExclusion, error) {
			assert.Equal(t, uint64(9), clinicID)
			return []model.StaffReservationExclusion{
				{ID: 1, StaffID: staffIDs[0], ReservationTypeID: 5},
				{ID: 2, StaffID: staffIDs[0], ReservationTypeID: 6},
				{ID: 3, StaffID: staffIDs[1], ReservationTypeID: 7},
			}, nil
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	result, err := svc.ListExcludedByStaffIDs(context.Background(), 9, []uint64{1, 2})

	assert.NoError(t, err)
	assert.Len(t, result[1], 2)
	assert.Len(t, result[2], 1)
}

func TestReservationStaffService_ListExcludedByStaffIDs_Empty(t *testing.T) {
	repo := &mockReservationStaffRepository{
		findExcludedReservationTypesByStaffIDs: func(_ context.Context, _ uint64, _ []uint64) ([]model.StaffReservationExclusion, error) {
			return []model.StaffReservationExclusion{}, nil
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	result, err := svc.ListExcludedByStaffIDs(context.Background(), 9, []uint64{})

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestReservationStaffService_ListExcludedByStaffIDs_RepoError(t *testing.T) {
	repo := &mockReservationStaffRepository{
		findExcludedReservationTypesByStaffIDs: func(_ context.Context, _ uint64, _ []uint64) ([]model.StaffReservationExclusion, error) {
			return nil, errors.New("db error")
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	result, err := svc.ListExcludedByStaffIDs(context.Background(), 9, []uint64{1})

	assert.Error(t, err)
	assert.Nil(t, result)
}
