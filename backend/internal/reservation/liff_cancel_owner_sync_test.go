package reservation

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestLiffService_CancelReservation_SyncsOwnerAtCancellation(t *testing.T) {
	// Deterministically model the interleaving: the notification snapshot sees A,
	// then the staff update to B finishes before the cancellation UPDATE.
	committed := false
	syncer := &commitAwareTagSyncer{t: t, committed: &committed}
	svc := &liffService{
		adminRepo: &mockLiffAdminRepository{
			findByIDForNotifyFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
				return &model.Reservation{ID: 10, ClinicID: 3, OwnerID: uint64Ptr(7)}, nil
			},
			cancelByIDFn: func(_ context.Context, clinicID, customerID, id uint64) (*uint64, error) {
				assert.Equal(t, uint64(3), clinicID)
				assert.Equal(t, uint64(1), customerID)
				assert.Equal(t, uint64(10), id)
				committed = true
				return uint64Ptr(8), nil
			},
		},
		notifier:         &mockLiffNotifier{},
		customerRepo:     &mockLiffCustomerRepository{},
		nextVisitTagSync: syncer,
	}

	require.NoError(t, svc.CancelReservation(context.Background(), 3, 1, 10))
	assert.True(t, committed)
	assert.Equal(t, []uint64{8}, syncer.calls, "sync must target B from the cancellation, not A from the notification snapshot")
}

func TestLiffService_CancelReservation_OwnerSyncFailurePaths(t *testing.T) {
	for _, tt := range []struct {
		name       string
		notifier   bool
		notifyErr  error
		ownerID    *uint64
		cancelErr  error
		syncErr    error
		wantOwners []uint64
	}{
		{name: "sync-only does not need notification read", ownerID: uint64Ptr(8), wantOwners: []uint64{8}},
		{name: "notification read failure cannot suppress sync", notifier: true, notifyErr: errors.New("read failed"), ownerID: uint64Ptr(8), wantOwners: []uint64{8}},
		{name: "nil cancellation owner does not use stale notification owner", notifier: true},
		{name: "zero cancellation owner is skipped", ownerID: uint64Ptr(0)},
		{name: "not found or unauthorized does not sync", notifier: true, ownerID: uint64Ptr(8), cancelErr: apperrors.ErrNotFound},
		{name: "write or commit failure does not sync", notifier: true, ownerID: uint64Ptr(8), cancelErr: errors.New("write failed")},
		{name: "tag failure does not reverse successful cancellation", ownerID: uint64Ptr(8), syncErr: errors.New("tag unavailable"), wantOwners: []uint64{8}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			committed, notified, cleaned := false, false, false
			syncer := &commitAwareTagSyncer{t: t, committed: &committed, err: tt.syncErr}
			svc := &liffService{
				adminRepo: &mockLiffAdminRepository{
					findByIDForNotifyFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
						assert.True(t, tt.notifier, "tag sync alone must not depend on a notification pre-read")
						if tt.notifyErr != nil {
							return nil, tt.notifyErr
						}
						return &model.Reservation{ID: 10, ClinicID: 3, OwnerID: uint64Ptr(7)}, nil
					},
					cancelByIDFn: func(_ context.Context, _, _, _ uint64) (*uint64, error) {
						assert.Empty(t, syncer.calls)
						committed = tt.cancelErr == nil
						return tt.ownerID, tt.cancelErr
					},
				},
				customerRepo:     &mockLiffCustomerRepository{},
				nextVisitTagSync: syncer,
				medicalRecord: &mockMedicalRecordService{deleteDraftFromReservationFn: func(_ context.Context, _, _ uint64) {
					assert.True(t, committed)
					cleaned = true
				}},
			}
			if tt.notifier {
				svc.notifier = &mockLiffNotifier{notifyCancelledFn: func(_ context.Context, _ *model.Reservation, _ *model.LineCustomer) {
					assert.True(t, committed)
					notified = true
				}}
			}
			err := svc.CancelReservation(context.Background(), 3, 1, 10)
			if tt.cancelErr != nil {
				require.ErrorIs(t, err, tt.cancelErr)
				assert.False(t, cleaned)
				assert.False(t, notified)
			} else {
				require.NoError(t, err)
				assert.True(t, cleaned)
				assert.Equal(t, tt.notifier && tt.notifyErr == nil, notified)
			}
			assert.Equal(t, tt.wantOwners, syncer.calls)
		})
	}
}
