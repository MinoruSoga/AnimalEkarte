package lstep

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

// Note: SyncVaccineDeadlineTag's opt-out / no-line-id / no-deadline / vacRepo-error branches are
// already covered by TestSyncVaccineDeadlineTag in lstep_health_tag_sync_test.go (which always
// exercises a nil lstep client). This file focuses on the branches only reachable with a real
// (mocked) lstep client: AddTag/RemoveTag success & failure.
//
// B-3: 旧公開ラッパー SyncVaccineDeadlineTag（thresholds fetch → syncVaccineDeadlineTagImpl 委譲）は
// 呼び出し元ゼロのため削除済み（batch は syncVaccineDeadlineTagImpl を thresholds 事前取得で直接呼ぶ）。
// ラッパー固有だった「shouldSkipSync 失敗」「thresholds 取得失敗」の分岐は syncVaccineDeadlineTagImpl の
// シグネチャ（thresholds を値で受け取る）には存在しないため、対応するテストは削除した。

func TestSyncVaccineDeadlineTagImpl_ClientInteraction(t *testing.T) {
	lineUID := "U_vaccine_deadline_test"
	near := time.Now().AddDate(0, 0, model.DefaultVaccineDeadlineDays-5)

	t.Run("returns wrapped error when checkOptOut fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) { return nil, errors.New("db error") },
			},
		}
		err := svc.syncVaccineDeadlineTagImpl(context.Background(), 1, 10, model.HealthPreventionThresholds{}.WithDefaults())
		assert.Error(t, err)
	})

	t.Run("returns wrapped error when vacRepo.FindByOwner fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepoForHealth{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
					return nil, errors.New("db error")
				},
			},
		}
		err := svc.syncVaccineDeadlineTagImpl(context.Background(), 1, 10, model.HealthPreventionThresholds{}.WithDefaults())
		assert.Error(t, err)
	})

	t.Run("deadline soon: adds the tag and upserts cache with the reason", func(t *testing.T) {
		var addedTag, upsertedTag, upsertedReason string
		client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, tagName string) error { addedTag = tagName; return nil }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepoForHealth{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
					return []model.Vaccination{{NextDate: &near}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				upsertTagFn: func(_ context.Context, _, _ uint64, tagName, _, reason string) error {
					upsertedTag = tagName
					upsertedReason = reason
					return nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}

		err := svc.syncVaccineDeadlineTagImpl(context.Background(), 1, 10, model.HealthPreventionThresholds{}.WithDefaults())
		assert.NoError(t, err)
		assert.Equal(t, PrevVaccineDeadlineTag, addedTag)
		assert.Equal(t, PrevVaccineDeadlineTag, upsertedTag)
		assert.NotEmpty(t, upsertedReason)
	})

	t.Run("deadline soon: AddTag failure is wrapped and returned", func(t *testing.T) {
		client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepoForHealth{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
					return []model.Vaccination{{NextDate: &near}}, nil
				},
			},
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.syncVaccineDeadlineTagImpl(context.Background(), 1, 10, model.HealthPreventionThresholds{}.WithDefaults())
		assert.Error(t, err)
	})

	t.Run("no deadline soon: removes the tag and deletes cache entry", func(t *testing.T) {
		var removedTag, deletedTag string
		client := &mockLstepAPIClient{removeTagFn: func(_ context.Context, _, tagName string) error { removedTag = tagName; return nil }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepoForHealth{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
					return nil, nil // no vaccination records → deadlineSoon=false
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				deleteTagFn: func(_ context.Context, _, _ uint64, tagName string) error {
					deletedTag = tagName
					return nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.syncVaccineDeadlineTagImpl(context.Background(), 1, 10, model.HealthPreventionThresholds{}.WithDefaults())
		assert.NoError(t, err)
		assert.Equal(t, PrevVaccineDeadlineTag, removedTag)
		assert.Equal(t, PrevVaccineDeadlineTag, deletedTag)
	})

	t.Run("no deadline soon: RemoveTag failure propagates (G2B-01)", func(t *testing.T) {
		client := &mockLstepAPIClient{removeTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepoForHealth{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) { return nil, nil },
			},
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.syncVaccineDeadlineTagImpl(context.Background(), 1, 10, model.HealthPreventionThresholds{}.WithDefaults())
		assert.Error(t, err)
	})

	t.Run("preloaded vaccinations skip vacRepo.FindByOwner (G2F-02 bulk)", func(t *testing.T) {
		var findCalls int
		client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, _ string) error { return nil }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepoForHealth{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
					findCalls++
					return nil, errors.New("should not fetch")
				},
			},
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		preloaded := []model.Vaccination{{NextDate: &near}}
		err := svc.syncVaccineDeadlineTagWithInputs(context.Background(), 1, 10, model.HealthPreventionThresholds{}.WithDefaults(), &preloaded)
		assert.NoError(t, err)
		assert.Equal(t, 0, findCalls)
	})
}

