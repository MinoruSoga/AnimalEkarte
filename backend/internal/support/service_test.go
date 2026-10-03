package support

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

// ---- mock Repository ----

type mockRepository struct {
	createFn           func(ctx context.Context, report *model.SupportBugReport) error
	findAllFn          func(ctx context.Context) ([]BugReportWithReporter, error)
	findByIDForClinicF func(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error)
	updateStatusFn     func(ctx context.Context, clinicID, id uint64, status model.SupportBugReportStatus) error
	setTicketFn        func(ctx context.Context, clinicID, id uint64, issueID, issueURL string) (bool, error)
	setSyncErrFn       func(ctx context.Context, clinicID, id uint64, syncErr string) error
	softDeleteFn       func(ctx context.Context, clinicID, id uint64) error
	createChatFn       func(ctx context.Context, messages []*model.SupportChatMessage) error
	listChatFn         func(ctx context.Context, clinicID, staffID uint64) ([]model.SupportChatMessage, error)
	listAllChatFn      func(ctx context.Context) ([]ChatMessageWithMeta, error)
	clearChatFn        func(ctx context.Context, clinicID, staffID uint64) error
	listOpenPlaneFn    func(ctx context.Context, limit int) ([]model.SupportBugReport, error)
}

func (m *mockRepository) Create(ctx context.Context, report *model.SupportBugReport) error {
	return m.createFn(ctx, report)
}
func (m *mockRepository) FindAll(ctx context.Context) ([]BugReportWithReporter, error) {
	return m.findAllFn(ctx)
}
func (m *mockRepository) FindByIDForClinic(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error) {
	return m.findByIDForClinicF(ctx, clinicID, id)
}
func (m *mockRepository) UpdateStatus(ctx context.Context, clinicID, id uint64, status model.SupportBugReportStatus) error {
	return m.updateStatusFn(ctx, clinicID, id, status)
}
func (m *mockRepository) SetPlaneTicket(ctx context.Context, clinicID, id uint64, issueID, issueURL string) (bool, error) {
	return m.setTicketFn(ctx, clinicID, id, issueID, issueURL)
}
func (m *mockRepository) SetPlaneSyncError(ctx context.Context, clinicID, id uint64, syncErr string) error {
	return m.setSyncErrFn(ctx, clinicID, id, syncErr)
}
func (m *mockRepository) SoftDeleteBugReport(ctx context.Context, clinicID, id uint64) error {
	return m.softDeleteFn(ctx, clinicID, id)
}
func (m *mockRepository) CreateChatMessages(ctx context.Context, messages []*model.SupportChatMessage) error {
	if m.createChatFn == nil {
		return nil
	}
	return m.createChatFn(ctx, messages)
}
func (m *mockRepository) ListChatHistory(ctx context.Context, clinicID, staffID uint64) ([]model.SupportChatMessage, error) {
	if m.listChatFn == nil {
		return nil, nil
	}
	return m.listChatFn(ctx, clinicID, staffID)
}
func (m *mockRepository) ListAllChatMessages(ctx context.Context) ([]ChatMessageWithMeta, error) {
	if m.listAllChatFn == nil {
		return nil, nil
	}
	return m.listAllChatFn(ctx)
}
func (m *mockRepository) ClearChatHistory(ctx context.Context, clinicID, staffID uint64) error {
	if m.clearChatFn == nil {
		return nil
	}
	return m.clearChatFn(ctx, clinicID, staffID)
}
func (m *mockRepository) ListOpenWithPlaneTicket(ctx context.Context, limit int) ([]model.SupportBugReport, error) {
	if m.listOpenPlaneFn == nil {
		return nil, nil
	}
	return m.listOpenPlaneFn(ctx, limit)
}

// ---- mock TicketCreator ----

type mockTicketCreator struct {
	calls  int
	result *PlaneIssue
	err    error
}

func (m *mockTicketCreator) CreateBugReportIssue(_ context.Context, _ *model.SupportBugReport) (*PlaneIssue, error) {
	m.calls++
	return m.result, m.err
}

// savedReportMock は repo.Create が report に ID=10 を採番する通常の成功形。
func savedReportMock(m *mockRepository) {
	m.createFn = func(_ context.Context, report *model.SupportBugReport) error {
		report.ID = 10
		return nil
	}
}

// ---- Create: Plane への外部送信は行わない（reviewable export は手動のみ） ----

