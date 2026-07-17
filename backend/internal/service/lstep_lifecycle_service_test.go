package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- LstepSettingsService モック ----

type mockLstepSettingsService struct {
	getRawCredentialsFn             func(ctx context.Context, clinicID uint64) (apiKey, baseURL, lineToken string, err error)
	getSettingsFn                   func(ctx context.Context, clinicID uint64) (*LstepSettingsResponse, error)
	updateSettingsFn                func(ctx context.Context, clinicID uint64, input *UpdateLstepSettingsInput) (*LstepSettingsResponse, error)
	deleteSettingsFn                func(ctx context.Context, clinicID uint64) error
	testConnectionFn                func(ctx context.Context, clinicID uint64) (*LstepConnectionTestResult, error)
	isSyncEnabledFn                 func(ctx context.Context, clinicID uint64) (bool, error)
	getHealthPreventionThresholdsFn func(ctx context.Context, clinicID uint64) (model.HealthPreventionThresholds, error)
	getDormantThresholdsFn          func(ctx context.Context, clinicID uint64) (model.DormantThresholds, error)
}

func (m *mockLstepSettingsService) GetRawCredentials(ctx context.Context, clinicID uint64) (apiKey, baseURL, lineToken string, err error) {
	if m.getRawCredentialsFn != nil {
		return m.getRawCredentialsFn(ctx, clinicID)
	}
	return "", "", "", nil
}
func (m *mockLstepSettingsService) GetSettings(ctx context.Context, clinicID uint64) (*LstepSettingsResponse, error) {
	if m.getSettingsFn != nil {
		return m.getSettingsFn(ctx, clinicID)
	}
	return &LstepSettingsResponse{}, nil
}
func (m *mockLstepSettingsService) UpdateSettings(ctx context.Context, clinicID uint64, input *UpdateLstepSettingsInput, actorID *uint64) (*LstepSettingsResponse, error) {
	if m.updateSettingsFn != nil {
		return m.updateSettingsFn(ctx, clinicID, input)
	}
	return &LstepSettingsResponse{}, nil
}
func (m *mockLstepSettingsService) DeleteSettings(ctx context.Context, clinicID uint64, actorID *uint64) error {
	if m.deleteSettingsFn != nil {
		return m.deleteSettingsFn(ctx, clinicID)
	}
	return nil
}
func (m *mockLstepSettingsService) TestConnection(ctx context.Context, clinicID uint64) (*LstepConnectionTestResult, error) {
	if m.testConnectionFn != nil {
		return m.testConnectionFn(ctx, clinicID)
	}
	return &LstepConnectionTestResult{}, nil
}
func (m *mockLstepSettingsService) IsSyncEnabled(ctx context.Context, clinicID uint64) (bool, error) {
	if m.isSyncEnabledFn != nil {
		return m.isSyncEnabledFn(ctx, clinicID)
	}
	return true, nil
}
func (m *mockLstepSettingsService) GetCPMVersion(_ context.Context, _ uint64) (string, error) {
	return "v1", nil
}
func (m *mockLstepSettingsService) GetDormantThresholds(ctx context.Context, clinicID uint64) (model.DormantThresholds, error) {
	if m.getDormantThresholdsFn != nil {
		return m.getDormantThresholdsFn(ctx, clinicID)
	}
	return model.DormantThresholds{}.WithDefaults(), nil
}
func (m *mockLstepSettingsService) GetCPMV2Thresholds(_ context.Context, _ uint64) (model.CPMV2Thresholds, error) {
	return model.CPMV2Thresholds{}.WithDefaults(), nil
}
func (m *mockLstepSettingsService) GetCPMV1Thresholds(_ context.Context, _ uint64) (model.CPMV1Thresholds, error) {
	return model.CPMV1Thresholds{}.WithDefaults(), nil
}

func (m *mockLstepSettingsService) GetHealthPreventionThresholds(ctx context.Context, clinicID uint64) (model.HealthPreventionThresholds, error) {
	if m.getHealthPreventionThresholdsFn != nil {
		return m.getHealthPreventionThresholdsFn(ctx, clinicID)
	}
	return model.HealthPreventionThresholds{}.WithDefaults(), nil
}

// ---- LstepTagCacheRepository モック ----

type mockLstepTagCacheRepository struct {
	upsertTagFn        func(ctx context.Context, clinicID, ownerID uint64, tagName, category, reason string) error
	deleteTagFn        func(ctx context.Context, clinicID, ownerID uint64, tagName string) error
	deleteAllByOwnerFn func(ctx context.Context, clinicID, ownerID uint64) error
	findByOwnerFn      func(ctx context.Context, clinicID, ownerID uint64) ([]*model.LstepTagCache, error)
	findByOwnersFn     func(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64][]*model.LstepTagCache, error)
}

func (m *mockLstepTagCacheRepository) UpsertTag(ctx context.Context, clinicID, ownerID uint64, tagName, category, reason string) error {
	if m.upsertTagFn != nil {
		return m.upsertTagFn(ctx, clinicID, ownerID, tagName, category, reason)
	}
	return nil
}
func (m *mockLstepTagCacheRepository) DeleteTag(ctx context.Context, clinicID, ownerID uint64, tagName string) error {
	if m.deleteTagFn != nil {
		return m.deleteTagFn(ctx, clinicID, ownerID, tagName)
	}
	return nil
}
func (m *mockLstepTagCacheRepository) DeleteAllByOwner(ctx context.Context, clinicID, ownerID uint64) error {
	if m.deleteAllByOwnerFn != nil {
		return m.deleteAllByOwnerFn(ctx, clinicID, ownerID)
	}
	return nil
}
func (m *mockLstepTagCacheRepository) FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]*model.LstepTagCache, error) {
	if m.findByOwnerFn != nil {
		return m.findByOwnerFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}
func (m *mockLstepTagCacheRepository) FindByOwners(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64][]*model.LstepTagCache, error) {
	if m.findByOwnersFn != nil {
		return m.findByOwnersFn(ctx, clinicID, ownerIDs)
	}
	return map[uint64][]*model.LstepTagCache{}, nil
}
func (m *mockLstepTagCacheRepository) TagSummary(ctx context.Context, clinicID uint64) ([]repository.TagSummaryRow, int64, error) {
	return nil, 0, nil
}
func (m *mockLstepTagCacheRepository) FindOwnersByTag(ctx context.Context, clinicID uint64, tagName, nameQuery string, offset, limit int) ([]repository.TagOwnerRow, int64, error) {
	return nil, 0, nil
}
func (m *mockLstepTagCacheRepository) BulkReplaceOwnerTags(ctx context.Context, clinicID, ownerID uint64, tags []repository.TagEntry) error {
	return nil
}
func (m *mockLstepTagCacheRepository) FindOwnerIDsByTag(_ context.Context, _ uint64, _ string) ([]uint64, error) {
	return nil, nil
}

