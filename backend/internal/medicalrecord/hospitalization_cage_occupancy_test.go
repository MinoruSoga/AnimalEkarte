package medicalrecord

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type cageOccupancyReaderFunc func(context.Context, uint64, *uint64) (hospitalizationCageOccupancyResponse, error)

func (f cageOccupancyReaderFunc) CageOccupancyForAssignment(ctx context.Context, clinic uint64, target *uint64) (hospitalizationCageOccupancyResponse, error) {
	return f(ctx, clinic, target)
}

type cageOccupancySourceFixture struct {
	listFn    func(context.Context, uint64, *uint64) ([]uint64, error)
	findAllFn func(context.Context, uint64, *string) ([]model.Cage, error)
}

func (f cageOccupancySourceFixture) ListOccupiedCageIDsExcludingHospitalization(ctx context.Context, clinic uint64, target *uint64) ([]uint64, error) {
	return f.listFn(ctx, clinic, target)
}

func (f cageOccupancySourceFixture) FindAll(ctx context.Context, clinic uint64, cageType *string) ([]model.Cage, error) {
	return f.findAllFn(ctx, clinic, cageType)
}

type cageAssignmentFinderFunc func(context.Context, uint64, uint64) (*model.Hospitalization, error)

func (f cageAssignmentFinderFunc) FindByID(ctx context.Context, clinic, id uint64) (*model.Hospitalization, error) {
	return f(ctx, clinic, id)
}

func cageOccupancyRequest(t *testing.T, query string, checker httpapi.ClinicPermissionChecker, reader HospitalizationCageOccupancyReader) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("clinic_id", "7")
	httpapi.SetClinicPermissionChecker(c, checker)
	c.Request = httptest.NewRequest(http.MethodGet, "/hospitalizations/cage-occupancy"+query, nil)
	serveHospitalizationCageOccupancy(c, reader)
	return w
}

func cageGrant(action string) httpapi.ClinicPermissionChecker {
	return func(_ *gin.Context, clinic uint64, resource, requested string) bool {
		return clinic == 7 && resource == "hospitalization" && requested == action
	}
}

