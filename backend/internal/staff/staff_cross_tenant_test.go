package staff

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type crossTenantStaffRepository struct {
	Repository
	createCalled bool
	updateCalled bool
}

func (r *crossTenantStaffRepository) Create(context.Context, *model.Staff) error {
	r.createCalled = true
	return nil
}

func (*crossTenantStaffRepository) LockActiveByIDForUpdateInClinic(
	_ context.Context,
	clinicID, id uint64,
) (*model.Staff, error) {
	return &model.Staff{ID: id, ClinicID: clinicID}, nil
}

func (r *crossTenantStaffRepository) Update(
	context.Context,
	uint64,
	uint64,
	UpdateStaffInput,
) error {
	r.updateCalled = true
	return nil
}

type crossTenantStaffAccountStore struct {
	createCalled bool
}

func (*crossTenantStaffAccountStore) FindByEmail(
	context.Context,
	string,
) (*model.Account, error) {
	return nil, apperrors.WrapNotFound("account", "email")
}

func (s *crossTenantStaffAccountStore) Create(context.Context, *model.Account) error {
	s.createCalled = true
	return nil
}

func (*crossTenantStaffAccountStore) UpdatePasswordHash(
	context.Context,
	uint64,
	string,
	time.Time,
) error {
	return nil
}

func (*crossTenantStaffAccountStore) DeletePasswordResetTokens(context.Context, uint64) error {
	return nil
}

type crossTenantStaffAssignmentRepository struct {
	StaffClinicAssignmentRepository
	createCalled bool
}

func (r *crossTenantStaffAssignmentRepository) Create(
	context.Context,
	*model.StaffClinicAssignment,
) error {
	r.createCalled = true
	return nil
}

func (*crossTenantStaffAssignmentRepository) LockActiveByStaff(
	_ context.Context,
	staffID uint64,
) ([]model.StaffClinicAssignment, error) {
	return []model.StaffClinicAssignment{{
		StaffID:  staffID,
		ClinicID: 1,
	}}, nil
}

type rejectingCrossTenantOccupationRepository struct {
	OccupationRepository
	lockCalled bool
}

func (r *rejectingCrossTenantOccupationRepository) LockActiveByIDForShare(
	context.Context,
	uint64,
	uint64,
) (*model.Occupation, error) {
	r.lockCalled = true
	return nil, apperrors.WrapNotFound("occupation", "foreign")
}

type stubReservationForStaff struct{}

func (*stubReservationForStaff) ExistsByStaffID(
	context.Context,
	uint64,
	uint64,
) (bool, error) {
	return false, nil
}

func (*stubReservationForStaff) FindClinicIDsByStaffID(
	context.Context,
	[]uint64,
	uint64,
) ([]uint64, error) {
	return nil, nil
}

func TestService_Create_RejectsCrossClinicOccupationID(t *testing.T) {
	occupationID := uint64(999)
	staffRepo := &crossTenantStaffRepository{}
	assignmentRepo := &crossTenantStaffAssignmentRepository{}
	occupationRepo := &rejectingCrossTenantOccupationRepository{}

	svc := NewService(
		staffRepo,
		nil,
		assignmentRepo,
		nil,
		nil,
		nil,
		nil,
		occupationRepo,
		nil,
		&mockTransactor{},
	)

	staff, err := svc.Create(context.Background(), &CreateStaffInput{
		ClinicID:     1,
		Name:         "clinic scoped occupation",
		OccupationID: &occupationID,
	})

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.True(t, occupationRepo.lockCalled)
	assert.False(t, staffRepo.createCalled)
	assert.False(t, assignmentRepo.createCalled)
}

func TestService_CreateWithAccount_RejectsCrossClinicOccupationID(t *testing.T) {
	occupationID := uint64(999)
	staffRepo := &crossTenantStaffRepository{}
	accountRepo := &crossTenantStaffAccountStore{}
	assignmentRepo := &crossTenantStaffAssignmentRepository{}
	occupationRepo := &rejectingCrossTenantOccupationRepository{}

	svc := NewService(
		staffRepo,
		accountRepo,
		assignmentRepo,
		nil,
		nil,
		nil,
		nil,
		occupationRepo,
		nil,
		&mockTransactor{},
	)

	staff, err := svc.CreateWithAccount(context.Background(), &CreateStaffWithAccountInput{
		ClinicID:     1,
		Name:         "clinic scoped occupation",
		Email:        "staff@example.com",
		Password:     "Passw0rd1",
		OccupationID: &occupationID,
	})

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.True(t, occupationRepo.lockCalled)
	assert.False(t, staffRepo.createCalled)
	assert.False(t, accountRepo.createCalled)
	assert.False(t, assignmentRepo.createCalled)
}

func TestService_Update_RejectsCrossClinicOccupationID(t *testing.T) {
	occupationID := uint64(999)
	staffRepo := &crossTenantStaffRepository{}
	assignmentRepo := &crossTenantStaffAssignmentRepository{}
	occupationRepo := &rejectingCrossTenantOccupationRepository{}

	svc := NewService(
		staffRepo,
		nil,
		assignmentRepo,
		nil,
		nil,
		nil,
		nil,
		occupationRepo,
		nil,
		&mockTransactor{},
	)

	staff, err := svc.Update(context.Background(), 1, 7, &UpdateStaffInput{
		OccupationID:        &occupationID,
		AuthorizedClinicIDs: []uint64{1},
	})

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.True(t, occupationRepo.lockCalled)
	assert.False(t, staffRepo.updateCalled)
}
