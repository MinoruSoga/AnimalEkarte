package lstep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mockLstepTagCodeMappingRepository ----

type mockLstepTagCodeMappingRepository struct {
	findAllByClinicIDFn        func(ctx context.Context, clinicID uint64) ([]*model.LstepTagCodeMapping, error)
	findByClinicIDAndTagNameFn func(ctx context.Context, clinicID uint64, tagName string) ([]*model.LstepTagCodeMapping, error)
	createFn                   func(ctx context.Context, mapping *model.LstepTagCodeMapping) error
	softDeleteFn               func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockLstepTagCodeMappingRepository) FindAllByClinicID(ctx context.Context, clinicID uint64) ([]*model.LstepTagCodeMapping, error) {
	if m.findAllByClinicIDFn != nil {
		return m.findAllByClinicIDFn(ctx, clinicID)
	}
	return nil, nil
}
func (m *mockLstepTagCodeMappingRepository) FindByClinicIDAndTagName(ctx context.Context, clinicID uint64, tagName string) ([]*model.LstepTagCodeMapping, error) {
	if m.findByClinicIDAndTagNameFn != nil {
		return m.findByClinicIDAndTagNameFn(ctx, clinicID, tagName)
	}
	return nil, nil
}
func (m *mockLstepTagCodeMappingRepository) Create(ctx context.Context, mapping *model.LstepTagCodeMapping) error {
	if m.createFn != nil {
		return m.createFn(ctx, mapping)
	}
	return nil
}
func (m *mockLstepTagCodeMappingRepository) SoftDelete(ctx context.Context, clinicID, id uint64) error {
	if m.softDeleteFn != nil {
		return m.softDeleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockLstepTagCodeMappingRepository) SoftDeleteByClinicIDAndTagName(_ context.Context, _ uint64, _ string) error {
	return nil
}

// ---- テストヘルパー ----

func defaultOwnerWithLineID() *model.Owner {
	lineUserID := "line-u-test-001"
	return &model.Owner{ID: 1, ClinicID: 10, LstepOptOut: false, LineUserID: &lineUserID}
}

func ownerRepoReturning(owner *model.Owner) *mockOwnerRepository {
	return &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return owner, nil
		},
	}
}

// buildHealthSvc は健診・予防・物販タグテスト用にサービスを構築する。
// settingsSvc: IsSyncEnabled=true, GetRawCredentials="" → buildClient→nil (APIコールなし)
// SyncXxxWithMappings（interface 外の非公開メソッド群）を直接呼ぶため具象型を返す
// （B-3: 呼び出し元ゼロの公開ラッパー削除に伴う変更）。
func buildHealthSvc(
	tagCodeRepo LstepTagCodeMappingRepository,
	ownerRepo tagSyncOwnerRepo,
	checkupRepo tagSyncCheckupRepo,
	medRecordRepo tagSyncMedicalRecordRepo,
	petRepo tagSyncPetRepo,
	billingItemRepo tagSyncBillingItemRepo,
) *lstepTagSyncService {
	settingsSvc := &mockLstepSettingsService{}
	var tagCache LstepTagCacheRepository = &mockLstepTagCacheRepository{}
	return NewLstepTagSyncService(
		settingsSvc,
		ownerRepo,
		nil, // vacRepo
		medRecordRepo,
		nil, // accountRepo
		tagCache,
		petRepo,
		nil, // prescriptionRepo
		checkupRepo,
		nil, // errorCounterRepo — client=nil なので notifyAPIFailure は到達しない
		tagCodeRepo,
		billingItemRepo,
		nil, // tagConfigRepo
	).(*lstepTagSyncService)
}

