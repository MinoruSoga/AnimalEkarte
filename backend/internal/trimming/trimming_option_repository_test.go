package trimming

// trimming_option_repository_test.go — TrimmingOptionRepository の統合テスト（カバレッジ向上）。
//
// 対象: FindAll / FindByID / Create / Update / Delete / Reorder / CountUsageByTrimmingOptionID
// 検証観点: 正常系、clinic_id 隔離、ソフトデリート除外、NotFound ラップ。
//
// CountUsageByTrimmingOptionID は appointment_trimming_options を appointments に JOIN するため、
// setup で reservation_types / appointments (model.Reservation) も migrate する（既存の
// preload_followup_clinic_isolation_test.go / reservation_clinic_isolation_test.go と同じ構成）。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// setupTrimmingOptionTestDB は trimming_options / appointment_trimming_options / appointments 周りを整備する。
func setupTrimmingOptionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	// TRUNCATE first: 他テストが残した orphan 行を除去してから AutoMigrate（FK 検証を通すため）。
	db.Exec("TRUNCATE TABLE appointment_trimming_options CASCADE")
	db.Exec("TRUNCATE TABLE appointment_trimming_details CASCADE")
	db.Exec("TRUNCATE TABLE trimming_options CASCADE")
	db.Exec("TRUNCATE TABLE reservation_types CASCADE") // appointments も連鎖クリア
	require.NoError(t, ensureAutoMigrated(db,
		&model.ReservationType{}, &model.Reservation{},
		&model.TrimmingCourse{}, &model.TrimmingOption{}, &model.AppointmentTrimmingOption{},
	))
	// 防御的クリーンアップ: model.AppointmentTrimmingDetail.Options の many2many タグ
	// (gorm:"many2many:appointment_trimming_options;joinForeignKey:AppointmentID;...")
	// は AppointmentTrimmingDetail と AppointmentTrimmingOption を同時に AutoMigrate する
	// 別テスト（例: preload_followup_clinic_isolation_test.go）が走ると、GORM が
	// appointment_trimming_options.appointment_id → appointment_trimming_details.id という
	// 実スキーマ（001_init.sql は appointments への FK のみ）に存在しない誤った FK 制約を
	// 共有テストDBに作成してしまう。制約は TRUNCATE では消えず後続テストに残留するため、
	// 本テストの INSERT（appointment_id = 実 reservations.id）が誤って弾かれる。
	// 実スキーマに存在しない制約なので、テストDB上でのみ明示的に除去する。
	db.Exec(`ALTER TABLE appointment_trimming_options
		DROP CONSTRAINT IF EXISTS fk_appointment_trimming_options_appointment_trimming_detail`)
	return db
}

func TestTrimmingMasterFindByID_HoldsShareLockForAmbientTransaction(t *testing.T) {
	db := setupTrimmingOptionTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)
	course := &model.TrimmingCourse{ClinicID: clinicID, Name: "共有ロックコース", IsActive: true}
	option := &model.TrimmingOption{ClinicID: clinicID, Name: "共有ロックオプション", IsActive: true}
	require.NoError(t, db.Create(course).Error)
	require.NoError(t, db.Create(option).Error)

	tests := []struct {
		name   string
		read   func(context.Context) error
		update func() error
	}{
		{
			name: "course",
			read: func(txCtx context.Context) error {
				_, err := NewTrimmingCourseRepository(db).FindByID(txCtx, clinicID, course.ID)
				return err
			},
			update: func() error {
				return db.Model(&model.TrimmingCourse{}).
					Where("clinic_id = ? AND name = ?", clinicID, "共有ロックコース").
					Update("is_active", false).Error
			},
		},
		{
			name: "option",
			read: func(txCtx context.Context) error {
				_, err := NewTrimmingOptionRepository(db).FindByID(txCtx, clinicID, option.ID)
				return err
			},
			update: func() error {
				return db.Model(&model.TrimmingOption{}).
					Where("clinic_id = ? AND name = ?", clinicID, "共有ロックオプション").
					Update("is_active", false).Error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locked := make(chan struct{})
			release := make(chan struct{})
			readDone := make(chan error, 1)
			go func() {
				readDone <- newTestTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
					if err := tt.read(txCtx); err != nil {
						return err
					}
					close(locked)
					<-release
					return nil
				})
			}()
			<-locked

			updateStarted := make(chan struct{})
			updateDone := make(chan error, 1)
			go func() {
				close(updateStarted)
				updateDone <- tt.update()
			}()
			<-updateStarted

			select {
			case err := <-updateDone:
				close(release)
				require.Failf(t, "master update was not serialized", "err=%v", err)
			case <-time.After(100 * time.Millisecond):
				close(release)
			}
			require.NoError(t, <-readDone)
			require.NoError(t, <-updateDone)
		})
	}
}

