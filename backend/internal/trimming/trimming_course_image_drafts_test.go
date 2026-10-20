package trimming

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type draftTestTx struct {
	t         *testing.T
	committed int
}

func (tx *draftTestTx) WithTx(ctx context.Context, fn func(context.Context) error) error {
	if persistence.TxFromContext(ctx) != nil {
		tx.t.Fatal("joined business transaction")
	}
	if err := fn(ctx); err != nil {
		return err
	}
	tx.committed++
	return nil
}

type draftTestRepo struct {
	row *model.TrimmingCourseImageUpload
	err error
}

func (r *draftTestRepo) Find(context.Context, ImageKey) (*model.TrimmingCourseImageUpload, error) {
	return r.row, r.err
}
func (r *draftTestRepo) Reserve(_ context.Context, key ImageKey, a uint64) error {
	if r.row != nil {
		return errors.New("duplicate")
	}
	r.row = &model.TrimmingCourseImageUpload{ClinicID: key.clinicID, ObjectKey: key.value, UploaderStaffID: a}
	return r.err
}
func (r *draftTestRepo) MarkUploaded(context.Context, ImageKey, uint64) error {
	now := time.Now()
	r.row.UploadedAt = &now
	return r.err
}
func (r *draftTestRepo) MarkSubmitted(context.Context, ImageKey) error {
	now := time.Now()
	r.row.SubmittedAt = &now
	return r.err
}

type draftTestLife struct {
	used, gone, reserved, locked bool
	err                          error
}

func (r *draftTestLife) LockImageKeys(context.Context, []ImageKey) error {
	r.locked = true
	return r.err
}
func (r *draftTestLife) ImageKeyDeleted(context.Context, ImageKey) (bool, error) {
	return r.gone, r.err
}
func (r *draftTestLife) MarkImageKeyDeleted(context.Context, ImageKey, uint64) error {
	if !r.locked {
		panic("unlocked deletion")
	}
	r.reserved = true
	r.gone = true
	return r.err
}
func (r *draftTestLife) ImageKeyInUse(context.Context, uint64, string, uint64) (bool, error) {
	return r.used, r.err
}

const draftKey = "trimming-images/7/123456_0123456789abcdef0123456789abcdef.jpg"

func newDraftTest(t *testing.T) (*courseImageDrafts, *draftTestRepo, *draftTestLife, *draftTestTx) {
	repo := &draftTestRepo{}
	life := &draftTestLife{}
	tx := &draftTestTx{t: t}
	return &courseImageDrafts{tx: tx, repo: repo, lifecycle: life}, repo, life, tx
}
func TestCourseImageDraft_OnlyCompletedOwnedUnsubmittedUnreferencedCanDiscard(t *testing.T) {
	for _, state := range []string{"eligible", "unknown", "pending", "submitted", "foreign actor", "referenced", "unreadable"} {
		t.Run(state, func(t *testing.T) {
			svc, repo, life, tx := newDraftTest(t)
			now := time.Now()
			repo.row = &model.TrimmingCourseImageUpload{UploaderStaffID: 12, UploadedAt: &now}
			switch state {
			case "unknown":
				repo.row = nil
			case "pending":
				repo.row.UploadedAt = nil
			case "submitted":
				repo.row.SubmittedAt = &now
			case "foreign actor":
				repo.row.UploaderStaffID = 99
			case "referenced":
				life.used = true
			case "unreadable":
				repo.err = errors.New("read failed")
			}
			ok, err := svc.ReserveDiscard(t.Context(), 7, 12, draftKey)
			if ok != (state == "eligible") || life.reserved != ok {
				t.Fatal("unsafe discard eligibility")
			}
			if state == "unreadable" && err == nil {
				t.Fatal("read failure allowed cleanup")
			}
			if ok && tx.committed != 1 {
				t.Fatal("tombstone did not commit")
			}
		})
	}
}
func TestCourseImageDraft_SubmissionSurvivesBusinessRejectionAndProtectsUnknown(t *testing.T) {
	svc, repo, _, tx := newDraftTest(t)
	if err := svc.ProtectSubmission(t.Context(), 7, 12, draftKey); err != nil || repo.row != nil {
		t.Fatal("legacy must stay unknown, never backfill")
	}
	if err := svc.ReserveUpload(t.Context(), 7, 12, draftKey); err != nil {
		t.Fatal(err)
	}
	if err := svc.CompleteUpload(t.Context(), 7, 12, draftKey); err != nil {
		t.Fatal(err)
	}
	if err := svc.ProtectSubmission(t.Context(), 7, 12, draftKey); err != nil {
		t.Fatal(err)
	}
	// Business field validation/save fails after this call; it owns no ledger rollback.
	if repo.row.SubmittedAt == nil || tx.committed != 4 {
		t.Fatal("submission was not committed independently")
	}
	if ok, _ := svc.ReserveDiscard(t.Context(), 7, 12, draftKey); ok {
		t.Fatal("rejected/unknown save re-enabled discard")
	}
}
func TestCourseImageDraft_OtherStaffDraftDeniedAfterProtectionCommit(t *testing.T) {
	svc, repo, life, tx := newDraftTest(t)
	now := time.Now()
	repo.row = &model.TrimmingCourseImageUpload{UploaderStaffID: 99, UploadedAt: &now}
	if err := svc.ProtectSubmission(t.Context(), 7, 12, draftKey); err == nil {
		t.Fatal("accepted another staff draft")
	}
	if repo.row.SubmittedAt == nil || tx.committed < 1 {
		t.Fatal("denial rolled back protection")
	}
	life.used = true
	if err := svc.ProtectSubmission(t.Context(), 7, 12, draftKey); err != nil {
		t.Fatal("saved/history reuse must remain allowed")
	}
}
func TestCourseImageDraft_TombstoneWinsAndCanonicalClinicRequired(t *testing.T) {
	svc, repo, life, _ := newDraftTest(t)
	life.gone = true
	if err := svc.ReserveUpload(t.Context(), 7, 12, draftKey); err == nil || repo.row != nil {
		t.Fatal("retired key overwritten")
	}
	if err := svc.ProtectSubmission(t.Context(), 7, 12, draftKey); err == nil {
		t.Fatal("retired key saved")
	}
	for _, key := range []string{"trimming-images/8/a.jpg", "/uploads/" + draftKey, "trimming-images/7/../a.jpg"} {
		if ok, err := svc.ReserveDiscard(t.Context(), 7, 12, key); err == nil || ok {
			t.Fatal("accepted key alias/foreign clinic")
		}
	}
}

