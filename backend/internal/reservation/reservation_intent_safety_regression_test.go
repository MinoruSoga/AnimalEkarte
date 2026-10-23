package reservation

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestClearNextVisitIntent_RejectsIncompleteGuard(t *testing.T) {
	for _, query := range []string{"?expected_reservation_id=10", "?expected_updated_at=2026-10-11T00:00:00Z", "?expected_reservation_id=0&expected_updated_at=2026-10-11T00:00:00Z", "?expected_reservation_id=10&expected_updated_at=invalid", "?expected_reservation_id=&expected_updated_at=", "?expected_reservation_id=10&expected_reservation_id=11&expected_updated_at=2026-10-11T00:00:00Z", "?expected_reservation_id=10&expected_updated_at=2026-10-11T00:00:00Z&expected_updated_at=2026-10-12T00:00:00Z"} {
		t.Run(query, func(t *testing.T) {
			called := false
			svc := &mockReservationService{clearNextVisitIntentFn: func(context.Context, uint64, uint64) error { called = true; return nil }}
			r := newClearNextVisitIntentRouter(svc)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/reservations/pets/10/next-visit-intent"+query, nil))
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.False(t, called)
		})
	}
}

type guardedClearTestService struct {
	*mockReservationService
	clear func(context.Context, uint64, uint64, ClearNextVisitIntentExpectation) error
}

func (s *guardedClearTestService) ClearNextVisitIntentIfMatches(ctx context.Context, clinicID, petID uint64, expected ClearNextVisitIntentExpectation) error {
	return s.clear(ctx, clinicID, petID, expected)
}

func TestClearNextVisitIntent_HTTPGuard(t *testing.T) {
	stamp := time.Date(2026, 10, 11, 0, 0, 0, 123456789, time.UTC)
	for _, conflict := range []bool{false, true} {
		t.Run(map[bool]string{false: "matched", true: "stale"}[conflict], func(t *testing.T) {
			svc := &guardedClearTestService{mockReservationService: &mockReservationService{}, clear: func(_ context.Context, clinicID, petID uint64, expected ClearNextVisitIntentExpectation) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(10), petID)
				assert.Equal(t, uint64(42), expected.ReservationID)
				assert.True(t, stamp.Equal(expected.UpdatedAt))
				if conflict {
					return apperrors.WrapConflict("stale")
				}
				return nil
			}}
			w := httptest.NewRecorder()
			newClearNextVisitIntentRouter(svc).ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/reservations/pets/10/next-visit-intent?expected_reservation_id=42&expected_updated_at=2026-10-11T00:00:00.123456789Z", nil))
			expected := http.StatusNoContent
			if conflict {
				expected = http.StatusConflict
			}
			assert.Equal(t, expected, w.Code)
		})
	}
	t.Run("missing guarded dependency fails closed", func(t *testing.T) {
		w := httptest.NewRecorder()
		newClearNextVisitIntentRouter(&mockReservationService{}).ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/reservations/pets/10/next-visit-intent?expected_reservation_id=42&expected_updated_at=2026-10-11T00:00:00Z", nil))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestClearNextVisitIntent_RevisionAndOwnership(t *testing.T) {
	stamp := time.Date(2026, 10, 11, 0, 0, 0, 123456789, time.UTC)
	for _, tc := range []struct {
		name           string
		current        *model.Reservation
		ownershipError bool
		failCommit     bool
		conflict       bool
	}{
		{name: "matches", current: &model.Reservation{ID: 42, UpdatedAt: stamp}},
		{name: "same ID newer revision", current: &model.Reservation{ID: 42, UpdatedAt: stamp.Add(time.Nanosecond)}, conflict: true},
		{name: "recreated ID", current: &model.Reservation{ID: 43, UpdatedAt: stamp}, conflict: true},
		{name: "already removed", conflict: true},
		{name: "foreign clinic", ownershipError: true},
		{name: "commit fails", current: &model.Reservation{ID: 42, UpdatedAt: stamp}, failCommit: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			locked, wrote, committed := false, false, false
			repo := &mockReservationRepository{
				acquireNextVisitIntentLockFn: func(_ context.Context, clinicID, petID uint64) error {
					assert.Equal(t, uint64(3), clinicID)
					assert.Equal(t, uint64(22), petID)
					locked = true
					return nil
				},
				findPetOwnerInClinicFn: func(context.Context, uint64, uint64) (uint64, error) {
					require.True(t, locked)
					if tc.ownershipError {
						return 0, apperrors.WrapNotFound("pet", "22")
					}
					return 7, nil
				},
				clearNextVisitIntentGuardedFn: func(_ context.Context, _, _ uint64, expected *ClearNextVisitIntentExpectation) (bool, *uint64, error) {
					require.True(t, locked)
					if err := checkClearNextVisitIntentExpectation(tc.current, expected); err != nil {
						return false, nil, err
					}
					wrote = true
					return true, uint64Ptr(6), nil
				},
			}
			tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
				if err := fn(ctx); err != nil {
					return err
				}
				if tc.failCommit {
					return errors.New("commit failed")
				}
				committed = true
				return nil
			}}
			syncer := &commitAwareTagSyncer{t: t, committed: &committed}
			svc := newReservationService(repo, nil, tx, nil, nil, nil, nil, nil, syncer)
			err := svc.ClearNextVisitIntentIfMatches(context.Background(), 3, 22, ClearNextVisitIntentExpectation{ReservationID: 42, UpdatedAt: stamp})
			if tc.conflict {
				require.True(t, apperrors.IsConflict(err))
				assert.False(t, wrote)
			} else if tc.ownershipError {
				require.True(t, apperrors.IsNotFound(err))
				assert.False(t, wrote)
			} else if tc.failCommit {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.True(t, wrote)
			}
			if committed {
				assert.Equal(t, []uint64{6, 7}, syncer.calls)
			} else {
				assert.Empty(t, syncer.calls)
			}
		})
	}
}

