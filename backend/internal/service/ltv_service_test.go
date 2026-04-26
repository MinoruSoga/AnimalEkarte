package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- mockLtvRepository ----

type mockLtvRepository struct {
	findOwnerLTVFn func(ctx context.Context, params repository.FindOwnerLTVParams) ([]repository.OwnerLTVRow, error)
}

func (m *mockLtvRepository) FindOwnerLTV(ctx context.Context, params repository.FindOwnerLTVParams) ([]repository.OwnerLTVRow, error) {
	if m.findOwnerLTVFn != nil {
		return m.findOwnerLTVFn(ctx, params)
	}
	return nil, nil
}

// ---- 純粋関数テスト ----

func TestCpmStageToAPI(t *testing.T) {
	cases := []struct {
		stage CPMStage
		want  string
	}{
		{CPMStageEncounter, "encounter"},
		{CPMStageGrowing, "growing"},
		{CPMStageCore, "core"},
		{CPMStageNoah, "noah"},
		{CPMStageSpot, "spot"},
		{CPMStageDormant, "dormant"},
	}
	for _, tc := range cases {
		t.Run(string(tc.stage), func(t *testing.T) {
			assert.Equal(t, tc.want, cpmStageToAPI(tc.stage))
		})
	}
}

func TestApiToCPMStage(t *testing.T) {
	cases := []struct {
		name   string
		want   CPMStage
		wantOK bool
	}{
		{"encounter", CPMStageEncounter, true},
		{"growing", CPMStageGrowing, true},
		{"core", CPMStageCore, true},
		{"noah", CPMStageNoah, true},
		{"spot", CPMStageSpot, true},
		{"dormant", CPMStageDormant, true},
		{"", "", false},
		{"invalid", "", false},
		{"ENCOUNTER", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := apiToCPMStage(tc.name)
			assert.Equal(t, tc.wantOK, ok)
			if tc.wantOK {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestMaskLineUserID(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"U1234abcd", "U****abcd"},
		{"", "U****"},
		{"abc", "U****"},
		{"1234", "U****1234"},
		{"U00000000000000000000000000000001", "U****0001"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.want, maskLineUserID(tc.input))
		})
	}
}

// ---- ListOwnerLTV テスト ----

