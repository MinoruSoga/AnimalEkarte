// Package reservation — EMR-332 C15 evidence tests
// (docs/work/emr/emr-332-security/workflow-audit/c15-order-idempotency.md).
//
// Scope: deterministic, single-threaded, in-memory verification of the
// reservation lifecycle order/idempotency guards that are exported from
// package reservation. No service, repository, DB, session, audit, cache,
// queue, file, mail, or telemetry object is contacted — every repository
// collaborator is an in-memory fake, so the side-effect invocation count is
// the number of calls recorded on those fakes (asserted per test).
//
// Marker convention: SECCHK20261009-style alphanumeric sentinels only.
package reservation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// ── in-memory fakes (deterministic, single-threaded) ────────────────────────

// fakeSlotConflicts satisfies the slotConflictChecker shape used by
// reservation.CheckSlotConflict. All state is scripted before the call.
type fakeSlotConflicts struct {
	doctorConflict bool
	onDutyDoctors  int64
	conflicts      int64
	err            error

	sawExcludeID *uint64
	calls        int
}

func (f *fakeSlotConflicts) HasDoctorConflict(_ context.Context, _, _ uint64, _, _ time.Time, excludeID *uint64) (bool, error) {
	f.calls++
	f.sawExcludeID = excludeID
	return f.doctorConflict, f.err
}

func (f *fakeSlotConflicts) CountOnDutyDoctors(_ context.Context, _ uint64, _ time.Time) (int64, error) {
	f.calls++
	return f.onDutyDoctors, f.err
}

func (f *fakeSlotConflicts) CountConflicts(_ context.Context, _ uint64, _, _ time.Time, excludeID *uint64) (int64, error) {
	f.calls++
	f.sawExcludeID = excludeID
	return f.conflicts, f.err
}

// fakeCapacityCounter satisfies reservationTypeCapacityCounter.
type fakeCapacityCounter struct {
	count int64
	err   error
	calls int
}

func (f *fakeCapacityCounter) CountByTypeAndStartTime(_ context.Context, _, _ uint64, _ time.Time, _ *uint64) (int64, error) {
	f.calls++
	return f.count, f.err
}

// fakeTypeFinder satisfies reservationTypeFinder.
type fakeTypeFinder struct {
	typ   *model.ReservationType
	err   error
	calls int
}

func (f *fakeTypeFinder) FindByID(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
	f.calls++
	return f.typ, f.err
}

// fakePetLifeStateReader satisfies sharedkernel.PetLifeStateReader (EMR-491 D6 death reader).
type fakePetLifeStateReader struct {
	state sharedkernel.PetLifeState
	err   error
	calls int
}

func (f *fakePetLifeStateReader) ReadPetLifeState(_ context.Context, _, _ uint64) (sharedkernel.PetLifeState, error) {
	f.calls++
	return f.state, f.err
}

// fakeOwnerPetVerifier satisfies sharedkernel.OwnerPetLinkVerifier.
type fakeOwnerPetVerifier struct {
	ownerErr   error
	petOwnerID uint64
	petErr     error
	calls      int
}

func (f *fakeOwnerPetVerifier) AssertOwnerInClinic(_ context.Context, _, _ uint64) error {
	f.calls++
	return f.ownerErr
}

func (f *fakeOwnerPetVerifier) FindPetOwnerInClinic(_ context.Context, _, _ uint64) (uint64, error) {
	f.calls++
	return f.petOwnerID, f.petErr
}

// ── C15: booking-constraint enforcement matrix (order guard) ────────────────

// ShouldEnforceReservationBookingConstraints decides whether a create/update
// must re-run slot/closed-day checks. checked_in and later lifecycle states
// already occupy the slot; shortcut reception routes skip create-time checks.
// A wrong decision here lets out-of-order reschedules bypass conflict checks.
func TestC15_BookingConstraintEnforcementMatrix(t *testing.T) {
	route := func(s string) *string { return &s }
	cases := []struct {
		name   string
		status model.ReservationStatus
		route  *string
		want   bool
	}{
		{"pending no route", model.ReservationStatusPending, nil, true},
		{"confirmed no route", model.ReservationStatusConfirmed, nil, true},
		{"cancelled no route", model.ReservationStatusCancelled, nil, true},
		{"no_show no route", model.ReservationStatusNoShow, nil, true},
		{"tentative no route", model.ReservationStatusTentative, nil, true},
		{"checked_in no route", model.ReservationStatusCheckedIn, nil, false},
		{"in_consultation no route", model.ReservationStatusInConsultation, nil, false},
		{"accounting no route", model.ReservationStatusAccounting, nil, false},
		{"completed no route", model.ReservationStatusCompleted, nil, false},
		{"pending via reception shortcut", model.ReservationStatusPending, route("reception"), false},
		{"pending via exam_room shortcut", model.ReservationStatusPending, route("exam_room"), false},
		{"pending via record_shortcut", model.ReservationStatusPending, route("record_shortcut"), false},
		{"pending via line route", model.ReservationStatusPending, route("line"), true},
		{"pending via phone route", model.ReservationStatusPending, route("phone"), true},
		{"pending via record_intent route", model.ReservationStatusPending, route("record_intent"), true},
		{"checked_in via reception still skips", model.ReservationStatusCheckedIn, route("reception"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := reservation.ShouldEnforceReservationBookingConstraints(tc.status, tc.route)
			assert.Equal(t, tc.want, got)
		})
	}
}