func mappingWith(tagName, codeType string, codes []string) *mockLstepTagCodeMappingRepository {
	return &mockLstepTagCodeMappingRepository{
		findByClinicIDAndTagNameFn: func(_ context.Context, _ uint64, tn string) ([]*model.LstepTagCodeMapping, error) {
			if tn == tagName {
				return []*model.LstepTagCodeMapping{{
					TagName:  tagName,
					CodeType: codeType,
					Codes:    pq.StringArray(codes),
				}}, nil
			}
			return nil, nil
		},
	}
}

func checkupRepoWithResult(checkups []model.Checkup, err error) *mockCheckupRepository {
	return &mockCheckupRepository{
		findByOwnerIDFn: func(_ context.Context, _, _ uint64) ([]model.Checkup, error) {
			return checkups, err
		},
	}
}

func visitSummaryRepo(annualCount int64) *mockMedicalRecordRepository {
	return &mockMedicalRecordRepository{
		findOwnerVisitSummaryFn: func(_ context.Context, _, _ uint64) (*medicalrecord.OwnerVisitSummary, error) {
			return &medicalrecord.OwnerVisitSummary{AnnualCount: annualCount}, nil
		},
	}
}

func petRepoWithPets(pets []model.Pet, err error) *mockPetRepository {
	return &mockPetRepository{
		findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) {
			return pets, err
		},
	}
}

func billingItemRepoReturning(hasItem, hasFood bool) *mockBillingItemRepository {
	return &mockBillingItemRepository{
		hasItemByOwnerSinceFn: func(_ context.Context, _, _ uint64, _ time.Time, _ []string) (bool, error) {
			return hasItem, nil
		},
		hasFoodPurchaseByOwnerSinceFn: func(_ context.Context, _, _ uint64, _ time.Time, _ []string) (bool, error) {
			return hasFood, nil
		},
	}
}

func recentCheckup(codeName string) model.Checkup {
	return model.Checkup{
		Date:        time.Now().AddDate(0, 0, -30), // 30日前 = lookback内
		CheckupType: &model.CheckupType{Name: codeName},
	}
}

func oldCheckup(codeName string) model.Checkup {
	return model.Checkup{
		Date:        time.Now().AddDate(0, 0, -(model.DefaultHealthPreventionLookbackDays + 10)), // lookback外
		CheckupType: &model.CheckupType{Name: codeName},
	}
}

func dogPet() model.Pet {
	return model.Pet{ID: 1, AnimalSpecies: &model.AnimalSpecies{Name: "犬（チワワ）"}}
}

func catPet() model.Pet {
	return model.Pet{ID: 2, AnimalSpecies: &model.AnimalSpecies{Name: "猫（アメショ）"}}
}

// ---- TestExtractTagCodes ----

func TestExtractTagCodes(t *testing.T) {
	t.Run("空mappings→空", func(t *testing.T) {
		got := extractTagCodes(nil, model.CodeTypeCheckupType)
		assert.Empty(t, got)
	})
	t.Run("一致code_type→コード返却", func(t *testing.T) {
		m := []*model.LstepTagCodeMapping{{CodeType: model.CodeTypeCheckupType, Codes: pq.StringArray{"A", "B"}}}
		got := extractTagCodes(m, model.CodeTypeCheckupType)
		assert.Equal(t, []string{"A", "B"}, got)
	})
	t.Run("不一致code_type→空", func(t *testing.T) {
		m := []*model.LstepTagCodeMapping{{CodeType: model.CodeTypePrescription, Codes: pq.StringArray{"RX1"}}}
		got := extractTagCodes(m, model.CodeTypeCheckupType)
		assert.Empty(t, got)
	})
	t.Run("複数エントリ同一type→連結", func(t *testing.T) {
		m := []*model.LstepTagCodeMapping{
			{CodeType: model.CodeTypeCheckupType, Codes: pq.StringArray{"A"}},
			{CodeType: model.CodeTypeCheckupType, Codes: pq.StringArray{"B", "C"}},
		}
		got := extractTagCodes(m, model.CodeTypeCheckupType)
		assert.Equal(t, []string{"A", "B", "C"}, got)
	})
	t.Run("混在type→一致のみ", func(t *testing.T) {
		m := []*model.LstepTagCodeMapping{
			{CodeType: model.CodeTypeCheckupType, Codes: pq.StringArray{"HC1"}},
			{CodeType: model.CodeTypePrescription, Codes: pq.StringArray{"RX1"}},
		}
		got := extractTagCodes(m, model.CodeTypeCheckupType)
		assert.Equal(t, []string{"HC1"}, got)
	})
}

