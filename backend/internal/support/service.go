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

// Service はバグ報告とヘルプチャット履歴のユースケースインターフェース
type Service interface {
	Create(ctx context.Context, clinicID, reporterStaffID uint64, input CreateBugReportInput) (*model.SupportBugReport, error)
	ListByClinic(ctx context.Context, clinicID uint64) ([]BugReportWithReporter, error)
	UpdateStatus(ctx context.Context, clinicID, id uint64, status model.SupportBugReportStatus) (*model.SupportBugReport, error)
	EnsurePlaneTicket(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error)
	Delete(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error)
	ListChatHistory(ctx context.Context, clinicID, staffID uint64) ([]model.SupportChatMessage, error)
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
	// 報告の保存が先・Plane 起票は後続の best-effort 副作用。
	// Plane 障害で報告を失わないことが主契約のため、ここは失敗しても 201 を返す。
	if s.tickets != nil {
		s.syncPlaneTicket(ctx, &report)
	}
	return &report, nil
}

func (s *service) ListByClinic(ctx context.Context, clinicID uint64) ([]BugReportWithReporter, error) {
	return s.repo.FindByClinicID(ctx, clinicID)
}

func (s *service) UpdateStatus(ctx context.Context, clinicID, id uint64, status model.SupportBugReportStatus) (*model.SupportBugReport, error) {
	if !model.IsValidSupportBugReportStatus(string(status)) {
		return nil, apperrors.WrapInvalidInput("invalid status")
	}
	if err := s.repo.UpdateStatus(ctx, clinicID, id, status); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, clinicID, id)
}

// EnsurePlaneTicket は管理画面からの手動起票・再送。起票済みなら現行値をそのまま返す（冪等）。
// Plane 未設定なら 501、起票失敗は同期状態を記録した上で 502 を返す。
func (s *service) EnsurePlaneTicket(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error) {
	report, err := s.repo.FindByID(ctx, clinicID, id)
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
	return s.repo.FindByID(ctx, clinicID, id)
}

// Delete は報告を論理削除する。返り値は削除前に読み取った行で、
// handler がスクショの後始末と監査記録に使う。
func (s *service) Delete(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error) {
	report, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SoftDeleteBugReport(ctx, clinicID, id); err != nil {
		return nil, err
	}
	return report, nil
}

// syncPlaneTicket は Plane 起票を best-effort で実行する。
// 失敗しても報告は保存済みなので、同期状態を DB に記録して WARN ログを残すのみ。
func (s *service) syncPlaneTicket(ctx context.Context, report *model.SupportBugReport) {
	issue, err := s.tickets.CreateBugReportIssue(ctx, report)
	if err != nil {
		s.recordPlaneFailure(ctx, report, err)
		return
	}
	s.claimPlaneTicket(ctx, report, issue)
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

// ClearChatHistory は指定スタッフの会話履歴をすべて soft delete する。
func (s *service) ClearChatHistory(ctx context.Context, clinicID, staffID uint64) error {
	return s.repo.ClearChatHistory(ctx, clinicID, staffID)
}
