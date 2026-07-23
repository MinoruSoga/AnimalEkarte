package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- local minimal mocks (scoped to this file to avoid touching concurrently-edited
//      staff_service_test.go / staff_service_account_test.go / staff_service_permissions_test.go) ----

type coreMockStaffRepository struct {
	findAllFn                          func(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error)
	findByIDFn                         func(ctx context.Context, id uint64) (*model.Staff, error)
	lockForUpdateFn                    func(ctx context.Context, id uint64) (*model.Staff, error)
	lockInClinicFn                     func(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	lockForShareFn                     func(ctx context.Context, id uint64) (*model.Staff, error)
	createFn                           func(ctx context.Context, staff *model.Staff) error
	updateFn                           func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn                           func(ctx context.Context, clinicID, id uint64) error
	reorderFn                          func(ctx context.Context, clinicID uint64, ids []uint64) error
	countBlockingReferencesByStaffIDFn func(ctx context.Context, clinicID, staffID uint64) ([]repository.StaffDependencyCount, error)
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
func (m *coreMockStaffRepository) CountBlockingReferencesByStaffID(ctx context.Context, clinicID, staffID uint64) ([]repository.StaffDependencyCount, error) {
	if m.countBlockingReferencesByStaffIDFn != nil {
		return m.countBlockingReferencesByStaffIDFn(ctx, clinicID, staffID)
	}
	return nil, nil
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
	updateFn func(ctx context.Context, id uint64, fields map[string]any) error
}

func (m *coreMockAccountRepository) FindByID(_ context.Context, _ uint64) (*model.Account, error) {
	return nil, nil
}
func (m *coreMockAccountRepository) FindByEmail(_ context.Context, _ string) (*model.Account, error) {
	return nil, nil
}
func (m *coreMockAccountRepository) Create(_ context.Context, _ *model.Account) error { return nil }
func (m *coreMockAccountRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, fields)
	}
	return nil
}

type coreMockStaffClinicAssignmentRepository struct {
	createFn     func(ctx context.Context, assignment *model.StaffClinicAssignment) error
	lockActiveFn func(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error)
}

func (m *coreMockStaffClinicAssignmentRepository) FindByStaffID(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
	return nil, nil
}
func (m *coreMockStaffClinicAssignmentRepository) CountByStaffAndClinic(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *coreMockStaffClinicAssignmentRepository) LockActiveByStaffAndClinic(
	_ context.Context,
	staffID, clinicID uint64,
) (*model.StaffClinicAssignment, error) {
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

type coreMockReservationQueryRepository struct {
	existsByStaffIDFn func(ctx context.Context, clinicID, staffID uint64) (bool, error)
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

func (m *coreMockReservationQueryRepository) AssertLineCustomerInClinic(_ context.Context, _, _ uint64) error {
	return nil
}

type coreMockShiftEntryRepository struct {
	existsByStaffIDFn func(ctx context.Context, clinicID, staffID uint64) (bool, error)
}

func (m *coreMockShiftEntryRepository) FindAll(_ context.Context, _ uint64, _ repository.ShiftEntryFilter) ([]model.ShiftEntry, error) {
	return nil, nil
}
func (m *coreMockShiftEntryRepository) FindByID(_ context.Context, _, _ uint64) (*model.ShiftEntry, error) {
	return nil, nil
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
func (m *coreMockShiftEntryRepository) ReplaceBreaks(_ context.Context, _ uint64, _ []model.ShiftEntryBreak) error {
	return nil
}
func (m *coreMockShiftEntryRepository) FindOnDutyStaffs(_ context.Context, _ uint64, _ time.Time) ([]model.Staff, error) {
	return nil, nil
}

func (m *coreMockShiftEntryRepository) SaveByStaffDate(_ context.Context, _ uint64, _ *model.ShiftEntry, _ []model.ShiftEntryBreak) error {
	return nil
}

func (m *coreMockShiftEntryRepository) DeleteByStaffDate(_ context.Context, _, _ uint64, _ time.Time) error {
	return nil
}

// coreFakeTransactor runs fn directly without a real transaction (WithTx passthrough).
type coreFakeTransactor struct {
	withTxErr error
}

func (t *coreFakeTransactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
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
	tx repository.Transactor,
) StaffService {
	return NewStaffService(repo, accountRepo, assignmentRepo, reservationRepo, shiftEntryRepo, nil, nil, nil, nil, tx)
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

// ---- Update ----

func TestStaffServiceCore_Update(t *testing.T) {
	name := "更新太郎"
	emptyName := ""
	password := "newpassword1"
	weakPassword := "short"

	tests := []struct {
		name           string
		input          *UpdateStaffInput
		findByIDFn     func(ctx context.Context, id uint64) (*model.Staff, error)
		updateFn       func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
		accountUpdate  func(ctx context.Context, id uint64, fields map[string]any) error
		wantErr        bool
		wantErrInvalid bool
	}{
		{
			name:    "updates profile fields successfully",
			input:   &UpdateStaffInput{Name: &name},
			wantErr: false,
		},
		{
			name:           "returns invalid input when no fields are set",
			input:          &UpdateStaffInput{},
			wantErr:        true,
			wantErrInvalid: true,
		},
		{
			name:  "returns error when staff not found",
			input: &UpdateStaffInput{Name: &name},
			findByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
				return nil, apperrors.WrapNotFound("staff", "1")
			},
			wantErr: true,
		},
		{
			name:    "returns error when name is empty",
			input:   &UpdateStaffInput{Name: &emptyName},
			wantErr: true,
		},
		{
			name:  "returns wrapped error when repo.Update fails",
			input: &UpdateStaffInput{Name: &name},
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				return errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:    "updates password when account exists",
			input:   &UpdateStaffInput{Password: &password},
			wantErr: false,
		},
		{
			name:    "returns error when password is too weak",
			input:   &UpdateStaffInput{Password: &weakPassword},
			wantErr: true,
		},
		{
			name:  "returns wrapped error when account password update fails",
			input: &UpdateStaffInput{Password: &password},
			accountUpdate: func(_ context.Context, _ uint64, _ map[string]any) error {
				return errors.New("account db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountID := uint64(5)
			findByIDFn := tt.findByIDFn
			if findByIDFn == nil {
				findByIDFn = func(_ context.Context, id uint64) (*model.Staff, error) {
					return &model.Staff{ID: id, ClinicID: 1, AccountID: &accountID}, nil
				}
			}
			repo := &coreMockStaffRepository{
				findByIDFn: findByIDFn,
				updateFn:   tt.updateFn,
			}
			accountRepo := &coreMockAccountRepository{updateFn: tt.accountUpdate}
			tx := &coreFakeTransactor{}
			svc := newCoreStaffService(repo, accountRepo, &coreMockStaffClinicAssignmentRepository{}, &coreMockReservationQueryRepository{}, &coreMockShiftEntryRepository{}, tx)

			staff, err := svc.Update(context.Background(), 1, 1, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, staff)
				if tt.wantErrInvalid {
					assert.True(t, apperrors.IsInvalidInput(err))
				}
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, staff)
		})
	}
}

func TestStaffServiceCore_Update_FindByIDErrorAfterProfileUpdate(t *testing.T) {
	// covers the second FindByID call (reload after profile update) failing.
	name := "更新太郎"
	callCount := 0
	repo := &coreMockStaffRepository{
		findByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
			callCount++
			if callCount == 1 {
				return &model.Staff{ID: id, ClinicID: 1}, nil
			}
			return nil, errors.New("reload failed")
		},
	}
	tx := &coreFakeTransactor{}
	svc := newCoreStaffService(repo, &coreMockAccountRepository{}, &coreMockStaffClinicAssignmentRepository{}, &coreMockReservationQueryRepository{}, &coreMockShiftEntryRepository{}, tx)

	staff, err := svc.Update(context.Background(), 1, 1, &UpdateStaffInput{Name: &name})
	assert.Error(t, err)
	assert.Nil(t, staff)
}

// ---- Delete ----

func TestStaffServiceCore_Delete(t *testing.T) {
	tests := []struct {
		name                    string
		findByIDFn              func(ctx context.Context, id uint64) (*model.Staff, error)
		existsByStaffIDFn       func(ctx context.Context, clinicID, staffID uint64) (bool, error)
		shiftExistsByStaffIDFn  func(ctx context.Context, clinicID, staffID uint64) (bool, error)
		countBlockingReferences func(ctx context.Context, clinicID, staffID uint64) ([]repository.StaffDependencyCount, error)
		deleteFn                func(ctx context.Context, clinicID, id uint64) error
		wantErr                 bool
		wantConflict            bool
	}{
		{
			name:    "deletes staff with no dependencies",
			wantErr: false,
		},
		{
			name: "returns error when staff not found",
			findByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
				return nil, apperrors.WrapNotFound("staff", "1")
			},
			wantErr: true,
		},
		{
			name: "returns conflict when staff has reservations",
			existsByStaffIDFn: func(_ context.Context, _, _ uint64) (bool, error) {
				return true, nil
			},
			wantErr:      true,
			wantConflict: true,
		},
		{
			name: "returns wrapped error when reservation dependency check fails",
			existsByStaffIDFn: func(_ context.Context, _, _ uint64) (bool, error) {
				return false, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name: "returns conflict when staff has shift entries",
			shiftExistsByStaffIDFn: func(_ context.Context, _, _ uint64) (bool, error) {
				return true, nil
			},
			wantErr:      true,
			wantConflict: true,
		},
		{
			name: "returns wrapped error when shift dependency check fails",
			shiftExistsByStaffIDFn: func(_ context.Context, _, _ uint64) (bool, error) {
				return false, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name: "returns wrapped error when blocking-reference check fails",
			countBlockingReferences: func(_ context.Context, _, _ uint64) ([]repository.StaffDependencyCount, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name: "returns conflict when staff has blocking references",
			countBlockingReferences: func(_ context.Context, _, _ uint64) ([]repository.StaffDependencyCount, error) {
				return []repository.StaffDependencyCount{{Label: "カルテ", Count: 3}}, nil
			},
			wantErr:      true,
			wantConflict: true,
		},
		{
			name: "returns wrapped error when repo.Delete fails",
			deleteFn: func(_ context.Context, _, _ uint64) error {
				return errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &coreMockStaffRepository{
				findByIDFn:                         tt.findByIDFn,
				countBlockingReferencesByStaffIDFn: tt.countBlockingReferences,
				deleteFn:                           tt.deleteFn,
			}
			reservationRepo := &coreMockReservationQueryRepository{existsByStaffIDFn: tt.existsByStaffIDFn}
			shiftRepo := &coreMockShiftEntryRepository{existsByStaffIDFn: tt.shiftExistsByStaffIDFn}
			svc := newCoreStaffService(repo, &coreMockAccountRepository{}, &coreMockStaffClinicAssignmentRepository{}, reservationRepo, shiftRepo, &coreFakeTransactor{})

			err := svc.Delete(context.Background(), 1, 1)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestStaffService_Delete_UsesCanonicalLockOrderAndTransactionContext(t *testing.T) {
	events := make([]string, 0, 7)
	repo := &coreMockStaffRepository{
		lockInClinicFn: func(ctx context.Context, clinicID, id uint64) (*model.Staff, error) {
			requireStaffSecurityTxContext(t, ctx)
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(7), id)
			events = append(events, "lock-staff")
			return &model.Staff{ID: id}, nil
		},
		countBlockingReferencesByStaffIDFn: func(ctx context.Context, clinicID, staffID uint64) ([]repository.StaffDependencyCount, error) {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, "dependencies")
			return nil, nil
		},
		deleteFn: func(ctx context.Context, clinicID, id uint64) error {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, "delete")
			return nil
		},
	}
	assignmentRepo := &coreMockStaffClinicAssignmentRepository{
		lockActiveFn: func(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, "lock-assignments")
			return []model.StaffClinicAssignment{{StaffID: staffID, ClinicID: 1, IsMain: true}}, nil
		},
	}
	reservationRepo := &coreMockReservationQueryRepository{
		existsByStaffIDFn: func(ctx context.Context, clinicID, staffID uint64) (bool, error) {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, "reservations")
			return false, nil
		},
	}
	shiftRepo := &coreMockShiftEntryRepository{
		existsByStaffIDFn: func(ctx context.Context, clinicID, staffID uint64) (bool, error) {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, "shifts")
			return false, nil
		},
	}
	svc := newCoreStaffService(
		repo,
		&coreMockAccountRepository{},
		assignmentRepo,
		reservationRepo,
		shiftRepo,
		markedStaffSecurityTransactor{},
	)

	err := svc.Delete(context.Background(), 1, 7)

	require.NoError(t, err)
	assert.Equal(t, []string{
		"lock-staff",
		"lock-assignments",
		"reservations",
		"shifts",
		"dependencies",
		"delete",
	}, events)
}

func TestStaffService_Delete_RejectsInvalidOrMultiClinicAssignmentStateBeforeDependencies(t *testing.T) {
	tests := []struct {
		name         string
		assignments  []model.StaffClinicAssignment
		wantNotFound bool
		wantConflict bool
	}{
		{
			name:         "no active assignments",
			assignments:  nil,
			wantNotFound: true,
		},
		{
			name: "only another clinic assignment",
			assignments: []model.StaffClinicAssignment{
				{StaffID: 7, ClinicID: 2},
			},
			wantNotFound: true,
		},
		{
			name: "multiple active assignments",
			assignments: []model.StaffClinicAssignment{
				{StaffID: 7, ClinicID: 1, IsMain: true},
				{StaffID: 7, ClinicID: 2},
			},
			wantConflict: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &coreMockStaffRepository{
				lockInClinicFn: func(_ context.Context, _, id uint64) (*model.Staff, error) {
					return &model.Staff{ID: id}, nil
				},
				countBlockingReferencesByStaffIDFn: func(_ context.Context, _, _ uint64) ([]repository.StaffDependencyCount, error) {
					t.Fatal("dependency state must not be disclosed before assignment validation")
					return nil, nil
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					t.Fatal("invalid assignment state must not delete staff")
					return nil
				},
			}
			assignmentRepo := &coreMockStaffClinicAssignmentRepository{
				lockActiveFn: func(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
					return tt.assignments, nil
				},
			}
			reservationRepo := &coreMockReservationQueryRepository{
				existsByStaffIDFn: func(_ context.Context, _, _ uint64) (bool, error) {
					t.Fatal("dependency state must not be disclosed before assignment validation")
					return false, nil
				},
			}
			shiftRepo := &coreMockShiftEntryRepository{
				existsByStaffIDFn: func(_ context.Context, _, _ uint64) (bool, error) {
					t.Fatal("dependency state must not be disclosed before assignment validation")
					return false, nil
				},
			}
			svc := newCoreStaffService(
				repo,
				&coreMockAccountRepository{},
				assignmentRepo,
				reservationRepo,
				shiftRepo,
				&coreFakeTransactor{},
			)

			err := svc.Delete(context.Background(), 1, 7)

			require.Error(t, err)
			if tt.wantNotFound {
				assert.True(t, apperrors.IsNotFound(err), "unexpected error: %v", err)
			}
			if tt.wantConflict {
				assert.True(t, apperrors.IsConflict(err), "unexpected error: %v", err)
			}
		})
	}
}

// ---- List / GetByID / Reorder (already high coverage elsewhere — smoke tests only) ----

func TestStaffServiceCore_List(t *testing.T) {
	repo := &coreMockStaffRepository{
		findAllFn: func(_ context.Context, _ uint64, _, _ int) ([]model.Staff, int64, error) {
			return []model.Staff{{ID: 1}}, 1, nil
		},
	}
	svc := newCoreStaffService(repo, &coreMockAccountRepository{}, &coreMockStaffClinicAssignmentRepository{}, &coreMockReservationQueryRepository{}, &coreMockShiftEntryRepository{}, &coreFakeTransactor{})
	staffs, total, err := svc.List(context.Background(), 1, 1, 20)
	assert.NoError(t, err)
	assert.Len(t, staffs, 1)
	assert.Equal(t, int64(1), total)

	repo.findAllFn = func(_ context.Context, _ uint64, _, _ int) ([]model.Staff, int64, error) {
		return nil, 0, errors.New("db error")
	}
	_, _, err = svc.List(context.Background(), 1, 1, 20)
	assert.Error(t, err)
}

func TestStaffServiceCore_GetByID(t *testing.T) {
	repo := &coreMockStaffRepository{
		findByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id}, nil
		},
	}
	svc := newCoreStaffService(repo, &coreMockAccountRepository{}, &coreMockStaffClinicAssignmentRepository{}, &coreMockReservationQueryRepository{}, &coreMockShiftEntryRepository{}, &coreFakeTransactor{})
	staff, err := svc.GetByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, staff)

	repo.findByIDFn = func(_ context.Context, _ uint64) (*model.Staff, error) {
		return nil, apperrors.WrapNotFound("staff", "1")
	}
	_, err = svc.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestStaffServiceCore_Reorder(t *testing.T) {
	repo := &coreMockStaffRepository{
		reorderFn: func(_ context.Context, _ uint64, _ []uint64) error { return nil },
	}
	svc := newCoreStaffService(repo, &coreMockAccountRepository{}, &coreMockStaffClinicAssignmentRepository{}, &coreMockReservationQueryRepository{}, &coreMockShiftEntryRepository{}, &coreFakeTransactor{})

	assert.NoError(t, svc.Reorder(context.Background(), 1, []uint64{2, 1}))
	assert.Error(t, svc.Reorder(context.Background(), 1, []uint64{}))

	repo.reorderFn = func(_ context.Context, _ uint64, _ []uint64) error { return errors.New("db error") }
	assert.Error(t, svc.Reorder(context.Background(), 1, []uint64{1}))
}
