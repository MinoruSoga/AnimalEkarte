package staff

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- local minimal mocks (scoped to this file to avoid touching concurrently-edited
//      staff_service_test.go / staff_service_account_test.go / staff_service_permissions_test.go) ----

type coreMockStaffRepository struct {
	findAllFn                          func(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error)
	findByIDFn                         func(ctx context.Context, id uint64) (*model.Staff, error)
	findByIDInClinicFn                 func(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	lockForUpdateFn                    func(ctx context.Context, id uint64) (*model.Staff, error)
	lockInClinicFn                     func(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	lockForShareFn                     func(ctx context.Context, id uint64) (*model.Staff, error)
	createFn                           func(ctx context.Context, staff *model.Staff) error
	updateFn                           func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn                           func(ctx context.Context, clinicID, id uint64) error
	reorderFn                          func(ctx context.Context, clinicID uint64, ids []uint64) error
	countBlockingReferencesByStaffIDFn func(ctx context.Context, clinicID, staffID uint64) ([]StaffDependencyCount, error)
	isActiveSystemAdminStaffFn         func(ctx context.Context, staffID uint64) (bool, error)
	countActiveSystemAdminStaffFn      func(ctx context.Context) (int64, error)
}

func (m *coreMockStaffRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID, page, limit)
	}
	return nil, 0, nil
}
func (m *coreMockStaffRepository) FindByID(ctx context.Context, id uint64) (*model.Staff, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return &model.Staff{ID: id}, nil
}
func (m *coreMockStaffRepository) FindByIDInClinic(
	ctx context.Context,
	clinicID, id uint64,
) (*model.Staff, error) {
	if m.findByIDInClinicFn != nil {
		return m.findByIDInClinicFn(ctx, clinicID, id)
	}
	return &model.Staff{ID: id, ClinicID: clinicID}, nil
}
func (m *coreMockStaffRepository) LockActiveByIDForUpdate(ctx context.Context, id uint64) (*model.Staff, error) {
	if m.lockForUpdateFn != nil {
		return m.lockForUpdateFn(ctx, id)
	}
	return &model.Staff{ID: id}, nil
}
func (m *coreMockStaffRepository) LockActiveByIDForUpdateInClinic(
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
func (m *coreMockStaffRepository) LockActiveByIDForShare(ctx context.Context, id uint64) (*model.Staff, error) {
	if m.lockForShareFn != nil {
		return m.lockForShareFn(ctx, id)
	}
	return &model.Staff{ID: id}, nil
}
func (m *coreMockStaffRepository) FindByAccountID(_ context.Context, _ uint64) (*model.Staff, error) {
	return nil, nil
}
func (m *coreMockStaffRepository) Create(ctx context.Context, staff *model.Staff) error {
	if m.createFn != nil {
		return m.createFn(ctx, staff)
	}
	return nil
}
func (m *coreMockStaffRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, fields)
	}
	return nil
}
func (m *coreMockStaffRepository) UpdatePrimaryClinicID(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *coreMockStaffRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}
func (m *coreMockStaffRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}
func (m *coreMockStaffRepository) CountBlockingReferencesByStaffID(ctx context.Context, clinicID, staffID uint64) ([]StaffDependencyCount, error) {
	if m.countBlockingReferencesByStaffIDFn != nil {
		return m.countBlockingReferencesByStaffIDFn(ctx, clinicID, staffID)
	}
	return nil, nil
}

func (m *coreMockStaffRepository) IsActiveSystemAdminStaff(ctx context.Context, staffID uint64) (bool, error) {
	if m.isActiveSystemAdminStaffFn != nil {
		return m.isActiveSystemAdminStaffFn(ctx, staffID)
	}
	return false, nil
}

func (m *coreMockStaffRepository) CountActiveSystemAdminStaff(ctx context.Context) (int64, error) {
	if m.countActiveSystemAdminStaffFn != nil {
		return m.countActiveSystemAdminStaffFn(ctx)
	}
	return 0, nil
}

