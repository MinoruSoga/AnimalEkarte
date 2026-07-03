package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	lineinfra "github.com/animal-ekarte/backend/internal/infra/line"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

func purposeTagNameFromPrefixes(purpose string, t time.Time, prefixes []*model.LstepSendPurposeTagPrefix) string {
	date := t.Format("2006-01-02")
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
	ownerRepo     repository.OwnerRepository
	sharedFile    SharedFileService
	tagCacheRepo  repository.LstepTagCacheRepository
	auditSvc      AuditService
	logRepo       repository.LineSendLogRepository
	tagConfigRepo repository.LstepTagConfigRepository
	// newLineClient は LINE Messaging API クライアントのファクトリ。
	// 本番では lineinfra.NewMessagingClient（実際に外部 API へ通信する）で初期化される。
	// テスト時にのみ差し替え可能にするためのテスト容易性シーム（振る舞いは変更しない）。
	newLineClient func(channelAccessToken string) lineinfra.MessagingClient
}

func NewLineSendService(
	lstepSettings LstepSettingsService,
	ownerRepo repository.OwnerRepository,
	sharedFile SharedFileService,
	tagCacheRepo repository.LstepTagCacheRepository,
	auditSvc AuditService,
	logRepo repository.LineSendLogRepository,
	tagConfigRepo repository.LstepTagConfigRepository,
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

func (s *lineSendService) Send(ctx context.Context, clinicID uint64, input *SendLineMessageInput) (*SendLineMessageResult, error) {
	owner, err := s.ownerRepo.FindByID(ctx, clinicID, input.OwnerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find owner for line send", "error", err)
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
		slog.ErrorContext(ctx, "failed to get line credentials", "error", err)
		return nil, apperrors.Wrap(err, "failed to get LINE credentials")
	}

	lineClient := s.newLineClient(lineToken)

	var contentSummary string
	var sendErr error

	switch input.MessageType {
	case "text":
		contentSummary = lineSendTruncate(input.Text, 50)
		sendErr = lineClient.PushText(ctx, *owner.LineUserID, input.Text)
	case "pdf_url", "image_url":
		if input.FileID == nil {
			return nil, apperrors.WrapInvalidInput("file_id は必須です")
		}
		fileURL, fErr := s.sharedFile.GetSignedURL(ctx, clinicID, *input.FileID)
		if fErr != nil {
			slog.ErrorContext(ctx, "failed to get signed URL for line send", "error", fErr)
			return nil, apperrors.Wrap(fErr, "failed to get file URL")
		}
		if input.MessageType == "image_url" {
			contentSummary = lineSendTruncate(fileURL, 50)
			sendErr = lineClient.PushImageURL(ctx, *owner.LineUserID, fileURL, fileURL)
		} else {
			altText := input.FileName
			if altText == "" {
				altText = "ファイル"
			}
			contentSummary = altText
			sendErr = lineClient.PushFileURL(ctx, *owner.LineUserID, fileURL, altText)
		}
	default:
		return nil, apperrors.WrapInvalidInput("message_type が無効です")
	}

	sentAt := time.Now()
	logStatus := "sent"
	var errMsg *string
	if sendErr != nil {
		logStatus = "failed"
		msg := sendErr.Error()
		errMsg = &msg
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
		return nil, apperrors.WrapBadGateway(fmt.Sprintf("LINE送信に失敗しました: %s", sendErr.Error()))
	}

	result := &SendLineMessageResult{SentAt: sentAt}
	var purposePrefixes []*model.LstepSendPurposeTagPrefix
	if s.tagConfigRepo != nil {
		if pp, ppErr := s.tagConfigRepo.FindAllSendPurposeTagPrefixes(ctx); ppErr != nil {
			slog.ErrorContext(ctx, "failed to load send purpose tag prefixes", "error", ppErr)
		} else {
			purposePrefixes = pp
		}
	}
	tagName := purposeTagNameFromPrefixes(input.Purpose, sentAt, purposePrefixes)
	if tagName != "" {
		if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, input.OwnerID, tagName, "auto", ""); upsertErr != nil {
			slog.ErrorContext(ctx, "failed to upsert line send tag", "error", upsertErr, "tag", tagName)
		} else {
			result.TagAdded = tagName
		}
	}

	return result, nil
}

func (s *lineSendService) GetSendLogs(ctx context.Context, clinicID, ownerID uint64) ([]*model.LineSendLog, error) {
	logs, err := s.logRepo.FindByOwner(ctx, clinicID, ownerID, 30)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find line send logs", "error", err)
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