// This exercises the production HTTP boundary. It does not establish the auth
// middleware's DB grants or execute repository SQL.
func TestGetCageOccupancyHandler_SelectedClinicWriteGrant(t *testing.T) {
	for _, tc := range []struct {
		name, action, query string
		want               int
	}{
		{"create only", "create", "", 200},
		{"edit only", "edit", "", 200},
		{"edit self exclusion", "edit", "?exclude_hospitalization_id=9", 200},
		{"create cannot exclude", "create", "?exclude_hospitalization_id=9", 403},
		{"view insufficient", "view", "", 403},
		{"master edit insufficient", "master-edit", "", 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			reader := cageOccupancyReaderFunc(func(_ context.Context, clinic uint64, target *uint64) (hospitalizationCageOccupancyResponse, error) {
				calls++
				require.Equal(t, uint64(7), clinic)
				require.Equal(t, tc.query != "", target != nil)
				return hospitalizationCageOccupancyResponse{ClinicID: "7", State: "unknown"}, nil
			})
			w := cageOccupancyRequest(t, tc.query, cageGrant(tc.action), reader)
			require.Equal(t, tc.want, w.Code)
			require.Equal(t, tc.want == 200, calls == 1)
		})
	}
	for _, check := range []httpapi.ClinicPermissionChecker{
		nil,
		func(_ *gin.Context, clinic uint64, _, _ string) bool { return clinic == 8 },
	} {
		w := cageOccupancyRequest(t, "", check, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/hospitalizations/cage-occupancy", nil)
	serveHospitalizationCageOccupancy(c, nil)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetCageOccupancyHandler_StrictBigintQuery(t *testing.T) {
	for _, query := range []string{
		"?exclude_hospitalization_id=", "?exclude_hospitalization_id=null",
		"?exclude_hospitalization_id=0", "?exclude_hospitalization_id=-1",
		"?exclude_hospitalization_id=01", "?exclude_hospitalization_id=%2B1",
		"?exclude_hospitalization_id=1.0", "?exclude_hospitalization_id=1e3",
		"?exclude_hospitalization_id=9223372036854775808",
		"?exclude_hospitalization_id=1&exclude_hospitalization_id=2",
		"?clinic_id=8", "?page=2", "?status=admitted", "?pet_id=1",
		"?exclude_hospitalization_id=%zz",
	} {
		t.Run(query, func(t *testing.T) {
			w := cageOccupancyRequest(t, query, cageGrant("edit"), nil)
			require.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
	var received uint64
	reader := cageOccupancyReaderFunc(func(_ context.Context, _ uint64, target *uint64) (hospitalizationCageOccupancyResponse, error) {
		received = *target
		return hospitalizationCageOccupancyResponse{ClinicID: "7", State: "unknown"}, nil
	})
	w := cageOccupancyRequest(t, "?exclude_hospitalization_id=9223372036854775807", cageGrant("edit"), reader)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, uint64(math.MaxInt64), received)
}

func TestGetCageOccupancyHandler_UnknownPreservesCurrentAssignmentAndFailuresAre500(t *testing.T) {
	targetID, currentID := uint64(9), uint64(9007199254740993)
	finder := cageAssignmentFinderFunc(func(_ context.Context, clinic, id uint64) (*model.Hospitalization, error) {
		require.Equal(t, uint64(7), clinic)
		require.Equal(t, targetID, id)
		return &model.Hospitalization{ID: id, ClinicID: clinic, CageID: &currentID}, nil
	})
	for _, tc := range []struct {
		name string
		err  error
		want int
	}{
		{"explicit unknown", &cageOccupancyUnknown{reason: "unresolved correction"}, 200},
		{"database failure", errors.New("private database failure"), 500},
		{"dependency unavailable", apperrors.WrapServiceUnavailable("private dependency"), 500},
		{"source not found is acquisition failure", apperrors.WrapNotFound("source", "1"), 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := cageOccupancySourceFixture{listFn: func(context.Context, uint64, *uint64) ([]uint64, error) { return nil, tc.err }}
			reader := cageOccupancyReaderFunc(func(ctx context.Context, clinic uint64, target *uint64) (hospitalizationCageOccupancyResponse, error) {
				return readHospitalizationCageOccupancy(ctx, clinic, target, finder, source)
			})
			w := cageOccupancyRequest(t, "?exclude_hospitalization_id=9", cageGrant("edit"), reader)
			require.Equal(t, tc.want, w.Code)
			if tc.want == 200 {
				require.JSONEq(t, `{"clinic_id":"7","exclude_hospitalization_id":"9","state":"unknown","occupied_cage_ids":null,"candidate_cage_ids":null,"current_assignment":{"state":"known","cage_id":"9007199254740993"}}`, w.Body.String())
			} else {
				require.JSONEq(t, `{"error":"internal server error"}`, w.Body.String())
			}
		})
	}
	w := cageOccupancyRequest(t, "", cageGrant("create"), nil)
	require.Equal(t, 500, w.Code)
}

func TestCageOccupancyRead_KnownGroupsSelfExclusionAndStringIDs(t *testing.T) {
	root, targetID, stayID := uint64(10), uint64(9), uint64(19)
	rows := []stayCageAssignment{
		{HospitalizationID: targetID, ClinicID: 7, StayCaseID: nil, CageID: 11},
		{HospitalizationID: 8, ClinicID: 7, StayCaseID: &stayID, CageID: 11},
	}
	ctx := context.WithValue(t.Context(), struct{}{}, "same-read-context")
	core := fixedStayCageReader{readFn: func(received context.Context, clinic uint64, ids []uint64) ([]StayCageEffectiveState, error) {
		require.Same(t, ctx, received)
		require.Equal(t, []uint64{stayID}, ids)
		return []StayCageEffectiveState{{ClinicID: clinic, StayCaseID: stayID, State: StayCageAdmitted, Evidence: StayCageEvidence{StayVersion: 1, MovementHeadRevision: 2}}}, nil
	}}
	source := cageOccupancySourceFixture{
		listFn: func(received context.Context, clinic uint64, excluded *uint64) ([]uint64, error) {
			require.Equal(t, targetID, *excluded)
			counts, err := countEffectiveStayCages(received, clinic, rows, []uint64{*excluded}, core)
			if err != nil {
				return nil, err
			}
			ids := make([]uint64, 0, len(counts))
			for id := range counts {
				ids = append(ids, id)
			}
			return ids, nil
		},
		findAllFn: func(received context.Context, clinic uint64, cageType *string) ([]model.Cage, error) {
			require.Same(t, ctx, received)
			require.Equal(t, uint64(7), clinic)
			require.Nil(t, cageType)
			return []model.Cage{
				{ID: 10, ClinicID: 7, IsActive: true, MergedIntoCageID: nil},
				{ID: 11, ClinicID: 7, IsActive: true, MergedIntoCageID: &root},
				{ID: 20, ClinicID: 7, IsActive: true, MergedIntoCageID: nil},
				{ID: 30, ClinicID: 7, IsActive: false, MergedIntoCageID: nil},
				{ID: math.MaxInt64, ClinicID: 7, IsActive: true, MergedIntoCageID: nil},
			}, nil
		},
	}
	current := uint64(11)
	finder := cageAssignmentFinderFunc(func(received context.Context, clinic, id uint64) (*model.Hospitalization, error) {
		require.Same(t, ctx, received)
		return &model.Hospitalization{ID: id, ClinicID: clinic, CageID: &current}, nil
	})
	result, err := readHospitalizationCageOccupancy(ctx, 7, &targetID, finder, source)
	require.NoError(t, err)
	require.Equal(t, "known", result.State)
	require.Equal(t, []string{"11"}, result.OccupiedCageIDs)
	require.Equal(t, []string{"20", "9223372036854775807"}, result.CandidateCageIDs)
	require.Equal(t, "11", *result.CurrentAssignment.CageID)
}

func TestCageOccupancyRead_TargetScopeAndEmptySets(t *testing.T) {
	target := uint64(9)
	for _, row := range []*model.Hospitalization{
		nil, {ID: 9, ClinicID: 8}, {ID: 10, ClinicID: 7},
		{ID: 9, ClinicID: 7, DeletedAt: gorm.DeletedAt{Valid: true}},
	} {
		finder := cageAssignmentFinderFunc(func(context.Context, uint64, uint64) (*model.Hospitalization, error) { return row, nil })
		_, err := readHospitalizationCageOccupancy(t.Context(), 7, &target, finder, cageOccupancySourceFixture{})
		require.ErrorIs(t, err, apperrors.ErrNotFound)
	}
	source := cageOccupancySourceFixture{
		listFn: func(context.Context, uint64, *uint64) ([]uint64, error) { return nil, nil },
		findAllFn: func(context.Context, uint64, *string) ([]model.Cage, error) { return nil, nil },
	}
	finder := cageAssignmentFinderFunc(func(_ context.Context, clinic, id uint64) (*model.Hospitalization, error) {
		return &model.Hospitalization{ID: id, ClinicID: clinic, CageID: nil}, nil
	})
	for _, excluded := range []*uint64{nil, &target} {
		result, err := readHospitalizationCageOccupancy(t.Context(), 7, excluded, finder, source)
		require.NoError(t, err)
		require.Equal(t, "known", result.State)
		require.NotNil(t, result.OccupiedCageIDs)
		require.NotNil(t, result.CandidateCageIDs)
		if excluded == nil {
			require.Nil(t, result.CurrentAssignment)
		} else {
			require.Equal(t, "known", result.CurrentAssignment.State)
			require.Nil(t, result.CurrentAssignment.CageID)
		}
		data, err := json.Marshal(result)
		require.NoError(t, err)
		require.Contains(t, string(data), `"occupied_cage_ids":[]`)
	}
	for _, cages := range [][]model.Cage{
		{{ID: 1, ClinicID: 8, IsActive: true}},
		{{ID: 1, ClinicID: 7, IsActive: true}, {ID: 1, ClinicID: 7, IsActive: true}},
		{{ID: 1, ClinicID: 7, IsActive: true, MergedIntoCageID: &target}},
	} {
		source.findAllFn = func(context.Context, uint64, *string) ([]model.Cage, error) { return cages, nil }
		result, err := readHospitalizationCageOccupancy(t.Context(), 7, &target, finder, source)
		require.NoError(t, err)
		require.Equal(t, "unknown", result.State)
		require.Nil(t, result.CandidateCageIDs)
		require.Nil(t, result.OccupiedCageIDs)
		require.Equal(t, "known", result.CurrentAssignment.State)
	}
	require.Equal(t, []string{strconv.FormatInt(math.MaxInt64, 10)}, cageProjectionStrings([]uint64{math.MaxInt64, math.MaxInt64}))
}

func TestCageOccupancyRead_EmptyHistoryKeepsZeroHeadAndCurrentAssignment(t *testing.T) {
	target, stayID, current := uint64(9), uint64(11), uint64(10)
	state := projectStayCageEffectiveState(model.StayCase{ID: stayID, ClinicID: 7, StayKind: "hotel", Version: 4, DeletedAt: nil}, nil)
	core := fixedStayCageReader{readFn: func(context.Context, uint64, []uint64) ([]StayCageEffectiveState, error) {
		return []StayCageEffectiveState{state}, nil
	}}
	source := cageOccupancySourceFixture{
		listFn: func(ctx context.Context, clinic uint64, _ *uint64) ([]uint64, error) {
			_, err := countEffectiveStayCages(ctx, clinic, []stayCageAssignment{
				{HospitalizationID: 8, ClinicID: clinic, StayCaseID: &stayID, CageID: current},
			}, nil, core)
			var unknown *cageOccupancyUnknown
			require.ErrorAs(t, err, &unknown)
			require.Equal(t, "stay occupancy is unknown", unknown.reason)
			return nil, err
		},
	}
	finder := cageAssignmentFinderFunc(func(_ context.Context, clinic, id uint64) (*model.Hospitalization, error) {
		return &model.Hospitalization{ID: id, ClinicID: clinic, CageID: &current}, nil
	})
	result, err := readHospitalizationCageOccupancy(t.Context(), 7, &target, finder, source)
	require.NoError(t, err)
	require.Equal(t, "unknown", result.State)
	require.Nil(t, result.OccupiedCageIDs)
	require.Nil(t, result.CandidateCageIDs)
	require.Equal(t, "10", *result.CurrentAssignment.CageID)
	require.Equal(t, StayCageEvidence{StayVersion: 4, MovementHeadRevision: 0}, state.Evidence)
}
