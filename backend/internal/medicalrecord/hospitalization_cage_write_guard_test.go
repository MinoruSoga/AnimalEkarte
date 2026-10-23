package medicalrecord

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// These fixtures exercise production save entry points without a database.
// They do not establish real database locking, rollback, or writer participation.
type cageWriteCageFixture struct {
	CageMergeRepository
	lock  func(context.Context, uint64, uint64) ([]model.Cage, error)
	read  func(context.Context, uint64, []uint64) ([]StayCageEffectiveState, error)
	count func(context.Context, uint64, []uint64, []uint64) (map[uint64]int64, error)
}

func (f cageWriteCageFixture) FindByID(_ context.Context, clinicID, id uint64) (*model.Cage, error) {
	return &model.Cage{ID: id, ClinicID: clinicID}, nil
}

func (f cageWriteCageFixture) LockCageGroupForUpdate(ctx context.Context, clinicID, id uint64) ([]model.Cage, error) {
	return f.lock(ctx, clinicID, id)
}

func (f cageWriteCageFixture) ReadEffectiveStatesForCageAssignment(ctx context.Context, clinicID uint64, ids []uint64) ([]StayCageEffectiveState, error) {
	return f.read(ctx, clinicID, ids)
}

func (f cageWriteCageFixture) CountOccupiedCagesExcludingHospitalizations(ctx context.Context, clinicID uint64, ids, excluded []uint64) (map[uint64]int64, error) {
	return f.count(ctx, clinicID, ids, excluded)
}

func cageWriteState(clinicID, id uint64, state StayCageOccupancyState) StayCageEffectiveState {
	return StayCageEffectiveState{ClinicID: clinicID, StayCaseID: id, State: state,
		Evidence: StayCageEvidence{StayVersion: 2, MovementHeadRevision: 1}}
}

func requireCageWriteCode(t *testing.T, err error, want string) {
	t.Helper()
	require.ErrorIs(t, err, apperrors.ErrConflict)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, want, appErr.Code)
}

func TestHospitalizationCreateCage_RejectsWithoutProofBeforeAnyWrite(t *testing.T) {
	cageID := uint64(10)
	t.Run("no stay history can prove acceptance", func(t *testing.T) {
		svc := &hospitalizationService{transactor: &mockTransactor{}}
		got, err := svc.Create(context.Background(), 7, &CreateHospitalizationInput{CageID: &cageID, StayKind: stayKindInpatient})
		require.Nil(t, got)
		requireCageWriteCode(t, err, codeOccupancyUnknown)
	})
}

func TestHospitalizationCreateCage_CagelessCreateStillPersists(t *testing.T) {
	writes := 0
	svc := &hospitalizationService{
		transactor:      &mockTransactor{},
		reservationRepo: &mockReservationRepository{},
		petRepo: &mockPetRepository{findByIDFn: func(_ context.Context, clinic, id uint64) (*model.Pet, error) {
			return &model.Pet{ID: id, ClinicID: clinic}, nil
		}},
		petLife: stubPetLifeReader{},
		hospRepo: &mockHospitalizationRepository{createFn: func(_ context.Context, row *model.Hospitalization) error {
			require.Nil(t, row.CageID)
			writes++
			return nil
		}},
	}
	got, err := svc.Create(context.Background(), 7, &CreateHospitalizationInput{})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Nil(t, got.CageID)
	require.Equal(t, 1, writes)
}

