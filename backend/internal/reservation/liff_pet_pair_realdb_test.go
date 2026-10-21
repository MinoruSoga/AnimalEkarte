package reservation

// liff_pet_pair_realdb_test.go — FindLiffPetPairs (EMR-520) の real DB 統合テスト。
//
// 保護する不変条件: 認可パス（同一医院・本人 line_customer・state='linked' 紐付け・
// 非削除 pet/owner・pet.owner_id = appointment.owner_id）を通る場合だけ pet ペアが返り、
// common 現行 revision 名とローカル pet_id が一致する。未解決は map に含まれない
// （= 応答の両方 null）。common UUID / revision ID は公開しない。

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupLiffPetPairTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.ReservationType{}, &model.Reservation{},
		&model.AnimalSpecies{}, &model.Pet{}, &model.Owner{}, &model.LineCustomer{},
		&model.Company{}, &model.Clinic{}, &model.Staff{},
		&model.CommonPetSubject{}, &model.CommonPetClinicBinding{}, &model.CommonPetBasicRevision{},
		&model.OwnerLineBindingState{},
	))
	testdb.Truncate(t, db,
		"reservation_types", "pets", "owners", "animal_species", "line_customers",
		"common_pet_basic_revisions", "common_pet_clinic_bindings",
		"common_pet_subjects", "owner_line_binding_states",
	)
	return db
}

// linkCommonPetForLiff はローカル pet を common subject + binding + 現行 revision に紐付ける。
// revisions は引数の順に revision=1..n として作り、最後のものを current にする
// （revision update で最新が勝つ経路の検証に使う）。
func linkCommonPetForLiff(t *testing.T, db *gorm.DB, clinicID, staffID uint64, pet *model.Pet, speciesID uint64, names ...string) {
	t.Helper()
	commonID := uuid.New()
	require.NoError(t, db.Create(&model.CommonPetSubject{ID: commonID}).Error)
	require.NoError(t, db.Create(&model.CommonPetClinicBinding{
		ClinicID: clinicID, CommonPetID: commonID, LocalPetID: pet.ID,
		ConfirmedBy: staffID, ConfirmedAt: time.Now(), OperationID: 1,
	}).Error)
	var last *model.CommonPetBasicRevision
	for i, name := range names {
		last = &model.CommonPetBasicRevision{
			CommonPetID: commonID, Revision: int64(i + 1),
			SourceClinicID: clinicID, SourceLocalPetID: pet.ID, OperationID: 1,
			Name: name, AnimalSpeciesID: speciesID,
			NeuteredStatus: model.PetNeuteredStatusUnknown,
			Provenance:     json.RawMessage(`{}`),
			RecordedBy:     staffID, RecordedAt: time.Now(),
		}
		require.NoError(t, db.Create(last).Error)
	}
	require.NoError(t, db.Model(&model.CommonPetSubject{}).
		Where("id = ?", commonID).
		Update("current_basic_revision_id", last.ID).Error)
}

func linkLineCustomerForLiff(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, customer *model.LineCustomer, state string) {
	t.Helper()
	require.NoError(t, db.Create(&model.OwnerLineBindingState{
		ClinicID: clinicID, OwnerID: ownerID, Revision: 1,
		State: state, LineCustomerID: &customer.ID, OperationID: 1,
	}).Error)
}

func makeLiffPairPet(t *testing.T, db *gorm.DB, clinicID, ownerID uint64) *model.Pet {
	t.Helper()
	pet := &model.Pet{ClinicID: clinicID, OwnerID: ownerID}
	require.NoError(t, db.WithContext(context.Background()).Create(pet).Error)
	return pet
}

func makeLiffPairSpecies(t *testing.T, db *gorm.DB) *model.AnimalSpecies {
	t.Helper()
	s := &model.AnimalSpecies{Name: "犬"}
	require.NoError(t, db.WithContext(context.Background()).Create(s).Error)
	return s
}

