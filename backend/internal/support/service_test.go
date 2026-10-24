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
	softDeleteFn       func(ctx context.Context, clinicID, id uint64) error
	createChatFn       func(ctx context.Context, messages []*model.SupportChatMessage) error
	listChatFn         func(ctx context.Context, clinicID, staffID uint64) ([]model.SupportChatMessage, error)
	listAllChatFn      func(ctx context.Context) ([]ChatMessageWithMeta, error)
	clearChatFn        func(ctx context.Context, clinicID, staffID uint64) error
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

// ---- mock TicketCreator ----

type mockTicketCreator struct {
	calls  int
	fn     func(ctx context.Context, report *model.SupportBugReport) // 呼び出し観測フック（順序・引数検証）
	result *PlaneIssue
	err    error
}

func (m *mockTicketCreator) CreateBugReportIssue(ctx context.Context, report *model.SupportBugReport) (*PlaneIssue, error) {
	m.calls++
	if m.fn != nil {
		m.fn(ctx, report)
	}
	return m.result, m.err
}

// savedReportMock は repo.Create が report に ID=10 を採番する通常の成功形。
func savedReportMock(m *mockRepository) {
	m.createFn = func(_ context.Context, report *model.SupportBugReport) error {
		report.ID = 10
		return nil
	}
}

// ---- mock BugReportNotifier ----

type mockBugReportNotifier struct {
	calls int
	err   error
	order *[]string // repo.Create との呼び出し順序検証用
}

func (m *mockBugReportNotifier) NotifyBugReport(_ context.Context, _ *model.SupportBugReport) error {
	m.calls++
	if m.order != nil {
		*m.order = append(*m.order, "notify")
	}
	return m.err
}

// ---- Create: Discord 通知は DB 保存前に best-effort で実行 ----

func TestServiceCreate_NotifiesDiscordBeforeSave(t *testing.T) {
	repo := &mockRepository{}
	var order []string
	notifier := &mockBugReportNotifier{order: &order}
	repo.createFn = func(_ context.Context, report *model.SupportBugReport) error {
		order = append(order, "create")
		report.ID = 10
		return nil
	}
	svc := NewService(repo, nil, notifier)

	report, err := svc.Create(context.Background(), 1, 5, CreateBugReportInput{
		Title:  "受付でエラー",
		Detail: "保存ボタンを押すと500になる",
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(10), report.ID)
	assert.Equal(t, 1, notifier.calls)
	assert.Equal(t, []string{"notify", "create"}, order)
}

func TestServiceCreate_DiscordFailureStillSaves(t *testing.T) {
	// Discord 送信失敗は WARN のみ — 報告作成自体は成功させる（通知障害で報告を失わせない）。
	repo := &mockRepository{}
	savedReportMock(repo)
	notifier := &mockBugReportNotifier{err: errors.New("discord webhook returned status 500")}
	svc := NewService(repo, nil, notifier)

	report, err := svc.Create(context.Background(), 1, 5, CreateBugReportInput{Title: "件名"})
	require.NoError(t, err)
	require.NotNil(t, report)
	assert.Equal(t, uint64(10), report.ID)
	assert.Equal(t, 1, notifier.calls)
}

// ---- Create: 配送契約の接続まで Plane へ送らない（EMR-549 C1） ----

func TestServiceCreate_DoesNotSendToPlaneBeforeDeliveryContract(t *testing.T) {
	repo := &mockRepository{}
	savedReportMock(repo)
	tickets := &mockTicketCreator{result: &PlaneIssue{ID: "must-not-be-created"}}
	svc := NewService(repo, tickets, nil)

	report, err := svc.Create(context.Background(), 1, 5, CreateBugReportInput{Title: "件名"})

	require.NoError(t, err)
	require.NotNil(t, report)
	assert.Equal(t, uint64(10), report.ID)
	assert.Equal(t, 0, tickets.calls)
}

// ---- EnsurePlaneTicket: 配送契約の接続まで固定の 422 で停止 ----

func TestServiceEnsurePlaneTicket_StopsWithStableCode(t *testing.T) {
	tickets := &mockTicketCreator{result: &PlaneIssue{ID: "must-not-be-created"}}
	repo := &mockRepository{
		findByIDForClinicF: func(_ context.Context, clinicID, id uint64) (*model.SupportBugReport, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(10), id)
			return &model.SupportBugReport{ID: 10, ClinicID: 1}, nil
		},
	}
	svc := NewService(repo, tickets, nil)

	report, err := svc.EnsurePlaneTicket(context.Background(), 1, 10)

	require.Error(t, err)
	assert.Nil(t, report)
	assert.ErrorIs(t, err, apperrors.ErrUnprocessable)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, codeSupportDeliveryPending, appErr.Code)
	assert.Equal(t, 0, tickets.calls, "plane must not be called before the delivery contract is connected")
}