// ---- LstepTagSyncService モック ----

type mockLstepTagSyncService struct {
	syncVaccineTagFn                   func(ctx context.Context, clinicID, ownerID, vaccinationID uint64) error
	syncVisitCompletionTagsFn          func(ctx context.Context, clinicID, ownerID uint64) error
	syncNextVisitTagFn                 func(ctx context.Context, clinicID, ownerID uint64) error
	syncOwnerAnimalClassificationTagFn func(ctx context.Context, clinicID, ownerID uint64) error
	syncPetBasicInfoTagsFn             func(ctx context.Context, clinicID, ownerID uint64) error
	syncCheckupTagFn                   func(ctx context.Context, clinicID, ownerID, checkupTypeID uint64, checkupDate time.Time, nextDate *time.Time) error
	syncPrescriptionTagFn              func(ctx context.Context, clinicID, ownerID uint64) error
	syncCPMStageTagFn                  func(ctx context.Context, clinicID, ownerID uint64) error
	resyncOwnerVaccineTagsFn           func(ctx context.Context, clinicID, ownerID uint64) error
	resyncOwnerCheckupTagsFn           func(ctx context.Context, clinicID, ownerID uint64) error
}

func (m *mockLstepTagSyncService) SyncVaccineTag(ctx context.Context, clinicID, ownerID, vaccinationID uint64) error {
	if m.syncVaccineTagFn != nil {
		return m.syncVaccineTagFn(ctx, clinicID, ownerID, vaccinationID)
	}
	return nil
}
func (m *mockLstepTagSyncService) SyncVisitCompletionTags(ctx context.Context, clinicID, ownerID uint64) error {
	if m.syncVisitCompletionTagsFn != nil {
		return m.syncVisitCompletionTagsFn(ctx, clinicID, ownerID)
	}
	return nil
}
func (m *mockLstepTagSyncService) SyncOwnerAnimalClassificationTags(ctx context.Context, clinicID, ownerID uint64) error {
	if m.syncOwnerAnimalClassificationTagFn != nil {
		return m.syncOwnerAnimalClassificationTagFn(ctx, clinicID, ownerID)
	}
	return nil
}
func (m *mockLstepTagSyncService) SyncPetBasicInfoTags(ctx context.Context, clinicID, ownerID uint64) error {
	if m.syncPetBasicInfoTagsFn != nil {
		return m.syncPetBasicInfoTagsFn(ctx, clinicID, ownerID)
	}
	return nil
}
func (m *mockLstepTagSyncService) SyncNextVisitTag(ctx context.Context, clinicID, ownerID uint64) error {
	if m.syncNextVisitTagFn != nil {
		return m.syncNextVisitTagFn(ctx, clinicID, ownerID)
	}
	return nil
}
func (m *mockLstepTagSyncService) SyncCheckupTag(ctx context.Context, clinicID, ownerID, checkupTypeID uint64, checkupDate time.Time, nextDate *time.Time) error {
	if m.syncCheckupTagFn != nil {
		return m.syncCheckupTagFn(ctx, clinicID, ownerID, checkupTypeID, checkupDate, nextDate)
	}
	return nil
}
func (m *mockLstepTagSyncService) SyncPrescriptionTag(ctx context.Context, clinicID, ownerID uint64) error {
	if m.syncPrescriptionTagFn != nil {
		return m.syncPrescriptionTagFn(ctx, clinicID, ownerID)
	}
	return nil
}
func (m *mockLstepTagSyncService) SyncChronicConditionTags(_ context.Context, _, _ uint64, _ []string) error {
	return nil
}
func (m *mockLstepTagSyncService) SyncCPMStageTag(ctx context.Context, clinicID, ownerID uint64) error {
	if m.syncCPMStageTagFn != nil {
		return m.syncCPMStageTagFn(ctx, clinicID, ownerID)
	}
	return nil
}

func (m *mockLstepTagSyncService) ResyncOwnerVaccineTags(ctx context.Context, clinicID, ownerID uint64) error {
	if m.resyncOwnerVaccineTagsFn != nil {
		return m.resyncOwnerVaccineTagsFn(ctx, clinicID, ownerID)
	}
	return nil
}

func (m *mockLstepTagSyncService) ResyncOwnerCheckupTags(ctx context.Context, clinicID, ownerID uint64) error {
	if m.resyncOwnerCheckupTagsFn != nil {
		return m.resyncOwnerCheckupTagsFn(ctx, clinicID, ownerID)
	}
	return nil
}

func (m *mockLstepTagSyncService) SyncLTVTopPercent(_ context.Context, _ uint64) (int, []error) {
	return 0, nil
}

func (m *mockLstepTagSyncService) SyncVisitDormantTags(_ context.Context, _, _ uint64, _ int) error {
	return nil
}

func (m *mockLstepTagSyncService) SyncExclusionTags(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockLstepTagSyncService) SyncHealthPreventionTagsForClinic(_ context.Context, _ uint64) (int, []error) {
	return 0, nil
}

func (m *mockLstepTagSyncService) SyncDormantTagsWithThresholds(_ context.Context, _, _ uint64, _ int, _ model.DormantThresholds) error {
	return nil
}

// ---- ヘルパー ----

func newLstepLifecycleSvc(
	settingsSvc *mockLstepSettingsService,
	ownerRepo *mockOwnerRepository,
	petRepo *mockPetRepository,
	tagCacheRepo *mockLstepTagCacheRepository,
	syncSvc *mockLstepTagSyncService,
) LstepLifecycleService {
	return NewLstepLifecycleService(settingsSvc, ownerRepo, petRepo, tagCacheRepo, syncSvc, &mockAuditService{}, nil)
}

func defaultLstepSettingsSvc() *mockLstepSettingsService {
	return &mockLstepSettingsService{
		getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
			return "", "", "", nil // Lステップ未設定
		},
	}
}

// ---- テスト: HandlePetDeath ----