func TestServiceCreate_DoesNotExportToPlane(t *testing.T) {
	// Plane 連携が設定されていても Create は外部へ何も送らない。
	// 外部エクスポートは EnsurePlaneTicket（POST /:id/plane-ticket の明示操作）に限定し、
	// 未分類コンテンツが自動で外部へ出る経路を閉じる。
	tests := []struct {
		name    string
		tickets *mockTicketCreator
	}{
		{name: "plane integration disabled", tickets: nil},
		{name: "plane integration configured", tickets: &mockTicketCreator{
			result: &PlaneIssue{ID: "issue-uuid", URL: "https://app.plane.so/ws/browse/EMR-1/"},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			savedReportMock(repo)

			var tickets TicketCreator
			if tt.tickets != nil {
				tickets = tt.tickets
			}
			svc := NewService(repo, tickets)

			report, err := svc.Create(context.Background(), 1, 5, CreateBugReportInput{Title: "受付でエラー"})
			require.NoError(t, err)
			require.NotNil(t, report)
			assert.Equal(t, uint64(10), report.ID)
			if tt.tickets != nil {
				assert.Zero(t, tt.tickets.calls)
			}
			assert.Nil(t, report.PlaneIssueID)
			assert.Nil(t, report.PlaneSyncError)
		})
	}
}

// ---- EnsurePlaneTicket: 手動起票/再送 ----

func TestServiceEnsurePlaneTicket(t *testing.T) {
	existingID := "issue-uuid-existing"
	tests := []struct {
		name          string
		report        *model.SupportBugReport
		findErr       error
		tickets       *mockTicketCreator
		setTicketFn   func(ctx context.Context, clinicID, id uint64, issueID, issueURL string) (bool, error)
		wantCalls     int
		wantErr       error
		wantIssueID   bool
		wantSyncError bool
	}{
		{
			name:        "returns existing ticket idempotently without calling plane",
			report:      &model.SupportBugReport{ID: 10, ClinicID: 1, PlaneIssueID: &existingID},
			tickets:     &mockTicketCreator{},
			wantCalls:   0,
			wantIssueID: true,
		},
		{
			name:    "returns not-implemented when plane is not configured",
			report:  &model.SupportBugReport{ID: 10, ClinicID: 1},
			tickets: nil,
			wantErr: apperrors.ErrNotImplemented,
		},
		{
			name:          "persists sync error and returns bad-gateway on upstream failure",
			report:        &model.SupportBugReport{ID: 10, ClinicID: 1},
			tickets:       &mockTicketCreator{err: &planeUpstreamError{status: 500}},
			wantCalls:     1,
			wantErr:       apperrors.ErrBadGateway,
			wantSyncError: true,
		},
		{
			name:    "claims ticket on success",
			report:  &model.SupportBugReport{ID: 10, ClinicID: 1},
			tickets: &mockTicketCreator{result: &PlaneIssue{ID: "new-uuid", URL: "https://app.plane.so/ws/browse/EMR-9/"}},
			setTicketFn: func(_ context.Context, _, _ uint64, _, _ string) (bool, error) {
				return true, nil
			},
			wantCalls:   1,
			wantIssueID: true,
		},
		{
			name:    "returns latest state when claim is lost",
			report:  &model.SupportBugReport{ID: 10, ClinicID: 1},
			tickets: &mockTicketCreator{result: &PlaneIssue{ID: "new-uuid", URL: "u"}},
			setTicketFn: func(_ context.Context, _, _ uint64, _, _ string) (bool, error) {
				return false, nil
			},
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				findByIDForClinicF: func(_ context.Context, clinicID, id uint64) (*model.SupportBugReport, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(10), id)
					if tt.findErr != nil {
						return nil, tt.findErr
					}
					return tt.report, nil
				},
				setSyncErrFn: func(_ context.Context, _, _ uint64, syncErr string) error {
					assert.NotEmpty(t, syncErr)
					return nil
				},
				setTicketFn: tt.setTicketFn,
			}

			var tickets TicketCreator
			if tt.tickets != nil {
				tickets = tt.tickets
			}
			svc := NewService(repo, tickets)

			report, err := svc.EnsurePlaneTicket(context.Background(), 1, 10)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr), "want %v, got %v", tt.wantErr, err)
				if tt.tickets != nil {
					assert.Equal(t, tt.wantCalls, tt.tickets.calls)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, report)
			if tt.tickets != nil {
				assert.Equal(t, tt.wantCalls, tt.tickets.calls)
			}
			assert.Equal(t, tt.wantIssueID, report.PlaneIssueID != nil)
			assert.Equal(t, tt.wantSyncError, report.PlaneSyncError != nil)
		})
	}
}

