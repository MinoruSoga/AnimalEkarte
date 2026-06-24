package service

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- mock LstepTagCacheRepository for summary tests ----
// Distinct from mockLstepTagCacheRepository (lstep_lifecycle_service_test.go)
// because that mock has hardcoded nil returns for TagSummary/FindOwnersByTag.

type mockTagCacheSummaryRepo struct {
	tagSummaryFn      func(ctx context.Context, clinicID uint64) ([]repository.TagSummaryRow, int64, error)
	findOwnersByTagFn func(ctx context.Context, clinicID uint64, tagName, nameQuery string, offset, limit int) ([]repository.TagOwnerRow, int64, error)
}

func (m *mockTagCacheSummaryRepo) UpsertTag(_ context.Context, _, _ uint64, _, _, _ string) error {
	return nil
}
func (m *mockTagCacheSummaryRepo) DeleteTag(_ context.Context, _, _ uint64, _ string) error {
	return nil
}
func (m *mockTagCacheSummaryRepo) DeleteAllByOwner(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *mockTagCacheSummaryRepo) FindByOwner(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
	return nil, nil
}
func (m *mockTagCacheSummaryRepo) CountByTag(_ context.Context, _ uint64, _ string) (int64, error) {
	return 0, nil
}
func (m *mockTagCacheSummaryRepo) TagSummary(ctx context.Context, clinicID uint64) ([]repository.TagSummaryRow, int64, error) {
	if m.tagSummaryFn != nil {
		return m.tagSummaryFn(ctx, clinicID)
	}
	return []repository.TagSummaryRow{}, 0, nil
}
func (m *mockTagCacheSummaryRepo) FindOwnersByTag(ctx context.Context, clinicID uint64, tagName, nameQuery string, offset, limit int) ([]repository.TagOwnerRow, int64, error) {
	if m.findOwnersByTagFn != nil {
		return m.findOwnersByTagFn(ctx, clinicID, tagName, nameQuery, offset, limit)
	}
	return []repository.TagOwnerRow{}, 0, nil
}
func (m *mockTagCacheSummaryRepo) BulkReplaceOwnerTags(_ context.Context, _, _ uint64, _ []repository.TagEntry) error {
	return nil
}
func (m *mockTagCacheSummaryRepo) FindOwnerIDsByTag(_ context.Context, _ uint64, _ string) ([]uint64, error) {
	return nil, nil
}

// ---- tests ----

func TestGetTagSummary(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockTagCacheSummaryRepo
		wantErr bool
	}{
		{
			name: "success",
			repo: &mockTagCacheSummaryRepo{
				tagSummaryFn: func(_ context.Context, _ uint64) ([]repository.TagSummaryRow, int64, error) {
					return []repository.TagSummaryRow{{TagName: "my_tag", OwnerCount: 3}}, 10, nil
				},
			},
		},
		{
			name: "repo error",
			repo: &mockTagCacheSummaryRepo{
				tagSummaryFn: func(_ context.Context, _ uint64) ([]repository.TagSummaryRow, int64, error) {
					return nil, 0, errors.New("db error")
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewLstepTagSummaryService(tt.repo)
			res, err := svc.GetTagSummary(context.Background(), 1)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, res.Tags, 1)
				assert.Equal(t, int64(10), res.TotalOwnersWithLstep)
			}
		})
	}
}

func TestListOwnersByTag(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockTagCacheSummaryRepo
		input   ListOwnersByTagInput
		wantErr bool
	}{
		{
			name: "success",
			repo: &mockTagCacheSummaryRepo{
				findOwnersByTagFn: func(_ context.Context, _ uint64, _, _ string, _, _ int) ([]repository.TagOwnerRow, int64, error) {
					return []repository.TagOwnerRow{{OwnerID: 1, OwnerName: "田中 太郎"}}, 1, nil
				},
			},
			input: ListOwnersByTagInput{TagName: "my_tag", Page: 1, PerPage: 20},
		},
		{
			name: "repo error",
			repo: &mockTagCacheSummaryRepo{
				findOwnersByTagFn: func(_ context.Context, _ uint64, _, _ string, _, _ int) ([]repository.TagOwnerRow, int64, error) {
					return nil, 0, errors.New("db error")
				},
			},
			input:   ListOwnersByTagInput{TagName: "my_tag", Page: 1, PerPage: 20},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewLstepTagSummaryService(tt.repo)
			res, err := svc.ListOwnersByTag(context.Background(), 1, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, res.Owners, 1)
				assert.Equal(t, int64(1), res.Total)
			}
		})
	}
}

func TestExportOwnersByTagCSV(t *testing.T) {
	repo := &mockTagCacheSummaryRepo{
		findOwnersByTagFn: func(_ context.Context, _ uint64, _, _ string, _, _ int) ([]repository.TagOwnerRow, int64, error) {
			return []repository.TagOwnerRow{{OwnerID: 1, OwnerName: "田中 太郎"}}, 1, nil
		},
	}
	svc := NewLstepTagSummaryService(repo)

	var buf bytes.Buffer
	err := svc.ExportOwnersByTagCSV(context.Background(), 1, "my_tag", "", &buf)
	assert.NoError(t, err)

	out := buf.Bytes()
	// #179 ③: Excel が Shift-JIS と誤認するのを防ぐため先頭に UTF-8 BOM を付与する
	assert.True(t, bytes.HasPrefix(out, []byte("\xEF\xBB\xBF")), "CSV は UTF-8 BOM で始まること")
	// 日本語の飼主名が UTF-8 のまま含まれる（文字化けしない）
	assert.Contains(t, buf.String(), "田中 太郎")
	// ヘッダ行が含まれる
	assert.Contains(t, buf.String(), "owner_id")
}
