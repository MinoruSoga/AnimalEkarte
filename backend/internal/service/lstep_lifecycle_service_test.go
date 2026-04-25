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
	updateSettingsFn    func(ctx context.Context, clinicID uint64, input UpdateLstepSettingsInput) (*LstepSettingsResponse, error)
	deleteSettingsFn    func(ctx context.Context, clinicID uint64) error
	testConnectionFn    func(ctx context.Context, clinicID uint64) (*LstepConnectionTestResult, error)
}

func (m *mockLstepSettingsService) GetRawCredentials(ctx context.Context, clinicID uint64) (string, string, string, error) {
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
func (m *mockLstepSettingsService) UpdateSettings(ctx context.Context, clinicID uint64, input UpdateLstepSettingsInput) (*LstepSettingsResponse, error) {
	if m.updateSettingsFn != nil {
		return m.updateSettingsFn(ctx, clinicID, input)
	}
	return &LstepSettingsResponse{}, nil
}
func (m *mockLstepSettingsService) DeleteSettings(ctx context.Context, clinicID uint64) error {
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

// ---- LstepTagCacheRepository モック ----

type mockLstepTagCacheRepository struct {
	upsertTagFn        func(ctx context.Context, clinicID, ownerID uint64, tagName, category string) error
	deleteTagFn        func(ctx context.Context, clinicID, ownerID uint64, tagName string) error
	deleteAllByOwnerFn func(ctx context.Context, clinicID, ownerID uint64) error
	findByOwnerFn      func(ctx context.Context, clinicID, ownerID uint64) ([]*model.LstepTagCache, error)
	countByTagFn       func(ctx context.Context, clinicID uint64, tagName string) (int64, error)
}

func (m *mockLstepTagCacheRepository) UpsertTag(ctx context.Context, clinicID, ownerID uint64, tagName, category string) error {
	if m.upsertTagFn != nil {
		return m.upsertTagFn(ctx, clinicID, ownerID, tagName, category)
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
func (m *mockLstepTagSyncService) SyncCheckupTag(_ context.Context, _, _ uint64, _ time.Time, _ *time.Time) error {
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

// ---- ヘルパー ----

func newLstepLifecycleSvc(
	settingsSvc *mockLstepSettingsService,
	ownerRepo *mockOwnerRepository,
	petRepo *mockPetRepository,
	tagCacheRepo *mockLstepTagCacheRepository,
	syncSvc *mockLstepTagSyncService,
) LstepLifecycleService {
	return NewLstepLifecycleService(settingsSvc, ownerRepo, petRepo, tagCacheRepo, syncSvc)
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
		name        string
		petFindFn   func(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
		petUpdateFn func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
		syncCPMFn   func(ctx context.Context, clinicID, ownerID uint64) error
		wantErr     bool
		wantErrIs   error
	}{
		{
			name: "正常: 死亡記録と CPM 再同期",
			petFindFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return &model.Pet{ID: 1, OwnerID: 10}, nil
			},
			petUpdateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
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
			name: "正常: CPM 同期失敗はエラーにならない（best-effort）",
			petFindFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return &model.Pet{ID: 1, OwnerID: 10}, nil
			},
			petUpdateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				return nil
			},
			syncCPMFn: func(_ context.Context, _, _ uint64) error {
				return errors.New("sync error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			petRepo := &mockPetRepository{
				findByIDFn: tt.petFindFn,
				updateFn:   tt.petUpdateFn,
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
		{"180日超 → dormant", CPMData{DaysSinceVisit: 181}, CPMStageDormant},
		{"90〜180日 → at_risk", CPMData{DaysSinceVisit: 100}, CPMStageAtRisk},
		{"90日以内・高LTV → loyal_high", CPMData{DaysSinceVisit: 30, LTVAmount: 100_000}, CPMStageLoyalHigh},
		{"90日以内・年6回 → loyal_high", CPMData{DaysSinceVisit: 10, AnnualVisitCount: 6}, CPMStageLoyalHigh},
		{"90日以内・累計4回 → regular", CPMData{DaysSinceVisit: 10, TotalVisitCount: 4}, CPMStageRegular},
		{"90日以内・累計2回 → step", CPMData{DaysSinceVisit: 10, TotalVisitCount: 2}, CPMStageStep},
		{"90日以内・初回 → new", CPMData{DaysSinceVisit: 5, TotalVisitCount: 1}, CPMStageNew},
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
	assert.Equal(t, "ltv_amount_500000plus", ltvBracketTag(500_000))
	assert.Equal(t, "ltv_amount_200000to500000", ltvBracketTag(200_000))
	assert.Equal(t, "ltv_amount_100000to200000", ltvBracketTag(100_000))
	assert.Equal(t, "ltv_amount_50000to100000", ltvBracketTag(50_000))
	assert.Equal(t, "ltv_amount_10000to50000", ltvBracketTag(10_000))
	assert.Equal(t, "ltv_amount_under10000", ltvBracketTag(9_999))
}

func TestVisitCountAnnualTag(t *testing.T) {
	assert.Equal(t, "visit_count_annual_12plus", visitCountAnnualTag(12))
	assert.Equal(t, "visit_count_annual_6to12", visitCountAnnualTag(6))
	assert.Equal(t, "visit_count_annual_3to6", visitCountAnnualTag(3))
	assert.Equal(t, "visit_count_annual_2", visitCountAnnualTag(2))
	assert.Equal(t, "visit_count_annual_0", visitCountAnnualTag(0))
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
