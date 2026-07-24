package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
)

var errReservationStaffReadback = errors.New("forced reservation staff readback failure")

type failingReservationStaffReadbackRepository struct {
	reservation.ReservationStaffRepository
	failFindCall int
	findCalls    int
	lockCalls    int
	failExcluded bool
}

func (r *failingReservationStaffReadbackRepository) LockForMutation(
	ctx context.Context,
	clinicID, staffID uint64,
) (*model.Staff, error) {
	r.lockCalls++
	return r.ReservationStaffRepository.LockForMutation(ctx, clinicID, staffID)
}

func (r *failingReservationStaffReadbackRepository) FindByID(
	ctx context.Context,
	clinicID, staffID uint64,
) (*model.Staff, error) {
	r.findCalls++
	if r.failFindCall > 0 && r.findCalls == r.failFindCall {
		return nil, errReservationStaffReadback
	}
	return r.ReservationStaffRepository.FindByID(ctx, clinicID, staffID)
}

func (r *failingReservationStaffReadbackRepository) FindAllExcludedReservationTypes(
	ctx context.Context,
	clinicID, staffID uint64,
) ([]model.StaffReservationExclusion, error) {
	if r.failExcluded {
		return nil, errReservationStaffReadback
	}
	return r.ReservationStaffRepository.FindAllExcludedReservationTypes(ctx, clinicID, staffID)
}

func TestReservationStaffService_Create_ReadbackFailureRollsBack(t *testing.T) {
	db := setupReservationStaffTxAtomicityTestDB(t)
	base := NewReservationStaffRepository(db)
	repo := &failingReservationStaffReadbackRepository{
		ReservationStaffRepository: base,
		failExcluded:               true,
	}
	service := reservation.NewReservationStaffService(repo, NewTransactor(db), nil)

	staff, exclusions, err := service.Create(
		context.Background(),
		1,
		&reservation.CreateReservationStaffInput{Name: "readback rollback create"},
	)

	require.ErrorIs(t, err, errReservationStaffReadback)
	assert.Nil(t, staff)
	assert.Nil(t, exclusions)
	var staffCount int64
	require.NoError(t, db.Model(&model.Staff{}).
		Where("name = ?", "readback rollback create").
		Count(&staffCount).Error)
	assert.Zero(t, staffCount)
	var assignmentCount int64
	require.NoError(t, db.Model(&model.StaffClinicAssignment{}).Count(&assignmentCount).Error)
	assert.Zero(t, assignmentCount)
}

func TestReservationStaffService_Update_StaffReadbackFailureRollsBack(t *testing.T) {
	db := setupReservationStaffTxAtomicityTestDB(t)
	const clinicID = uint64(1)
	staff := makeDoctorAssignedToClinic(t, db, clinicID, "readback rollback update before")
	base := NewReservationStaffRepository(db)
	repo := &failingReservationStaffReadbackRepository{
		ReservationStaffRepository: base,
		failFindCall:               1,
	}
	service := reservation.NewReservationStaffService(repo, NewTransactor(db), nil)
	updatedName := "readback rollback update after"

	updated, exclusions, err := service.Update(
		context.Background(),
		clinicID,
		staff.ID,
		&reservation.UpdateReservationStaffInput{Name: &updatedName},
	)

	require.ErrorIs(t, err, errReservationStaffReadback)
	assert.Nil(t, updated)
	assert.Nil(t, exclusions)
	assert.Equal(t, 1, repo.lockCalls, "ownership lock must succeed before readback failure")
	assert.Equal(t, 1, repo.findCalls, "failure must be injected into the only post-update readback")
	var reloaded model.Staff
	require.NoError(t, db.First(&reloaded, staff.ID).Error)
	assert.Equal(t, "readback rollback update before", reloaded.Name)
}

func TestReservationStaffService_PatchStatus_ExclusionReadbackFailureRollsBack(t *testing.T) {
	db := setupReservationStaffTxAtomicityTestDB(t)
	const clinicID = uint64(1)
	staff := makeDoctorAssignedToClinic(t, db, clinicID, "readback rollback patch")
	require.NoError(t, db.Model(&model.Staff{}).
		Where("id = ?", staff.ID).
		Update("is_active", true).Error)
	base := NewReservationStaffRepository(db)
	repo := &failingReservationStaffReadbackRepository{
		ReservationStaffRepository: base,
		failExcluded:               true,
	}
	service := reservation.NewReservationStaffService(repo, NewTransactor(db), nil)

	updated, exclusions, err := service.PatchStatus(
		context.Background(),
		clinicID,
		staff.ID,
		false,
	)

	require.ErrorIs(t, err, errReservationStaffReadback)
	assert.Nil(t, updated)
	assert.Nil(t, exclusions)
	var reloaded model.Staff
	require.NoError(t, db.First(&reloaded, staff.ID).Error)
	assert.True(t, reloaded.IsActive)
}
