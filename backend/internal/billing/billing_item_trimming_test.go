package billing

// billing_item_trimming_test.go — G11-3 テスト負債解消
//
// テスト対象: BillingItemRepository.FindUnbilledTrimmingItemsByPetID
//   （70行 raw SQL、UNION ALL + NOT EXISTS）、CountNonAccountingTrimmingByPetAndDate
// 保護する不変条件:
//   - コース/オプションの請求済み除外は appointment_id + course_id(or option_id) の枝別対応で判定される
//   - cancelled のみの請求は「未請求」として再取得対象になる（b.status != cancelled）
//   - price=0/NULL は自動取り込み対象外
//   - 並び順は appointment_id → sort_order(course=0, option=100+ato.sort_order) → origin_id
//
// このテストは NOT EXISTS の結合条件（trimming_course_id/trimming_option_id の枝別対応）を
// 崩す、または price>0 条件を削除すると必ず失敗するよう設計されている。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupBillingItemTrimmingTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.AnimalSpecies{}, &model.Pet{},
		&model.ReservationType{}, &model.Reservation{},
		&model.TrimmingCourse{}, &model.TrimmingOption{},
		&model.AppointmentTrimmingDetail{}, &model.AppointmentTrimmingOption{},
		&model.BillingItem{},
	))
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	db.Exec("TRUNCATE TABLE reservation_types CASCADE")
	db.Exec("TRUNCATE TABLE trimming_courses CASCADE")
	db.Exec("TRUNCATE TABLE trimming_options CASCADE")
	return db
}

func priceOf(v int64) *int64 { return &v }

func makeTrimmingReservationType(t *testing.T, db *gorm.DB, clinicID uint64) *model.ReservationType {
	t.Helper()
	rt := &model.ReservationType{ClinicID: clinicID, Name: "トリミング", Category: model.ReservationTypeCategoryTrimming}
	require.NoError(t, db.WithContext(context.Background()).Create(rt).Error)
	return rt
}

func makeTrimmingCourse(t *testing.T, db *gorm.DB, clinicID uint64, name string, price *int64) *model.TrimmingCourse {
	t.Helper()
	c := &model.TrimmingCourse{ClinicID: clinicID, Name: name, Price: price}
	require.NoError(t, db.WithContext(context.Background()).Create(c).Error)
	return c
}

func makeTrimmingOption(t *testing.T, db *gorm.DB, clinicID uint64, name string, price *int64) *model.TrimmingOption {
	t.Helper()
	o := &model.TrimmingOption{ClinicID: clinicID, Name: name, Price: price}
	require.NoError(t, db.WithContext(context.Background()).Create(o).Error)
	return o
}

// makeTrimmingAppointment はトリミング区分の appointment を1件作成する。
func makeTrimmingAppointment(t *testing.T, db *gorm.DB, clinicID, petID, reservationTypeID uint64, status model.ReservationStatus) *model.Reservation {
	t.Helper()
	start := time.Date(2026, 6, 10, 3, 0, 0, 0, time.UTC)
	res := &model.Reservation{
		ClinicID:          clinicID,
		StartTime:         start,
		EndTime:           start.Add(60 * time.Minute),
		PetID:             &petID,
		ReservationTypeID: reservationTypeID,
		VisitType:         model.VisitTypeRevisit,
		Status:            status,
		Source:            model.ReservationSourceManual,
		CustomerFields:    []byte(`{}`),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(res).Error)
	return res
}

func attachTrimmingCourse(t *testing.T, db *gorm.DB, clinicID, appointmentID, courseID uint64) {
	t.Helper()
	d := &model.AppointmentTrimmingDetail{ClinicID: clinicID, AppointmentID: appointmentID, CourseID: &courseID}
	require.NoError(t, db.WithContext(context.Background()).Create(d).Error)
}

func attachTrimmingOption(t *testing.T, db *gorm.DB, appointmentID, optionID uint64, sortOrder int) {
	t.Helper()
	o := &model.AppointmentTrimmingOption{AppointmentID: appointmentID, OptionID: optionID, SortOrder: sortOrder}
	require.NoError(t, db.WithContext(context.Background()).Create(o).Error)
}

