package support

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CreateBugReportInput はサービス層への作成入力（handler で検証済みの値）
type CreateBugReportInput struct {
	Title         string
	Detail        string
	PageURL       string
	RoutePath     string
	UserAgent     string
	Viewport      string
	AppVersion    string
	ScreenshotKey *string
}

// Service はバグ報告とヘルプチャット履歴のユースケースインターフェース。
// バグ報告の一覧・作成は全医院横断だが、status 更新・Plane 起票・削除は
// 報告元 clinicID スコープ（共有ボードは閲覧のみ公開）。Create は provenance
// として clinic_id/reporter_staff_id を記録する。チャット履歴のみ
// clinic×staff スコープ。
type Service interface {
	Create(ctx context.Context, clinicID, reporterStaffID uint64, input CreateBugReportInput) (*model.SupportBugReport, error)
	ListAll(ctx context.Context) ([]BugReportWithReporter, error)
	// Get は報告元 clinicID スコープで 1 件を返す。他医院は NotFound（404）。
	Get(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error)
	UpdateStatus(ctx context.Context, clinicID, id uint64, status model.SupportBugReportStatus) (*model.SupportBugReport, error)
	EnsurePlaneTicket(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error)
	Delete(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error)
	ListChatHistory(ctx context.Context, clinicID, staffID uint64) ([]model.SupportChatMessage, error)
	// ListChatExchanges は全医院の質問+回答ペアを新しい順で返す（共有一覧ページ用）
	ListChatExchanges(ctx context.Context) ([]ChatExchange, error)
	// ListChatRooms は全医院の発言ストリームを疑似会話ルーム単位で返す（共有一覧ページ用）
	ListChatRooms(ctx context.Context) ([]ChatRoom, error)
	RecordChatExchange(ctx context.Context, clinicID, staffID uint64, userMessage, assistantReply string, sources []ChatSource) error
	ClearChatHistory(ctx context.Context, clinicID, staffID uint64) error
	// SyncPlaneTicketStates は scheduled job 用 — Plane 側 completed → resolved の
	// 一方向同期。全医院横断で動き、各報告の clinic_id で UpdateStatus する。
	SyncPlaneTicketStates(ctx context.Context) PlaneSyncResult
}

const (
	// planeAutoAuditAction は自動起票フロー専用の監査アクション。
	// 手動起票は handler 側の既存監査（support_bug_report.plane_ticket）で
	// 記録されるため、本アクションは post-save 起票のみが出力する。
	planeAutoAuditAction = "support_bug_report.plane_ticket_auto"

	// 自動起票監査の outcome 分類（固定値のみ。任意文字列は埋め込まない）。
	planeAutoOutcomeClaimed        = "claimed"
	planeAutoOutcomeCreationFailed = "creation_failed"
	planeAutoOutcomeClaimLost      = "claim_lost"
	planeAutoOutcomeClaimError     = "claim_error"
)

const (
	// defaultPlanePostSaveTimeout は保存後 best-effort 起票の外部呼出 budget
	// 既定値（PLANE_TIMEOUT_MS の既定 10000ms に合わせる）。
	defaultPlanePostSaveTimeout = 10 * time.Second
	// planePostSaveDBTimeout / planePostSaveAuditTimeout は保存後処理の
	// DB 操作・監査書込みそれぞれの個別 budget。外部 HTTP の deadline とは
	// 分離する（HTTP timeout が失敗理由の永続化まで潰さないため）。
	planePostSaveDBTimeout    = 5 * time.Second
	planePostSaveAuditTimeout = 5 * time.Second
)

type service struct {
	repo     Repository
	tickets  TicketCreator     // nil = Plane 連携無効（ローカル保存のみ）
	notifier BugReportNotifier // nil = Discord 通知無効
	// audit は自動 Plane 起票結果の専用監査記録先（nil = 記録しない）。
	audit AuditLogger
	// planeTimeout は保存後 best-effort 起票の外部呼出 timeout。
	planeTimeout time.Duration
}

// ServiceOption は NewService へのオプション依存を注入する。
type ServiceOption func(*service)

