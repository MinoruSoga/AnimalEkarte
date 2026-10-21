package manualarticle

import (
	"encoding/hex"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/operation"
)

// ManualArticleResponse はマニュアル記事の HTTP レスポンス（既存 wire 維持）
type ManualArticleResponse struct {
	ID               uint64    `json:"id"`
	Category         string    `json:"category"`
	Slug             string    `json:"slug"`
	Title            string    `json:"title"`
	OrderValue       float64   `json:"order_value"`
	Section          string    `json:"section"`
	BodyMarkdown     string    `json:"body_markdown"`
	UpdatedByStaffID *uint64   `json:"updated_by_staff_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ManualArticleListResponse はマニュアル記事一覧の HTTP レスポンス
type ManualArticleListResponse struct {
	Data []ManualArticleResponse `json:"data"`
}

// ManualArticleVersionListResponse は編集履歴一覧の HTTP レスポンス（既存 wire 維持）
type ManualArticleVersionListResponse struct {
	Data []ManualArticleVersionResponse `json:"data"`
}

// ManualArticleVersionResponse は編集履歴のレスポンス（既存 wire 維持）
type ManualArticleVersionResponse struct {
	ID              uint64    `json:"id"`
	ArticleID       uint64    `json:"article_id"`
	Title           string    `json:"title"`
	OrderValue      float64   `json:"order_value"`
	Section         string    `json:"section"`
	BodyMarkdown    string    `json:"body_markdown"`
	EditedByStaffID *uint64   `json:"edited_by_staff_id,omitempty"`
	EditedAt        time.Time `json:"edited_at"`
}

// ManualEditContentDTO は ManualArticleContent（表示/履歴内容の値 object）。
type ManualEditContentDTO struct {
	Title        string  `json:"title"`
	OrderValue   float64 `json:"order_value"`
	Section      string  `json:"section"`
	BodyMarkdown string  `json:"body_markdown"`
}

// ManualArticleVersionDTO は api.yaml ManualArticleVersion（応答専用の不変内容版）。
// ID は正数十進文字列、content_digest は 32byte digest の 64 lower-hex。
// source_kind の旧 NULL は legacy_unknown として返す。digest/revision の NULL は
// 未解決出所を表し、架空値で埋めない。
type ManualArticleVersionDTO struct {
	ID              string               `json:"id"`
	ArticleID       string               `json:"article_id"`
	ContentRevision *int64               `json:"content_revision"`
	ContentDigest   *string              `json:"content_digest"`
	SourceKind      string               `json:"source_kind"`
	RestoredFromID  *string              `json:"restored_from_id"`
	Content         ManualEditContentDTO `json:"content"`
	EditedByStaffID *uint64              `json:"edited_by_staff_id"`
	EditedAt        time.Time            `json:"edited_at"`
}

// ManualArticleEditViewResponse は api.yaml ManualArticleEditView
// （編集権限者限定。version は記事 CAS 版、current は現在内容版、
// 未公開 current は公開版 read へ混ぜない）。
type ManualArticleEditViewResponse struct {
	ID                 string                   `json:"id"`
	Category           string                   `json:"category"`
	Slug               string                   `json:"slug"`
	Version            int64                    `json:"version"`
	Current            *ManualArticleVersionDTO `json:"current"`
	PublishedVersionID *string                  `json:"published_version_id"`
	Hidden             bool                     `json:"hidden"`
}

// ManualEditReceiptResponse は api.yaml ManualEditReceipt
// （業務変更と原子的に確定した receipt）。
type ManualEditReceiptResponse struct {
	OperationKey       string  `json:"operation_key"`
	ArticleID          string  `json:"article_id"`
	Intent             string  `json:"intent"`
	PreviousVersion    int64   `json:"previous_version"`
	ResultingVersion   int64   `json:"resulting_version"`
	ContentVersionID   string  `json:"content_version_id"`
	ContentDigest      string  `json:"content_digest"`
	PublishedVersionID *string `json:"published_version_id"`
}

// ManualEditOperationResultResponse は api.yaml ManualEditOperationResult
// （succeeded は receipt 非NULL、unknown は NULL）。
type ManualEditOperationResultResponse struct {
	OperationKey string                     `json:"operation_key"`
	State        string                     `json:"state"`
	Receipt      *ManualEditReceiptResponse `json:"receipt"`
}

// ManualEditConflictBody は api.yaml ManualEditConflict（{error, code}）。
type ManualEditConflictBody struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func exactNumericFloat(v model.ExactNumeric) float64 {
	f8, err := v.Float64Value()
	if err != nil {
		return 0
	}
	return f8.Float64
}

func toManualArticleListResponse(articles []PublishedArticle) ManualArticleListResponse {
	data := httpapi.MapSlice(articles, toManualArticleResponse)
	return ManualArticleListResponse{Data: data}
}

func toManualArticleVersionListResponse(versions []model.ManualArticleVersion) ManualArticleVersionListResponse {
	data := httpapi.MapSlice(versions, toManualArticleVersionResponse)
	return ManualArticleVersionListResponse{Data: data}
}

func toManualArticleResponse(a *PublishedArticle) ManualArticleResponse {
	return ManualArticleResponse{
		ID:               a.ID,
		Category:         string(a.Category),
		Slug:             a.Slug,
		Title:            a.Title,
		OrderValue:       exactNumericFloat(a.OrderValue),
		Section:          a.Section,
		BodyMarkdown:     a.BodyMarkdown,
		UpdatedByStaffID: a.EditedByStaffID,
		CreatedAt:        httpapi.LocalTime(a.CreatedAt),
		UpdatedAt:        httpapi.LocalTime(a.UpdatedAt),
	}
}

func toManualArticleVersionResponse(v *model.ManualArticleVersion) ManualArticleVersionResponse {
	return ManualArticleVersionResponse{
		ID:              v.ID,
		ArticleID:       v.ArticleID,
		Title:           v.Title,
		OrderValue:      exactNumericFloat(v.OrderValue),
		Section:         v.Section,
		BodyMarkdown:    v.BodyMarkdown,
		EditedByStaffID: v.EditedByStaffID,
		EditedAt:        httpapi.LocalTime(v.EditedAt),
	}
}

func decimalStringPtr(id *uint64) *string {
	if id == nil {
		return nil
	}
	s := strconv.FormatUint(*id, 10)
	return &s
}

// toManualArticleVersionDTO は不変内容版の api.yaml 応答 projection。
// source_kind NULL → legacy_unknown、digest NULL → null（架空値で埋めない）。
func toManualArticleVersionDTO(v *model.ManualArticleVersion) ManualArticleVersionDTO {
	var digest *string
	if v.ContentDigest != nil {
		h := hex.EncodeToString(*v.ContentDigest)
		digest = &h
	}
	sourceKind := "legacy_unknown"
	if v.SourceKind != nil {
		sourceKind = *v.SourceKind
	}
	return ManualArticleVersionDTO{
		ID:              strconv.FormatUint(v.ID, 10),
		ArticleID:       strconv.FormatUint(v.ArticleID, 10),
		ContentRevision: v.ContentRevision,
		ContentDigest:   digest,
		SourceKind:      sourceKind,
		RestoredFromID:  decimalStringPtr(v.RestoredFromID),
		Content: ManualEditContentDTO{
			Title:        v.Title,
			OrderValue:   exactNumericFloat(v.OrderValue),
			Section:      v.Section,
			BodyMarkdown: v.BodyMarkdown,
		},
		EditedByStaffID: v.EditedByStaffID,
		EditedAt:        httpapi.LocalTime(v.EditedAt),
	}
}

func toManualArticleEditViewResponse(view *EditView) ManualArticleEditViewResponse {
	res := ManualArticleEditViewResponse{
		ID:                 strconv.FormatUint(view.Article.ID, 10),
		Category:           string(view.Article.Category),
		Slug:               view.Article.Slug,
		Version:            view.Article.Version,
		PublishedVersionID: decimalStringPtr(view.Article.PublishedVersionID),
		Hidden:             view.Article.HiddenAt != nil,
	}
	if view.Current != nil {
		current := toManualArticleVersionDTO(view.Current)
		res.Current = &current
	}
	return res
}

// receiptFromStoredResult は永続化済み operation_result の refs/versions から
// 公開 receipt を組み立てる。後日の最新状態で書き換えない。
// 必須 member が欠ける保存済み receipt は (nil, false)（非開示 → 500）。
func receiptFromStoredResult(key uuid.UUID, res *operation.Result) (*ManualEditReceiptResponse, bool) {
	refs := res.ResultRefs
	versions := res.CommittedVersions
	previous, err1 := strconv.ParseInt(versions["previous_version"], 10, 64)
	resulting, err2 := strconv.ParseInt(versions["resulting_version"], 10, 64)
	if err1 != nil || err2 != nil ||
		refs["article_id"] == "" || refs["content_version_id"] == "" || refs["content_digest"] == "" {
		return nil, false
	}
	receipt := &ManualEditReceiptResponse{
		OperationKey:     key.String(),
		ArticleID:        refs["article_id"],
		Intent:           publicReceiptIntentEdit,
		PreviousVersion:  previous,
		ResultingVersion: resulting,
		ContentVersionID: refs["content_version_id"],
		ContentDigest:    refs["content_digest"],
	}
	if v, ok := refs["published_version_id"]; ok && v != "" {
		receipt.PublishedVersionID = &v
	}
	return receipt, true
}
