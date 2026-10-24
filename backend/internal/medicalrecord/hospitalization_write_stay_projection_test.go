package medicalrecord

// hospitalization_write_stay_projection_test.go — F-01: 入院の書き込み（Create / Update / SwapCages）は
// 滞在 projection を書き込みと同じトランザクションの中で導出し、導出に失敗したら書き込みごとロールバックする。
// F-02: 滞在に紐づく入院の患者変更は拒否し、入院行と患者が食い違う滞在を projection は採用しない。
// DB を使わない単体テスト。commit / rollback は WithTx の callback が error を返したかどうかで代理しており、
// 実 DB のロック・ロールバックの証明ではない。

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type writeTxMarkerKey struct{}

// writeTxProbe は WithTx の結果を記録する。callback が error を返したらロールバック（commit されなかった）扱い。
type writeTxProbe struct {
	committed bool
}

// transactor は ambient tx と tx 内マーカーを載せた ctx で callback を実行する mock transactor を返す。
func (p *writeTxProbe) transactor() *mockTransactor {
	return &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
		txCtx := context.WithValue(persistence.WithTxValue(ctx, &gorm.DB{}), writeTxMarkerKey{}, true)
		err := fn(txCtx)
		p.committed = err == nil
		return err
	}}
}

func inWriteTx(ctx context.Context) bool {
	inTx, _ := ctx.Value(writeTxMarkerKey{}).(bool)
	return inTx
}

// stayReadLog は FindStayCasesByIDs の呼び出しを記録する（書き込み tx の中で呼ばれたか・どの滞在 ID を読んだか）。
type stayReadLog struct {
	inTx  int
	other int
	ids   []uint64
}

// fn は FindStayCasesByIDs のテスト実装を返す。readErr があればそれを返し、なければ petByStay の患者を持つ滞在を返す
// （未指定の滞在の患者は 0）。
func (l *stayReadLog) fn(readErr error, petByStay map[uint64]uint64) func(context.Context, uint64, []uint64) (map[uint64]model.StayCase, error) {
	return func(ctx context.Context, clinicID uint64, ids []uint64) (map[uint64]model.StayCase, error) {
		if inWriteTx(ctx) {
			l.inTx++
		} else {
			l.other++
		}
		l.ids = append(l.ids, ids...)
		if readErr != nil {
			return nil, readErr
		}
		stays := make(map[uint64]model.StayCase, len(ids))
		for _, id := range ids {
			stays[id] = model.StayCase{ID: id, ClinicID: clinicID, PetID: petByStay[id], StayKind: stayKindHotel}
		}
		return stays, nil
	}
}

func newWriteProjectionCreateService(repo *mockHospitalizationRepository, tx *mockTransactor) HospitalizationService {
	reservations := &mockReservationRepository{
		assertOwnerInClinicFn:  func(context.Context, uint64, uint64) error { return nil },
		findPetOwnerInClinicFn: func(context.Context, uint64, uint64) (uint64, error) { return 2, nil },
	}
	pets := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) { return &model.Pet{ID: id}, nil },
	}
	return NewHospitalizationService(repo, reservations, pets, acceptAnyCageRepo(), nil, nil, nil, tx, WithPetLifeStateReader(stubPetLifeReader{}))
}

func TestHospitalizationService_Create_StayProjection(t *testing.T) {
	const ownerID, petID = uint64(2), uint64(5)
	stayID := 900 + ownerID // mockHospitalizationRepository.CreateStayCase の採番
	input := &CreateHospitalizationInput{OwnerID: ownerID, PetID: petID, StayKind: stayKindHotel}
	newRepo := func(reads *stayReadLog, readErr error) *mockHospitalizationRepository {
		repo := &mockHospitalizationRepository{
			createFn: func(_ context.Context, h *model.Hospitalization) error {
				h.ID = 42
				return nil
			},
		}
		repo.findStayCasesByIDsFn = reads.fn(readErr, map[uint64]uint64{stayID: petID})
		return repo
	}

	t.Run("derives the projection inside the create transaction and returns it", func(t *testing.T) {
		probe, reads := &writeTxProbe{}, &stayReadLog{}
		svc := newWriteProjectionCreateService(newRepo(reads, nil), probe.transactor())

		got, err := svc.Create(context.Background(), 1, input)

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, uint64(42), got.ID)
		require.NotNil(t, got.Stay.StayKind)
		assert.Equal(t, stayKindHotel, *got.Stay.StayKind)
		assert.True(t, probe.committed)
		assert.Equal(t, 1, reads.inTx, "projection is read once, inside the write transaction")
		assert.Zero(t, reads.other, "no read outside the write transaction")
		assert.Equal(t, []uint64{stayID}, reads.ids)
	})

	t.Run("a projection failure rolls the create back and reports an error", func(t *testing.T) {
		probe, reads := &writeTxProbe{}, &stayReadLog{}
		repo := newRepo(reads, errors.New("stay read failed"))
		creates := 0
		repo.createFn = func(_ context.Context, h *model.Hospitalization) error {
			creates++
			h.ID = 42
			return nil
		}
		svc := newWriteProjectionCreateService(repo, probe.transactor())

		got, err := svc.Create(context.Background(), 1, input)

		require.Error(t, err)
		assert.Nil(t, got)
		assert.Equal(t, 1, creates, "the create was attempted inside the transaction")
		assert.False(t, probe.committed, "the transaction must not commit when the projection fails")
		assert.Equal(t, 1, reads.inTx)
	})
}

