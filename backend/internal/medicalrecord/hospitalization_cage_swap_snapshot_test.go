package medicalrecord

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type swapSnapshotPetRepository struct{}

func (swapSnapshotPetRepository) FindByID(context.Context, uint64, uint64) (*model.Pet, error) {
	panic("swap must use the locked pet read")
}

func (swapSnapshotPetRepository) LockByIDForUpdate(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
	return &model.Pet{ID: id, ClinicID: clinicID, Status: model.PetStatusAlive}, nil
}

type swapSnapshotCageRepository struct {
	CageMergeRepository
}

func (swapSnapshotCageRepository) FindByID(context.Context, uint64, uint64) (*model.Cage, error) {
	panic("swap must use the locked cage group")
}

func (swapSnapshotCageRepository) LockCageGroupForUpdate(_ context.Context, clinicID, id uint64) ([]model.Cage, error) {
	return []model.Cage{{ID: id, ClinicID: clinicID}}, nil
}

func TestHospitalizationSwapCages_RequiresDisplayedCageSnapshot(t *testing.T) {
	cage1, cage2, cage3, cage4 := uint64(1), uint64(2), uint64(3), uint64(4)
	for _, tc := range []struct {
		name          string
		currentA      *uint64
		currentB      *uint64
		reverse       bool
		foreignTarget bool
		wantConflict  bool
	}{
		{name: "unchanged snapshot", currentA: &cage1, currentB: &cage2},
		{name: "unchanged reversed input", currentA: &cage1, currentB: &cage2, reverse: true},
		{name: "target moved", currentA: &cage1, currentB: &cage3, wantConflict: true},
		{name: "source moved", currentA: &cage4, currentB: &cage2, wantConflict: true},
		{name: "target moved reversed input", currentA: &cage1, currentB: &cage3, reverse: true, wantConflict: true},
		{name: "both moved", currentA: &cage4, currentB: &cage3, wantConflict: true},
		{name: "source assignment cleared", currentA: nil, currentB: &cage2, wantConflict: true},
		{name: "target assignment cleared", currentA: &cage1, currentB: nil, wantConflict: true},
		{name: "foreign clinic target", currentA: &cage1, currentB: &cage2, foreignTarget: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const clinicID = uint64(1)
			rows := map[uint64]*model.Hospitalization{
				11: {ID: 11, ClinicID: clinicID, PetID: 21, Status: model.HospitalizationStatusAdmitted, CageID: tc.currentA},
				12: {ID: 12, ClinicID: clinicID, PetID: 22, Status: model.HospitalizationStatusAdmitted, CageID: tc.currentB},
			}
			if tc.foreignTarget {
				rows[12].ClinicID = 2
			}
			var lockedIDs []uint64
			writes, inTx := 0, false
			repo := &mockHospitalizationRepository{
				findByIDFn: func(_ context.Context, clinic, id uint64) (*model.Hospitalization, error) {
					require.True(t, inTx)
					require.Equal(t, clinicID, clinic)
					lockedIDs = append(lockedIDs, id)
					if rows[id].ClinicID != clinic {
						return nil, apperrors.WrapNotFound("hospitalization", "id")
					}
					return rows[id], nil // The current assignment after the row lock, not the board snapshot.
				},
				updateFieldsFn: func(_ context.Context, clinic, id uint64, input UpdateHospitalizationInput) (*model.Hospitalization, error) {
					require.True(t, inTx)
					require.Equal(t, clinicID, clinic)
					require.Equal(t, []uint64{11, 12}, lockedIDs, "both rows must be locked before writes")
					writes++
					updated := *rows[id]
					updated.CageID = input.CageID
					rows[id] = &updated
					return &updated, nil
				},
			}
			svc := &hospitalizationService{
				hospRepo: repo, petRepo: swapSnapshotPetRepository{}, cageRepo: swapSnapshotCageRepository{},
				transactor: &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
					inTx = true
					defer func() { inTx = false }()
					return fn(ctx)
				}},
			}
			idA, idB, expected := uint64(11), uint64(12), [2]uint64{cage1, cage2}
			if tc.reverse {
				idA, idB, expected = idB, idA, [2]uint64{cage2, cage1}
			}
			gotA, gotB, err := svc.SwapCages(context.Background(), clinicID, idA, idB, expected)
			assert.Equal(t, []uint64{11, 12}, lockedIDs, "request order must not change lock order")
			if tc.wantConflict || tc.foreignTarget {
				require.Error(t, err)
				if tc.foreignTarget {
					assert.True(t, apperrors.IsNotFound(err))
				} else {
					assert.True(t, apperrors.IsConflict(err), "stale cage assignment must return 409: %v", err)
				}
				assert.Nil(t, gotA)
				assert.Nil(t, gotB)
				assert.Zero(t, writes)
				assert.Equal(t, tc.currentA, rows[11].CageID)
				assert.Equal(t, tc.currentB, rows[12].CageID)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, gotA)
			require.NotNil(t, gotB)
			assert.Equal(t, idA, gotA.ID)
			assert.Equal(t, idB, gotB.ID)
			assert.Equal(t, expected[1], *gotA.CageID)
			assert.Equal(t, expected[0], *gotB.CageID)
			assert.Equal(t, 2, writes)
		})
	}
}

func TestHospitalizationSwapCages_RejectsMissingExpectedCages(t *testing.T) {
	for _, expected := range [][2]uint64{{0, 2}, {1, 0}, {0, 0}} {
		t.Run("missing expected cage", func(t *testing.T) {
			svc := &hospitalizationService{}
			gotA, gotB, err := svc.SwapCages(context.Background(), 1, 11, 12, expected)
			require.Error(t, err)
			assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
			assert.Nil(t, gotA)
			assert.Nil(t, gotB)
		})
	}
}
