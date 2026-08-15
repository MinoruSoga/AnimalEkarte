package reservation

// reservation_clinic_isolation_test.go — #196 clinic_id テナント隔離回帰テスト
//
// テスト対象: ReservationRepository の clinic_id 境界
// 保護する不変条件: clinic A のスコープで clinic B の予約を
//   読み取れない、更新できない、削除できない。
//
// このテストは clinicScope を削除すると必ず失敗するよう設計されている。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupReservationIsolationTestDB は予約 clinic_id 隔離テスト用の DB を整備する。
// setupTestDB が owners/billings 系をクリアし、ここで reservation_types/appointments を追加する。
// ekarte_db_test は本番 migration 適用済みのため fk_appointments_reservation_type FK 制約が存在する。
// TRUNCATE reservation_types CASCADE で appointments も連鎖クリアされる。
func setupReservationIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	// setupTestDB は per-call で ENUM を DROP しない。ensureAutoMigrated で reservation 系テーブルを揃える。
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{},
		&model.Clinic{},
		&model.Owner{},
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.ReservationTypeGroup{},
		&model.ReservationType{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.LineCustomer{},
		&model.Reservation{},
	))
	// reservation_types CASCADE により fk_appointments_reservation_type 経由で
	// appointments も連鎖クリアされる。
	require.NoError(t, db.Exec(`
		TRUNCATE TABLE
			appointments,
			staff_clinic_assignments,
			staffs,
			reservation_types,
			reservation_type_groups,
			pets,
			animal_species,
			line_customers,
			owners
		CASCADE
	`).Error)
	return db
}

func makeReservationClinicCorrelationClinic(t *testing.T, db *gorm.DB, id uint64, name string) {
	t.Helper()
	company := &model.Company{Name: name + "法人"}
	require.NoError(t, db.Create(company).Error)
	clinic := &model.Clinic{ID: id, CompanyID: company.ID, Name: name}
	require.NoError(t, db.Clauses(clause.OnConflict{DoNothing: true}).Create(clinic).Error)
}

func makeReservationClinicCorrelationPet(
	t *testing.T,
	db *gorm.DB,
	clinicID, ownerID, speciesID uint64,
	name string,
) *model.Pet {
	t.Helper()
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         ownerID,
		AnimalSpeciesID: speciesID,
		Name:            name,
	}
	require.NoError(t, db.Create(pet).Error)
	return pet
}

func makeReservationClinicCorrelationStaff(
	t *testing.T,
	db *gorm.DB,
	primaryClinicID, assignedClinicID uint64,
	name string,
) *model.Staff {
	t.Helper()
	staff := makeDoctor(t, db, primaryClinicID, name)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID:  staff.ID,
		ClinicID: assignedClinicID,
	}).Error)
	return staff
}

func makeReservationClinicCorrelationAppointment(
	t *testing.T,
	db *gorm.DB,
	clinicID, reservationTypeID, ownerID, petID, doctorID, createdByID uint64,
) *model.Reservation {
	t.Helper()
	reservation := makeReservation(t, db, clinicID)
	updates := struct {
		ReservationTypeID uint64
		OwnerID           *uint64
		PetID             *uint64
		DoctorID          *uint64
		CreatedBy         *uint64
	}{
		ReservationTypeID: reservationTypeID,
		OwnerID:           &ownerID,
		PetID:             &petID,
		DoctorID:          &doctorID,
		CreatedBy:         &createdByID,
	}
	require.NoError(t, db.Model(reservation).Updates(updates).Error)
	return reservation
}