func TestServiceEnsurePlaneTicket_ForeignOrMissingReportIsNotFound(t *testing.T) {
	tickets := &mockTicketCreator{}
	repo := &mockRepository{
		findByIDForClinicF: func(context.Context, uint64, uint64) (*model.SupportBugReport, error) {
			return nil, apperrors.WrapNotFound("support_bug_report", "10")
		},
	}
	svc := NewService(repo, tickets, nil)

	_, err := svc.EnsurePlaneTicket(context.Background(), 2, 10)

	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Equal(t, 0, tickets.calls)
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
		svc := NewService(repo, nil, nil)

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
		svc := NewService(repo, nil, nil)

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
		svc := NewService(repo, nil, nil)

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
		svc := NewService(repo, nil, nil)

		got, err := svc.ListChatExchanges(context.Background())
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "質問", got[0].UserMessage.Content)
	})

	t.Run("empty history returns empty list", func(t *testing.T) {
		svc := NewService(&mockRepository{}, nil, nil)
		got, err := svc.ListChatExchanges(context.Background())
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestListChatRooms(t *testing.T) {
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

	t.Run("同一スタッフで間隔がchatRoomGap以下なら同一ルームにまとまる", func(t *testing.T) {
		repo := &mockRepository{
			listAllChatFn: func(_ context.Context) ([]ChatMessageWithMeta, error) {
				return []ChatMessageWithMeta{
					msg(4, model.SupportChatRoleAssistant, "回答2", 1, 1, 40),
					msg(3, model.SupportChatRoleUser, "質問2", 1, 1, 39),
					msg(2, model.SupportChatRoleAssistant, "回答1", 1, 1, 10),
					msg(1, model.SupportChatRoleUser, "質問1", 1, 1, 9),
				}, nil
			},
		}
		svc := NewService(repo, nil, nil)

		rooms, err := svc.ListChatRooms(context.Background())
		require.NoError(t, err)
		require.Len(t, rooms, 1)
		assert.Equal(t, "1-1-1", rooms[0].ID)
		assert.Equal(t, "医院A", rooms[0].ClinicName)
		assert.Equal(t, "スタッフA", rooms[0].StaffName)
		require.Len(t, rooms[0].Turns, 4)
		// ルーム内は古い順
		assert.Equal(t, "質問1", rooms[0].Turns[0].Content)
		assert.Equal(t, "回答2", rooms[0].Turns[3].Content)
		assert.Equal(t, base.Add(9*time.Minute), rooms[0].StartedAt)
		assert.Equal(t, base.Add(40*time.Minute), rooms[0].LastMessageAt)
	})

	t.Run("chatRoomGap超過で別ルームに分かれる", func(t *testing.T) {
		repo := &mockRepository{
			listAllChatFn: func(_ context.Context) ([]ChatMessageWithMeta, error) {
				return []ChatMessageWithMeta{
					msg(4, model.SupportChatRoleAssistant, "回答2", 1, 1, 100),
					msg(3, model.SupportChatRoleUser, "質問2", 1, 1, 99),
					msg(2, model.SupportChatRoleAssistant, "回答1", 1, 1, 10),
					msg(1, model.SupportChatRoleUser, "質問1", 1, 1, 9),
				}, nil
			},
		}
		svc := NewService(repo, nil, nil)

		rooms, err := svc.ListChatRooms(context.Background())
		require.NoError(t, err)
		require.Len(t, rooms, 2)
		// 最終発言が新しいルームが先頭
		assert.Equal(t, "1-1-3", rooms[0].ID)
		require.Len(t, rooms[0].Turns, 2)
		assert.Equal(t, "質問2", rooms[0].Turns[0].Content)
		assert.Equal(t, "1-1-1", rooms[1].ID)
		require.Len(t, rooms[1].Turns, 2)
	})

	t.Run("clinic×staff が違えば別ルーム — 同一時間帯でも混ざらない", func(t *testing.T) {
		repo := &mockRepository{
			listAllChatFn: func(_ context.Context) ([]ChatMessageWithMeta, error) {
				return []ChatMessageWithMeta{
					msg(4, model.SupportChatRoleAssistant, "他院回答", 2, 7, 21),
					msg(3, model.SupportChatRoleUser, "他院質問", 2, 7, 20),
					msg(2, model.SupportChatRoleAssistant, "回答1", 1, 1, 10),
					msg(1, model.SupportChatRoleUser, "質問1", 1, 1, 9),
				}, nil
			},
		}
		svc := NewService(repo, nil, nil)

		rooms, err := svc.ListChatRooms(context.Background())
		require.NoError(t, err)
		require.Len(t, rooms, 2)
		assert.Equal(t, "医院B", rooms[0].ClinicName)
		assert.Equal(t, "スタッフG", rooms[0].StaffName)
		assert.Equal(t, "医院A", rooms[1].ClinicName)
	})

	t.Run("空履歴は空リスト", func(t *testing.T) {
		svc := NewService(&mockRepository{}, nil, nil)
		rooms, err := svc.ListChatRooms(context.Background())
		require.NoError(t, err)
		assert.Empty(t, rooms)
	})

	t.Run("repo エラーはそのまま返す", func(t *testing.T) {
		repo := &mockRepository{
			listAllChatFn: func(context.Context) ([]ChatMessageWithMeta, error) {
				return nil, assert.AnError
			},
		}
		svc := NewService(repo, nil, nil)
		_, err := svc.ListChatRooms(context.Background())
		assert.ErrorIs(t, err, assert.AnError)
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

func TestServiceSyncPlaneTicketStates_DisabledWhenNoReader(t *testing.T) {
	// TicketCreator のみ（reader 未実装 = 旧クライアント相当）は no-op。
	svc := NewService(&mockRepository{}, &mockTicketCreator{}, nil)

	result := svc.SyncPlaneTicketStates(context.Background())

	assert.Equal(t, PlaneSyncResult{}, result)
}

func TestServiceSyncPlaneTicketStates_NilTicketsIsNoOp(t *testing.T) {
	svc := NewService(&mockRepository{}, nil, nil)
	assert.Equal(t, PlaneSyncResult{}, svc.SyncPlaneTicketStates(context.Background()))
}

func TestServiceSyncPlaneTicketStates_ReaderConfiguredFailsClosed(t *testing.T) {
	// 配送結果 reader が未接続の間は、外部 ID を持たない報告を「未送信」と推測して空振りさせず、
	// 実行ごとに 1 件の失敗として報告する（Processed == Succeeded + Failed を維持）。
	fetched := false
	svc := NewService(&mockRepository{}, &mockTicketReader{fetchFn: func(context.Context, string) (string, error) {
		fetched = true
		return "completed", nil
	}}, nil)

	result := svc.SyncPlaneTicketStates(context.Background())

	assert.Equal(t, PlaneSyncResult{Processed: 1, Failed: 1}, result)
	assert.False(t, fetched, "plane must not be queried without delivery results")
}
