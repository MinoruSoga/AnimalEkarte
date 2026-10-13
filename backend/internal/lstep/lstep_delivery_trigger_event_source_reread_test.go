package lstep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-330: イベントソース再読の回帰テスト。
// 非 tentative トリガーは candidate 抽出後に
//   1. scheduled claim 前
//   2. delivery owner lock 内の write/send 直前
// の 2 回、イベントソース由来の適格性を owner 単位で再読しなければならない。
// 初診/休眠系は生存ペット由来カルテの boundary 再計算、ワクチンは alive
// deadline イベント、誕生日は alive birthday イベント、タグ系は scoped
// living-owner チェック、ウェルカムはイベントペット再読を行う。
// 抽出時点の candidate がその後に死亡しても外部 AddTag へ到達してはならず、
// 再読エラーや capability 欠落は fail-closed（送信しない）でなければならない。

// reReadSequence は eligibility fn を N 回目ごとの戻り値で制御するヘルパー。
// results[0] が claim 前再読、results[1] が write 直前再読に対応する。
// 呼出回数は *calls に記録する。
func reReadSequence(results []bool, calls *int) func() (bool, error) {
	return func() (bool, error) {
		i := *calls
		*calls = i + 1
		if i < len(results) {
			return results[i], nil
		}
		return true, nil
	}
}

type ownerScopedReReadCase struct {
	name          string
	okAtClaim     bool
	okAtWrite     bool
	wantFired     int
	wantAddTag    int
	wantMinChecks int
}

var ownerScopedReReadCases = []ownerScopedReReadCase{
	{"抽出後に死亡: claim 前の再読で不適格となり AddTag しない", false, false, 0, 0, 1},
	{"claim 後 write 前に死亡: lock 内の再読で不適格となり AddTag しない", true, false, 0, 0, 2},
	{"イベントソースが有効なままなら AddTag する", true, true, 1, 1, 2},
}

func rereadOwnerRepo() *mockOwnerRepoForDelivery {
	return &mockOwnerRepoForDelivery{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
			return defaultOwnerWithLine(id), nil
		},
	}
}

func TestLstepDeliveryTriggerService_FirstVisitFollowUp_EventSourceReRead(t *testing.T) {
	asOf := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	const clinicID, ownerID = uint64(1), uint64(10)

	for _, tt := range ownerScopedReReadCases {
		t.Run(tt.name, func(t *testing.T) {
			checks := 0
			seq := reReadSequence([]bool{tt.okAtClaim, tt.okAtWrite}, &checks)
			medRepo := &mockMedicalRecordRepository{
				findOwnersByFirstVisitFn: func(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
					return []uint64{ownerID}, nil
				},
				ownerHasAliveFirstVisitOnFn: func(_ context.Context, _, _ uint64, _ time.Time) (bool, error) {
					return seq()
				},
			}
			addTagCalls := 0
			// CountLivingByOwner=1 を返しても owner 単位の「他に生存ペットがいる」
			// だけでは契約を満たさない（イベントソース再読が必須）。
			petRepo := &mockPetRepoForDelivery{
				countLivingByOwnerFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return 1, nil
				},
			}
			svc := buildDeliverySvc(rereadOwnerRepo(), medRepo, nil, petRepo, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
			injectTestClient(svc, &mockLstepClientForDelivery{
				addTagFn: func(_ context.Context, _, _ string) error {
					addTagCalls++
					return nil
				},
			})

			count, errs := svc.TriggerFirstVisitFollowUp3D(context.Background(), clinicID, asOf)

			require.Empty(t, errs)
			assert.Equal(t, tt.wantFired, count)
			assert.Equal(t, tt.wantAddTag, addTagCalls, "死亡したイベントソース由来の candidate に外部 AddTag してはいけない")
			assert.GreaterOrEqual(t, checks, tt.wantMinChecks, "claim 前と write 直前にイベントソースを再読すること")
		})
	}
}