func TestReservationRepository_MultiClinicReadsCorrelateRelationsToParentClinic(t *testing.T) {
	db := setupReservationIsolationTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	authorizedClinicIDs := []uint64{clinicA, clinicB}

	makeReservationClinicCorrelationClinic(t, db, clinicA, "医院A")
	makeReservationClinicCorrelationClinic(t, db, clinicB, "医院B")

	species := &model.AnimalSpecies{Name: "犬", IsActive: true}
	require.NoError(t, db.Create(species).Error)
	ownerA := testdb.MakeTestOwner(t, db, clinicA, "医院A飼主")
	otherOwnerA := testdb.MakeTestOwner(t, db, clinicA, "医院A別飼主")
	ownerB := testdb.MakeTestOwner(t, db, clinicB, "医院B飼主")
	petA := makeReservationClinicCorrelationPet(t, db, clinicA, ownerA.ID, species.ID, "医院A患者")
	petAForOtherOwner := makeReservationClinicCorrelationPet(
		t, db, clinicA, otherOwnerA.ID, species.ID, "医院A別飼主の患者",
	)
	petB := makeReservationClinicCorrelationPet(t, db, clinicB, ownerB.ID, species.ID, "医院B患者")
	petAWithForeignOwner := makeReservationClinicCorrelationPet(
		t,
		db,
		clinicA,
		ownerB.ID,
		species.ID,
		"医院A所属・医院B飼主の破損患者",
	)

	groupA := &model.ReservationTypeGroup{ClinicID: clinicA, Name: "医院Aグループ", IsActive: true}
	groupB := &model.ReservationTypeGroup{ClinicID: clinicB, Name: "医院Bグループ", IsActive: true}
	require.NoError(t, db.Create(groupA).Error)
	require.NoError(t, db.Create(groupB).Error)
	typeA := &model.ReservationType{
		ClinicID: clinicA, Name: "医院A区分", Category: model.ReservationTypeCategoryGeneral, GroupID: &groupA.ID,
	}
	typeB := &model.ReservationType{
		ClinicID: clinicB, Name: "医院B区分", Category: model.ReservationTypeCategoryGeneral, GroupID: &groupB.ID,
	}
	typeAWithForeignGroup := &model.ReservationType{
		ClinicID: clinicA, Name: "医院A所属・医院Bグループの破損区分",
		Category: model.ReservationTypeCategoryGeneral, GroupID: &groupB.ID,
	}
	require.NoError(t, db.Create(typeA).Error)
	require.NoError(t, db.Create(typeB).Error)
	require.NoError(t, db.Create(typeAWithForeignGroup).Error)

	doctorA := makeReservationClinicCorrelationStaff(t, db, clinicA, clinicA, "医院A医師")
	doctorB := makeReservationClinicCorrelationStaff(t, db, clinicB, clinicB, "医院B医師")
	creatorA := makeReservationClinicCorrelationStaff(t, db, clinicA, clinicA, "医院A作成者")
	creatorB := makeReservationClinicCorrelationStaff(t, db, clinicB, clinicB, "医院B作成者")
	lineCustomerA := makeLineCustomerForAdmin(t, db, clinicA, "line-correlation-a")
	lineCustomerB := makeLineCustomerForAdmin(t, db, clinicB, "line-correlation-b")

	valid := makeReservationClinicCorrelationAppointment(
		t, db, clinicA, typeA.ID, ownerA.ID, petA.ID, doctorA.ID, creatorA.ID,
	)
	require.NoError(t, db.Model(valid).Update("line_customer_id", lineCustomerA.ID).Error)

	tests := []struct {
		name          string
		reservationID uint64
	}{
		{
			name: "foreign owner",
			reservationID: makeReservationClinicCorrelationAppointment(
				t, db, clinicA, typeA.ID, ownerB.ID, petA.ID, doctorA.ID, creatorA.ID,
			).ID,
		},
		{
			name: "foreign pet",
			reservationID: makeReservationClinicCorrelationAppointment(
				t, db, clinicA, typeA.ID, ownerA.ID, petB.ID, doctorA.ID, creatorA.ID,
			).ID,
		},
		{
			name: "same-clinic pet with foreign owner",
			reservationID: makeReservationClinicCorrelationAppointment(
				t, db, clinicA, typeA.ID, ownerA.ID, petAWithForeignOwner.ID, doctorA.ID, creatorA.ID,
			).ID,
		},
		{
			name: "same-clinic owner and pet belonging to different owners",
			reservationID: makeReservationClinicCorrelationAppointment(
				t, db, clinicA, typeA.ID, ownerA.ID, petAForOtherOwner.ID, doctorA.ID, creatorA.ID,
			).ID,
		},
		{
			name: "foreign reservation type",
			reservationID: makeReservationClinicCorrelationAppointment(
				t, db, clinicA, typeB.ID, ownerA.ID, petA.ID, doctorA.ID, creatorA.ID,
			).ID,
		},
		{
			name: "same-clinic reservation type with foreign group",
			reservationID: makeReservationClinicCorrelationAppointment(
				t, db, clinicA, typeAWithForeignGroup.ID, ownerA.ID, petA.ID, doctorA.ID, creatorA.ID,
			).ID,
		},
		{
			name: "doctor assigned only to foreign clinic",
			reservationID: makeReservationClinicCorrelationAppointment(
				t, db, clinicA, typeA.ID, ownerA.ID, petA.ID, doctorB.ID, creatorA.ID,
			).ID,
		},
		{
			name: "creator assigned only to foreign clinic",
			reservationID: makeReservationClinicCorrelationAppointment(
				t, db, clinicA, typeA.ID, ownerA.ID, petA.ID, doctorA.ID, creatorB.ID,
			).ID,
		},
		{
			name: "foreign LINE customer",
			reservationID: func() uint64 {
				reservation := makeReservationClinicCorrelationAppointment(
					t, db, clinicA, typeA.ID, ownerA.ID, petA.ID, doctorA.ID, creatorA.ID,
				)
				require.NoError(t, db.Model(reservation).Update("line_customer_id", lineCustomerB.ID).Error)
				return reservation.ID
			}(),
		},
	}

	t.Run("list excludes corrupted parents even when both clinics are authorized", func(t *testing.T) {
		got, total, err := repo.FindAll(
			ctx, authorizedClinicIDs, 1, 100, nil, nil, nil, nil, nil, nil, nil,
		)
		require.NoError(t, err)
		assert.EqualValues(t, 1, total)
		require.Len(t, got, 1)
		assert.Equal(t, valid.ID, got[0].ID)
		require.NotNil(t, got[0].Owner)
		require.NotNil(t, got[0].Pet)
		require.NotNil(t, got[0].Pet.Owner)
		require.NotNil(t, got[0].ReservationType)
		require.NotNil(t, got[0].ReservationType.Group)
		require.NotNil(t, got[0].Doctor)
	})

	t.Run("by ID preserves same-clinic relations", func(t *testing.T) {
		got, err := repo.FindByIDForClinics(ctx, authorizedClinicIDs, valid.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.Owner)
		require.NotNil(t, got.Pet)
		require.NotNil(t, got.Pet.Owner)
		require.NotNil(t, got.ReservationType)
		require.NotNil(t, got.ReservationType.Group)
		require.NotNil(t, got.Doctor)
		require.NotNil(t, got.CreatedByStaff)
		require.NotNil(t, got.LineCustomerID)
		assert.Equal(t, lineCustomerA.ID, *got.LineCustomerID)
	})

	for _, tt := range tests {
		t.Run("by ID excludes "+tt.name, func(t *testing.T) {
			got, err := repo.FindByIDForClinics(ctx, authorizedClinicIDs, tt.reservationID)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.True(t, apperrors.IsNotFound(err), "error must be NotFound: %v", err)
		})
	}
}

