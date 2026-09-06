package reservation

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/model"
	staffpkg "github.com/animal-ekarte/backend/internal/staff"
)

type reservationJunctionKind string

const (
	reservationJunctionExclusion  reservationJunctionKind = "exclusion"
	reservationJunctionCapability reservationJunctionKind = "capability"
)

func replaceReservationJunction(
	ctx context.Context,
	repo ReservationStaffRepository,
	kind reservationJunctionKind,
	clinicID, staffID, typeID uint64,
) error {
	if kind == reservationJunctionExclusion {
		return repo.UpdateExcludedReservationTypes(ctx, clinicID, staffID, []uint64{typeID})
	}
	return repo.UpdateReservationCapabilities(ctx, clinicID, staffID, []uint64{typeID})
}

func countReservationJunction(
	t *testing.T,
	db *gorm.DB,
	kind reservationJunctionKind,
	staffID uint64,
) int64 {
	t.Helper()
	var count int64
	// Stage B: exclusion facade writes capabilities only. Both kinds count capabilities.
	_ = kind
	require.NoError(t, db.Model(&model.StaffReservationCapability{}).
		Where("staff_id = ?", staffID).
		Count(&count).Error)
	return count
}

func revokeStaffAssignmentWithIdentityLock(
	db *gorm.DB,
	clinicID, staffID uint64,
) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var staff model.Staff
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND deleted_at IS NULL", staffID).
			First(&staff).Error; err != nil {
			return err
		}
		result := tx.
			Where("staff_id = ? AND clinic_id = ? AND deleted_at IS NULL", staffID, clinicID).
			Delete(&model.StaffClinicAssignment{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return fmt.Errorf("revoke assignment: got %d rows, want 1", result.RowsAffected)
		}
		return nil
	})
}

func TestReservationStaffRepository_JunctionReplacement_RevocationWins(t *testing.T) {
	for _, kind := range []reservationJunctionKind{
		reservationJunctionExclusion,
		reservationJunctionCapability,
	} {
		t.Run(string(kind), func(t *testing.T) {
			db := setupReservationStaffTxAtomicityTestDB(t)
			repo := NewReservationStaffRepository(db, staffpkg.NewRepository(db))
			const clinicID = uint64(1)
			staff := makeDoctorAssignedToClinic(t, db, clinicID, "revocation wins "+string(kind))
			reservationType := makeReservationType(t, db, clinicID)

			revocationTx := db.Begin()
			require.NoError(t, revocationTx.Error)
			defer revocationTx.Rollback()
			var locked model.Staff
			require.NoError(t, revocationTx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ? AND deleted_at IS NULL", staff.ID).
				First(&locked).Error)
			require.NoError(t, revocationTx.
				Where("staff_id = ? AND clinic_id = ?", staff.ID, clinicID).
				Delete(&model.StaffClinicAssignment{}).Error)

			writeResult := make(chan error, 1)
			go func() {
				writeResult <- replaceReservationJunction(
					context.Background(),
					repo,
					kind,
					clinicID,
					staff.ID,
					reservationType.ID,
				)
			}()

			select {
			case err := <-writeResult:
				t.Fatalf("junction writer bypassed the pending revocation lock: %v", err)
			case <-time.After(100 * time.Millisecond):
			}
			require.NoError(t, revocationTx.Commit().Error)

			err := <-writeResult
			require.Error(t, err)
			assert.Zero(t, countReservationJunction(t, db, kind, staff.ID))
		})
	}
}

func TestReservationStaffRepository_JunctionReplacement_WriteWins(t *testing.T) {
	for _, kind := range []reservationJunctionKind{
		reservationJunctionExclusion,
		reservationJunctionCapability,
	} {
		t.Run(string(kind), func(t *testing.T) {
			db := setupReservationStaffTxAtomicityTestDB(t)
			repo := NewReservationStaffRepository(db, staffpkg.NewRepository(db))
			const clinicID = uint64(1)
			staff := makeDoctorAssignedToClinic(t, db, clinicID, "write wins "+string(kind))
			reservationType := makeReservationType(t, db, clinicID)
			// Extra type so exclusion facade leave ≥1 capability row when excluding one type.
			otherType := makeReservationType(t, db, clinicID)
			_ = otherType
			writerReady := make(chan struct{})
			releaseWriter := make(chan struct{})
			writerResult := make(chan error, 1)

			go func() {
				writerResult <- testNewTransactor(db).WithTx(
					context.Background(),
					func(txCtx context.Context) error {
						if err := replaceReservationJunction(
							txCtx,
							repo,
							kind,
							clinicID,
							staff.ID,
							reservationType.ID,
						); err != nil {
							return err
						}
						close(writerReady)
						<-releaseWriter
						return nil
					},
				)
			}()
			<-writerReady

			revokeResult := make(chan error, 1)
			go func() {
				revokeResult <- revokeStaffAssignmentWithIdentityLock(db, clinicID, staff.ID)
			}()
			select {
			case err := <-revokeResult:
				close(releaseWriter)
				require.NoError(t, <-writerResult)
				t.Fatalf("assignment revocation bypassed the junction writer lock: %v", err)
			case <-time.After(100 * time.Millisecond):
			}

			close(releaseWriter)
			require.NoError(t, <-writerResult)
			require.NoError(t, <-revokeResult)
			// capability path: capable={reservationType} → count 1
			// exclusion facade: exclude reservationType → capable={otherType} → count 1
			assert.EqualValues(t, 1, countReservationJunction(t, db, kind, staff.ID))

			var activeAssignments int64
			require.NoError(t, db.Model(&model.StaffClinicAssignment{}).
				Where("staff_id = ? AND clinic_id = ?", staff.ID, clinicID).
				Count(&activeAssignments).Error)
			assert.Zero(t, activeAssignments)
		})
	}
}
