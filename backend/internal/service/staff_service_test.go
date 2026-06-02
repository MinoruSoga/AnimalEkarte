package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// mockStaffRepository は StaffRepository のテスト用モック実装
type mockStaffRepository struct {
	findAllFn           func(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error)
	findByIDFn          func(ctx context.Context, id uint64) (*model.Staff, error)
	findByAccountIDFn   func(ctx context.Context, accountID uint64) (*model.Staff, error)
	createFn            func(ctx context.Context, staff *model.Staff) error
	updateFn            func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn            func(ctx context.Context, clinicID, id uint64) error
	reorderErr          error
	countBlockingRefsFn func(ctx context.Context, clinicID, staffID uint64) ([]repository.StaffDependencyCount, error)
}

func (m *mockStaffRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error) {
	return m.findAllFn(ctx, clinicID, page, limit)
}

func (m *mockStaffRepository) FindByID(ctx context.Context, id uint64) (*model.Staff, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockStaffRepository) FindByAccountID(ctx context.Context, accountID uint64) (*model.Staff, error) {
	if m.findByAccountIDFn != nil {
		return m.findByAccountIDFn(ctx, accountID)
	}
	return nil, apperrors.WrapNotFound("staff", "account_id")
}

func (m *mockStaffRepository) Create(ctx context.Context, staff *model.Staff) error {
	return m.createFn(ctx, staff)
}

func (m *mockStaffRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return m.updateFn(ctx, clinicID, id, fields)
}

func (m *mockStaffRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockStaffRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}

func (m *mockStaffRepository) CountBlockingReferencesByStaffID(ctx context.Context, clinicID, staffID uint64) ([]repository.StaffDependencyCount, error) {
	if m.countBlockingRefsFn == nil {
		return nil, nil
	}
	return m.countBlockingRefsFn(ctx, clinicID, staffID)
}

// mockReservationForStaff は Staff テストで使用する ReservationQueryRepository のスタブ
type mockReservationForStaff struct {
	existsByStaffIDFn func(ctx context.Context, clinicID, staffID uint64) (bool, error)
}

func (m *mockReservationForStaff) ExistsByReservationTypeID(_ context.Context, _, _ uint64) (bool, error) {
	return false, nil
}
func (m *mockReservationForStaff) ExistsByStaffID(ctx context.Context, clinicID, staffID uint64) (bool, error) {
	if m.existsByStaffIDFn != nil {
		return m.existsByStaffIDFn(ctx, clinicID, staffID)
	}
	return false, nil
}
func (m *mockReservationForStaff) CountMedicalRecordsByReservationID(_ context.Context, _ uint64) (int64, error) {
	return 0, nil
}
func (m *mockReservationForStaff) CountByCustomerAndDateRange(_ context.Context, _, _ uint64, _, _ time.Time) (int64, error) {
	return 0, nil
}
func (m *mockReservationForStaff) CountByDateAndSource(_ context.Context, _ uint64, _ time.Time, _ model.ReservationSource) (int64, error) {
	return 0, nil
}
func (m *mockReservationForStaff) FindAllByCategory(_ context.Context, _ uint64, _ model.ReservationTypeCategory, _, _ *uint64, _, _ *string, _, _ int) ([]model.Reservation, int64, error) {
	return nil, 0, nil
}

func (m *mockReservationForStaff) FindNoShowCandidates(_ context.Context, _ uint64) ([]model.Reservation, error) {
	return nil, nil
}
func (m *mockReservationForStaff) HasReservationByOwnerInRange(_ context.Context, _, _ uint64, _, _ time.Time) (bool, error) {
	return false, nil
}

// mockShiftEntryForStaff は Staff テストで使用する ShiftEntryRepository のスタブ
type mockShiftEntryForStaff struct {
	existsByStaffIDFn func(ctx context.Context, clinicID, staffID uint64) (bool, error)
}

func (m *mockShiftEntryForStaff) FindAll(_ context.Context, _ uint64, _ repository.ShiftEntryFilter) ([]model.ShiftEntry, error) {
	return nil, nil
}
func (m *mockShiftEntryForStaff) FindByID(_ context.Context, _, _ uint64) (*model.ShiftEntry, error) {
	return nil, nil
}
func (m *mockShiftEntryForStaff) Create(_ context.Context, _ *model.ShiftEntry) error {
	return nil
}
func (m *mockShiftEntryForStaff) Update(_ context.Context, _, _ uint64, _ map[string]any) error {
	return nil
}
func (m *mockShiftEntryForStaff) Delete(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *mockShiftEntryForStaff) ReplaceBreaks(_ context.Context, _ uint64, _ []model.ShiftEntryBreak) error {
	return nil
}
func (m *mockShiftEntryForStaff) ExistsByStaffID(ctx context.Context, clinicID, staffID uint64) (bool, error) {
	if m.existsByStaffIDFn != nil {
		return m.existsByStaffIDFn(ctx, clinicID, staffID)
	}
	return false, nil
}

