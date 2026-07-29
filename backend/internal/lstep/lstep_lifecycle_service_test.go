package lstep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

type mockTransactor struct{}

func (*mockTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
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
	audit := &mockAuditService{}
	return NewLstepLifecycleService(settingsSvc, ownerRepo, petRepo, tagCacheRepo, syncSvc, audit, nil, &mockTransactor{}, audit)
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

func TestLstepLifecycleService_HandlePetDeath_AlreadyDeceasedConflictNoSideEffects(t *testing.T) {
	var updateCalls int
	var animalClassificationSyncCalls int
	var petBasicInfoSyncCalls int
	var cpmStageSyncCalls int
	var secondaryAuditCalls int

	originalDeceasedAt := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	originalDeceasedReason := "自然死"
	pet := &model.Pet{
		ID:             1,
		OwnerID:        10,
		Status:         model.PetStatusDeceased,
		DeceasedAt:     &originalDeceasedAt,
		DeceasedReason: &originalDeceasedReason,
	}

	petRepo := &mockPetRepository{
		findByIDFn: func(_ context.Context, clinicID, petID uint64) (*model.Pet, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(1), petID)
			return pet, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			updateCalls++
			return nil
		},
		findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) {
			return []model.Pet{{ID: 2, OwnerID: 10, Status: model.PetStatusAlive}}, nil
		},
	}
	syncSvc := &mockLstepTagSyncService{
		syncOwnerAnimalClassificationTagFn: func(_ context.Context, _, _ uint64) error {
			animalClassificationSyncCalls++
			return nil
		},
		syncPetBasicInfoTagsFn: func(_ context.Context, _, _ uint64) error {
			petBasicInfoSyncCalls++
			return nil
		},
		syncCPMStageTagFn: func(_ context.Context, _, _ uint64) error {
			cpmStageSyncCalls++
			return nil
		},
	}
	audit := &mockAuditService{
		logLstepOperationFn: func(_ context.Context, _ uint64, _ *uint64, _, _ string, _ *uint64) error {
			secondaryAuditCalls++
			return nil
		},
	}
	svc := NewLstepLifecycleService(
		defaultLstepSettingsSvc(),
		&mockOwnerRepository{},
		petRepo,
		&mockLstepTagCacheRepository{},
		syncSvc,
		audit,
		nil,
		&mockTransactor{},
		audit,
	)

	err := svc.HandlePetDeath(context.Background(), 1, 1, time.Now(), "病気", nil)

	assert.ErrorIs(t, err, apperrors.ErrConflict)
	assert.Equal(t, model.PetStatusDeceased, pet.Status)
	assert.Equal(t, &originalDeceasedAt, pet.DeceasedAt)
	assert.Equal(t, &originalDeceasedReason, pet.DeceasedReason)
	assert.Equal(t, 0, updateCalls)
	assert.Equal(t, 0, animalClassificationSyncCalls)
	assert.Equal(t, 0, petBasicInfoSyncCalls)
	assert.Equal(t, 0, cpmStageSyncCalls)
	assert.Empty(t, audit.entries)
	assert.Equal(t, 0, secondaryAuditCalls)
	t.Logf(
		"update_calls=%d animal_classification_sync_calls=%d pet_basic_info_sync_calls=%d cpm_stage_sync_calls=%d primary_audit_entries=%d secondary_audit_calls=%d",
		updateCalls,
		animalClassificationSyncCalls,
		petBasicInfoSyncCalls,
		cpmStageSyncCalls,
		len(audit.entries),
		secondaryAuditCalls,
	)
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
				return &model.Pet{ID: 1, OwnerID: 10, Status: model.PetStatusDeceased}, nil
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