// makeTrimmingBilling は makeBillingWith（billing_test_fixtures_test.go）への thin wrapper。
func makeTrimmingBilling(t *testing.T, db *gorm.DB, clinicID uint64, status model.BillingStatus) *model.Billing {
	t.Helper()
	return makeBillingWith(t, db, billingFixtureOpts{
		ClinicID:      clinicID,
		Status:        status,
		ScheduledDate: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
	})
}

func makeTrimmingBillingItem(t *testing.T, db *gorm.DB, billingID, appointmentID uint64, courseID, optionID *uint64) *model.BillingItem {
	t.Helper()
	item := &model.BillingItem{
		BillingID:     billingID,
		Category:      model.ItemCategoryTrimming,
		AppointmentID: &appointmentID,
	}
	if courseID != nil {
		item.TrimmingCourseID = courseID
	}
	if optionID != nil {
		item.TrimmingOptionID = optionID
	}
	require.NoError(t, db.WithContext(context.Background()).Create(item).Error)
	return item
}

func TestBillingItemRepository_FindUnbilledTrimmingItemsByPetID(t *testing.T) {
	ctx := context.Background()
	const clinicA = uint64(1)
	const clinicB = uint64(2)

	t.Run("status=accountingのトリミング予約でコース+オプションが結合結果に出る", func(t *testing.T) {
		db := setupBillingItemTrimmingTestDB(t)
		repo := NewBillingItemRepository(db)
		owner := testdb.MakeTestOwner(t, db, clinicA, "G11-3飼主1")
		pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "G11-3ペット1")
		rt := makeTrimmingReservationType(t, db, clinicA)
		appt := makeTrimmingAppointment(t, db, clinicA, pet.ID, rt.ID, model.ReservationStatusAccounting)

		course := makeTrimmingCourse(t, db, clinicA, "カットコース", priceOf(3000))
		attachTrimmingCourse(t, db, clinicA, appt.ID, course.ID)
		opt := makeTrimmingOption(t, db, clinicA, "爪切り", priceOf(500))
		attachTrimmingOption(t, db, appt.ID, opt.ID, 0)

		items, err := repo.FindUnbilledTrimmingItemsByPetID(ctx, clinicA, pet.ID)
		require.NoError(t, err)
		require.Len(t, items, 2)

		var courseItem, optionItem *model.BillingItem
		for i := range items {
			if items[i].TrimmingCourseID != nil {
				courseItem = &items[i]
			}
			if items[i].TrimmingOptionID != nil {
				optionItem = &items[i]
			}
		}
		require.NotNil(t, courseItem, "コース枝の行が含まれる")
		require.NotNil(t, optionItem, "オプション枝の行が含まれる")
		assert.Equal(t, "カットコース", courseItem.Name)
		assert.Equal(t, int64(3000), courseItem.UnitPrice)
		assert.Equal(t, course.ID, *courseItem.TrimmingCourseID)
		assert.Nil(t, courseItem.TrimmingOptionID, "コース行はtrimming_option_idを持たない")
		assert.Equal(t, "爪切り", optionItem.Name)
		assert.Equal(t, int64(500), optionItem.UnitPrice)
		assert.Equal(t, opt.ID, *optionItem.TrimmingOptionID)
		assert.Nil(t, optionItem.TrimmingCourseID, "オプション行はtrimming_course_idを持たない")
	})

	t.Run("並び順はappointment_id→sort_order(コース0/オプション100+ato.sort_order)昇順", func(t *testing.T) {
		db := setupBillingItemTrimmingTestDB(t)
		repo := NewBillingItemRepository(db)
		owner := testdb.MakeTestOwner(t, db, clinicA, "G11-3飼主2")
		pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "G11-3ペット2")
		rt := makeTrimmingReservationType(t, db, clinicA)

		// appointment #1（先に作成 = 小さいID）: コース + 2オプション（join sort_order 逆順で挿入）
		appt1 := makeTrimmingAppointment(t, db, clinicA, pet.ID, rt.ID, model.ReservationStatusAccounting)
		course1 := makeTrimmingCourse(t, db, clinicA, "コース1", priceOf(3000))
		attachTrimmingCourse(t, db, clinicA, appt1.ID, course1.ID)
		optLate := makeTrimmingOption(t, db, clinicA, "後オプション", priceOf(300))
		attachTrimmingOption(t, db, appt1.ID, optLate.ID, 1)
		optEarly := makeTrimmingOption(t, db, clinicA, "先オプション", priceOf(200))
		attachTrimmingOption(t, db, appt1.ID, optEarly.ID, 0)

		// appointment #2（後で作成 = 大きいID）: コースのみ
		appt2 := makeTrimmingAppointment(t, db, clinicA, pet.ID, rt.ID, model.ReservationStatusAccounting)
		course2 := makeTrimmingCourse(t, db, clinicA, "コース2", priceOf(4000))
		attachTrimmingCourse(t, db, clinicA, appt2.ID, course2.ID)

		items, err := repo.FindUnbilledTrimmingItemsByPetID(ctx, clinicA, pet.ID)
		require.NoError(t, err)
		require.Len(t, items, 4)
		gotNames := make([]string, len(items))
		for i, it := range items {
			gotNames[i] = it.Name
		}
		assert.Equal(t, []string{"コース1", "先オプション", "後オプション", "コース2"}, gotNames,
			"appt1のコース(sort=0)→先オプション(sort=100)→後オプション(sort=101)→appt2のコース")
	})

	t.Run("請求済み除外: 有効なbillingに同一appointment_id+course_idのbilling_itemがあると除外される", func(t *testing.T) {
		db := setupBillingItemTrimmingTestDB(t)
		repo := NewBillingItemRepository(db)
		owner := testdb.MakeTestOwner(t, db, clinicA, "G11-3飼主3")
		pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "G11-3ペット3")
		rt := makeTrimmingReservationType(t, db, clinicA)
		appt := makeTrimmingAppointment(t, db, clinicA, pet.ID, rt.ID, model.ReservationStatusAccounting)
		course := makeTrimmingCourse(t, db, clinicA, "請求済みコース", priceOf(3000))
		attachTrimmingCourse(t, db, clinicA, appt.ID, course.ID)

		billing := makeTrimmingBilling(t, db, clinicA, model.BillingStatusWaiting)
		makeTrimmingBillingItem(t, db, billing.ID, appt.ID, &course.ID, nil)

		items, err := repo.FindUnbilledTrimmingItemsByPetID(ctx, clinicA, pet.ID)
		require.NoError(t, err)
		assert.Empty(t, items, "waitingのbillingに既に請求済みのコースは再取得されない")
	})

	t.Run("cancelled請求のみの場合は再取得対象になる", func(t *testing.T) {
		db := setupBillingItemTrimmingTestDB(t)
		repo := NewBillingItemRepository(db)
		owner := testdb.MakeTestOwner(t, db, clinicA, "G11-3飼主4")
		pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "G11-3ペット4")
		rt := makeTrimmingReservationType(t, db, clinicA)
		appt := makeTrimmingAppointment(t, db, clinicA, pet.ID, rt.ID, model.ReservationStatusAccounting)
		course := makeTrimmingCourse(t, db, clinicA, "キャンセル済みコース", priceOf(3000))
		attachTrimmingCourse(t, db, clinicA, appt.ID, course.ID)

		cancelledBilling := makeTrimmingBilling(t, db, clinicA, model.BillingStatusCancelled)
		makeTrimmingBillingItem(t, db, cancelledBilling.ID, appt.ID, &course.ID, nil)

		items, err := repo.FindUnbilledTrimmingItemsByPetID(ctx, clinicA, pet.ID)
		require.NoError(t, err)
		require.Len(t, items, 1, "cancelledのみの請求は未請求扱いで再取得される")
		assert.Equal(t, "キャンセル済みコース", items[0].Name)
	})

	t.Run("price=0/NULLのコース・オプションは除外", func(t *testing.T) {
		db := setupBillingItemTrimmingTestDB(t)
		repo := NewBillingItemRepository(db)
		owner := testdb.MakeTestOwner(t, db, clinicA, "G11-3飼主5")
		pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "G11-3ペット5")
		rt := makeTrimmingReservationType(t, db, clinicA)

		apptZero := makeTrimmingAppointment(t, db, clinicA, pet.ID, rt.ID, model.ReservationStatusAccounting)
		courseZero := makeTrimmingCourse(t, db, clinicA, "0円コース", priceOf(0))
		attachTrimmingCourse(t, db, clinicA, apptZero.ID, courseZero.ID)

		apptNil := makeTrimmingAppointment(t, db, clinicA, pet.ID, rt.ID, model.ReservationStatusAccounting)
		courseNil := makeTrimmingCourse(t, db, clinicA, "NULL価格コース", nil)
		attachTrimmingCourse(t, db, clinicA, apptNil.ID, courseNil.ID)

		apptOptZero := makeTrimmingAppointment(t, db, clinicA, pet.ID, rt.ID, model.ReservationStatusAccounting)
		optZero := makeTrimmingOption(t, db, clinicA, "0円オプション", priceOf(0))
		attachTrimmingOption(t, db, apptOptZero.ID, optZero.ID, 0)

		// 対照群: 有効な価格を持つコース
		apptValid := makeTrimmingAppointment(t, db, clinicA, pet.ID, rt.ID, model.ReservationStatusAccounting)
		courseValid := makeTrimmingCourse(t, db, clinicA, "有効コース", priceOf(1500))
		attachTrimmingCourse(t, db, clinicA, apptValid.ID, courseValid.ID)

		items, err := repo.FindUnbilledTrimmingItemsByPetID(ctx, clinicA, pet.ID)
		require.NoError(t, err)
		require.Len(t, items, 1, "price=0/NULLの3件は除外され有効コースのみ残る")
		assert.Equal(t, "有効コース", items[0].Name)
	})

	t.Run("クリニック/ペット/status/カテゴリ不一致は除外", func(t *testing.T) {
		db := setupBillingItemTrimmingTestDB(t)
		repo := NewBillingItemRepository(db)
		owner := testdb.MakeTestOwner(t, db, clinicA, "G11-3飼主6")
		pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "G11-3ペット6")
		otherPet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "別ペット")
		rtTrimming := makeTrimmingReservationType(t, db, clinicA)
		rtGeneral := makeReservationType(t, db, clinicA) // カテゴリ: general（一般区分）

		otherClinicOwner := testdb.MakeTestOwner(t, db, clinicB, "別クリニック飼主")
		otherClinicPet := makeSpeciesAndPet(t, db, clinicB, otherClinicOwner.ID, "別クリニックペット")
		rtOtherClinic := makeTrimmingReservationType(t, db, clinicB)

		// 別クリニックの appointment（同一クエリのclinicIDでは見えないはず）
		apptOtherClinic := makeTrimmingAppointment(t, db, clinicB, otherClinicPet.ID, rtOtherClinic.ID, model.ReservationStatusAccounting)
		courseOtherClinic := makeTrimmingCourse(t, db, clinicB, "別クリニックコース", priceOf(1000))
		attachTrimmingCourse(t, db, clinicB, apptOtherClinic.ID, courseOtherClinic.ID)

		// 別ペットの appointment
		apptOtherPet := makeTrimmingAppointment(t, db, clinicA, otherPet.ID, rtTrimming.ID, model.ReservationStatusAccounting)
		courseOtherPet := makeTrimmingCourse(t, db, clinicA, "別ペットコース", priceOf(1000))
		attachTrimmingCourse(t, db, clinicA, apptOtherPet.ID, courseOtherPet.ID)

		// status≠accounting
		apptPending := makeTrimmingAppointment(t, db, clinicA, pet.ID, rtTrimming.ID, model.ReservationStatusPending)
		coursePending := makeTrimmingCourse(t, db, clinicA, "未会計コース", priceOf(1000))
		attachTrimmingCourse(t, db, clinicA, apptPending.ID, coursePending.ID)

		// カテゴリ≠trimming
		apptGeneral := makeTrimmingAppointment(t, db, clinicA, pet.ID, rtGeneral.ID, model.ReservationStatusAccounting)
		courseGeneral := makeTrimmingCourse(t, db, clinicA, "一般区分コース", priceOf(1000))
		attachTrimmingCourse(t, db, clinicA, apptGeneral.ID, courseGeneral.ID)

		items, err := repo.FindUnbilledTrimmingItemsByPetID(ctx, clinicA, pet.ID)
		require.NoError(t, err)
		assert.Empty(t, items, "別クリニック/別ペット/status≠accounting/カテゴリ≠trimmingはすべて除外される")
	})

	t.Run("foreign master/detail corruption is excluded and foreign billing cannot suppress valid items", func(t *testing.T) {
		db := setupBillingItemTrimmingTestDB(t)
		repo := NewBillingItemRepository(db)
		owner := testdb.MakeTestOwner(t, db, clinicA, "tenant-corruption-owner")
		pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "tenant-corruption-pet")
		localType := makeTrimmingReservationType(t, db, clinicA)
		foreignType := makeTrimmingReservationType(t, db, clinicB)

		validAppointment := makeTrimmingAppointment(t, db, clinicA, pet.ID, localType.ID, model.ReservationStatusAccounting)
		validCourse := makeTrimmingCourse(t, db, clinicA, "valid local course", priceOf(3000))
		validOption := makeTrimmingOption(t, db, clinicA, "valid local option", priceOf(500))
		attachTrimmingCourse(t, db, clinicA, validAppointment.ID, validCourse.ID)
		attachTrimmingOption(t, db, validAppointment.ID, validOption.ID, 0)

		// A foreign-clinic billing_item may satisfy the raw FK constraints, but
		// it must not mark this clinic's valid appointment items as billed.
		foreignBilling := makeTrimmingBilling(t, db, clinicB, model.BillingStatusWaiting)
		makeTrimmingBillingItem(t, db, foreignBilling.ID, validAppointment.ID, &validCourse.ID, &validOption.ID)

		foreignTypeAppointment := makeTrimmingAppointment(t, db, clinicA, pet.ID, foreignType.ID, model.ReservationStatusAccounting)
		foreignTypeCourse := makeTrimmingCourse(t, db, clinicA, "foreign reservation type leak", priceOf(1000))
		attachTrimmingCourse(t, db, clinicA, foreignTypeAppointment.ID, foreignTypeCourse.ID)

		foreignCourseAppointment := makeTrimmingAppointment(t, db, clinicA, pet.ID, localType.ID, model.ReservationStatusAccounting)
		foreignCourse := makeTrimmingCourse(t, db, clinicB, "foreign course leak", priceOf(1100))
		attachTrimmingCourse(t, db, clinicA, foreignCourseAppointment.ID, foreignCourse.ID)

		foreignDetailAppointment := makeTrimmingAppointment(t, db, clinicA, pet.ID, localType.ID, model.ReservationStatusAccounting)
		foreignDetailCourse := makeTrimmingCourse(t, db, clinicA, "foreign detail leak", priceOf(1200))
		attachTrimmingCourse(t, db, clinicB, foreignDetailAppointment.ID, foreignDetailCourse.ID)

		foreignOptionAppointment := makeTrimmingAppointment(t, db, clinicA, pet.ID, localType.ID, model.ReservationStatusAccounting)
		require.NoError(t, db.Create(&model.AppointmentTrimmingDetail{
			ClinicID:      clinicA,
			AppointmentID: foreignOptionAppointment.ID,
		}).Error)
		foreignOption := makeTrimmingOption(t, db, clinicB, "foreign option leak", priceOf(600))
		attachTrimmingOption(t, db, foreignOptionAppointment.ID, foreignOption.ID, 0)

		items, err := repo.FindUnbilledTrimmingItemsByPetID(ctx, clinicA, pet.ID)

		require.NoError(t, err)
		require.Len(t, items, 2)
		assert.ElementsMatch(t, []string{"valid local course", "valid local option"}, []string{items[0].Name, items[1].Name})
	})
}

