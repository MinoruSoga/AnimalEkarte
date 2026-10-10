package lstep

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestDeliveryClaimCache_OnlyKnownUnsentFailureIsRetryable(t *testing.T) {
	now := time.Now()
	marker, apiFailure := preSendRevalidationFailure, "lstep_add_tag_failed"
	for _, tc := range []struct {
		name    string
		log     model.LstepDeliveryTriggerLog
		blocked bool
	}{
		{"unsent", model.LstepDeliveryTriggerLog{Status: model.TriggerStatusFailed, ExcludedReason: &marker}, false},
		{"API failure", model.LstepDeliveryTriggerLog{Status: model.TriggerStatusFailed, ExcludedReason: &apiFailure}, true},
		{"unknown failure", model.LstepDeliveryTriggerLog{Status: model.TriggerStatusFailed}, true},
		{"fired evidence", model.LstepDeliveryTriggerLog{Status: model.TriggerStatusFailed, ExcludedReason: &marker, FiredAt: &now}, true},
		{"suppressed", model.LstepDeliveryTriggerLog{Status: model.TriggerStatusFailed, ExcludedReason: &marker, SuppressedByPriority: true}, true},
		{"scheduled", model.LstepDeliveryTriggerLog{Status: model.TriggerStatusScheduled, ExcludedReason: &marker}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.log.TriggerType = model.TriggerTypeNextVisitTentativeReminder0d
			cache := &deliveryBatchTriggerLogCache{loaded: true, dayLogsByOwner: map[uint64][]model.LstepDeliveryTriggerLog{1: {tc.log}}}
			blocked, err := cache.ExistsTodayByOwnerAndType(context.Background(), 1, 1, tc.log.TriggerType, now)
			require.NoError(t, err)
			assert.Equal(t, tc.blocked, blocked)
		})
	}
}

func TestTentativeReminderFailure_DoesNotReleaseAfterSendStarts(t *testing.T) {
	for _, commitFailure := range []bool{false, true} {
		t.Run(map[bool]string{false: "API error", true: "serialization commit error after send"}[commitFailure], func(t *testing.T) {
			var statuses []string
			var reasons []string
			logs := &mockDeliveryTriggerLogRepository{updateStatusFn: func(_ context.Context, _ uint64, status string, _ *time.Time, reason *string) error {
				statuses = append(statuses, status)
				if reason != nil {
					reasons = append(reasons, *reason)
				}
				return nil
			}}
			appt := &mockDeliveryAppointmentRepository{
				findOwnerIDsWithTentativeFn: func(context.Context, uint64, time.Time) ([]uint64, error) { return []uint64{1}, nil },
				withReminderSerializationFn: func(ctx context.Context, _ uint64, _ []uint64, fn func(context.Context) error) error {
					if err := fn(ctx); err != nil {
						return err
					}
					if commitFailure {
						return errors.New("commit failed")
					}
					return nil
				},
			}
			svc := buildDeliverySvc(validOwnerRepoForDelivery(), nil, nil, nil, &mockTagCacheRepoForDelivery{}, logs, enabledSettings(), appt)
			injectTestClient(svc, &mockLstepClientForDelivery{addTagFn: func(context.Context, string, string) error {
				if !commitFailure {
					return errors.New("delivery response lost")
				}
				return nil
			}})
			count, errs := svc.TriggerNextVisitTentativeReminder0D(context.Background(), 1, time.Now())
			assert.Zero(t, count)
			require.Len(t, errs, 1)
			assert.NotContains(t, reasons, preSendRevalidationFailure)
			if commitFailure {
				assert.Equal(t, []string{model.TriggerStatusFired}, statuses)
			} else {
				assert.Equal(t, []string{model.TriggerStatusFailed}, statuses)
			}
		})
	}
}

func TestTentativeReminderFailure_ReleasesBeforeSendStarts(t *testing.T) {
	for _, failurePoint := range []string{"candidate lookup", "serialization lock"} {
		t.Run(failurePoint, func(t *testing.T) {
			var statuses, reasons []string
			logs := &mockDeliveryTriggerLogRepository{updateStatusFn: func(_ context.Context, _ uint64, status string, _ *time.Time, reason *string) error {
				statuses = append(statuses, status)
				if reason != nil {
					reasons = append(reasons, *reason)
				}
				return nil
			}}
			appt := &mockDeliveryAppointmentRepository{
				findOwnerIDsWithTentativeFn: func(context.Context, uint64, time.Time) ([]uint64, error) { return []uint64{1}, nil },
				findTentativesByOwnerFn: func(context.Context, uint64, uint64, time.Time) ([]model.Reservation, error) {
					if failurePoint == "candidate lookup" {
						return nil, errors.New("lookup failed")
					}
					petID := uint64(500)
					return []model.Reservation{{PetID: &petID, Status: model.ReservationStatusTentative}}, nil
				},
				withReminderSerializationFn: func(context.Context, uint64, []uint64, func(context.Context) error) error {
					return errors.New("pet lock unavailable")
				},
			}
			svc := buildDeliverySvc(validOwnerRepoForDelivery(), nil, nil, nil, &mockTagCacheRepoForDelivery{}, logs, enabledSettings(), appt)
			sent := 0
			injectTestClient(svc, &mockLstepClientForDelivery{addTagFn: func(context.Context, string, string) error { sent++; return nil }})
			count, errs := svc.TriggerNextVisitTentativeReminder0D(context.Background(), 1, time.Now())
			assert.Zero(t, count)
			require.Len(t, errs, 1)
			assert.Zero(t, sent)
			assert.Equal(t, []string{model.TriggerStatusFailed}, statuses)
			assert.Equal(t, []string{preSendRevalidationFailure}, reasons)
		})
	}
}

