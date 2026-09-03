package lstep

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	lineinfra "github.com/animal-ekarte/backend/internal/infra/line"
	"github.com/animal-ekarte/backend/internal/model"
)

func purposeTagNameFromPrefixes(purpose string, t time.Time, prefixes []*model.LstepSendPurposeTagPrefix) string {
	date := t.Format(time.DateOnly)
	for _, p := range prefixes {
		if p.Purpose == purpose {
			return p.TagPrefix + date
		}
	}
	return ""
}

type SendLineMessageInput struct {
	OwnerID     uint64
	StaffID     uint64
	MessageType string
	Text        string
	FileID      *uint64
	FileName    string
	Purpose     string
}

type SendLineMessageResult struct {
	SentAt   time.Time
	TagAdded string
}

type LineSendService interface {
	Send(ctx context.Context, clinicID uint64, input *SendLineMessageInput) (*SendLineMessageResult, error)
	GetSendLogs(ctx context.Context, clinicID, ownerID uint64) ([]*model.LineSendLog, error)
}

type lineSendService struct {
	lstepSettings LstepSettingsService
	ownerRepo     lstepOwnerRepo
	sharedFile    lstepSharedFileService
	tagCacheRepo  LstepTagCacheRepository
	auditSvc      lstepAuditLogger
	logRepo       LineSendLogRepository
	tagConfigRepo LstepTagConfigRepository
	// newLineClient は LINE Messaging API クライアントのファクトリ。
	// 本番では lineinfra.NewMessagingClient（実際に外部 API へ通信する）で初期化される。
	// テスト時にのみ差し替え可能にするためのテスト容易性シーム（振る舞いは変更しない）。
	newLineClient func(channelAccessToken string) lineinfra.MessagingClient
}

func NewLineSendService(
	lstepSettings LstepSettingsService,
	ownerRepo lstepOwnerRepo,
	sharedFile lstepSharedFileService,
	tagCacheRepo LstepTagCacheRepository,
	auditSvc lstepAuditLogger,
	logRepo LineSendLogRepository,
	tagConfigRepo LstepTagConfigRepository,
) LineSendService {
	return &lineSendService{
		lstepSettings: lstepSettings,
		ownerRepo:     ownerRepo,
		sharedFile:    sharedFile,
		tagCacheRepo:  tagCacheRepo,
		auditSvc:      auditSvc,
		logRepo:       logRepo,
		tagConfigRepo: tagConfigRepo,
		newLineClient: lineinfra.NewMessagingClient,
	}
}

// dispatchLineMessage は input.MessageType に応じてLINEメッセージを送信する
// （BE-refactor.md E-8）。pdf/image 分岐の検証エラー（file_id 必須・署名URL取得失敗・
// message_type 不正）は fail-fast の別経路のため、通常の送信エラー（sendErr）とは
// 戻り値を分ける（現行の分岐構造を厳密に保つ）。
func (s *lineSendService) dispatchLineMessage(ctx context.Context, lineClient lineinfra.MessagingClient, clinicID uint64, lineUserID string, input *SendLineMessageInput) (contentSummary string, validationErr, sendErr error) {
	switch input.MessageType {
	case "text":
		contentSummary = lineSendTruncate(input.Text, 50)
		sendErr = lineClient.PushText(ctx, lineUserID, input.Text)
	case "pdf_url", "image_url":
		if input.FileID == nil {
			validationErr = apperrors.WrapInvalidInput("file_id は必須です")
			return
		}
		fileURL, fErr := s.sharedFile.GetSignedURL(ctx, clinicID, *input.FileID)
		if fErr != nil {
			validationErr = apperrors.Wrap(fErr, "failed to get file URL")
			return
		}
		if input.MessageType == "image_url" {
			contentSummary = lineSendTruncate(fileURL, 50)
			sendErr = lineClient.PushImageURL(ctx, lineUserID, fileURL, fileURL)
		} else {
			altText := input.FileName
			if altText == "" {
				altText = "ファイル"
			}
			contentSummary = altText
			sendErr = lineClient.PushFileURL(ctx, lineUserID, fileURL, altText)
		}
	default:
		validationErr = apperrors.WrapInvalidInput("message_type が無効です")
	}
	return
}

