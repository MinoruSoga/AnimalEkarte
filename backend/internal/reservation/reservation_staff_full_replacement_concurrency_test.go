package reservation

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type reservationStaffReplacementKind string

const (
	reservationStaffReplacementExclusion  reservationStaffReplacementKind = "exclusion"
	reservationStaffReplacementCapability reservationStaffReplacementKind = "capability"
)

func replaceReservationStaffTypes(
	ctx context.Context,
	repo ReservationStaffRepository,
	kind reservationStaffReplacementKind,
	clinicID, staffID, reservationTypeID uint64,
) error {
	if kind == reservationStaffReplacementExclusion {
		return repo.UpdateExcludedReservationTypes(
			ctx,
			clinicID,
			staffID,
			[]uint64{reservationTypeID},
		)
	}
	return repo.UpdateReservationCapabilities(
		ctx,
		clinicID,
		staffID,
		[]uint64{reservationTypeID},
	)
}

func reservationStaffReplacementTypeIDs(
	t *testing.T,
	db *gorm.DB,
	kind reservationStaffReplacementKind,
	clinicID, staffID uint64,
) []uint64 {
	t.Helper()
	typeIDs := make([]uint64, 0)
	query := db.Order("reservation_type_id ASC")
	// Stage B: exclusion facade persists capabilities only. For both kinds the
	// durable junction is staff_reservation_capabilities; exclusion path stores
	// capable = universe \ excluded, so callers assert the intended end-state
	// of that table (or derived excluded via FindAllExcluded in dedicated tests).
	var rows []model.StaffReservationCapability
	require.NoError(t, query.
		Where("clinic_id = ? AND staff_id = ?", clinicID, staffID).
		Find(&rows).Error)
	for i := range rows {
		typeIDs = append(typeIDs, rows[i].ReservationTypeID)
	}
	if kind == reservationStaffReplacementExclusion {
		// Requested excluded set is not stored; return capable IDs for SoT checks.
		return typeIDs
	}
	return typeIDs
}

func TestReservationStaffRepository_FullReplacementOwnsExclusiveMutationLock(t *testing.T) {
	tests := []struct {
		name string
		kind reservationStaffReplacementKind
	}{
		{name: "exclusions", kind: reservationStaffReplacementExclusion},
		{name: "capabilities", kind: reservationStaffReplacementCapability},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupReservationRepoTestDB(t)
			repo := NewReservationStaffRepository(db, nil)
			const clinicID = uint64(1)
			staff := makeDoctor(t, db, clinicID, "full replacement "+tt.name)
			require.NoError(t, db.Create(&model.StaffClinicAssignment{
				StaffID:  staff.ID,
				ClinicID: clinicID,
				IsMain:   true,
			}).Error)
			firstType := makeReservationType(t, db, clinicID)
			secondType := makeReservationType(t, db, clinicID)

			holder := db.Begin()
			require.NoError(t, holder.Error)
			defer holder.Rollback()
			require.NoError(t, replaceReservationStaffTypes(
				persistence.WithTxValue(context.Background(), holder),
				repo,
				tt.kind,
				clinicID,
				staff.ID,
				firstType.ID,
			))

			contender := db.Begin()
			require.NoError(t, contender.Error)
			defer contender.Rollback()
			require.NoError(t, contender.Exec("SET LOCAL lock_timeout = '100ms'").Error)
			contenderErr := replaceReservationStaffTypes(
				persistence.WithTxValue(context.Background(), contender),
				repo,
				tt.kind,
				clinicID,
				staff.ID,
				secondType.ID,
			)

			require.Error(t, contenderErr, "second full replacement must wait for exclusive ownership")
			assert.True(
				t,
				strings.Contains(contenderErr.Error(), "55P03") ||
					strings.Contains(contenderErr.Error(), "lock timeout"),
				"expected PostgreSQL lock timeout, got: %v",
				contenderErr,
			)
			require.NoError(t, contender.Rollback().Error)
			require.NoError(t, holder.Commit().Error)

			require.NoError(t, replaceReservationStaffTypes(
				context.Background(),
				repo,
				tt.kind,
				clinicID,
				staff.ID,
				secondType.ID,
			))
			got := reservationStaffReplacementTypeIDs(t, db, tt.kind, clinicID, staff.ID)
			if tt.kind == reservationStaffReplacementExclusion {
				// UpdateExcluded([second]) ⇒ capable = universe\{second} = {first}
				assert.Equal(t, []uint64{firstType.ID}, got,
					"exclusion facade must replace capabilities as universe\\excluded, not dual-write exclusions")
				var exclN int64
				require.NoError(t, db.Model(&model.StaffReservationExclusion{}).
					Where("staff_id = ?", staff.ID).Count(&exclN).Error)
				assert.Zero(t, exclN, "Stage B: no production rows in staff_reservation_exclusions")
			} else {
				assert.Equal(t, []uint64{secondType.ID}, got,
					"a full replacement must end at exactly one request, never their union")
			}
		})
	}
}
