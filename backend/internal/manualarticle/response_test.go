package manualarticle

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToManualArticleResponse(t *testing.T) {
	now := time.Now()
	staffID := uint64(5)

	tests := []struct {
		name string
		in   *model.ManualArticle
	}{
		{
			name: "converts full article",
			in: &model.ManualArticle{
				ID:               1,
				Category:         "general",
				Slug:             "intro",
				Title:            "はじめに",
				OrderValue:       1.5,
				Section:          "基本",
				BodyMarkdown:     "# はじめに",
				UpdatedByStaffID: &staffID,
				CreatedAt:        now,
				UpdatedAt:        now,
			},
		},
		{
			name: "converts zero-value article with nil UpdatedByStaffID",
			in:   &model.ManualArticle{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toManualArticleResponse(tt.in)

			assert.Equal(t, tt.in.ID, got.ID)
			assert.Equal(t, string(tt.in.Category), got.Category)
			assert.Equal(t, tt.in.Slug, got.Slug)
			assert.Equal(t, tt.in.Title, got.Title)
			assert.Equal(t, tt.in.OrderValue, got.OrderValue)
			assert.Equal(t, tt.in.Section, got.Section)
			assert.Equal(t, tt.in.BodyMarkdown, got.BodyMarkdown)
			assert.Equal(t, tt.in.UpdatedByStaffID, got.UpdatedByStaffID)
		})
	}
}

func TestToManualArticleListResponse(t *testing.T) {
	tests := []struct {
		name string
		in   []model.ManualArticle
	}{
		{
			name: "converts multiple articles",
			in: []model.ManualArticle{
				{ID: 1, Category: "general", Slug: "a", Title: "A"},
				{ID: 2, Category: "general", Slug: "b", Title: "B"},
			},
		},
		{
			name: "converts empty slice",
			in:   []model.ManualArticle{},
		},
		{
			name: "converts nil slice",
			in:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toManualArticleListResponse(tt.in)
			assert.Len(t, got.Data, len(tt.in))
			for i, article := range tt.in {
				assert.Equal(t, article.ID, got.Data[i].ID)
				assert.Equal(t, article.Title, got.Data[i].Title)
			}
		})
	}
}

func TestToManualArticleVersionResponse(t *testing.T) {
	now := time.Now()
	staffID := uint64(9)

	tests := []struct {
		name string
		in   *model.ManualArticleVersion
	}{
		{
			name: "converts full version",
			in: &model.ManualArticleVersion{
				ID:              10,
				ArticleID:       1,
				Title:           "旧版",
				OrderValue:      2.0,
				Section:         "基本",
				BodyMarkdown:    "# 旧版",
				EditedByStaffID: &staffID,
				EditedAt:        now,
			},
		},
		{
			name: "converts zero-value version with nil EditedByStaffID",
			in:   &model.ManualArticleVersion{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toManualArticleVersionResponse(tt.in)

			assert.Equal(t, tt.in.ID, got.ID)
			assert.Equal(t, tt.in.ArticleID, got.ArticleID)
			assert.Equal(t, tt.in.Title, got.Title)
			assert.Equal(t, tt.in.OrderValue, got.OrderValue)
			assert.Equal(t, tt.in.Section, got.Section)
			assert.Equal(t, tt.in.BodyMarkdown, got.BodyMarkdown)
			assert.Equal(t, tt.in.EditedByStaffID, got.EditedByStaffID)
		})
	}
}

func TestToManualArticleVersionListResponse(t *testing.T) {
	tests := []struct {
		name string
		in   []model.ManualArticleVersion
	}{
		{
			name: "converts multiple versions",
			in: []model.ManualArticleVersion{
				{ID: 1, ArticleID: 1, Title: "v1"},
				{ID: 2, ArticleID: 1, Title: "v2"},
			},
		},
		{
			name: "converts empty slice",
			in:   []model.ManualArticleVersion{},
		},
		{
			name: "converts nil slice",
			in:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toManualArticleVersionListResponse(tt.in)
			assert.Len(t, got.Data, len(tt.in))
			for i, version := range tt.in {
				assert.Equal(t, version.ID, got.Data[i].ID)
				assert.Equal(t, version.Title, got.Data[i].Title)
			}
		})
	}
}
