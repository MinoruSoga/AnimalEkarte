package lstep

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock LstepTagCacheRepository for summary tests ----
// Distinct from mockLstepTagCacheRepository (lstep_lifecycle_service_test.go)
// because that mock has hardcoded nil returns for TagSummary/FindOwnersByTag.

type mockTagCacheSummaryRepo struct {
	tagSummaryFn      func(ctx context.Context, clinicID uint64) ([]TagSummaryRow, int64, error)
	findOwnersByTagFn func(ctx context.Context, clinicID uint64, tagName, nameQuery string, offset, limit int) ([]TagOwnerRow, int64, error)
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
func (m *mockTagCacheSummaryRepo) FindByOwners(_ context.Context, _ uint64, _ []uint64) (map[uint64][]*model.LstepTagCache, error) {
	return map[uint64][]*model.LstepTagCache{}, nil
}
func (m *mockTagCacheSummaryRepo) TagSummary(ctx context.Context, clinicID uint64) ([]TagSummaryRow, int64, error) {
	if m.tagSummaryFn != nil {
		return m.tagSummaryFn(ctx, clinicID)
	}
	return []TagSummaryRow{}, 0, nil
}
func (m *mockTagCacheSummaryRepo) FindOwnersByTag(ctx context.Context, clinicID uint64, tagName, nameQuery string, offset, limit int) ([]TagOwnerRow, int64, error) {
	if m.findOwnersByTagFn != nil {
		return m.findOwnersByTagFn(ctx, clinicID, tagName, nameQuery, offset, limit)
	}
	return []TagOwnerRow{}, 0, nil
}
func (m *mockTagCacheSummaryRepo) BulkReplaceOwnerTags(_ context.Context, _, _ uint64, _ []TagEntry) error {
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
				tagSummaryFn: func(_ context.Context, _ uint64) ([]TagSummaryRow, int64, error) {
					return []TagSummaryRow{{TagName: "my_tag", OwnerCount: 3}}, 10, nil
				},
			},
		},
		{
			name: "repo error",
			repo: &mockTagCacheSummaryRepo{
				tagSummaryFn: func(_ context.Context, _ uint64) ([]TagSummaryRow, int64, error) {
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
				findOwnersByTagFn: func(_ context.Context, _ uint64, _, _ string, _, _ int) ([]TagOwnerRow, int64, error) {
					return []TagOwnerRow{{OwnerID: 1, OwnerName: "田中 太郎"}}, 1, nil
				},
			},
			input: ListOwnersByTagInput{TagName: "my_tag", Page: 1, PerPage: 20},
		},
		{
			name: "repo error",
			repo: &mockTagCacheSummaryRepo{
				findOwnersByTagFn: func(_ context.Context, _ uint64, _, _ string, _, _ int) ([]TagOwnerRow, int64, error) {
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
		findOwnersByTagFn: func(_ context.Context, _ uint64, _, _ string, _, _ int) ([]TagOwnerRow, int64, error) {
			return []TagOwnerRow{{OwnerID: 1, OwnerName: "田中 太郎"}}, 1, nil
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

func TestExportOwnersByTagCSV_SanitizesFormulaCPMStage(t *testing.T) {
	repo := &mockTagCacheSummaryRepo{
		findOwnersByTagFn: func(_ context.Context, _ uint64, _, _ string, _, _ int) ([]TagOwnerRow, int64, error) {
			return []TagOwnerRow{{
				OwnerID:   1,
				OwnerName: "田中",
				Tags:      []string{"cpm_=CMD()"},
			}}, 1, nil
		},
	}
	svc := NewLstepTagSummaryService(repo)

	var buf bytes.Buffer
	err := svc.ExportOwnersByTagCSV(context.Background(), 1, "my_tag", "", &buf)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "'=CMD()")
	assert.NotContains(t, buf.String(), ",=CMD()")
}

// BUG-464: total above hard cap must fail closed before any CSV body is written.
func TestExportOwnersByTagCSV_FailsClosedWhenTotalExceedsCap(t *testing.T) {
	repo := &mockTagCacheSummaryRepo{
		findOwnersByTagFn: func(_ context.Context, _ uint64, _, _ string, offset, limit int) ([]TagOwnerRow, int64, error) {
			assert.Equal(t, 0, offset)
			assert.Equal(t, exportOwnersByTagCSVMaxRows, limit)
			// Simulate 5001 total while returning only the capped page.
			rows := make([]TagOwnerRow, exportOwnersByTagCSVMaxRows)
			for i := range rows {
				rows[i] = TagOwnerRow{OwnerID: uint64(i + 1), OwnerName: "x"}
			}
			return rows, int64(exportOwnersByTagCSVMaxRows) + 1, nil
		},
	}
	svc := NewLstepTagSummaryService(repo)

	var buf bytes.Buffer
	err := svc.ExportOwnersByTagCSV(context.Background(), 1, "my_tag", "", &buf)
	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "over-cap export must be Conflict, got %v", err)
	assert.Empty(t, buf.Bytes(), "must not write CSV body when fail-closed")
}

// ---- sanitizeCSVCell / extractLastVisitDate / extractCPMStage 直接テスト ----

func TestSanitizeCSVCell(t *testing.T) {
	tests := []struct {
		name string
		cell string
		want string
	}{
		{name: "空文字はそのまま", cell: "", want: ""},
		{name: "通常の文字列はそのまま", cell: "田中 太郎", want: "田中 太郎"},
		{name: "=で始まる場合は引用符を前置", cell: "=SUM(A1:A2)", want: "'=SUM(A1:A2)"},
		{name: "+で始まる場合は引用符を前置", cell: "+1234", want: "'+1234"},
		{name: "-で始まる場合は引用符を前置", cell: "-1234", want: "'-1234"},
		{name: "@で始まる場合は引用符を前置", cell: "@cmd", want: "'@cmd"},
		{name: "先頭以外に記号があっても対象外", cell: "a=b", want: "a=b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, sanitizeCSVCell(tt.cell))
		})
	}
}

func TestExtractLastVisitDate(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want *string
	}{
		{name: "該当タグなし", tags: []string{"cpm_active"}, want: nil},
		{name: "タグなし(空スライス)", tags: []string{}, want: nil},
		{name: "last_visit_タグあり", tags: []string{"cpm_active", "last_visit_2024-01-15"}, want: strPtr("2024-01-15")},
		{name: "日付長さが10文字でない場合は無視", tags: []string{"last_visit_2024-1-1"}, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, extractLastVisitDate(tt.tags))
		})
	}
}

