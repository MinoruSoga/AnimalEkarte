package repository

// cross_clinic_preload_isolation_test.go
// クロステナント READ IDOR remediation follow-up — (c) cross-clinic(#86)変種の master Preload 隔離。
//
// owner/pet の Insurance・reservation の ReservationType/Group は単一 clinicID だけでなく
// 拠点横断(#86)の clinicID 集合でも読まれる。Preload は `clinic_id IN ?`(認可集合)でスコープし、
//   (i)  別テナント単独の混入を拒否（FindByID）
//   (ii) 拠点横断ユーザは認可済み複数 clinic のマスタを引き続き読める（FindByIDForClinics([A,B])）
//   (iii)認可外 clinic のマスタは #86 でも隠れる（FindByIDForClinics([A]) で B のマスタは nil）
// の3点を同時に満たすこと。naive `clinic_id = ?` だと (ii) を壊す。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

func makeInsuranceMaster(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Insurance {
	t.Helper()
	ins := &model.Insurance{ClinicID: clinicID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(ins).Error)
	return ins
}

func makePetWithInsurance(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, insuranceID *uint64, name string) *model.Pet {
	t.Helper()
	species := &model.AnimalSpecies{Name: "犬"}
	require.NoError(t, db.WithContext(context.Background()).Create(species).Error)
	pet := &model.Pet{ClinicID: clinicID, OwnerID: ownerID, AnimalSpeciesID: species.ID, Name: name, InsuranceID: insuranceID}
	require.NoError(t, db.WithContext(context.Background()).Create(pet).Error)
	return pet
}

func setupInsurancePreloadTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	db.Exec("TRUNCATE TABLE insurances CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	require.NoError(t, db.AutoMigrate(&model.AnimalSpecies{}, &model.Pet{}, &model.Insurance{}))
	return db
}

// --- (c1) pet: Insurance ---

func TestPetRepository_Insurance_CrossClinicPreloadIsolation(t *testing.T) {
	db := setupInsurancePreloadTestDB(t)
	repo := NewPetRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	insA := makeInsuranceMaster(t, db, clinicA, "医院Aの保険")
	insB := makeInsuranceMaster(t, db, clinicB, "医院Bの保険")
	ownerA := makeTestOwner(t, db, clinicA, "保険飼主A")
	petLegit := makePetWithInsurance(t, db, clinicA, ownerA.ID, &insA.ID, "正規保険ペット")
	petCross := makePetWithInsurance(t, db, clinicA, ownerA.ID, &insB.ID, "越境保険ペット") // 別clinicの保険FKを植え付け

	// (legit) 単一 clinic の正規保険は Preload される
	gotLegit, err := repo.FindByID(ctx, clinicA, petLegit.ID)
	require.NoError(t, err)
	require.NotNil(t, gotLegit.Insurance)
	assert.Equal(t, insA.ID, gotLegit.Insurance.ID)

	// (i) 別テナント保険は FindByID(単一)で混入しない
	gotCross, err := repo.FindByID(ctx, clinicA, petCross.ID)
	require.NoError(t, err)
	assert.Nil(t, gotCross.Insurance, "別クリニックの保険マスタが混入してはならない")

	// (ii) #86: 認可集合 [A,B] なら B の保険は見える
	gotForBoth, err := repo.FindByIDForClinics(ctx, []uint64{clinicA, clinicB}, petCross.ID)
	require.NoError(t, err)
	require.NotNil(t, gotForBoth.Insurance, "#86 で B も認可済みなら B の保険は見えるべき")
	assert.Equal(t, insB.ID, gotForBoth.Insurance.ID)

	// (iii) #86: 認可集合 [A] のみなら B の保険は隠れる
	gotForAOnly, err := repo.FindByIDForClinics(ctx, []uint64{clinicA}, petCross.ID)
	require.NoError(t, err)
	assert.Nil(t, gotForAOnly.Insurance, "認可外 clinic の保険マスタは #86 でも漏れてはならない")
}

// --- (c2) owner: Pets.Insurance (nested) ---

func TestOwnerRepository_PetsInsurance_CrossClinicPreloadIsolation(t *testing.T) {
	db := setupInsurancePreloadTestDB(t)
	repo := NewOwnerRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	insB := makeInsuranceMaster(t, db, clinicB, "医院Bの保険(owner)")
	ownerCross := makeTestOwner(t, db, clinicA, "越境保険飼主")
	makePetWithInsurance(t, db, clinicA, ownerCross.ID, &insB.ID, "越境保険ペット(owner)")

	// (i) FindByID(単一)で nested Pets.Insurance に別テナント保険が混入しない
	got, err := repo.FindByID(ctx, clinicA, ownerCross.ID)
	require.NoError(t, err)
	require.Len(t, got.Pets, 1)
	assert.Nil(t, got.Pets[0].Insurance, "別クリニックの保険マスタが Pets.Insurance に混入してはならない")

	// (ii) #86 [A,B] なら B の保険は見える
	gotBoth, err := repo.FindByIDForClinics(ctx, []uint64{clinicA, clinicB}, ownerCross.ID)
	require.NoError(t, err)
	require.Len(t, gotBoth.Pets, 1)
	require.NotNil(t, gotBoth.Pets[0].Insurance, "#86 認可済みなら B の保険は見えるべき")
	assert.Equal(t, insB.ID, gotBoth.Pets[0].Insurance.ID)

	// (iii) #86 [A] のみなら隠れる
	gotA, err := repo.FindByIDForClinics(ctx, []uint64{clinicA}, ownerCross.ID)
	require.NoError(t, err)
	require.Len(t, gotA.Pets, 1)
	assert.Nil(t, gotA.Pets[0].Insurance, "認可外 clinic の保険は #86 でも漏れてはならない")
}

// --- (c3) reservation: ReservationType / ReservationType.Group ---

func TestReservationRepository_ReservationType_CrossClinicPreloadIsolation(t *testing.T) {
	db := setupTestDB(t)
	db.Exec("TRUNCATE TABLE appointments CASCADE")
	db.Exec("TRUNCATE TABLE reservation_types CASCADE")
	db.Exec("TRUNCATE TABLE reservation_type_groups CASCADE")
	require.NoError(t, db.AutoMigrate(&model.ReservationTypeGroup{}, &model.ReservationType{}, &model.Reservation{}))
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rtA := makeReservationType(t, db, clinicA)
	groupB := &model.ReservationTypeGroup{ClinicID: clinicB, Name: "医院Bのグループ"}
	require.NoError(t, db.WithContext(ctx).Create(groupB).Error)
	rtB := &model.ReservationType{ClinicID: clinicB, Name: "医院Bの診療区分", Category: model.ReservationTypeCategoryGeneral, GroupID: &groupB.ID}
	require.NoError(t, db.WithContext(ctx).Create(rtB).Error)

	resLegit := makeReservationWithType(t, db, clinicA, rtA.ID)
	resCross := makeReservationWithType(t, db, clinicA, rtB.ID) // 別clinicの診療区分FKを植え付け

	// (legit) 正規の診療区分は Preload される
	gotLegit, err := repo.FindByID(ctx, clinicA, resLegit.ID)
	require.NoError(t, err)
	require.NotNil(t, gotLegit.ReservationType)
	assert.Equal(t, rtA.ID, gotLegit.ReservationType.ID)

	// (i) 別テナント診療区分は FindByID(単一)で混入しない
	gotCross, err := repo.FindByID(ctx, clinicA, resCross.ID)
	require.NoError(t, err)
	assert.Nil(t, gotCross.ReservationType, "別クリニックの診療区分が混入してはならない")

	// (ii) #86 [A,B] なら B の診療区分 + グループは見える
	gotBoth, err := repo.FindByIDForClinics(ctx, []uint64{clinicA, clinicB}, resCross.ID)
	require.NoError(t, err)
	require.NotNil(t, gotBoth.ReservationType, "#86 認可済みなら B の診療区分は見えるべき")
	assert.Equal(t, rtB.ID, gotBoth.ReservationType.ID)
	require.NotNil(t, gotBoth.ReservationType.Group, "#86 認可済みなら B のグループも見えるべき")
	assert.Equal(t, groupB.ID, gotBoth.ReservationType.Group.ID)

	// (iii) #86 [A] のみなら B の診療区分は隠れる
	gotA, err := repo.FindByIDForClinics(ctx, []uint64{clinicA}, resCross.ID)
	require.NoError(t, err)
	assert.Nil(t, gotA.ReservationType, "認可外 clinic の診療区分は #86 でも漏れてはならない")
}

func makeReservationWithType(t *testing.T, db *gorm.DB, clinicID, reservationTypeID uint64) *model.Reservation {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Minute)
	res := &model.Reservation{
		ClinicID:          clinicID,
		StartTime:         now,
		EndTime:           now.Add(15 * time.Minute),
		ReservationTypeID: reservationTypeID,
		VisitType:         model.VisitTypeRevisit,
		Status:            model.ReservationStatusPending,
		Source:            model.ReservationSourceManual,
		CustomerFields:    []byte(`{}`),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(res).Error)
	return res
}
