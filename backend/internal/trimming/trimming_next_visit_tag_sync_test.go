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
	"github.com/animal-ekarte/backend/internal/reservation"
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

func TestService_CreateExisting_NextVisitTagChanges(t *testing.T) {
	for _, tt := range []struct {
		name      string
		configure func(*trimmingTagMutationFixture, *CreateTrimmingInput)
		want      []uint64
	}{
		{name: "detail only"},
		{name: "same appointment values", configure: func(f *trimmingTagMutationFixture, in *CreateTrimmingInput) {
			in.StartTime, in.EndTime, in.Status = f.locked.StartTime, f.locked.EndTime, f.locked.Status
		}},
		{name: "detail course options images weight", configure: func(_ *trimmingTagMutationFixture, in *CreateTrimmingInput) {
			weight := 5.0
			in.CourseID, in.OptionIDs, in.StyleImage, in.BodyWeight = ptrUint64(4), []uint64{5}, "new-image", &weight
		}},
		{name: "doctor and end only", configure: func(f *trimmingTagMutationFixture, in *CreateTrimmingInput) {
			in.StaffID, in.EndTime = ptrUint64(8), f.locked.EndTime.Add(time.Hour)
		}},
		{name: "date changed", configure: func(f *trimmingTagMutationFixture, in *CreateTrimmingInput) {
			in.StartTime, in.EndTime = f.locked.StartTime.Add(24*time.Hour), f.locked.EndTime.Add(24*time.Hour)
		}, want: []uint64{7}},
		{name: "status changed", configure: func(_ *trimmingTagMutationFixture, in *CreateTrimmingInput) {
			in.Status = model.ReservationStatusConfirmed
		}, want: []uint64{7}},
		{name: "missing owner repaired", configure: func(f *trimmingTagMutationFixture, _ *CreateTrimmingInput) {
			f.locked.OwnerID, f.after.OwnerID = nil, nil
		}, want: []uint64{7}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := newTrimmingTagMutationFixture(t)
			in := &CreateTrimmingInput{ActorID: ptrUint64(1), AppointmentID: ptrUint64(99)}
			if tt.configure != nil {
				tt.configure(f, in)
			}
			got, err := f.service(false).Create(context.Background(), 3, in)
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.want, f.syncer.calls)
		})
	}
}

func TestService_CreateExisting_NextVisitTagRollback(t *testing.T) {
	for _, phase := range []string{"lock", "appointment", "detail read", "detail create", "options", "result read", "audit", "commit"} {
		t.Run(phase, func(t *testing.T) {
			f := newTrimmingTagMutationFixture(t)
			f.failure = phase
			got, err := f.service(false).Create(context.Background(), 3, &CreateTrimmingInput{
				ActorID: ptrUint64(1), AppointmentID: ptrUint64(99),
				StartTime: f.locked.StartTime.Add(24 * time.Hour), EndTime: f.locked.EndTime.Add(24 * time.Hour), OptionIDs: []uint64{5},
			})
			require.Error(t, err)
			assert.Nil(t, got)
			assert.Empty(t, f.syncer.calls)
		})
	}
}