func TestExtractCPMStage(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want string
	}{
		{name: "cpm_タグあり", tags: []string{"last_visit_2024-01-01", "cpm_dormant"}, want: "dormant"},
		{name: "cpm_タグなし", tags: []string{"last_visit_2024-01-01"}, want: ""},
		{name: "空スライス", tags: []string{}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, extractCPMStage(tt.tags))
		})
	}
}

// ---- ListOwnersByTag ページネーション補正 ----

func TestListOwnersByTag_Pagination(t *testing.T) {
	tests := []struct {
		name        string
		input       ListOwnersByTagInput
		wantPage    int
		wantPerPage int
	}{
		{
			name:        "PerPageが0以下の場合は20に補正",
			input:       ListOwnersByTagInput{TagName: "t", PerPage: 0},
			wantPage:    1,
			wantPerPage: 20,
		},
		{
			name:        "PerPageが100超の場合は100に補正",
			input:       ListOwnersByTagInput{TagName: "t", PerPage: 500},
			wantPage:    1,
			wantPerPage: 100,
		},
		{
			name:        "Pageが0以下の場合は1に補正",
			input:       ListOwnersByTagInput{TagName: "t", Page: 0, PerPage: 10},
			wantPage:    1,
			wantPerPage: 10,
		},
		{
			name:        "PageとPerPageが指定通り",
			input:       ListOwnersByTagInput{TagName: "t", Page: 3, PerPage: 10},
			wantPage:    3,
			wantPerPage: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotOffset, gotLimit int
			repo := &mockTagCacheSummaryRepo{
				findOwnersByTagFn: func(_ context.Context, _ uint64, _, _ string, offset, limit int) ([]TagOwnerRow, int64, error) {
					gotOffset = offset
					gotLimit = limit
					return []TagOwnerRow{{OwnerID: 1, OwnerName: "a", Reason: strPtr("R")}}, 1, nil
				},
			}
			svc := NewLstepTagSummaryService(repo)

			res, err := svc.ListOwnersByTag(context.Background(), 1, tt.input)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantPage, res.Page)
			assert.Equal(t, tt.wantPerPage, res.PerPage)
			assert.Equal(t, (tt.wantPage-1)*tt.wantPerPage, gotOffset)
			assert.Equal(t, tt.wantPerPage, gotLimit)
			assert.Len(t, res.Owners, 1)
			assert.Equal(t, "R", *res.Owners[0].Reason)
		})
	}
}

