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

// ---- lstep.Client モック ----

type mockLstepClientForDelivery struct {
	addTagFn func(ctx context.Context, lineUserID, tagName string) error
}

func (m *mockLstepClientForDelivery) AddTag(ctx context.Context, lineUserID, tagName string) error {
	if m.addTagFn != nil {
		return m.addTagFn(ctx, lineUserID, tagName)
	}
	return nil
}
func (m *mockLstepClientForDelivery) RemoveTag(_ context.Context, _, _ string) error { return nil }
func (m *mockLstepClientForDelivery) AddTagBulk(_ context.Context, _ []string, _ string) error {
	return nil
}
func (m *mockLstepClientForDelivery) GetUserTags(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}
func (m *mockLstepClientForDelivery) GetUser(_ context.Context, _ string) (*lstep.UserInfo, error) {
	return nil, nil
}
func (m *mockLstepClientForDelivery) SetProperty(_ context.Context, _, _, _ string) error {
	return nil
}

// ---- LstepDeliveryTriggerLogRepository モック ----

type mockDeliveryTriggerLogRepository struct {
	createFn              func(ctx context.Context, log *model.LstepDeliveryTriggerLog) error
	existsTodayFn         func(ctx context.Context, clinicID, ownerID uint64, triggerType string, date time.Time) (bool, error)
	updateStatusFn        func(ctx context.Context, id uint64, status string, firedAt *time.Time, excludedReason *string) error
	findByClinicAndDateFn func(ctx context.Context, clinicID uint64, date time.Time) ([]model.LstepDeliveryTriggerLog, error)
	markKnownUnsentFn     func(ctx context.Context, clinicID, id uint64, reason string) error
}

func (m *mockDeliveryTriggerLogRepository) Create(ctx context.Context, log *model.LstepDeliveryTriggerLog) error {
	if m.createFn != nil {
		return m.createFn(ctx, log)
	}
	log.ID = 1
	return nil
}
func (m *mockDeliveryTriggerLogRepository) CreateIfAbsentToday(ctx context.Context, log *model.LstepDeliveryTriggerLog) (bool, error) {
	if err := m.Create(ctx, log); err != nil {
		return false, err
	}
	return true, nil
}
func (m *mockDeliveryTriggerLogRepository) ExistsTodayByOwnerAndType(ctx context.Context, clinicID, ownerID uint64, triggerType string, date time.Time) (bool, error) {
	if m.existsTodayFn != nil {
		return m.existsTodayFn(ctx, clinicID, ownerID, triggerType, date)
	}
	return false, nil
}
func (m *mockDeliveryTriggerLogRepository) UpdateStatus(ctx context.Context, clinicID, id uint64, status string, firedAt *time.Time, excludedReason *string) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status, firedAt, excludedReason)
	}
	return nil
}
func (m *mockDeliveryTriggerLogRepository) FindByClinicAndDate(ctx context.Context, clinicID uint64, date time.Time) ([]model.LstepDeliveryTriggerLog, error) {
	if m.findByClinicAndDateFn != nil {
		return m.findByClinicAndDateFn(ctx, clinicID, date)
	}
	return nil, nil
}
func (m *mockDeliveryTriggerLogRepository) CountByStatusAndDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (map[string]int64, error) {
	return map[string]int64{}, nil
}
func (m *mockDeliveryTriggerLogRepository) CountExcludedReasonByDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (map[string]int64, error) {
	return map[string]int64{}, nil
}
func (m *mockDeliveryTriggerLogRepository) CountSuppressedByPriorityDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (int64, error) {
	return 0, nil
}

// FindByDateRangeWithFilters deliberately fails so runBatch bulk day-log preload
// falls back to per-owner ExistsTodayByOwnerAndType (existsTodayFn). Returning
// (nil, 0, nil) would install an empty authoritative day-log cache and break
// already-fired tests that only stub ExistsToday.
func (m *mockDeliveryTriggerLogRepository) FindByDateRangeWithFilters(_ context.Context, _ uint64, _, _ time.Time, _, _ string, _, _ int) ([]DeliveryTriggerLogRow, int64, error) {
	return nil, 0, errors.New("mock: bulk day-log not configured; use per-owner ExistsToday")
}
func (m *mockDeliveryTriggerLogRepository) CountByTypeAndStatus(_ context.Context, _ uint64, _, _ time.Time) ([]DeliveryStatsRow, error) {
	return nil, nil
}
func (m *mockDeliveryTriggerLogRepository) CountVisitConversionsByType(_ context.Context, _ uint64, _, _ time.Time, _ int) ([]VisitConversionRow, error) {
	return nil, nil
}
func (m *mockDeliveryTriggerLogRepository) FindByOwnerAndDate(_ context.Context, _, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
	return nil, nil
}
func (m *mockDeliveryTriggerLogRepository) UpdateSuppressed(_ context.Context, _, _ uint64, _ string) error {
	return nil
}
func (m *mockDeliveryTriggerLogRepository) MarkKnownUnsent(ctx context.Context, clinicID, id uint64, reason string) error {
	if m.markKnownUnsentFn != nil {
		return m.markKnownUnsentFn(ctx, clinicID, id, reason)
	}
	return m.UpdateStatus(ctx, clinicID, id, model.TriggerStatusFailed, nil, &reason)
}

// ---- OwnerRepository モック（delivery trigger 用）----

type mockOwnerRepoForDelivery struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
}

func (m *mockOwnerRepoForDelivery) FindAll(_ context.Context, _ []uint64, _, _ int, _ string) ([]model.Owner, int64, error) {
	return nil, 0, nil
}
func (m *mockOwnerRepoForDelivery) FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}
func (m *mockOwnerRepoForDelivery) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.Owner, error) {
	return nil, nil
}
func (m *mockOwnerRepoForDelivery) FindByEmail(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
	return nil, nil
}
func (m *mockOwnerRepoForDelivery) FindByPhone(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
	return nil, nil
}
func (m *mockOwnerRepoForDelivery) FindByNameAndPhone(_ context.Context, _ uint64, _, _ string) (*model.Owner, error) {
	return nil, nil
}
func (m *mockOwnerRepoForDelivery) FindByLineUserID(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
	return nil, nil
}
func (m *mockOwnerRepoForDelivery) FindAllWithLineUserID(_ context.Context, _ uint64) ([]model.Owner, error) {
	return nil, nil
}
func (m *mockOwnerRepoForDelivery) FindAllWithLineUserIDCursor(_ context.Context, _, _ uint64, _ int) ([]model.Owner, error) {
	return nil, nil
}
func (m *mockOwnerRepoForDelivery) CreateWithPets(_ context.Context, _ *model.Owner, _ []model.Pet) error {
	return nil
}
func (m *mockOwnerRepoForDelivery) Update(_ context.Context, _, _ uint64, _ map[string]any) error {
	return nil
}
func (m *mockOwnerRepoForDelivery) UpdateLineUserID(_ context.Context, _, _ uint64, _ *string) error {
	return nil
}
func (m *mockOwnerRepoForDelivery) UpdateLineFollowedAt(
	_ context.Context,
	_, _ uint64,
	_ string,
	_ time.Time,
) (bool, error) {
	return true, nil
}
func (m *mockOwnerRepoForDelivery) UpdateLineBlockedAt(
	_ context.Context,
	_, _ uint64,
	_ string,
	_ time.Time,
) (bool, error) {
	return true, nil
}
func (m *mockOwnerRepoForDelivery) Delete(_ context.Context, _, _ uint64) error { return nil }
func (m *mockOwnerRepoForDelivery) CountPetsByOwnerID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

// FindByIDs materializes via FindByID so runBatch bulk install (ce79e0c23) does not
// install an empty owner cache that later returns (nil, nil) on miss+fallback.
func (m *mockOwnerRepoForDelivery) FindByIDs(ctx context.Context, clinicID uint64, ownerIDs []uint64) ([]*model.Owner, error) {
	out := make([]*model.Owner, 0, len(ownerIDs))
	for _, id := range ownerIDs {
		o, err := m.FindByID(ctx, clinicID, id)
		if err != nil {
			return nil, err
		}
		if o != nil {
			out = append(out, o)
		}
	}
	return out, nil
}

