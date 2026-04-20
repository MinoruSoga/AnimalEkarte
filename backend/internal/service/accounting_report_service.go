package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/repository"
)

// AccountingReportService は月次売上レポートのビジネスロジックインターフェース
type AccountingReportService interface {
	GetMonthly(ctx context.Context, clinicID uint64, year, month int) (*repository.MonthlyReportResult, error)
	ExportCSV(ctx context.Context, clinicID uint64, year, month int) (io.Reader, error)
}

type accountingReportService struct {
	repo repository.AccountingRepository
}

// NewAccountingReportService は AccountingReportService を初期化して返す
func NewAccountingReportService(repo repository.AccountingRepository) AccountingReportService {
	return &accountingReportService{repo: repo}
}

func (s *accountingReportService) GetMonthly(ctx context.Context, clinicID uint64, year, month int) (*repository.MonthlyReportResult, error) {
	if month < 1 || month > 12 {
		return nil, apperrors.WrapInvalidInput("month は 1〜12 の範囲で指定してください")
	}
	result, err := s.repo.GetMonthlyReport(ctx, clinicID, year, month)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get monthly report")
	}
	return result, nil
}

func (s *accountingReportService) ExportCSV(ctx context.Context, clinicID uint64, year, month int) (io.Reader, error) {
	if month < 1 || month > 12 {
		return nil, apperrors.WrapInvalidInput("month は 1〜12 の範囲で指定してください")
	}
	result, err := s.repo.GetMonthlyReport(ctx, clinicID, year, month)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get monthly report for CSV export")
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// BOM（Excel の UTF-8 認識用）
	buf.WriteString("\xEF\xBB\xBF")

	// ヘッダー行
	header := []string{"日付", "区分(AM/PM)", "カテゴリ", "支払方法ID", "純売上(円)"}
	if err := w.Write(header); err != nil {
		return nil, apperrors.Wrap(err, "failed to write CSV header")
	}

	for _, row := range result.Rows {
		pmID := ""
		if row.PaymentMethodID != nil {
			pmID = fmt.Sprintf("%d", *row.PaymentMethodID)
		}
		record := []string{
			row.Date,
			row.Period,
			row.Category,
			pmID,
			fmt.Sprintf("%d", row.NetAmount),
		}
		if err := w.Write(record); err != nil {
			return nil, apperrors.Wrap(err, "failed to write CSV row")
		}
	}

	// 合計行
	if err := w.Write([]string{"合計", "", "", "", fmt.Sprintf("%d", result.GrandTotal)}); err != nil {
		return nil, apperrors.Wrap(err, "failed to write CSV total row")
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, apperrors.Wrap(err, "failed to flush CSV writer")
	}

	return &buf, nil
}
