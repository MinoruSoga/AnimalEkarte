package medicalrecord

// EMR-499 U1: the hospitalization read wire and the discharge movement side are derived from
// stay_cases / stay movement events (the EMR-499 core), not from the retired hospitalizations columns.
// These are database-free unit tests: they establish the mapping and the fail-closed rules only. They are
// not evidence of real DB locking, receipts or writer participation.

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func hospServiceWithStay(repo *mockHospitalizationRepository, cages cageFinder) *hospitalizationService {
	return &hospitalizationService{
		hospRepo:        repo,
		cageRepo:        cages,
		transactor:      &mockTransactor{},
		reservationRepo: &mockReservationRepository{assertOwnerInClinicFn: func(context.Context, uint64, uint64) error { return nil }, findPetOwnerInClinicFn: func(_ context.Context, _, _ uint64) (uint64, error) { return 2, nil }},
	}
}

func requireCoded422(t *testing.T, err error, code string) {
	t.Helper()
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, code, appErr.Code)
	require.ErrorIs(t, err, apperrors.ErrUnprocessable)
}

func TestLifecycleOf(t *testing.T) {
	for _, tc := range []struct {
		name  string
		state StayCageEffectiveState
		want  string
	}{
		{"reserved", StayCageEffectiveState{State: StayCageReserved}, lifecycleReserved},
		{"admitted", StayCageEffectiveState{State: StayCageAdmitted}, lifecycleAdmitted},
		{"released by discharge", StayCageEffectiveState{State: StayCageReleased, ReleaseKind: StayCageReleasedByDischarge}, lifecycleDischarged},
		{"released by cancel is not a discharge", StayCageEffectiveState{State: StayCageReleased, ReleaseKind: StayCageReleasedByCancel}, lifecycleCancelled},
		{"released by no-show is not a discharge", StayCageEffectiveState{State: StayCageReleased, ReleaseKind: StayCageReleasedByNoShow}, lifecycleNoShow},
		{"released with no known cause is unknown", StayCageEffectiveState{State: StayCageReleased}, lifecycleUnknown},
		{"unknown", StayCageEffectiveState{State: StayCageUnknown}, lifecycleUnknown},
		{"empty state is unknown", StayCageEffectiveState{}, lifecycleUnknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, lifecycleOf(tc.state))
		})
	}
}

func TestStayProjections(t *testing.T) {
	const clinic = uint64(7)
	entry := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	exit := entry.Add(72 * time.Hour)
	admitted := entry.Add(time.Hour)
	stayA, stayB, stayC := uint64(11), uint64(12), uint64(13)
	rows := []model.Hospitalization{
		{ID: 1, ClinicID: clinic, StayCaseID: &stayA},
		{ID: 2, ClinicID: clinic, StayCaseID: &stayB},
		{ID: 3, ClinicID: clinic},                         // legacy row without a stay case
		{ID: 4, ClinicID: clinic + 1, StayCaseID: &stayC}, // foreign clinic row is never projected
	}
	repo := &mockHospitalizationRepository{
		findStayCasesByIDsFn: func(_ context.Context, c uint64, ids []uint64) (map[uint64]model.StayCase, error) {
			require.Equal(t, clinic, c)
			require.Equal(t, []uint64{stayA, stayB}, ids, "only same-clinic, non-NULL stay IDs are read, deduplicated")
			return map[uint64]model.StayCase{
				stayA: {ID: stayA, ClinicID: clinic, StayKind: stayKindInpatient, PlannedEntryAt: &entry, PlannedExitAt: &exit},
				stayB: {ID: stayB, ClinicID: clinic, StayKind: "boarding"}, // unrecognised kind is not guessed
			}, nil
		},
	}
	cages := cageWriteCageFixture{read: func(_ context.Context, c uint64, ids []uint64) ([]StayCageEffectiveState, error) {
		return []StayCageEffectiveState{
			{ClinicID: c, StayCaseID: stayA, State: StayCageAdmitted, AdmittedAt: &admitted},
			{ClinicID: c, StayCaseID: stayB, State: StayCageUnknown},
		}, nil
	}}

	got, err := hospServiceWithStay(repo, cages).StayProjections(context.Background(), clinic, rows)

	require.NoError(t, err)
	require.Len(t, got, 4)
	a := got[1]
	require.NotNil(t, a.StayKind)
	assert.Equal(t, stayKindInpatient, *a.StayKind)
	assert.Equal(t, &entry, a.PlannedEntryAt, "planned interval is passed through, not converted to dates")
	assert.Equal(t, &exit, a.PlannedExitAt)
	assert.Equal(t, lifecycleAdmitted, a.Lifecycle)
	assert.Equal(t, &admitted, a.AdmittedAt)
	assert.Nil(t, a.DischargedAt)
	assert.Nil(t, got[2].StayKind, "an unrecognised stay kind is not mapped")
	assert.Equal(t, lifecycleUnknown, got[2].Lifecycle)
	assert.Equal(t, unknownHospitalizationStay(), got[3], "no stay case => unknown, nothing defaulted")
	assert.Equal(t, unknownHospitalizationStay(), got[4], "foreign clinic row => unknown")
}