// ---- TagCacheRepository モック（delivery trigger 用）----

type mockTagCacheRepoForDelivery struct {
	findByOwnerFn           func(ctx context.Context, clinicID, ownerID uint64) ([]*model.LstepTagCache, error)
	findOwnerIDsByTagFn     func(ctx context.Context, clinicID uint64, tagName string) ([]uint64, error)
	withDeliveryOwnerLockFn func(ctx context.Context, clinicID, ownerID uint64, fn func(context.Context) error) error
}

func (m *mockTagCacheRepoForDelivery) FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]*model.LstepTagCache, error) {
	if m.findByOwnerFn != nil {
		return m.findByOwnerFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}

// FindByOwners materializes via FindByOwner so bulk tag cache install does not
// treat missing keys as authoritative empty (which would hide EXCL tags).
func (m *mockTagCacheRepoForDelivery) FindByOwners(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64][]*model.LstepTagCache, error) {
	out := make(map[uint64][]*model.LstepTagCache, len(ownerIDs))
	for _, id := range ownerIDs {
		tags, err := m.FindByOwner(ctx, clinicID, id)
		if err != nil {
			return nil, err
		}
		if tags == nil {
			tags = []*model.LstepTagCache{}
		}
		out[id] = tags
	}
	return out, nil
}
func (m *mockTagCacheRepoForDelivery) WithDeliveryOwnerLock(ctx context.Context, clinicID, ownerID uint64, fn func(context.Context) error) error {
	if m.withDeliveryOwnerLockFn != nil {
		return m.withDeliveryOwnerLockFn(ctx, clinicID, ownerID, fn)
	}
	return fn(ctx)
}

func (m *mockTagCacheRepoForDelivery) FindOwnerIDsByTag(ctx context.Context, clinicID uint64, tagName string) ([]uint64, error) {
	if m.findOwnerIDsByTagFn != nil {
		return m.findOwnerIDsByTagFn(ctx, clinicID, tagName)
	}
	return nil, nil
}
func (m *mockTagCacheRepoForDelivery) UpsertTag(_ context.Context, _, _ uint64, _, _, _ string) error {
	return nil
}
func (m *mockTagCacheRepoForDelivery) DeleteTag(_ context.Context, _, _ uint64, _ string) error {
	return nil
}
func (m *mockTagCacheRepoForDelivery) DeleteAllByOwner(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *mockTagCacheRepoForDelivery) TagSummary(_ context.Context, _ uint64) ([]TagSummaryRow, int64, error) {
	return nil, 0, nil
}
func (m *mockTagCacheRepoForDelivery) FindOwnersByTag(_ context.Context, _ uint64, _, _ string, _, _ int) ([]TagOwnerRow, int64, error) {
	return nil, 0, nil
}
func (m *mockTagCacheRepoForDelivery) BulkReplaceOwnerTags(_ context.Context, _, _ uint64, _ []TagEntry) error {
	return nil
}

// ---- VaccinationRepository モック（delivery trigger 用）----

type mockVaccinationRepoForDelivery struct {
	findOwnersByVaccineDeadlineFn func(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error)
}

func (m *mockVaccinationRepoForDelivery) FindAll(_ context.Context, _ uint64, _, _ *uint64, _, _ *string, _, _ int) ([]model.Vaccination, int64, error) {
	return nil, 0, nil
}
func (m *mockVaccinationRepoForDelivery) FindByID(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
	return nil, nil
}
func (m *mockVaccinationRepoForDelivery) FindByOwner(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
	return nil, nil
}
func (m *mockVaccinationRepoForDelivery) Create(_ context.Context, _ *model.Vaccination) error {
	return nil
}
func (m *mockVaccinationRepoForDelivery) Update(_ context.Context, _, _ uint64, _ map[string]any) (*model.Vaccination, error) {
	return nil, nil
}
func (m *mockVaccinationRepoForDelivery) Delete(_ context.Context, _, _ uint64) error { return nil }
func (m *mockVaccinationRepoForDelivery) FindOwnersByVaccineDeadline(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error) {
	if m.findOwnersByVaccineDeadlineFn != nil {
		return m.findOwnersByVaccineDeadlineFn(ctx, clinicID, targetDate)
	}
	return nil, nil
}

// ---- PetRepository モック（delivery trigger 用）----

type mockPetRepoForDelivery struct {
	findOwnersByPetBirthdayFn func(ctx context.Context, clinicID uint64, month, day int) ([]uint64, error)
}

func (m *mockPetRepoForDelivery) FindByID(_ context.Context, _, _ uint64) (*model.Pet, error) {
	return nil, nil
}
func (m *mockPetRepoForDelivery) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.Pet, error) {
	return nil, nil
}
func (m *mockPetRepoForDelivery) FindLivingByOwner(_ context.Context, _, _ uint64) ([]model.Pet, error) {
	return nil, nil
}
func (m *mockPetRepoForDelivery) CountByOwner(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *mockPetRepoForDelivery) CountLivingByOwner(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *mockPetRepoForDelivery) CountUsageByAnimalSpeciesID(_ context.Context, _ uint64) (int64, error) {
	return 0, nil
}
func (m *mockPetRepoForDelivery) Create(_ context.Context, _ *model.Pet) error { return nil }
func (m *mockPetRepoForDelivery) Update(_ context.Context, _, _ uint64, _ map[string]any) error {
	return nil
}
func (m *mockPetRepoForDelivery) Delete(_ context.Context, _, _ uint64) error { return nil }
func (m *mockPetRepoForDelivery) FindOwnersByPetBirthday(ctx context.Context, clinicID uint64, month, day int) ([]uint64, error) {
	if m.findOwnersByPetBirthdayFn != nil {
		return m.findOwnersByPetBirthdayFn(ctx, clinicID, month, day)
	}
	return nil, nil
}
func (m *mockPetRepoForDelivery) CountLivingByOwnerIDs(_ context.Context, _ uint64, _ []uint64) (map[uint64]int64, error) {
	return map[uint64]int64{}, nil
}

// ---- AppointmentRepository モック（EMR-308 tentative reminder 用）----

type mockDeliveryAppointmentRepository struct {
	findOwnerIDsWithTentativeFn func(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error)
	findTentativesByOwnerFn     func(ctx context.Context, clinicID, ownerID uint64, targetDate time.Time) ([]model.Reservation, error)
	withReminderSerializationFn func(ctx context.Context, clinicID uint64, petIDs []uint64, fn func(txCtx context.Context) error) error
}

func (m *mockDeliveryAppointmentRepository) FindOwnerIDsWithTentativeIntendedDate(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error) {
	if m.findOwnerIDsWithTentativeFn != nil {
		return m.findOwnerIDsWithTentativeFn(ctx, clinicID, targetDate)
	}
	return nil, nil
}

func (m *mockDeliveryAppointmentRepository) FindTentativesByOwnerAndIntendedDate(ctx context.Context, clinicID, ownerID uint64, targetDate time.Time) ([]model.Reservation, error) {
	if m.findTentativesByOwnerFn != nil {
		return m.findTentativesByOwnerFn(ctx, clinicID, ownerID, targetDate)
	}
	// デフォルトは active tentative 生存＝再検証通過（再検証ケース以外のセットアップを簡潔にする）。
	// PetID は直列化ロックのキー解決に必須のためスタブ値を付ける。
	petID := uint64(9001)
	return []model.Reservation{{OwnerID: &ownerID, PetID: &petID, Status: model.ReservationStatusTentative}}, nil
}

func (m *mockDeliveryAppointmentRepository) WithTentativeReminderSerialization(ctx context.Context, clinicID uint64, petIDs []uint64, fn func(txCtx context.Context) error) error {
	if m.withReminderSerializationFn != nil {
		return m.withReminderSerializationFn(ctx, clinicID, petIDs, fn)
	}
	// デフォルトはロックなし pass-through —— 実装側の直列化は repo 統合テストで担保し、
	// ここでは hook 呼出の有無だけを検証可能にする。
	return fn(ctx)
}