func TestHospitalizationService_Update_StayProjection(t *testing.T) {
	const petID = uint64(5)
	stayID := uint64(31)
	memo := "memo"
	row := func() *model.Hospitalization {
		return &model.Hospitalization{ID: 1, ClinicID: 1, OwnerID: 2, PetID: petID, StayCaseID: &stayID}
	}
	newSvc := func(reads *stayReadLog, readErr error, tx *mockTransactor, updates *int) HospitalizationService {
		repo := &mockHospitalizationRepository{
			findByIDFn: func(context.Context, uint64, uint64) (*model.Hospitalization, error) {
				return row(), nil
			},
			updateFieldsFn: func(context.Context, uint64, uint64, UpdateHospitalizationInput) (*model.Hospitalization, error) {
				*updates++
				return row(), nil
			},
		}
		repo.findStayCasesByIDsFn = reads.fn(readErr, map[uint64]uint64{stayID: petID})
		return NewHospitalizationService(repo, nil, nil, nil, nil, nil, nil, tx, WithPetLifeStateReader(stubPetLifeReader{}))
	}

	t.Run("derives the projection inside the update transaction and returns it", func(t *testing.T) {
		probe, reads, updates := &writeTxProbe{}, &stayReadLog{}, 0
		svc := newSvc(reads, nil, probe.transactor(), &updates)

		got, err := svc.Update(context.Background(), 1, 1, &UpdateHospitalizationInput{Memo: &memo})

		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.Stay.StayKind)
		assert.Equal(t, stayKindHotel, *got.Stay.StayKind)
		assert.True(t, probe.committed)
		assert.Equal(t, 1, updates)
		assert.Equal(t, 1, reads.inTx, "projection is read once, inside the write transaction")
		assert.Zero(t, reads.other, "no read outside the write transaction")
	})

	t.Run("a projection failure rolls the update back and reports an error", func(t *testing.T) {
		probe, reads, updates := &writeTxProbe{}, &stayReadLog{}, 0
		svc := newSvc(reads, errors.New("stay read failed"), probe.transactor(), &updates)

		got, err := svc.Update(context.Background(), 1, 1, &UpdateHospitalizationInput{Memo: &memo})

		require.Error(t, err)
		assert.Nil(t, got)
		assert.Equal(t, 1, updates, "the update was attempted inside the transaction")
		assert.False(t, probe.committed, "the transaction must not commit when the projection fails")
		assert.Equal(t, 1, reads.inTx)
	})
}