func TestLstepLifecycleService_HandlePetRevival_LivingPetConflictNoSideEffects(t *testing.T) {
	var updateCalls int
	var animalClassificationSyncCalls int
	var petBasicInfoSyncCalls int
	var cpmStageSyncCalls int
	var secondaryAuditCalls int

	pet := &model.Pet{
		ID:             1,
		OwnerID:        10,
		Status:         model.PetStatusAlive,
		DeceasedAt:     nil,
		DeceasedReason: nil,
	}

	petRepo := &mockPetRepository{
		findByIDFn: func(_ context.Context, clinicID, petID uint64) (*model.Pet, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(1), petID)
			return pet, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			updateCalls++
			return nil
		},
	}
	syncSvc := &mockLstepTagSyncService{
		syncOwnerAnimalClassificationTagFn: func(_ context.Context, _, _ uint64) error {
			animalClassificationSyncCalls++
			return nil
		},
		syncPetBasicInfoTagsFn: func(_ context.Context, _, _ uint64) error {
			petBasicInfoSyncCalls++
			return nil
		},
		syncCPMStageTagFn: func(_ context.Context, _, _ uint64) error {
			cpmStageSyncCalls++
			return nil
		},
	}
	audit := &mockAuditService{
		logLstepOperationFn: func(_ context.Context, _ uint64, _ *uint64, _, _ string, _ *uint64) error {
			secondaryAuditCalls++
			return nil
		},
	}
	svc := NewLstepLifecycleService(
		defaultLstepSettingsSvc(),
		&mockOwnerRepository{},
		petRepo,
		&mockLstepTagCacheRepository{},
		syncSvc,
		audit,
		nil,
		&mockTransactor{},
		audit,
	)

	err := svc.HandlePetRevival(context.Background(), 1, 1, nil)

	assert.ErrorIs(t, err, apperrors.ErrConflict)
	assert.Equal(t, model.PetStatusAlive, pet.Status)
	assert.Nil(t, pet.DeceasedAt)
	assert.Nil(t, pet.DeceasedReason)
	assert.Equal(t, 0, updateCalls)
	assert.Equal(t, 0, animalClassificationSyncCalls)
	assert.Equal(t, 0, petBasicInfoSyncCalls)
	assert.Equal(t, 0, cpmStageSyncCalls)
	assert.Empty(t, audit.entries)
	assert.Equal(t, 0, secondaryAuditCalls)
	t.Logf(
		"update_calls=%d animal_classification_sync_calls=%d pet_basic_info_sync_calls=%d cpm_stage_sync_calls=%d primary_audit_entries=%d secondary_audit_calls=%d",
		updateCalls,
		animalClassificationSyncCalls,
		petBasicInfoSyncCalls,
		cpmStageSyncCalls,
		len(audit.entries),
		secondaryAuditCalls,
	)
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

			err := svc.HandleOwnerOptOut(context.Background(), 1, 10, "希望なし", nil)
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

			err := svc.HandleOwnerOptIn(context.Background(), 1, 10, nil)
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

// ---- テスト: loadPetDerivedPrefixes ----

func TestLstepLifecycleService_LoadPetDerivedPrefixes(t *testing.T) {
	t.Run("tagConfigRepo nil → fallback prefixes", func(t *testing.T) {
		svc := &lstepLifecycleService{tagConfigRepo: nil}
		prefixes := svc.loadPetDerivedPrefixes(context.Background())
		assert.Equal(t, []string{"vaccine_", TagPrefixCheckupDone}, prefixes)
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
		assert.Equal(t, []string{"vaccine_", TagPrefixCheckupDone}, prefixes)
	})

	t.Run("DB error → fallback", func(t *testing.T) {
		repo := &mockLstepTagConfigRepository{
			findAllAutoManagedPrefixesFn: func(_ context.Context) ([]*model.LstepAutoManagedPrefix, error) {
				return nil, errors.New("db error")
			},
		}
		svc := &lstepLifecycleService{tagConfigRepo: repo}
		prefixes := svc.loadPetDerivedPrefixes(context.Background())
		assert.Equal(t, []string{"vaccine_", TagPrefixCheckupDone}, prefixes)
	})
}