func TestSyncVaccineDeadlineTag_VaccineSupersession(t *testing.T) {
	ctx := context.Background()
	const clinicID uint64 = 1
	const otherClinicID uint64 = 2
	const ownerID uint64 = 10
	lineUID := "U_vaccine_supersession"
	now := time.Now()
	oldDate := now.AddDate(0, 0, -30)
	newDate := now.AddDate(0, 0, -10)
	inWindow := now.AddDate(0, 0, model.DefaultVaccineDeadlineDays-5)
	inWindowEarlier := now.AddDate(0, 0, 5)
	beyondWindow := now.AddDate(0, 0, model.DefaultVaccineDeadlineDays+10)
	petID := uint64(100)
	otherPetID := uint64(200)
	vaccineID := uint64(7)
	otherVaccineID := uint64(8)

	regularVaccine := &model.Vaccine{ClinicID: clinicID, Name: "混合ワクチン"}
	genericVaccine := &model.Vaccine{ClinicID: clinicID, Name: "他院接種"}
	genericVaccineOtherClinic := &model.Vaccine{ClinicID: otherClinicID, Name: "他院接種"}

	vac := func(pet *uint64, vacID uint64, clinic uint64, date time.Time, next *time.Time, vaccine *model.Vaccine, deleted bool) model.Vaccination {
		v := model.Vaccination{ClinicID: clinic, PetID: pet, VaccineID: vacID, Date: date, NextDate: next, Vaccine: vaccine}
		if deleted {
			v.DeletedAt = gorm.DeletedAt{Time: now, Valid: true}
		}
		return v
	}

	cases := []struct {
		name       string
		records    []model.Vaccination
		wantAdd    bool
		wantReason *time.Time
	}{
		{
			name: "newer same pet+vaccine with nil NextDate supersedes old due record",
			records: []model.Vaccination{
				vac(&petID, vaccineID, clinicID, oldDate, &inWindow, regularVaccine, false),
				vac(&petID, vaccineID, clinicID, newDate, nil, regularVaccine, false),
			},
			wantAdd: false,
		},
		{
			name: "newer same pet+vaccine with NextDate beyond window supersedes old due record",
			records: []model.Vaccination{
				vac(&petID, vaccineID, clinicID, oldDate, &inWindow, regularVaccine, false),
				vac(&petID, vaccineID, clinicID, newDate, &beyondWindow, regularVaccine, false),
			},
			wantAdd: false,
		},
		{
			name: "different vaccine does not supersede",
			records: []model.Vaccination{
				vac(&petID, vaccineID, clinicID, oldDate, &inWindow, regularVaccine, false),
				vac(&petID, otherVaccineID, clinicID, newDate, nil, regularVaccine, false),
			},
			wantAdd:    true,
			wantReason: &inWindow,
		},
		{
			name: "different pet does not supersede",
			records: []model.Vaccination{
				vac(&petID, vaccineID, clinicID, oldDate, &inWindow, regularVaccine, false),
				vac(&otherPetID, vaccineID, clinicID, newDate, nil, regularVaccine, false),
			},
			wantAdd:    true,
			wantReason: &inWindow,
		},
		{
			name: "generic 他院 vaccine keeps each record current",
			records: []model.Vaccination{
				vac(&petID, vaccineID, clinicID, oldDate, &inWindow, genericVaccine, false),
				vac(&petID, vaccineID, clinicID, newDate, nil, genericVaccine, false),
			},
			wantAdd:    true,
			wantReason: &inWindow,
		},
		{
			name: "generic vaccine master from another clinic does not mark records generic",
			records: []model.Vaccination{
				vac(&petID, vaccineID, clinicID, oldDate, &inWindow, genericVaccineOtherClinic, false),
				vac(&petID, vaccineID, clinicID, newDate, nil, genericVaccineOtherClinic, false),
			},
			wantAdd: false,
		},
		{
			name: "same-date records are both retained",
			records: []model.Vaccination{
				vac(&petID, vaccineID, clinicID, oldDate, nil, regularVaccine, false),
				vac(&petID, vaccineID, clinicID, oldDate, &inWindow, regularVaccine, false),
			},
			wantAdd:    true,
			wantReason: &inWindow,
		},
		{
			name: "deleted due-only record never produces a tag",
			records: []model.Vaccination{
				vac(&petID, vaccineID, clinicID, oldDate, &inWindow, regularVaccine, true),
			},
			wantAdd: false,
		},
		{
			name: "deleted later record does not supersede",
			records: []model.Vaccination{
				vac(&petID, vaccineID, clinicID, oldDate, &inWindow, regularVaccine, false),
				vac(&petID, vaccineID, clinicID, newDate, nil, regularVaccine, true),
			},
			wantAdd:    true,
			wantReason: &inWindow,
		},
		{
			name: "foreign clinic newer record does not suppress",
			records: []model.Vaccination{
				vac(&petID, vaccineID, clinicID, oldDate, &inWindow, regularVaccine, false),
				vac(&petID, vaccineID, otherClinicID, newDate, nil, regularVaccine, false),
			},
			wantAdd:    true,
			wantReason: &inWindow,
		},
		{
			name: "zero vaccine id records are exempt from supersession",
			records: []model.Vaccination{
				vac(&petID, 0, clinicID, oldDate, &inWindow, nil, false),
				vac(&petID, 0, clinicID, newDate, nil, nil, false),
			},
			wantAdd:    true,
			wantReason: &inWindow,
		},
		{
			name: "nil pet id records are exempt from supersession",
			records: []model.Vaccination{
				vac(nil, vaccineID, clinicID, oldDate, &inWindow, regularVaccine, false),
				vac(nil, vaccineID, clinicID, newDate, nil, regularVaccine, false),
			},
			wantAdd:    true,
			wantReason: &inWindow,
		},
		{
			name: "reason reports the earliest eligible NextDate among current records only",
			records: []model.Vaccination{
				vac(&petID, vaccineID, clinicID, oldDate, &inWindowEarlier, regularVaccine, false),
				vac(&petID, vaccineID, clinicID, newDate, &inWindow, regularVaccine, false),
			},
			wantAdd:    true,
			wantReason: &inWindow,
		},
		{
			name: "nil Vaccine association does not mark records generic",
			records: []model.Vaccination{
				vac(&petID, vaccineID, clinicID, oldDate, &inWindow, nil, false),
				vac(&petID, vaccineID, clinicID, newDate, nil, nil, false),
			},
			wantAdd: false,
		},
	}

	for _, tc := range cases {
		for _, viaPreloaded := range []bool{false, true} {
			name := tc.name
			if viaPreloaded {
				name += " (preloaded)"
			}
			t.Run(name, func(t *testing.T) {
				var addedTag, removedTag, upsertedReason string
				client := &mockLstepAPIClient{
					addTagFn:    func(_ context.Context, _, tagName string) error { addedTag = tagName; return nil },
					removeTagFn: func(_ context.Context, _, tagName string) error { removedTag = tagName; return nil },
				}
				svc := &lstepTagSyncService{
					settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
					ownerRepo: &mockOwnerRepository{
						findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
							return &model.Owner{ID: id, LineUserID: &lineUID}, nil
						},
					},
					vacRepo: &mockVaccinationRepoForHealth{
						findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
							if viaPreloaded {
								return nil, errors.New("should not fetch when preloaded")
							}
							return tc.records, nil
						},
					},
					tagCacheRepo: &mockLstepTagCacheRepository{
						upsertTagFn: func(_ context.Context, _, _ uint64, _, _, reason string) error {
							upsertedReason = reason
							return nil
						},
					},
					buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
				}

				inputSnapshot := append([]model.Vaccination(nil), tc.records...)

				var err error
				if viaPreloaded {
					preloaded := tc.records
					err = svc.syncVaccineDeadlineTagWithInputs(ctx, clinicID, ownerID, model.HealthPreventionThresholds{}.WithDefaults(), &preloaded)
				} else {
					err = svc.syncVaccineDeadlineTagImpl(ctx, clinicID, ownerID, model.HealthPreventionThresholds{}.WithDefaults())
				}
				assert.NoError(t, err)
				assert.Equal(t, inputSnapshot, tc.records, "input slice must not be mutated")
				if tc.wantAdd {
					assert.Equal(t, PrevVaccineDeadlineTag, addedTag, "expected AddTag")
					assert.Empty(t, removedTag)
					if tc.wantReason != nil {
						assert.Equal(t, fmt.Sprintf("次回期限: %s", tc.wantReason.Format(time.DateOnly)), upsertedReason)
					}
				} else {
					assert.Equal(t, PrevVaccineDeadlineTag, removedTag, "expected RemoveTag")
					assert.Empty(t, addedTag)
				}
			})
		}
	}
}
