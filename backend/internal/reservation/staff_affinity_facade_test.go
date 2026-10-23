package reservation

// staff_affinity_facade_test.go — TASK-021 Stage B integration: inverse mapping (staff_reservation_exclusions table no longer exists).

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	staffpkg "github.com/animal-ekarte/backend/internal/staff"
)

func TestStageB_ExclusionFacade_InverseRead(t *testing.T) {
	db := setupCapabilityIsolationTestDB(t)
	repo := NewReservationStaffRepository(db, staffpkg.NewRepository(db))
	ctx := context.Background()
	const clinicA = uint64(1)

	staff := makeDoctorAssignedToClinic(t, db, clinicA, "StageB facade staff")
	rt1 := makeReservationType(t, db, clinicA)
	rt2 := makeReservationType(t, db, clinicA)
	rt3 := makeReservationType(t, db, clinicA)

	// PUT excluded = {rt2} → capable = {rt1, rt3}
	require.NoError(t, repo.UpdateExcludedReservationTypes(ctx, clinicA, staff.ID, []uint64{rt2.ID}))

	caps, err := repo.FindAllReservationCapabilities(ctx, clinicA, staff.ID)
	require.NoError(t, err)
	gotCap := map[uint64]bool{}
	for _, c := range caps {
		gotCap[c.ReservationTypeID] = true
	}
	assert.True(t, gotCap[rt1.ID])
	assert.False(t, gotCap[rt2.ID])
	assert.True(t, gotCap[rt3.ID])
	assert.Len(t, caps, 2)

	// GET excluded facade = universe \ capable = {rt2}
	excluded, err := repo.FindAllExcludedReservationTypes(ctx, clinicA, staff.ID)
	require.NoError(t, err)
	require.Len(t, excluded, 1)
	assert.Equal(t, rt2.ID, excluded[0].ReservationTypeID)
	require.NotNil(t, excluded[0].ReservationType)

	// Direct capable PUT then derived excluded agrees
	require.NoError(t, repo.UpdateReservationCapabilities(ctx, clinicA, staff.ID, []uint64{rt1.ID}))
	excluded, err = repo.FindAllExcludedReservationTypes(ctx, clinicA, staff.ID)
	require.NoError(t, err)
	gotEx := map[uint64]bool{}
	for _, e := range excluded {
		gotEx[e.ReservationTypeID] = true
	}
	assert.False(t, gotEx[rt1.ID])
	assert.True(t, gotEx[rt2.ID])
	assert.True(t, gotEx[rt3.ID])

}

func TestStageB_EmptyCapable_DerivesFullUniverseExcluded(t *testing.T) {
	db := setupCapabilityIsolationTestDB(t)
	repo := NewReservationStaffRepository(db, staffpkg.NewRepository(db))
	ctx := context.Background()
	const clinicA = uint64(1)

	staff := makeDoctorAssignedToClinic(t, db, clinicA, "empty capable staff")
	rt1 := makeReservationType(t, db, clinicA)
	rt2 := makeReservationType(t, db, clinicA)

	excluded, err := repo.FindAllExcludedReservationTypes(ctx, clinicA, staff.ID)
	require.NoError(t, err)
	require.Len(t, excluded, 2, "no capabilities ⇒ fail-closed full universe excluded")
	ids := map[uint64]bool{}
	for _, e := range excluded {
		ids[e.ReservationTypeID] = true
	}
	assert.True(t, ids[rt1.ID])
	assert.True(t, ids[rt2.ID])
}