func TestLstepDeliveryTriggerService_DormantPrevention_EventSourceReRead(t *testing.T) {
	asOf := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	const clinicID, ownerID = uint64(1), uint64(60)

	for _, tt := range ownerScopedReReadCases {
		t.Run(tt.name, func(t *testing.T) {
			checks := 0
			seq := reReadSequence([]bool{tt.okAtClaim, tt.okAtWrite}, &checks)
			medRepo := &mockMedicalRecordRepository{
				findOwnersByLastVisitDaysFn: func(_ context.Context, _ uint64, _ int, _ time.Time) ([]uint64, error) {
					return []uint64{ownerID}, nil
				},
				ownerHasAliveLastVisitOnFn: func(_ context.Context, _, _ uint64, _ time.Time) (bool, error) {
					return seq()
				},
			}
			addTagCalls := 0
			svc := buildDeliverySvc(rereadOwnerRepo(), medRepo, nil, &mockPetRepoForDelivery{}, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
			injectTestClient(svc, &mockLstepClientForDelivery{
				addTagFn: func(_ context.Context, _, _ string) error {
					addTagCalls++
					return nil
				},
			})

			count, errs := svc.TriggerDormantPrevention180(context.Background(), clinicID, asOf)

			require.Empty(t, errs)
			assert.Equal(t, tt.wantFired, count)
			assert.Equal(t, tt.wantAddTag, addTagCalls)
			assert.GreaterOrEqual(t, checks, tt.wantMinChecks)
		})
	}
}

func TestLstepDeliveryTriggerService_VaccineDeadline_EventSourceReRead(t *testing.T) {
	asOf := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	const clinicID, ownerID = uint64(1), uint64(40)

	for _, tt := range ownerScopedReReadCases {
		t.Run(tt.name, func(t *testing.T) {
			checks := 0
			seq := reReadSequence([]bool{tt.okAtClaim, tt.okAtWrite}, &checks)
			vacRepo := &mockVaccinationRepoForDelivery{
				findOwnersByVaccineDeadlineFn: func(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
					return []uint64{ownerID}, nil
				},
				ownerHasAliveVaccineDeadlineFn: func(_ context.Context, _, _ uint64, _ time.Time) (bool, error) {
					return seq()
				},
			}
			addTagCalls := 0
			svc := buildDeliverySvc(rereadOwnerRepo(), nil, vacRepo, &mockPetRepoForDelivery{}, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
			injectTestClient(svc, &mockLstepClientForDelivery{
				addTagFn: func(_ context.Context, _, _ string) error {
					addTagCalls++
					return nil
				},
			})

			count, errs := svc.TriggerVaccineDeadline30(context.Background(), clinicID, asOf)

			require.Empty(t, errs)
			assert.Equal(t, tt.wantFired, count)
			assert.Equal(t, tt.wantAddTag, addTagCalls)
			assert.GreaterOrEqual(t, checks, tt.wantMinChecks)
		})
	}
}

func TestLstepDeliveryTriggerService_BirthdayMessage_EventSourceReRead(t *testing.T) {
	asOf := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	const clinicID, ownerID = uint64(1), uint64(50)

	for _, tt := range ownerScopedReReadCases {
		t.Run(tt.name, func(t *testing.T) {
			checks := 0
			seq := reReadSequence([]bool{tt.okAtClaim, tt.okAtWrite}, &checks)
			petRepo := &mockPetRepoForDelivery{
				findOwnersByPetBirthdayFn: func(_ context.Context, _ uint64, _, _ int) ([]uint64, error) {
					return []uint64{ownerID}, nil
				},
				ownerHasAlivePetBirthdayFn: func(_ context.Context, _, _ uint64, _, _ int) (bool, error) {
					return seq()
				},
			}
			addTagCalls := 0
			svc := buildDeliverySvc(rereadOwnerRepo(), nil, nil, petRepo, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
			injectTestClient(svc, &mockLstepClientForDelivery{
				addTagFn: func(_ context.Context, _, _ string) error {
					addTagCalls++
					return nil
				},
			})

			count, errs := svc.TriggerBirthdayMessage(context.Background(), clinicID, asOf)

			require.Empty(t, errs)
			assert.Equal(t, tt.wantFired, count)
			assert.Equal(t, tt.wantAddTag, addTagCalls)
			assert.GreaterOrEqual(t, checks, tt.wantMinChecks)
		})
	}
}

// TestLstepDeliveryTriggerService_LivingOwnerCheck はタグ系・健診系の
// scoped living-owner 再検証を検証する（EMR-330）。
func TestLstepDeliveryTriggerService_LivingOwnerCheck(t *testing.T) {
	asOf := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	const clinicID, ownerID = uint64(1), uint64(70)

	t.Run("タグ由来 candidate の全ペット死亡は AddTag しない", func(t *testing.T) {
		tagRepo := &mockTagCacheRepoForDelivery{
			findOwnerIDsByTagFn: func(_ context.Context, _ uint64, _ string) ([]uint64, error) {
				return []uint64{ownerID}, nil
			},
		}
		petRepo := &mockPetRepoForDelivery{
			countLivingByOwnerFn: func(_ context.Context, _, _ uint64) (int64, error) {
				return 0, nil
			},
		}
		addTagCalls := 0
		svc := buildDeliverySvc(rereadOwnerRepo(), nil, nil, petRepo, tagRepo, &mockDeliveryTriggerLogRepository{}, enabledSettings())
		injectTestClient(svc, &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, _ string) error {
				addTagCalls++
				return nil
			},
		})

		count, errs := svc.TriggerFilariaAlert(context.Background(), clinicID, asOf)

		require.Empty(t, errs)
		assert.Equal(t, 0, count)
		assert.Equal(t, 0, addTagCalls)
	})

	t.Run("タグ由来 candidate の生存ペットありは AddTag する", func(t *testing.T) {
		tagRepo := &mockTagCacheRepoForDelivery{
			findOwnerIDsByTagFn: func(_ context.Context, _ uint64, _ string) ([]uint64, error) {
				return []uint64{ownerID}, nil
			},
		}
		addTagCalls := 0
		svc := buildDeliverySvc(rereadOwnerRepo(), nil, nil, &mockPetRepoForDelivery{}, tagRepo, &mockDeliveryTriggerLogRepository{}, enabledSettings())
		injectTestClient(svc, &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, _ string) error {
				addTagCalls++
				return nil
			},
		})

		count, errs := svc.TriggerFilariaAlert(context.Background(), clinicID, asOf)

		require.Empty(t, errs)
		assert.Equal(t, 1, count)
		assert.Equal(t, 1, addTagCalls)
	})

	t.Run("petRepo 未設定は fail-closed で AddTag しない", func(t *testing.T) {
		tagRepo := &mockTagCacheRepoForDelivery{
			findOwnerIDsByTagFn: func(_ context.Context, _ uint64, _ string) ([]uint64, error) {
				return []uint64{ownerID}, nil
			},
		}
		addTagCalls := 0
		svc := buildDeliverySvc(rereadOwnerRepo(), nil, nil, nil, tagRepo, &mockDeliveryTriggerLogRepository{}, enabledSettings())
		injectTestClient(svc, &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, _ string) error {
				addTagCalls++
				return nil
			},
		})

		count, errs := svc.TriggerFilariaAlert(context.Background(), clinicID, asOf)

		assert.Equal(t, 0, count)
		assert.NotEmpty(t, errs, "living-owner 再読の capability 欠落は error として扱う")
		assert.Equal(t, 0, addTagCalls)
	})

	t.Run("健診フォローアップの petRepo 未設定は fail-closed", func(t *testing.T) {
		addTagCalls := 0
		svc := buildDeliverySvc(rereadOwnerRepo(), nil, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
		injectTestClient(svc, &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, _ string) error {
				addTagCalls++
				return nil
			},
		})

		err := svc.TriggerCheckupFollowUp(context.Background(), clinicID, ownerID, 301)

		assert.Error(t, err)
		assert.Equal(t, 0, addTagCalls)
	})
}