func TestLstepLifecycleService_HandlePetDeath(t *testing.T) {
	tests := []struct {
		name                string
		petFindFn           func(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
		petUpdateFn         func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
		findLivingByOwnerFn func(ctx context.Context, clinicID, ownerID uint64) ([]model.Pet, error)
		syncCPMFn           func(ctx context.Context, clinicID, ownerID uint64) error
		wantErr             bool
		wantErrIs           error
	}{
		{
			name: "正常: 生存ペットあり — タグ再同期",
			petFindFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return &model.Pet{ID: 1, OwnerID: 10}, nil
			},
			petUpdateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				return nil
			},
			findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) {
				return []model.Pet{{ID: 2, OwnerID: 10}}, nil
			},
		},
		{
			name: "正常: 全ペット死亡 — owner nil でも panic しない",
			petFindFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return &model.Pet{ID: 1, OwnerID: 10}, nil
			},
			petUpdateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				return nil
			},
			// findLivingByOwnerFn nil → default returns nil, nil (all dead)
			// ownerRepo.FindByID returns nil, nil — nil guard must prevent panic
		},
		{
			name: "エラー: ペットが見つからない → ErrNotFound",
			petFindFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return nil, apperrors.WrapNotFound("pet", "1")
			},
			wantErr:   true,
			wantErrIs: apperrors.ErrNotFound,
		},
		{
			name: "エラー: ペット更新失敗",
			petFindFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return &model.Pet{ID: 1, OwnerID: 10}, nil
			},
			petUpdateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				return errors.New("db error")
			},
			wantErr: true,
		},
		{
			name: "正常: 生存ペットあり — CPM 同期失敗は best-effort",
			petFindFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return &model.Pet{ID: 1, OwnerID: 10}, nil
			},
			petUpdateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				return nil
			},
			findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) {
				return []model.Pet{{ID: 2, OwnerID: 10}}, nil
			},
			syncCPMFn: func(_ context.Context, _, _ uint64) error {
				return errors.New("sync error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			petRepo := &mockPetRepository{
				findByIDFn:          tt.petFindFn,
				updateFn:            tt.petUpdateFn,
				findLivingByOwnerFn: tt.findLivingByOwnerFn,
			}
			syncSvc := &mockLstepTagSyncService{
				syncCPMStageTagFn: tt.syncCPMFn,
			}
			svc := newLstepLifecycleSvc(
				defaultLstepSettingsSvc(),
				&mockOwnerRepository{},
				petRepo,
				&mockLstepTagCacheRepository{},
				syncSvc,
			)

			err := svc.HandlePetDeath(context.Background(), 1, 1, time.Now(), "病気", nil)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrIs != nil {
					assert.True(t, errors.Is(err, tt.wantErrIs))
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}

// BUG-407: HandlePetDeath は deceased_at/reason に加えて status も
// 同一 Update で "deceased" に書き換える。status が deceased_at と
// 独立した二重管理フィールドのままだと、外側フォームの生死ラジオが
// 追従せず、次に外側「更新」を押した際に status="alive" で上書きされ
// deceased_at だけが残る不整合状態を再現してしまう。
func TestLstepLifecycleService_HandlePetDeath_SetsStatusDeceased(t *testing.T) {
	var capturedFields map[string]any
	petRepo := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			return &model.Pet{ID: 1, OwnerID: 10}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, fields map[string]any) error {
			capturedFields = fields
			return nil
		},
		findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) {
			return []model.Pet{{ID: 2, OwnerID: 10}}, nil
		},
	}
	svc := newLstepLifecycleSvc(
		defaultLstepSettingsSvc(),
		&mockOwnerRepository{},
		petRepo,
		&mockLstepTagCacheRepository{},
		&mockLstepTagSyncService{},
	)

	err := svc.HandlePetDeath(context.Background(), 1, 1, time.Now(), "老衰", nil)

	assert.NoError(t, err)
	assert.Equal(t, model.PetStatusDeceased, capturedFields["status"])
}

// ---- テスト: HandlePetRevival ----

func TestLstepLifecycleService_HandlePetRevival(t *testing.T) {
	tests := []struct {
		name        string
		petFindFn   func(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
		petUpdateFn func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
		wantErr     bool
		wantErrIs   error
	}{
		{
			name: "正常: 死亡取り消し",
			petFindFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return &model.Pet{ID: 1, OwnerID: 10}, nil
			},
			petUpdateFn: func(_ context.Context, _, _ uint64, fields map[string]any) error {
				assert.Nil(t, fields["deceased_at"])
				assert.Nil(t, fields["deceased_reason"])
				// BUG-407: 死亡取り消し時も status を "alive" に戻し、
				// deceased_at と status の二重管理不整合を防ぐ。
				assert.Equal(t, model.PetStatusAlive, fields["status"])
				return nil
			},
		},
		{
			name: "エラー: ペットが見つからない → ErrNotFound",
			petFindFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return nil, apperrors.WrapNotFound("pet", "1")
			},
			wantErr:   true,
			wantErrIs: apperrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			petRepo := &mockPetRepository{
				findByIDFn: tt.petFindFn,
				updateFn:   tt.petUpdateFn,
			}
			svc := newLstepLifecycleSvc(
				defaultLstepSettingsSvc(),
				&mockOwnerRepository{},
				petRepo,
				&mockLstepTagCacheRepository{},
				&mockLstepTagSyncService{},
			)

			err := svc.HandlePetRevival(context.Background(), 1, 1, nil)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrIs != nil {
					assert.True(t, errors.Is(err, tt.wantErrIs))
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}

// ---- テスト: HandleOwnerOptOut ----

func TestLstepLifecycleService_HandleOwnerOptOut(t *testing.T) {
	lineUserID := "Uabc123"

	tests := []struct {
		name           string
		ownerFindFn    func(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
		ownerUpdateFn  func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
		tagCacheFindFn func(ctx context.Context, clinicID, ownerID uint64) ([]*model.LstepTagCache, error)
		wantErr        bool
		wantErrIs      error
	}{
		{
			name: "正常: LINE未連携 → タグ解除スキップ",
			ownerFindFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return &model.Owner{ID: 10}, nil // LineUserID nil
			},
			ownerUpdateFn: func(_ context.Context, _, _ uint64, fields map[string]any) error {
				assert.Equal(t, true, fields["lstep_opt_out"])
				return nil
			},
		},
		{
			name: "正常: LINE連携あり → タグ解除実行（best-effort）",
			ownerFindFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return &model.Owner{ID: 10, LineUserID: &lineUserID}, nil
			},
			ownerUpdateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				return nil
			},
			tagCacheFindFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
				return []*model.LstepTagCache{{TagName: "cpm_new"}}, nil
			},
		},
		{
			name: "エラー: オーナーが見つからない → ErrNotFound",
			ownerFindFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return nil, apperrors.WrapNotFound("owner", "10")
			},
			wantErr:   true,
			wantErrIs: apperrors.ErrNotFound,
		},
		{
			name: "エラー: オーナー更新失敗",
			ownerFindFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return &model.Owner{ID: 10}, nil
			},
			ownerUpdateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				return errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ownerRepo := &mockOwnerRepository{
				findByIDFn: tt.ownerFindFn,
				updateFn: func(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
					if tt.ownerUpdateFn != nil {
						return tt.ownerUpdateFn(ctx, clinicID, id, fields)
					}
					return nil
				},
			}
			tagCacheRepo := &mockLstepTagCacheRepository{
				findByOwnerFn: tt.tagCacheFindFn,
			}
			svc := newLstepLifecycleSvc(
				defaultLstepSettingsSvc(),
				ownerRepo,
				&mockPetRepository{},
				tagCacheRepo,
				&mockLstepTagSyncService{},
			)

			err := svc.HandleOwnerOptOut(context.Background(), 1, 10, "希望なし")
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrIs != nil {
					assert.True(t, errors.Is(err, tt.wantErrIs))
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}

