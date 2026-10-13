package support

// create_plane_context_test.go — 保存後 best-effort 起票の context 分離（PR650-R2）。
//
// repo.Create 成功後の post-save 処理は context.WithoutCancel(requestCtx) を
// 起点に、外部 HTTP / DB 永続化 / 監査書込みそれぞれに fresh bounded context を
// 張る。request のキャンセルが保存済み報告の起票を潰さない一方、goroutine は
// 使わず同期的に実行し、外部呼出の deadline が失敗理由の DB 記録まで潰さない
// ことを固定する。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

type postSaveTestCtxKey struct{}

// INSERT 成功のコールバック内で request ctx がキャンセルされても、
// 起票と claim 永続化は継続する（保存済み報告の契約を破らない）。
func TestServiceCreate_PlaneWorkSurvivesRequestCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var claimCalled bool
	repo := &mockRepository{
		createFn: func(_ context.Context, r *model.SupportBugReport) error {
			r.ID = 10
			cancel() // INSERT 成功直後に request ctx がキャンセルされる
			return nil
		},
		setTicketFn: func(c context.Context, _, _ uint64, _, _ string) (bool, error) {
			claimCalled = true
			assert.NoError(t, c.Err(), "claim persistence must run on a live context")
			return true, nil
		},
	}
	tickets := &mockTicketCreator{result: &PlaneIssue{ID: "issue-1"}}
	svc := NewService(repo, tickets, nil)

	report, err := svc.Create(ctx, 1, 5, CreateBugReportInput{Title: "t"})

	require.NoError(t, err)
	require.NotNil(t, report)
	assert.Equal(t, 1, tickets.calls, "post-save plane work must survive request cancellation")
	assert.True(t, claimCalled, "claim persistence must run on a fresh bounded context")
}

// post-save 各操作の ctx は request の値を引き継ぎ（WithoutCancel）、
// 個別の deadline を持ち、request キャンセル後も生きている。
func TestServiceCreate_PostSaveContextsKeepValuesAndAreBounded(t *testing.T) {
	ctx := context.WithValue(context.Background(), postSaveTestCtxKey{}, "req-42")

	var planeValue, claimValue any
	var planeHasDeadline, planeAlive, claimHasDeadline, claimAlive bool

	repo := &mockRepository{
		createFn: func(_ context.Context, r *model.SupportBugReport) error {
			r.ID = 10
			return nil
		},
		setTicketFn: func(c context.Context, _, _ uint64, _, _ string) (bool, error) {
			claimValue = c.Value(postSaveTestCtxKey{})
			_, claimHasDeadline = c.Deadline()
			claimAlive = c.Err() == nil
			return true, nil
		},
	}
	tickets := &mockTicketCreator{
		result: &PlaneIssue{ID: "issue-1"},
		fn: func(c context.Context, _ *model.SupportBugReport) {
			planeValue = c.Value(postSaveTestCtxKey{})
			_, planeHasDeadline = c.Deadline()
			planeAlive = c.Err() == nil
		},
	}
	svc := NewService(repo, tickets, nil)

	_, err := svc.Create(ctx, 1, 5, CreateBugReportInput{Title: "t"})
	require.NoError(t, err)

	assert.Equal(t, "req-42", planeValue)
	assert.True(t, planeHasDeadline, "external call must run on a bounded context")
	assert.True(t, planeAlive)
	assert.Equal(t, "req-42", claimValue)
	assert.True(t, claimHasDeadline, "claim persistence must run on a bounded context")
	assert.True(t, claimAlive)
}

// 外部 budget が切れても、失敗理由の DB 記録は別の fresh bounded context で
// 書かれる（HTTP deadline が persistence を潰さない）。
func TestServiceCreate_ExpiredPlaneBudgetStillRecordsFailure(t *testing.T) {
	var planeCtxErr error
	var syncErrRecorded string
	var syncErrHasDeadline, syncErrAlive bool

	repo := &mockRepository{
		createFn: func(_ context.Context, r *model.SupportBugReport) error {
			r.ID = 10
			return nil
		},
		setCreationSyncErrFn: func(c context.Context, _, _ uint64, syncErr string) (bool, error) {
			_, syncErrHasDeadline = c.Deadline()
			syncErrAlive = c.Err() == nil
			syncErrRecorded = syncErr
			return true, nil
		},
	}
	tickets := &mockTicketCreator{
		err: context.DeadlineExceeded,
		fn: func(c context.Context, _ *model.SupportBugReport) {
			planeCtxErr = c.Err()
		},
	}
	// 極小の外部 budget: planeCtx は呼出時点で既に deadline exceeded。
	svc := NewService(repo, tickets, nil, WithPlaneTimeout(time.Nanosecond))

	report, err := svc.Create(context.Background(), 1, 5, CreateBugReportInput{Title: "t"})
	require.NoError(t, err)
	require.NotNil(t, report)

	assert.ErrorIs(t, planeCtxErr, context.DeadlineExceeded)
	assert.Equal(t, "plane request failed", syncErrRecorded)
	assert.True(t, syncErrHasDeadline, "failure persistence must run on a bounded context")
	assert.True(t, syncErrAlive, "external timeout must not poison the failure-write context")
}

// INSERT 前の request キャンセルは従来どおり失敗 — post-save 起票にも進まない。
func TestServiceCreate_CancelBeforeInsertDoesNotCreateTicket(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := &mockRepository{
		createFn: func(c context.Context, _ *model.SupportBugReport) error {
			return c.Err()
		},
	}
	tickets := &mockTicketCreator{result: &PlaneIssue{ID: "issue-1"}}
	svc := NewService(repo, tickets, nil)

	report, err := svc.Create(ctx, 1, 5, CreateBugReportInput{Title: "t"})

	require.Error(t, err)
	assert.Nil(t, report)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 0, tickets.calls)
}