func TestCourseImageDraft_StorageFailureKeepsCommittedRetryProof(t *testing.T) {
	svc, repo, life, tx := newDraftTest(t)
	if err := svc.ReserveUpload(t.Context(), 7, 12, draftKey); err != nil {
		t.Fatal(err)
	}
	if err := svc.CompleteUpload(t.Context(), 7, 12, draftKey); err != nil {
		t.Fatal(err)
	}
	ok, err := svc.ReserveDiscard(t.Context(), 7, 12, draftKey)
	if err != nil || !ok || tx.committed != 3 {
		t.Fatal("no committed reservation before external delete")
	}
	// Simulated external Delete fails; no registry or tombstone rollback occurs.
	if !life.gone || repo.row.SubmittedAt != nil {
		t.Fatal("storage error destroyed retry proof")
	}
	if err := svc.ProtectSubmission(t.Context(), 7, 12, draftKey); err == nil {
		t.Fatal("save passed committed tombstone")
	}
	ok, err = svc.ReserveDiscard(t.Context(), 7, 12, draftKey)
	if err != nil || !ok {
		t.Fatal("eligible same-owner retry blocked")
	}
}

func TestCourseImageDraft_RealStoredNameAccepted(t *testing.T) {
	name, err := trimmingImageConstraints.NewStoredName("image/jpeg", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	svc, repo, _, _ := newDraftTest(t)
	key := "trimming-images/7/" + name
	if err := svc.ReserveUpload(t.Context(), 7, 12, key); err != nil || repo.row == nil {
		t.Fatalf("real generated upload rejected: %s %v", key, err)
	}
}

func TestCourseImageDraft_PendingSubmissionRemainsProtectedAfterLateCompletion(t *testing.T) {
	svc, repo, _, tx := newDraftTest(t)
	if err := svc.ReserveUpload(t.Context(), 7, 12, draftKey); err != nil {
		t.Fatal(err)
	}
	if err := svc.ProtectSubmission(t.Context(), 7, 12, draftKey); err == nil {
		t.Fatal("unknown pending upload accepted")
	}
	if repo.row.SubmittedAt == nil || tx.committed != 2 {
		t.Fatal("pending submission protection rolled back")
	}
	if err := svc.CompleteUpload(t.Context(), 7, 12, draftKey); err != nil {
		t.Fatal(err)
	}
	if ok, _ := svc.ReserveDiscard(t.Context(), 7, 12, draftKey); ok {
		t.Fatal("late completion resurrected discard eligibility")
	}
}