func TestBillingItemRepository_FindUnbilledTrimmingItemsByPetID_RejectsCorruptCrossClinicPetRelation(
	t *testing.T,
) {
	db := setupBillingItemTrimmingTestDB(t)
	repo := NewBillingItemRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerB := testdb.MakeTestOwner(t, db, clinicB, "cross-clinic trimming owner")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "cross-clinic trimming pet")
	reservationTypeA := makeTrimmingReservationType(t, db, clinicA)
	appointment := makeTrimmingAppointment(
		t,
		db,
		clinicA,
		petB.ID,
		reservationTypeA.ID,
		model.ReservationStatusAccounting,
	)
	course := makeTrimmingCourse(t, db, clinicA, "corrupt relation course", priceOf(3000))
	option := makeTrimmingOption(t, db, clinicA, "corrupt relation option", priceOf(500))
	attachTrimmingCourse(t, db, clinicA, appointment.ID, course.ID)
	attachTrimmingOption(t, db, appointment.ID, option.ID, 0)

	items, err := repo.FindUnbilledTrimmingItemsByPetID(ctx, clinicA, petB.ID)

	require.NoError(t, err)
	assert.Empty(t, items, "clinic A appointment must not resolve a clinic B pet")
}

func TestBillingItemRepository_CountNonAccountingTrimmingByPetAndDate(t *testing.T) {
	ctx := context.Background()
	const clinicA = uint64(1)
	targetDate := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)

	t.Run("JST日付境界と対象status(accounting/completed/cancelled以外)の判定", func(t *testing.T) {
		db := setupBillingItemTrimmingTestDB(t)
		repo := NewBillingItemRepository(db)
		owner := testdb.MakeTestOwner(t, db, clinicA, "G11-3カウント飼主")
		pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "G11-3カウントペット")
		rt := makeTrimmingReservationType(t, db, clinicA)

		// 対象: 同日(JST境界内)・status=pending(未会計候補)
		makeTrimmingAppointment(t, db, clinicA, pet.ID, rt.ID, model.ReservationStatusPending)
		// 対象: UTC前日15:00 = JST当日0:00
		apptBoundary := makeTrimmingAppointment(t, db, clinicA, pet.ID, rt.ID, model.ReservationStatusConfirmed)
		require.NoError(t, db.Model(&model.Reservation{}).Where("id = ?", apptBoundary.ID).
			Update("start_time", time.Date(2026, 6, 9, 15, 0, 0, 0, time.UTC)).Error)

		// 除外: UTC当日15:00 = JST翌日0:00（対象日外）
		apptNextDay := makeTrimmingAppointment(t, db, clinicA, pet.ID, rt.ID, model.ReservationStatusPending)
		require.NoError(t, db.Model(&model.Reservation{}).Where("id = ?", apptNextDay.ID).
			Update("start_time", time.Date(2026, 6, 10, 15, 0, 0, 0, time.UTC)).Error)

		// 除外: status=accounting（会計待ちに既に進んでいる=取り残しではない）
		makeTrimmingAppointment(t, db, clinicA, pet.ID, rt.ID, model.ReservationStatusAccounting)
		// 除外: ステータスが completed
		makeTrimmingAppointment(t, db, clinicA, pet.ID, rt.ID, model.ReservationStatusCompleted)
		// 除外: ステータスが cancelled
		makeTrimmingAppointment(t, db, clinicA, pet.ID, rt.ID, model.ReservationStatusCancelled)

		count, err := repo.CountNonAccountingTrimmingByPetAndDate(ctx, clinicA, pet.ID, targetDate)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count, "同日pending 1件 + JST境界内confirmed 1件 = 2件（accounting/completed/cancelled/翌日は除外）")
	})

	t.Run("foreign-clinic reservation type cannot classify a local appointment as trimming", func(t *testing.T) {
		db := setupBillingItemTrimmingTestDB(t)
		repo := NewBillingItemRepository(db)
		owner := testdb.MakeTestOwner(t, db, clinicA, "count-tenant-owner")
		pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "count-tenant-pet")
		localType := makeTrimmingReservationType(t, db, clinicA)
		foreignType := makeTrimmingReservationType(t, db, 2)

		makeTrimmingAppointment(t, db, clinicA, pet.ID, localType.ID, model.ReservationStatusPending)
		makeTrimmingAppointment(t, db, clinicA, pet.ID, foreignType.ID, model.ReservationStatusPending)

		count, err := repo.CountNonAccountingTrimmingByPetAndDate(ctx, clinicA, pet.ID, targetDate)

		require.NoError(t, err)
		assert.Equal(t, int64(1), count, "only the clinic-owned reservation type may classify the appointment")
	})
}

func TestBillingItemRepository_CountNonAccountingTrimmingByPetAndDate_RejectsCorruptCrossClinicPetRelation(
	t *testing.T,
) {
	db := setupBillingItemTrimmingTestDB(t)
	repo := NewBillingItemRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerB := testdb.MakeTestOwner(t, db, clinicB, "cross-clinic trimming count owner")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "cross-clinic trimming count pet")
	reservationTypeA := makeTrimmingReservationType(t, db, clinicA)
	makeTrimmingAppointment(
		t,
		db,
		clinicA,
		petB.ID,
		reservationTypeA.ID,
		model.ReservationStatusPending,
	)

	count, err := repo.CountNonAccountingTrimmingByPetAndDate(
		ctx,
		clinicA,
		petB.ID,
		time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
	)

	require.NoError(t, err)
	assert.Zero(t, count, "clinic A appointment must not count a clinic B pet")
}