// applySendPurposeTag は purpose に対応する自動タグを best-effort で付与する
// （BE-refactor.md E-8）。付与したタグ名を返す（""=付与なし）。
func (s *lineSendService) applySendPurposeTag(ctx context.Context, clinicID, ownerID uint64, purpose string, sentAt time.Time) string {
	var purposePrefixes []*model.LstepSendPurposeTagPrefix
	if s.tagConfigRepo != nil {
		if pp, ppErr := s.tagConfigRepo.FindAllSendPurposeTagPrefixes(ctx); ppErr != nil {
			slog.ErrorContext(ctx, "failed to load send purpose tag prefixes", "error", ppErr)
		} else {
			purposePrefixes = pp
		}
	}
	tagName := purposeTagNameFromPrefixes(purpose, sentAt, purposePrefixes)
	if tagName == "" {
		return ""
	}
	if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, tagName, "auto", ""); upsertErr != nil {
		slog.ErrorContext(ctx, "failed to upsert line send tag", "error", upsertErr, "tag", tagName)
		return ""
	}
	return tagName
}

func (s *lineSendService) Send(ctx context.Context, clinicID uint64, input *SendLineMessageInput) (*SendLineMessageResult, error) {
	owner, err := s.ownerRepo.FindByID(ctx, clinicID, input.OwnerID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find owner")
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil, apperrors.WrapInvalidInput("飼い主にLINE User IDが設定されていません")
	}
	if owner.LstepOptOut {
		return nil, apperrors.WrapInvalidInput("この飼い主はLINEメッセージの受信を拒否しています")
	}

	_, _, lineToken, err := s.lstepSettings.GetRawCredentials(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get LINE credentials")
	}

	lineClient := s.newLineClient(lineToken)

	contentSummary, validationErr, sendErr := s.dispatchLineMessage(ctx, lineClient, clinicID, *owner.LineUserID, input)
	if validationErr != nil {
		return nil, validationErr
	}

	sentAt := time.Now()
	logStatus := "sent"
	var errMsg *string
	if sendErr != nil {
		logStatus = "failed"
		// LSA-09: persist only a stable classification code; raw provider errors stay in slog.
		code := classifyLineSendError(sendErr)
		errMsg = &code
		slog.ErrorContext(ctx, "LINE messaging API send failed",
			"error", sendErr,
			"clinic_id", clinicID,
			"owner_id", input.OwnerID,
			"error_class", code,
		)
	}

	logRecord := &model.LineSendLog{
		ClinicID:       clinicID,
		OwnerID:        input.OwnerID,
		SentByUserID:   input.StaffID,
		MessageType:    input.MessageType,
		ContentSummary: contentSummary,
		Status:         logStatus,
		ErrorMessage:   errMsg,
		SentAt:         sentAt,
	}
	if logErr := s.logRepo.Create(ctx, logRecord); logErr != nil {
		slog.ErrorContext(ctx, "failed to save line send log", "error", logErr)
	}

	// best-effort audit log
	staffID := input.StaffID
	ownerID := input.OwnerID
	if err := s.auditSvc.LogLstepOperation(ctx, clinicID, &staffID, "owner.line.send", "owner", &ownerID); err != nil {
		slog.WarnContext(ctx, "audit log failed for line send", "error", err, "owner_id", ownerID)
	}

	if sendErr != nil {
		// LSA-09: fixed client-facing message; do not embed raw upstream Error().
		return nil, apperrors.WrapBadGateway("LINE送信に失敗しました")
	}

	result := &SendLineMessageResult{SentAt: sentAt}
	result.TagAdded = s.applySendPurposeTag(ctx, clinicID, input.OwnerID, input.Purpose, sentAt)

	return result, nil
}

func (s *lineSendService) GetSendLogs(ctx context.Context, clinicID, ownerID uint64) ([]*model.LineSendLog, error) {
	logs, err := s.logRepo.FindByOwner(ctx, clinicID, ownerID, 30)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find line send logs")
	}
	return logs, nil
}

func lineSendTruncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}

// classifyLineSendError maps provider/transport failures to stable codes for API/DB (LSA-09).
// Raw details must not leave slog.
func classifyLineSendError(err error) string {
	if err == nil {
		return "line_api_error"
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return "timeout"
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "timeout"
	}
	// Fallback only: remaining Contains must not proliferate. Prefer typed errors.Is/As above.
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline"):
		return "timeout"
	case strings.Contains(msg, "unauthorized") || strings.Contains(msg, "401") || strings.Contains(msg, "403"):
		return "unauthorized"
	case strings.Contains(msg, "connection") || strings.Contains(msg, "network") || strings.Contains(msg, "dial "):
		return "unreachable"
	default:
		return "line_api_error"
	}
}