// WithAuditLogger は自動 Plane 起票結果を記録する監査ロガーを設定する。
// 監査書込み自体も best-effort であり、失敗は ERROR ログに留めて
// 報告作成の成功を変えない。
func WithAuditLogger(logger AuditLogger) ServiceOption {
	return func(s *service) { s.audit = logger }
}

// WithPlaneTimeout は保存後 Plane 起票の外部呼出 timeout を設定する。
// 0 以下は無視して既定値を維持する。
func WithPlaneTimeout(timeout time.Duration) ServiceOption {
	return func(s *service) {
		if timeout > 0 {
			s.planeTimeout = timeout
		}
	}
}

// NewService は Service を初期化して返す
func NewService(repo Repository, tickets TicketCreator, notifier BugReportNotifier, opts ...ServiceOption) Service {
	s := &service{
		repo:         repo,
		tickets:      tickets,
		notifier:     notifier,
		planeTimeout: defaultPlanePostSaveTimeout,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *service) Create(ctx context.Context, clinicID, reporterStaffID uint64, input CreateBugReportInput) (*model.SupportBugReport, error) {
	report := model.SupportBugReport{
		ClinicID:        clinicID,
		ReporterStaffID: reporterStaffID,
		Title:           input.Title,
		Detail:          input.Detail,
		PageURL:         input.PageURL,
		RoutePath:       input.RoutePath,
		UserAgent:       input.UserAgent,
		Viewport:        input.Viewport,
		AppVersion:      input.AppVersion,
		ScreenshotKey:   input.ScreenshotKey,
		Status:          model.SupportBugReportStatusOpen,
	}
	// Discord 通知は DB 保存前に best-effort で実行する（報告原文の外部転写を優先）。
	// 失敗しても報告作成は続行する（Discord 障害で報告自体を失わせない）。
	// ID は未採番のため通知に含まれない。
	if s.notifier != nil {
		if err := s.notifier.NotifyBugReport(ctx, &report); err != nil {
			slog.WarnContext(ctx, "failed to send bug report discord notification",
				"clinic_id", clinicID, "reporter_staff_id", reporterStaffID, "error", err)
		}
	}
	if err := s.repo.Create(ctx, &report); err != nil {
		return nil, err
	}
	// 保存成功後に Plane へ best-effort で自動起票する（2026-10-11 owner 指示で復帰）。
	// 起票・記録の失敗は plane_sync_error / ログに留め、作成応答は成功のまま返す。
	if s.tickets != nil {
		s.syncPlaneTicket(ctx, &report)
	}
	return &report, nil
}

// syncPlaneTicket は保存済み報告の Plane 起票を best-effort で試行する。
// request ctx のキャンセルからは切り離す（WithoutCancel — context 値は引き継ぐ）
// が、goroutine は使わず同期的に実行する。外部 HTTP・DB 永続化・監査書込みは
// それぞれ fresh bounded context を張り、外部呼出の deadline が失敗理由の
// 永続化まで潰さないようにする。
// 上流障害は recordPlaneFailure が plane_sync_error に記録する。起票自体は
// claim 前に Plane 側へ作成済みとなるため、claim 敗退（並行起票の後着や
// DB 書き込み失敗）では孤立 issue が残りうる既知の制限がある — 呼出側の
// 作成成功にはいずれも影響しない。
func (s *service) syncPlaneTicket(requestCtx context.Context, report *model.SupportBugReport) {
	postSaveCtx := context.WithoutCancel(requestCtx)

	planeCtx, cancel := context.WithTimeout(postSaveCtx, s.planeTimeout)
	issue, err := s.tickets.CreateBugReportIssue(planeCtx, report)
	cancel()
	if err != nil {
		s.recordPlaneFailure(postSaveCtx, report, err)
		s.logPlaneAutoAudit(postSaveCtx, report, planeAutoOutcomeCreationFailed, "")
		return
	}
	claimed, claimErr := s.claimPlaneTicket(postSaveCtx, report, issue)
	switch {
	case claimed:
		s.logPlaneAutoAudit(postSaveCtx, report, planeAutoOutcomeClaimed, issue.ID)
	case claimErr != nil:
		// 外部には作成済みだがローカル claim を永続化できなかった孤立 issue。
		s.logPlaneAutoAudit(postSaveCtx, report, planeAutoOutcomeClaimError, issue.ID)
	default:
		// 並行経路が先に claim した。こちら側の外部 issue も孤立しうるため
		// 外部 ID を残して追跡可能にする。
		s.logPlaneAutoAudit(postSaveCtx, report, planeAutoOutcomeClaimLost, issue.ID)
	}
}

func (s *service) ListAll(ctx context.Context) ([]BugReportWithReporter, error) {
	return s.repo.FindAll(ctx)
}

func (s *service) Get(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error) {
	return s.repo.FindByIDForClinic(ctx, clinicID, id)
}

func (s *service) UpdateStatus(ctx context.Context, clinicID, id uint64, status model.SupportBugReportStatus) (*model.SupportBugReport, error) {
	if !model.IsValidSupportBugReportStatus(string(status)) {
		return nil, apperrors.WrapInvalidInput("invalid status")
	}
	if err := s.repo.UpdateStatus(ctx, clinicID, id, status); err != nil {
		return nil, err
	}
	return s.repo.FindByIDForClinic(ctx, clinicID, id)
}

// EnsurePlaneTicket は一覧からの手動起票・再送。起票済みなら現行値をそのまま返す（冪等）。
// Plane 未設定なら 501、起票失敗は同期状態を記録した上で 502 を返す。
func (s *service) EnsurePlaneTicket(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error) {
	report, err := s.repo.FindByIDForClinic(ctx, clinicID, id)
	if err != nil {
		return nil, err
	}
	if report.PlaneIssueID != nil {
		return report, nil
	}
	if s.tickets == nil {
		return nil, apperrors.WrapNotImplemented("plane integration is not configured")
	}
	issue, err := s.tickets.CreateBugReportIssue(ctx, report)
	if err != nil {
		if s.recordPlaneFailure(ctx, report, err) {
			return report, nil
		}
		return nil, apperrors.WrapBadGateway("plane ticket creation failed")
	}
	if claimed, _ := s.claimPlaneTicket(ctx, report, issue); claimed {
		return report, nil
	}
	// claim 失敗（並行起票の後着 or DB 書き込み失敗）は最新状態を返して実態に合わせる。
	return s.repo.FindByIDForClinic(ctx, clinicID, id)
}

// Delete は報告を論理削除する。返り値は削除前に読み取った行で、
// handler がスクショの後始末と監査記録に使う。
func (s *service) Delete(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error) {
	report, err := s.repo.FindByIDForClinic(ctx, clinicID, id)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SoftDeleteBugReport(ctx, clinicID, id); err != nil {
		return nil, err
	}
	return report, nil
}

// recordPlaneFailure は起票失敗を plane_sync_error に記録し WARN を残す。
// 並行した成功claimを再取得できた場合は true。記録・再取得失敗は ERROR ログで可視化する。
func (s *service) recordPlaneFailure(ctx context.Context, report *model.SupportBugReport, syncErr error) bool {
	if s.persistPlaneFailure(ctx, report, planeSyncErrorMessage(syncErr)) {
		return true
	}
	slog.WarnContext(ctx, "plane ticket creation failed", "error", syncErr, "report_id", report.ID, "clinic_id", report.ClinicID)
	return false
}

// persistPlaneFailure records a creation failure only while the report is unclaimed.
// A guarded no-op reloads the clinic-scoped winner so responses reflect its success.
// Persistence/reload failures remain visible in logs; automatic creation stays best-effort.
// Each DB operation gets its own fresh bounded context so an exhausted external
// HTTP deadline cannot poison the failure record.
func (s *service) persistPlaneFailure(ctx context.Context, report *model.SupportBugReport, msg string) bool {
	writeCtx, cancel := context.WithTimeout(ctx, planePostSaveDBTimeout)
	written, err := s.repo.SetPlaneSyncError(writeCtx, report.ClinicID, report.ID, msg)
	cancel()
	if err != nil {
		slog.ErrorContext(ctx, "failed to persist plane sync error", "error", err, "report_id", report.ID)
	} else if !written {
		reloadCtx, reloadCancel := context.WithTimeout(ctx, planePostSaveDBTimeout)
		latest, reloadErr := s.repo.FindByIDForClinic(reloadCtx, report.ClinicID, report.ID)
		reloadCancel()
		if reloadErr != nil {
			slog.ErrorContext(ctx, "failed to reload competing plane claim", "error", reloadErr, "report_id", report.ID)
		} else if latest != nil && latest.PlaneIssueID != nil {
			*report = *latest
			return true
		}
	}
	report.PlaneSyncError = &msg
	return false
}

// claimPlaneTicket は起票成功を DB に記録する。claim できた場合のみ (true, nil)。
// (false, nil) は別経路が先に記録済み（二重起票の防止）。err 非 nil は
// DB 書き込み失敗 — Plane 側に孤立チケットが残りうるため ERROR ログを残す。
func (s *service) claimPlaneTicket(ctx context.Context, report *model.SupportBugReport, issue *PlaneIssue) (bool, error) {
	claimCtx, cancel := context.WithTimeout(ctx, planePostSaveDBTimeout)
	defer cancel()
	claimed, err := s.repo.SetPlaneTicket(claimCtx, report.ClinicID, report.ID, issue.ID, issue.URL)
	if err != nil {
		slog.ErrorContext(ctx, "failed to persist plane ticket; plane-side issue may be orphaned",
			"error", err, "report_id", report.ID, "plane_issue_id", issue.ID)
		return false, err
	}
	if !claimed {
		return false, nil
	}
	report.PlaneIssueID = &issue.ID
	if issue.URL != "" {
		report.PlaneIssueURL = &issue.URL
	}
	report.PlaneSyncError = nil
	return true, nil
}

// logPlaneAutoAudit は自動起票の結果を固定 outcome で監査記録する（best-effort）。
// 記録するのは固定分類と安全なローカル/外部 ID のみ — 報告本文や upstream の
// エラー詳細は監査イベントにも出さない。書込み失敗は ERROR ログで可視化し、
// 報告作成の成功は変えない（DB-atomic ではない既知の制限）。
func (s *service) logPlaneAutoAudit(postSaveCtx context.Context, report *model.SupportBugReport, outcome, planeIssueID string) {
	if s.audit == nil {
		return
	}
	auditCtx, cancel := context.WithTimeout(postSaveCtx, planePostSaveAuditTimeout)
	defer cancel()
	newValue := map[string]any{"outcome": outcome}
	if planeIssueID != "" {
		newValue["plane_issue_id"] = planeIssueID
	}
	if err := s.audit.LogEntry(auditCtx, AuditEntry{
		ClinicID:   &report.ClinicID,
		ActorID:    &report.ReporterStaffID,
		ActorType:  model.AuditActorTypeStaff,
		Action:     planeAutoAuditAction,
		Resource:   "support_bug_report",
		ResourceID: &report.ID,
		NewValue:   newValue,
	}); err != nil {
		slog.ErrorContext(auditCtx, "failed to write plane auto-ticket audit log",
			"error", err, "report_id", report.ID, "clinic_id", report.ClinicID)
	}
}

// ListChatHistory は指定スタッフの会話履歴を古い順で返す。
func (s *service) ListChatHistory(ctx context.Context, clinicID, staffID uint64) ([]model.SupportChatMessage, error) {
	return s.repo.ListChatHistory(ctx, clinicID, staffID)
}

// RecordChatExchange は回答が成立した「質問+回答」のペアを1回の INSERT で保存する。
// LLM 呼び出しが失敗したやり取りは呼び出し側から渡されない前提。
func (s *service) RecordChatExchange(ctx context.Context, clinicID, staffID uint64, userMessage, assistantReply string, sources []ChatSource) error {
	var sourcesJSON json.RawMessage
	if len(sources) > 0 {
		b, err := json.Marshal(sources)
		if err != nil {
			return apperrors.Wrap(err, "failed to encode chat sources")
		}
		sourcesJSON = b
	}
	return s.repo.CreateChatMessages(ctx, []*model.SupportChatMessage{
		{ClinicID: clinicID, StaffID: staffID, Role: model.SupportChatRoleUser, Content: userMessage},
		{ClinicID: clinicID, StaffID: staffID, Role: model.SupportChatRoleAssistant, Content: assistantReply, Sources: sourcesJSON},
	})
}

// ChatExchange は一覧表示用の質問+回答ペア（回答行に結合したスタッフ/医院名つき）。
type ChatExchange struct {
	UserMessage      model.SupportChatMessage
	AssistantMessage model.SupportChatMessage
	StaffName        string
	ClinicName       string
}

// ListChatExchanges は全医院の履歴を user→assistant のペアにして新しい順で返す。
// RecordChatExchange が質問+回答を同一 batch で保存するため、同一 clinic×staff 内で
// assistant の直前にある user 行がその質問。対が欠けた行（窓の境界など）は除外する。
func (s *service) ListChatExchanges(ctx context.Context) ([]ChatExchange, error) {
	messages, err := s.repo.ListAllChatMessages(ctx)
	if err != nil {
		return nil, err
	}
	// 新しい順で来るので古い順に反転してからペア化する
	type exchangeKey struct{ clinicID, staffID uint64 }
	pending := make(map[exchangeKey]*ChatMessageWithMeta, len(messages))
	exchanges := make([]ChatExchange, 0, len(messages)/2)
	for i := len(messages) - 1; i >= 0; i-- {
		m := &messages[i]
		key := exchangeKey{clinicID: m.ClinicID, staffID: m.StaffID}
		switch m.Role {
		case model.SupportChatRoleUser:
			pending[key] = m
		case model.SupportChatRoleAssistant:
			user, ok := pending[key]
			if !ok {
				continue
			}
			delete(pending, key)
			exchanges = append(exchanges, ChatExchange{
				UserMessage:      user.SupportChatMessage,
				AssistantMessage: m.SupportChatMessage,
				StaffName:        m.StaffName,
				ClinicName:       m.ClinicName,
			})
		}
	}
	// 古い順で組み立てたので新しい順に反転して返す
	for i, j := 0, len(exchanges)-1; i < j; i, j = i+1, j-1 {
		exchanges[i], exchanges[j] = exchanges[j], exchanges[i]
	}
	return exchanges, nil
}

// chatRoomGap は会話区切りとして扱う無発言時間。同一 clinic×staff の発言が
// この間隔を超えて空いたら別ルームとみなす（会話区切りはDBスキーマに存在しないためのヒューリスティック）。
const chatRoomGap = 30 * time.Minute

// ChatRoomTurn はルーム内の1発言（user / assistant）。
type ChatRoomTurn struct {
	ID        uint64
	Role      model.SupportChatMessageRole
	Content   string
	Sources   json.RawMessage
	CreatedAt time.Time
}

// ChatRoom は同一 clinic×staff の時間的に連続した会話まとまり（擬似ルーム）。
type ChatRoom struct {
	// ID は導出キー "clinicID-staffID-先頭メッセージID"。DB に永続化されない。
	ID            string
	ClinicName    string
	StaffName     string
	StartedAt     time.Time
	LastMessageAt time.Time
	Turns         []ChatRoomTurn
}

// ListChatRooms は全医院のサポートチャット履歴を、同一 clinic×staff の発言ストリームを
// chatRoomGap を区切りにした「擬似会話ルーム」単位で返す。ルームは最終発言の新しい順、
// ルーム内の発言は古い順。閲覧範囲は ListChatExchanges と同じ共有ボード方針。
func (s *service) ListChatRooms(ctx context.Context) ([]ChatRoom, error) {
	messages, err := s.repo.ListAllChatMessages(ctx)
	if err != nil {
		return nil, err
	}
	type key struct{ clinicID, staffID uint64 }
	rooms := make([]ChatRoom, 0, len(messages)/2)
	open := map[key]int{}
	// ListAllChatMessages は新しい順 — 古い順に走査して発言をルームへ積む
	for i := len(messages) - 1; i >= 0; i-- {
		m := &messages[i]
		k := key{m.ClinicID, m.StaffID}
		idx, ok := open[k]
		if !ok || m.CreatedAt.Sub(rooms[idx].LastMessageAt) > chatRoomGap {
			idx = len(rooms)
			rooms = append(rooms, ChatRoom{
				ID:         fmt.Sprintf("%d-%d-%d", m.ClinicID, m.StaffID, m.ID),
				ClinicName: m.ClinicName,
				StaffName:  m.StaffName,
				StartedAt:  m.CreatedAt,
				Turns:      []ChatRoomTurn{},
			})
			open[k] = idx
		}
		rooms[idx].Turns = append(rooms[idx].Turns, ChatRoomTurn{
			ID:        m.ID,
			Role:      m.Role,
			Content:   m.Content,
			Sources:   m.Sources,
			CreatedAt: m.CreatedAt,
		})
		rooms[idx].LastMessageAt = m.CreatedAt
	}
	sort.Slice(rooms, func(i, j int) bool {
		return rooms[i].LastMessageAt.After(rooms[j].LastMessageAt)
	})
	return rooms, nil
}

// ClearChatHistory は指定スタッフの会話履歴をすべて soft delete する。
func (s *service) ClearChatHistory(ctx context.Context, clinicID, staffID uint64) error {
	return s.repo.ClearChatHistory(ctx, clinicID, staffID)
}

// planeStateSyncBatchLimit は 1 実行あたりの同期対象上限（防御的上限 — 通常は数件）。
// 上限超過分は次回 cron に繰り越される（id ASC で安定順のため取り残しはない）。
const planeStateSyncBatchLimit = 200

// planeStateCompletedGroup は Plane の「完了」状態グループ名（Done 相当）。
// cancelled は resolved に倒さない — 起票済みのまま open で残し、人が判断する。
const planeStateCompletedGroup = "completed"

// PlaneSyncResult は scheduled plane_sync ジョブの件数サマリ。
// scheduler.Result の契約に合わせ Processed == Succeeded + Failed を維持する
// （fetch/update の失敗だけが Failed — 未完了スキップは Succeeded に含める）。
type PlaneSyncResult struct {
	Processed int
	Succeeded int
	Failed    int
}

// SyncPlaneTicketStates は Plane 側で completed になった起票済み報告を resolved に揃える。
// Plane→ローカルの一方向のみ（ローカルの open/resolved 操作は外部へ反映しない）。
// 連携無効（tickets が TicketStateReader を実装しない = nil 含む）は zero result で no-op。
func (s *service) SyncPlaneTicketStates(ctx context.Context) PlaneSyncResult {
	reader, ok := s.tickets.(TicketStateReader)
	if !ok || reader == nil {
		return PlaneSyncResult{}
	}
	reports, err := s.repo.ListOpenWithPlaneTicket(ctx, planeStateSyncBatchLimit)
	if err != nil {
		slog.ErrorContext(ctx, "plane sync: failed to list open reports", "error", err)
		return PlaneSyncResult{Processed: 1, Failed: 1}
	}
	var result PlaneSyncResult
	for i := range reports {
		report := &reports[i]
		result.Processed++
		group, err := reader.FetchIssueStateGroup(ctx, *report.PlaneIssueID)
		if err != nil {
			result.Failed++
			// 404 含む upstream 失敗は報告行に記録して一覧から追跡できるようにする。
			// planeSyncErrorMessage は upstream body を落とした安全な文言のみ返す。
			if persistErr := s.repo.SetPlaneStateSyncError(ctx, report.ClinicID, report.ID, planeSyncErrorMessage(err)); persistErr != nil {
				slog.ErrorContext(ctx, "failed to persist plane state sync error", "error", persistErr, "report_id", report.ID)
			}
			slog.WarnContext(ctx, "plane sync: failed to fetch issue state",
				"error", err, "report_id", report.ID, "clinic_id", report.ClinicID)
			continue
		}
		completed := group == planeStateCompletedGroup
		if err := s.repo.ApplyPlaneStateSync(ctx, report.ClinicID, report.ID, completed); err != nil {
			result.Failed++
			slog.WarnContext(ctx, "plane sync: failed to persist fetched state",
				"error", err, "report_id", report.ID, "clinic_id", report.ClinicID)
			continue
		}
		result.Succeeded++
		if completed {
			slog.InfoContext(ctx, "plane sync: report resolved via plane completed",
				"report_id", report.ID, "clinic_id", report.ClinicID)
		}
	}
	return result
}