// ---- ビルドコンパイル確認: LstepTagCacheRepository インターフェース実装 ----

var _ lifecycleTagCacheRepository = (*mockLstepTagCacheRepository)(nil)
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
			// LSB-02: client nil must NOT wipe local cache (retry evidence retained)
			name:        "client nil (sync disabled) keeps cache without calling DeleteAll",
			settingsSvc: defaultLstepSettingsSvc(),
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{{TagName: "cpm_new"}}, nil
				},
				deleteAllByOwnerFn: func(_ context.Context, _, _ uint64) error {
					t.Fatal("DeleteAllByOwner must not run when lstep client is nil")
					return nil
				},
			},
			wantErr: false,
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
			return &model.Pet{ID: 1, OwnerID: 10, Status: model.PetStatusAlive}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error { return nil },
		findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) {
			return nil, errors.New("db error")
		},
	}
	svc := newLstepLifecycleSvc(defaultLstepSettingsSvc(), &mockOwnerRepository{}, petRepo, &mockLstepTagCacheRepository{}, &mockLstepTagSyncService{})

	// LSB-03: post-commit FindLiving failure must not invert the successful death write.
	err := svc.HandlePetDeath(context.Background(), 1, 1, time.Now(), "病気", nil)

	assert.NoError(t, err)
}

func TestLstepLifecycleService_HandlePetDeath_AllPetsDead_OwnerFindError(t *testing.T) {
	petRepo := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			return &model.Pet{ID: 1, OwnerID: 10, Status: model.PetStatusAlive}, nil
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
	// defaultLstepSettingsSvc returns empty API key → client nil.
	// LSB-02: without a remote client we keep cache for retry (no DeleteAllByOwner).
	lineUserID := "Uabc123"
	deleteAllCalled := false
	petRepo := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			return &model.Pet{ID: 1, OwnerID: 10, Status: model.PetStatusAlive}, nil
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
	assert.False(t, deleteAllCalled, "LSB-02: nil client must not wipe tag cache")
}

func TestLstepLifecycleService_HandlePetDeath_SurvivingPets_PetDerivedCleanupBestEffort(t *testing.T) {
	lineUserID := "Uabc123"
	petRepo := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			return &model.Pet{ID: 1, OwnerID: 10, Status: model.PetStatusAlive}, nil
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

// BUG-407 (audit fail-closed): 旧実装は監査書込失敗を best-effort として飲み込み、status/deceased_at
// の更新を確定させていた（fail-open）。#211 返金の fail-closed 先例に合わせ、一次監査ログ
// （action="pet_death"）の書込失敗は Transactor.WithTx 経由で status 更新ごとロールバックされ、
// HandlePetDeath はエラーを返すようになった。この service レベルのテストは「エラーが伝播すること」
// (fail-closed の契約) を証明する — status 更新が実 DB でロールバックされること自体の証明は
// mock では不可能なため repository/pet_repository_tx_atomicity_test.go の
// TestPetRepository_Update_RollsBackWhenAmbientTxFails が担う（healthcare-reviewer 指摘どおりの役割分担）。
func TestLstepLifecycleService_HandlePetDeath_AuditLogFailureRollsBackDeathRecord(t *testing.T) {
	petRepo := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			return &model.Pet{ID: 1, OwnerID: 10}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error { return nil },
		findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) {
			return []model.Pet{{ID: 2, OwnerID: 10}}, nil
		},
	}
	audit := &mockAuditService{logEntryTxErr: errors.New("audit db error")}
	svc := NewLstepLifecycleService(
		defaultLstepSettingsSvc(),
		&mockOwnerRepository{},
		petRepo,
		&mockLstepTagCacheRepository{},
		&mockLstepTagSyncService{},
		audit,
		nil,
		&mockTransactor{},
		audit,
	)

	err := svc.HandlePetDeath(context.Background(), 1, 1, time.Now(), "病気", nil)

	assert.Error(t, err, "audit log failure must roll back the death record (fail-closed), not be swallowed as best-effort")
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
				// pet_death_tag_sync (best-effort, surviving-pets branch only) still goes
				// through LogLstepOperation — captured here for parity with the pre-existing shape.
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
				&mockTransactor{},
				auditSvc,
			)

			err := svc.HandlePetDeath(context.Background(), 1, 1, time.Now(), "病気", &actor)
			assert.NoError(t, err)

			// BUG-407: the primary user-action audit (action="pet_death") is now written via
			// AuditTxLogger.LogEntryTx (same tx as the status update, fail-closed), not
			// LogLstepOperation — assert against auditSvc.entries instead of `calls`.
			var found bool
			for _, e := range auditSvc.entries {
				if e.Action != "pet_death" {
					continue
				}
				found = true
				assert.Equal(t, "pet", e.Resource)
				if assert.NotNil(t, e.ActorID) {
					assert.Equal(t, actor, *e.ActorID)
				}
				if assert.NotNil(t, e.ResourceID) {
					assert.Equal(t, uint64(1), *e.ResourceID)
				}
			}
			assert.True(t, found, "expected a pet_death audit log entry (via LogEntryTx) with the real actor")
		})
	}
}