func (m *mockShiftEntryForStaff) FindOnDutyStaffs(_ context.Context, _ uint64, _ time.Time) ([]model.Staff, error) {
	return nil, nil
}

// mockAccountForStaff は Staff テストで使用する AccountRepository のスタブ
type mockAccountForStaff struct {
	findByEmailFn  func(ctx context.Context, email string) (*model.Account, error)
	getByIDFn      func(ctx context.Context, id uint64) (*model.Account, error)
	createFn       func(ctx context.Context, account *model.Account) error
	updateFieldsFn func(ctx context.Context, id uint64, fields map[string]any) error
}

func (m *mockAccountForStaff) FindByID(ctx context.Context, id uint64) (*model.Account, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return &model.Account{ID: id}, nil
}
func (m *mockAccountForStaff) FindByEmail(ctx context.Context, email string) (*model.Account, error) {
	if m.findByEmailFn != nil {
		return m.findByEmailFn(ctx, email)
	}
	return nil, apperrors.WrapNotFound("account", email)
}
func (m *mockAccountForStaff) Create(ctx context.Context, account *model.Account) error {
	if m.createFn != nil {
		return m.createFn(ctx, account)
	}
	account.ID = 1
	return nil
}
func (m *mockAccountForStaff) Update(ctx context.Context, id uint64, fields map[string]any) error {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, id, fields)
	}
	return nil
}
func (m *mockAccountForStaff) Delete(_ context.Context, _ uint64) error { return nil }

// mockAssignmentForStaff は Staff テストで使用する StaffClinicAssignmentRepository のスタブ
type mockAssignmentForStaff struct {
	deleteByStaffIDFn func(ctx context.Context, staffID uint64) error
	createFn          func(ctx context.Context, a *model.StaffClinicAssignment) error
}

func (m *mockAssignmentForStaff) FindByStaffID(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
	return nil, nil
}
func (m *mockAssignmentForStaff) CountByStaffAndClinic(_ context.Context, _, _ uint64) (int64, error) {
	return 1, nil
}
func (m *mockAssignmentForStaff) FindByClinicID(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
	return nil, nil
}
func (m *mockAssignmentForStaff) Create(ctx context.Context, a *model.StaffClinicAssignment) error {
	if m.createFn != nil {
		return m.createFn(ctx, a)
	}
	return nil
}
func (m *mockAssignmentForStaff) Delete(ctx context.Context, staffID uint64) error {
	if m.deleteByStaffIDFn != nil {
		return m.deleteByStaffIDFn(ctx, staffID)
	}
	return nil
}

// mockPermissionGroupForStaff は Staff テストで使用する PermissionGroupRepository のスタブ
type mockPermissionGroupForStaff struct{}

func (m *mockPermissionGroupForStaff) FindAll(_ context.Context, _ uint64) ([]model.PermissionGroup, error) {
	return nil, nil
}
func (m *mockPermissionGroupForStaff) FindByID(_ context.Context, _, _ uint64) (*model.PermissionGroup, error) {
	return nil, nil
}
func (m *mockPermissionGroupForStaff) Create(_ context.Context, _ *model.PermissionGroup) error {
	return nil
}
func (m *mockPermissionGroupForStaff) Update(_ context.Context, _, _ uint64, _ map[string]any) (*model.PermissionGroup, error) {
	return &model.PermissionGroup{}, nil
}
func (m *mockPermissionGroupForStaff) Delete(_ context.Context, _, _ uint64) error { return nil }
func (m *mockPermissionGroupForStaff) UpdateRules(_ context.Context, _ uint64, _ []model.PermissionGroupRule) error {
	return nil
}
func (m *mockPermissionGroupForStaff) CountUsageByGroupID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *mockPermissionGroupForStaff) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return nil
}
func (m *mockPermissionGroupForStaff) FindAllEffectivePermissionsByStaffID(_ context.Context, _, _ uint64) ([]model.PermissionGroupRule, error) {
	return nil, nil
}
func (m *mockPermissionGroupForStaff) FindAllGroupIDsByStaffID(_ context.Context, _ uint64) ([]uint64, error) {
	return nil, nil
}
func (m *mockPermissionGroupForStaff) UpdateStaffGroups(_ context.Context, _ uint64, _ []uint64) error {
	return nil
}