// ---- helpers ----

func buildDeliverySvc(
	ownerRepo deliveryOwnerRepository,
	medRepo deliveryMedicalRecordRepository,
	vacRepo deliveryVaccinationRepository,
	petRepo deliveryPetRepository,
	tagRepo deliveryTagCacheRepository,
	logRepo LstepDeliveryTriggerLogRepository,
	settingsSvc LstepSettingsService,
	appointmentRepos ...deliveryAppointmentRepository,
) LstepDeliveryTriggerService {
	var appointmentRepo deliveryAppointmentRepository
	if len(appointmentRepos) > 0 {
		appointmentRepo = appointmentRepos[0]
	}
	return NewLstepDeliveryTriggerService(ownerRepo, medRepo, vacRepo, petRepo, tagRepo, logRepo, settingsSvc, nil, appointmentRepo)
}

// injectTestClient はテスト用の lstep.Client モックをサービスに注入する。
// HTTP呼び出しを避けて AddTag の成否を制御できる。
func injectTestClient(svc LstepDeliveryTriggerService, client lstep.Client) {
	svc.(*lstepDeliveryTriggerService).clientBuilderFn = func(_ context.Context, _ uint64) (lstep.Client, error) {
		return client, nil
	}
}

func defaultOwnerWithLine(ownerID uint64) *model.Owner {
	lineID := "U_test_line_id"
	return &model.Owner{ID: ownerID, ClinicID: 1, LineUserID: &lineID}
}

func defaultOwnerNoLine(ownerID uint64) *model.Owner {
	return &model.Owner{ID: ownerID, ClinicID: 1, LineUserID: nil}
}

func defaultOwnerExcluded(ownerID uint64) *model.Owner {
	lineID := "U_excluded"
	return &model.Owner{ID: ownerID, ClinicID: 1, LineUserID: &lineID, DeliveryExcluded: true}
}

// enabledSettings は IsSyncEnabled=true、GetRawCredentials で有効な apiKey を返すモック。
func enabledSettings() *mockLstepSettingsService {
	return &mockLstepSettingsService{
		isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
		getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
			return "test_api_key", "https://api.example.com", "", nil
		},
	}
}

// disabledSettings は IsSyncEnabled=false を返すモック（クライアント未生成）。
func disabledSettings() *mockLstepSettingsService {
	return &mockLstepSettingsService{
		isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil },
	}
}

// ---- Tests ----

func TestLstepDeliveryTriggerService_TriggerFirstVisitFollowUp3D(t *testing.T) {
	asOf := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	expected3DagsAgo := time.Date(2026, 5, 7, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		ownerIDs     []uint64
		repoErr      error
		alreadyFired bool
		excluded     bool
		noLineID     bool
		addTagErr    error
		wantCount    int
		wantErrCount int
	}{
		{
			name:      "fires for matching owners",
			ownerIDs:  []uint64{10},
			wantCount: 1,
		},
		{
			name:         "skips already fired today",
			ownerIDs:     []uint64{10},
			alreadyFired: true,
			wantCount:    0,
		},
		{
			name:      "skips delivery_excluded owners",
			ownerIDs:  []uint64{10},
			excluded:  true,
			wantCount: 0,
		},
		{
			name:      "skips owners without line_user_id",
			ownerIDs:  []uint64{10},
			noLineID:  true,
			wantCount: 0,
		},
		{
			name:         "returns error when repo fails",
			repoErr:      errors.New("db error"),
			wantCount:    0,
			wantErrCount: 1,
		},
		{
			name:      "fires for multiple owners",
			ownerIDs:  []uint64{10, 11, 12},
			wantCount: 3,
		},
		{
			name:         "returns error when AddTag fails",
			ownerIDs:     []uint64{10},
			addTagErr:    errors.New("lstep error"),
			wantCount:    0,
			wantErrCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedTargetDate time.Time
			medRepo := &mockMedicalRecordRepository{
				findOwnersByFirstVisitFn: func(_ context.Context, _ uint64, targetDate time.Time) ([]uint64, error) {
					capturedTargetDate = targetDate
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return tt.ownerIDs, nil
				},
			}
			ownerRepo := &mockOwnerRepoForDelivery{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					if tt.noLineID {
						return defaultOwnerNoLine(id), nil
					}
					if tt.excluded {
						return defaultOwnerExcluded(id), nil
					}
					return defaultOwnerWithLine(id), nil
				},
			}
			tagRepo := &mockTagCacheRepoForDelivery{}
			logRepo := &mockDeliveryTriggerLogRepository{}
			settings := &mockLstepSettingsService{
				isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
				getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
					return "key", "https://api.example.com", "", nil
				},
			}

			svc := buildDeliverySvc(ownerRepo, medRepo, nil, nil, tagRepo, logRepo, settings)
			injectTestClient(svc, &mockLstepClientForDelivery{
				addTagFn: func(_ context.Context, _, _ string) error { return tt.addTagErr },
			})

			if tt.alreadyFired {
				logRepo.existsTodayFn = func(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
					return true, nil
				}
			}

			count, errs := svc.TriggerFirstVisitFollowUp3D(context.Background(), 1, asOf)

			if tt.repoErr == nil && len(tt.ownerIDs) > 0 {
				assert.Equal(t, expected3DagsAgo.Format("2006-01-02"), capturedTargetDate.Format("2006-01-02"))
			}
			assert.Equal(t, tt.wantCount, count)
			assert.Len(t, errs, tt.wantErrCount)
		})
	}
}

func TestLstepDeliveryTriggerService_TriggerFirstVisitFollowUp7D(t *testing.T) {
	asOf := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	expected7DagsAgo := time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC)

	medRepo := &mockMedicalRecordRepository{
		findOwnersByFirstVisitFn: func(_ context.Context, _ uint64, targetDate time.Time) ([]uint64, error) {
			assert.Equal(t, expected7DagsAgo.Format("2006-01-02"), targetDate.Format("2006-01-02"))
			return []uint64{20}, nil
		},
	}
	ownerRepo := &mockOwnerRepoForDelivery{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
			return defaultOwnerWithLine(id), nil
		},
	}
	svc := buildDeliverySvc(ownerRepo, medRepo, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
	injectTestClient(svc, &mockLstepClientForDelivery{})
	count, errs := svc.TriggerFirstVisitFollowUp7D(context.Background(), 1, asOf)
	assert.Equal(t, 1, count)
	assert.Empty(t, errs)
}