func TestReservationResponse_TentativeCurrentOwnerAndRevision(t *testing.T) {
	stamp := time.Date(2026, 10, 11, 0, 0, 0, 123456789, time.UTC)
	for _, status := range []model.ReservationStatus{model.ReservationStatusTentative, model.ReservationStatusPending} {
		t.Run(string(status), func(t *testing.T) {
			row := &model.Reservation{ID: 42, ClinicID: 3, Status: status, OwnerID: uint64Ptr(6), Owner: &model.Owner{ID: 6, ClinicID: 3, Name: "former"}, PetID: uint64Ptr(22), Pet: &model.Pet{ID: 22, ClinicID: 3, OwnerID: 7, Owner: &model.Owner{ID: 7, ClinicID: 3, Name: "current"}}, UpdatedAt: stamp}
			response := toReservationResponse(row)
			expected := uint64(6)
			if status == model.ReservationStatusTentative {
				expected = 7
			}
			require.NotNil(t, response.Owner)
			require.NotNil(t, response.OwnerID)
			assert.Equal(t, expected, *response.OwnerID)
			assert.Equal(t, expected, response.Owner.ID)
			data, err := json.Marshal(response)
			require.NoError(t, err)
			var wire struct {
				UpdatedAt string `json:"updated_at"`
			}
			require.NoError(t, json.Unmarshal(data, &wire))
			revision, err := time.Parse(time.RFC3339Nano, wire.UpdatedAt)
			require.NoError(t, err)
			assert.True(t, stamp.Equal(revision))
			assert.Equal(t, uint64(6), *row.OwnerID, "response must not rewrite historical persistence")
		})
	}
	t.Run("corrupt current owner summary hidden", func(t *testing.T) {
		response := toReservationResponse(&model.Reservation{ClinicID: 3, Status: model.ReservationStatusTentative, Pet: &model.Pet{ClinicID: 3, OwnerID: 7, Owner: &model.Owner{ID: 8, ClinicID: 4}}})
		assert.Nil(t, response.Owner)
	})
}

func TestReservationUpdate_PetRebindInvalidatesIntentProvenance(t *testing.T) {
	for _, conflictCheck := range []bool{false, true} {
		t.Run(map[bool]string{false: "link only", true: "schedule also changed"}[conflictCheck], func(t *testing.T) {
			intended := time.Date(2030, 1, 10, 0, 0, 0, 0, time.UTC)
			original := &model.Reservation{ID: 10, ClinicID: 3, OwnerID: uint64Ptr(20), PetID: uint64Ptr(22), Status: model.ReservationStatusPending, IntendedDate: &intended, SourceMedicalRecordID: uint64Ptr(99), StartTime: intended.Add(10 * time.Hour), EndTime: intended.Add(11 * time.Hour)}
			var saved map[string]any
			repo := &mockReservationRepository{
				countOnDutyDoctorsFn:   func(context.Context, uint64, time.Time) (int64, error) { return 1, nil },
				findByIDFn:             func(context.Context, uint64, uint64) (*model.Reservation, error) { return original, nil },
				lockAndFindByIDFn:      func(context.Context, uint64, uint64) (*model.Reservation, error) { return original, nil },
				findPetOwnerInClinicFn: func(context.Context, uint64, uint64) (uint64, error) { return 20, nil },
				findPetByIDInClinicFn: func(context.Context, uint64, uint64) (*model.Pet, error) {
					return &model.Pet{ID: 23, ClinicID: 3, OwnerID: 20}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.Reservation, error) {
					saved = fields
					return original, nil
				},
			}
			svc := newReservationService(repo, nil, &mockTransactor{}, nil, nil, nil, nil, nil, nil)
			svc.holidayFinder = openDayHolidayFinder()
			input := &UpdateReservationInput{PetID: uint64Ptr(23)}
			if conflictCheck {
				start := original.StartTime.Add(time.Hour)
				end := original.EndTime.Add(time.Hour)
				input.StartTime = &start
				input.EndTime = &end
			}
			_, err := svc.Update(context.Background(), 3, 10, input)
			require.NoError(t, err)
			require.Contains(t, saved, "intended_date")
			assert.Nil(t, saved["intended_date"])
			require.Contains(t, saved, "source_medical_record_id")
			assert.Nil(t, saved["source_medical_record_id"])
		})
	}
}

