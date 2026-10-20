package trimming

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestCourseImageHistory_UpdateReadsLockedBeforeAndAuditsInSameTransaction(t *testing.T) {
	for _, next := range []string{"trimming-images/1/new.jpg", ""} {
		t.Run(next, func(t *testing.T) {
			oldKey := "trimming-images/1/original.jpg"
			journal := []string{}
			tx := &trimmingAuditRollbackTransactor{journal: &journal}
			audit := &trimmingAuditRecorder{}
			locked := false
			repo := &mockTrimmingCourseRepository{
				findByIDFn: func(ctx context.Context, _, _ uint64) (*model.TrimmingCourse, error) {
					if ctx.Value(trimmingAuditTxContextKey{}) == true {
						locked = true
						return &model.TrimmingCourse{ID: 10, ClinicID: 1, ImageObjectKey: &oldKey}, nil
					}
					// An unlocked earlier read is not the history source.
					return &model.TrimmingCourse{ID: 10, ClinicID: 1, ImageObjectKey: courseAttrStrPtr("stale")}, nil
				},
				updateFieldsFn: func(ctx context.Context, _, _ uint64, _ UpdateTrimmingCourseInput) (*model.TrimmingCourse, error) {
					require.True(t, locked)
					require.Equal(t, true, ctx.Value(trimmingAuditTxContextKey{}))
					journal = append(journal, "updated")
					var key *string
					if next != "" {
						key = &next
					}
					return &model.TrimmingCourse{ID: 10, ClinicID: 1, ImageObjectKey: key}, nil
				},
			}
			lifecycle := &courseImageLifecycleFake{}
			svc := NewTrimmingCourseService(repo, &mockMinimalCourseTypeRepo{}, tx, WithCourseImageLifecycle(lifecycle), WithCourseImageAudit(audit))
			_, err := svc.Update(t.Context(), 1, 10, &UpdateTrimmingCourseInput{ImageObjectKey: &next, ActorStaffID: 7})
			require.NoError(t, err)
			require.True(t, audit.sawTxCtx)
			require.Len(t, audit.entries, 1)
			entry := audit.entries[0]
			require.Equal(t, courseImageAuditResource, entry.Resource)
			require.Equal(t, uint64(7), *entry.ActorID)
			require.Equal(t, map[string]any{"image_object_key": oldKey}, entry.OldValue)
			var want any
			if next != "" {
				want = next
			}
			require.Equal(t, map[string]any{"image_object_key": want}, entry.NewValue)
			require.Empty(t, lifecycle.tombstoned)
			require.Contains(t, lifecycle.locked, ImageKey{clinicID: 1, value: oldKey})
		})
	}
}

func TestCourseImageHistory_AuditFailureRollsBackImageWrite(t *testing.T) {
	journal := []string{}
	tx := &trimmingAuditRollbackTransactor{journal: &journal}
	audit := &trimmingAuditRecorder{err: errors.New("audit unavailable")}
	repo := &mockTrimmingCourseRepository{createFn: func(_ context.Context, course *model.TrimmingCourse) error {
		course.ID = 10
		journal = append(journal, "course created")
		return nil
	}}
	svc := NewTrimmingCourseService(repo, &mockMinimalCourseTypeRepo{}, tx, WithCourseImageLifecycle(&courseImageLifecycleFake{}), WithCourseImageAudit(audit))
	_, err := svc.Create(t.Context(), 1, &CreateTrimmingCourseInput{Name: "A", ImageObjectKey: "trimming-images/1/new.jpg", ActorStaffID: 7})
	require.Error(t, err)
	require.True(t, tx.rolledBack)
	require.Empty(t, journal)
	require.True(t, audit.sawTxCtx)
}

func TestCourseImageHistory_FailsClosedWithoutActorOrAudit(t *testing.T) {
	for _, tc := range []struct {
		name  string
		actor uint64
	}{{"missing actor", 0}, {"missing audit", 7}} {
		t.Run(tc.name, func(t *testing.T) {
			created := false
			repo := &mockTrimmingCourseRepository{createFn: func(context.Context, *model.TrimmingCourse) error { created = true; return nil }}
			svc := NewTrimmingCourseService(repo, &mockMinimalCourseTypeRepo{}, &mockTransactor{}, WithCourseImageLifecycle(&courseImageLifecycleFake{}))
			_, err := svc.Create(t.Context(), 1, &CreateTrimmingCourseInput{Name: "A", ImageObjectKey: "trimming-images/1/new.jpg", ActorStaffID: tc.actor})
			require.Error(t, err)
			require.False(t, created)
		})
	}
}

func TestCourseImageHistory_UpdateAuditFailureRollsBack(t *testing.T) {
	journal := []string{}
	tx := &trimmingAuditRollbackTransactor{journal: &journal}
	audit := &trimmingAuditRecorder{err: errors.New("audit unavailable")}
	oldKey := "trimming-images/1/original.jpg"
	next := "trimming-images/1/new.jpg"
	repo := &mockTrimmingCourseRepository{
		findByIDFn: func(context.Context, uint64, uint64) (*model.TrimmingCourse, error) {
			return &model.TrimmingCourse{ID: 10, ClinicID: 1, ImageObjectKey: &oldKey}, nil
		},
		updateFieldsFn: func(context.Context, uint64, uint64, UpdateTrimmingCourseInput) (*model.TrimmingCourse, error) {
			journal = append(journal, "image updated")
			return &model.TrimmingCourse{ID: 10, ClinicID: 1, ImageObjectKey: &next}, nil
		},
	}
	svc := NewTrimmingCourseService(repo, &mockMinimalCourseTypeRepo{}, tx, WithCourseImageLifecycle(&courseImageLifecycleFake{}), WithCourseImageAudit(audit))
	_, err := svc.Update(t.Context(), 1, 10, &UpdateTrimmingCourseInput{ImageObjectKey: &next, ActorStaffID: 7})
	require.Error(t, err)
	require.True(t, tx.rolledBack)
	require.Empty(t, journal)
}