// ---- テスト: HandleOwnerOptIn ----

func TestLstepLifecycleService_HandleOwnerOptIn(t *testing.T) {
	tests := []struct {
		name          string
		ownerFindFn   func(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
		ownerUpdateFn func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
		wantErr       bool
		wantErrIs     error
	}{
		{
			name: "正常: オプトイン → opt_out フラグ解除",
			ownerFindFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return &model.Owner{ID: 10, LstepOptOut: true}, nil
			},
			ownerUpdateFn: func(_ context.Context, _, _ uint64, fields map[string]any) error {
				assert.Equal(t, false, fields["lstep_opt_out"])
				assert.Nil(t, fields["lstep_opt_out_at"])
				return nil
			},
		},
		{
			name: "エラー: オーナーが見つからない → ErrNotFound",
			ownerFindFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return nil, apperrors.WrapNotFound("owner", "10")
			},
			wantErr:   true,
			wantErrIs: apperrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ownerRepo := &mockOwnerRepository{
				findByIDFn: tt.ownerFindFn,
				updateFn: func(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
					if tt.ownerUpdateFn != nil {
						return tt.ownerUpdateFn(ctx, clinicID, id, fields)
					}
					return nil
				},
			}
			svc := newLstepLifecycleSvc(
				defaultLstepSettingsSvc(),
				ownerRepo,
				&mockPetRepository{},
				&mockLstepTagCacheRepository{},
				&mockLstepTagSyncService{},
			)

			err := svc.HandleOwnerOptIn(context.Background(), 1, 10)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrIs != nil {
					assert.True(t, errors.Is(err, tt.wantErrIs))
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}

// ---- テスト: HandleOwnerDeletion ----

func TestLstepLifecycleService_HandleOwnerDeletion(t *testing.T) {
	lineUserID := "Uabc123"

	tests := []struct {
		name        string
		ownerFindFn func(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
		wantErr     bool
		wantErrIs   error
	}{
		{
			name: "正常: LINE未連携 → タグ解除スキップ",
			ownerFindFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return &model.Owner{ID: 10}, nil
			},
		},
		{
			name: "正常: LINE連携あり → タグ解除（best-effort）",
			ownerFindFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return &model.Owner{ID: 10, LineUserID: &lineUserID}, nil
			},
		},
		{
			name: "エラー: オーナーが見つからない → ErrNotFound",
			ownerFindFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return nil, apperrors.WrapNotFound("owner", "10")
			},
			wantErr:   true,
			wantErrIs: apperrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ownerRepo := &mockOwnerRepository{
				findByIDFn: tt.ownerFindFn,
			}
			svc := newLstepLifecycleSvc(
				defaultLstepSettingsSvc(),
				ownerRepo,
				&mockPetRepository{},
				&mockLstepTagCacheRepository{},
				&mockLstepTagSyncService{},
			)

			err := svc.HandleOwnerDeletion(context.Background(), 1, 10)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrIs != nil {
					assert.True(t, errors.Is(err, tt.wantErrIs))
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}

// ---- テスト: CalculateCPMStage ----

func TestCalculateCPMStage(t *testing.T) {
	tests := []struct {
		name  string
		data  CPMData
		stage CPMStage
	}{
		{"来院なし → dormant", CPMData{
			TotalVisitCount:      0,
			AnnualVisitCount:     0,
			DaysSinceVisit:       -1,
			LTVAmount:            0,
			FirstVisitDaysSince:  0,
			MaxSingleVisitAmount: 0,
		}, CPMStageDormant},
		{"240日以上 → dormant", CPMData{
			TotalVisitCount:      4,
			AnnualVisitCount:     0,
			DaysSinceVisit:       240,
			LTVAmount:            40_000,
			FirstVisitDaysSince:  500,
			MaxSingleVisitAmount: 15_000,
		}, CPMStageDormant},
		{"noah条件満たす → noah", CPMData{
			TotalVisitCount:      3,
			AnnualVisitCount:     3,
			DaysSinceVisit:       30,
			LTVAmount:            80_000,
			FirstVisitDaysSince:  365,
			MaxSingleVisitAmount: 0,
		}, CPMStageNoah},
		{"core条件満たす → core", CPMData{
			TotalVisitCount:      2,
			AnnualVisitCount:     2,
			DaysSinceVisit:       30,
			LTVAmount:            50_000,
			FirstVisitDaysSince:  180,
			MaxSingleVisitAmount: 0,
		}, CPMStageCore},
		{"spot条件 高単価・90日超 → spot", CPMData{
			TotalVisitCount:      1,
			AnnualVisitCount:     0,
			DaysSinceVisit:       100,
			LTVAmount:            0,
			FirstVisitDaysSince:  200,
			MaxSingleVisitAmount: 30_000,
		}, CPMStageSpot},
		{"2〜3回来院・90日以内 → growing", CPMData{
			TotalVisitCount:      2,
			AnnualVisitCount:     0,
			DaysSinceVisit:       10,
			LTVAmount:            25_000,
			FirstVisitDaysSince:  45,
			MaxSingleVisitAmount: 0,
		}, CPMStageGrowing},
		// encounter: 来院1回 AND LTV 20,000円未満（仕様書 §3 明示判定）
		{"初回来院 LTV=0 → encounter", CPMData{
			TotalVisitCount:      1,
			AnnualVisitCount:     0,
			DaysSinceVisit:       5,
			LTVAmount:            0,
			FirstVisitDaysSince:  5,
			MaxSingleVisitAmount: 0,
		}, CPMStageEncounter},
		{"来院1回 LTV=19999 → encounter", CPMData{
			TotalVisitCount:      1,
			AnnualVisitCount:     1,
			DaysSinceVisit:       10,
			LTVAmount:            19_999,
			FirstVisitDaysSince:  10,
			MaxSingleVisitAmount: 19_999,
		}, CPMStageEncounter},
		// unclassified: 全6ステージのいずれにも該当しない異常データ
		{"来院1回 LTV>=20k → unclassified", CPMData{
			TotalVisitCount:      1,
			AnnualVisitCount:     1,
			DaysSinceVisit:       10,
			LTVAmount:            20_000,
			FirstVisitDaysSince:  10,
			MaxSingleVisitAmount: 20_000,
		}, CPMStageUnclassified},
		{"累計4回・全条件未達 → unclassified", CPMData{
			TotalVisitCount:      4,
			AnnualVisitCount:     2,
			DaysSinceVisit:       10,
			LTVAmount:            15_000,
			FirstVisitDaysSince:  120,
			MaxSingleVisitAmount: 0,
		}, CPMStageUnclassified},
		{"来院0回・非dormant → unclassified", CPMData{
			TotalVisitCount:      0,
			AnnualVisitCount:     0,
			DaysSinceVisit:       50,
			LTVAmount:            100_000,
			FirstVisitDaysSince:  50,
			MaxSingleVisitAmount: 0,
		}, CPMStageUnclassified},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateCPMStage(tt.data)
			assert.Equal(t, tt.stage, got)
		})
	}
}

// ---- テスト: ltvBracketTag / visitCountAnnualTag ----

func TestLtvBracketTag(t *testing.T) {
	assert.Equal(t, "ltv_amount_8", ltvBracketTag(80_000))
	assert.Equal(t, "ltv_amount_8", ltvBracketTag(200_000))
	assert.Equal(t, "ltv_amount_5", ltvBracketTag(50_000))
	assert.Equal(t, "ltv_amount_5", ltvBracketTag(79_999))
	assert.Equal(t, "ltv_amount_2", ltvBracketTag(20_000))
	assert.Equal(t, "ltv_amount_2", ltvBracketTag(49_999))
	assert.Equal(t, "ltv_amount_0", ltvBracketTag(0))
	assert.Equal(t, "ltv_amount_0", ltvBracketTag(19_999))
}

func TestVisitCountAnnualTag(t *testing.T) {
	assert.Equal(t, "visit_count_annual_10", visitCountAnnualTag(10))
	assert.Equal(t, "visit_count_annual_10", visitCountAnnualTag(15))
	assert.Equal(t, "visit_count_annual_5", visitCountAnnualTag(5))
	assert.Equal(t, "visit_count_annual_5", visitCountAnnualTag(9))
	assert.Equal(t, "visit_count_annual_3", visitCountAnnualTag(3))
	assert.Equal(t, "visit_count_annual_3", visitCountAnnualTag(4))
	assert.Equal(t, "visit_count_annual_2", visitCountAnnualTag(2))
	assert.Equal(t, "visit_count_annual_1", visitCountAnnualTag(0))
	assert.Equal(t, "visit_count_annual_1", visitCountAnnualTag(1))
}

// ---- テスト: vaccineTagNames / isRabiesVaccine ----

func TestVaccineTagNames(t *testing.T) {
	dogSpecies := model.VaccineSpeciesDog
	catSpecies := model.VaccineSpeciesCat
	bothSpecies := model.VaccineSpeciesBoth

	baseDate := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)

	t.Run("dog species → vaccine_dog_date tag", func(t *testing.T) {
		vac := &model.Vaccination{
			Date: baseDate,
			Vaccine: &model.Vaccine{
				Name:    "混合ワクチン",
				Species: &dogSpecies,
			},
		}
		tags := vaccineTagNames(vac)
		assert.Equal(t, []string{"vaccine_dog_2026-04-25"}, tags)
	})

	t.Run("cat species → vaccine_cat_date tag", func(t *testing.T) {
		vac := &model.Vaccination{
			Date: baseDate,
			Vaccine: &model.Vaccine{
				Name:    "猫ワクチン",
				Species: &catSpecies,
			},
		}
		tags := vaccineTagNames(vac)
		assert.Equal(t, []string{"vaccine_cat_2026-04-25"}, tags)
	})

	t.Run("both species → dog + cat tags", func(t *testing.T) {
		vac := &model.Vaccination{
			Date: baseDate,
			Vaccine: &model.Vaccine{
				Name:    "共通ワクチン",
				Species: &bothSpecies,
			},
		}
		tags := vaccineTagNames(vac)
		assert.Equal(t, []string{"vaccine_dog_2026-04-25", "vaccine_cat_2026-04-25"}, tags)
	})

	t.Run("rabies vaccine → rabies tag added", func(t *testing.T) {
		vac := &model.Vaccination{
			Date: baseDate,
			Vaccine: &model.Vaccine{
				Name:    "狂犬病ワクチン",
				Species: &dogSpecies,
			},
		}
		tags := vaccineTagNames(vac)
		assert.Contains(t, tags, "vaccine_dog_2026-04-25")
		assert.Contains(t, tags, "vaccine_rabies_2026-04-25")
	})

	t.Run("nil vaccine → nil tags", func(t *testing.T) {
		tags := vaccineTagNames(&model.Vaccination{Date: baseDate})
		assert.Nil(t, tags)
	})
}