func TestFindLiffPetPairs(t *testing.T) {
	db := setupLiffPetPairTestDB(t)
	repo := NewReservationAdminRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	start := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)

	staff := testdb.MakeDoctor(t, db, clinicA, "Dr.A")
	species := makeLiffPairSpecies(t, db)
	owner := testdb.MakeTestOwner(t, db, clinicA, "山田")
	customer := makeLineCustomerForAdmin(t, db, clinicA, "U-petpair-001")
	linkLineCustomerForLiff(t, db, clinicA, owner.ID, customer, "linked")
	pet := makeLiffPairPet(t, db, clinicA, owner.ID)
	linkCommonPetForLiff(t, db, clinicA, staff.ID, pet, species.ID, "タマ")

	newAppt := func(ownerID, petID, customerID *uint64) *model.Reservation {
		return makeAdminReservationAt(t, db, clinicA, start, ownerID, petID, nil, customerID)
	}

	t.Run("認可パス成立で local pet_id と現行 revision 名が返る（customer_fields の名前は使わない）", func(t *testing.T) {
		appt := newAppt(&owner.ID, &pet.ID, &customer.ID)
		appt.CustomerFields = json.RawMessage(`{"pets":[{"name":"ポチA"}]}`)
		require.NoError(t, db.Save(appt).Error)

		got, err := repo.FindLiffPetPairs(ctx, clinicA, customer.ID, []uint64{appt.ID})
		require.NoError(t, err)
		require.Contains(t, got, appt.ID)
		assert.Equal(t, LiffPetPair{PetID: pet.ID, PetName: "タマ"}, got[appt.ID])
	})

	t.Run("同名の別ペットは別の pet_id で解決される", func(t *testing.T) {
		pet2 := makeLiffPairPet(t, db, clinicA, owner.ID)
		linkCommonPetForLiff(t, db, clinicA, staff.ID, pet2, species.ID, "タマ")
		a1 := newAppt(&owner.ID, &pet.ID, &customer.ID)
		a2 := newAppt(&owner.ID, &pet2.ID, &customer.ID)

		got, err := repo.FindLiffPetPairs(ctx, clinicA, customer.ID, []uint64{a1.ID, a2.ID})
		require.NoError(t, err)
		assert.Equal(t, pet.ID, got[a1.ID].PetID)
		assert.Equal(t, pet2.ID, got[a2.ID].PetID)
		assert.Equal(t, "タマ", got[a1.ID].PetName)
		assert.Equal(t, "タマ", got[a2.ID].PetName)
	})

	t.Run("pet が binding を持たない（pet 差し替え相当）→ 含まれない", func(t *testing.T) {
		unbound := makeLiffPairPet(t, db, clinicA, owner.ID)
		appt := newAppt(&owner.ID, &unbound.ID, &customer.ID)

		got, err := repo.FindLiffPetPairs(ctx, clinicA, customer.ID, []uint64{appt.ID})
		require.NoError(t, err)
		assert.NotContains(t, got, appt.ID)
	})

	t.Run("pet_id 未解決（NULL）→ 含まれない", func(t *testing.T) {
		appt := newAppt(&owner.ID, nil, &customer.ID)
		got, err := repo.FindLiffPetPairs(ctx, clinicA, customer.ID, []uint64{appt.ID})
		require.NoError(t, err)
		assert.NotContains(t, got, appt.ID)
	})

	t.Run("binding が unlinked / unlink_pending → 含まれない", func(t *testing.T) {
		ownerU := testdb.MakeTestOwner(t, db, clinicA, "非連携")
		customerU := makeLineCustomerForAdmin(t, db, clinicA, "U-petpair-unlinked")
		petU := makeLiffPairPet(t, db, clinicA, ownerU.ID)
		linkCommonPetForLiff(t, db, clinicA, staff.ID, petU, species.ID, "コロ")
		appt := newAppt(&ownerU.ID, &petU.ID, &customerU.ID)

		for _, state := range []string{"unlinked", "unlink_pending"} {
			linkLineCustomerForLiff(t, db, clinicA, ownerU.ID, customerU, state)
			got, err := repo.FindLiffPetPairs(ctx, clinicA, customerU.ID, []uint64{appt.ID})
			require.NoError(t, err, "state=%s", state)
			assert.NotContains(t, got, appt.ID, "state=%s", state)
			require.NoError(t, db.Where("clinic_id = ? AND owner_id = ?", clinicA, ownerU.ID).
				Delete(&model.OwnerLineBindingState{}).Error)
		}
	})

	t.Run("別医院の binding/local ID は混ざらない", func(t *testing.T) {
		ownerB := testdb.MakeTestOwner(t, db, clinicB, "他院")
		customerB := makeLineCustomerForAdmin(t, db, clinicB, "U-petpair-clinicB")
		linkLineCustomerForLiff(t, db, clinicB, ownerB.ID, customerB, "linked")
		petB := makeLiffPairPet(t, db, clinicB, ownerB.ID)
		linkCommonPetForLiff(t, db, clinicB, staff.ID, petB, species.ID, "タマ")

		// clinicB の local pet_id と同じ ID の pet を clinicA で bound にしても、
		// clinicA の予約は clinicB の pet を解決しない（clinic_id 等号で隔離）。
		apptA := newAppt(&owner.ID, &pet.ID, &customer.ID)
		got, err := repo.FindLiffPetPairs(ctx, clinicA, customer.ID, []uint64{apptA.ID})
		require.NoError(t, err)
		assert.Equal(t, pet.ID, got[apptA.ID].PetID)
		assert.Equal(t, "タマ", got[apptA.ID].PetName)

		gotB, err := repo.FindLiffPetPairs(ctx, clinicB, customerB.ID, []uint64{apptA.ID})
		require.NoError(t, err)
		assert.Empty(t, gotB)
	})

	t.Run("revision 更新では最新の current revision が勝つ", func(t *testing.T) {
		petV := makeLiffPairPet(t, db, clinicA, owner.ID)
		linkCommonPetForLiff(t, db, clinicA, staff.ID, petV, species.ID, "旧名", "新名")
		appt := newAppt(&owner.ID, &petV.ID, &customer.ID)

		got, err := repo.FindLiffPetPairs(ctx, clinicA, customer.ID, []uint64{appt.ID})
		require.NoError(t, err)
		assert.Equal(t, "新名", got[appt.ID].PetName)
	})

	t.Run("他人の予約は返らない", func(t *testing.T) {
		other := makeLineCustomerForAdmin(t, db, clinicA, "U-petpair-other")
		appt := newAppt(&owner.ID, &pet.ID, &customer.ID)

		got, err := repo.FindLiffPetPairs(ctx, clinicA, other.ID, []uint64{appt.ID})
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("予約の pet が別 pet に差し替え → 旧名は出ず、認可済みなら新 pet のみ", func(t *testing.T) {
		petNew := makeLiffPairPet(t, db, clinicA, owner.ID)
		linkCommonPetForLiff(t, db, clinicA, staff.ID, petNew, species.ID, "ハク")
		appt := newAppt(&owner.ID, &pet.ID, &customer.ID)
		require.NoError(t, db.Model(appt).Update("pet_id", petNew.ID).Error)

		got, err := repo.FindLiffPetPairs(ctx, clinicA, customer.ID, []uint64{appt.ID})
		require.NoError(t, err)
		require.Contains(t, got, appt.ID)
		assert.Equal(t, petNew.ID, got[appt.ID].PetID)
		assert.Equal(t, "ハク", got[appt.ID].PetName)

		unbound := makeLiffPairPet(t, db, clinicA, owner.ID)
		require.NoError(t, db.Model(appt).Update("pet_id", unbound.ID).Error)
		got, err = repo.FindLiffPetPairs(ctx, clinicA, customer.ID, []uint64{appt.ID})
		require.NoError(t, err)
		assert.NotContains(t, got, appt.ID)
	})

	t.Run("pet.owner_id が予約の owner と不一致 → 含まれない", func(t *testing.T) {
		owner2 := testdb.MakeTestOwner(t, db, clinicA, "別飼主")
		petOtherOwner := makeLiffPairPet(t, db, clinicA, owner2.ID)
		linkCommonPetForLiff(t, db, clinicA, staff.ID, petOtherOwner, species.ID, "チビ")
		appt := newAppt(&owner.ID, &petOtherOwner.ID, &customer.ID)

		got, err := repo.FindLiffPetPairs(ctx, clinicA, customer.ID, []uint64{appt.ID})
		require.NoError(t, err)
		assert.NotContains(t, got, appt.ID)
	})

	t.Run("soft-deleted pet → 含まれない", func(t *testing.T) {
		petDel := makeLiffPairPet(t, db, clinicA, owner.ID)
		linkCommonPetForLiff(t, db, clinicA, staff.ID, petDel, species.ID, "モモ")
		appt := newAppt(&owner.ID, &petDel.ID, &customer.ID)
		require.NoError(t, db.Delete(&model.Pet{}, petDel.ID).Error)

		got, err := repo.FindLiffPetPairs(ctx, clinicA, customer.ID, []uint64{appt.ID})
		require.NoError(t, err)
		assert.NotContains(t, got, appt.ID)
	})

	t.Run("soft-deleted owner → 含まれない", func(t *testing.T) {
		ownerDel := testdb.MakeTestOwner(t, db, clinicA, "削除飼主")
		customerDel := makeLineCustomerForAdmin(t, db, clinicA, "U-petpair-delowner")
		linkLineCustomerForLiff(t, db, clinicA, ownerDel.ID, customerDel, "linked")
		petDel := makeLiffPairPet(t, db, clinicA, ownerDel.ID)
		linkCommonPetForLiff(t, db, clinicA, staff.ID, petDel, species.ID, "ルナ")
		appt := newAppt(&ownerDel.ID, &petDel.ID, &customerDel.ID)
		require.NoError(t, db.Delete(&model.Owner{}, ownerDel.ID).Error)

		got, err := repo.FindLiffPetPairs(ctx, clinicA, customerDel.ID, []uint64{appt.ID})
		require.NoError(t, err)
		assert.NotContains(t, got, appt.ID)
	})

	t.Run("linked でも line_customer_id が別顧客 → 含まれない", func(t *testing.T) {
		ownerX := testdb.MakeTestOwner(t, db, clinicA, "別顧客飼主")
		customerX := makeLineCustomerForAdmin(t, db, clinicA, "U-petpair-x")
		otherCustomer := makeLineCustomerForAdmin(t, db, clinicA, "U-petpair-y")
		linkLineCustomerForLiff(t, db, clinicA, ownerX.ID, otherCustomer, "linked")
		petX := makeLiffPairPet(t, db, clinicA, ownerX.ID)
		linkCommonPetForLiff(t, db, clinicA, staff.ID, petX, species.ID, "クロ")
		appt := newAppt(&ownerX.ID, &petX.ID, &customerX.ID)

		got, err := repo.FindLiffPetPairs(ctx, clinicA, customerX.ID, []uint64{appt.ID})
		require.NoError(t, err)
		assert.NotContains(t, got, appt.ID)
	})

	t.Run("appointmentIDs が空なら空 map（クエリを発行しない）", func(t *testing.T) {
		got, err := repo.FindLiffPetPairs(ctx, clinicA, customer.ID, nil)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}