func (m *coreMockStaffRepository) CreateForReservation(_ context.Context, _ *model.Staff, _ uint64) error {
	return nil
}

func (m *coreMockStaffRepository) UpdateForReservation(_ context.Context, _, _ uint64, _ map[string]any) error {
	return nil
}

func (m *coreMockStaffRepository) DeleteForReservation(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *coreMockStaffRepository) SwapSortOrderForReservation(_ context.Context, _, _ uint64, _ string) error {
	return nil
}

type coreMockAccountRepository struct {
	updatePasswordHashFn func(
		ctx context.Context,
		id uint64,
		newHash string,
		updatedAt time.Time,
	) error
	deletePasswordResetTokensFn func(ctx context.Context, id uint64) error
}

func (m *coreMockAccountRepository) FindByID(_ context.Context, _ uint64) (*model.Account, error) {
	return nil, nil
}
func (m *coreMockAccountRepository) FindByEmail(_ context.Context, _ string) (*model.Account, error) {
	return nil, nil
}
func (m *coreMockAccountRepository) Create(_ context.Context, _ *model.Account) error { return nil }
func (m *coreMockAccountRepository) UpdatePasswordHash(
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
func (m *coreMockAccountRepository) DeletePasswordResetTokens(
	ctx context.Context,
	id uint64,
) error {
	if m.deletePasswordResetTokensFn != nil {
		return m.deletePasswordResetTokensFn(ctx, id)
	}
	return nil
}

type coreMockStaffClinicAssignmentRepository struct {
	createFn             func(ctx context.Context, assignment *model.StaffClinicAssignment) error
	lockActiveFn         func(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error)
	lockStaffAndClinicFn func(ctx context.Context, staffID, clinicID uint64) (*model.StaffClinicAssignment, error)
}

func (m *coreMockStaffClinicAssignmentRepository) FindByStaffID(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
	return nil, nil
}
func (m *coreMockStaffClinicAssignmentRepository) FindByStaffAndClinic(
	_ context.Context,
	staffID, clinicID uint64,
) (*model.StaffClinicAssignment, error) {
	return &model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID}, nil
}
func (m *coreMockStaffClinicAssignmentRepository) CountByStaffAndClinic(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *coreMockStaffClinicAssignmentRepository) LockActiveByStaffAndClinic(
	ctx context.Context,
	staffID, clinicID uint64,
) (*model.StaffClinicAssignment, error) {
	if m.lockStaffAndClinicFn != nil {
		return m.lockStaffAndClinicFn(ctx, staffID, clinicID)
	}
	return &model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID}, nil
}
func (m *coreMockStaffClinicAssignmentRepository) LockActiveByStaff(
	ctx context.Context,
	staffID uint64,
) ([]model.StaffClinicAssignment, error) {
	if m.lockActiveFn != nil {
		return m.lockActiveFn(ctx, staffID)
	}
	return []model.StaffClinicAssignment{{StaffID: staffID, ClinicID: 1, IsMain: true}}, nil
}
func (m *coreMockStaffClinicAssignmentRepository) Create(ctx context.Context, assignment *model.StaffClinicAssignment) error {
	if m.createFn != nil {
		return m.createFn(ctx, assignment)
	}
	return nil
}
func (m *coreMockStaffClinicAssignmentRepository) RestoreOrCreate(
	ctx context.Context,
	assignment *model.StaffClinicAssignment,
) error {
	return m.Create(ctx, assignment)
}
func (m *coreMockStaffClinicAssignmentRepository) Delete(_ context.Context, _ uint64) error {
	return nil
}
func (m *coreMockStaffClinicAssignmentRepository) DeleteByStaffAndClinicIDs(
	_ context.Context,
	_ uint64,
	_ []uint64,
) error {
	return nil
}