func TestHospitalizationUpdateCage_ValidatesSubjectBeforeSelfExclusion(t *testing.T) {
	stayID, cageID := uint64(31), uint64(10)
	for _, tc := range []struct {
		name     string
		state    StayCageOccupancyState
		noStay   bool
		badScope bool
		headZero bool
		occupied bool
		readErr  error
		wantCode string
		want400  bool
	}{
		{name: "admitted", state: StayCageAdmitted},
		{name: "reserved", state: StayCageReserved},
		{name: "missing own stay", noStay: true, wantCode: codeOccupancyUnknown},
		{name: "unknown", state: StayCageUnknown, wantCode: codeOccupancyUnknown},
		{name: "released is known ineligible", state: StayCageReleased, want400: true},
		{name: "foreign stay", state: StayCageAdmitted, badScope: true, wantCode: codeOccupancyUnknown},
		{name: "empty history lacks accepted reservation proof", state: StayCageReserved, headZero: true, wantCode: codeOccupancyUnknown},
		{name: "third party occupies standalone", state: StayCageAdmitted, occupied: true, wantCode: codeCageOccupied},
		{name: "typed unresolved read", readErr: &cageOccupancyUnknown{reason: "legacy"}, wantCode: codeOccupancyUnknown},
		{name: "acquisition failure", readErr: errors.New("database failed")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx := &gorm.DB{}
			var calls []string
			row := &model.Hospitalization{ID: 11, ClinicID: 7, StayCaseID: &stayID}
			if tc.noStay {
				row.StayCaseID = nil
			}
			check := func(ctx context.Context, clinic uint64) {
				require.Same(t, tx, persistence.TxFromContext(ctx))
				require.Equal(t, uint64(7), clinic)
			}
			cages := cageWriteCageFixture{
				lock: func(ctx context.Context, clinic, id uint64) ([]model.Cage, error) {
					check(ctx, clinic)
					calls = append(calls, "cage")
					return []model.Cage{{ID: id, ClinicID: clinic}}, nil
				},
				read: func(ctx context.Context, clinic uint64, ids []uint64) ([]StayCageEffectiveState, error) {
					check(ctx, clinic)
					require.Equal(t, []string{"cage"}, calls)
					require.Equal(t, []uint64{stayID}, ids)
					calls = append(calls, "stay")
					state := cageWriteState(clinic, stayID, tc.state)
					if tc.badScope {
						state.ClinicID++
					}
					if tc.headZero {
						state.Evidence.MovementHeadRevision = 0
					}
					return []StayCageEffectiveState{state}, tc.readErr
				},
				count: func(ctx context.Context, clinic uint64, ids, excluded []uint64) (map[uint64]int64, error) {
					check(ctx, clinic)
					require.Equal(t, []string{"cage", "stay"}, calls)
					require.Equal(t, []uint64{cageID}, ids)
					require.Equal(t, []uint64{row.ID}, excluded)
					calls = append(calls, "others")
					if tc.occupied {
						return map[uint64]int64{cageID: 1}, nil
					}
					return map[uint64]int64{}, nil
				},
			}
			repo := &mockHospitalizationRepository{
				findByIDFn: func(ctx context.Context, clinic, _ uint64) (*model.Hospitalization, error) {
					check(ctx, clinic)
					return row, nil
				},
				updateFieldsFn: func(ctx context.Context, clinic, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
					check(ctx, clinic)
					calls = append(calls, "write")
					return row, nil
				},
			}
			svc := &hospitalizationService{hospRepo: repo, cageRepo: cages,
				transactor: &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
					return fn(persistence.WithTxValue(ctx, tx))
				}},
			}
			got, err := svc.Update(context.Background(), 7, 11, &UpdateHospitalizationInput{CageID: &cageID, ExpectedCageIsNull: true})
			if tc.wantCode != "" || tc.want400 || tc.readErr != nil {
				require.Nil(t, got)
				require.NotContains(t, calls, "write")
				if tc.wantCode != "" {
					requireCageWriteCode(t, err, tc.wantCode)
				} else if tc.want400 {
					require.ErrorIs(t, err, apperrors.ErrInvalidInput)
				} else {
					require.ErrorIs(t, err, tc.readErr)
					require.NotErrorIs(t, err, apperrors.ErrConflict)
				}
				return
			}
			require.NoError(t, err)
			require.Equal(t, []string{"cage", "stay", "others", "write"}, calls)
		})
	}
}

func TestHospitalizationUpdateCage_OmitRetainClearPreserveUnknownAssignment(t *testing.T) {
	cageID, staleID, memo := uint64(10), uint64(20), "memo"
	for _, tc := range []struct {
		name  string
		input UpdateHospitalizationInput
		stale bool
	}{
		{"omitted", UpdateHospitalizationInput{Memo: &memo}, false},
		{"same locked cage", UpdateHospitalizationInput{CageID: &cageID, ExpectedCageID: &cageID}, false},
		{"explicit clear", UpdateHospitalizationInput{ClearCageID: true, ExpectedCageID: &cageID}, false},
		{"stale clear", UpdateHospitalizationInput{ClearCageID: true, ExpectedCageID: &staleID}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			writes := 0
			row := &model.Hospitalization{ID: 11, ClinicID: 7, CageID: &cageID}
			svc := &hospitalizationService{cageRepo: cageWriteCageFixture{}, transactor: &mockTransactor{},
				hospRepo: &mockHospitalizationRepository{
					findByIDFn: func(context.Context, uint64, uint64) (*model.Hospitalization, error) { return row, nil },
					updateFieldsFn: func(context.Context, uint64, uint64, UpdateHospitalizationInput) (*model.Hospitalization, error) {
						writes++
						return row, nil
					},
				},
			}
			_, err := svc.Update(context.Background(), 7, 11, &tc.input)
			if tc.stale {
				requireCageWriteCode(t, err, codeCageAssignmentChanged)
				require.Zero(t, writes)
			} else {
				require.NoError(t, err)
				require.Equal(t, 1, writes)
			}
		})
	}
}

