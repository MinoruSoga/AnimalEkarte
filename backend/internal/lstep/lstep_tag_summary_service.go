package lstep

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// TagSummaryItem はタグ集計の1エントリ。
type TagSummaryItem struct {
	TagName    string
	OwnerCount int64
	Category   string
}

// TagSummaryResponse は GET /lstep/tag-summary のレスポンス。
type TagSummaryResponse struct {
	Tags                 []TagSummaryItem
	TotalOwnersWithLstep int64
	AsOf                 time.Time
}

// ListOwnersByTagInput は ListOwnersByTag の入力パラメータ。
type ListOwnersByTagInput struct {
	TagName   string
	NameQuery string
	Page      int
	PerPage   int
}

// TagOwnerItem は GET /lstep/owners の1件。
type TagOwnerItem struct {
	OwnerID       uint64
	OwnerName     string
	LastVisitDate *string
	AllTags       []string
	Reason        *string
}

// TagOwnerListResponse は GET /lstep/owners のレスポンス。
type TagOwnerListResponse struct {
	Owners  []TagOwnerItem
	Total   int64
	Page    int
	PerPage int
}

// LstepTagSummaryService はタグ集計・タグ別飼い主検索のサービスインターフェース。
type LstepTagSummaryService interface {
	GetTagSummary(ctx context.Context, clinicID uint64) (TagSummaryResponse, error)
	ListOwnersByTag(ctx context.Context, clinicID uint64, input ListOwnersByTagInput) (TagOwnerListResponse, error)
	ExportOwnersByTagCSV(ctx context.Context, clinicID uint64, tagName, nameQuery string, w io.Writer) error
}

type lstepTagSummaryService struct {
	tagCache tagSummaryRepo
}

// NewLstepTagSummaryService は LstepTagSummaryService を初期化して返す。
func NewLstepTagSummaryService(tagCache tagSummaryRepo) LstepTagSummaryService {
	return &lstepTagSummaryService{tagCache: tagCache}
}

func normalizePagination(page, perPage, defaultPerPage, maxPerPage int) (outPage, outPerPage, outOffset int) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}
	return page, perPage, (page - 1) * perPage
}

// sanitizeCSVCell はセルがスプレッドシート数式と解釈されるのを防ぐ。
// 先頭文字が = + - @ の場合、単一引用符 ' で前置する。
func sanitizeCSVCell(cell string) string {
	if cell != "" {
		switch cell[0] {
		case '=', '+', '-', '@':
			return "'" + cell
		}
	}
	return cell
}

func (s *lstepTagSummaryService) GetTagSummary(ctx context.Context, clinicID uint64) (TagSummaryResponse, error) {
	rows, total, err := s.tagCache.TagSummary(ctx, clinicID)
	if err != nil {
		return TagSummaryResponse{}, apperrors.Wrap(err, "failed to get tag summary")
	}
	items := make([]TagSummaryItem, len(rows))
	for i, r := range rows {
		items[i] = TagSummaryItem{TagName: r.TagName, OwnerCount: r.OwnerCount, Category: r.Category}
	}
	return TagSummaryResponse{Tags: items, TotalOwnersWithLstep: total, AsOf: time.Now()}, nil
}

func (s *lstepTagSummaryService) ListOwnersByTag(ctx context.Context, clinicID uint64, input ListOwnersByTagInput) (TagOwnerListResponse, error) {
	page, perPage, offset := normalizePagination(input.Page, input.PerPage, 20, 100)

	rows, total, err := s.tagCache.FindOwnersByTag(ctx, clinicID, input.TagName, input.NameQuery, offset, perPage)
	if err != nil {
		return TagOwnerListResponse{}, apperrors.Wrap(err, "failed to list owners by tag")
	}

	items := make([]TagOwnerItem, len(rows))
	for i, r := range rows {
		item := TagOwnerItem{
			OwnerID:       r.OwnerID,
			OwnerName:     r.OwnerName,
			AllTags:       r.Tags,
			LastVisitDate: extractLastVisitDate(r.Tags),
			Reason:        r.Reason,
		}
		items[i] = item
	}
	return TagOwnerListResponse{Owners: items, Total: total, Page: page, PerPage: perPage}, nil
}

// exportOwnersByTagCSVMaxRows is the hard cap for tag-owner CSV export.
// Exceeding this limit is fail-closed (Conflict) rather than silent truncate.
const exportOwnersByTagCSVMaxRows = 5000

func (s *lstepTagSummaryService) ExportOwnersByTagCSV(ctx context.Context, clinicID uint64, tagName, nameQuery string, w io.Writer) error {
	// Fetch one page at the hard cap and require total to fit — never drop rows silently.
	rows, total, err := s.tagCache.FindOwnersByTag(ctx, clinicID, tagName, nameQuery, 0, exportOwnersByTagCSVMaxRows)
	if err != nil {
		return apperrors.Wrap(err, "failed to export owners by tag csv")
	}
	if total > int64(exportOwnersByTagCSVMaxRows) {
		return apperrors.WrapConflict(fmt.Sprintf(
			"tag owner export exceeds the %d-row limit (total=%d); narrow the tag or name filter",
			exportOwnersByTagCSVMaxRows, total,
		))
	}

	// UTF-8 BOM（Excel が Shift-JIS と誤認して日本語が文字化けするのを防ぐ）。
	// 月次集計レポート CSV と同一方針（#179 ③）。
	if _, err := w.Write([]byte("\xEF\xBB\xBF")); err != nil {
		return apperrors.Wrap(err, "failed to write csv BOM")
	}

	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"owner_id", "owner_name", "last_visit_date", "cpm_stage"}); err != nil {
		return apperrors.Wrap(err, "failed to write csv header")
	}
	for _, r := range rows {
		lvd := ""
		if d := extractLastVisitDate(r.Tags); d != nil {
			lvd = *d
		}
		cw.Write([]string{ //nolint:errcheck // csv.Writer error is captured by cw.Error() after Flush()
			fmt.Sprintf("%d", r.OwnerID),
			sanitizeCSVCell(r.OwnerName),
			sanitizeCSVCell(lvd),
			sanitizeCSVCell(extractCPMStage(r.Tags)),
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		return apperrors.Wrap(err, "csv writer error")
	}
	return nil
}

// extractLastVisitDate はタグ一覧から last_visit_YYYY-MM-DD を抽出する。
func extractLastVisitDate(tags []string) *string {
	const prefix = "last_visit_"
	for _, t := range tags {
		if strings.HasPrefix(t, prefix) {
			date := t[len(prefix):]
			if len(date) == 10 { // YYYY-MM-DD
				return &date
			}
		}
	}
	return nil
}

// extractCPMStage はタグ一覧から cpm_* タグを抽出して返す。
func extractCPMStage(tags []string) string {
	for _, t := range tags {
		if strings.HasPrefix(t, "cpm_") {
			return t[4:] // "cpm_" を除いた stage 名
		}
	}
	return ""
}