func TestReservationRepository_MultiClinicReadsKeepHistoricalParentWithSoftDeletedRelations(t *testing.T) {
	db := setupReservationIsolationTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	makeReservationClinicCorrelationClinic(t, db, clinicID, "履歴医院")
	species := &model.AnimalSpecies{Name: "履歴犬", IsActive: true}
	require.NoError(t, db.Create(species).Error)
	owner := testdb.MakeTestOwner(t, db, clinicID, "削除済み飼主")
	pet := makeReservationClinicCorrelationPet(t, db, clinicID, owner.ID, species.ID, "削除済み患者")
	group := &model.ReservationTypeGroup{
		ClinicID: clinicID,
		Name:     "削除済みグループ",
		IsActive: true,
	}
	require.NoError(t, db.Create(group).Error)
	reservationType := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "削除済み区分",
		Category: model.ReservationTypeCategoryGeneral,
		GroupID:  &group.ID,
	}
	require.NoError(t, db.Create(reservationType).Error)
	staff := makeReservationClinicCorrelationStaff(t, db, clinicID, clinicID, "削除済み担当者")
	reservation := makeReservationClinicCorrelationAppointment(
		t,
		db,
		clinicID,
		reservationType.ID,
		owner.ID,
		pet.ID,
		staff.ID,
		staff.ID,
	)

	require.NoError(t, db.Delete(reservationType).Error)
	require.NoError(t, db.Delete(group).Error)
	require.NoError(t, db.Delete(pet).Error)
	require.NoError(t, db.Delete(owner).Error)
	assignmentDelete := db.
		Where("staff_id = ? AND clinic_id = ?", staff.ID, clinicID).
		Delete(&model.StaffClinicAssignment{})
	require.NoError(t, assignmentDelete.Error)
	require.EqualValues(t, 1, assignmentDelete.RowsAffected)
	require.NoError(t, db.Delete(staff).Error)

	tests := []struct {
		name   string
		isList bool
	}{
		{name: "list keeps historical parent and count", isList: true},
		{name: "by ID keeps historical parent", isList: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got *model.Reservation
			if tt.isList {
				items, total, err := repo.FindAll(
					ctx, []uint64{clinicID}, 1, 10, nil, nil, nil, nil, nil, nil, nil,
				)
				require.NoError(t, err)
				assert.EqualValues(t, 1, total)
				require.Len(t, items, 1)
				got = &items[0]
			} else {
				var err error
				got, err = repo.FindByIDForClinics(ctx, []uint64{clinicID}, reservation.ID)
				require.NoError(t, err)
				require.NotNil(t, got)
			}

			assert.Equal(t, reservation.ID, got.ID)
			assert.Nil(t, got.Owner)
			assert.Nil(t, got.Pet)
			assert.Nil(t, got.ReservationType)
			assert.Nil(t, got.Doctor)
			assert.Nil(t, got.CreatedByStaff)
		})
	}
}