func TestIsRabiesVaccine(t *testing.T) {
	assert.True(t, isRabiesVaccine("rabies"))
	assert.True(t, isRabiesVaccine("RABIES vaccine"))
	assert.True(t, isRabiesVaccine("狂犬病予防ワクチン"))
	assert.False(t, isRabiesVaccine("混合ワクチン"))
	assert.False(t, isRabiesVaccine("フィラリア"))
}

// ---- テスト: loadPetDerivedPrefixes ----

func TestLstepLifecycleService_LoadPetDerivedPrefixes(t *testing.T) {
	t.Run("tagConfigRepo nil → fallback prefixes", func(t *testing.T) {
		svc := &lstepLifecycleService{tagConfigRepo: nil}
		prefixes := svc.loadPetDerivedPrefixes(context.Background())
		assert.Equal(t, []string{"vaccine_", tagPrefixCheckupDone}, prefixes)
	})

	t.Run("DB returns C2 prefixes → use DB values", func(t *testing.T) {
		repo := &mockLstepTagConfigRepository{
			findAllAutoManagedPrefixesFn: func(_ context.Context) ([]*model.LstepAutoManagedPrefix, error) {
				return []*model.LstepAutoManagedPrefix{
					{Category: "C2", Prefix: "vaccine_"},
					{Category: "C2", Prefix: "checkup_done_"},
					{Category: "B", Prefix: "next_visit_"},
				}, nil
			},
		}
		svc := &lstepLifecycleService{tagConfigRepo: repo}
		prefixes := svc.loadPetDerivedPrefixes(context.Background())
		assert.ElementsMatch(t, []string{"vaccine_", "checkup_done_"}, prefixes)
	})

	t.Run("DB returns no C2 prefixes → fallback", func(t *testing.T) {
		repo := &mockLstepTagConfigRepository{
			findAllAutoManagedPrefixesFn: func(_ context.Context) ([]*model.LstepAutoManagedPrefix, error) {
				return []*model.LstepAutoManagedPrefix{
					{Category: "B", Prefix: "next_visit_"},
				}, nil
			},
		}
		svc := &lstepLifecycleService{tagConfigRepo: repo}
		prefixes := svc.loadPetDerivedPrefixes(context.Background())
		assert.Equal(t, []string{"vaccine_", tagPrefixCheckupDone}, prefixes)
	})

	t.Run("DB error → fallback", func(t *testing.T) {
		repo := &mockLstepTagConfigRepository{
			findAllAutoManagedPrefixesFn: func(_ context.Context) ([]*model.LstepAutoManagedPrefix, error) {
				return nil, errors.New("db error")
			},
		}
		svc := &lstepLifecycleService{tagConfigRepo: repo}
		prefixes := svc.loadPetDerivedPrefixes(context.Background())
		assert.Equal(t, []string{"vaccine_", tagPrefixCheckupDone}, prefixes)
	})
}