// TestLstepDeliveryTriggerService_TentativeReminderStages は EMR-308 の tentative
// リマインダー2ステージ（3日前/当日）を検証する。candidate 抽出は intended_date、
// 送信直前に同一 intended_date の active tentative を再検証する。
func TestLstepDeliveryTriggerService_TentativeReminderStages(t *testing.T) {
	asOf := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)

	t.Run("3d stage queries intended_date=asOf+3 and fires stage-qualified tag", func(t *testing.T) {
		var capturedTarget time.Time
		var addedTags []string
		var capturedLogTypes []string
		apptRepo := &mockDeliveryAppointmentRepository{
			findOwnerIDsWithTentativeFn: func(_ context.Context, _ uint64, targetDate time.Time) ([]uint64, error) {
				capturedTarget = targetDate
				return []uint64{30, 31}, nil
			},
		}
		ownerRepo := &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return defaultOwnerWithLine(id), nil
			},
		}
		logRepo := &mockDeliveryTriggerLogRepository{
			createFn: func(_ context.Context, log *model.LstepDeliveryTriggerLog) error {
				capturedLogTypes = append(capturedLogTypes, log.TriggerType)
				log.ID = 1
				return nil
			},
		}
		svc := buildDeliverySvc(ownerRepo, nil, nil, nil, &mockTagCacheRepoForDelivery{}, logRepo, enabledSettings(), apptRepo)
		injectTestClient(svc, &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, tagName string) error {
				addedTags = append(addedTags, tagName)
				return nil
			},
		})

		count, errs := svc.TriggerNextVisitTentativeReminder3D(context.Background(), 1, asOf)

		assert.Equal(t, asOf.AddDate(0, 0, 3).Format("2006-01-02"), capturedTarget.Format("2006-01-02"))
		assert.Equal(t, 2, count)
		assert.Empty(t, errs)
		for _, tag := range addedTags {
			assert.Equal(t, string(model.TriggerTypeNextVisitTentativeReminder3d), tag)
		}
		for _, tt := range capturedLogTypes {
			assert.Equal(t, string(model.TriggerTypeNextVisitTentativeReminder3d), tt)
		}
	})

	t.Run("0d stage queries intended_date=asOf and fires stage-qualified tag", func(t *testing.T) {
		var capturedTarget time.Time
		var addedTags []string
		apptRepo := &mockDeliveryAppointmentRepository{
			findOwnerIDsWithTentativeFn: func(_ context.Context, _ uint64, targetDate time.Time) ([]uint64, error) {
				capturedTarget = targetDate
				return []uint64{30}, nil
			},
		}
		ownerRepo := &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return defaultOwnerWithLine(id), nil
			},
		}
		svc := buildDeliverySvc(ownerRepo, nil, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings(), apptRepo)
		injectTestClient(svc, &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, tagName string) error {
				addedTags = append(addedTags, tagName)
				return nil
			},
		})

		count, errs := svc.TriggerNextVisitTentativeReminder0D(context.Background(), 1, asOf)

		assert.Equal(t, asOf.Format("2006-01-02"), capturedTarget.Format("2006-01-02"))
		assert.Equal(t, 1, count)
		assert.Empty(t, errs)
		assert.Equal(t, []string{string(model.TriggerTypeNextVisitTentativeReminder0d)}, addedTags)
	})

	t.Run("stage-qualified idempotency: 0d fired does not suppress 3d", func(t *testing.T) {
		apptRepo := &mockDeliveryAppointmentRepository{
			findOwnerIDsWithTentativeFn: func(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
				return []uint64{30}, nil
			},
		}
		ownerRepo := &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return defaultOwnerWithLine(id), nil
			},
		}
		// 当日ステージのみ発火済み。3日前ステージの冪等キーは trigger_type 別なので送信される。
		logRepo := &mockDeliveryTriggerLogRepository{
			existsTodayFn: func(_ context.Context, _, _ uint64, triggerType string, _ time.Time) (bool, error) {
				return triggerType == string(model.TriggerTypeNextVisitTentativeReminder0d), nil
			},
		}
		svc := buildDeliverySvc(ownerRepo, nil, nil, nil, &mockTagCacheRepoForDelivery{}, logRepo, enabledSettings(), apptRepo)
		injectTestClient(svc, &mockLstepClientForDelivery{})

		count3d, errs := svc.TriggerNextVisitTentativeReminder3D(context.Background(), 1, asOf)
		assert.Equal(t, 1, count3d)
		assert.Empty(t, errs)
	})

	t.Run("idempotency: same stage already fired skips", func(t *testing.T) {
		apptRepo := &mockDeliveryAppointmentRepository{
			findOwnerIDsWithTentativeFn: func(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
				return []uint64{30}, nil
			},
		}
		ownerRepo := &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return defaultOwnerWithLine(id), nil
			},
		}
		logRepo := &mockDeliveryTriggerLogRepository{
			existsTodayFn: func(_ context.Context, _, _ uint64, triggerType string, _ time.Time) (bool, error) {
				return triggerType == string(model.TriggerTypeNextVisitTentativeReminder3d), nil
			},
		}
		svc := buildDeliverySvc(ownerRepo, nil, nil, nil, &mockTagCacheRepoForDelivery{}, logRepo, enabledSettings(), apptRepo)
		injectTestClient(svc, &mockLstepClientForDelivery{})

		count, errs := svc.TriggerNextVisitTentativeReminder3D(context.Background(), 1, asOf)
		assert.Equal(t, 0, count)
		assert.Empty(t, errs)
	})

	t.Run("pre-send revalidation excludes stale tentative (cancelled/promoted/deleted)", func(t *testing.T) {
		var addTagCalls int
		var capturedReason *string
		var capturedStatus string
		apptRepo := &mockDeliveryAppointmentRepository{
			findOwnerIDsWithTentativeFn: func(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
				return []uint64{30}, nil
			},
			// candidate 抽出後に tentative が消えた（取消・昇格・削除）ケース: 再検証 nil。
			findTentativesByOwnerFn: func(_ context.Context, _, _ uint64, _ time.Time) ([]model.Reservation, error) {
				return nil, nil
			},
		}
		ownerRepo := &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return defaultOwnerWithLine(id), nil
			},
		}
		logRepo := &mockDeliveryTriggerLogRepository{
			updateStatusFn: func(_ context.Context, _ uint64, status string, _ *time.Time, excludedReason *string) error {
				capturedStatus = status
				capturedReason = excludedReason
				return nil
			},
		}
		svc := buildDeliverySvc(ownerRepo, nil, nil, nil, &mockTagCacheRepoForDelivery{}, logRepo, enabledSettings(), apptRepo)
		injectTestClient(svc, &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, _ string) error {
				addTagCalls++
				return nil
			},
		})

		count, errs := svc.TriggerNextVisitTentativeReminder3D(context.Background(), 1, asOf)

		assert.Equal(t, 0, count, "stale candidate は送信しない")
		assert.Empty(t, errs)
		assert.Equal(t, 0, addTagCalls, "LSTEP へのタグ送信は行われない")
		assert.Equal(t, model.TriggerStatusExcluded, capturedStatus)
		if assert.NotNil(t, capturedReason) {
			assert.Equal(t, "stale_tentative_appointment", *capturedReason)
		}
	})

	t.Run("revalidation error propagates as per-owner error", func(t *testing.T) {
		apptRepo := &mockDeliveryAppointmentRepository{
			findOwnerIDsWithTentativeFn: func(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
				return []uint64{30}, nil
			},
			findTentativesByOwnerFn: func(_ context.Context, _, _ uint64, _ time.Time) ([]model.Reservation, error) {
				return nil, errors.New("revalidation db error")
			},
		}
		ownerRepo := &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return defaultOwnerWithLine(id), nil
			},
		}
		svc := buildDeliverySvc(ownerRepo, nil, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings(), apptRepo)
		injectTestClient(svc, &mockLstepClientForDelivery{})

		count, errs := svc.TriggerNextVisitTentativeReminder3D(context.Background(), 1, asOf)
		assert.Equal(t, 0, count)
		assert.Len(t, errs, 1)
	})

	t.Run("candidate query error propagates", func(t *testing.T) {
		apptRepo := &mockDeliveryAppointmentRepository{
			findOwnerIDsWithTentativeFn: func(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
				return nil, errors.New("appointment db error")
			},
		}
		svc := buildDeliverySvc(&mockOwnerRepoForDelivery{}, nil, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings(), apptRepo)
		injectTestClient(svc, &mockLstepClientForDelivery{})

		count, errs := svc.TriggerNextVisitTentativeReminder3D(context.Background(), 1, asOf)
		assert.Equal(t, 0, count)
		assert.Len(t, errs, 1)
	})

	t.Run("locks every candidate pet and revalidates under all locks (Codex #22)", func(t *testing.T) {
		// 同一 owner の対象日 tentative が複数ペットに跨るケース。1ペットだけを
		// ロックして owner+date で再検索すると、ロック外ペットの intent で
		// 再検証を通過し得るため、全 candidate pet をロック対象にする。
		petA, petB := uint64(9001), uint64(9002)
		ownerID := uint64(30)
		ownerScoped := func(_ context.Context, _, oid uint64, _ time.Time) ([]model.Reservation, error) {
			if oid != ownerID {
				return nil, nil
			}
			return []model.Reservation{
				{ID: 501, OwnerID: &ownerID, PetID: &petA, Status: model.ReservationStatusTentative},
				{ID: 502, OwnerID: &ownerID, PetID: &petB, Status: model.ReservationStatusTentative},
			}, nil
		}
		var capturedPetIDs []uint64
		var revalidateUnderLock bool
		var addTagCalls int
		apptRepo := &mockDeliveryAppointmentRepository{
			findOwnerIDsWithTentativeFn: func(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
				return []uint64{ownerID}, nil
			},
			findTentativesByOwnerFn: ownerScoped,
			withReminderSerializationFn: func(ctx context.Context, _ uint64, petIDs []uint64, fn func(txCtx context.Context) error) error {
				capturedPetIDs = append([]uint64(nil), petIDs...)
				// fn 内で呼ばれる再検証は全ロック保持中に走る —— フラグで検証する。
				revalidateUnderLock = true
				defer func() { revalidateUnderLock = false }()
				return fn(ctx)
			},
		}
		ownerRepo := &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return defaultOwnerWithLine(id), nil
			},
		}
		svc := buildDeliverySvc(ownerRepo, nil, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings(), apptRepo)
		injectTestClient(svc, &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, _ string) error {
				addTagCalls++
				return nil
			},
		})

		count, errs := svc.TriggerNextVisitTentativeReminder3D(context.Background(), 1, asOf)

		assert.Equal(t, 1, count)
		assert.Empty(t, errs)
		assert.ElementsMatch(t, []uint64{petA, petB}, capturedPetIDs,
			"Codex #22: 送信根拠の tentative を持つ全 candidate pet がロック対象")
		assert.False(t, revalidateUnderLock, "serialization fn は正常に return している")
		assert.Equal(t, 1, addTagCalls)
	})

	t.Run("serialization ctx revalidation sees lock-scoped data (Codex #22)", func(t *testing.T) {
		// ロック取得後の再検証が txCtx（tx スコープ）で行われることを検証する。
		// ここでは「ロック内再検証が走ること」「再検証が空なら送信しない」を固定する。
		ownerID := uint64(30)
		petID := uint64(9001)
		var firstCallDone bool
		var lockHeldDuringRequery bool
		var addTagCalls int
		apptRepo := &mockDeliveryAppointmentRepository{
			findOwnerIDsWithTentativeFn: func(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
				return []uint64{ownerID}, nil
			},
			findTentativesByOwnerFn: func(_ context.Context, _, _ uint64, _ time.Time) ([]model.Reservation, error) {
				if !firstCallDone {
					firstCallDone = true
					return []model.Reservation{{ID: 501, OwnerID: &ownerID, PetID: &petID, Status: model.ReservationStatusTentative}}, nil
				}
				// ロック保持中の再検証 —— lockHeldDuringRequery が true のはず。
				assert.True(t, lockHeldDuringRequery, "再検証は serialization lock のスコープ内で実行される")
				return nil, nil // ロック中に確認すると tentative は消えている（昇格済み）
			},
			withReminderSerializationFn: func(ctx context.Context, _ uint64, _ []uint64, fn func(txCtx context.Context) error) error {
				lockHeldDuringRequery = true
				defer func() { lockHeldDuringRequery = false }()
				return fn(ctx)
			},
		}
		ownerRepo := &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return defaultOwnerWithLine(id), nil
			},
		}
		svc := buildDeliverySvc(ownerRepo, nil, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings(), apptRepo)
		injectTestClient(svc, &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, _ string) error {
				addTagCalls++
				return nil
			},
		})

		count, errs := svc.TriggerNextVisitTentativeReminder3D(context.Background(), 1, asOf)

		assert.Equal(t, 0, count, "ロック中の再検証で tentative が消えていれば送信しない")
		assert.Empty(t, errs)
		assert.Equal(t, 0, addTagCalls)
	})

	t.Run("nil appointment repo fails closed", func(t *testing.T) {
		// EMR-308: appointment 状態が source of truth の tentative リマインダーは
		// repo 未注入時に黙ってスキップすると配信漏れが検知できないためエラーを返す。
		svc := buildDeliverySvc(&mockOwnerRepoForDelivery{}, nil, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
		injectTestClient(svc, &mockLstepClientForDelivery{})

		count3d, errs3d := svc.TriggerNextVisitTentativeReminder3D(context.Background(), 1, asOf)
		count0d, errs0d := svc.TriggerNextVisitTentativeReminder0D(context.Background(), 1, asOf)
		assert.Equal(t, 0, count3d)
		assert.Equal(t, 0, count0d)
		assert.Len(t, errs3d, 1)
		assert.Len(t, errs0d, 1)
	})

	t.Run("re-locks the expanded pet set when the locked re-fetch grows (Codex #22 rereview)", func(t *testing.T) {
		// 初回候補の列挙〜ロック取得の間に別 pet の tentative が commit されたケース。
		// ロック中の再検索がロック済み集合外の pet を返したら送信せず、拡大後の
		// 集合でロックを取り直して収束すれば送信する。
		ownerID := uint64(30)
		petA, petB := uint64(9001), uint64(9002)
		bothPets := []model.Reservation{
			{ID: 501, OwnerID: &ownerID, PetID: &petA, Status: model.ReservationStatusTentative},
			{ID: 502, OwnerID: &ownerID, PetID: &petB, Status: model.ReservationStatusTentative},
		}
		inLock := false
		var lockSets [][]uint64
		var addTagCalls int
		apptRepo := &mockDeliveryAppointmentRepository{
			findOwnerIDsWithTentativeFn: func(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
				return []uint64{ownerID}, nil
			},
			findTentativesByOwnerFn: func(_ context.Context, _, _ uint64, _ time.Time) ([]model.Reservation, error) {
				if !inLock && len(lockSets) == 0 {
					// 1回目の候補列挙は petA のみ（petB はこの後に commit される想定）。
					return bothPets[:1], nil
				}
				// ロック中の再検索・2回目以降の列挙は petA+petB を返す。
				return bothPets, nil
			},
			withReminderSerializationFn: func(ctx context.Context, _ uint64, petIDs []uint64, fn func(txCtx context.Context) error) error {
				lockSets = append(lockSets, append([]uint64(nil), petIDs...))
				inLock = true
				defer func() { inLock = false }()
				return fn(ctx)
			},
		}
		ownerRepo := &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return defaultOwnerWithLine(id), nil
			},
		}
		svc := buildDeliverySvc(ownerRepo, nil, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings(), apptRepo)
		injectTestClient(svc, &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, _ string) error {
				addTagCalls++
				return nil
			},
		})

		count, errs := svc.TriggerNextVisitTentativeReminder3D(context.Background(), 1, asOf)

		assert.Equal(t, 1, count)
		assert.Empty(t, errs)
		require.Len(t, lockSets, 2, "集合拡大を検出してロックを取り直す")
		assert.ElementsMatch(t, []uint64{petA}, lockSets[0], "1回目のロックは列挙時点の集合")
		assert.ElementsMatch(t, []uint64{petA, petB}, lockSets[1], "2回目は拡大後の集合で再ロック")
		assert.Equal(t, 1, addTagCalls, "収束後の送信は1回だけ")
	})

	t.Run("re-locks when the locked re-fetch returns a row without pet id (Codex #22 rereview)", func(t *testing.T) {
		// ロック中の再検索が pet 未解決行（pet_id NULL）を含む場合、その行の mutation は
		// (clinic,pet) advisory lock で直列化できないため送信根拠にしない。行が収束した
		// 2回目の attempt で送信する。
		ownerID := uint64(30)
		petA := uint64(9001)
		inLock := false
		var lockCalls int
		var addTagCalls int
		apptRepo := &mockDeliveryAppointmentRepository{
			findOwnerIDsWithTentativeFn: func(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
				return []uint64{ownerID}, nil
			},
			findTentativesByOwnerFn: func(_ context.Context, _, _ uint64, _ time.Time) ([]model.Reservation, error) {
				if inLock && lockCalls == 1 {
					// 1回目のロック中再検索だけ pet 未解決行が混ざる。
					return []model.Reservation{
						{ID: 501, OwnerID: &ownerID, PetID: &petA, Status: model.ReservationStatusTentative},
						{ID: 599, OwnerID: &ownerID, PetID: nil, Status: model.ReservationStatusTentative},
					}, nil
				}
				return []model.Reservation{
					{ID: 501, OwnerID: &ownerID, PetID: &petA, Status: model.ReservationStatusTentative},
				}, nil
			},
			withReminderSerializationFn: func(ctx context.Context, _ uint64, _ []uint64, fn func(txCtx context.Context) error) error {
				lockCalls++
				inLock = true
				defer func() { inLock = false }()
				return fn(ctx)
			},
		}
		ownerRepo := &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return defaultOwnerWithLine(id), nil
			},
		}
		svc := buildDeliverySvc(ownerRepo, nil, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings(), apptRepo)
		injectTestClient(svc, &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, _ string) error {
				addTagCalls++
				return nil
			},
		})

		count, errs := svc.TriggerNextVisitTentativeReminder3D(context.Background(), 1, asOf)

		assert.Equal(t, 1, count)
		assert.Empty(t, errs)
		assert.Equal(t, 2, lockCalls, "pet 未解決行を検出してロックを取り直す")
		assert.Equal(t, 1, addTagCalls)
	})

	t.Run("fails closed after max lock attempts under non-converging candidate churn", func(t *testing.T) {
		// candidate 集合が拡大し続ける異常 churn: 上限まで再ロックを試み、収束しない
		// 場合は未送信のまま per-owner error で表面化し claim 済みスロットを failed にする。
		ownerID := uint64(30)
		petA, petB := uint64(9001), uint64(9002)
		inLock := false
		var lockCalls int
		var addTagCalls int
		var capturedStatus string
		apptRepo := &mockDeliveryAppointmentRepository{
			findOwnerIDsWithTentativeFn: func(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
				return []uint64{ownerID}, nil
			},
			findTentativesByOwnerFn: func(_ context.Context, _, _ uint64, _ time.Time) ([]model.Reservation, error) {
				if !inLock {
					// ロック前の列挙は常に petA のみ → ロック集合は収束しない。
					return []model.Reservation{
						{ID: 501, OwnerID: &ownerID, PetID: &petA, Status: model.ReservationStatusTentative},
					}, nil
				}
				// ロック中の再検索は毎回ロック外の petB を返し続ける。
				return []model.Reservation{
					{ID: 501, OwnerID: &ownerID, PetID: &petA, Status: model.ReservationStatusTentative},
					{ID: 502, OwnerID: &ownerID, PetID: &petB, Status: model.ReservationStatusTentative},
				}, nil
			},
			withReminderSerializationFn: func(ctx context.Context, _ uint64, _ []uint64, fn func(txCtx context.Context) error) error {
				lockCalls++
				inLock = true
				defer func() { inLock = false }()
				return fn(ctx)
			},
		}
		ownerRepo := &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return defaultOwnerWithLine(id), nil
			},
		}
		logRepo := &mockDeliveryTriggerLogRepository{
			updateStatusFn: func(_ context.Context, _ uint64, status string, _ *time.Time, _ *string) error {
				capturedStatus = status
				return nil
			},
		}
		svc := buildDeliverySvc(ownerRepo, nil, nil, nil, &mockTagCacheRepoForDelivery{}, logRepo, enabledSettings(), apptRepo)
		injectTestClient(svc, &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, _ string) error {
				addTagCalls++
				return nil
			},
		})

		count, errs := svc.TriggerNextVisitTentativeReminder3D(context.Background(), 1, asOf)

		assert.Equal(t, 0, count)
		assert.Len(t, errs, 1, "非収束 churn は per-owner error で表面化")
		assert.Equal(t, maxTentativeReminderLockAttempts, lockCalls, "上限回数だけ再ロックを試みる")
		assert.Equal(t, 0, addTagCalls, "未送信")
		assert.Equal(t, model.TriggerStatusFailed, capturedStatus, "claim 済みスロットを failed で解放")
	})
}