// TestReservationRepository_FindByID_ClinicIsolation は
// clinic A の予約を clinic B の clinicID で取得できないことを検証する。
// clinicScope を reservation_repository.go から削除すると「別クリニックIDでは取得できない」が失敗する。
func TestReservationRepository_FindByID_ClinicIsolation(t *testing.T) {
	db := setupReservationIsolationTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	resA := makeReservation(t, db, clinicA)

	t.Run("同一クリニックIDでは取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, resA.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, resA.ID, got.ID)
		assert.Equal(t, clinicA, got.ClinicID)
	})

	t.Run("別クリニックIDでは取得できない（clinic_id 隔離）", func(t *testing.T) {
		// clinicScope が有効なら clinic B から clinic A の予約は見えない。
		got, err := repo.FindByID(ctx, clinicB, resA.ID)
		assert.Error(t, err, "clinic B から clinic A の予約を取得できてはならない")
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})
}

// TestReservationRepository_Update_ClinicIsolation は
// 別クリニックIDからの Update が NotFound を返し、行が変更されないことを検証する。
// clinicScope を削除すると「行が変更されていない」が失敗する。
func TestReservationRepository_Update_ClinicIsolation(t *testing.T) {
	db := setupReservationIsolationTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	resA := makeReservation(t, db, clinicA)

	t.Run("別クリニックIDからの Update は NotFound を返す", func(t *testing.T) {
		_, err := repo.update(ctx, clinicB, resA.ID, map[string]any{"notes": "不正書き換え"})
		require.Error(t, err, "clinic B から clinic A の予約を更新できてはならない")
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("行が変更されていない（データ改ざん防止）", func(t *testing.T) {
		// プリロードを使わず直接 DB 読み取りで notes が変化していないことを確認する。
		var r model.Reservation
		require.NoError(t, db.Where("id = ? AND clinic_id = ?", resA.ID, clinicA).First(&r).Error)
		assert.Equal(t, "", r.Notes, "別クリニックからの Update で notes が変わってはならない")
	})

	t.Run("正しいクリニックIDからの Update は成功する", func(t *testing.T) {
		got, err := repo.update(ctx, clinicA, resA.ID, map[string]any{"notes": "正常更新"})
		require.NoError(t, err)
		assert.Equal(t, "正常更新", got.Notes)
	})
}

// TestReservationRepository_Delete_ClinicIsolation は
// 別クリニックIDからの Delete が NotFound を返し、予約が削除されないことを検証する。
// clinicScope を削除すると「予約はまだ存在する」が失敗する。
func TestReservationRepository_Delete_ClinicIsolation(t *testing.T) {
	db := setupReservationIsolationTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	resA := makeReservation(t, db, clinicA)

	t.Run("別クリニックIDからの Delete は NotFound を返す", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, resA.ID)
		require.Error(t, err, "clinic B から clinic A の予約を削除できてはならない")
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("予約はまだ存在する（不正削除防止）", func(t *testing.T) {
		// プリロードを使わず直接 DB 読み取りで行が残存することを確認する。
		var r model.Reservation
		require.NoError(t, db.Where("id = ? AND clinic_id = ?", resA.ID, clinicA).First(&r).Error)
		assert.Equal(t, resA.ID, r.ID, "clinic A の予約は別クリニックの Delete で消えてはならない")
	})
}