// ---- Delete: 論理削除 + 削除前行の返却 ----

func TestServiceDelete(t *testing.T) {
	t.Run("returns pre-delete report after soft delete", func(t *testing.T) {
		deleted := false
		repo := &mockRepository{
			findByIDForClinicF: func(_ context.Context, clinicID, _ uint64) (*model.SupportBugReport, error) {
				assert.Equal(t, uint64(1), clinicID)
				return &model.SupportBugReport{ID: 10, ClinicID: 1, Title: "対象"}, nil
			},
			softDeleteFn: func(_ context.Context, clinicID, id uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(10), id)
				deleted = true
				return nil
			},
		}
		svc := NewService(repo, nil)

		report, err := svc.Delete(context.Background(), 1, 10)
		require.NoError(t, err)
		assert.True(t, deleted)
		assert.Equal(t, uint64(10), report.ID)
	})

	t.Run("does not delete when report is missing", func(t *testing.T) {
		softDeleteCalled := false
		repo := &mockRepository{
			findByIDForClinicF: func(_ context.Context, _, _ uint64) (*model.SupportBugReport, error) {
				return nil, apperrors.WrapNotFound("support_bug_report", "99")
			},
			softDeleteFn: func(_ context.Context, _, _ uint64) error {
				softDeleteCalled = true
				return nil
			},
		}
		svc := NewService(repo, nil)

		_, err := svc.Delete(context.Background(), 2, 99)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.False(t, softDeleteCalled)
	})
}