func TestTrimmingOptionRepository_FindAll_ClinicIsolationAndSortOrder(t *testing.T) {
	db := setupTrimmingOptionTestDB(t)
	repo := NewTrimmingOptionRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	oB := &model.TrimmingOption{ClinicID: clinicB, Name: "医院Bのオプション", SortOrder: 1}
	require.NoError(t, db.WithContext(ctx).Create(oB).Error)
	second := &model.TrimmingOption{ClinicID: clinicA, Name: "爪切り", SortOrder: 2}
	require.NoError(t, db.WithContext(ctx).Create(second).Error)
	first := &model.TrimmingOption{ClinicID: clinicA, Name: "耳掃除", SortOrder: 1}
	require.NoError(t, db.WithContext(ctx).Create(first).Error)

	got, err := repo.FindAll(ctx, clinicA)
	require.NoError(t, err)
	require.Len(t, got, 2, "clinic A のオプションのみ返る")
	assert.Equal(t, first.ID, got[0].ID, "sort_order 昇順で先頭に来る")
	assert.Equal(t, second.ID, got[1].ID)
	for _, x := range got {
		assert.NotEqual(t, oB.ID, x.ID, "別クリニックのオプションが混入してはならない")
	}
}

func TestTrimmingOptionRepository_FindByID(t *testing.T) {
	db := setupTrimmingOptionTestDB(t)
	repo := NewTrimmingOptionRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	price := int64(800)
	o := &model.TrimmingOption{ClinicID: clinicA, Name: "肛門腺絞り", Price: &price}
	require.NoError(t, db.WithContext(ctx).Create(o).Error)

	t.Run("同一クリニックで取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, o.ID)
		require.NoError(t, err)
		assert.Equal(t, "肛門腺絞り", got.Name)
		require.NotNil(t, got.Price)
		assert.Equal(t, price, *got.Price)
	})

	t.Run("別クリニックからは NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicB, o.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("存在しない ID は NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestTrimmingOptionRepository_Create(t *testing.T) {
	db := setupTrimmingOptionTestDB(t)
	repo := NewTrimmingOptionRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	o := &model.TrimmingOption{ClinicID: clinicA, Name: "新規オプション"}
	require.NoError(t, repo.Create(ctx, o))
	assert.NotZero(t, o.ID)

	got, err := repo.FindByID(ctx, clinicA, o.ID)
	require.NoError(t, err)
	assert.Equal(t, "新規オプション", got.Name)
}

// BUG-455-S8: gorm default:true omits zero bools from INSERT.
func TestTrimmingOptionRepository_Create_BoolDefaultsFalsePersist(t *testing.T) {
	db := setupTrimmingOptionTestDB(t)
	repo := NewTrimmingOptionRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	o := &model.TrimmingOption{
		ClinicID:     clinicA,
		Name:         "inactive noncombinable option",
		IsActive:     false,
		IsCombinable: false,
	}
	require.NoError(t, repo.Create(ctx, o))
	assert.False(t, o.IsActive)
	assert.False(t, o.IsCombinable)

	got, err := repo.FindByID(ctx, clinicA, o.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive)
	assert.False(t, got.IsCombinable)

	var rawActive, rawCombinable bool
	require.NoError(t, db.WithContext(ctx).
		Model(&model.TrimmingOption{}).
		Select("is_active").
		Where("id = ?", o.ID).
		Scan(&rawActive).Error)
	require.NoError(t, db.WithContext(ctx).
		Model(&model.TrimmingOption{}).
		Select("is_combinable").
		Where("id = ?", o.ID).
		Scan(&rawCombinable).Error)
	assert.False(t, rawActive, "raw is_active must be false")
	assert.False(t, rawCombinable, "raw is_combinable must be false")
}

func TestTrimmingOptionRepository_Update(t *testing.T) {
	db := setupTrimmingOptionTestDB(t)
	repo := NewTrimmingOptionRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	o := &model.TrimmingOption{ClinicID: clinicA, Name: "旧名称"}
	require.NoError(t, db.WithContext(ctx).Create(o).Error)

	t.Run("同一クリニックで更新できる", func(t *testing.T) {
		name := "新名称"
		got, err := repo.Update(ctx, clinicA, o.ID, UpdateTrimmingOptionInput{Name: &name})
		require.NoError(t, err)
		assert.Equal(t, "新名称", got.Name)
	})

	t.Run("別クリニックからの更新は NotFound", func(t *testing.T) {
		name := "乗っ取り"
		_, err := repo.Update(ctx, clinicB, o.ID, UpdateTrimmingOptionInput{Name: &name})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の更新は NotFound", func(t *testing.T) {
		name := "x"
		_, err := repo.Update(ctx, clinicA, 999999, UpdateTrimmingOptionInput{Name: &name})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestTrimmingOptionRepository_Delete(t *testing.T) {
	db := setupTrimmingOptionTestDB(t)
	repo := NewTrimmingOptionRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	o := &model.TrimmingOption{ClinicID: clinicA, Name: "削除対象"}
	require.NoError(t, db.WithContext(ctx).Create(o).Error)

	t.Run("別クリニックからの削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, o.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("同一クリニックで削除でき、ソフトデリートされる", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, o.ID))

		_, err := repo.FindByID(ctx, clinicA, o.ID)
		assert.True(t, apperrors.IsNotFound(err))

		all, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		for _, x := range all {
			assert.NotEqual(t, o.ID, x.ID)
		}

		var raw model.TrimmingOption
		require.NoError(t, db.WithContext(ctx).Unscoped().Where("id = ?", o.ID).First(&raw).Error)
		assert.True(t, raw.DeletedAt.Valid, "deleted_at が設定されているべき（物理行は残る）")
	})
}

func TestTrimmingOptionRepository_CountUsageByTrimmingOptionID(t *testing.T) {
	db := setupTrimmingOptionTestDB(t)
	repo := NewTrimmingOptionRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	option := &model.TrimmingOption{ClinicID: clinicA, Name: "使用中のオプション"}
	require.NoError(t, db.WithContext(ctx).Create(option).Error)

	t.Run("使用実績が無ければ 0", func(t *testing.T) {
		count, err := repo.CountUsageByTrimmingOptionID(ctx, clinicA, option.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	// appointment_trimming_options.appointment_id は appointments(=Reservation) への FK のため実予約を作る。
	appt := makeReservation(t, db, clinicA)
	link := &model.AppointmentTrimmingOption{AppointmentID: appt.ID, OptionID: option.ID}
	require.NoError(t, db.WithContext(ctx).Create(link).Error)

	t.Run("生存する appointment が使用していれば 1", func(t *testing.T) {
		count, err := repo.CountUsageByTrimmingOptionID(ctx, clinicA, option.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("別クリニックIDでは 0（クロステナント越境なし）", func(t *testing.T) {
		count, err := repo.CountUsageByTrimmingOptionID(ctx, clinicB, option.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("ソフトデリートされた appointment は除外される（P2）", func(t *testing.T) {
		require.NoError(t, db.WithContext(ctx).Delete(appt).Error)
		count, err := repo.CountUsageByTrimmingOptionID(ctx, clinicA, option.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestTrimmingOptionRepository_Reorder(t *testing.T) {
	db := setupTrimmingOptionTestDB(t)
	repo := NewTrimmingOptionRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	o1 := &model.TrimmingOption{ClinicID: clinicA, Name: "O1"}
	require.NoError(t, db.WithContext(ctx).Create(o1).Error)
	o2 := &model.TrimmingOption{ClinicID: clinicA, Name: "O2"}
	require.NoError(t, db.WithContext(ctx).Create(o2).Error)
	o3 := &model.TrimmingOption{ClinicID: clinicA, Name: "O3"}
	require.NoError(t, db.WithContext(ctx).Create(o3).Error)

	t.Run("指定順に sort_order が振り直される", func(t *testing.T) {
		require.NoError(t, repo.Reorder(ctx, clinicA, []uint64{o3.ID, o1.ID, o2.ID}))

		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, 3)
		assert.Equal(t, o3.ID, got[0].ID)
		assert.Equal(t, o1.ID, got[1].ID)
		assert.Equal(t, o2.ID, got[2].ID)
	})

	t.Run("別クリニックの ID を含むと失敗する", func(t *testing.T) {
		other := &model.TrimmingOption{ClinicID: clinicB, Name: "他院"}
		require.NoError(t, db.WithContext(ctx).Create(other).Error)
		err := repo.Reorder(ctx, clinicA, []uint64{o1.ID, other.ID})
		require.Error(t, err)
	})
}