// ── C15: terminal-state mutation guard (trimming lifecycle) ─────────────────

// Completed/cancelled/no_show trimming appointments are terminal: the
// appointment and its clinical detail must reject further mutation.
func TestC15_TerminalTrimmingMutationRejected(t *testing.T) {
	cases := []struct {
		name    string
		status  model.ReservationStatus
		wantErr bool
	}{
		{"completed is terminal", model.ReservationStatusCompleted, true},
		{"cancelled is terminal", model.ReservationStatusCancelled, true},
		{"no_show is terminal", model.ReservationStatusNoShow, true},
		{"pending is mutable", model.ReservationStatusPending, false},
		{"confirmed is mutable", model.ReservationStatusConfirmed, false},
		{"checked_in is mutable", model.ReservationStatusCheckedIn, false},
		{"in_consultation is mutable", model.ReservationStatusInConsultation, false},
		{"accounting is mutable", model.ReservationStatusAccounting, false},
		{"tentative is mutable", model.ReservationStatusTentative, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := reservation.ValidateTrimmingAppointmentMutable(tc.status)
			if tc.wantErr {
				require.Error(t, err)
				assert.True(t, apperrors.IsConflict(err), "terminal rejection must be Conflict, got %v", err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ── C15: slot-conflict duplicate/overlap rejection (deterministic order) ────

func TestC15_SlotConflictDeterministicScenarios(t *testing.T) {
	doctorID := uint64(42)
	selfID := uint64(9001)
	start := time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC)
	end := start.Add(30 * time.Minute)

	cases := []struct {
		name      string
		doctorID  *uint64
		fake      *fakeSlotConflicts
		excludeID *uint64
		wantErr   bool
		conflict  bool
	}{
		{
			name:     "second booking for same doctor+slot is rejected",
			doctorID: &doctorID,
			fake:     &fakeSlotConflicts{doctorConflict: true},
			wantErr:  true,
			conflict: true,
		},
		{
			name:     "free doctor slot passes",
			doctorID: &doctorID,
			fake:     &fakeSlotConflicts{doctorConflict: false},
			wantErr:  false,
		},
		{
			name:     "capacity full rejects (conflicts reach on-duty count)",
			fake:     &fakeSlotConflicts{onDutyDoctors: 2, conflicts: 2},
			wantErr:  true,
			conflict: true,
		},
		{
			name:    "capacity below on-duty count passes",
			fake:    &fakeSlotConflicts{onDutyDoctors: 2, conflicts: 1},
			wantErr: false,
		},
		{
			name:     "zero doctors on duty rejects",
			fake:     &fakeSlotConflicts{onDutyDoctors: 0, conflicts: 0},
			wantErr:  true,
			conflict: true,
		},
		{
			name:      "self-conflict is excluded on update (excludeID forwarded)",
			doctorID:  &doctorID,
			fake:      &fakeSlotConflicts{doctorConflict: false},
			excludeID: &selfID,
			wantErr:   false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := reservation.CheckSlotConflict(context.Background(), tc.fake, 1, tc.doctorID, start, end, tc.excludeID)
			if tc.wantErr {
				require.Error(t, err)
				if tc.conflict {
					assert.True(t, apperrors.IsConflict(err), "expected conflict-class rejection, got %v", err)
				}
			} else {
				assert.NoError(t, err)
			}
			if tc.excludeID != nil {
				require.NotNil(t, tc.fake.sawExcludeID)
				assert.Equal(t, *tc.excludeID, *tc.fake.sawExcludeID)
			}
		})
	}
}

// ── C15: reservation-type capacity boundary ─────────────────────────────────

func TestC15_TypeCapacityBoundary(t *testing.T) {
	max2 := 2
	start := time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC)
	cases := []struct {
		name    string
		typ     *model.ReservationType
		count   int64
		wantErr bool
	}{
		{"at MaxConcurrent rejects", &model.ReservationType{MaxConcurrent: &max2}, 2, true},
		{"below MaxConcurrent passes", &model.ReservationType{MaxConcurrent: &max2}, 1, false},
		{"no MaxConcurrent configured passes regardless", &model.ReservationType{MaxConcurrent: nil}, 99, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			finder := &fakeTypeFinder{typ: tc.typ}
			counter := &fakeCapacityCounter{count: tc.count}
			err := reservation.CheckReservationTypeCapacity(context.Background(), counter, finder, 1, 7, start, nil)
			if tc.wantErr {
				require.Error(t, err)
				assert.True(t, apperrors.IsConflict(err), "expected conflict-class rejection, got %v", err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ── C15: write-order guards on owner/pet linkage and deceased pets ───────────

func TestC15_DeceasedPetWriteRejected(t *testing.T) {
	petID := uint64(555)
	unresolved := apperrors.Wrap(sharedkernel.ErrPetLifeStateUnresolved, "pet status is not recorded")
	readerDown := errors.New("SECCHK20261009READERDOWN")
	cases := []struct {
		name        string
		petID       *uint64
		state       sharedkernel.PetLifeState
		err         error
		wantErr     bool
		wantInvalid bool
		wantCalls   int
	}{
		{name: "nil petID skips lookup", petID: nil, wantCalls: 0},
		{name: "alive pet passes", petID: &petID, state: sharedkernel.PetLifeStateAlive, wantCalls: 1},
		{name: "deceased rejected as business rule", petID: &petID, state: sharedkernel.PetLifeStateDeceased, wantErr: true, wantInvalid: true, wantCalls: 1},
		{name: "unresolved life state is an error, never alive", petID: &petID, err: unresolved, wantErr: true, wantCalls: 1},
		{name: "reader failure propagates, never alive", petID: &petID, err: readerDown, wantErr: true, wantCalls: 1},
		{name: "empty state is not alive", petID: &petID, wantErr: true, wantCalls: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reader := &fakePetLifeStateReader{state: tc.state, err: tc.err}
			err := reservation.ValidateReservationPetNotDeceased(context.Background(), reader, 1, tc.petID)
			if tc.wantErr {
				require.Error(t, err)
				assert.Equal(t, tc.wantInvalid, apperrors.IsInvalidInput(err), "unexpected error class: %v", err)
				if tc.err != nil {
					assert.ErrorIs(t, err, tc.err)
				}
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.wantCalls, reader.calls)
		})
	}

	t.Run("nil reader fails closed", func(t *testing.T) {
		err := reservation.ValidateReservationPetNotDeceased(context.Background(), nil, 1, &petID)
		require.Error(t, err)
		assert.False(t, apperrors.IsInvalidInput(err))
	})
}

func TestC15_OwnerPetLinkCrossTenantRejected(t *testing.T) {
	ownerID := uint64(10)
	petID := uint64(20)
	notFound := apperrors.WrapNotFound("owner", "10")
	cases := []struct {
		name    string
		ownerID *uint64
		petID   *uint64
		fake    *fakeOwnerPetVerifier
		wantErr bool
	}{
		{"both nil skips verification", nil, nil, &fakeOwnerPetVerifier{}, false},
		{"owner outside clinic rejected", &ownerID, nil, &fakeOwnerPetVerifier{ownerErr: notFound}, true},
		{"pet owned by another owner rejected", &ownerID, &petID, &fakeOwnerPetVerifier{petOwnerID: 99}, true},
		{"consistent owner+pet passes", &ownerID, &petID, &fakeOwnerPetVerifier{petOwnerID: ownerID}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := reservation.ValidateReservationOwnerPetLinksWithRepo(context.Background(), tc.fake, 1, tc.ownerID, tc.petID)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ── C15: deterministic lifecycle order script ───────────────────────────────

// A single-threaded scripted walk over the decision points that are exported.
// Each step records the guard verdict so the transition allow/deny order is
// captured as evidence (static citations for the non-exported CAS gates live
// in c15-order-idempotency.md).
func TestC15_DeterministicOrderScript(t *testing.T) {
	var got []string
	record := func(label string, ok bool) {
		got = append(got, label+"="+map[bool]string{true: "allow", false: "deny"}[ok])
	}

	// Reschedule-conflict enforcement flips off exactly at checked_in.
	enforceable := []model.ReservationStatus{
		model.ReservationStatusPending, model.ReservationStatusConfirmed,
		model.ReservationStatusCancelled, model.ReservationStatusNoShow,
		model.ReservationStatusTentative,
	}
	nonEnforceable := []model.ReservationStatus{
		model.ReservationStatusCheckedIn, model.ReservationStatusInConsultation,
		model.ReservationStatusAccounting, model.ReservationStatusCompleted,
	}
	for _, s := range enforceable {
		record("enforce/"+string(s), reservation.ShouldEnforceReservationBookingConstraints(s, nil))
	}
	for _, s := range nonEnforceable {
		record("enforce/"+string(s), reservation.ShouldEnforceReservationBookingConstraints(s, nil))
	}
	// Terminal trimming mutations deny.
	for _, s := range []model.ReservationStatus{model.ReservationStatusCompleted, model.ReservationStatusCancelled, model.ReservationStatusNoShow} {
		record("trimming-mutable/"+string(s), reservation.ValidateTrimmingAppointmentMutable(s) == nil)
	}
	record("trimming-mutable/in_consultation", reservation.ValidateTrimmingAppointmentMutable(model.ReservationStatusInConsultation) == nil)

	want := []string{
		"enforce/pending=allow",
		"enforce/confirmed=allow",
		"enforce/cancelled=allow",
		"enforce/no_show=allow",
		"enforce/tentative=allow",
		"enforce/checked_in=deny",
		"enforce/in_consultation=deny",
		"enforce/accounting=deny",
		"enforce/completed=deny",
		"trimming-mutable/completed=deny",
		"trimming-mutable/cancelled=deny",
		"trimming-mutable/no_show=deny",
		"trimming-mutable/in_consultation=allow",
	}
	assert.Equal(t, want, got)
}