// ---- HandlePetRevival: additional branches ----

func TestLstepLifecycleService_HandlePetRevival_UpdateError(t *testing.T) {
	petRepo := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			return &model.Pet{ID: 1, OwnerID: 10, Status: model.PetStatusDeceased}, nil
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
			return &model.Pet{ID: 1, OwnerID: 10, Status: model.PetStatusDeceased}, nil
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
	audit := &mockAuditService{logLstepOperationErr: errors.New("audit db error")}
	svc := NewLstepLifecycleService(
		defaultLstepSettingsSvc(),
		&mockOwnerRepository{},
		petRepo,
		&mockLstepTagCacheRepository{},
		syncSvc,
		audit,
		nil,
		&mockTransactor{},
		audit,
	)

	err := svc.HandlePetRevival(context.Background(), 1, 1, nil)

	// logLstepOperationErr only affects the secondary "pet_revival_tag_sync" best-effort audit
	// (still LogLstepOperation, unaffected by BUG-407) — the primary "pet_revival" audit now
	// goes through LogEntryTx (no error injected here), so HandlePetRevival still succeeds.
	assert.NoError(t, err)
}

// BUG-407 (audit fail-closed): symmetric with HandlePetDeath — a failure writing the primary
// user-action audit (action="pet_revival") via AuditTxLogger.LogEntryTx rolls back the status
// update in the same tx and HandlePetRevival returns an error, instead of the old best-effort
// swallow. Real-DB rollback proof lives in
// repository/pet_repository_tx_atomicity_test.go (TestPetRepository_Update_RollsBackWhenAmbientTxFails).
func TestLstepLifecycleService_HandlePetRevival_AuditLogFailureRollsBackRevivalRecord(t *testing.T) {
	petRepo := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			return &model.Pet{ID: 1, OwnerID: 10, Status: model.PetStatusDeceased}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error { return nil },
	}
	audit := &mockAuditService{logEntryTxErr: errors.New("audit db error")}
	svc := NewLstepLifecycleService(
		defaultLstepSettingsSvc(),
		&mockOwnerRepository{},
		petRepo,
		&mockLstepTagCacheRepository{},
		&mockLstepTagSyncService{},
		audit,
		nil,
		&mockTransactor{},
		audit,
	)

	err := svc.HandlePetRevival(context.Background(), 1, 1, nil)

	assert.Error(t, err, "audit log failure must roll back the revival record (fail-closed), not be swallowed as best-effort")
}