func TestStayProjections_NoReaderIsUnknownAndReadFailureIsAnError(t *testing.T) {
	const clinic = uint64(7)
	stay := uint64(11)
	rows := []model.Hospitalization{{ID: 1, ClinicID: clinic, StayCaseID: &stay}}
	repo := &mockHospitalizationRepository{findStayCasesByIDsFn: func(context.Context, uint64, []uint64) (map[uint64]model.StayCase, error) {
		return map[uint64]model.StayCase{stay: {ID: stay, ClinicID: clinic, StayKind: stayKindHotel}}, nil
	}}

	t.Run("no effective-state reader wired => lifecycle unknown, kind still shown", func(t *testing.T) {
		got, err := hospServiceWithStay(repo, acceptAnyCageRepo()).StayProjections(context.Background(), clinic, rows)
		require.NoError(t, err)
		assert.Equal(t, lifecycleUnknown, got[1].Lifecycle)
		require.NotNil(t, got[1].StayKind)
	})
	t.Run("reader failure is returned, never turned into unknown", func(t *testing.T) {
		cages := cageWriteCageFixture{read: func(context.Context, uint64, []uint64) ([]StayCageEffectiveState, error) {
			return nil, errors.New("db down")
		}}
		got, err := hospServiceWithStay(repo, cages).StayProjections(context.Background(), clinic, rows)
		require.Error(t, err)
		assert.Nil(t, got)
	})
	t.Run("stay case read failure is returned", func(t *testing.T) {
		failing := &mockHospitalizationRepository{findStayCasesByIDsFn: func(context.Context, uint64, []uint64) (map[uint64]model.StayCase, error) {
			return nil, errors.New("db down")
		}}
		_, err := hospServiceWithStay(failing, acceptAnyCageRepo()).StayProjections(context.Background(), clinic, rows)
		require.Error(t, err)
	})
}

