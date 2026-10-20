package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type fixedStayCageReader struct {
	readFn func(context.Context, uint64, []uint64) ([]StayCageEffectiveState, error)
}

func (r fixedStayCageReader) ReadEffectiveStatesForCageAssignment(ctx context.Context, clinicID uint64, ids []uint64) ([]StayCageEffectiveState, error) {
	return r.readFn(ctx, clinicID, ids)
}

func TestEffectiveStayCageAdapter_PreservesCoreDecisionAndSelfExclusion(t *testing.T) {
	stay1, stay2, stay3 := uint64(11), uint64(12), uint64(13)
	rows := []stayCageAssignment{
		{HospitalizationID: 1, ClinicID: 7, StayCaseID: &stay1, CageID: 10},
		{HospitalizationID: 2, ClinicID: 7, StayCaseID: &stay2, CageID: 10},
		{HospitalizationID: 3, ClinicID: 7, StayCaseID: &stay3, CageID: 20},
	}
	for _, tc := range []struct {
		name  string
		state StayCageOccupancyState
		want  map[uint64]int64
	}{
		{name: "reserved", state: StayCageReserved, want: map[uint64]int64{10: 1}},
		{name: "admitted beyond planned end remains occupied", state: StayCageAdmitted, want: map[uint64]int64{10: 1}},
		{name: "core resolved discharge cancel or correction releases", state: StayCageReleased, want: map[uint64]int64{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reader := fixedStayCageReader{readFn: func(_ context.Context, clinic uint64, ids []uint64) ([]StayCageEffectiveState, error) {
				require.Equal(t, uint64(7), clinic)
				require.Equal(t, []uint64{12, 13}, ids)
				return []StayCageEffectiveState{
					{ClinicID: 7, StayCaseID: 12, State: tc.state, EvidenceVersion: 5},
					{ClinicID: 7, StayCaseID: 13, State: StayCageReleased, EvidenceVersion: 4},
				}, nil
			}}
			got, err := countEffectiveStayCages(t.Context(), 7, rows, []uint64{1}, reader)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
func TestEffectiveStayCageAdapter_UnresolvedNeverMeansVacant(t *testing.T) {
	stay := uint64(11)
	base := stayCageAssignment{HospitalizationID: 1, ClinicID: 7, StayCaseID: &stay, CageID: 10}
	for _, mode := range []string{"nil-reader", "reader-error", "missing-state", "unknown", "invalid-state", "wrong-clinic", "zero-evidence", "duplicate-state", "foreign-state", "missing-stay", "duplicate-assignment"} {
		t.Run(mode, func(t *testing.T) {
			rows := []stayCageAssignment{base}
			states := []StayCageEffectiveState{{ClinicID: 7, StayCaseID: 11, State: StayCageReserved, EvidenceVersion: 1}}
			var readErr error
			switch mode {
			case "reader-error":
				readErr = errors.New("core unavailable")
			case "missing-state":
				states = nil
			case "unknown":
				states[0].State = StayCageUnknown
			case "invalid-state":
				states[0].State = "correction"
			case "wrong-clinic":
				states[0].ClinicID = 8
			case "zero-evidence":
				states[0].EvidenceVersion = 0
			case "duplicate-state":
				states = append(states, states[0])
			case "foreign-state":
				states[0].StayCaseID = 99
			case "missing-stay":
				rows[0].StayCaseID = nil
			case "duplicate-assignment":
				rows = append(rows, base)
			}
			var reader StayCageEffectiveStateReader = fixedStayCageReader{readFn: func(context.Context, uint64, []uint64) ([]StayCageEffectiveState, error) { return states, readErr }}
			if mode == "nil-reader" {
				reader = nil
			}
			got, err := countEffectiveStayCages(t.Context(), 7, rows, nil, reader)
			require.Error(t, err)
			require.Nil(t, got)
			var unknown *cageOccupancyUnknown
			require.Equal(t, mode != "nil-reader" && mode != "reader-error", errors.As(err, &unknown))
		})
	}
}
func TestEffectiveStayCageAdapter_PreservesCallerTransactionContext(t *testing.T) {
	type transactionMarker struct{}
	ctx := context.WithValue(t.Context(), transactionMarker{}, "caller-transaction")
	stay := uint64(11)
	reader := fixedStayCageReader{readFn: func(received context.Context, clinic uint64, ids []uint64) ([]StayCageEffectiveState, error) {
		require.Same(t, ctx, received)
		return []StayCageEffectiveState{{ClinicID: 7, StayCaseID: 11, State: StayCageAdmitted, EvidenceVersion: 1}}, nil
	}}
	got, err := countEffectiveStayCages(ctx, 7, []stayCageAssignment{{HospitalizationID: 1, ClinicID: 7, StayCaseID: &stay, CageID: 10}}, nil, reader)
	require.NoError(t, err)
	require.Equal(t, map[uint64]int64{10: 1}, got)
}