func TestHospitalizationSwapCages_BothSubjectsBeforeExclusion(t *testing.T) {
	cageA, cageB, stayA, stayB := uint64(10), uint64(20), uint64(31), uint64(32)
	for _, tc := range []struct {
		name       string
		second     StayCageOccupancyState
		countError error
		occupied   bool
		wantCode   string
	}{
		{name: "both admitted", second: StayCageAdmitted},
		{name: "second unknown", second: StayCageUnknown, wantCode: codeOccupancyUnknown},
		{name: "second released", second: StayCageReleased},
		{name: "other occupancy unknown", second: StayCageAdmitted, countError: &cageOccupancyUnknown{reason: "legacy"}, wantCode: codeOccupancyUnknown},
		{name: "third party occupies", second: StayCageAdmitted, occupied: true, wantCode: codeCageOccupied},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx := &gorm.DB{}
			var calls []string
			rows := map[uint64]*model.Hospitalization{
				11: {ID: 11, ClinicID: 7, PetID: 21, CageID: &cageA, StayCaseID: &stayA},
				12: {ID: 12, ClinicID: 7, PetID: 22, CageID: &cageB, StayCaseID: &stayB},
			}
			cages := cageWriteCageFixture{
				lock: func(ctx context.Context, clinic, id uint64) ([]model.Cage, error) {
					require.Same(t, tx, persistence.TxFromContext(ctx))
					calls = append(calls, "cage"+strconv.FormatUint(id, 10))
					return []model.Cage{{ID: id, ClinicID: clinic}}, nil
				},
				read: func(ctx context.Context, clinic uint64, ids []uint64) ([]StayCageEffectiveState, error) {
					require.Same(t, tx, persistence.TxFromContext(ctx))
					require.Equal(t, []string{"cage10", "cage20"}, calls)
					require.Equal(t, []uint64{stayA, stayB}, ids)
					calls = append(calls, "both stays")
					return []StayCageEffectiveState{cageWriteState(clinic, stayA, StayCageAdmitted), cageWriteState(clinic, stayB, tc.second)}, nil
				},
				count: func(ctx context.Context, _ uint64, ids, excluded []uint64) (map[uint64]int64, error) {
					require.Same(t, tx, persistence.TxFromContext(ctx))
					require.Contains(t, calls, "both stays")
					require.Equal(t, []uint64{11, 12}, excluded)
					calls = append(calls, "others")
					if tc.occupied {
						return map[uint64]int64{ids[0]: 1}, nil
					}
					return map[uint64]int64{}, tc.countError
				},
			}
			svc := &hospitalizationService{cageRepo: cages, petRepo: swapSnapshotPetRepository{},
				hospRepo: &mockHospitalizationRepository{
					findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) { return rows[id], nil },
					updateFieldsFn: func(_ context.Context, _, id uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
						calls = append(calls, "write")
						return rows[id], nil
					},
				},
				transactor: &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
					return fn(persistence.WithTxValue(ctx, tx))
				}},
			}
			_, _, err := svc.SwapCages(context.Background(), 7, 12, 11, [2]uint64{cageB, cageA})
			if tc.wantCode != "" || tc.second == StayCageReleased {
				require.NotContains(t, calls, "write")
				if tc.wantCode != "" {
					requireCageWriteCode(t, err, tc.wantCode)
				} else {
					require.ErrorIs(t, err, apperrors.ErrInvalidInput)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, []string{"cage10", "cage20", "both stays", "others", "others", "write", "write"}, calls)
			}
			require.Equal(t, cageA, *rows[11].CageID)
			require.Equal(t, cageB, *rows[12].CageID)
		})
	}
}

func TestHospitalizationCageWriteError_EmitsContractConflictCode(t *testing.T) {
	for _, code := range []string{codeCageAssignmentChanged, codeCageOccupied, codeOccupancyUnknown} {
		t.Run(code, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			respondHospitalizationCageWriteError(c, hospitalizationCageConflict(code, "safe message"))
			require.Equal(t, http.StatusConflict, w.Code)
			require.JSONEq(t, `{"error":"safe message","code":"`+code+`"}`, w.Body.String())
		})
	}
}