func TestHospitalizationService_SwapCages_StayProjection(t *testing.T) {
	const clinicID = uint64(1)
	cage1, cage2 := uint64(1), uint64(2)
	stayA, stayB := uint64(31), uint64(32)
	newSvc := func(reads *stayReadLog, readErr error, tx *mockTransactor, writes *int) HospitalizationService {
		rows := map[uint64]*model.Hospitalization{
			11: {ID: 11, ClinicID: clinicID, PetID: 21, StayCaseID: &stayA, CageID: &cage1},
			12: {ID: 12, ClinicID: clinicID, PetID: 22, StayCaseID: &stayB, CageID: &cage2},
		}
		repo := &mockHospitalizationRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
				return rows[id], nil
			},
			updateFieldsFn: func(_ context.Context, _, id uint64, input UpdateHospitalizationInput) (*model.Hospitalization, error) {
				*writes++
				updated := *rows[id]
				updated.CageID = input.CageID
				return &updated, nil
			},
		}
		repo.findStayCasesByIDsFn = reads.fn(readErr, map[uint64]uint64{stayA: 21, stayB: 22})
		return &hospitalizationService{
			hospRepo: repo, petRepo: swapSnapshotPetRepository{}, cageRepo: swapSnapshotCageRepository{},
			petLife: stubPetLifeReader{}, transactor: tx,
		}
	}

	t.Run("derives both projections inside the swap transaction and returns them in input order", func(t *testing.T) {
		probe, reads, writes := &writeTxProbe{}, &stayReadLog{}, 0
		svc := newSvc(reads, nil, probe.transactor(), &writes)

		gotA, gotB, err := svc.SwapCages(context.Background(), clinicID, 11, 12, [2]uint64{cage1, cage2})

		require.NoError(t, err)
		require.NotNil(t, gotA)
		require.NotNil(t, gotB)
		assert.Equal(t, uint64(11), gotA.ID)
		assert.Equal(t, uint64(12), gotB.ID)
		assert.Equal(t, cage2, *gotA.CageID)
		assert.Equal(t, cage1, *gotB.CageID)
		assert.Equal(t, lifecycleAdmitted, gotA.Stay.Lifecycle)
		assert.Equal(t, lifecycleAdmitted, gotB.Stay.Lifecycle)
		assert.True(t, probe.committed)
		assert.Equal(t, 2, writes)
		assert.Equal(t, 1, reads.inTx, "both stays are read once, inside the swap transaction")
		assert.Zero(t, reads.other, "no read outside the swap transaction")
		assert.Equal(t, []uint64{stayA, stayB}, reads.ids)
	})

	t.Run("a projection failure rolls the swap back and reports an error", func(t *testing.T) {
		probe, reads, writes := &writeTxProbe{}, &stayReadLog{}, 0
		svc := newSvc(reads, errors.New("stay read failed"), probe.transactor(), &writes)

		gotA, gotB, err := svc.SwapCages(context.Background(), clinicID, 11, 12, [2]uint64{cage1, cage2})

		require.Error(t, err)
		assert.Nil(t, gotA)
		assert.Nil(t, gotB)
		assert.Equal(t, 2, writes, "both cage updates were attempted inside the transaction")
		assert.False(t, probe.committed, "the transaction must not commit when the projection fails")
		assert.Equal(t, 1, reads.inTx)
	})
}

func TestHospitalizationService_Update_PetChangeOnStayLinkedHospitalization(t *testing.T) {
	stayID := uint64(31)
	for _, tc := range []struct {
		name         string
		stayCaseID   *uint64
		newPetID     uint64
		wantConflict bool
	}{
		{name: "stay-linked row cannot change its patient", stayCaseID: &stayID, newPetID: 9, wantConflict: true},
		{name: "stay-linked row may resend the same patient", stayCaseID: &stayID, newPetID: 5},
		{name: "row without a stay case is not stay-owned", newPetID: 9},
	} {
		t.Run(tc.name, func(t *testing.T) {
			updates, linkChecks := 0, 0
			repo := &mockHospitalizationRepository{
				findByIDFn: func(context.Context, uint64, uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{ID: 1, ClinicID: 1, OwnerID: 2, PetID: 5, StayCaseID: tc.stayCaseID}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, cmd UpdateHospitalizationInput) (*model.Hospitalization, error) {
					updates++
					return &model.Hospitalization{ID: 1, ClinicID: 1, OwnerID: 2, PetID: *cmd.PetID, StayCaseID: tc.stayCaseID}, nil
				},
			}
			reservations := &mockReservationRepository{
				assertOwnerInClinicFn: func(context.Context, uint64, uint64) error {
					linkChecks++
					return nil
				},
				findPetOwnerInClinicFn: func(context.Context, uint64, uint64) (uint64, error) {
					linkChecks++
					return 2, nil
				},
			}
			svc := NewHospitalizationService(repo, reservations, nil, nil, nil, nil, nil, &mockTransactor{}, WithPetLifeStateReader(stubPetLifeReader{}))

			got, err := svc.Update(context.Background(), 1, 1, &UpdateHospitalizationInput{PetID: &tc.newPetID})

			if tc.wantConflict {
				require.Error(t, err)
				assert.True(t, apperrors.IsConflict(err), "want conflict: %v", err)
				assert.Nil(t, got)
				assert.Zero(t, updates, "nothing is written")
				assert.Zero(t, linkChecks, "rejected before any owner/pet lookup")
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, 1, updates)
		})
	}
}