// ---- ビルドコンパイル確認: LstepTagCacheRepository インターフェース実装 ----

var _ repository.LstepTagCacheRepository = (*mockLstepTagCacheRepository)(nil)
var _ LstepSettingsService = (*mockLstepSettingsService)(nil)
var _ LstepTagSyncService = (*mockLstepTagSyncService)(nil)

// mockLstepClientForLifecycle is a lstep.Client mock local to this file so that
// removePetDerivedTagsFromLstep can be tested without any real HTTP calls (the client
// is a plain function parameter on that method, so it can be injected directly).
type mockLstepClientForLifecycle struct {
	removeTagFn func(ctx context.Context, lineUserID, tagName string) error
}

func (m *mockLstepClientForLifecycle) AddTag(_ context.Context, _, _ string) error { return nil }
func (m *mockLstepClientForLifecycle) RemoveTag(ctx context.Context, lineUserID, tagName string) error {
	if m.removeTagFn != nil {
		return m.removeTagFn(ctx, lineUserID, tagName)
	}
	return nil
}
func (m *mockLstepClientForLifecycle) GetUserTags(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}
func (m *mockLstepClientForLifecycle) AddTagBulk(_ context.Context, _ []string, _ string) error {
	return nil
}
func (m *mockLstepClientForLifecycle) GetUser(_ context.Context, _ string) (*lstep.UserInfo, error) {
	return nil, nil
}
func (m *mockLstepClientForLifecycle) SetProperty(_ context.Context, _, _, _ string) error {
	return nil
}

var _ lstep.Client = (*mockLstepClientForLifecycle)(nil)

// ---- buildClient ----

func TestLstepLifecycleService_BuildClient(t *testing.T) {
	tests := []struct {
		name          string
		isSyncEnabled bool
		syncErr       error
		apiKey        string
		credErr       error
		wantErr       bool
		wantNilClient bool
	}{
		{
			name:          "sync disabled -> nil client, nil error",
			isSyncEnabled: false,
			wantNilClient: true,
		},
		{
			name:          "IsSyncEnabled error -> wrapped error",
			isSyncEnabled: false,
			syncErr:       errors.New("db error"),
			wantErr:       true,
			wantNilClient: true,
		},
		{
			name:          "GetRawCredentials error -> wrapped error",
			isSyncEnabled: true,
			credErr:       errors.New("db error"),
			wantErr:       true,
			wantNilClient: true,
		},
		{
			name:          "empty api key -> nil client, nil error",
			isSyncEnabled: true,
			apiKey:        "",
			wantNilClient: true,
		},
		{
			name:          "valid credentials -> non-nil client",
			isSyncEnabled: true,
			apiKey:        "secret-key",
			wantNilClient: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settingsSvc := &mockLstepSettingsService{
				isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) {
					return tt.isSyncEnabled, tt.syncErr
				},
				getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
					return tt.apiKey, "https://example.com", "", tt.credErr
				},
			}
			svc := &lstepLifecycleService{settingsSvc: settingsSvc}

			client, err := svc.buildClient(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.wantNilClient {
				assert.Nil(t, client)
			} else {
				assert.NotNil(t, client)
			}
		})
	}
}

// ---- removeAllTagsFromLstep ----

func TestLstepLifecycleService_RemoveAllTagsFromLstep(t *testing.T) {
	tests := []struct {
		name         string
		settingsSvc  *mockLstepSettingsService
		tagCacheRepo *mockLstepTagCacheRepository
		wantErr      bool
	}{
		{
			name: "buildClient error propagates",
			settingsSvc: &mockLstepSettingsService{
				isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) {
					return false, errors.New("settings db error")
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{},
			wantErr:      true,
		},
		{
			name:        "tag cache FindByOwner error is wrapped",
			settingsSvc: defaultLstepSettingsSvc(),
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return nil, errors.New("db error")
				},
			},
			wantErr: true,
		},
		{
			name:        "client nil (sync disabled) deletes cache without calling RemoveTag",
			settingsSvc: defaultLstepSettingsSvc(),
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{{TagName: "cpm_new"}}, nil
				},
			},
			wantErr: false,
		},
		{
			name:        "DeleteAllByOwner error is wrapped",
			settingsSvc: defaultLstepSettingsSvc(),
			tagCacheRepo: &mockLstepTagCacheRepository{
				deleteAllByOwnerFn: func(_ context.Context, _, _ uint64) error {
					return errors.New("db error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &lstepLifecycleService{
				settingsSvc:  tt.settingsSvc,
				tagCacheRepo: tt.tagCacheRepo,
			}

			err := svc.removeAllTagsFromLstep(context.Background(), 1, 10, "Uabc123")

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- removePetDerivedTagsFromLstep ----

func TestLstepLifecycleService_RemovePetDerivedTagsFromLstep(t *testing.T) {
	tests := []struct {
		name             string
		tagCacheRepo     *mockLstepTagCacheRepository
		removeTagFn      func(ctx context.Context, lineUserID, tagName string) error
		wantDeleteCalled bool
	}{
		{
			name: "removes matching pet-derived tags (fallback prefixes), skips unrelated tags",
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{
						{TagName: "vaccine_dog_2026-01-01"},
						{TagName: "cpm_new"},
					}, nil
				},
			},
			wantDeleteCalled: true,
		},
		{
			name: "FindByOwner error returns early without panic",
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return nil, errors.New("db error")
				},
			},
		},
		{
			name: "RemoveTag error is best-effort, does not panic",
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{{TagName: "checkup_done_1"}}, nil
				},
			},
			removeTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") },
		},
		{
			name: "DeleteTag error is logged, does not panic",
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{{TagName: "vaccine_cat_2026-01-01"}}, nil
				},
				deleteTagFn: func(_ context.Context, _, _ uint64, _ string) error {
					return errors.New("db error")
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleteCalled := false
			if tt.tagCacheRepo.deleteTagFn == nil {
				tt.tagCacheRepo.deleteTagFn = func(_ context.Context, _, _ uint64, _ string) error {
					deleteCalled = true
					return nil
				}
			}
			svc := &lstepLifecycleService{tagCacheRepo: tt.tagCacheRepo}
			client := &mockLstepClientForLifecycle{removeTagFn: tt.removeTagFn}

			assert.NotPanics(t, func() {
				svc.removePetDerivedTagsFromLstep(context.Background(), client, 1, 10, "Uabc123")
			})
			if tt.wantDeleteCalled {
				assert.True(t, deleteCalled)
			}
		})
	}
}