func TestDeliveryClaimSuppression_IgnoresOnlyKnownUnsentFailure(t *testing.T) {
	for _, reason := range []string{preSendRevalidationFailure, "lstep_add_tag_failed"} {
		t.Run(reason, func(t *testing.T) {
			logs := &mockDeliveryTriggerLogRepoForBatch{
				findByOwnerAndDateFn: func(context.Context, uint64, uint64, time.Time) ([]model.LstepDeliveryTriggerLog, error) {
					return []model.LstepDeliveryTriggerLog{{ID: 1, TriggerType: "higher", Status: model.TriggerStatusFailed, ExcludedReason: &reason}}, nil
				},
				updateSuppressedFn: func(context.Context, uint64, uint64, string) error {
					t.Fatal("an unsent claim must not be demoted")
					return nil
				},
			}
			priorityCalls := 0
			svc := &lstepDeliveryTriggerService{triggerLogRepo: logs, prioritySvc: &mockLstepTriggerPriorityServiceForBatch{
				getPriorityForFn: func(_ context.Context, _ uint64, trigger string) (int, error) {
					priorityCalls++
					if trigger == "higher" {
						return 1, nil
					}
					return 2, nil
				},
			}}
			suppressed, err := svc.applySuppression(context.Background(), 1, 1, "current", time.Now())
			require.NoError(t, err)
			if reason == preSendRevalidationFailure {
				assert.False(t, suppressed)
				assert.Zero(t, priorityCalls)
			} else {
				assert.True(t, suppressed)
				assert.Equal(t, 2, priorityCalls)
			}
		})
	}
}

func TestDeliveryClaimRepository_ReclaimsKnownUnsentFailureOnce(t *testing.T) {
	db := setupLstepDeliveryTriggerLogTestDB(t)
	ensureDailyUniqueIndex(t, db)
	repo := NewLstepDeliveryTriggerLogRepository(db)
	ctx := context.Background()
	asOf := time.Now()
	claim := makeDeliveryTriggerLog(t, db, 1, 100, model.TriggerTypeNextVisitTentativeReminder0d, model.TriggerStatusFailed, asOf)
	require.NoError(t, repo.UpdateStatus(ctx, 1, claim.ID, model.TriggerStatusFailed, nil, strPtr(preSendRevalidationFailure)))
	blocked, err := repo.ExistsTodayByOwnerAndType(ctx, 1, 100, claim.TriggerType, asOf)
	require.NoError(t, err)
	assert.False(t, blocked)
	var wg sync.WaitGroup
	results := make(chan bool, 2)
	errs := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			next := &model.LstepDeliveryTriggerLog{ClinicID: 1, OwnerID: 100, TriggerType: claim.TriggerType, Status: model.TriggerStatusScheduled, ScheduledAt: asOf}
			created, claimErr := repo.CreateIfAbsentToday(ctx, next)
			if created {
				assert.Equal(t, claim.ID, next.ID)
			}
			results <- created
			errs <- claimErr
		}()
	}
	wg.Wait()
	winners := 0
	for range 2 {
		require.NoError(t, <-errs)
		if <-results {
			winners++
		}
	}
	assert.Equal(t, 1, winners)
	var stored model.LstepDeliveryTriggerLog
	require.NoError(t, db.First(&stored, claim.ID).Error)
	assert.Equal(t, model.TriggerStatusScheduled, stored.Status)
	assert.Nil(t, stored.ExcludedReason)
	// A failed API attempt is still a blocking claim.
	require.NoError(t, repo.UpdateStatus(ctx, 1, claim.ID, model.TriggerStatusFailed, nil, strPtr("lstep_add_tag_failed")))
	blocked, err = repo.ExistsTodayByOwnerAndType(ctx, 1, 100, claim.TriggerType, asOf)
	require.NoError(t, err)
	assert.True(t, blocked)
}

func TestTentativeReminder_RevalidationFailureCanRetrySameDay(t *testing.T) {
	db := setupLstepDeliveryTriggerLogTestDB(t)
	ensureDailyUniqueIndex(t, db)
	owner := testdb.MakeTestOwner(t, db, 1, "retry reminder owner")
	logs := NewLstepDeliveryTriggerLogRepository(db)
	fail := true
	appt := &mockDeliveryAppointmentRepository{
		findOwnerIDsWithTentativeFn: func(context.Context, uint64, time.Time) ([]uint64, error) { return []uint64{owner.ID}, nil },
		findTentativesByOwnerFn: func(context.Context, uint64, uint64, time.Time) ([]model.Reservation, error) {
			if fail {
				return nil, errors.New("transient lookup failure")
			}
			petID := uint64(500)
			return []model.Reservation{{PetID: &petID, Status: model.ReservationStatusTentative}}, nil
		},
	}
	owners := &mockOwnerRepoForDelivery{findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) { return defaultOwnerWithLine(id), nil }}
	svc := buildDeliverySvc(owners, nil, nil, nil, &mockTagCacheRepoForDelivery{}, logs, enabledSettings(), appt)
	sent := 0
	injectTestClient(svc, &mockLstepClientForDelivery{addTagFn: func(context.Context, string, string) error { sent++; return nil }})
	asOf := time.Now()
	count, errs := svc.TriggerNextVisitTentativeReminder0D(context.Background(), 1, asOf)
	assert.Zero(t, count)
	require.Len(t, errs, 1)
	assert.Zero(t, sent)
	fail = false
	count, errs = svc.TriggerNextVisitTentativeReminder0D(context.Background(), 1, asOf)
	assert.Equal(t, 1, count)
	assert.Empty(t, errs)
	count, errs = svc.TriggerNextVisitTentativeReminder0D(context.Background(), 1, asOf)
	assert.Zero(t, count)
	assert.Empty(t, errs)
	assert.Equal(t, 1, sent)
}
