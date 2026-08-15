package lstep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// This file extends checkup_sync_service_test.go's PreviewCheckupSync coverage with branches not
// exercised there: nil input, repository error, thresholds lookup error, and tag-cache lookup
// error for line-linked owners.

func TestCheckupSyncService_PreviewCheckupSync_NilInput(t *testing.T) {
	svc := newCheckupSyncSvcForPreview(nil)
	result, err := svc.PreviewCheckupSync(context.Background(), 1, nil, nil)
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, result)
}

func TestCheckupSyncPreview_Bounds(t *testing.T) {
	// BUG-032: hard bounds so preview cannot hang unbounded.
	assert.Equal(t, 15*time.Second, CheckupSyncPreviewTimeout)
	assert.Equal(t, 100, CheckupSyncPreviewOwnerCap)
	assert.Equal(t, 500, CheckupSyncPreviewRowLimit)
	assert.LessOrEqual(t, CheckupSyncPreviewOwnerCap, CheckupSyncPreviewRowLimit)
}

func TestCheckupSyncService_PreviewCheckupSync_RepositoryError(t *testing.T) {
	repo := &mockCheckupSyncRepository{
		findCheckupSyncPreviewFn: func(_ context.Context, _ *FindCheckupSyncPreviewParams) ([]CheckupSyncPreviewRow, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewCheckupSyncService(repo, &mockOwnerRepository{}, &mockPetRepository{}, &mockLstepTagCacheRepository{}, &mockLstepSettingsService{}, &mockAuditService{})
	result, err := svc.PreviewCheckupSync(context.Background(), 1, &PreviewCheckupSyncInput{CheckupType: "annual"}, nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestCheckupSyncService_PreviewCheckupSync_DeadlineReturnsInvalidInput(t *testing.T) {
	// BUG-030: deadline must not become opaque 500.
	repo := &mockCheckupSyncRepository{
		findCheckupSyncPreviewFn: func(_ context.Context, _ *FindCheckupSyncPreviewParams) ([]CheckupSyncPreviewRow, error) {
			return nil, context.DeadlineExceeded
		},
	}
	svc := NewCheckupSyncService(repo, &mockOwnerRepository{}, &mockPetRepository{}, &mockLstepTagCacheRepository{}, &mockLstepSettingsService{}, &mockAuditService{})
	result, err := svc.PreviewCheckupSync(context.Background(), 1, &PreviewCheckupSyncInput{CheckupType: "annual"}, nil)
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, result)
}

// thresholdsFailingSettingsService embeds mockLstepSettingsService (stable) and overrides
// GetCPMV1Thresholds, which the base mock hardcodes to always succeed with no injection hook.
type thresholdsFailingSettingsService struct {
	*mockLstepSettingsService
}

func (m *thresholdsFailingSettingsService) GetCPMV1Thresholds(_ context.Context, _ uint64) (model.CPMV1Thresholds, error) {
	return model.CPMV1Thresholds{}, errors.New("db error")
}

func TestCheckupSyncService_PreviewCheckupSync_ThresholdsError(t *testing.T) {
	repo := &mockCheckupSyncRepository{
		findCheckupSyncPreviewFn: func(_ context.Context, _ *FindCheckupSyncPreviewParams) ([]CheckupSyncPreviewRow, error) {
			return nil, nil
		},
	}
	svc := NewCheckupSyncService(
		repo,
		&mockOwnerRepository{},
		&mockPetRepository{},
		&mockLstepTagCacheRepository{},
		&thresholdsFailingSettingsService{mockLstepSettingsService: &mockLstepSettingsService{}},
		&mockAuditService{},
	)
	result, err := svc.PreviewCheckupSync(context.Background(), 1, &PreviewCheckupSyncInput{CheckupType: "annual"}, nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestCheckupSyncService_PreviewCheckupSync_TagCacheLookupError(t *testing.T) {
	lineID1 := "U_test_1"
	lineID2 := "U_test_2"
	rows := []CheckupSyncPreviewRow{
		{OwnerID: 1, OwnerName: "line-linked-1", LineUserID: &lineID1, LstepOptOut: false, LivingPetCount: 1},
		{OwnerID: 2, OwnerName: "line-linked-2", LineUserID: &lineID2, LstepOptOut: false, LivingPetCount: 1},
	}
	repo := &mockCheckupSyncRepository{
		findCheckupSyncPreviewFn: func(_ context.Context, _ *FindCheckupSyncPreviewParams) ([]CheckupSyncPreviewRow, error) {
			return rows, nil
		},
	}
	tagCacheRepo := &mockLstepTagCacheRepository{
		findByOwnersFn: func(_ context.Context, _ uint64, _ []uint64) (map[uint64][]*model.LstepTagCache, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewCheckupSyncService(repo, &mockOwnerRepository{}, &mockPetRepository{}, tagCacheRepo, &mockLstepSettingsService{}, &mockAuditService{})
	result, err := svc.PreviewCheckupSync(context.Background(), 1, &PreviewCheckupSyncInput{CheckupType: "annual"}, nil)
	// G7-2: タグキャッシュはバッチ取得(FindByOwners)に変更されたため、失敗は per-owner ではなく
	// 全体に及ぶ。line連携済み全員の currentTags が空スライスにフォールバックし、プレビュー自体は継続する
	// （この一括失敗時の挙動差は BE-refactor.md G7-2 の実装手順で仕様として固定されている）。
	assert.NoError(t, err)
	if assert.NotNil(t, result) && assert.Len(t, result.Owners, 2) {
		assert.Empty(t, result.Owners[0].CurrentTags)
		assert.Empty(t, result.Owners[1].CurrentTags)
	}
}

func TestCheckupSyncService_PreviewCheckupSync_BatchLoadsTagsWithinClinicScope(t *testing.T) {
	const clinicID = uint64(17)
	lineID1 := "U_test_1"
	lineID3 := "U_test_3"
	rows := []CheckupSyncPreviewRow{
		{OwnerID: 1, LineUserID: &lineID1, LivingPetCount: 1},
		{OwnerID: 2, LineUserID: nil, LivingPetCount: 1},
		{OwnerID: 3, LineUserID: &lineID3, LivingPetCount: 1},
	}

	repo := &mockCheckupSyncRepository{
		findCheckupSyncPreviewFn: func(_ context.Context, params *FindCheckupSyncPreviewParams) ([]CheckupSyncPreviewRow, error) {
			assert.Equal(t, clinicID, params.ClinicID)
			return rows, nil
		},
	}
	settings := &mockLstepSettingsService{
		getCPMV1ThresholdsFn: func(_ context.Context, gotClinicID uint64) (model.CPMV1Thresholds, error) {
			assert.Equal(t, clinicID, gotClinicID)
			return model.CPMV1Thresholds{}.WithDefaults(), nil
		},
	}
	findByOwnersCalls := 0
	cache := &mockLstepTagCacheRepository{
		findByOwnersFn: func(_ context.Context, gotClinicID uint64, ownerIDs []uint64) (map[uint64][]*model.LstepTagCache, error) {
			findByOwnersCalls++
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, []uint64{1, 3}, ownerIDs)
			return map[uint64][]*model.LstepTagCache{
				1: {{TagName: "tag-one"}},
				3: {{TagName: "tag-three"}},
			}, nil
		},
	}
	audit := &mockAuditService{
		logLstepOperationWithMetadataFn: func(_ context.Context, gotClinicID uint64, _ *uint64, _, _ string, _ *uint64, _ any) error {
			assert.Equal(t, clinicID, gotClinicID)
			return nil
		},
	}
	svc := NewCheckupSyncService(repo, &mockOwnerRepository{}, &mockPetRepository{}, cache, settings, audit)

	result, err := svc.PreviewCheckupSync(context.Background(), clinicID, &PreviewCheckupSyncInput{}, nil)

	assert.NoError(t, err)
	assert.Equal(t, 1, findByOwnersCalls)
	if assert.NotNil(t, result) && assert.Len(t, result.Owners, 3) {
		assert.Equal(t, []string{"tag-one"}, result.Owners[0].CurrentTags)
		assert.Empty(t, result.Owners[1].CurrentTags)
		assert.Equal(t, []string{"tag-three"}, result.Owners[2].CurrentTags)
	}
}