type coreMockReservationQueryRepository struct {
	existsByStaffIDFn        func(ctx context.Context, clinicID, staffID uint64) (bool, error)
	findClinicIDsByStaffIDFn func(ctx context.Context, clinicIDs []uint64, staffID uint64) ([]uint64, error)
}

func (m *coreMockReservationQueryRepository) ExistsByReservationTypeID(_ context.Context, _, _ uint64) (bool, error) {
	return false, nil
}
func (m *coreMockReservationQueryRepository) ExistsByStaffID(ctx context.Context, clinicID, staffID uint64) (bool, error) {
	if m.existsByStaffIDFn != nil {
		return m.existsByStaffIDFn(ctx, clinicID, staffID)
	}
	return false, nil
}
func (m *coreMockReservationQueryRepository) FindClinicIDsByStaffID(
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
func (m *coreMockReservationQueryRepository) CountMedicalRecordsByReservationID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *coreMockReservationQueryRepository) CountByCustomerAndDateRange(_ context.Context, _, _ uint64, _, _ time.Time) (int64, error) {
	return 0, nil
}
func (m *coreMockReservationQueryRepository) CountByDateAndSource(_ context.Context, _ uint64, _ time.Time, _ model.ReservationSource) (int64, error) {
	return 0, nil
}
func (m *coreMockReservationQueryRepository) FindAllByCategory(_ context.Context, _ uint64, _ model.ReservationTypeCategory, _, _ *uint64, _, _ *string, _, _ int) ([]model.Reservation, int64, error) {
	return nil, 0, nil
}
func (m *coreMockReservationQueryRepository) FindNoShowCandidates(_ context.Context, _ uint64) ([]model.Reservation, error) {
	return nil, nil
}

func (m *coreMockReservationQueryRepository) AssertOwnerInClinic(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *coreMockReservationQueryRepository) FindPetOwnerInClinic(_ context.Context, _, _ uint64) (uint64, error) {
	return 0, nil
}

func (m *coreMockReservationQueryRepository) FindPetByIDInClinic(_ context.Context, _, petID uint64) (*model.Pet, error) {
	return &model.Pet{ID: petID, Status: model.PetStatusAlive}, nil
}

func (m *coreMockReservationQueryRepository) AssertLineCustomerInClinic(_ context.Context, _, _ uint64) error {
	return nil
}

type coreMockShiftEntryRepository struct {
	existsByStaffIDFn        func(ctx context.Context, clinicID, staffID uint64) (bool, error)
	findClinicIDsByStaffIDFn func(ctx context.Context, clinicIDs []uint64, staffID uint64) ([]uint64, error)
}

func (m *coreMockShiftEntryRepository) FindAll(_ context.Context, _ uint64, _ ShiftEntryFilter) ([]model.ShiftEntry, error) {
	return nil, nil
}
func (m *coreMockShiftEntryRepository) FindByID(_ context.Context, _, _ uint64) (*model.ShiftEntry, error) {
	return nil, nil
}
func (m *coreMockShiftEntryRepository) LockActiveByIDForUpdate(
	_ context.Context,
	clinicID, id uint64,
) (*model.ShiftEntry, error) {
	return &model.ShiftEntry{ID: id, ClinicID: clinicID}, nil
}
func (m *coreMockShiftEntryRepository) Create(_ context.Context, _ *model.ShiftEntry) error {
	return nil
}
func (m *coreMockShiftEntryRepository) Update(_ context.Context, _, _ uint64, _ map[string]any) error {
	return nil
}
func (m *coreMockShiftEntryRepository) Delete(_ context.Context, _, _ uint64) error { return nil }
func (m *coreMockShiftEntryRepository) ExistsByStaffID(ctx context.Context, clinicID, staffID uint64) (bool, error) {
	if m.existsByStaffIDFn != nil {
		return m.existsByStaffIDFn(ctx, clinicID, staffID)
	}
	return false, nil
}
func (m *coreMockShiftEntryRepository) FindClinicIDsByStaffID(
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
func (m *coreMockShiftEntryRepository) ReplaceBreaks(_ context.Context, _ uint64, _ []model.ShiftEntryBreak) error {
	return nil
}
func (m *coreMockShiftEntryRepository) FindOnDutyStaffs(_ context.Context, _ uint64, _ time.Time) ([]model.Staff, error) {
	return nil, nil
}

func (m *coreMockShiftEntryRepository) SaveByStaffDate(
	_ context.Context,
	_ uint64,
	_ *model.ShiftEntry,
	_ []model.ShiftEntryBreak,
) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error) {
	return nil, nil, false, nil
}

