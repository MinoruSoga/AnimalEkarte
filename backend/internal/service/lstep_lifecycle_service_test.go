package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- LstepSettingsService モック ----

type mockLstepSettingsService struct {
	getRawCredentialsFn func(ctx context.Context, clinicID uint64) (apiKey, baseURL, lineToken string, err error)
	getSettingsFn       func(ctx context.Context, clinicID uint64) (*LstepSettingsResponse, error)
	updateSettingsFn    func(ctx context.Context, clinicID uint64, input *UpdateLstepSettingsInput) (*LstepSettingsResponse, error)
	deleteSettingsFn    func(ctx context.Context, clinicID uint64) error
	testConnectionFn    func(ctx context.Context, clinicID uint64) (*LstepConnectionTestResult, error)
	isSyncEnabledFn     func(ctx context.Context, clinicID uint64) (bool, error)
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
func (m *mockLstepSettingsService) GetDormantThresholds(_ context.Context, _ uint64) (model.DormantThresholds, error) {
	return model.DormantThresholds{}.WithDefaults(), nil
}

// ---- LstepTagCacheRepository モック ----

type mockLstepTagCacheRepository struct {
	upsertTagFn        func(ctx context.Context, clinicID, ownerID uint64, tagName, category, reason string) error
	deleteTagFn        func(ctx context.Context, clinicID, ownerID uint64, tagName string) error
	deleteAllByOwnerFn func(ctx context.Context, clinicID, ownerID uint64) error
	findByOwnerFn      func(ctx context.Context, clinicID, ownerID uint64) ([]*model.LstepTagCache, error)
	countByTagFn       func(ctx context.Context, clinicID uint64, tagName string) (int64, error)
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
func (m *mockLstepTagCacheRepository) CountByTag(ctx context.Context, clinicID uint64, tagName string) (int64, error) {
	if m.countByTagFn != nil {
		return m.countByTagFn(ctx, clinicID, tagName)
	}
	return 0, nil
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
	syncVaccineTagFn          func(ctx context.Context, clinicID, ownerID, vaccinationID uint64) error
	syncVisitCompletionTagsFn func(ctx context.Context, clinicID, ownerID uint64) error
	syncCPMStageTagFn         func(ctx context.Context, clinicID, ownerID uint64) error
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
func (m *mockLstepTagSyncService) SyncOwnerAnimalClassificationTags(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *mockLstepTagSyncService) SyncPetBasicInfoTags(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *mockLstepTagSyncService) SyncNextVisitTag(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *mockLstepTagSyncService) SyncReservationTag(_ context.Context, _, _ uint64, _ time.Time) error {
	return nil
}
func (m *mockLstepTagSyncService) SyncCancellationTag(_ context.Context, _, _ uint64, _ time.Time) error {
	return nil
}
func (m *mockLstepTagSyncService) SyncCheckupTag(_ context.Context, _, _, _ uint64, _ time.Time, _ *time.Time) error {
	return nil
}
func (m *mockLstepTagSyncService) SyncPrescriptionTag(_ context.Context, _, _ uint64) error {
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

func (m *mockLstepTagSyncService) SyncNoShowTag(_ context.Context, _, _ uint64, _ time.Time) error {
	return nil
}

func (m *mockLstepTagSyncService) SyncDormantTags(_ context.Context, _, _ uint64, _ int) error {
	return nil
}

func (m *mockLstepTagSyncService) ResyncOwnerVaccineTags(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockLstepTagSyncService) ResyncOwnerCheckupTags(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockLstepTagSyncService) ResyncOwnerReservationTags(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockLstepTagSyncService) SyncCPMStageTagV2(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockLstepTagSyncService) SyncLTVTopPercent(_ context.Context, _ uint64) (int, []error) {
	return 0, nil
}

func (m *mockLstepTagSyncService) SyncVisitDormantTags(_ context.Context, _, _ uint64, _ int) error {
	return nil
}

func (m *mockLstepTagSyncService) SyncPetSpeciesTags(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockLstepTagSyncService) SyncSeniorTag(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockLstepTagSyncService) SyncExclusionTags(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockLstepTagSyncService) SyncHealthcheckTags(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockLstepTagSyncService) SyncAnnual4CheckupTag(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockLstepTagSyncService) SyncVaccineDeadlineTag(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockLstepTagSyncService) SyncFilariaTag(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockLstepTagSyncService) SyncFleaTickTag(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockLstepTagSyncService) SyncFoodPurchaseTag(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockLstepTagSyncService) SyncSpecialCheckupCandidateTag(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockLstepTagSyncService) SyncHealthPreventionTagsForClinic(_ context.Context, _ uint64) (int, []error) {
	return 0, nil
}

// ---- AuditService モック ----

type mockAuditService struct{}

func (m *mockAuditService) Log(_ context.Context, _ *model.AuditLog) error { return nil }
func (m *mockAuditService) LogAuthLogin(_ context.Context, _, _ *uint64, _, _, _ string) error {
	return nil
}
func (m *mockAuditService) LogLstepOperation(_ context.Context, _ uint64, _ *uint64, _, _ string, _ *uint64) error {
	return nil
}

func (m *mockAuditService) LogLstepOperationWithMetadata(_ context.Context, _ uint64, _ *uint64, _, _ string, _ *uint64, _ any) error {
	return nil
}
func (m *mockAuditService) LogMedicalRecordChange(_ context.Context, _ uint64, _ *uint64, _ string, _ uint64, _, _ map[string]any) error {
	return nil
}
func (m *mockAuditService) LogVitalChange(_ context.Context, _ uint64, _ *uint64, _ string, _, _ uint64, _, _ map[string]any) error {
	return nil
}
func (m *mockAuditService) LogAddendumCreate(_ context.Context, _ uint64, _ *uint64, _, _ uint64, _ *model.MedicalRecordAddendum) error {
	return nil
}
func (m *mockAuditService) LogClinicSwitch(_ context.Context, _ *uint64, _, _ uint64, _, _ string) error {
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
	return NewLstepLifecycleService(settingsSvc, ownerRepo, petRepo, tagCacheRepo, syncSvc, &mockAuditService{})
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

			err := svc.HandlePetDeath(context.Background(), 1, 1, time.Now(), "病気")
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

			err := svc.HandlePetRevival(context.Background(), 1, 1)
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
		{"来院なし → dormant", CPMData{DaysSinceVisit: -1}, CPMStageDormant},
		{"240日以上 → dormant", CPMData{DaysSinceVisit: 240}, CPMStageDormant},
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

// ---- ビルドコンパイル確認: LstepTagCacheRepository インターフェース実装 ----

var _ repository.LstepTagCacheRepository = (*mockLstepTagCacheRepository)(nil)
var _ LstepSettingsService = (*mockLstepSettingsService)(nil)
var _ LstepTagSyncService = (*mockLstepTagSyncService)(nil)
