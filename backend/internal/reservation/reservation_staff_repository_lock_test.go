package reservation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestReservationStaffCapabilityValidation_HoldsShareLocks(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationStaffRepository(db, nil)
	ctx := context.Background()
	const clinicID = uint64(1)

	staff := makeDoctor(t, db, clinicID, "予約対応能力共有ロック獣医")
	assignment := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicID, IsMain: true}
	require.NoError(t, db.Create(assignment).Error)
	reservationType := makeReservationType(t, db, clinicID)
	require.NoError(t, db.Create(&model.StaffReservationCapability{
		ClinicID:          clinicID,
		StaffID:           staff.ID,
		ReservationTypeID: reservationType.ID,
	}).Error)

	locked := make(chan struct{})
	release := make(chan struct{})
	holderDone := make(chan error, 1)
	go func() {
		holderDone <- testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			if err := ValidateReservationStaffCapability(txCtx, repo, clinicID, &staff.ID, reservationType.ID); err != nil {
				return err
			}
			close(locked)
			<-release
			return nil
		})
	}()
	<-locked

	mutationDone := make(chan error, 1)
	go func() {
		mutationDone <- repo.UpdateReservationCapabilities(ctx, clinicID, staff.ID, nil)
	}()

	var mutationErr error
	completedEarly := false
	select {
	case mutationErr = <-mutationDone:
		completedEarly = true
		close(release)
		t.Errorf("capability replacement was not serialized: err=%v", mutationErr)
	case <-time.After(100 * time.Millisecond):
		close(release)
	}
	require.NoError(t, <-holderDone)
	if !completedEarly {
		mutationErr = <-mutationDone
	}
	require.NoError(t, mutationErr)
}