func TestListOwnerLTV_InvalidCPMFilter(t *testing.T) {
	svc := NewLtvService(&mockLtvRepository{}, &mockLstepTagCacheRepository{})
	_, err := svc.ListOwnerLTV(context.Background(), 1, ListOwnerLTVInput{CPMStageFilter: "unknown"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cpm_stage")
}

func TestListOwnerLTV_CPMFilter(t *testing.T) {
	lineID := "U1234567890"
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	// 2 rows: one encounter (total=5000, annual=1), one core (total=100000, annual=10)
	rows := []repository.OwnerLTVRow{
		{OwnerID: 1, OwnerName: "飼主A", LineUserID: &lineID, TotalAmount: 5000, TotalVisitCount: 1, AnnualVisitCount: 1, LastVisitDate: &yesterday},
		{OwnerID: 2, OwnerName: "飼主B", LineUserID: &lineID, TotalAmount: 100000, TotalVisitCount: 20, AnnualVisitCount: 10, LastVisitDate: &yesterday},
	}

	repo := &mockLtvRepository{
		findOwnerLTVFn: func(_ context.Context, _ repository.FindOwnerLTVParams) ([]repository.OwnerLTVRow, error) {
			return rows, nil
		},
	}
	svc := NewLtvService(repo, &mockLstepTagCacheRepository{})

	result, err := svc.ListOwnerLTV(context.Background(), 1, ListOwnerLTVInput{CPMStageFilter: "core", PerPage: 50})
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	require.Len(t, result.Items, 1)
	assert.Equal(t, uint64(2), result.Items[0].OwnerID)
	assert.Equal(t, "core", result.Items[0].CPMStageAPI)
}

func TestListOwnerLTV_Pagination(t *testing.T) {
	lineID := "Utest"
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	rows := make([]repository.OwnerLTVRow, 10)
	for i := range rows {
		id := uint64(i + 1)
		rows[i] = repository.OwnerLTVRow{
			OwnerID: id, OwnerName: "飼主", LineUserID: &lineID,
			TotalAmount: 0, TotalVisitCount: 1, AnnualVisitCount: 1, LastVisitDate: &yesterday,
		}
	}

	repo := &mockLtvRepository{
		findOwnerLTVFn: func(_ context.Context, _ repository.FindOwnerLTVParams) ([]repository.OwnerLTVRow, error) {
			return rows, nil
		},
	}
	svc := NewLtvService(repo, &mockLstepTagCacheRepository{})

	result, err := svc.ListOwnerLTV(context.Background(), 1, ListOwnerLTVInput{Page: 2, PerPage: 3})
	require.NoError(t, err)
	assert.Equal(t, 10, result.Total)
	assert.Equal(t, 2, result.Page)
	assert.Equal(t, 3, result.PerPage)
	assert.Len(t, result.Items, 3)
	assert.Equal(t, uint64(4), result.Items[0].OwnerID)
}

func TestListOwnerLTV_LineIDMasked(t *testing.T) {
	lineID := "U1234567890abcdef"
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	repo := &mockLtvRepository{
		findOwnerLTVFn: func(_ context.Context, _ repository.FindOwnerLTVParams) ([]repository.OwnerLTVRow, error) {
			return []repository.OwnerLTVRow{
				{OwnerID: 1, OwnerName: "飼主A", LineUserID: &lineID, LastVisitDate: &yesterday, TotalVisitCount: 1, AnnualVisitCount: 1},
			}, nil
		},
	}
	svc := NewLtvService(repo, &mockLstepTagCacheRepository{})

	result, err := svc.ListOwnerLTV(context.Background(), 1, ListOwnerLTVInput{PerPage: 50})
	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	require.NotNil(t, result.Items[0].LineUserIDMasked)
	assert.Equal(t, "U****cdef", *result.Items[0].LineUserIDMasked)
	assert.True(t, result.Items[0].HasLine)
}

func TestListOwnerLTV_RepoError(t *testing.T) {
	repo := &mockLtvRepository{
		findOwnerLTVFn: func(_ context.Context, _ repository.FindOwnerLTVParams) ([]repository.OwnerLTVRow, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewLtvService(repo, &mockLstepTagCacheRepository{})
	_, err := svc.ListOwnerLTV(context.Background(), 1, ListOwnerLTVInput{PerPage: 50})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find owner ltv")
}

// ---- SyncLtvTags テスト ----

func TestSyncLtvTags_AutoManagedTag(t *testing.T) {
	svc := NewLtvService(&mockLtvRepository{}, &mockLstepTagCacheRepository{})
	_, err := svc.SyncLtvTags(context.Background(), 1, SyncLtvTagsInput{TagName: "vaccine_2024-01"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "自動管理タグ")
}

func TestSyncLtvTags_DryRun(t *testing.T) {
	lineID := "Utest"
	rows := []repository.OwnerLTVRow{
		{OwnerID: 1, LineUserID: &lineID},
		{OwnerID: 2, LineUserID: &lineID},
	}

	upsertCalled := 0
	cache := &mockLstepTagCacheRepository{
		upsertTagFn: func(_ context.Context, _, _ uint64, _, _ string) error {
			upsertCalled++
			return nil
		},
	}
	repo := &mockLtvRepository{
		findOwnerLTVFn: func(_ context.Context, _ repository.FindOwnerLTVParams) ([]repository.OwnerLTVRow, error) {
			return rows, nil
		},
	}
	svc := NewLtvService(repo, cache)

	result, err := svc.SyncLtvTags(context.Background(), 1, SyncLtvTagsInput{TagName: "campaign_spring", DryRun: true})
	require.NoError(t, err)
	assert.True(t, result.DryRun)
	assert.Equal(t, 2, result.Total)
	assert.Equal(t, 2, result.Synced)
	assert.Equal(t, 0, result.Skipped)
	assert.Equal(t, 0, upsertCalled, "dry-run should not call UpsertTag")
}

func TestSyncLtvTags_SkipsOptOut(t *testing.T) {
	lineID := "Utest"
	rows := []repository.OwnerLTVRow{
		{OwnerID: 1, LineUserID: &lineID, LstepOptOut: true},
		{OwnerID: 2, LineUserID: nil},
		{OwnerID: 3, LineUserID: &lineID},
	}

	repo := &mockLtvRepository{
		findOwnerLTVFn: func(_ context.Context, _ repository.FindOwnerLTVParams) ([]repository.OwnerLTVRow, error) {
			return rows, nil
		},
	}
	svc := NewLtvService(repo, &mockLstepTagCacheRepository{})

	result, err := svc.SyncLtvTags(context.Background(), 1, SyncLtvTagsInput{TagName: "campaign_spring", DryRun: false})
	require.NoError(t, err)
	assert.Equal(t, 3, result.Total)
	assert.Equal(t, 1, result.Synced)
	assert.Equal(t, 2, result.Skipped)
}

func TestSyncLtvTags_RepoError(t *testing.T) {
	repo := &mockLtvRepository{
		findOwnerLTVFn: func(_ context.Context, _ repository.FindOwnerLTVParams) ([]repository.OwnerLTVRow, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewLtvService(repo, &mockLstepTagCacheRepository{})
	_, err := svc.SyncLtvTags(context.Background(), 1, SyncLtvTagsInput{TagName: "campaign_spring"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find owner ltv for sync")
}