func TestListByLifecycleState(t *testing.T) {
	const clinic = uint64(7)
	ids := []uint64{21, 22, 23, 24}
	var rows []model.Hospitalization
	for i := range ids {
		stay := ids[i]
		rows = append(rows, model.Hospitalization{ID: uint64(i + 1), ClinicID: clinic, StayCaseID: &stay})
	}
	rows = append(rows, model.Hospitalization{ID: 5, ClinicID: clinic}) // legacy row, no stay case
	repo := &mockHospitalizationRepository{
		findAllFn: func(_ context.Context, _ uint64, _, _ *uint64, _, _ *string, page, limit int) ([]model.Hospitalization, int64, error) {
			require.Equal(t, 1, page)
			require.Equal(t, hospitalizationSnapshotMaxRows, limit, "the candidate set is read in full, never truncated")
			return rows, int64(len(rows)), nil
		},
		findStayCasesByIDsFn: func(_ context.Context, c uint64, in []uint64) (map[uint64]model.StayCase, error) {
			out := map[uint64]model.StayCase{}
			for _, id := range in {
				out[id] = model.StayCase{ID: id, ClinicID: c, StayKind: stayKindHotel}
			}
			return out, nil
		},
	}
	cages := cageWriteCageFixture{read: func(_ context.Context, c uint64, in []uint64) ([]StayCageEffectiveState, error) {
		states := map[uint64]StayCageOccupancyState{21: StayCageAdmitted, 22: StayCageAdmitted, 23: StayCageUnknown, 24: StayCageAdmitted}
		var out []StayCageEffectiveState
		for _, id := range in {
			out = append(out, StayCageEffectiveState{ClinicID: c, StayCaseID: id, State: states[id]})
		}
		return out, nil
	}}
	svc := hospServiceWithStay(repo, cages)
	admitted, unknown, bad := lifecycleAdmitted, lifecycleUnknown, "released"

	t.Run("only rows whose DERIVED state matches, paginated after filtering", func(t *testing.T) {
		page1, total, err := svc.List(context.Background(), clinic, nil, nil, &admitted, nil, nil, 1, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		require.Len(t, page1, 2)
		page2, total2, err := svc.List(context.Background(), clinic, nil, nil, &admitted, nil, nil, 2, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total2)
		require.Len(t, page2, 1)
		assert.Equal(t, uint64(4), page2[0].ID)
		past, _, err := svc.List(context.Background(), clinic, nil, nil, &admitted, nil, nil, 3, 2)
		require.NoError(t, err)
		assert.Empty(t, past)
	})
	t.Run("a page beyond the data is empty with the total preserved and an overflowing page cannot panic", func(t *testing.T) {
		// 一致 3 件 / limit 3 => 最終有効ページは 1。
		last, total, err := svc.List(context.Background(), clinic, nil, nil, &admitted, nil, nil, 1, 3)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, last, 3)
		// limit 1 => 最終有効ページは 3。その次のページは start == len(matched) で空になる。
		lastOfOne, _, err := svc.List(context.Background(), clinic, nil, nil, &admitted, nil, nil, 3, 1)
		require.NoError(t, err)
		require.Len(t, lastOfOne, 1)
		assert.Equal(t, uint64(4), lastOfOne[0].ID)
		pastOfOne, totalOfOne, err := svc.List(context.Background(), clinic, nil, nil, &admitted, nil, nil, 4, 1)
		require.NoError(t, err)
		assert.Empty(t, pastOfOne)
		assert.Equal(t, int64(3), totalOfOne)
		// (page-1)*100 が int を溢れる page でも panic せず、空ページと総数が返る。
		for _, page := range []int{2, 3, math.MaxInt/100 + 2, math.MaxInt} {
			got, gotTotal, err := svc.List(context.Background(), clinic, nil, nil, &admitted, nil, nil, page, 100)
			require.NoError(t, err)
			assert.Empty(t, got)
			assert.Equal(t, int64(3), gotTotal)
		}
	})
	t.Run("unknown rows (unresolvable history and legacy rows) match only lifecycle_state=unknown", func(t *testing.T) {
		got, total, err := svc.List(context.Background(), clinic, nil, nil, &unknown, nil, nil, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		require.Len(t, got, 2)
		assert.ElementsMatch(t, []uint64{3, 5}, []uint64{got[0].ID, got[1].ID})
	})
	t.Run("an invalid state is refused", func(t *testing.T) {
		_, _, err := svc.List(context.Background(), clinic, nil, nil, &bad, nil, nil, 1, 10)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})
	t.Run("a candidate set over the cap is refused, not truncated", func(t *testing.T) {
		over := &mockHospitalizationRepository{findAllFn: func(context.Context, uint64, *uint64, *uint64, *string, *string, int, int) ([]model.Hospitalization, int64, error) {
			return nil, int64(hospitalizationSnapshotMaxRows + 1), nil
		}}
		_, _, err := hospServiceWithStay(over, cages).List(context.Background(), clinic, nil, nil, &admitted, nil, nil, 1, 10)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})
}

func TestDischargeWithBilling_MovementSideFailsClosed(t *testing.T) {
	const clinic, id = uint64(7), uint64(10)
	stay := uint64(31)
	admitted := time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC)
	dischargeDay := admitted.Add(48 * time.Hour)
	newSvc := func(row *model.Hospitalization, state *StayCageEffectiveState, readErr error) (*hospitalizationService, *int) {
		locks := 0
		repo := &mockHospitalizationRepository{findByIDFn: func(context.Context, uint64, uint64) (*model.Hospitalization, error) {
			locks++
			return row, nil
		}}
		cages := cageWriteCageFixture{read: func(_ context.Context, c uint64, in []uint64) ([]StayCageEffectiveState, error) {
			if readErr != nil {
				return nil, readErr
			}
			if state == nil {
				return nil, nil
			}
			st := *state
			st.ClinicID, st.StayCaseID = c, in[0]
			return []StayCageEffectiveState{st}, nil
		}}
		return hospServiceWithStay(repo, cages), &locks
	}
	row := &model.Hospitalization{ID: id, ClinicID: clinic, OwnerID: 2, PetID: 5, StayCaseID: &stay}

	t.Run("create_accounting=true is refused before any read or write", func(t *testing.T) {
		svc, locks := newSvc(row, nil, nil)
		got, err := svc.DischargeWithBilling(context.Background(), clinic, id, DischargeWithBillingInput{DischargeDate: dischargeDay, CreateAccounting: true})
		require.Nil(t, got)
		requireCoded422(t, err, codeStayDischargeBillingUnavailable)
		assert.Zero(t, *locks, "nothing is locked or read when the billing builder is unavailable")
	})
	t.Run("a legacy row without a stay case is unknown (no backfill, no guess)", func(t *testing.T) {
		svc, _ := newSvc(&model.Hospitalization{ID: id, ClinicID: clinic, OwnerID: 2, PetID: 5}, nil, nil)
		_, err := svc.DischargeWithBilling(context.Background(), clinic, id, DischargeWithBillingInput{DischargeDate: dischargeDay})
		requireCoded422(t, err, codeStayLifecycleUnknown)
	})
	t.Run("an unresolved or missing stay state is unknown", func(t *testing.T) {
		for name, st := range map[string]*StayCageEffectiveState{
			"unknown":  {State: StayCageUnknown},
			"not read": nil,
		} {
			t.Run(name, func(t *testing.T) {
				svc, _ := newSvc(row, st, nil)
				_, err := svc.DischargeWithBilling(context.Background(), clinic, id, DischargeWithBillingInput{DischargeDate: dischargeDay})
				requireCoded422(t, err, codeStayLifecycleUnknown)
			})
		}
	})
	t.Run("reader failure is surfaced", func(t *testing.T) {
		svc, _ := newSvc(row, nil, errors.New("db down"))
		_, err := svc.DischargeWithBilling(context.Background(), clinic, id, DischargeWithBillingInput{DischargeDate: dischargeDay})
		require.Error(t, err)
		assert.NotErrorIs(t, err, apperrors.ErrUnprocessable)
	})
	t.Run("already discharged is a 400", func(t *testing.T) {
		svc, _ := newSvc(row, &StayCageEffectiveState{State: StayCageReleased, ReleaseKind: StayCageReleasedByDischarge}, nil)
		_, err := svc.DischargeWithBilling(context.Background(), clinic, id, DischargeWithBillingInput{DischargeDate: dischargeDay})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})
	t.Run("a cancelled or no-show stay cannot be discharged", func(t *testing.T) {
		for _, kind := range []StayCageReleaseKind{StayCageReleasedByCancel, StayCageReleasedByNoShow} {
			svc, _ := newSvc(row, &StayCageEffectiveState{State: StayCageReleased, ReleaseKind: kind}, nil)
			_, err := svc.DischargeWithBilling(context.Background(), clinic, id, DischargeWithBillingInput{DischargeDate: dischargeDay})
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err))
		}
	})
	t.Run("a stay not yet admitted is refused", func(t *testing.T) {
		svc, _ := newSvc(row, &StayCageEffectiveState{State: StayCageReserved}, nil)
		_, err := svc.DischargeWithBilling(context.Background(), clinic, id, DischargeWithBillingInput{DischargeDate: dischargeDay})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})
	t.Run("discharge before the actual admission day is refused", func(t *testing.T) {
		svc, _ := newSvc(row, &StayCageEffectiveState{State: StayCageAdmitted, AdmittedAt: &admitted}, nil)
		_, err := svc.DischargeWithBilling(context.Background(), clinic, id, DischargeWithBillingInput{DischargeDate: admitted.AddDate(0, 0, -1)})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})
	t.Run("a proven admitted stay still gets the stable writer-unavailable 422 (no movement event is written)", func(t *testing.T) {
		svc, _ := newSvc(row, &StayCageEffectiveState{State: StayCageAdmitted, AdmittedAt: &admitted}, nil)
		got, err := svc.DischargeWithBilling(context.Background(), clinic, id, DischargeWithBillingInput{DischargeDate: dischargeDay})
		require.Nil(t, got)
		requireCoded422(t, err, codeStayDischargeWriterUnavailable)
	})
}

