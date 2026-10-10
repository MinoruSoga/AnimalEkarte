package trimming

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestService_Create_NextVisitTagSync(t *testing.T) {
	for _, tt := range []struct {
		name                                    string
		previous                                *uint64
		existing                                bool
		commitErr, auditErr, detailErr, syncErr error
		want                                    []uint64
	}{
		{name: "new appointment", want: []uint64{7}},
		{name: "absorbed transferred pet", previous: ptrUint64(6), want: []uint64{6, 7}},
		{name: "absorbed same owner", previous: ptrUint64(7), want: []uint64{7}},
		{name: "zero previous owner", previous: ptrUint64(0), want: []uint64{7}},
		{name: "sync failure keeps successful write", previous: ptrUint64(6), syncErr: errors.New("sync failed"), want: []uint64{6, 7}},
		{name: "commit failure", previous: ptrUint64(6), commitErr: errors.New("commit failed")},
		{name: "audit failure", previous: ptrUint64(6), auditErr: errors.New("audit failed")},
		{name: "detail failure", previous: ptrUint64(6), detailErr: errors.New("detail failed")},
		{name: "existing detail date changed", existing: true, previous: ptrUint64(7), want: []uint64{7}},
		{name: "existing detail owner repaired", existing: true, want: []uint64{7}},
		{name: "existing detail commit failure", existing: true, commitErr: errors.New("commit failed")},
		{name: "existing detail audit failure", existing: true, auditErr: errors.New("audit failed")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			committed := false
			start := time.Date(2030, 1, 10, 10, 0, 0, 0, time.UTC)
			locked := &model.Reservation{ID: 99, ClinicID: 3, ReservationTypeID: 9, Status: model.ReservationStatusPending, OwnerID: tt.previous, StartTime: start.Add(-24 * time.Hour), EndTime: start.Add(-23 * time.Hour), ReservationType: &model.ReservationType{ID: 9, ClinicID: 3, Category: model.ReservationTypeCategoryTrimming}}
			result := *locked
			result.OwnerID = ptrUint64(7)
			result.StartTime = start
			result.EndTime = start.Add(time.Hour)
			reads := 0
			repo := &mockTrimmingReservationRepository{previousAbsorbedOwnerID: tt.previous, createFn: func(_ context.Context, b *model.Reservation) error { b.ID = 99; b.OwnerID = ptrUint64(7); return nil }, lockAndFindByIDFn: func(context.Context, uint64, uint64) (*model.Reservation, error) { return locked, nil }, findByIDFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
				reads++
				if tt.existing && reads == 1 {
					return locked, nil
				}
				return &result, nil
			}}
			detail := &mockTrimmingDetailRepository{findByAppointmentIDFn: func(context.Context, uint64, uint64) (*model.AppointmentTrimmingDetail, error) {
				return nil, apperrors.ErrNotFound
			}, createFn: func(context.Context, *model.AppointmentTrimmingDetail) error { return tt.detailErr }}
			tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
				if err := fn(ctx); err != nil {
					return err
				}
				if tt.commitErr != nil {
					return tt.commitErr
				}
				committed = true
				return nil
			}}
			syncer := &trimmingCommitAwareTagSyncer{t: t, committed: &committed, err: tt.syncErr}
			svc := NewServiceWithAudit(repo, &mockTrimmingReservationTypeRepository{}, newAcceptingTrimmingStaffRepository(), &mockTrimmingUnavailableTimeRepository{}, detail, nil, nil, tx, &trimmingAuditRecorder{err: tt.auditErr}, syncer)
			input := &CreateTrimmingInput{ActorID: ptrUint64(1), ReservationTypeID: 9, StartTime: start, EndTime: start.Add(time.Hour)}
			if tt.existing {
				input.AppointmentID = ptrUint64(99)
			}
			got, err := svc.Create(context.Background(), 3, input)
			if tt.commitErr != nil || tt.auditErr != nil || tt.detailErr != nil {
				require.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
			}
			assert.Equal(t, tt.want, syncer.calls)
		})
	}
}

type trimmingCommitAwareTagSyncer struct {
	t         *testing.T
	committed *bool
	calls     []uint64
	err       error
}

func (s *trimmingCommitAwareTagSyncer) SyncNextVisitTag(_ context.Context, clinicID, ownerID uint64) error {
	assert.True(s.t, *s.committed, "sync must run after outer commit")
	assert.Equal(s.t, uint64(3), clinicID)
	s.calls = append(s.calls, ownerID)
	return s.err
}