func TestStayProjections_PetMismatchIsUnknown(t *testing.T) {
	const clinic = uint64(7)
	stayA, stayB := uint64(11), uint64(12)
	rows := []model.Hospitalization{
		{ID: 1, ClinicID: clinic, PetID: 5, StayCaseID: &stayA},
		{ID: 2, ClinicID: clinic, PetID: 6, StayCaseID: &stayB}, // 滞在の患者が入院行の患者と食い違う
	}
	repo := &mockHospitalizationRepository{
		findStayCasesByIDsFn: func(_ context.Context, c uint64, _ []uint64) (map[uint64]model.StayCase, error) {
			return map[uint64]model.StayCase{
				stayA: {ID: stayA, ClinicID: c, PetID: 5, StayKind: stayKindInpatient},
				stayB: {ID: stayB, ClinicID: c, PetID: 99, StayKind: stayKindHotel},
			}, nil
		},
	}
	cages := cageWriteCageFixture{read: func(_ context.Context, c uint64, ids []uint64) ([]StayCageEffectiveState, error) {
		states := make([]StayCageEffectiveState, 0, len(ids))
		for _, id := range ids {
			states = append(states, StayCageEffectiveState{ClinicID: c, StayCaseID: id, State: StayCageAdmitted})
		}
		return states, nil
	}}

	got, err := hospServiceWithStay(repo, cages).StayProjections(context.Background(), clinic, rows)

	require.NoError(t, err)
	require.NotNil(t, got[1].StayKind)
	assert.Equal(t, stayKindInpatient, *got[1].StayKind)
	assert.Equal(t, lifecycleAdmitted, got[1].Lifecycle)
	assert.Equal(t, unknownHospitalizationStay(), got[2], "a stay of another patient is not adopted: kind, interval and lifecycle stay unknown")
}

// snapshotCountingCages は read snapshot runner の呼び出し回数を数える。
type snapshotCountingCages struct {
	cageWriteCageFixture
	snapshots *int
}

func (c snapshotCountingCages) WithCageOccupancyReadSnapshot(ctx context.Context, read func(context.Context) error) error {
	*c.snapshots++
	return read(ctx)
}

// 書き込み tx は ambient tx で滞在を読み（runner は repeatable read 以外の ambient tx を拒否する）、
// 読み取り経路は従来どおり read snapshot runner を使う。
func TestHospitalizationService_StayProjection_SnapshotRunnerOnlyOnReadPaths(t *testing.T) {
	stayID := uint64(31)
	memo := "memo"
	snapshots := 0
	cages := snapshotCountingCages{
		cageWriteCageFixture: cageWriteCageFixture{
			read: func(_ context.Context, clinic uint64, ids []uint64) ([]StayCageEffectiveState, error) {
				states := make([]StayCageEffectiveState, 0, len(ids))
				for _, id := range ids {
					states = append(states, cageWriteState(clinic, id, StayCageAdmitted))
				}
				return states, nil
			},
		},
		snapshots: &snapshots,
	}
	row := func() *model.Hospitalization {
		return &model.Hospitalization{ID: 1, ClinicID: 1, OwnerID: 2, PetID: 5, StayCaseID: &stayID}
	}
	repo := &mockHospitalizationRepository{
		findByIDFn: func(context.Context, uint64, uint64) (*model.Hospitalization, error) {
			return row(), nil
		},
		updateFieldsFn: func(context.Context, uint64, uint64, UpdateHospitalizationInput) (*model.Hospitalization, error) {
			return row(), nil
		},
	}
	repo.findStayCasesByIDsFn = (&stayReadLog{}).fn(nil, map[uint64]uint64{stayID: 5})
	svc := NewHospitalizationService(repo, nil, nil, cages, nil, nil, nil, &mockTransactor{}, WithPetLifeStateReader(stubPetLifeReader{}))

	got, err := svc.Update(context.Background(), 1, 1, &UpdateHospitalizationInput{Memo: &memo})

	require.NoError(t, err)
	assert.Equal(t, lifecycleAdmitted, got.Stay.Lifecycle)
	assert.Zero(t, snapshots, "a write transaction never goes through the read snapshot runner")

	projections, err := svc.StayProjections(context.Background(), 1, []model.Hospitalization{*got.Hospitalization})

	require.NoError(t, err)
	assert.Equal(t, 1, snapshots, "read paths keep using the read snapshot runner")
	assert.Equal(t, lifecycleAdmitted, projections[got.ID].Lifecycle)
}
