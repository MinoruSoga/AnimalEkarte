package manualarticle

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/operation"
)

func numericForTest(t *testing.T, text string) model.ExactNumeric {
	t.Helper()
	var n pgtype.Numeric
	if err := n.Scan(text); err != nil {
		t.Fatalf("numeric scan: %v", err)
	}
	return model.ExactNumeric{Numeric: n}
}

func TestToManualArticleResponse(t *testing.T) {
	now := time.Now()
	staffID := uint64(5)

	tests := []struct {
		name string
		in   *PublishedArticle
	}{
		{
			name: "converts full published article",
			in: &PublishedArticle{
				ID:              1,
				Category:        "screens",
				Slug:            "intro",
				Title:           "はじめに",
				OrderValue:      numericForTest(t, "1.5"),
				Section:         "基本",
				BodyMarkdown:    "# はじめに",
				EditedByStaffID: &staffID,
				EditedAt:        now,
				CreatedAt:       now,
				UpdatedAt:       now,
			},
		},
		{
			name: "converts zero-value article with nil staff",
			in:   &PublishedArticle{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toManualArticleResponse(tt.in)
			assert.Equal(t, tt.in.ID, got.ID)
			assert.Equal(t, string(tt.in.Category), got.Category)
			assert.Equal(t, tt.in.Slug, got.Slug)
			assert.Equal(t, tt.in.Title, got.Title)
			assert.Equal(t, tt.in.Section, got.Section)
			assert.Equal(t, tt.in.BodyMarkdown, got.BodyMarkdown)
			assert.Equal(t, tt.in.EditedByStaffID, got.UpdatedByStaffID)
		})
	}
}

func TestToManualArticleVersionDTO_SourceKindAndDigest(t *testing.T) {
	digest := make([]byte, 32)
	revision := int64(2)
	kind := "edited"
	v := &model.ManualArticleVersion{
		ID:              9,
		ArticleID:       4,
		Title:           "t",
		OrderValue:      numericForTest(t, "0"),
		Section:         "s",
		BodyMarkdown:    "b",
		ContentDigest:   &digest,
		ContentRevision: &revision,
		SourceKind:      &kind,
	}
	dto := toManualArticleVersionDTO(v)
	assert.Equal(t, "9", dto.ID)
	assert.Equal(t, "4", dto.ArticleID)
	assert.Equal(t, "edited", dto.SourceKind)
	assert.NotNil(t, dto.ContentDigest)
	assert.Len(t, *dto.ContentDigest, 64)
	assert.Equal(t, revision, *dto.ContentRevision)
}

func TestToManualArticleVersionDTO_LegacyUnknownAndNullDigest(t *testing.T) {
	v := &model.ManualArticleVersion{ID: 1, ArticleID: 1}
	dto := toManualArticleVersionDTO(v)
	assert.Equal(t, "legacy_unknown", dto.SourceKind)
	assert.Nil(t, dto.ContentDigest)
	assert.Nil(t, dto.ContentRevision)
	assert.Nil(t, dto.RestoredFromID)
}

func TestToManualArticleEditViewResponse(t *testing.T) {
	publishedID := uint64(3)
	view := &EditView{
		Article: model.ManualArticle{
			ID:                 4,
			Category:           model.ManualCategoryScreens,
			Slug:               "intro",
			Version:            7,
			PublishedVersionID: &publishedID,
		},
		Current: nil,
	}
	res := toManualArticleEditViewResponse(view)
	assert.Equal(t, "4", res.ID)
	assert.Equal(t, "screens", res.Category)
	assert.Equal(t, int64(7), res.Version)
	assert.Nil(t, res.Current)
	assert.Equal(t, "3", *res.PublishedVersionID)
	assert.False(t, res.Hidden)
}

func TestReceiptFromStoredResult(t *testing.T) {
	key := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	res := &operation.Result{
		ResultRefs: map[string]string{
			"article_id":            "4",
			"operation_history_id":  "9",
			"content_version_id":    "10",
			"content_digest":        "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			"published_version_id":  "3",
		},
		CommittedVersions: map[string]string{
			"previous_version":  "2",
			"resulting_version": "3",
			"content_revision":  "2",
		},
	}
	receipt, ok := receiptFromStoredResult(key, res)
	assert.True(t, ok)
	assert.Equal(t, "11111111-1111-4111-8111-111111111111", receipt.OperationKey)
	assert.Equal(t, "4", receipt.ArticleID)
	assert.Equal(t, "edit", receipt.Intent)
	assert.Equal(t, int64(2), receipt.PreviousVersion)
	assert.Equal(t, int64(3), receipt.ResultingVersion)
	assert.Equal(t, "10", receipt.ContentVersionID)
	assert.Equal(t, "3", *receipt.PublishedVersionID)
}

func TestReceiptFromStoredResult_NullPublishedAndMissingMembers(t *testing.T) {
	key := uuid.New()
	res := &operation.Result{
		ResultRefs: map[string]string{
			"article_id":         "4",
			"content_version_id": "10",
			"content_digest":     "a",
		},
		CommittedVersions: map[string]string{
			"previous_version":  "2",
			"resulting_version": "3",
		},
	}
	receipt, ok := receiptFromStoredResult(key, res)
	assert.True(t, ok)
	assert.Nil(t, receipt.PublishedVersionID)

	_, ok = receiptFromStoredResult(key, &operation.Result{})
	assert.False(t, ok, "missing stored members must not produce a fabricated receipt")
}
