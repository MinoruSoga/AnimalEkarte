package support

import (
	"context"
	"encoding/json"
	"log/slog"

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
	UpdateStatus(ctx context.Context, clinicID, id uint64, status model.SupportBugReportStatus) (*model.SupportBugReport, error)
	EnsurePlaneTicket(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error)
	Delete(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error)
	ListChatHistory(ctx context.Context, clinicID, staffID uint64) ([]model.SupportChatMessage, error)
	// ListChatExchanges は全医院の質問+回答ペアを新しい順で返す（共有一覧ページ用）
	ListChatExchanges(ctx context.Context) ([]ChatExchange, error)
	RecordChatExchange(ctx context.Context, clinicID, staffID uint64, userMessage, assistantReply string, sources []ChatSource) error
	ClearChatHistory(ctx context.Context, clinicID, staffID uint64) error
}

type service struct {
	repo    Repository
	tickets TicketCreator // nil = Plane 連携無効（ローカル保存のみ）
}

// NewService は Service を初期化して返す
func NewService(repo Repository, tickets TicketCreator) Service {
	return &service{repo: repo, tickets: tickets}
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
	if err := s.repo.Create(ctx, &report); err != nil {
		return nil, err
	}
	// Plane への外部エクスポートは作成時に自動実行しない。スタッフが内容を
	// 確認した上で明示操作（EnsurePlaneTicket = POST /:id/plane-ticket）する
	// reviewable export のみとし、未分類データが外部へ黙って出る経路を閉じる。
	return &report, nil
}

func (s *service) ListAll(ctx context.Context) ([]BugReportWithReporter, error) {
	return s.repo.FindAll(ctx)
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
		s.recordPlaneFailure(ctx, report, err)
		return nil, apperrors.WrapBadGateway("plane ticket creation failed")
	}
	if s.claimPlaneTicket(ctx, report, issue) {
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
// 記録自体の失敗（DB 障害）は追加の ERROR ログのみ（起票失敗の報告を悪化させない）。
func (s *service) recordPlaneFailure(ctx context.Context, report *model.SupportBugReport, syncErr error) {
	msg := planeSyncErrorMessage(syncErr)
	if err := s.repo.SetPlaneSyncError(ctx, report.ClinicID, report.ID, msg); err != nil {
		slog.ErrorContext(ctx, "failed to persist plane sync error", "error", err, "report_id", report.ID)
	}
	report.PlaneSyncError = &msg
	slog.WarnContext(ctx, "plane ticket creation failed", "error", syncErr, "report_id", report.ID, "clinic_id", report.ClinicID)
}

// claimPlaneTicket は起票成功を DB に記録する。claim できた場合のみ true。
// false の場合、別経路が先に記録済み（二重起票の防止）か DB 書き込み失敗。
// 後者は Plane 側に孤立チケットが残りうるため ERROR ログを残す。
func (s *service) claimPlaneTicket(ctx context.Context, report *model.SupportBugReport, issue *PlaneIssue) bool {
	claimed, err := s.repo.SetPlaneTicket(ctx, report.ClinicID, report.ID, issue.ID, issue.URL)
	if err != nil {
		slog.ErrorContext(ctx, "failed to persist plane ticket; plane-side issue may be orphaned",
			"error", err, "report_id", report.ID, "plane_issue_id", issue.ID)
		return false
	}
	if !claimed {
		return false
	}
	report.PlaneIssueID = &issue.ID
	if issue.URL != "" {
		report.PlaneIssueURL = &issue.URL
	}
	report.PlaneSyncError = nil
	return true
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

// ClearChatHistory は指定スタッフの会話履歴をすべて soft delete する。
func (s *service) ClearChatHistory(ctx context.Context, clinicID, staffID uint64) error {
	return s.repo.ClearChatHistory(ctx, clinicID, staffID)
}