// ---- HandlePetDeath: additional branches ----

func TestLstepLifecycleService_HandlePetDeath_FindLivingByOwnerError(t *testing.T) {
	petRepo := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			return &model.Pet{ID: 1, OwnerID: 10}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error { return nil },
		findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) {
			return nil, errors.New("db error")
		},
	}
	svc := newLstepLifecycleSvc(defaultLstepSettingsSvc(), &mockOwnerRepository{}, petRepo, &mockLstepTagCacheRepository{}, &mockLstepTagSyncService{})

	err := svc.HandlePetDeath(context.Background(), 1, 1, time.Now(), "病気", nil)

	assert.Error(t, err)
}

func TestLstepLifecycleService_HandlePetDeath_AllPetsDead_OwnerFindError(t *testing.T) {
	petRepo := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			return &model.Pet{ID: 1, OwnerID: 10}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error { return nil },
		findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) {
			return []model.Pet{}, nil
		},
	}
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return nil, errors.New("db error")
		},
	}
	svc := newLstepLifecycleSvc(defaultLstepSettingsSvc(), ownerRepo, petRepo, &mockLstepTagCacheRepository{}, &mockLstepTagSyncService{})

	// owner lookup failure on the all-pets-dead path is best-effort: HandlePetDeath must
	// still return nil (the death was already recorded successfully).
	err := svc.HandlePetDeath(context.Background(), 1, 1, time.Now(), "病気", nil)

	assert.NoError(t, err)
}

func TestLstepLifecycleService_HandlePetDeath_AllPetsDead_LineLinkedRemovesTags(t *testing.T) {
	lineUserID := "Uabc123"
	deleteAllCalled := false
	petRepo := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			return &model.Pet{ID: 1, OwnerID: 10}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error { return nil },
		findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) {
			return []model.Pet{}, nil
		},
	}
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 10, LineUserID: &lineUserID}, nil
		},
	}
	tagCacheRepo := &mockLstepTagCacheRepository{
		findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
			return []*model.LstepTagCache{{TagName: "cpm_new"}}, nil
		},
		deleteAllByOwnerFn: func(_ context.Context, _, _ uint64) error {
			deleteAllCalled = true
			return nil
		},
	}
	svc := newLstepLifecycleSvc(defaultLstepSettingsSvc(), ownerRepo, petRepo, tagCacheRepo, &mockLstepTagSyncService{})

	err := svc.HandlePetDeath(context.Background(), 1, 1, time.Now(), "病気", nil)

	assert.NoError(t, err)
	assert.True(t, deleteAllCalled)
}

func TestLstepLifecycleService_HandlePetDeath_SurvivingPets_PetDerivedCleanupBestEffort(t *testing.T) {
	lineUserID := "Uabc123"
	petRepo := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			return &model.Pet{ID: 1, OwnerID: 10}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error { return nil },
		findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) {
			return []model.Pet{{ID: 2, OwnerID: 10}}, nil
		},
	}
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 10, LineUserID: &lineUserID}, nil
		},
	}
	// Sync enabled with a real api key: buildClient returns a non-nil client, exercising
	// the "cleanup path builds a client" branch of HandlePetDeath. tagCacheRepo returns no
	// tags, so removePetDerivedTagsFromLstep never actually calls RemoveTag (no network hit).
	settingsSvc := &mockLstepSettingsService{
		isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
		getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
			return "api-key", "https://example.com", "", nil
		},
	}
	svc := newLstepLifecycleSvc(settingsSvc, ownerRepo, petRepo, &mockLstepTagCacheRepository{}, &mockLstepTagSyncService{})

	err := svc.HandlePetDeath(context.Background(), 1, 1, time.Now(), "病気", nil)

	assert.NoError(t, err)
}

func TestLstepLifecycleService_HandlePetDeath_AuditLogFailureIsBestEffort(t *testing.T) {
	petRepo := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			return &model.Pet{ID: 1, OwnerID: 10}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error { return nil },
		findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) {
			return []model.Pet{{ID: 2, OwnerID: 10}}, nil
		},
	}
	svc := NewLstepLifecycleService(
		defaultLstepSettingsSvc(),
		&mockOwnerRepository{},
		petRepo,
		&mockLstepTagCacheRepository{},
		&mockLstepTagSyncService{},
		&mockAuditService{logLstepOperationErr: errors.New("audit db error")},
		nil,
	)

	err := svc.HandlePetDeath(context.Background(), 1, 1, time.Now(), "病気", nil)

	assert.NoError(t, err)
}

// BUG-407 follow-up (healthcare-reviewer HIGH finding): HandlePetDeath must record the
// user-driven death event itself in the audit trail (action="pet_death") with the real
// actor, on EVERY code path — including the all-pets-dead early-return branch, which
// previously skipped audit entirely (only the best-effort "pet_death_tag_sync" fired, and
// only on the living-pets branch).
func TestLstepLifecycleService_HandlePetDeath_LogsUserActionAudit(t *testing.T) {
	tests := []struct {
		name                string
		findLivingByOwnerFn func(ctx context.Context, clinicID, ownerID uint64) ([]model.Pet, error)
	}{
		{
			name: "living pets remain",
			findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) {
				return []model.Pet{{ID: 2, OwnerID: 10}}, nil
			},
		},
		{
			name: "all pets dead (early return branch)",
			findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) {
				return []model.Pet{}, nil
			},
		},
	}

	type capturedCall struct {
		actorID    *uint64
		action     string
		resource   string
		resourceID *uint64
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actor := uint64(42)
			var calls []capturedCall

			petRepo := &mockPetRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
					return &model.Pet{ID: 1, OwnerID: 10}, nil
				},
				updateFn:            func(_ context.Context, _, _ uint64, _ map[string]any) error { return nil },
				findLivingByOwnerFn: tt.findLivingByOwnerFn,
			}
			auditSvc := &mockAuditService{
				logLstepOperationFn: func(_ context.Context, _ uint64, actorID *uint64, action, resource string, resourceID *uint64) error {
					calls = append(calls, capturedCall{actorID: actorID, action: action, resource: resource, resourceID: resourceID})
					return nil
				},
			}
			svc := NewLstepLifecycleService(
				defaultLstepSettingsSvc(),
				&mockOwnerRepository{},
				petRepo,
				&mockLstepTagCacheRepository{},
				&mockLstepTagSyncService{},
				auditSvc,
				nil,
			)

			err := svc.HandlePetDeath(context.Background(), 1, 1, time.Now(), "病気", &actor)
			assert.NoError(t, err)

			var found bool
			for _, c := range calls {
				if c.action != "pet_death" {
					continue
				}
				found = true
				assert.Equal(t, "pet", c.resource)
				if assert.NotNil(t, c.actorID) {
					assert.Equal(t, actor, *c.actorID)
				}
				if assert.NotNil(t, c.resourceID) {
					assert.Equal(t, uint64(1), *c.resourceID)
				}
			}
			assert.True(t, found, "expected a pet_death audit log call with the real actor")
		})
	}
}