func TestReservationReadGuard_TransferredTentativeRemainsClinicScoped(t *testing.T) {
	db, err := gorm.Open(postgres.Open("host=disabled.invalid user=test dbname=test sslmode=disable"), &gorm.Config{DisableAutomaticPing: true, DryRun: true, SkipDefaultTransaction: true})
	require.NoError(t, err)
	var rows []model.Reservation
	query := db.Model(&model.Reservation{}).Scopes(reservationRelationsMatchParentClinic).Find(&rows)
	sql := query.Statement.SQL.String()
	assert.Contains(t, sql, "appointments.status = 'tentative'")
	assert.Contains(t, sql, "OR scoped_pet.owner_id = appointments.owner_id")
	assert.Contains(t, sql, "scoped_pet.clinic_id = appointments.clinic_id")
	assert.Contains(t, sql, "scoped_pet_owner.clinic_id = appointments.clinic_id")
	assert.Contains(t, sql, "scoped_owner.clinic_id = appointments.clinic_id")
}

func TestReservationUpdate_SamePetPreservesIntentProvenance(t *testing.T) {
	intended := time.Date(2030, 1, 10, 0, 0, 0, 0, time.UTC)
	row := &model.Reservation{ID: 10, ClinicID: 3, OwnerID: uint64Ptr(20), PetID: uint64Ptr(22), Status: model.ReservationStatusPending, IntendedDate: &intended, SourceMedicalRecordID: uint64Ptr(99)}
	var saved map[string]any
	repo := &mockReservationRepository{
		findByIDFn:             func(context.Context, uint64, uint64) (*model.Reservation, error) { return row, nil },
		lockAndFindByIDFn:      func(context.Context, uint64, uint64) (*model.Reservation, error) { return row, nil },
		findPetOwnerInClinicFn: func(context.Context, uint64, uint64) (uint64, error) { return 20, nil },
		findPetByIDInClinicFn: func(context.Context, uint64, uint64) (*model.Pet, error) {
			return &model.Pet{ID: 22, ClinicID: 3, OwnerID: 20}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.Reservation, error) {
			saved = fields
			return row, nil
		},
	}
	svc := newReservationService(repo, nil, &mockTransactor{}, nil, nil, nil, nil, nil, nil)
	updated, err := svc.Update(context.Background(), 3, 10, &UpdateReservationInput{PetID: uint64Ptr(22)})
	require.NoError(t, err)
	assert.NotContains(t, saved, "intended_date")
	assert.NotContains(t, saved, "source_medical_record_id")
	assert.True(t, toReservationResponse(updated).HadIntendedDate)
}

func TestReservationOwnerFilter_CurrentTentativeOwnerSQL(t *testing.T) {
	db, err := gorm.Open(postgres.Open("host=disabled.invalid user=test dbname=test sslmode=disable"), &gorm.Config{DisableAutomaticPing: true, DryRun: true, SkipDefaultTransaction: true})
	require.NoError(t, err)
	var rows []model.Reservation
	query := db.Model(&model.Reservation{}).Scopes(reservationOwnerFilter(7)).Find(&rows)
	sql := query.Statement.SQL.String()
	assert.Contains(t, sql, "appointments.status <> 'tentative' AND appointments.owner_id =")
	assert.Contains(t, sql, "appointments.status = 'tentative' AND EXISTS")
	assert.Contains(t, sql, "intent_pet.clinic_id = appointments.clinic_id")
	assert.Contains(t, sql, "intent_owner.clinic_id = appointments.clinic_id")
	assert.Contains(t, sql, "intent_pet.owner_id =")
	assert.Equal(t, []any{uint64(7), uint64(7)}, query.Statement.Vars)
}