func TestService_Update_NextVisitTagChanges(t *testing.T) {
	for _, tt := range []struct {
		name      string
		configure func(*trimmingTagMutationFixture, *UpdateTrimmingInput)
		want      []uint64
	}{
		{name: "detail only"},
		{name: "detail course options images weight", configure: func(_ *trimmingTagMutationFixture, in *UpdateTrimmingInput) {
			weight, image, options := 5.0, "new-image", []uint64{5}
			in.CourseID, in.OptionIDs, in.StyleImage, in.BodyWeight = ptrUint64(4), &options, &image, &weight
		}},
		{name: "same appointment values with another timezone", configure: func(f *trimmingTagMutationFixture, in *UpdateTrimmingInput) {
			start := f.locked.StartTime.In(time.FixedZone("JST", 9*60*60))
			in.StartTime, in.EndTime, in.Status, in.PetID = &start, &f.locked.EndTime, &f.locked.Status, ptrUint64(5)
		}},
		{name: "doctor and end only", configure: func(f *trimmingTagMutationFixture, in *UpdateTrimmingInput) {
			end := f.locked.EndTime.Add(time.Hour)
			in.StaffID, in.EndTime = ptrUint64(8), &end
		}},
		{name: "date changed", configure: func(f *trimmingTagMutationFixture, in *UpdateTrimmingInput) {
			start, end := f.locked.StartTime.Add(24*time.Hour), f.locked.EndTime.Add(24*time.Hour)
			in.StartTime, in.EndTime = &start, &end
		}, want: []uint64{7}},
		{name: "writer preserves semantic appointment values", configure: func(f *trimmingTagMutationFixture, in *UpdateTrimmingInput) {
			start, end := f.locked.StartTime.Add(24*time.Hour), f.locked.EndTime.Add(24*time.Hour)
			in.StartTime, in.EndTime, f.preserveStart = &start, &end, true
		}},
		{name: "status changed", configure: func(_ *trimmingTagMutationFixture, in *UpdateTrimmingInput) {
			status := model.ReservationStatusCancelled
			in.Status = &status
		}, want: []uint64{7}},
		{name: "linked pet changed", configure: func(_ *trimmingTagMutationFixture, in *UpdateTrimmingInput) {
			in.PetID = ptrUint64(6)
		}, want: []uint64{7}},
		{name: "missing owner repaired", configure: func(f *trimmingTagMutationFixture, _ *UpdateTrimmingInput) {
			f.locked.OwnerID, f.after.OwnerID = nil, nil
		}, want: []uint64{7}},
		{name: "returned owner changes and sync failure continues", configure: func(f *trimmingTagMutationFixture, in *UpdateTrimmingInput) {
			in.PetID = ptrUint64(6)
			f.after.OwnerID, f.syncer.err = ptrUint64(8), errors.New("sync failed")
		}, want: []uint64{7, 8}},
		{name: "writer mutates locked pointer", configure: func(f *trimmingTagMutationFixture, in *UpdateTrimmingInput) {
			in.PetID, f.after.OwnerID, f.mutateLocked = ptrUint64(6), ptrUint64(8), true
		}, want: []uint64{7, 8}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := newTrimmingTagMutationFixture(t)
			in := &UpdateTrimmingInput{ActorID: ptrUint64(1)}
			if tt.configure != nil {
				tt.configure(f, in)
			}
			got, replaced, err := f.service(true).Update(context.Background(), 3, 99, in)
			require.NoError(t, err)
			require.NotNil(t, got)
			require.NotNil(t, replaced)
			assert.Equal(t, tt.want, f.syncer.calls)
		})
	}
}

func TestService_Update_NextVisitTagRollback(t *testing.T) {
	for _, phase := range []string{"lock", "appointment", "detail read", "detail update", "options", "result read", "audit", "commit"} {
		t.Run(phase, func(t *testing.T) {
			f := newTrimmingTagMutationFixture(t)
			f.failure = phase
			start, end, options := f.locked.StartTime.Add(24*time.Hour), f.locked.EndTime.Add(24*time.Hour), []uint64{5}
			got, replaced, err := f.service(true).Update(context.Background(), 3, 99, &UpdateTrimmingInput{ActorID: ptrUint64(1), StartTime: &start, EndTime: &end, OptionIDs: &options})
			require.Error(t, err)
			assert.Nil(t, got)
			assert.Nil(t, replaced)
			assert.Empty(t, f.syncer.calls)
		})
	}
}

func TestService_Delete_NextVisitTagSync(t *testing.T) {
	for _, tt := range []struct {
		name, failure string
		owner         *uint64
		want          []uint64
		syncErr       error
	}{
		{name: "locked owner", owner: ptrUint64(7), want: []uint64{7}},
		{name: "writer mutates locked owner", owner: ptrUint64(7), want: []uint64{7}},
		{name: "missing owner"},
		{name: "zero owner", owner: ptrUint64(0)},
		{name: "external failure keeps delete success", owner: ptrUint64(7), want: []uint64{7}, syncErr: errors.New("sync failed")},
		{name: "lock failure", owner: ptrUint64(7), failure: "lock"},
		{name: "detail failure", owner: ptrUint64(7), failure: "detail read"},
		{name: "delete failure", owner: ptrUint64(7), failure: "delete"},
		{name: "audit failure", owner: ptrUint64(7), failure: "audit"},
		{name: "commit failure", owner: ptrUint64(7), failure: "commit"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := newTrimmingTagMutationFixture(t)
			f.locked.OwnerID, f.failure, f.syncer.err = tt.owner, tt.failure, tt.syncErr
			f.mutateLocked = tt.name == "writer mutates locked owner"
			err := f.service(true).Delete(context.Background(), 3, 99, ptrUint64(1))
			if tt.failure != "" {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, f.syncer.calls)
		})
	}
}