func (m *coreMockShiftEntryRepository) DeleteByStaffDate(_ context.Context, _, _ uint64, _ time.Time) error {
	return nil
}

// coreFakeTransactor runs fn directly without a real transaction (WithTx passthrough).
type coreFakeTransactor struct {
	withTxErr error
	calls     int
}

func (t *coreFakeTransactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	t.calls++
	if t.withTxErr != nil {
		return t.withTxErr
	}
	return fn(ctx)
}

func newCoreStaffService(
	repo *coreMockStaffRepository,
	accountRepo *coreMockAccountRepository,
	assignmentRepo *coreMockStaffClinicAssignmentRepository,
	reservationRepo *coreMockReservationQueryRepository,
	shiftEntryRepo *coreMockShiftEntryRepository,
	tx Transactor,
) StaffService {
	return NewStaffServiceWithCredentialAudit(
		repo,
		accountRepo,
		assignmentRepo,
		reservationRepo,
		shiftEntryRepo,
		nil,
		nil,
		nil,
		nil,
		tx,
		noopStaffCredentialAuditTxLogger{},
	)
}

func TestStaffServiceCore_GetByIDInClinicUsesScopedRepositoryRead(t *testing.T) {
	repo := &coreMockStaffRepository{
		findByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
			t.Fatal("HTTP-scoped staff read must not use global FindByID")
			return nil, nil
		},
		findByIDInClinicFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
			assert.Equal(t, uint64(20), clinicID)
			return &model.Staff{
				ID:           id,
				ClinicID:     10,
				OccupationID: nil,
				Occupation:   nil,
			}, nil
		},
	}
	svc := newCoreStaffService(
		repo,
		&coreMockAccountRepository{},
		&coreMockStaffClinicAssignmentRepository{},
		&coreMockReservationQueryRepository{},
		&coreMockShiftEntryRepository{},
		&coreFakeTransactor{},
	)
	scoped, ok := svc.(interface {
		GetByIDInClinic(context.Context, uint64, uint64) (*model.Staff, error)
	})
	require.True(t, ok, "staff service must expose a clinic-scoped HTTP read")

	got, err := scoped.GetByIDInClinic(context.Background(), 20, 7)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Nil(t, got.OccupationID)
}

// ---- Create ----