// mockResStaffForStaff は Staff テストで使用する ReservationStaffRepository のスタブ
type mockResStaffForStaff struct{}

func (m *mockResStaffForStaff) FindAll(_ context.Context, _ uint64) ([]model.Staff, error) {
	return nil, nil
}
func (m *mockResStaffForStaff) FindByID(_ context.Context, _, _ uint64) (*model.Staff, error) {
	return nil, nil
}
func (m *mockResStaffForStaff) Create(_ context.Context, _ *model.Staff, _ uint64) error { return nil }
func (m *mockResStaffForStaff) Update(_ context.Context, _, _ uint64, _ map[string]any) error {
	return nil
}
func (m *mockResStaffForStaff) Delete(_ context.Context, _, _ uint64) error { return nil }
func (m *mockResStaffForStaff) CountUsageByStaffID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *mockResStaffForStaff) UpdateSortOrder(_ context.Context, _, _ uint64, _ string) error {
	return nil
}
func (m *mockResStaffForStaff) FindAllExcludedReservationTypes(_ context.Context, _ uint64) ([]model.StaffReservationExclusion, error) {
	return nil, nil
}
func (m *mockResStaffForStaff) FindAllExcludedReservationTypesByStaffIDs(_ context.Context, _ []uint64) ([]model.StaffReservationExclusion, error) {
	return nil, nil
}
func (m *mockResStaffForStaff) UpdateExcludedReservationTypes(_ context.Context, _ uint64, _ []uint64) error {
	return nil
}
func (m *mockResStaffForStaff) FindAllReservationCapabilities(_ context.Context, _, _ uint64) ([]model.StaffReservationCapability, error) {
	return nil, nil
}
func (m *mockResStaffForStaff) FindAllReservationCapabilitiesByStaffIDs(_ context.Context, _ uint64, _ []uint64) ([]model.StaffReservationCapability, error) {
	return nil, nil
}
func (m *mockResStaffForStaff) UpdateReservationCapabilities(_ context.Context, _, _ uint64, _ []uint64) error {
	return nil
}
func (m *mockResStaffForStaff) SupportsReservationType(_ context.Context, _, _, _ uint64) (bool, error) {
	return true, nil
}

// noopTransactor はテスト用のトランザクションモック。fn を直接実行するだけ。
type noopTransactor struct{}

func (noopTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx) //nolint:contextcheck // テスト用: 親 context をそのまま伝播
}

func newTestStaffService(repo *mockStaffRepository) StaffService {
	return NewStaffService(repo, &mockAccountForStaff{}, &mockAssignmentForStaff{}, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupForStaff{}, &mockResStaffForStaff{}, noopTransactor{})
}

func TestStaffService_List(t *testing.T) {
	tests := []struct {
		name       string
		clinicID   uint64
		repoStaffs []model.Staff
		repoTotal  int64
		repoErr    error
		wantLen    int
		wantTotal  int64
		wantErr    bool
	}{
		{
			name:     "returns all staffs",
			clinicID: 1,
			repoStaffs: []model.Staff{
				{ID: 1, Name: "山田 太郎"},
				{ID: 2, Name: "鈴木 花子"},
			},
			repoTotal: 2,
			repoErr:   nil,
			wantLen:   2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:       "returns empty list when no staffs exist",
			clinicID:   1,
			repoStaffs: []model.Staff{},
			repoTotal:  0,
			repoErr:    nil,
			wantLen:    0,
			wantTotal:  0,
			wantErr:    false,
		},
		{
			name:       "propagates repository error",
			clinicID:   1,
			repoStaffs: nil,
			repoTotal:  0,
			repoErr:    errors.New("db connection error"),
			wantLen:    0,
			wantTotal:  0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capturedPage := 0
			capturedLimit := 0
			repo := &mockStaffRepository{
				findAllFn: func(_ context.Context, _ uint64, page, limit int) ([]model.Staff, int64, error) {
					capturedPage = page
					capturedLimit = limit
					return tt.repoStaffs, tt.repoTotal, tt.repoErr
				},
			}
			svc := newTestStaffService(repo)

			staffs, total, err := svc.List(context.Background(), tt.clinicID, 1, 20)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, int64(0), total)
			} else {
				assert.NoError(t, err)
				assert.Len(t, staffs, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
				assert.Equal(t, 1, capturedPage)
				assert.Equal(t, 20, capturedLimit)
			}
		})
	}
}

