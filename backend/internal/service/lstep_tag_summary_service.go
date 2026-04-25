package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/repository"
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
	OwnerID          uint64
	OwnerName        string
	LineUserIDMasked *string
	LastVisitDate    *string
	AllTags          []string
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
	tagCache repository.LstepTagCacheRepository
}

// NewLstepTagSummaryService は LstepTagSummaryService を初期化して返す。
func NewLstepTagSummaryService(tagCache repository.LstepTagCacheRepository) LstepTagSummaryService {
	return &lstepTagSummaryService{tagCache: tagCache}
}

func (s *lstepTagSummaryService) GetTagSummary(ctx context.Context, clinicID uint64) (TagSummaryResponse, error) {
	rows, total, err := s.tagCache.TagSummary(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get tag summary", "clinic_id", clinicID, "error", err)
		return TagSummaryResponse{}, apperrors.Wrap(err, "failed to get tag summary")
	}
	items := make([]TagSummaryItem, len(rows))
	for i, r := range rows {
		items[i] = TagSummaryItem{TagName: r.TagName, OwnerCount: r.OwnerCount, Category: r.Category}
	}
	return TagSummaryResponse{Tags: items, TotalOwnersWithLstep: total, AsOf: time.Now()}, nil
}

func (s *lstepTagSummaryService) ListOwnersByTag(ctx context.Context, clinicID uint64, input ListOwnersByTagInput) (TagOwnerListResponse, error) {
	perPage := input.PerPage
	if perPage <= 0 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	page := input.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * perPage

	rows, total, err := s.tagCache.FindOwnersByTag(ctx, clinicID, input.TagName, input.NameQuery, offset, perPage)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list owners by tag", "clinic_id", clinicID, "tag", input.TagName, "error", err)
		return TagOwnerListResponse{}, apperrors.Wrap(err, "failed to list owners by tag")
	}

	items := make([]TagOwnerItem, len(rows))
	for i, r := range rows {
		item := TagOwnerItem{
			OwnerID:       r.OwnerID,
			OwnerName:     r.OwnerName,
			AllTags:       r.Tags,
			LastVisitDate: extractLastVisitDate(r.Tags),
		}
		if r.LineUserID != nil {
			masked := maskLineUserID(*r.LineUserID)
			item.LineUserIDMasked = &masked
		}
		items[i] = item
	}
	return TagOwnerListResponse{Owners: items, Total: total, Page: page, PerPage: perPage}, nil
}

func (s *lstepTagSummaryService) ExportOwnersByTagCSV(ctx context.Context, clinicID uint64, tagName, nameQuery string, w io.Writer) error {
	// ページネーションなし全件取得（最大5000件で安全上限）
	rows, _, err := s.tagCache.FindOwnersByTag(ctx, clinicID, tagName, nameQuery, 0, 5000)
	if err != nil {
		slog.ErrorContext(ctx, "failed to export owners by tag csv", "clinic_id", clinicID, "tag", tagName, "error", err)
		return apperrors.Wrap(err, "failed to export owners by tag csv")
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
		cw.Write([]string{ //nolint:errcheck
			fmt.Sprintf("%d", r.OwnerID),
			r.OwnerName,
			lvd,
			extractCPMStage(r.Tags),
		})
	}
	cw.Flush()
	return cw.Error()
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