func TestStaffServiceCore_Create(t *testing.T) {
	tests := []struct {
		name      string
		input     *CreateStaffInput
		createFn  func(ctx context.Context, staff *model.Staff) error
		assignFn  func(ctx context.Context, assignment *model.StaffClinicAssignment) error
		withTxErr error
		wantErr   bool
	}{
		{
			name:  "creates staff and assigns to clinic successfully",
			input: &CreateStaffInput{ClinicID: 1, Name: "山田太郎"},
			createFn: func(_ context.Context, staff *model.Staff) error {
				staff.ID = 100
				return nil
			},
			assignFn: func(_ context.Context, _ *model.StaffClinicAssignment) error { return nil },
			wantErr:  false,
		},
		{
			name:    "returns error when name is empty",
			input:   &CreateStaffInput{ClinicID: 1, Name: ""},
			wantErr: true,
		},
		{
			name:    "returns error when clinic_id is zero",
			input:   &CreateStaffInput{ClinicID: 0, Name: "山田太郎"},
			wantErr: true,
		},
		{
			name:  "returns wrapped error when repo.Create fails",
			input: &CreateStaffInput{ClinicID: 1, Name: "山田太郎"},
			createFn: func(_ context.Context, _ *model.Staff) error {
				return errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:  "returns wrapped error when assignmentRepo.Create fails",
			input: &CreateStaffInput{ClinicID: 1, Name: "山田太郎"},
			createFn: func(_ context.Context, staff *model.Staff) error {
				staff.ID = 100
				return nil
			},
			assignFn: func(_ context.Context, _ *model.StaffClinicAssignment) error {
				return errors.New("assignment db error")
			},
			wantErr: true,
		},
		{
			name:      "propagates transaction failure",
			input:     &CreateStaffInput{ClinicID: 1, Name: "山田太郎"},
			withTxErr: errors.New("tx failed"),
			wantErr:   true,
		},
		{
			name: "defaults staff_type to doctor when not provided",
			input: &CreateStaffInput{
				ClinicID: 1, Name: "山田太郎",
			},
			createFn: func(_ context.Context, staff *model.Staff) error {
				assert.Equal(t, model.StaffTypeDoctor, staff.StaffType)
				staff.ID = 101
				return nil
			},
			assignFn: func(_ context.Context, _ *model.StaffClinicAssignment) error { return nil },
			wantErr:  false,
		},
		{
			name: "trims whitespace from name",
			input: &CreateStaffInput{
				ClinicID: 1, Name: "  田中花子  ",
			},
			createFn: func(_ context.Context, staff *model.Staff) error {
				assert.Equal(t, "田中花子", staff.Name)
				staff.ID = 102
				return nil
			},
			assignFn: func(_ context.Context, _ *model.StaffClinicAssignment) error { return nil },
			wantErr:  false,
		},
		{
			name: "reservation_visible defaults to true when nil",
			input: &CreateStaffInput{
				ClinicID: 1, Name: "山田太郎",
			},
			createFn: func(_ context.Context, staff *model.Staff) error {
				assert.True(t, staff.ReservationVisible)
				staff.ID = 103
				return nil
			},
			assignFn: func(_ context.Context, _ *model.StaffClinicAssignment) error { return nil },
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &coreMockStaffRepository{createFn: tt.createFn}
			assignRepo := &coreMockStaffClinicAssignmentRepository{createFn: tt.assignFn}
			tx := &coreFakeTransactor{withTxErr: tt.withTxErr}
			svc := newCoreStaffService(repo, &coreMockAccountRepository{}, assignRepo, &coreMockReservationQueryRepository{}, &coreMockShiftEntryRepository{}, tx)

			staff, err := svc.Create(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, staff)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, staff)
		})
	}
}

func TestStaffServiceCore_Create_LocksOccupationInsideWriteTransaction(t *testing.T) {
	occupationID := uint64(30)
	findCalled := false
	lockCalled := false
	occupationRepo := &mockOccupationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Occupation, error) {
			findCalled = true
			return &model.Occupation{ID: occupationID, ClinicID: 10}, nil
		},
		lockForShareFn: func(_ context.Context, clinicID, id uint64) (*model.Occupation, error) {
			lockCalled = true
			return &model.Occupation{ID: id, ClinicID: clinicID}, nil
		},
	}
	repo := &coreMockStaffRepository{
		createFn: func(_ context.Context, staff *model.Staff) error {
			staff.ID = 99
			return nil
		},
	}
	svc := NewStaffService(
		repo,
		&coreMockAccountRepository{},
		&coreMockStaffClinicAssignmentRepository{},
		&coreMockReservationQueryRepository{},
		&coreMockShiftEntryRepository{},
		nil,
		nil,
		occupationRepo,
		nil,
		&coreFakeTransactor{},
	)

	created, err := svc.Create(context.Background(), &CreateStaffInput{
		ClinicID:     10,
		Name:         "山田太郎",
		OccupationID: &occupationID,
	})

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.True(t, lockCalled)
	assert.False(t, findCalled)
}

// ---- Update ----