// TestLstepDeliveryTriggerService_CheckupFollowUp_EventPetEligibility は
// 健診フォローアップがイベントペットを typed 引数で受け取り、claim/write 前に
// その生存・現飼主・clinic 整合を再読することを検証する（EMR-330 R1）。
// 死亡した健診ペットは、同一飼主に生存の兄弟ペットがいても配信しない。
func TestLstepDeliveryTriggerService_CheckupFollowUp_EventPetEligibility(t *testing.T) {
	const clinicID, ownerID, petID = uint64(1), uint64(10), uint64(301)

	livingEventPet := func() *model.Pet {
		return &model.Pet{ID: petID, ClinicID: clinicID, OwnerID: ownerID, Status: model.PetStatusAlive}
	}
	buildSvc := func(petRepo *mockPetRepoForDelivery) (LstepDeliveryTriggerService, *int) {
		addTagCalls := 0
		svc := buildDeliverySvc(rereadOwnerRepo(), nil, nil, petRepo, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
		injectTestClient(svc, &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, _ string) error {
				addTagCalls++
				return nil
			},
		})
		return svc, &addTagCalls
	}

	t.Run("健診ペット生存なら送信する（positive control）", func(t *testing.T) {
		svc, addTagCalls := buildSvc(&mockPetRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID, OwnerID: ownerID, Status: model.PetStatusAlive}, nil
			},
		})

		err := svc.TriggerCheckupFollowUp(context.Background(), clinicID, ownerID, petID)

		assert.NoError(t, err)
		assert.Equal(t, 1, *addTagCalls)
	})

	t.Run("死亡した健診ペットは生存の兄弟ペットがいても送信しない", func(t *testing.T) {
		svc, addTagCalls := buildSvc(&mockPetRepoForDelivery{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				pet := livingEventPet()
				pet.Status = model.PetStatusDeceased
				return pet, nil
			},
			countLivingByOwnerFn: func(_ context.Context, _, _ uint64) (int64, error) {
				return 1, nil // 生存の兄弟ペットがいても救済しない
			},
		})

		err := svc.TriggerCheckupFollowUp(context.Background(), clinicID, ownerID, petID)

		assert.NoError(t, err)
		assert.Equal(t, 0, *addTagCalls)
	})

	t.Run("claim 前に死亡した健診ペットは送信しない", func(t *testing.T) {
		svc, addTagCalls := buildSvc(&mockPetRepoForDelivery{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				// 最初の再読（claim 前）で既に死亡
				pet := livingEventPet()
				pet.Status = model.PetStatusDeceased
				return pet, nil
			},
		})

		err := svc.TriggerCheckupFollowUp(context.Background(), clinicID, ownerID, petID)

		assert.NoError(t, err)
		assert.Equal(t, 0, *addTagCalls)
	})

	t.Run("claim 後・送信直前に死亡した健診ペットは送信しない", func(t *testing.T) {
		calls := 0
		svc, addTagCalls := buildSvc(&mockPetRepoForDelivery{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				calls++
				if calls == 1 {
					return livingEventPet(), nil
				}
				pet := livingEventPet()
				pet.Status = model.PetStatusDeceased
				return pet, nil
			},
		})

		err := svc.TriggerCheckupFollowUp(context.Background(), clinicID, ownerID, petID)

		assert.NoError(t, err)
		assert.Equal(t, 0, *addTagCalls)
		assert.GreaterOrEqual(t, calls, 2, "claim 前と送信直前の 2 回再読が必要")
	})

	t.Run("イベントペットの現飼主が一致しないなら送信しない", func(t *testing.T) {
		svc, addTagCalls := buildSvc(&mockPetRepoForDelivery{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return &model.Pet{ID: petID, ClinicID: clinicID, OwnerID: 999, Status: model.PetStatusAlive}, nil
			},
		})

		err := svc.TriggerCheckupFollowUp(context.Background(), clinicID, ownerID, petID)

		assert.NoError(t, err)
		assert.Equal(t, 0, *addTagCalls)
	})

	t.Run("イベントペットが存在しないなら送信しない", func(t *testing.T) {
		svc, addTagCalls := buildSvc(&mockPetRepoForDelivery{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return nil, nil
			},
		})

		err := svc.TriggerCheckupFollowUp(context.Background(), clinicID, ownerID, petID)

		assert.NoError(t, err)
		assert.Equal(t, 0, *addTagCalls)
	})
}

