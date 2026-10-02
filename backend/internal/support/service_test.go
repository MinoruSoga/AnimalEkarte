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
	createFn       func(ctx context.Context, report *model.SupportBugReport) error
	findAllFn      func(ctx context.Context) ([]BugReportWithReporter, error)
	findByIDFn     func(ctx context.Context, id uint64) (*model.SupportBugReport, error)
	updateStatusFn func(ctx context.Context, id uint64, status model.SupportBugReportStatus) error
	setTicketFn    func(ctx context.Context, id uint64, issueID, issueURL string) (bool, error)
	setSyncErrFn   func(ctx context.Context, id uint64, syncErr string) error
	softDeleteFn   func(ctx context.Context, id uint64) error
	createChatFn   func(ctx context.Context, messages []*model.SupportChatMessage) error
	listChatFn     func(ctx context.Context, clinicID, staffID uint64) ([]model.SupportChatMessage, error)
	listAllChatFn  func(ctx context.Context) ([]ChatMessageWithMeta, error)
	clearChatFn    func(ctx context.Context, clinicID, staffID uint64) error
}

func (m *mockRepository) Create(ctx context.Context, report *model.SupportBugReport) error {
	return m.createFn(ctx, report)
}
func (m *mockRepository) FindAll(ctx context.Context) ([]BugReportWithReporter, error) {
	return m.findAllFn(ctx)
}
func (m *mockRepository) FindByID(ctx context.Context, id uint64) (*model.SupportBugReport, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockRepository) UpdateStatus(ctx context.Context, id uint64, status model.SupportBugReportStatus) error {
	return m.updateStatusFn(ctx, id, status)
}
func (m *mockRepository) SetPlaneTicket(ctx context.Context, id uint64, issueID, issueURL string) (bool, error) {
	return m.setTicketFn(ctx, id, issueID, issueURL)
}
func (m *mockRepository) SetPlaneSyncError(ctx context.Context, id uint64, syncErr string) error {
	return m.setSyncErrFn(ctx, id, syncErr)
}
func (m *mockRepository) SoftDeleteBugReport(ctx context.Context, id uint64) error {
	return m.softDeleteFn(ctx, id)
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

// ---- Create: Plane 連携は best-effort ----

func TestServiceCreate_PlaneSync(t *testing.T) {
	tests := []struct {
		name            string
		tickets         *mockTicketCreator
		setTicketResult bool
		setTicketErr    error
		wantCalls       int
		wantIssueID     bool
		wantSyncErr     bool
	}{
		{
			name:    "local save only when plane integration disabled",
			tickets: nil,
		},
		{
			name: "records plane fields on successful creation",
			tickets: &mockTicketCreator{
				result: &PlaneIssue{ID: "issue-uuid", URL: "https://app.plane.so/ws/browse/EMR-1/"},
			},
			setTicketResult: true,
			wantCalls:       1,
			wantIssueID:     true,
		},
		{
			name:        "records sync error but keeps report on plane failure",
			tickets:     &mockTicketCreator{err: &planeUpstreamError{status: 503}},
			wantCalls:   1,
			wantSyncErr: true,
		},
		{
			name: "does not set fields when claim lost to a concurrent creator",
			tickets: &mockTicketCreator{
				result: &PlaneIssue{ID: "issue-uuid", URL: "https://app.plane.so/ws/browse/EMR-1/"},
			},
			setTicketResult: false,
			wantCalls:       1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			savedReportMock(repo)
			repo.setTicketFn = func(_ context.Context, id uint64, issueID, issueURL string) (bool, error) {
				assert.Equal(t, uint64(10), id)
				return tt.setTicketResult, tt.setTicketErr
			}
			repo.setSyncErrFn = func(_ context.Context, _ uint64, syncErr string) error {
				assert.NotEmpty(t, syncErr)
				return nil
			}

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
				assert.Equal(t, tt.wantCalls, tt.tickets.calls)
			}
			assert.Equal(t, tt.wantIssueID, report.PlaneIssueID != nil)
			assert.Equal(t, tt.wantSyncErr, report.PlaneSyncError != nil)
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
		setTicketFn   func(ctx context.Context, id uint64, issueID, issueURL string) (bool, error)
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
			setTicketFn: func(_ context.Context, _ uint64, _, _ string) (bool, error) {
				return true, nil
			},
			wantCalls:   1,
			wantIssueID: true,
		},
		{
			name:    "returns latest state when claim is lost",
			report:  &model.SupportBugReport{ID: 10, ClinicID: 1},
			tickets: &mockTicketCreator{result: &PlaneIssue{ID: "new-uuid", URL: "u"}},
			setTicketFn: func(_ context.Context, _ uint64, _, _ string) (bool, error) {
				return false, nil
			},
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				findByIDFn: func(_ context.Context, id uint64) (*model.SupportBugReport, error) {
					if tt.findErr != nil {
						return nil, tt.findErr
					}
					return tt.report, nil
				},
				setSyncErrFn: func(_ context.Context, _ uint64, syncErr string) error {
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

			report, err := svc.EnsurePlaneTicket(context.Background(), 10)
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
			findByIDFn: func(_ context.Context, _ uint64) (*model.SupportBugReport, error) {
				return &model.SupportBugReport{ID: 10, ClinicID: 1, Title: "対象"}, nil
			},
			softDeleteFn: func(_ context.Context, id uint64) error {
				assert.Equal(t, uint64(10), id)
				deleted = true
				return nil
			},
		}
		svc := NewService(repo, nil)

		report, err := svc.Delete(context.Background(), 10)
		require.NoError(t, err)
		assert.True(t, deleted)
		assert.Equal(t, uint64(10), report.ID)
	})

	t.Run("does not delete when report is missing", func(t *testing.T) {
		softDeleteCalled := false
		repo := &mockRepository{
			findByIDFn: func(_ context.Context, _ uint64) (*model.SupportBugReport, error) {
				return nil, apperrors.WrapNotFound("support_bug_report", "99")
			},
			softDeleteFn: func(_ context.Context, _ uint64) error {
				softDeleteCalled = true
				return nil
			},
		}
		svc := NewService(repo, nil)

		_, err := svc.Delete(context.Background(), 99)
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