func TestLstepDeliveryTriggerService_TriggerVaccineDeadlines(t *testing.T) {
	asOf := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		triggerFn   func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error)
		expectedDay time.Time
	}{
		{
			name: "60d deadline triggers for 60 days from now",
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerVaccineDeadline60(ctx, 1, asOf)
			},
			expectedDay: asOf.AddDate(0, 0, 60),
		},
		{
			name: "30d deadline triggers for 30 days from now",
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerVaccineDeadline30(ctx, 1, asOf)
			},
			expectedDay: asOf.AddDate(0, 0, 30),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedDate time.Time
			vacRepo := &mockVaccinationRepoForDelivery{
				findOwnersByVaccineDeadlineFn: func(_ context.Context, _ uint64, targetDate time.Time) ([]uint64, error) {
					capturedDate = targetDate
					return []uint64{40}, nil
				},
			}
			ownerRepo := &mockOwnerRepoForDelivery{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return defaultOwnerWithLine(id), nil
				},
			}
			svc := buildDeliverySvc(ownerRepo, nil, vacRepo, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
			injectTestClient(svc, &mockLstepClientForDelivery{})
			count, errs := tt.triggerFn(svc, context.Background())
			assert.Equal(t, 1, count)
			assert.Empty(t, errs)
			assert.Equal(t, tt.expectedDay.Format("2006-01-02"), capturedDate.Format("2006-01-02"))
		})
	}
}