func TestStaffService_GetByID(t *testing.T) {
	tests := []struct {
		name      string
		id        uint64
		repoStaff *model.Staff
		repoErr   error
		wantStaff *model.Staff
		wantErr   error
	}{
		{
			name:      "returns staff when found",
			id:        10,
			repoStaff: &model.Staff{ID: 10, Name: "山田 太郎"},
			repoErr:   nil,
			wantStaff: &model.Staff{ID: 10, Name: "山田 太郎"},
			wantErr:   nil,
		},
		{
			name:      "returns not found error when staff does not exist",
			id:        999,
			repoStaff: nil,
			repoErr:   apperrors.WrapNotFound("staff", "999"),
			wantStaff: nil,
			wantErr:   apperrors.ErrNotFound,
		},
		{
			name:      "returns error on repository failure",
			id:        10,
			repoStaff: nil,
			repoErr:   errors.New("db error"),
			wantStaff: nil,
			wantErr:   errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockStaffRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
					return tt.repoStaff, tt.repoErr
				},
			}
			svc := newTestStaffService(repo)

			staff, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				assert.Error(t, err)
				if errors.Is(tt.wantErr, apperrors.ErrNotFound) {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantStaff, staff)
			}
		})
	}
}

func TestStaffService_GetByID_NotFound(t *testing.T) {
	repo := &mockStaffRepository{
		findByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
			return nil, apperrors.WrapNotFound("staff", "999")
		},
	}
	svc := newTestStaffService(repo)

	staff, err := svc.GetByID(context.Background(), 999)

	assert.Nil(t, staff)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

// TestStaffService_Create_Success はスタッフ作成の正常ケースをテストする。
func TestStaffService_Create_Success(t *testing.T) {
	repo := &mockStaffRepository{
		createFn: func(_ context.Context, staff *model.Staff) error {
			// IDをシミュレート
			staff.ID = 1
			return nil
		},
	}
	svc := newTestStaffService(repo)

	input := &CreateStaffInput{
		Name: "新規 スタッフ",
	}

	staff, err := svc.Create(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, staff)
	assert.Equal(t, "新規 スタッフ", staff.Name)
	assert.True(t, staff.IsActive)
}

func TestStaffService_Create_RepositoryError(t *testing.T) {
	repo := &mockStaffRepository{
		createFn: func(_ context.Context, _ *model.Staff) error {
			return errors.New("db connection error")
		},
	}
	svc := newTestStaffService(repo)

	input := &CreateStaffInput{
		Name: "エラー スタッフ",
	}

	staff, err := svc.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, staff)
}

func TestStaffService_Create_DuplicateName(t *testing.T) {
	repo := &mockStaffRepository{
		createFn: func(_ context.Context, _ *model.Staff) error {
			return apperrors.WrapAlreadyExists("staff", "existing@example.com")
		},
	}
	svc := newTestStaffService(repo)

	input := &CreateStaffInput{
		Name: "重複 スタッフ",
	}

	staff, err := svc.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.True(t, apperrors.IsAlreadyExists(err))
}

func TestStaffService_Update(t *testing.T) {
	name := "更新後 スタッフ"
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		input    *UpdateStaffInput
		repoErr  error
		wantErr  bool
		wantNF   bool
	}{
		{
			name:     "updates staff successfully",
			clinicID: 1,
			id:       1,
			input:    &UpdateStaffInput{Name: &name},
			repoErr:  nil,
			wantErr:  false,
		},
		{
			name:     "returns not found error when staff does not exist",
			clinicID: 1,
			id:       999,
			input:    &UpdateStaffInput{Name: &name},
			repoErr:  apperrors.WrapNotFound("staff", "999"),
			wantErr:  true,
			wantNF:   true,
		},
		{
			name:     "returns error when no field provided",
			clinicID: 1,
			id:       1,
			input:    &UpdateStaffInput{},
			repoErr:  nil,
			wantErr:  true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       1,
			input:    &UpdateStaffInput{Name: &name},
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockStaffRepository{
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
					return tt.repoErr
				},
				findByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
					return &model.Staff{ID: id}, nil
				},
			}
			svc := newTestStaffService(repo)

			staff, err := svc.Update(context.Background(), tt.clinicID, tt.id, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, staff)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, staff)
			}
		})
	}
}