func TestStayCageEffectiveState_DescribesReleaseCauseAndActualTimes(t *testing.T) {
	admit := time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC)
	out := admit.Add(24 * time.Hour)
	hosp := uint64(7)
	stay := model.StayCase{ID: 11, ClinicID: 3, StayKind: "hotel", Version: 1}
	admitEvent := model.StayMovementEvent{ID: 1, ClinicID: 3, StayCaseID: 11, Revision: 1, EventKind: "admit", ActualAt: &admit, HospitalizationID: &hosp}
	discharge := model.StayMovementEvent{ID: 2, ClinicID: 3, StayCaseID: 11, Revision: 2, EventKind: "discharge", ActualAt: &out}

	got := projectStayCageEffectiveState(stay, []model.StayMovementEvent{admitEvent})
	require.Equal(t, StayCageAdmitted, got.State)
	require.NotNil(t, got.AdmittedAt)
	assert.True(t, got.AdmittedAt.Equal(admit))
	assert.Nil(t, got.DischargedAt)
	assert.Empty(t, got.ReleaseKind)

	got = projectStayCageEffectiveState(stay, []model.StayMovementEvent{admitEvent, discharge})
	require.Equal(t, StayCageReleased, got.State)
	assert.Equal(t, StayCageReleasedByDischarge, got.ReleaseKind)
	require.NotNil(t, got.DischargedAt)
	assert.True(t, got.DischargedAt.Equal(out))

	for kind, want := range map[string]StayCageReleaseKind{"cancel": StayCageReleasedByCancel, "no_show": StayCageReleasedByNoShow} {
		got = projectStayCageEffectiveState(stay, []model.StayMovementEvent{{ID: 3, ClinicID: 3, StayCaseID: 11, Revision: 1, EventKind: kind}})
		require.Equal(t, StayCageReleased, got.State)
		assert.Equal(t, want, got.ReleaseKind)
		assert.Nil(t, got.AdmittedAt)
		assert.Nil(t, got.DischargedAt)
	}

	got = projectStayCageEffectiveState(stay, nil)
	assert.Equal(t, StayCageUnknown, got.State)
	assert.Empty(t, got.ReleaseKind, "unknown carries no cause or time")
	assert.Nil(t, got.AdmittedAt)
}