// TestLstepDeliveryTriggerService_FirstVisitWelcome_EventPetEligibility は
// ウェルカム配信がイベントペットを typed 引数で受け取り、claim/write 前に
// その生存・所有・clinic 整合を再読することを検証する（EMR-330）。
func TestLstepDeliveryTriggerService_FirstVisitWelcome_EventPetEligibility(t *testing.T) {
	const clinicID, ownerID, petID = uint64(1), uint64(10), uint64(90)

	t.Run("イベントペットが生存・所有一致なら送信する", func(t *testing.T) {
		petRepo := &mockPetRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID, OwnerID: ownerID, Status: model.PetStatusAlive}, nil
			},
		}
		addTagCalls := 0
		svc := buildDeliverySvc(rereadOwnerRepo(), nil, nil, petRepo, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
		injectTestClient(svc, &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, _ string) error {
				addTagCalls++
				return nil
			},
		})

		err := svc.TriggerFirstVisitWelcome(context.Background(), clinicID, ownerID, petID)

		assert.NoError(t, err)
		assert.Equal(t, 1, addTagCalls)
	})

	t.Run("イベントペットが死亡していれば AddTag しない", func(t *testing.T) {
		deceasedAt := time.Now()
		petRepo := &mockPetRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID, OwnerID: ownerID, Status: model.PetStatusDeceased, DeceasedAt: &deceasedAt}, nil
			},
		}
		addTagCalls := 0
		svc := buildDeliverySvc(rereadOwnerRepo(), nil, nil, petRepo, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
		injectTestClient(svc, &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, _ string) error {
				addTagCalls++
				return nil
			},
		})

		err := svc.TriggerFirstVisitWelcome(context.Background(), clinicID, ownerID, petID)

		assert.NoError(t, err)
		assert.Equal(t, 0, addTagCalls)
	})

	t.Run("イベントペットが別飼い主なら AddTag しない", func(t *testing.T) {
		petRepo := &mockPetRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID, OwnerID: 999, Status: model.PetStatusAlive}, nil
			},
		}
		addTagCalls := 0
		svc := buildDeliverySvc(rereadOwnerRepo(), nil, nil, petRepo, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
		injectTestClient(svc, &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, _ string) error {
				addTagCalls++
				return nil
			},
		})

		err := svc.TriggerFirstVisitWelcome(context.Background(), clinicID, ownerID, petID)

		assert.NoError(t, err)
		assert.Equal(t, 0, addTagCalls)
	})

	t.Run("イベントペット再読エラーは fail-closed", func(t *testing.T) {
		petRepo := &mockPetRepoForDelivery{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return nil, errors.New("pet read error")
			},
		}
		addTagCalls := 0
		svc := buildDeliverySvc(rereadOwnerRepo(), nil, nil, petRepo, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
		injectTestClient(svc, &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, _ string) error {
				addTagCalls++
				return nil
			},
		})

		err := svc.TriggerFirstVisitWelcome(context.Background(), clinicID, ownerID, petID)

		assert.Error(t, err)
		assert.Equal(t, 0, addTagCalls)
	})

	t.Run("petRepo 未設定は fail-closed", func(t *testing.T) {
		addTagCalls := 0
		svc := buildDeliverySvc(rereadOwnerRepo(), nil, nil, nil, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
		injectTestClient(svc, &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, _ string) error {
				addTagCalls++
				return nil
			},
		})

		err := svc.TriggerFirstVisitWelcome(context.Background(), clinicID, ownerID, petID)

		assert.Error(t, err)
		assert.Equal(t, 0, addTagCalls)
	})

	t.Run("status=deceased のみ（deceased_at NULL）の半死亡ペットは送信しない", func(t *testing.T) {
		petRepo := &mockPetRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID, OwnerID: ownerID, Status: model.PetStatusDeceased}, nil
			},
		}
		addTagCalls := 0
		svc := buildDeliverySvc(rereadOwnerRepo(), nil, nil, petRepo, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
		injectTestClient(svc, &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, _ string) error {
				addTagCalls++
				return nil
			},
		})

		err := svc.TriggerFirstVisitWelcome(context.Background(), clinicID, ownerID, petID)

		assert.NoError(t, err)
		assert.Equal(t, 0, addTagCalls)
	})

	t.Run("deceased_at のみ非 NULL（status=alive）の半死亡ペットは送信しない", func(t *testing.T) {
		deceasedAt := time.Now()
		petRepo := &mockPetRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID, OwnerID: ownerID, Status: model.PetStatusAlive, DeceasedAt: &deceasedAt}, nil
			},
		}
		addTagCalls := 0
		svc := buildDeliverySvc(rereadOwnerRepo(), nil, nil, petRepo, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
		injectTestClient(svc, &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, _ string) error {
				addTagCalls++
				return nil
			},
		})

		err := svc.TriggerFirstVisitWelcome(context.Background(), clinicID, ownerID, petID)

		assert.NoError(t, err)
		assert.Equal(t, 0, addTagCalls)
	})

	t.Run("claim 前は生存・送信直前再読で死亡したイベントペットは送信しない", func(t *testing.T) {
		calls := 0
		petRepo := &mockPetRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
				calls++
				pet := &model.Pet{ID: id, ClinicID: clinicID, OwnerID: ownerID, Status: model.PetStatusAlive}
				if calls > 1 {
					pet.Status = model.PetStatusDeceased
				}
				return pet, nil
			},
		}
		addTagCalls := 0
		svc := buildDeliverySvc(rereadOwnerRepo(), nil, nil, petRepo, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
		injectTestClient(svc, &mockLstepClientForDelivery{
			addTagFn: func(_ context.Context, _, _ string) error {
				addTagCalls++
				return nil
			},
		})

		err := svc.TriggerFirstVisitWelcome(context.Background(), clinicID, ownerID, petID)

		assert.NoError(t, err)
		assert.Equal(t, 0, addTagCalls)
		assert.GreaterOrEqual(t, calls, 2, "claim 前と送信直前の 2 回再読が必要")
	})
}