func TestStaffService_Reorder(t *testing.T) {
	tests := []struct {
		name             string
		ids              []uint64
		repoErr          error
		wantErr          bool
		wantInvalidInput bool
	}{
		{
			name:    "reorders successfully",
			ids:     []uint64{3, 1, 2},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:             "returns invalid input when ids is empty",
			ids:              []uint64{},
			wantErr:          true,
			wantInvalidInput: true,
		},
		{
			name:    "propagates repository error",
			ids:     []uint64{1, 2},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockStaffRepository{reorderErr: tt.repoErr}
			svc := newTestStaffService(repo)

			err := svc.Reorder(context.Background(), 1, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantInvalidInput {
					assert.True(t, apperrors.IsInvalidInput(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStaffService_Delete(t *testing.T) {
	tests := []struct {
		name                string
		clinicID            uint64
		id                  uint64
		findByIDErr         error
		reservationExists   bool
		shiftExists         bool
		checkReservationErr error
		checkShiftErr       error
		blockingRefs        []repository.StaffDependencyCount
		blockingErr         error
		repoErr             error
		wantErr             bool
		wantNF              bool
		wantConflict        bool
	}{
		{
			name:                "deletes staff successfully when no dependencies exist",
			clinicID:            1,
			id:                  10,
			reservationExists:   false,
			shiftExists:         false,
			checkReservationErr: nil,
			checkShiftErr:       nil,
			repoErr:             nil,
			wantErr:             false,
		},
		{
			name:        "returns not found error when staff does not exist",
			clinicID:    1,
			id:          999,
			findByIDErr: apperrors.WrapNotFound("staff", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name:                "returns conflict error when staff has reservations",
			clinicID:            1,
			id:                  10,
			reservationExists:   true,
			shiftExists:         false,
			checkReservationErr: nil,
			checkShiftErr:       nil,
			repoErr:             nil,
			wantErr:             true,
			wantConflict:        true,
		},
		{
			name:                "returns conflict error when staff has shift entries",
			clinicID:            1,
			id:                  10,
			reservationExists:   false,
			shiftExists:         true,
			checkReservationErr: nil,
			checkShiftErr:       nil,
			repoErr:             nil,
			wantErr:             true,
			wantConflict:        true,
		},
		{
			name:                "returns conflict error when staff has both reservations and shifts",
			clinicID:            1,
			id:                  10,
			reservationExists:   true,
			shiftExists:         true,
			checkReservationErr: nil,
			checkShiftErr:       nil,
			repoErr:             nil,
			wantErr:             true,
			wantConflict:        true,
		},
		{
			name:                "returns error when reservation check fails",
			clinicID:            1,
			id:                  10,
			reservationExists:   false,
			shiftExists:         false,
			checkReservationErr: errors.New("db error"),
			checkShiftErr:       nil,
			repoErr:             nil,
			wantErr:             true,
		},
		{
			name:                "returns error when shift check fails",
			clinicID:            1,
			id:                  10,
			reservationExists:   false,
			shiftExists:         false,
			checkReservationErr: nil,
			checkShiftErr:       errors.New("db error"),
			repoErr:             nil,
			wantErr:             true,
		},
		{
			name:         "returns conflict error when staff has blocking dependencies",
			clinicID:     1,
			id:           10,
			blockingRefs: []repository.StaffDependencyCount{{Label: "カルテ追記", Count: 1}},
			wantErr:      true,
			wantConflict: true,
		},
		{
			name:        "returns error when dependency check fails",
			clinicID:    1,
			id:          10,
			blockingErr: errors.New("db error"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockStaffRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.Staff{ID: tt.id}, nil
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
				countBlockingRefsFn: func(_ context.Context, _, _ uint64) ([]repository.StaffDependencyCount, error) {
					return tt.blockingRefs, tt.blockingErr
				},
			}
			reservationRepo := &mockReservationForStaff{
				existsByStaffIDFn: func(_ context.Context, _, _ uint64) (bool, error) {
					return tt.reservationExists, tt.checkReservationErr
				},
			}
			shiftRepo := &mockShiftEntryForStaff{
				existsByStaffIDFn: func(_ context.Context, _, _ uint64) (bool, error) {
					return tt.shiftExists, tt.checkShiftErr
				},
			}
			svc := NewStaffService(repo, &mockAccountForStaff{}, &mockAssignmentForStaff{}, reservationRepo, shiftRepo, &mockPermissionGroupForStaff{}, &mockResStaffForStaff{}, noopTransactor{})

			err := svc.Delete(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
