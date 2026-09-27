package support

import (
	"context"

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

// Service はバグ報告のユースケースインターフェース
type Service interface {
	Create(ctx context.Context, clinicID, reporterStaffID uint64, input CreateBugReportInput) (*model.SupportBugReport, error)
	ListByClinic(ctx context.Context, clinicID uint64) ([]BugReportWithReporter, error)
	UpdateStatus(ctx context.Context, clinicID, id uint64, status model.SupportBugReportStatus) (*model.SupportBugReport, error)
}

type service struct {
	repo Repository
}

// NewService は Service を初期化して返す
func NewService(repo Repository) Service {
	return &service{repo: repo}
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