// ---- ExportOwnersByTagCSV 追加分岐 ----

func TestExportOwnersByTagCSV_RepoError(t *testing.T) {
	repo := &mockTagCacheSummaryRepo{
		findOwnersByTagFn: func(_ context.Context, _ uint64, _, _ string, _, _ int) ([]TagOwnerRow, int64, error) {
			return nil, 0, errors.New("db error")
		},
	}
	svc := NewLstepTagSummaryService(repo)

	var buf bytes.Buffer
	err := svc.ExportOwnersByTagCSV(context.Background(), 1, "my_tag", "", &buf)
	assert.Error(t, err)
}

func TestExportOwnersByTagCSV_FormulaInjectionPrevention(t *testing.T) {
	repo := &mockTagCacheSummaryRepo{
		findOwnersByTagFn: func(_ context.Context, _ uint64, _, _ string, _, _ int) ([]TagOwnerRow, int64, error) {
			return []TagOwnerRow{{OwnerID: 1, OwnerName: "=SUM(A1:A2)"}}, 1, nil
		},
	}
	svc := NewLstepTagSummaryService(repo)

	var buf bytes.Buffer
	err := svc.ExportOwnersByTagCSV(context.Background(), 1, "my_tag", "", &buf)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "'=SUM(A1:A2)")
}

func TestExportOwnersByTagCSV_CPMStageAndLastVisit(t *testing.T) {
	repo := &mockTagCacheSummaryRepo{
		findOwnersByTagFn: func(_ context.Context, _ uint64, _, _ string, _, _ int) ([]TagOwnerRow, int64, error) {
			return []TagOwnerRow{
				{OwnerID: 1, OwnerName: "田中 太郎", Tags: []string{"last_visit_2024-01-01", "cpm_dormant"}},
				{OwnerID: 2, OwnerName: "鈴木 花子", Tags: []string{}},
			}, 2, nil
		},
	}
	svc := NewLstepTagSummaryService(repo)

	var buf bytes.Buffer
	err := svc.ExportOwnersByTagCSV(context.Background(), 1, "my_tag", "", &buf)
	assert.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "2024-01-01")
	assert.Contains(t, out, "dormant")
}

// failWriter は指定回数目以降の Write 呼び出しでエラーを返す io.Writer テストダブル。
type failWriter struct {
	failAt int
	calls  int
}

func (w *failWriter) Write(p []byte) (int, error) {
	w.calls++
	if w.calls >= w.failAt {
		return 0, errors.New("write failed")
	}
	return len(p), nil
}

func TestExportOwnersByTagCSV_BOMWriteError(t *testing.T) {
	repo := &mockTagCacheSummaryRepo{
		findOwnersByTagFn: func(_ context.Context, _ uint64, _, _ string, _, _ int) ([]TagOwnerRow, int64, error) {
			return []TagOwnerRow{}, 0, nil
		},
	}
	svc := NewLstepTagSummaryService(repo)

	w := &failWriter{failAt: 1}
	err := svc.ExportOwnersByTagCSV(context.Background(), 1, "my_tag", "", w)
	assert.Error(t, err)
}