type trimmingTagMutationFixture struct {
	t                             *testing.T
	locked, after                 *model.Reservation
	failure                       string
	committed, inTx, mutateLocked bool
	preserveStart                 bool
	syncer                        *trimmingCommitAwareTagSyncer
}

func newTrimmingTagMutationFixture(t *testing.T) *trimmingTagMutationFixture {
	start := time.Date(2030, 1, 10, 10, 0, 0, 0, time.UTC)
	locked := &model.Reservation{ID: 99, ClinicID: 3, ReservationTypeID: 9, Status: model.ReservationStatusPending, OwnerID: ptrUint64(7), PetID: ptrUint64(5), StartTime: start, EndTime: start.Add(time.Hour), ReservationType: &model.ReservationType{ID: 9, ClinicID: 3, Category: model.ReservationTypeCategoryTrimming}}
	after := *locked
	f := &trimmingTagMutationFixture{t: t, locked: locked, after: &after}
	f.syncer = &trimmingCommitAwareTagSyncer{t: t, committed: &f.committed}
	return f
}

func (f *trimmingTagMutationFixture) err(phase string) error {
	if f.failure == phase {
		return errors.New(phase + " failed")
	}
	return nil
}

func (f *trimmingTagMutationFixture) service(hasDetail bool) Service {
	reads := 0
	repo := &mockTrimmingReservationRepository{
		findPetOwnerFn: func(context.Context, uint64, uint64) (uint64, error) { return 7, nil },
		lockAndFindByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			assert.True(f.t, f.inTx)
			assert.Equal(f.t, uint64(3), clinicID)
			assert.Equal(f.t, uint64(99), id)
			return f.locked, f.err("lock")
		},
		findByIDFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
			reads++
			if reads == 1 {
				// A stale preflight owner must never become a sync target.
				preflight := *f.locked
				preflight.OwnerID = ptrUint64(999)
				return &preflight, nil
			}
			return f.after, f.err("result read")
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.Reservation, error) {
			assert.True(f.t, f.inTx)
			if err := f.err("appointment"); err != nil {
				return nil, err
			}
			if start, ok := fields["start_time"].(time.Time); ok && !f.preserveStart {
				f.after.StartTime = start
			}
			if end, ok := fields["end_time"].(time.Time); ok {
				f.after.EndTime = end
			}
			if pet, ok := fields["pet_id"].(uint64); ok {
				f.after.PetID = &pet
			}
			if status, ok := fields["status"].(model.ReservationStatus); ok {
				f.after.Status = status
			}
			if f.after.OwnerID == nil {
				f.after.OwnerID = ptrUint64(7)
			}
			if f.mutateLocked {
				*f.locked = *f.after
			}
			return f.after, nil
		},
		deleteFn: func(context.Context, uint64, uint64) error {
			assert.True(f.t, f.inTx)
			if f.mutateLocked {
				*f.locked.OwnerID = 8
			}
			return f.err("delete")
		},
	}
	detail := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(_ context.Context, _, id uint64) (*model.AppointmentTrimmingDetail, error) {
			if err := f.err("detail read"); err != nil {
				return nil, err
			}
			if hasDetail || reads > 1 {
				return &model.AppointmentTrimmingDetail{AppointmentID: id}, nil
			}
			return nil, apperrors.ErrNotFound
		},
		createFn:     func(context.Context, *model.AppointmentTrimmingDetail) error { return f.err("detail create") },
		updateFn:     func(context.Context, *model.AppointmentTrimmingDetail) error { return f.err("detail update") },
		setOptionsFn: func(context.Context, uint64, uint64, []uint64) error { return f.err("options") },
	}
	tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
		f.inTx = true
		err := fn(ctx)
		f.inTx = false
		if err != nil {
			return err
		}
		if err := f.err("commit"); err != nil {
			return err
		}
		f.committed = true
		return nil
	}}
	return NewServiceWithAudit(repo, &mockTrimmingReservationTypeRepository{}, newAcceptingTrimmingStaffRepository(), &mockTrimmingUnavailableTimeRepository{}, detail, newActiveTrimmingCourseRepository(), newActiveTrimmingOptionRepository(), tx, &trimmingAuditRecorder{err: f.err("audit")}, f.syncer)
}

var _ reservation.NextVisitTagSyncer = (*trimmingCommitAwareTagSyncer)(nil)