func TestLstepDeliveryTriggerService_TriggerBirthdayMessage(t *testing.T) {
	asOf := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)

	var capturedMonth, capturedDay int
	petRepo := &mockPetRepoForDelivery{
		findOwnersByPetBirthdayFn: func(_ context.Context, _ uint64, month, day int) ([]uint64, error) {
			capturedMonth = month
			capturedDay = day
			return []uint64{50}, nil
		},
	}
	ownerRepo := &mockOwnerRepoForDelivery{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
			return defaultOwnerWithLine(id), nil
		},
	}
	svc := buildDeliverySvc(ownerRepo, nil, nil, petRepo, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
	injectTestClient(svc, &mockLstepClientForDelivery{})
	count, errs := svc.TriggerBirthdayMessage(context.Background(), 1, asOf)
	assert.Equal(t, 1, count)
	assert.Empty(t, errs)
	assert.Equal(t, 5, capturedMonth)
	assert.Equal(t, 10, capturedDay)
}

func TestLstepDeliveryTriggerService_TriggerDormantPrevention(t *testing.T) {
	asOf := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		triggerFn func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error)
		wantDays  int
	}{
		{
			name: "180d dormant",
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerDormantPrevention180(ctx, 1, asOf)
			},
			wantDays: 180,
		},
		{
			name: "210d dormant",
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerDormantPrevention210(ctx, 1, asOf)
			},
			wantDays: 210,
		},
		{
			name: "240d dormant",
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerDormantPrevention240(ctx, 1, asOf)
			},
			wantDays: 240,
		},
		{
			name: "365d dormant",
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerDormantPrevention365(ctx, 1, asOf)
			},
			wantDays: 365,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedDays int
			medRepo := &mockMedicalRecordRepository{
				findOwnersByLastVisitDaysFn: func(_ context.Context, _ uint64, exactDays int, _ time.Time) ([]uint64, error) {
					capturedDays = exactDays
					return []uint64{60}, nil
				},
			}
			ownerRepo := &mockOwnerRepoForDelivery{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return defaultOwnerWithLine(id), nil
				},
			}
			svc := buildDeliverySvc(ownerRepo, medRepo, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
			injectTestClient(svc, &mockLstepClientForDelivery{})
			count, errs := tt.triggerFn(svc, context.Background())
			assert.Equal(t, 1, count)
			assert.Empty(t, errs)
			assert.Equal(t, tt.wantDays, capturedDays)
		})
	}
}

