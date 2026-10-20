package medicalrecord

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestMasterWriteRowIDs_GlobalOrder(t *testing.T) {
	for _, tc := range []struct {
		target, parent uint64
		want           []uint64
	}{
		{9, 2, []uint64{2, 9}}, {2, 9, []uint64{2, 9}}, {9, 9, []uint64{9}}, {0, 9, []uint64{9}},
	} {
		if got := masterWriteRowIDs(tc.target, &tc.parent); !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("got %v want %v", got, tc.want)
		}
	}
}

// Controlled service interleavings only; these do not demonstrate DB row locking.
func TestMasterParentWrite_DeletedParentStopsCreateAfterLock(t *testing.T) {
	actor, parent := uint64(12), uint64(2)
	for _, entity := range []string{"consultation", "procedure"} {
		t.Run(entity, func(t *testing.T) {
			locked, created := false, false
			lock := func(_ context.Context, cid uint64, ids []uint64) error {
				if cid != 7 || !reflect.DeepEqual(ids, []uint64{2}) {
					t.Fatal("wrong lock set")
				}
				locked = true
				return nil
			}
			missing := func() error {
				if !locked {
					t.Fatal("parent read before lock")
				}
				return apperrors.WrapNotFound("parent", "2")
			}
			var err error
			if entity == "consultation" {
				repo := &mockConsultationRepository{lockWriteRowsFn: lock, findByIDFn: func(context.Context, uint64, uint64) (*model.Consultation, error) { return nil, missing() }, createFn: func(context.Context, *model.Consultation) error { created = true; return nil }}
				_, err = NewConsultationServiceWithAudit(repo, masterTestWriteDependencies()).Create(t.Context(), 7, &CreateConsultationInput{ActorID: &actor, Name: "child", ParentID: &parent})
			} else {
				repo := &mockProcedureRepository{lockWriteRowsFn: lock, findByIDFn: func(context.Context, uint64, uint64) (*model.Procedure, error) { return nil, missing() }, createFn: func(context.Context, *model.Procedure) error { created = true; return nil }}
				_, err = NewProcedureServiceWithAudit(repo, masterTestWriteDependencies()).Create(t.Context(), 7, &CreateProcedureInput{ActorID: &actor, Name: "child", ParentID: &parent})
			}
			if err == nil || created {
				t.Fatal("deleted parent reached create")
			}
		})
	}
}
func TestMasterParentWrite_UpdateLocksCompleteSetBeforeBeforeImage(t *testing.T) {
	actor, parent := uint64(12), uint64(2)
	stop := errors.New("stop after ordered lock")
	for _, entity := range []string{"consultation", "procedure"} {
		t.Run(entity, func(t *testing.T) {
			called := false
			lock := func(_ context.Context, cid uint64, ids []uint64) error {
				called = true
				if cid != 7 || !reflect.DeepEqual(ids, []uint64{2, 9}) {
					t.Fatal("wrong lock set")
				}
				return stop
			}
			var err error
			if entity == "consultation" {
				repo := &mockConsultationRepository{lockWriteRowsFn: lock}
				_, err = NewConsultationServiceWithAudit(repo, masterTestWriteDependencies()).Update(t.Context(), 7, 9, &UpdateConsultationInput{Version: 1, ActorID: &actor, ParentID: &parent})
			} else {
				repo := &mockProcedureRepository{lockWriteRowsFn: lock}
				_, err = NewProcedureServiceWithAudit(repo, masterTestWriteDependencies()).Update(t.Context(), 7, 9, &UpdateProcedureInput{Version: 1, ActorID: &actor, ParentID: &parent})
			}
			if !called || !errors.Is(err, stop) {
				t.Fatal("write escaped lock failure")
			}
		})
	}
}
func TestMasterParentWrite_DeleteRechecksChildAfterLockWait(t *testing.T) {
	for _, entity := range []string{"consultation", "procedure"} {
		t.Run(entity, func(t *testing.T) {
			locked, deleted := false, false
			lock := func(_ context.Context, cid uint64, ids []uint64) error {
				if cid != 7 || !reflect.DeepEqual(ids, []uint64{2}) {
					t.Fatal("wrong delete lock")
				}
				locked = true
				return nil
			}
			children := func(context.Context, uint64, uint64) (int64, error) {
				if !locked {
					t.Fatal("guard before target lock")
				}
				return 1, nil
			}
			var err error
			if entity == "consultation" {
				repo := &mockConsultationRepository{lockWriteRowsFn: lock, findByIDFn: func(context.Context, uint64, uint64) (*model.Consultation, error) {
					return &model.Consultation{ID: 2}, nil
				}, countChildrenByParentIDFn: children, deleteFn: func(context.Context, uint64, uint64) error { deleted = true; return nil }}
				err = NewConsultationServiceWithAudit(repo, masterTestWriteDependencies()).Delete(t.Context(), 7, 2)
			} else {
				repo := &mockProcedureRepository{lockWriteRowsFn: lock, findByIDFn: func(context.Context, uint64, uint64) (*model.Procedure, error) { return &model.Procedure{ID: 2}, nil }, countChildrenByParentIDFn: children, deleteFn: func(context.Context, uint64, uint64) error { deleted = true; return nil }}
				err = NewProcedureServiceWithAudit(repo, masterTestWriteDependencies()).Delete(t.Context(), 7, 2)
			}
			if err == nil || deleted {
				t.Fatal("delete ignored child committed before lock acquired")
			}
		})
	}
}