// ---- HandlePetRevival: additional branches ----

func TestLstepLifecycleService_HandlePetRevival_UpdateError(t *testing.T) {
	petRepo := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			return &model.Pet{ID: 1, OwnerID: 10}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			return errors.New("db error")
		},
	}
	svc := newLstepLifecycleSvc(defaultLstepSettingsSvc(), &mockOwnerRepository{}, petRepo, &mockLstepTagCacheRepository{}, &mockLstepTagSyncService{})

	err := svc.HandlePetRevival(context.Background(), 1, 1, nil)

	assert.Error(t, err)
}

func TestLstepLifecycleService_HandlePetRevival_SyncErrorsAreBestEffort(t *testing.T) {
	petRepo := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			return &model.Pet{ID: 1, OwnerID: 10}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error { return nil },
	}
	syncSvc := &mockLstepTagSyncService{
		syncOwnerAnimalClassificationTagFn: func(_ context.Context, _, _ uint64) error {
			return errors.New("sync error")
		},
		syncPetBasicInfoTagsFn: func(_ context.Context, _, _ uint64) error {
			return errors.New("sync error")
		},
		syncCPMStageTagFn: func(_ context.Context, _, _ uint64) error {
			return errors.New("sync error")
		},
	}
	svc := NewLstepLifecycleService(
		defaultLstepSettingsSvc(),
		&mockOwnerRepository{},
		petRepo,
		&mockLstepTagCacheRepository{},
		syncSvc,
		&mockAuditService{logLstepOperationErr: errors.New("audit db error")},
		nil,
	)

	err := svc.HandlePetRevival(context.Background(), 1, 1, nil)

	assert.NoError(t, err)
}

// BUG-407 follow-up (healthcare-reviewer HIGH finding): HandlePetRevival must record the
// user-driven revival event itself in the audit trail (action="pet_revival") with the
// real actor (symmetric with HandlePetDeath's new "pet_death" audit).
func TestLstepLifecycleService_HandlePetRevival_LogsUserActionAudit(t *testing.T) {
	type capturedCall struct {
		actorID    *uint64
		action     string
		resource   string
		resourceID *uint64
	}
	actor := uint64(77)
	var calls []capturedCall

	petRepo := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			return &model.Pet{ID: 1, OwnerID: 10}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error { return nil },
	}
	auditSvc := &mockAuditService{
		logLstepOperationFn: func(_ context.Context, _ uint64, actorID *uint64, action, resource string, resourceID *uint64) error {
			calls = append(calls, capturedCall{actorID: actorID, action: action, resource: resource, resourceID: resourceID})
			return nil
		},
	}
	svc := NewLstepLifecycleService(
		defaultLstepSettingsSvc(),
		&mockOwnerRepository{},
		petRepo,
		&mockLstepTagCacheRepository{},
		&mockLstepTagSyncService{},
		auditSvc,
		nil,
	)

	err := svc.HandlePetRevival(context.Background(), 1, 1, &actor)
	assert.NoError(t, err)

	var found bool
	for _, c := range calls {
		if c.action != "pet_revival" {
			continue
		}
		found = true
		assert.Equal(t, "pet", c.resource)
		if assert.NotNil(t, c.actorID) {
			assert.Equal(t, actor, *c.actorID)
		}
		if assert.NotNil(t, c.resourceID) {
			assert.Equal(t, uint64(1), *c.resourceID)
		}
	}
	assert.True(t, found, "expected a pet_revival audit log call with the real actor")
}

// ---- HandleOwnerOptIn: additional branches ----

func TestLstepLifecycleService_HandleOwnerOptIn_UpdateError(t *testing.T) {
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 10, LstepOptOut: true}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			return errors.New("db error")
		},
	}
	svc := newLstepLifecycleSvc(defaultLstepSettingsSvc(), ownerRepo, &mockPetRepository{}, &mockLstepTagCacheRepository{}, &mockLstepTagSyncService{})

	err := svc.HandleOwnerOptIn(context.Background(), 1, 10)

	assert.Error(t, err)
}

func TestLstepLifecycleService_HandleOwnerOptIn_SyncErrorIsBestEffort(t *testing.T) {
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 10, LstepOptOut: true}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error { return nil },
	}
	syncSvc := &mockLstepTagSyncService{
		syncCPMStageTagFn: func(_ context.Context, _, _ uint64) error {
			return errors.New("sync error")
		},
	}
	svc := newLstepLifecycleSvc(defaultLstepSettingsSvc(), ownerRepo, &mockPetRepository{}, &mockLstepTagCacheRepository{}, syncSvc)

	err := svc.HandleOwnerOptIn(context.Background(), 1, 10)

	assert.NoError(t, err)
}

// ---- HandleOwnerDeletion: additional branches ----

func TestLstepLifecycleService_HandleOwnerDeletion_RemoveTagsErrorIsBestEffort(t *testing.T) {
	lineUserID := "Uabc123"
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 10, LineUserID: &lineUserID}, nil
		},
	}
	tagCacheRepo := &mockLstepTagCacheRepository{
		findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
			return nil, errors.New("db error")
		},
	}
	svc := newLstepLifecycleSvc(defaultLstepSettingsSvc(), ownerRepo, &mockPetRepository{}, tagCacheRepo, &mockLstepTagSyncService{})

	// removeAllTagsFromLstep failing must not abort the owner-deletion flow.
	err := svc.HandleOwnerDeletion(context.Background(), 1, 10)

	assert.NoError(t, err)
}