func TestLstepDeliveryTriggerService_TriggerTagBased(t *testing.T) {
	tests := []struct {
		name      string
		tagName   string
		triggerFn func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error)
	}{
		{
			name:    "filaria alert uses correct tag",
			tagName: tagFilariaAlert,
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerFilariaAlert(ctx, 1, time.Now())
			},
		},
		{
			name:    "flea tick alert uses correct tag",
			tagName: tagFleaTickAlert,
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerFleaTickAlert(ctx, 1, time.Now())
			},
		},
		{
			name:    "food refill reminder uses correct tag",
			tagName: tagFoodRefill,
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerFoodRefillReminder(ctx, 1, time.Now())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedTagName string
			tagRepo := &mockTagCacheRepoForDelivery{
				findOwnerIDsByTagFn: func(_ context.Context, _ uint64, tagName string) ([]uint64, error) {
					capturedTagName = tagName
					return []uint64{70}, nil
				},
			}
			ownerRepo := &mockOwnerRepoForDelivery{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return defaultOwnerWithLine(id), nil
				},
			}
			svc := buildDeliverySvc(ownerRepo, nil, nil, nil, tagRepo, &mockDeliveryTriggerLogRepository{}, enabledSettings())
			injectTestClient(svc, &mockLstepClientForDelivery{})
			count, errs := tt.triggerFn(svc, context.Background())
			assert.Equal(t, 1, count)
			assert.Empty(t, errs)
			assert.Equal(t, tt.tagName, capturedTagName)
		})
	}
}

func TestLstepDeliveryTriggerService_SyncDisabled(t *testing.T) {
	medRepo := &mockMedicalRecordRepository{
		findOwnersByFirstVisitFn: func(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
			return []uint64{10}, nil
		},
	}
	svc := buildDeliverySvc(&mockOwnerRepoForDelivery{}, medRepo, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, disabledSettings())
	count, errs := svc.TriggerFirstVisitFollowUp3D(context.Background(), 1, time.Now())
	assert.Equal(t, 0, count)
	assert.Empty(t, errs)
}

func TestLstepDeliveryTriggerService_ExclTagStopsDelivery(t *testing.T) {
	ownerID := uint64(80)
	lineID := "U_excl_tag"
	medRepo := &mockMedicalRecordRepository{
		findOwnersByFirstVisitFn: func(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
			return []uint64{ownerID}, nil
		},
	}
	ownerRepo := &mockOwnerRepoForDelivery{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
			return &model.Owner{ID: id, ClinicID: 1, LineUserID: &lineID}, nil
		},
	}
	tagRepo := &mockTagCacheRepoForDelivery{
		findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
			return []*model.LstepTagCache{
				{TagName: exclTagDeliveryStop},
			}, nil
		},
	}
	var statusUpdated string
	logRepo := &mockDeliveryTriggerLogRepository{
		updateStatusFn: func(_ context.Context, _ uint64, status string, _ *time.Time, _ *string) error {
			statusUpdated = status
			return nil
		},
	}
	svc := buildDeliverySvc(ownerRepo, medRepo, nil, nil, tagRepo, logRepo, enabledSettings())
	count, errs := svc.TriggerFirstVisitFollowUp3D(context.Background(), 1, time.Now())
	assert.Equal(t, 0, count)
	assert.Empty(t, errs)
	assert.Equal(t, model.TriggerStatusExcluded, statusUpdated)
}

func TestLstepDeliveryTriggerService_AlreadyFiredSkips(t *testing.T) {
	medRepo := &mockMedicalRecordRepository{
		findOwnersByFirstVisitFn: func(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
			return []uint64{10}, nil
		},
	}
	ownerRepo := &mockOwnerRepoForDelivery{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
			return defaultOwnerWithLine(id), nil
		},
	}
	logRepo := &mockDeliveryTriggerLogRepository{
		existsTodayFn: func(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
			return true, nil
		},
	}
	svc := buildDeliverySvc(ownerRepo, medRepo, nil, nil, &mockTagCacheRepoForDelivery{}, logRepo, enabledSettings())
	count, errs := svc.TriggerFirstVisitFollowUp3D(context.Background(), 1, time.Now())
	assert.Equal(t, 0, count)
	assert.Empty(t, errs)
}

func TestLstepDeliveryTriggerService_TriggerLogRecorded(t *testing.T) {
	var createdLog *model.LstepDeliveryTriggerLog
	logRepo := &mockDeliveryTriggerLogRepository{
		createFn: func(_ context.Context, log *model.LstepDeliveryTriggerLog) error {
			log.ID = 99
			createdLog = log
			return nil
		},
	}
	medRepo := &mockMedicalRecordRepository{
		findOwnersByFirstVisitFn: func(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
			return []uint64{10}, nil
		},
	}
	ownerRepo := &mockOwnerRepoForDelivery{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
			return defaultOwnerWithLine(id), nil
		},
	}
	svc := buildDeliverySvc(ownerRepo, medRepo, nil, nil, &mockTagCacheRepoForDelivery{}, logRepo, enabledSettings())
	injectTestClient(svc, &mockLstepClientForDelivery{})
	_, _ = svc.TriggerFirstVisitFollowUp3D(context.Background(), 1, time.Now())
	assert.NotNil(t, createdLog)
	assert.Equal(t, model.TriggerTypeFirstVisitFollowUp3D, createdLog.TriggerType)
	assert.Equal(t, model.TriggerStatusScheduled, createdLog.Status)
	assert.Equal(t, uint64(10), createdLog.OwnerID)
}

func TestLstepDeliveryTriggerService_RepoErrorDoesNotStopOtherOwners(t *testing.T) {
	callCount := 0
	medRepo := &mockMedicalRecordRepository{
		findOwnersByFirstVisitFn: func(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
			return []uint64{10, 11}, nil
		},
	}
	ownerRepo := &mockOwnerRepoForDelivery{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
			callCount++
			if id == 10 {
				return nil, apperrors.WrapNotFound("owner", "10")
			}
			return defaultOwnerWithLine(id), nil
		},
	}
	svc := buildDeliverySvc(ownerRepo, medRepo, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
	injectTestClient(svc, &mockLstepClientForDelivery{})
	count, errs := svc.TriggerFirstVisitFollowUp3D(context.Background(), 1, time.Now())
	// owner 10 fails (NotFound) but owner 11 succeeds
	assert.Equal(t, 1, count)
	assert.Len(t, errs, 1)
}

func TestLstepDeliveryTriggerService_TriggerFirstVisitWelcome(t *testing.T) {
	ownerID := uint64(10)

	t.Run("sync disabled returns nil (noop)", func(t *testing.T) {
		svc := buildDeliverySvc(&mockOwnerRepoForDelivery{}, &mockMedicalRecordRepository{}, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, disabledSettings())
		assert.NoError(t, svc.TriggerFirstVisitWelcome(context.Background(), 1, ownerID))
	})

	t.Run("happy path fires tag for owner with line_user_id", func(t *testing.T) {
		ownerRepo := &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return defaultOwnerWithLine(id), nil
			},
		}
		var firedTag string
		client := &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, tagName string) error {
				firedTag = tagName
				return nil
			},
		}
		svc := buildDeliverySvc(ownerRepo, &mockMedicalRecordRepository{}, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
		injectTestClient(svc, client)
		err := svc.TriggerFirstVisitWelcome(context.Background(), 1, ownerID)
		assert.NoError(t, err)
		assert.Equal(t, model.TriggerTypeFirstVisitWelcome, firedTag)
	})

	t.Run("buildClient error is returned", func(t *testing.T) {
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) {
				return false, errors.New("settings error")
			},
		}
		svc := buildDeliverySvc(&mockOwnerRepoForDelivery{}, &mockMedicalRecordRepository{}, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, settingsSvc)
		err := svc.TriggerFirstVisitWelcome(context.Background(), 1, ownerID)
		assert.Error(t, err)
	})

	t.Run("AddTag error is returned from processSingleOwner", func(t *testing.T) {
		ownerRepo := &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return defaultOwnerWithLine(id), nil
			},
		}
		client := &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, _ string) error {
				return errors.New("lstep api error")
			},
		}
		svc := buildDeliverySvc(ownerRepo, &mockMedicalRecordRepository{}, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
		injectTestClient(svc, client)
		err := svc.TriggerFirstVisitWelcome(context.Background(), 1, ownerID)
		assert.Error(t, err)
	})
}