// BUG-407 follow-up (healthcare-reviewer HIGH finding): HandlePetRevival must record the
// user-driven revival event itself in the audit trail (action="pet_revival") with the
// real actor (symmetric with HandlePetDeath's new "pet_death" audit).
func TestLstepLifecycleService_HandlePetRevival_LogsUserActionAudit(t *testing.T) {
	actor := uint64(77)

	petRepo := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			return &model.Pet{ID: 1, OwnerID: 10, Status: model.PetStatusDeceased}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error { return nil },
	}
	auditSvc := &mockAuditService{}
	svc := NewLstepLifecycleService(
		defaultLstepSettingsSvc(),
		&mockOwnerRepository{},
		petRepo,
		&mockLstepTagCacheRepository{},
		&mockLstepTagSyncService{},
		auditSvc,
		nil,
		&mockTransactor{},
		auditSvc,
	)

	err := svc.HandlePetRevival(context.Background(), 1, 1, &actor)
	assert.NoError(t, err)

	// BUG-407: the primary user-action audit (action="pet_revival") is now written via
	// AuditTxLogger.LogEntryTx (same tx as the status update, fail-closed), not
	// LogLstepOperation — assert against auditSvc.entries instead of a LogLstepOperation capture.
	var found bool
	for _, e := range auditSvc.entries {
		if e.Action != "pet_revival" {
			continue
		}
		found = true
		assert.Equal(t, "pet", e.Resource)
		if assert.NotNil(t, e.ActorID) {
			assert.Equal(t, actor, *e.ActorID)
		}
		if assert.NotNil(t, e.ResourceID) {
			assert.Equal(t, uint64(1), *e.ResourceID)
		}
	}
	assert.True(t, found, "expected a pet_revival audit log entry (via LogEntryTx) with the real actor")
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

	err := svc.HandleOwnerOptIn(context.Background(), 1, 10, nil)

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

	err := svc.HandleOwnerOptIn(context.Background(), 1, 10, nil)

	assert.NoError(t, err)
}

// LSA-05: opt-out/in/deletion leave audit trails with actor when provided.
func TestLstepLifecycleService_HandleOwnerOptOut_LogsAuditWithActor(t *testing.T) {
	actor := uint64(7)
	var gotAction string
	var gotActor *uint64
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 10}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error { return nil },
	}
	audit := &mockAuditService{
		logLstepOperationFn: func(_ context.Context, _ uint64, actorID *uint64, action, _ string, _ *uint64) error {
			gotAction = action
			gotActor = actorID
			return nil
		},
	}
	svc := NewLstepLifecycleService(defaultLstepSettingsSvc(), ownerRepo, &mockPetRepository{}, &mockLstepTagCacheRepository{}, &mockLstepTagSyncService{}, audit, nil, &mockTransactor{}, audit)

	err := svc.HandleOwnerOptOut(context.Background(), 1, 10, "希望なし", &actor)
	require.NoError(t, err)
	assert.Equal(t, "owner_lstep_opt_out", gotAction)
	if assert.NotNil(t, gotActor) {
		assert.Equal(t, actor, *gotActor)
	}
}

func TestLstepLifecycleService_HandleOwnerOptIn_LogsAuditWithActor(t *testing.T) {
	actor := uint64(9)
	var gotAction string
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 10, LstepOptOut: true}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error { return nil },
	}
	audit := &mockAuditService{
		logLstepOperationFn: func(_ context.Context, _ uint64, _ *uint64, action, _ string, _ *uint64) error {
			gotAction = action
			return nil
		},
	}
	svc := NewLstepLifecycleService(defaultLstepSettingsSvc(), ownerRepo, &mockPetRepository{}, &mockLstepTagCacheRepository{}, &mockLstepTagSyncService{}, audit, nil, &mockTransactor{}, audit)

	err := svc.HandleOwnerOptIn(context.Background(), 1, 10, &actor)
	require.NoError(t, err)
	assert.Equal(t, "owner_lstep_opt_in", gotAction)
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