// TestLstepDeliveryTriggerService_EventSourceReRead_Error はイベントソース
// 再読のエラーが fail-closed（送信せず error として返る）ことを検証する。
func TestLstepDeliveryTriggerService_EventSourceReRead_Error(t *testing.T) {
	asOf := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	const clinicID, ownerID = uint64(1), uint64(10)

	medRepo := &mockMedicalRecordRepository{
		findOwnersByFirstVisitFn: func(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
			return []uint64{ownerID}, nil
		},
		ownerHasAliveFirstVisitOnFn: func(_ context.Context, _, _ uint64, _ time.Time) (bool, error) {
			return false, errors.New("event source re-read error")
		},
	}
	addTagCalls := 0
	svc := buildDeliverySvc(rereadOwnerRepo(), medRepo, nil, &mockPetRepoForDelivery{}, &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
	injectTestClient(svc, &mockLstepClientForDelivery{
		addTagFn: func(_ context.Context, _, _ string) error {
			addTagCalls++
			return nil
		},
	})

	count, errs := svc.TriggerFirstVisitFollowUp3D(context.Background(), clinicID, asOf)

	assert.Equal(t, 0, count)
	assert.NotEmpty(t, errs, "再読エラーは送信せず error として扱う")
	assert.Equal(t, 0, addTagCalls)
}