func TestLstepDeliveryTriggerService_TriggerCheckupFollowUp(t *testing.T) {
	ownerID := uint64(20)

	t.Run("sync disabled returns nil (noop)", func(t *testing.T) {
		svc := buildDeliverySvc(&mockOwnerRepoForDelivery{}, &mockMedicalRecordRepository{}, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, disabledSettings())
		assert.NoError(t, svc.TriggerCheckupFollowUp(context.Background(), 1, ownerID))
	})

	t.Run("happy path fires tag for owner with line_user_id", func(t *testing.T) {
		ownerRepo := &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return defaultOwnerWithLine(id), nil
			},
		}
		var firedTag string
		client := &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, tagName string) error {
				firedTag = tagName
				return nil
			},
		}
		svc := buildDeliverySvc(ownerRepo, &mockMedicalRecordRepository{}, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
		injectTestClient(svc, client)
		err := svc.TriggerCheckupFollowUp(context.Background(), 1, ownerID)
		assert.NoError(t, err)
		assert.Equal(t, model.TriggerTypeCheckupFollowUp, firedTag)
	})

	t.Run("buildClient error is returned", func(t *testing.T) {
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) {
				return false, errors.New("settings error")
			},
		}
		svc := buildDeliverySvc(&mockOwnerRepoForDelivery{}, &mockMedicalRecordRepository{}, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, settingsSvc)
		err := svc.TriggerCheckupFollowUp(context.Background(), 1, ownerID)
		assert.Error(t, err)
	})

	t.Run("AddTag error is returned from processSingleOwner", func(t *testing.T) {
		ownerRepo := &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return defaultOwnerWithLine(id), nil
			},
		}
		client := &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, _ string) error {
				return errors.New("lstep api error")
			},
		}
		svc := buildDeliverySvc(ownerRepo, &mockMedicalRecordRepository{}, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
		injectTestClient(svc, client)
		err := svc.TriggerCheckupFollowUp(context.Background(), 1, ownerID)
		assert.Error(t, err)
	})
}

func TestLstepDeliveryTriggerService_EmptyOwnerListIsNoop(t *testing.T) {
	medRepo := &mockMedicalRecordRepository{
		findOwnersByFirstVisitFn: func(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
			return []uint64{}, nil
		},
	}
	svc := buildDeliverySvc(&mockOwnerRepoForDelivery{}, medRepo, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
	count, errs := svc.TriggerFirstVisitFollowUp3D(context.Background(), 1, time.Now())
	assert.Equal(t, 0, count)
	assert.Empty(t, errs)
}

// TestTriggerSyncTagConstantsAlignment は sync service と trigger service のタグ定数が一致することを保証する回帰テスト。
// FEAT-379 命名移行で sync 側だけ更新して trigger 側が取り残されたバグ (P0) を再発防止する。
func TestTriggerSyncTagConstantsAlignment(t *testing.T) {
	assert.Equal(t, PrevFilariaTag, tagFilariaAlert, "filaria tag must match between sync and trigger")
	assert.Equal(t, PrevFleaTickTag, tagFleaTickAlert, "flea tick tag must match between sync and trigger")
	assert.Equal(t, LtvFoodPurchaseTag, tagFoodRefill, "food purchase tag must match between sync and trigger")
}

// TestLstepDeliveryTriggerService_FindOwnersErrorPropagates は各トリガーメソッドの
// owner検索リポジトリエラー時に count=0・エラー1件でラップ返却されることを検証する。
func TestLstepDeliveryTriggerService_FindOwnersErrorPropagates(t *testing.T) {
	asOf := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	repoErr := errors.New("db error")

	tests := []struct {
		name      string
		triggerFn func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error)
	}{
		{
			name: "TriggerFirstVisitFollowUp7D",
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerFirstVisitFollowUp7D(ctx, 1, asOf)
			},
		},
		{
			name: "TriggerNextVisitTentativeReminder3D",
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerNextVisitTentativeReminder3D(ctx, 1, asOf)
			},
		},
		{
			name: "TriggerVaccineDeadline60",
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerVaccineDeadline60(ctx, 1, asOf)
			},
		},
		{
			name: "TriggerVaccineDeadline30",
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerVaccineDeadline30(ctx, 1, asOf)
			},
		},
		{
			name: "TriggerBirthdayMessage",
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerBirthdayMessage(ctx, 1, asOf)
			},
		},
		{
			name: "TriggerDormantPrevention180",
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerDormantPrevention180(ctx, 1, asOf)
			},
		},
		{
			name: "TriggerDormantPrevention210",
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerDormantPrevention210(ctx, 1, asOf)
			},
		},
		{
			name: "TriggerDormantPrevention240",
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerDormantPrevention240(ctx, 1, asOf)
			},
		},
		{
			name: "TriggerDormantPrevention365",
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerDormantPrevention365(ctx, 1, asOf)
			},
		},
		{
			name: "TriggerFilariaAlert",
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerFilariaAlert(ctx, 1, asOf)
			},
		},
		{
			name: "TriggerFleaTickAlert",
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerFleaTickAlert(ctx, 1, asOf)
			},
		},
		{
			name: "TriggerFoodRefillReminder",
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerFoodRefillReminder(ctx, 1, asOf)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			medRepo := &mockMedicalRecordRepository{
				findOwnersByFirstVisitFn: func(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
					return nil, repoErr
				},
				findOwnersByLastVisitDaysFn: func(_ context.Context, _ uint64, _ int, _ time.Time) ([]uint64, error) {
					return nil, repoErr
				},
			}
			vacRepo := &mockVaccinationRepoForDelivery{
				findOwnersByVaccineDeadlineFn: func(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
					return nil, repoErr
				},
			}
			petRepo := &mockPetRepoForDelivery{
				findOwnersByPetBirthdayFn: func(_ context.Context, _ uint64, _, _ int) ([]uint64, error) {
					return nil, repoErr
				},
			}
			tagRepo := &mockTagCacheRepoForDelivery{
				findOwnerIDsByTagFn: func(_ context.Context, _ uint64, _ string) ([]uint64, error) {
					return nil, repoErr
				},
			}
			apptRepo := &mockDeliveryAppointmentRepository{
				findOwnerIDsWithTentativeFn: func(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
					return nil, repoErr
				},
			}
			svc := buildDeliverySvc(&mockOwnerRepoForDelivery{}, medRepo, vacRepo, petRepo, tagRepo, &mockDeliveryTriggerLogRepository{}, enabledSettings(), apptRepo)
			injectTestClient(svc, &mockLstepClientForDelivery{})

			count, errs := tt.triggerFn(svc, context.Background())

			assert.Equal(t, 0, count)
			assert.Len(t, errs, 1)
		})
	}
}

// TestLstepDeliveryTriggerService_DormantThresholdsErrorPropagates は dormant 系トリガーの
// settingsSvc.GetDormantThresholds エラー時にエラーが伝播することを検証する（他トリガーには無い分岐）。
func TestLstepDeliveryTriggerService_DormantThresholdsErrorPropagates(t *testing.T) {
	asOf := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	thresholdsErr := errors.New("settings db error")
	settingsSvc := &mockLstepSettingsService{
		getDormantThresholdsFn: func(_ context.Context, _ uint64) (model.DormantThresholds, error) {
			return model.DormantThresholds{}, thresholdsErr
		},
	}

	tests := []struct {
		name      string
		triggerFn func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error)
	}{
		{
			name: "TriggerDormantPrevention180",
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerDormantPrevention180(ctx, 1, asOf)
			},
		},
		{
			name: "TriggerDormantPrevention210",
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerDormantPrevention210(ctx, 1, asOf)
			},
		},
		{
			name: "TriggerDormantPrevention240",
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerDormantPrevention240(ctx, 1, asOf)
			},
		},
		{
			name: "TriggerDormantPrevention365",
			triggerFn: func(svc LstepDeliveryTriggerService, ctx context.Context) (int, []error) {
				return svc.TriggerDormantPrevention365(ctx, 1, asOf)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := buildDeliverySvc(&mockOwnerRepoForDelivery{}, &mockMedicalRecordRepository{}, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, settingsSvc)
			injectTestClient(svc, &mockLstepClientForDelivery{})

			count, errs := tt.triggerFn(svc, context.Background())

			assert.Equal(t, 0, count)
			assert.Len(t, errs, 1)
		})
	}
}