func TestListChatExchanges(t *testing.T) {
	base := time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)
	msg := func(id uint64, role model.SupportChatMessageRole, content string, clinicID, staffID uint64, min int) ChatMessageWithMeta {
		return ChatMessageWithMeta{
			SupportChatMessage: model.SupportChatMessage{
				ID: id, ClinicID: clinicID, StaffID: staffID, Role: role, Content: content,
				CreatedAt: base.Add(time.Duration(min) * time.Minute),
			},
			StaffName:  "スタッフ" + string(rune('A')+rune(staffID-1)),
			ClinicName: "医院" + string(rune('A')+rune(clinicID-1)),
		}
	}

	t.Run("pairs user+assistant per clinic/staff and returns newest first", func(t *testing.T) {
		repo := &mockRepository{
			listAllChatFn: func(_ context.Context) ([]ChatMessageWithMeta, error) {
				// repo は新しい順で返す — id 降順（新→旧）
				return []ChatMessageWithMeta{
					msg(6, model.SupportChatRoleAssistant, "回答2", 1, 1, 30),
					msg(5, model.SupportChatRoleUser, "質問2", 1, 1, 29),
					msg(4, model.SupportChatRoleAssistant, "他院回答", 2, 7, 20),
					msg(3, model.SupportChatRoleUser, "他院質問", 2, 7, 19),
					msg(2, model.SupportChatRoleAssistant, "回答1", 1, 1, 10),
					msg(1, model.SupportChatRoleUser, "質問1", 1, 1, 9),
				}, nil
			},
		}
		svc := NewService(repo, nil)

		got, err := svc.ListChatExchanges(context.Background())
		require.NoError(t, err)
		require.Len(t, got, 3)
		// 新しい順: 質問2 → 他院 → 質問1
		assert.Equal(t, "質問2", got[0].UserMessage.Content)
		assert.Equal(t, "回答2", got[0].AssistantMessage.Content)
		assert.Equal(t, "他院質問", got[1].UserMessage.Content)
		assert.Equal(t, "医院B", got[1].ClinicName)
		assert.Equal(t, "質問1", got[2].UserMessage.Content)
	})

	t.Run("orphan assistant without preceding user is dropped", func(t *testing.T) {
		repo := &mockRepository{
			listAllChatFn: func(_ context.Context) ([]ChatMessageWithMeta, error) {
				return []ChatMessageWithMeta{
					msg(3, model.SupportChatRoleAssistant, "回答", 1, 1, 20),
					msg(2, model.SupportChatRoleUser, "質問", 1, 1, 19),
					msg(1, model.SupportChatRoleAssistant, "窓外の孤立回答", 1, 1, 5),
				}, nil
			},
		}
		svc := NewService(repo, nil)

		got, err := svc.ListChatExchanges(context.Background())
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "質問", got[0].UserMessage.Content)
	})

	t.Run("empty history returns empty list", func(t *testing.T) {
		svc := NewService(&mockRepository{}, nil)
		got, err := svc.ListChatExchanges(context.Background())
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

// ---- SyncPlaneTicketStates: Plane completed → resolved の一方向同期 ----

// mockTicketReader は TicketCreator + TicketStateReader の両方を満たす同期用スタブ。
type mockTicketReader struct {
	fetchFn func(ctx context.Context, issueID string) (string, error)
}

func (m *mockTicketReader) CreateBugReportIssue(context.Context, *model.SupportBugReport) (*PlaneIssue, error) {
	return nil, errors.New("not implemented")
}
func (m *mockTicketReader) FetchIssueStateGroup(ctx context.Context, issueID string) (string, error) {
	return m.fetchFn(ctx, issueID)
}

func strPtr(s string) *string { return &s }

func openReportWithTicket(id, clinicID uint64, issueID string) model.SupportBugReport {
	return model.SupportBugReport{
		ID:           id,
		ClinicID:     clinicID,
		Status:       model.SupportBugReportStatusOpen,
		PlaneIssueID: strPtr(issueID),
	}
}

func TestServiceSyncPlaneTicketStates_DisabledWhenNoReader(t *testing.T) {
	// TicketCreator のみ（reader 未実装 = 旧クライアント相当）は no-op。
	repo := &mockRepository{}
	svc := NewService(repo, &mockTicketCreator{})

	result := svc.SyncPlaneTicketStates(context.Background())

	assert.Equal(t, PlaneSyncResult{}, result)
	assert.Nil(t, repo.listOpenPlaneFn, "repo must not be queried")
}

func TestServiceSyncPlaneTicketStates_NilTicketsIsNoOp(t *testing.T) {
	svc := NewService(&mockRepository{}, nil)
	assert.Equal(t, PlaneSyncResult{}, svc.SyncPlaneTicketStates(context.Background()))
}

func TestServiceSyncPlaneTicketStates_ListErrorFailsClosed(t *testing.T) {
	repo := &mockRepository{
		listOpenPlaneFn: func(context.Context, int) ([]model.SupportBugReport, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewService(repo, &mockTicketReader{fetchFn: func(context.Context, string) (string, error) {
		return "completed", nil
	}})

	result := svc.SyncPlaneTicketStates(context.Background())

	// 対象一覧すら取れない = ジョブ全体失敗として 1 件の failed を報告する。
	assert.Equal(t, PlaneSyncResult{Processed: 1, Failed: 1}, result)
}

func TestServiceSyncPlaneTicketStates_CompletedResolvesWithReportClinic(t *testing.T) {
	var updateClinic, updateID uint64
	var updateStatus model.SupportBugReportStatus
	repo := &mockRepository{
		listOpenPlaneFn: func(_ context.Context, limit int) ([]model.SupportBugReport, error) {
			assert.Equal(t, planeStateSyncBatchLimit, limit)
			return []model.SupportBugReport{openReportWithTicket(7, 42, "issue-uuid-1")}, nil
		},
		updateStatusFn: func(_ context.Context, clinicID, id uint64, status model.SupportBugReportStatus) error {
			updateClinic, updateID, updateStatus = clinicID, id, status
			return nil
		},
	}
	var fetchedID string
	svc := NewService(repo, &mockTicketReader{fetchFn: func(_ context.Context, issueID string) (string, error) {
		fetchedID = issueID
		return "completed", nil
	}})

	result := svc.SyncPlaneTicketStates(context.Background())

	assert.Equal(t, PlaneSyncResult{Processed: 1, Succeeded: 1}, result)
	assert.Equal(t, "issue-uuid-1", fetchedID)
	// UpdateStatus は必ず報告の clinic_id スコープで行う（cross-clinic 更新にならない）。
	assert.Equal(t, uint64(42), updateClinic)
	assert.Equal(t, uint64(7), updateID)
	assert.Equal(t, model.SupportBugReportStatusResolved, updateStatus)
}

func TestServiceSyncPlaneTicketStates_NonCompletedSkipped(t *testing.T) {
	for _, group := range []string{"started", "backlog", "cancelled", ""} {
		t.Run("group="+group, func(t *testing.T) {
			updateCalled := false
			repo := &mockRepository{
				listOpenPlaneFn: func(context.Context, int) ([]model.SupportBugReport, error) {
					return []model.SupportBugReport{openReportWithTicket(7, 42, "issue-1")}, nil
				},
				updateStatusFn: func(context.Context, uint64, uint64, model.SupportBugReportStatus) error {
					updateCalled = true
					return nil
				},
			}
			svc := NewService(repo, &mockTicketReader{fetchFn: func(context.Context, string) (string, error) {
				return group, nil
			}})

			result := svc.SyncPlaneTicketStates(context.Background())

			// 未完了スキップは失敗ではなく succeeded に計上（対象を正しく処理済み）。
			assert.Equal(t, PlaneSyncResult{Processed: 1, Succeeded: 1}, result)
			assert.False(t, updateCalled)
		})
	}
}

func TestServiceSyncPlaneTicketStates_FetchErrorRecordsSyncError(t *testing.T) {
	var syncErrClinic, syncErrID uint64
	var syncErrMsg string
	repo := &mockRepository{
		listOpenPlaneFn: func(context.Context, int) ([]model.SupportBugReport, error) {
			return []model.SupportBugReport{openReportWithTicket(9, 5, "issue-gone")}, nil
		},
		setSyncErrFn: func(_ context.Context, clinicID, id uint64, msg string) error {
			syncErrClinic, syncErrID, syncErrMsg = clinicID, id, msg
			return nil
		},
		updateStatusFn: func(context.Context, uint64, uint64, model.SupportBugReportStatus) error {
			t.Fatal("UpdateStatus must not run on fetch failure")
			return nil
		},
	}
	svc := NewService(repo, &mockTicketReader{fetchFn: func(context.Context, string) (string, error) {
		return "", &planeUpstreamError{status: 404}
	}})

	result := svc.SyncPlaneTicketStates(context.Background())

	assert.Equal(t, PlaneSyncResult{Processed: 1, Failed: 1}, result)
	// 404 も他の upstream 失敗と同じく plane_sync_error に記録（報告行から追跡可能）。
	assert.Equal(t, uint64(5), syncErrClinic)
	assert.Equal(t, uint64(9), syncErrID)
	assert.Equal(t, "plane api error (status 404)", syncErrMsg)
}

func TestServiceSyncPlaneTicketStates_UpdateErrorCountsFailed(t *testing.T) {
	repo := &mockRepository{
		listOpenPlaneFn: func(context.Context, int) ([]model.SupportBugReport, error) {
			return []model.SupportBugReport{openReportWithTicket(7, 42, "issue-1")}, nil
		},
		updateStatusFn: func(context.Context, uint64, uint64, model.SupportBugReportStatus) error {
			return errors.New("write conflict")
		},
	}
	svc := NewService(repo, &mockTicketReader{fetchFn: func(context.Context, string) (string, error) {
		return "completed", nil
	}})

	result := svc.SyncPlaneTicketStates(context.Background())
	assert.Equal(t, PlaneSyncResult{Processed: 1, Failed: 1}, result)
}

func TestServiceSyncPlaneTicketStates_MixedBatch(t *testing.T) {
	var resolved []uint64
	var syncErrs []uint64
	repo := &mockRepository{
		listOpenPlaneFn: func(context.Context, int) ([]model.SupportBugReport, error) {
			return []model.SupportBugReport{
				openReportWithTicket(1, 10, "done-1"),
				openReportWithTicket(2, 20, "wip-2"),
				openReportWithTicket(3, 30, "gone-3"),
			}, nil
		},
		updateStatusFn: func(_ context.Context, _, id uint64, _ model.SupportBugReportStatus) error {
			resolved = append(resolved, id)
			return nil
		},
		setSyncErrFn: func(_ context.Context, _, id uint64, _ string) error {
			syncErrs = append(syncErrs, id)
			return nil
		},
	}
	svc := NewService(repo, &mockTicketReader{fetchFn: func(_ context.Context, issueID string) (string, error) {
		switch issueID {
		case "done-1":
			return "completed", nil
		case "wip-2":
			return "started", nil
		default:
			return "", errors.New("boom")
		}
	}})

	result := svc.SyncPlaneTicketStates(context.Background())

	assert.Equal(t, PlaneSyncResult{Processed: 3, Succeeded: 2, Failed: 1}, result)
	assert.Equal(t, []uint64{1}, resolved)
	assert.Equal(t, []uint64{3}, syncErrs)
}
