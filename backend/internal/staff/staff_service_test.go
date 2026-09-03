package staff

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockStaffRepository は StaffRepository のテスト用モック実装
type mockStaffRepository struct {
	findAllFn           func(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error)
	findByIDFn          func(ctx context.Context, id uint64) (*model.Staff, error)
	lockForUpdateFn     func(ctx context.Context, id uint64) (*model.Staff, error)
	lockInClinicFn      func(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	lockForShareFn      func(ctx context.Context, id uint64) (*model.Staff, error)
	findByAccountIDFn   func(ctx context.Context, accountID uint64) (*model.Staff, error)
	createFn            func(ctx context.Context, staff *model.Staff) error
	updateFn            func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	updatePrimaryFn     func(ctx context.Context, id, clinicID uint64) error
	deleteFn            func(ctx context.Context, clinicID, id uint64) error
	reorderErr          error
	countBlockingRefsFn func(ctx context.Context, clinicID, staffID uint64) ([]StaffDependencyCount, error)
}

func (m *mockStaffRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error) {
	return m.findAllFn(ctx, clinicID, page, limit)
}

func (m *mockStaffRepository) FindByID(ctx context.Context, id uint64) (*model.Staff, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockStaffRepository) FindByIDInClinic(
	ctx context.Context,
	clinicID, id uint64,
) (*model.Staff, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return &model.Staff{ID: id, ClinicID: clinicID}, nil
}

func (m *mockStaffRepository) LockActiveByIDForUpdate(ctx context.Context, id uint64) (*model.Staff, error) {
	if m.lockForUpdateFn != nil {
		return m.lockForUpdateFn(ctx, id)
	}
	return &model.Staff{ID: id}, nil
}

func (m *mockStaffRepository) LockActiveByIDForUpdateInClinic(
	ctx context.Context,
	clinicID, id uint64,
) (*model.Staff, error) {
	if m.lockInClinicFn != nil {
		return m.lockInClinicFn(ctx, clinicID, id)
	}
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return &model.Staff{ID: id}, nil
}

func (m *mockStaffRepository) LockActiveByIDForShare(ctx context.Context, id uint64) (*model.Staff, error) {
	if m.lockForShareFn != nil {
		return m.lockForShareFn(ctx, id)
	}
	return &model.Staff{ID: id}, nil
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

func (m *mockStaffRepository) UpdatePrimaryClinicID(ctx context.Context, id, clinicID uint64) error {
	if m.updatePrimaryFn != nil {
		return m.updatePrimaryFn(ctx, id, clinicID)
	}
	return nil
}

func (m *mockStaffRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockStaffRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}

func (m *mockStaffRepository) CountBlockingReferencesByStaffID(ctx context.Context, clinicID, staffID uint64) ([]StaffDependencyCount, error) {
	if m.countBlockingRefsFn == nil {
		return nil, nil
	}
	return m.countBlockingRefsFn(ctx, clinicID, staffID)
}

func (m *mockStaffRepository) IsActiveSystemAdminStaff(_ context.Context, _ uint64) (bool, error) {
	return false, nil
}

func (m *mockStaffRepository) CountActiveSystemAdminStaff(_ context.Context) (int64, error) {
	return 0, nil
}

// 予約用途 write（ADR-006 論点#1 案A）は staff service からは呼ばれない no-op スタブ。
func (m *mockStaffRepository) CreateForReservation(_ context.Context, _ *model.Staff, _ uint64) error {
	return nil
}

func (m *mockStaffRepository) UpdateForReservation(_ context.Context, _, _ uint64, _ ReservationStaffUpdate) error {
	return nil
}

func (m *mockStaffRepository) DeleteForReservation(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockStaffRepository) SwapSortOrderForReservation(_ context.Context, _, _ uint64, _ string) error {
	return nil
}

// mockReservationForStaff は Staff テストで使用する ReservationQueryRepository のスタブ
type mockReservationForStaff struct {
	existsByStaffIDFn        func(ctx context.Context, clinicID, staffID uint64) (bool, error)
	findClinicIDsByStaffIDFn func(ctx context.Context, clinicIDs []uint64, staffID uint64) ([]uint64, error)
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
func (m *mockReservationForStaff) FindClinicIDsByStaffID(
	ctx context.Context,
	clinicIDs []uint64,
	staffID uint64,
) ([]uint64, error) {
	if m.findClinicIDsByStaffIDFn != nil {
		return m.findClinicIDsByStaffIDFn(ctx, clinicIDs, staffID)
	}
	result := make([]uint64, 0)
	for _, clinicID := range clinicIDs {
		exists, err := m.ExistsByStaffID(ctx, clinicID, staffID)
		if err != nil {
			return nil, err
		}
		if exists {
			result = append(result, clinicID)
		}
	}
	return result, nil
}
func (m *mockReservationForStaff) CountMedicalRecordsByReservationID(_ context.Context, _, _ uint64) (int64, error) {
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

func (m *mockReservationForStaff) AssertOwnerInClinic(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockReservationForStaff) FindPetOwnerInClinic(_ context.Context, _, _ uint64) (uint64, error) {
	return 0, nil
}

func (m *mockReservationForStaff) FindPetByIDInClinic(_ context.Context, _, petID uint64) (*model.Pet, error) {
	return &model.Pet{ID: petID, Status: model.PetStatusAlive}, nil
}

func (m *mockReservationForStaff) AssertLineCustomerInClinic(_ context.Context, _, _ uint64) error {
	return nil
}

// mockShiftEntryForStaff は Staff テストで使用する ShiftEntryRepository のスタブ
type mockShiftEntryForStaff struct {
	existsByStaffIDFn        func(ctx context.Context, clinicID, staffID uint64) (bool, error)
	findClinicIDsByStaffIDFn func(ctx context.Context, clinicIDs []uint64, staffID uint64) ([]uint64, error)
}

func (m *mockShiftEntryForStaff) FindAll(_ context.Context, _ uint64, _ ShiftEntryFilter) ([]model.ShiftEntry, error) {
	return nil, nil
}
func (m *mockShiftEntryForStaff) FindByID(_ context.Context, _, _ uint64) (*model.ShiftEntry, error) {
	return nil, nil
}
func (m *mockShiftEntryForStaff) LockActiveByIDForUpdate(
	_ context.Context,
	clinicID, id uint64,
) (*model.ShiftEntry, error) {
	return &model.ShiftEntry{ID: id, ClinicID: clinicID}, nil
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
func (m *mockShiftEntryForStaff) FindClinicIDsByStaffID(
	ctx context.Context,
	clinicIDs []uint64,
	staffID uint64,
) ([]uint64, error) {
	if m.findClinicIDsByStaffIDFn != nil {
		return m.findClinicIDsByStaffIDFn(ctx, clinicIDs, staffID)
	}
	result := make([]uint64, 0)
	for _, clinicID := range clinicIDs {
		exists, err := m.ExistsByStaffID(ctx, clinicID, staffID)
		if err != nil {
			return nil, err
		}
		if exists {
			result = append(result, clinicID)
		}
	}
	return result, nil
}

func (m *mockShiftEntryForStaff) FindOnDutyStaffs(_ context.Context, _ uint64, _ time.Time) ([]model.Staff, error) {
	return nil, nil
}

func (m *mockShiftEntryForStaff) SaveByStaffDate(
	_ context.Context,
	_ uint64,
	_ *model.ShiftEntry,
	_ []model.ShiftEntryBreak,
) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error) {
	return nil, nil, false, nil
}

func (m *mockShiftEntryForStaff) DeleteByStaffDate(_ context.Context, _, _ uint64, _ time.Time) error {
	return nil
}

// mockAccountForStaff は Staff テストで使用する AccountRepository のスタブ
type mockAccountForStaff struct {
	findByEmailFn        func(ctx context.Context, email string) (*model.Account, error)
	getByIDFn            func(ctx context.Context, id uint64) (*model.Account, error)
	createFn             func(ctx context.Context, account *model.Account) error
	updatePasswordHashFn func(
		ctx context.Context,
		id uint64,
		newHash string,
		updatedAt time.Time,
	) error
	deletePasswordResetTokensFn func(ctx context.Context, id uint64) error
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
func (m *mockAccountForStaff) UpdatePasswordHash(
	ctx context.Context,
	id uint64,
	newHash string,
	updatedAt time.Time,
) error {
	if m.updatePasswordHashFn != nil {
		return m.updatePasswordHashFn(ctx, id, newHash, updatedAt)
	}
	return nil
}
func (m *mockAccountForStaff) DeletePasswordResetTokens(
	ctx context.Context,
	id uint64,
) error {
	if m.deletePasswordResetTokensFn != nil {
		return m.deletePasswordResetTokensFn(ctx, id)
	}
	return nil
}
func (m *mockAccountForStaff) Delete(_ context.Context, _ uint64) error { return nil }

// mockAssignmentForStaff は Staff テストで使用する StaffClinicAssignmentRepository のスタブ
type mockAssignmentForStaff struct {
	deleteByStaffIDFn   func(ctx context.Context, staffID uint64) error
	deleteByClinicIDsFn func(ctx context.Context, staffID uint64, clinicIDs []uint64) error
	createFn            func(ctx context.Context, a *model.StaffClinicAssignment) error
	lockActiveFn        func(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error)
	restoreOrCreateFn   func(ctx context.Context, a *model.StaffClinicAssignment) error
}

func (m *mockAssignmentForStaff) FindByStaffID(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
	return nil, nil
}
func (m *mockAssignmentForStaff) FindByStaffAndClinic(
	_ context.Context,
	staffID, clinicID uint64,
) (*model.StaffClinicAssignment, error) {
	return &model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID}, nil
}
func (m *mockAssignmentForStaff) CountByStaffAndClinic(_ context.Context, _, _ uint64) (int64, error) {
	return 1, nil
}
func (m *mockAssignmentForStaff) LockActiveByStaffAndClinic(
	_ context.Context,
	staffID, clinicID uint64,
) (*model.StaffClinicAssignment, error) {
	return &model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID}, nil
}
func (m *mockAssignmentForStaff) LockActiveByStaff(
	ctx context.Context,
	staffID uint64,
) ([]model.StaffClinicAssignment, error) {
	if m.lockActiveFn != nil {
		return m.lockActiveFn(ctx, staffID)
	}
	return nil, nil
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
func (m *mockAssignmentForStaff) RestoreOrCreate(ctx context.Context, a *model.StaffClinicAssignment) error {
	if m.restoreOrCreateFn != nil {
		return m.restoreOrCreateFn(ctx, a)
	}
	return m.Create(ctx, a)
}
func (m *mockAssignmentForStaff) Delete(ctx context.Context, staffID uint64) error {
	if m.deleteByStaffIDFn != nil {
		return m.deleteByStaffIDFn(ctx, staffID)
	}
	return nil
}
func (m *mockAssignmentForStaff) DeleteByStaffAndClinicIDs(
	ctx context.Context,
	staffID uint64,
	clinicIDs []uint64,
) error {
	if m.deleteByClinicIDsFn != nil {
		return m.deleteByClinicIDsFn(ctx, staffID, clinicIDs)
	}
	if len(clinicIDs) == 0 {
		return nil
	}
	if m.deleteByStaffIDFn != nil {
		return m.deleteByStaffIDFn(ctx, staffID)
	}
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
func (m *mockResStaffForStaff) FindAllExcludedReservationTypes(
	_ context.Context,
	_, _ uint64,
) ([]model.StaffReservationExclusion, error) {
	return nil, nil
}
func (m *mockResStaffForStaff) FindAllExcludedReservationTypesByStaffIDs(_ context.Context, _ uint64, _ []uint64) ([]model.StaffReservationExclusion, error) {
	return nil, nil
}
func (m *mockResStaffForStaff) UpdateExcludedReservationTypes(_ context.Context, _, _ uint64, _ []uint64) error {
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
	return newTestStaffServiceWithAssignmentRepo(repo, &mockAssignmentForStaff{})
}

func newTestStaffServiceWithAssignmentRepo(
	repo *mockStaffRepository,
	assignmentRepo StaffClinicAssignmentRepository,
) StaffService {
	return NewStaffService(repo, &mockAccountForStaff{}, assignmentRepo, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, nil, noopTransactor{})
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
		ClinicID: 1,
		Name:     "新規 スタッフ",
	}

	staff, err := svc.Create(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, staff)
	assert.Equal(t, "新規 スタッフ", staff.Name)
	assert.Equal(t, uint64(1), staff.ClinicID)
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
		ClinicID: 1,
		Name:     "エラー スタッフ",
	}

	staff, err := svc.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, staff)
}

func TestStaffService_Create_RequiresClinicID(t *testing.T) {
	repo := &mockStaffRepository{
		createFn: func(_ context.Context, _ *model.Staff) error {
			t.Fatal("repository must not be called when clinic_id is missing")
			return nil
		},
	}
	svc := newTestStaffService(repo)

	staff, err := svc.Create(context.Background(), &CreateStaffInput{Name: "clinic なし"})

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestStaffService_Create_DuplicateName(t *testing.T) {
	repo := &mockStaffRepository{
		createFn: func(_ context.Context, _ *model.Staff) error {
			return apperrors.WrapAlreadyExists("staff", "existing@example.com")
		},
	}
	svc := newTestStaffService(repo)

	input := &CreateStaffInput{
		ClinicID: 1,
		Name:     "重複 スタッフ",
	}

	staff, err := svc.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.True(t, apperrors.IsAlreadyExists(err))
}

func TestStaffService_SetClinicAssignments_UpdatesPrimaryClinicID(t *testing.T) {
	var created []model.StaffClinicAssignment
	var primaryClinicID uint64
	repo := &mockStaffRepository{
		updatePrimaryFn: func(_ context.Context, id, clinicID uint64) error {
			assert.Equal(t, uint64(10), id)
			primaryClinicID = clinicID
			return nil
		},
	}
	assignmentRepo := &mockAssignmentForStaff{
		createFn: func(_ context.Context, a *model.StaffClinicAssignment) error {
			created = append(created, *a)
			return nil
		},
	}
	svc := NewStaffService(repo, &mockAccountForStaff{}, assignmentRepo, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, existingClinicLookupForStaffAssignments(), noopTransactor{})

	err := svc.SetClinicAssignments(context.Background(), &SetClinicAssignmentsInput{
		StaffID:             10,
		ClinicIDs:           []uint64{2, 4},
		AuthorizedClinicIDs: []uint64{2, 4},
	})

	assert.NoError(t, err)
	if assert.Len(t, created, 2) {
		assert.True(t, created[0].IsMain)
		assert.False(t, created[1].IsMain)
	}
	assert.Equal(t, uint64(2), primaryClinicID)
}

func TestStaffService_SetClinicAssignments_RequiresClinicIDs(t *testing.T) {
	repo := &mockStaffRepository{
		updatePrimaryFn: func(_ context.Context, _, _ uint64) error {
			t.Fatal("repository must not be called when clinic_ids is empty")
			return nil
		},
	}
	svc := NewStaffService(repo, &mockAccountForStaff{}, &mockAssignmentForStaff{}, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, nil, noopTransactor{})

	err := svc.SetClinicAssignments(context.Background(), &SetClinicAssignmentsInput{
		StaffID:             10,
		ClinicIDs:           nil,
		AuthorizedClinicIDs: []uint64{1},
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}