// ---- TestStrSet ----

func TestStrSet(t *testing.T) {
	t.Run("空→空マップ", func(t *testing.T) {
		assert.Empty(t, strSet(nil))
	})
	t.Run("重複→dedup", func(t *testing.T) {
		m := strSet([]string{"a", "b", "a"})
		assert.Len(t, m, 2)
		_, hasA := m["a"]
		_, hasB := m["b"]
		assert.True(t, hasA)
		assert.True(t, hasB)
	})
}

// ---- TestSyncHealthcheckTags ----

func TestSyncHealthcheckTags(t *testing.T) {
	ctx := context.Background()
	const clinicID, ownerID uint64 = 10, 1

	t.Run("nil tagCodeRepo→noop", func(t *testing.T) {
		svc := buildHealthSvc(nil, ownerRepoReturning(defaultOwnerWithLineID()), nil, nil, nil, nil)
		assert.NoError(t, svc.SyncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("コードなし→noop", func(t *testing.T) {
		tagRepo := &mockLstepTagCodeMappingRepository{} // FindByClinicIDAndTagName → nil, nil
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()), nil, nil, nil, nil)
		assert.NoError(t, svc.SyncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("オプトアウト→noop", func(t *testing.T) {
		owner := &model.Owner{ID: 1, LstepOptOut: true}
		tagRepo := mappingWith(HlthHealthcheckDoneTag, model.CodeTypeCheckupType, []string{"健診A"})
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(owner), checkupRepoWithResult(nil, nil), nil, nil, nil)
		assert.NoError(t, svc.SyncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("LineUserIDなし→noop", func(t *testing.T) {
		owner := &model.Owner{ID: 1, LstepOptOut: false, LineUserID: nil}
		tagRepo := mappingWith(HlthHealthcheckDoneTag, model.CodeTypeCheckupType, []string{"健診A"})
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(owner), checkupRepoWithResult(nil, nil), nil, nil, nil)
		assert.NoError(t, svc.SyncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("lookback内で一致→hasHealthcheck=true→nil(client=nil)", func(t *testing.T) {
		tagRepo := mappingWith(HlthHealthcheckDoneTag, model.CodeTypeCheckupType, []string{"健診A"})
		checkups := []model.Checkup{recentCheckup("健診A")}
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(checkups, nil), nil, nil, nil)
		assert.NoError(t, svc.SyncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("lookback外の健診→hasHealthcheck=false→nil", func(t *testing.T) {
		tagRepo := mappingWith(HlthHealthcheckDoneTag, model.CodeTypeCheckupType, []string{"健診A"})
		checkups := []model.Checkup{oldCheckup("健診A")}
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(checkups, nil), nil, nil, nil)
		assert.NoError(t, svc.SyncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("コードが一致しない→hasHealthcheck=false→nil", func(t *testing.T) {
		tagRepo := mappingWith(HlthHealthcheckDoneTag, model.CodeTypeCheckupType, []string{"健診B"})
		checkups := []model.Checkup{recentCheckup("健診A")} // 別コード
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(checkups, nil), nil, nil, nil)
		assert.NoError(t, svc.SyncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("checkupRepo エラー→エラー返却", func(t *testing.T) {
		tagRepo := mappingWith(HlthHealthcheckDoneTag, model.CodeTypeCheckupType, []string{"健診A"})
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(nil, errors.New("db error")), nil, nil, nil)
		assert.Error(t, svc.SyncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("tagCodeRepo エラー→エラー返却", func(t *testing.T) {
		tagRepo := &mockLstepTagCodeMappingRepository{
			findByClinicIDAndTagNameFn: func(_ context.Context, _ uint64, _ string) ([]*model.LstepTagCodeMapping, error) {
				return nil, errors.New("db error")
			},
		}
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()), nil, nil, nil, nil)
		assert.Error(t, svc.SyncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("ownerRepo エラー→エラー返却", func(t *testing.T) {
		tagRepo := mappingWith(HlthHealthcheckDoneTag, model.CodeTypeCheckupType, []string{"健診A"})
		ownerRepo := &mockOwnerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return nil, errors.New("db error")
			},
		}
		svc := buildHealthSvc(tagRepo, ownerRepo, nil, nil, nil, nil)
		assert.Error(t, svc.SyncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, nil, nil))
	})
}

// ---- TestSyncAnnual4CheckupTag ----

func TestSyncAnnual4CheckupTag(t *testing.T) {
	ctx := context.Background()
	const clinicID, ownerID uint64 = 10, 1

	t.Run("nil tagCodeRepo→noop", func(t *testing.T) {
		svc := buildHealthSvc(nil, ownerRepoReturning(defaultOwnerWithLineID()), nil, nil, nil, nil)
		assert.NoError(t, svc.SyncAnnual4CheckupTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("コードなし→noop", func(t *testing.T) {
		svc := buildHealthSvc(&mockLstepTagCodeMappingRepository{},
			ownerRepoReturning(defaultOwnerWithLineID()), nil, nil, nil, nil)
		assert.NoError(t, svc.SyncAnnual4CheckupTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("healthcheck=true AnnualCount>=2→qualified→nil", func(t *testing.T) {
		tagRepo := mappingWith(HlthHealthcheckDoneTag, model.CodeTypeCheckupType, []string{"健診A"})
		checkups := []model.Checkup{recentCheckup("健診A")}
		medRepo := visitSummaryRepo(2)
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(checkups, nil), medRepo, nil, nil)
		assert.NoError(t, svc.SyncAnnual4CheckupTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("healthcheck=true AnnualCount<2→not qualified→nil", func(t *testing.T) {
		tagRepo := mappingWith(HlthHealthcheckDoneTag, model.CodeTypeCheckupType, []string{"健診A"})
		checkups := []model.Checkup{recentCheckup("健診A")}
		medRepo := visitSummaryRepo(1)
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(checkups, nil), medRepo, nil, nil)
		assert.NoError(t, svc.SyncAnnual4CheckupTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("healthcheck=false AnnualCount>=2→not qualified→nil", func(t *testing.T) {
		tagRepo := mappingWith(HlthHealthcheckDoneTag, model.CodeTypeCheckupType, []string{"健診A"})
		checkups := []model.Checkup{recentCheckup("健診B")} // コード不一致
		medRepo := visitSummaryRepo(5)
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(checkups, nil), medRepo, nil, nil)
		assert.NoError(t, svc.SyncAnnual4CheckupTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("medRecordRepo エラー→エラー返却", func(t *testing.T) {
		tagRepo := mappingWith(HlthHealthcheckDoneTag, model.CodeTypeCheckupType, []string{"健診A"})
		checkups := []model.Checkup{recentCheckup("健診A")}
		medRepo := &mockMedicalRecordRepository{
			findOwnerVisitSummaryFn: func(_ context.Context, _, _ uint64) (*medicalrecord.OwnerVisitSummary, error) {
				return nil, errors.New("db error")
			},
		}
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(checkups, nil), medRepo, nil, nil)
		assert.Error(t, svc.SyncAnnual4CheckupTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("checkupRepo エラー→エラー返却", func(t *testing.T) {
		tagRepo := mappingWith(HlthHealthcheckDoneTag, model.CodeTypeCheckupType, []string{"健診A"})
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(nil, errors.New("db error")), visitSummaryRepo(3), nil, nil)
		assert.Error(t, svc.SyncAnnual4CheckupTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})
}

// ---- TestSyncFilariaTag ----

func TestSyncFilariaTag(t *testing.T) {
	ctx := context.Background()
	const clinicID, ownerID uint64 = 10, 1

	t.Run("nil tagCodeRepo→noop", func(t *testing.T) {
		svc := buildHealthSvc(nil, ownerRepoReturning(defaultOwnerWithLineID()), nil, nil, nil, nil)
		assert.NoError(t, svc.SyncFilariaTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("testCodes/rxCodes両方空→noop", func(t *testing.T) {
		// FindByClinicIDAndTagName → 空スライス (CodeType が checkup_type でも prescription でもない)
		tagRepo := &mockLstepTagCodeMappingRepository{
			findByClinicIDAndTagNameFn: func(_ context.Context, _ uint64, _ string) ([]*model.LstepTagCodeMapping, error) {
				return []*model.LstepTagCodeMapping{
					{CodeType: model.CodeTypeMerchandiseItem, Codes: pq.StringArray{"food1"}},
				}, nil
			},
		}
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()), nil, nil, nil, nil)
		assert.NoError(t, svc.SyncFilariaTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("犬なし→noop(nil返却)", func(t *testing.T) {
		tagRepo := mappingWith(PrevFilariaTag, model.CodeTypeCheckupType, []string{"フィラリア検査"})
		petRepo := petRepoWithPets([]model.Pet{catPet()}, nil)
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(nil, nil), nil, petRepo, nil)
		assert.NoError(t, svc.SyncFilariaTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("犬あり+testDone=true→complete→nil", func(t *testing.T) {
		tagRepo := mappingWith(PrevFilariaTag, model.CodeTypeCheckupType, []string{"フィラリア検査"})
		checkups := []model.Checkup{recentCheckup("フィラリア検査")}
		petRepo := petRepoWithPets([]model.Pet{dogPet()}, nil)
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(checkups, nil), nil, petRepo, nil)
		assert.NoError(t, svc.SyncFilariaTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("犬あり+rxDone=true→complete→nil", func(t *testing.T) {
		tagRepo := mappingWith(PrevFilariaTag, model.CodeTypePrescription, []string{"フィラリア薬"})
		petRepo := petRepoWithPets([]model.Pet{dogPet()}, nil)
		billingRepo := billingItemRepoReturning(true, false)
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(nil, nil), nil, petRepo, billingRepo)
		assert.NoError(t, svc.SyncFilariaTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("犬あり+両方未完了→incomplete→nil(client=nil)", func(t *testing.T) {
		tagRepo := &mockLstepTagCodeMappingRepository{
			findByClinicIDAndTagNameFn: func(_ context.Context, _ uint64, _ string) ([]*model.LstepTagCodeMapping, error) {
				return []*model.LstepTagCodeMapping{
					{CodeType: model.CodeTypeCheckupType, Codes: pq.StringArray{"フィラリア検査"}},
					{CodeType: model.CodeTypePrescription, Codes: pq.StringArray{"フィラリア薬"}},
				}, nil
			},
		}
		checkups := []model.Checkup{recentCheckup("別検査")} // コード不一致
		petRepo := petRepoWithPets([]model.Pet{dogPet()}, nil)
		billingRepo := billingItemRepoReturning(false, false)
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(checkups, nil), nil, petRepo, billingRepo)
		assert.NoError(t, svc.SyncFilariaTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("petRepo エラー→エラー返却", func(t *testing.T) {
		tagRepo := mappingWith(PrevFilariaTag, model.CodeTypeCheckupType, []string{"フィラリア検査"})
		petRepo := petRepoWithPets(nil, errors.New("db error"))
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			nil, nil, petRepo, nil)
		assert.Error(t, svc.SyncFilariaTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("checkupRepo エラー→エラー返却", func(t *testing.T) {
		tagRepo := mappingWith(PrevFilariaTag, model.CodeTypeCheckupType, []string{"フィラリア検査"})
		petRepo := petRepoWithPets([]model.Pet{dogPet()}, nil)
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(nil, errors.New("db error")), nil, petRepo, nil)
		assert.Error(t, svc.SyncFilariaTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("billingItemRepo エラー→エラー返却", func(t *testing.T) {
		tagRepo := mappingWith(PrevFilariaTag, model.CodeTypePrescription, []string{"フィラリア薬"})
		petRepo := petRepoWithPets([]model.Pet{dogPet()}, nil)
		billingRepo := &mockBillingItemRepository{
			hasItemByOwnerSinceFn: func(_ context.Context, _, _ uint64, _ time.Time, _ []string) (bool, error) {
				return false, errors.New("db error")
			},
		}
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(nil, nil), nil, petRepo, billingRepo)
		assert.Error(t, svc.SyncFilariaTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})
}

// ---- TestSyncFleaTickTag ----

func TestSyncFleaTickTag(t *testing.T) {
	ctx := context.Background()
	const clinicID, ownerID uint64 = 10, 1

	t.Run("nil tagCodeRepo→noop", func(t *testing.T) {
		svc := buildHealthSvc(nil, ownerRepoReturning(defaultOwnerWithLineID()), nil, nil, nil, billingItemRepoReturning(false, false))
		assert.NoError(t, svc.SyncFleaTickTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("nil billingItemRepo→noop", func(t *testing.T) {
		tagRepo := mappingWith(PrevFleaTickTag, model.CodeTypePrescription, []string{"ノミダニ薬"})
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()), nil, nil, nil, nil)
		assert.NoError(t, svc.SyncFleaTickTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("rxCodesなし→noop", func(t *testing.T) {
		tagRepo := &mockLstepTagCodeMappingRepository{} // FindByClinicIDAndTagName → nil
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			nil, nil, nil, billingItemRepoReturning(false, false))
		assert.NoError(t, svc.SyncFleaTickTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("処方なし→add tag→nil(client=nil)", func(t *testing.T) {
		tagRepo := mappingWith(PrevFleaTickTag, model.CodeTypePrescription, []string{"ノミダニ薬"})
		billingRepo := billingItemRepoReturning(false, false)
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			nil, nil, nil, billingRepo)
		assert.NoError(t, svc.SyncFleaTickTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("処方あり→remove tag→nil(client=nil)", func(t *testing.T) {
		tagRepo := mappingWith(PrevFleaTickTag, model.CodeTypePrescription, []string{"ノミダニ薬"})
		billingRepo := billingItemRepoReturning(true, false)
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			nil, nil, nil, billingRepo)
		assert.NoError(t, svc.SyncFleaTickTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("billingItemRepo エラー→エラー返却", func(t *testing.T) {
		tagRepo := mappingWith(PrevFleaTickTag, model.CodeTypePrescription, []string{"ノミダニ薬"})
		billingRepo := &mockBillingItemRepository{
			hasItemByOwnerSinceFn: func(_ context.Context, _, _ uint64, _ time.Time, _ []string) (bool, error) {
				return false, errors.New("db error")
			},
		}
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			nil, nil, nil, billingRepo)
		assert.Error(t, svc.SyncFleaTickTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})
}

// ---- TestSyncFoodPurchaseTag ----

func TestSyncFoodPurchaseTag(t *testing.T) {
	ctx := context.Background()
	const clinicID, ownerID uint64 = 10, 1

	t.Run("nil tagCodeRepo→noop", func(t *testing.T) {
		svc := buildHealthSvc(nil, ownerRepoReturning(defaultOwnerWithLineID()), nil, nil, nil, billingItemRepoReturning(false, false))
		assert.NoError(t, svc.SyncFoodPurchaseTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("nil billingItemRepo→noop", func(t *testing.T) {
		tagRepo := mappingWith(LtvFoodPurchaseTag, model.CodeTypeMerchandiseItem, []string{"フード1"})
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()), nil, nil, nil, nil)
		assert.NoError(t, svc.SyncFoodPurchaseTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("itemCodesなし→noop対象外(HasFoodPurchase呼ぶ)→nil", func(t *testing.T) {
		// itemCodesが空でもHasFoodPurchaseByOwnerSinceを呼ぶ(category='food'フォールバック)
		tagRepo := &mockLstepTagCodeMappingRepository{} // codes=空
		billingRepo := billingItemRepoReturning(false, false)
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			nil, nil, nil, billingRepo)
		assert.NoError(t, svc.SyncFoodPurchaseTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("フード購入あり→add tag→nil(client=nil)", func(t *testing.T) {
		tagRepo := mappingWith(LtvFoodPurchaseTag, model.CodeTypeMerchandiseItem, []string{"フード1"})
		billingRepo := billingItemRepoReturning(false, true)
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			nil, nil, nil, billingRepo)
		assert.NoError(t, svc.SyncFoodPurchaseTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("フード購入なし→remove tag→nil(client=nil)", func(t *testing.T) {
		tagRepo := mappingWith(LtvFoodPurchaseTag, model.CodeTypeMerchandiseItem, []string{"フード1"})
		billingRepo := billingItemRepoReturning(false, false)
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			nil, nil, nil, billingRepo)
		assert.NoError(t, svc.SyncFoodPurchaseTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})

	t.Run("billingItemRepo エラー→エラー返却", func(t *testing.T) {
		tagRepo := mappingWith(LtvFoodPurchaseTag, model.CodeTypeMerchandiseItem, []string{"フード1"})
		billingRepo := &mockBillingItemRepository{
			hasFoodPurchaseByOwnerSinceFn: func(_ context.Context, _, _ uint64, _ time.Time, _ []string) (bool, error) {
				return false, errors.New("db error")
			},
		}
		svc := buildHealthSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			nil, nil, nil, billingRepo)
		assert.Error(t, svc.SyncFoodPurchaseTagWithMappings(ctx, clinicID, ownerID, nil, nil))
	})
}

// ---- mockVaccinationRepoForHealth ----

type mockVaccinationRepoForHealth struct {
	findByOwnerFn func(ctx context.Context, clinicID, ownerID uint64) ([]model.Vaccination, error)
}

func (m *mockVaccinationRepoForHealth) FindAll(_ context.Context, _ uint64, _, _ *uint64, _, _ *string, _, _ int) ([]model.Vaccination, int64, error) {
	return nil, 0, nil
}
func (m *mockVaccinationRepoForHealth) FindByID(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
	return nil, nil
}
func (m *mockVaccinationRepoForHealth) FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Vaccination, error) {
	if m.findByOwnerFn != nil {
		return m.findByOwnerFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}
func (m *mockVaccinationRepoForHealth) Create(_ context.Context, _ *model.Vaccination) error {
	return nil
}
func (m *mockVaccinationRepoForHealth) Update(_ context.Context, _, _ uint64, _ map[string]any) (*model.Vaccination, error) {
	return nil, nil
}
func (m *mockVaccinationRepoForHealth) Delete(_ context.Context, _, _ uint64) error { return nil }
func (m *mockVaccinationRepoForHealth) FindOwnersByVaccineDeadline(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
	return nil, nil
}

// buildVaccineSvc は syncVaccineDeadlineTagImpl（非公開）を直接呼ぶテスト用に具象型を返す。
func buildVaccineSvc(ownerRepo tagSyncOwnerRepo, vacRepo tagSyncVaccinationRepo) *lstepTagSyncService {
	settingsSvc := &mockLstepSettingsService{}
	var tagCache LstepTagCacheRepository = &mockLstepTagCacheRepository{}
	return NewLstepTagSyncService(
		settingsSvc,
		ownerRepo,
		vacRepo,
		nil, nil,
		tagCache,
		nil, nil, nil, nil, nil, nil,
		nil, // tagConfigRepo
	).(*lstepTagSyncService)
}

// ---- TestSyncVaccineDeadlineTag ----

func TestSyncVaccineDeadlineTag(t *testing.T) {
	ctx := context.Background()
	const clinicID uint64 = 10
	const ownerID uint64 = 1

	t.Run("オプトアウト→noop", func(t *testing.T) {
		owner := &model.Owner{ID: ownerID, ClinicID: clinicID, LstepOptOut: true}
		svc := buildVaccineSvc(ownerRepoReturning(owner), &mockVaccinationRepoForHealth{})
		assert.NoError(t, svc.syncVaccineDeadlineTagImpl(ctx, clinicID, ownerID, model.HealthPreventionThresholds{}.WithDefaults()))
	})

	t.Run("LineUserIDなし→noop", func(t *testing.T) {
		owner := &model.Owner{ID: ownerID, ClinicID: clinicID, LstepOptOut: false, LineUserID: nil}
		svc := buildVaccineSvc(ownerRepoReturning(owner), &mockVaccinationRepoForHealth{})
		assert.NoError(t, svc.syncVaccineDeadlineTagImpl(ctx, clinicID, ownerID, model.HealthPreventionThresholds{}.WithDefaults()))
	})

	t.Run("NextDate nil→期限なし→nil(client=nil)", func(t *testing.T) {
		vacs := []model.Vaccination{{NextDate: nil}}
		vacRepo := &mockVaccinationRepoForHealth{
			findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
				return vacs, nil
			},
		}
		svc := buildVaccineSvc(ownerRepoReturning(defaultOwnerWithLineID()), vacRepo)
		assert.NoError(t, svc.syncVaccineDeadlineTagImpl(ctx, clinicID, ownerID, model.HealthPreventionThresholds{}.WithDefaults()))
	})

	t.Run("期限なし→nil(client=nil)", func(t *testing.T) {
		farFuture := time.Now().AddDate(0, 0, model.DefaultVaccineDeadlineDays+10)
		vacs := []model.Vaccination{{NextDate: &farFuture}}
		vacRepo := &mockVaccinationRepoForHealth{
			findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
				return vacs, nil
			},
		}
		svc := buildVaccineSvc(ownerRepoReturning(defaultOwnerWithLineID()), vacRepo)
		assert.NoError(t, svc.syncVaccineDeadlineTagImpl(ctx, clinicID, ownerID, model.HealthPreventionThresholds{}.WithDefaults()))
	})

	t.Run("期限あり→nil(client=nil)", func(t *testing.T) {
		withinDeadline := time.Now().AddDate(0, 0, model.DefaultVaccineDeadlineDays-5)
		vacs := []model.Vaccination{{NextDate: &withinDeadline}}
		vacRepo := &mockVaccinationRepoForHealth{
			findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
				return vacs, nil
			},
		}
		svc := buildVaccineSvc(ownerRepoReturning(defaultOwnerWithLineID()), vacRepo)
		assert.NoError(t, svc.syncVaccineDeadlineTagImpl(ctx, clinicID, ownerID, model.HealthPreventionThresholds{}.WithDefaults()))
	})

	t.Run("vacRepo エラー→エラー返却", func(t *testing.T) {
		vacRepo := &mockVaccinationRepoForHealth{
			findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
				return nil, errors.New("db error")
			},
		}
		svc := buildVaccineSvc(ownerRepoReturning(defaultOwnerWithLineID()), vacRepo)
		assert.Error(t, svc.syncVaccineDeadlineTagImpl(ctx, clinicID, ownerID, model.HealthPreventionThresholds{}.WithDefaults()))
	})
}
