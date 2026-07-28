package trimming

// trimming_repository_test.go — AppointmentTrimmingDetailRepository の統合テスト（カバレッジ向上）。
//
// 対象: FindByAppointmentID / Create / Update / SetOptions
// 検証観点: 正常系（Course/Options Preload含む、CourseID nil、Options空、Options非空)、clinic_id 隔離、
//           NotFound ラップ、FK違反(23503: appointment_id/course_id)/UNIQUE違反(23505)の apperrors
//           マッピング、SetOptions の Clear + Replace（全置換／空スライスでの解除）。
//
// FindByAppointmentID のマスタ Preload clinic 隔離は
// preload_followup_clinic_isolation_test.go の
// TestAppointmentTrimmingDetailRepository_FindByAppointmentID_MasterPreloadClinicIsolation で
// 別途検証済みのため、本ファイルでは同一クリニックの正常系と NotFound 系を中心に補完する。
//
// #212 修正済み: Options many2many タグに foreignKey:AppointmentID / references:ID を追加したことで
// （backend/internal/model/trimming.go）、GORM の Association()/Preload がソース側結合キーとして
// AppointmentID を使うようになり、旧来 SetOptions が出していた "primary key required" 失敗と
// Preload の常時空返却（appointment_trimming_options.appointment_id を
// appointment_trimming_details.id と誤って突合していた）が解消された（2026-07-13）。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// setupAppointmentTrimmingDetailTestDB は appointment_trimming_details / appointment_trimming_options /
// trimming_courses / trimming_options / appointments 周りを整備する。
//
// setupIsolatedTestDB を使う: 共有プール上の並行 TRUNCATE（reservation_types CASCADE 等）が
// 他テストの接続状態を破壊し TestAppointmentTrimmingDetail* をフレークさせる実測があるため
// （TEST-FLAKE-P2 / #236）。TRUNCATE のみでも並行破壊するため隔離する。
func setupAppointmentTrimmingDetailTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupIsolatedTestDB(t)
	// TRUNCATE first: 他テストが残した orphan 行を除去してから AutoMigrate（FK 検証を通すため）。
	db.Exec("TRUNCATE TABLE appointment_trimming_options CASCADE")
	db.Exec("TRUNCATE TABLE appointment_trimming_details CASCADE")
	db.Exec("TRUNCATE TABLE trimming_courses CASCADE")
	db.Exec("TRUNCATE TABLE trimming_options CASCADE")
	db.Exec("TRUNCATE TABLE reservation_types CASCADE") // appointments も連鎖クリア
	require.NoError(t, ensureAutoMigrated(db,
		&model.ReservationType{}, &model.Reservation{},
		&model.TrimmingCourse{}, &model.TrimmingOption{},
		&model.AppointmentTrimmingDetail{}, &model.AppointmentTrimmingOption{},
	))
	return db
}

func TestAppointmentTrimmingDetailRepository_FindByAppointmentID_Success(t *testing.T) {
	db := setupAppointmentTrimmingDetailTestDB(t)
	repo := NewAppointmentTrimmingDetailRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	appt := makeReservation(t, db, clinicA)
	course := &model.TrimmingCourse{ClinicID: clinicA, Name: "全身コース"}
	require.NoError(t, db.WithContext(ctx).Create(course).Error)
	detail := &model.AppointmentTrimmingDetail{
		ClinicID: clinicA, AppointmentID: appt.ID, CourseID: &course.ID, StyleRequest: "ふわふわに",
	}
	require.NoError(t, db.WithContext(ctx).Create(detail).Error)

	t.Run("同一クリニックで取得しCourseがPreloadされる", func(t *testing.T) {
		got, err := repo.FindByAppointmentID(ctx, clinicA, appt.ID)
		require.NoError(t, err)
		require.NotNil(t, got.Course, "同一クリニックのコースは Preload されるべき")
		assert.Equal(t, course.ID, got.Course.ID)
		assert.Equal(t, "ふわふわに", got.StyleRequest)
	})

	t.Run("別クリニックからは NotFound", func(t *testing.T) {
		_, err := repo.FindByAppointmentID(ctx, clinicB, appt.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない appointment_id は NotFound", func(t *testing.T) {
		_, err := repo.FindByAppointmentID(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("Optionsが未設定の場合は空で返る", func(t *testing.T) {
		got, err := repo.FindByAppointmentID(ctx, clinicA, appt.ID)
		require.NoError(t, err)
		assert.Empty(t, got.Options, "オプションを一度も設定していない appointment は Options が空で返る")
	})

	t.Run("CourseIDがnilの場合はCourseもnilで返る", func(t *testing.T) {
		apptNoCourse := makeReservation(t, db, clinicA)
		detailNoCourse := &model.AppointmentTrimmingDetail{
			ClinicID: clinicA, AppointmentID: apptNoCourse.ID, StyleRequest: "コース未指定",
		}
		require.NoError(t, db.WithContext(ctx).Create(detailNoCourse).Error)

		got, err := repo.FindByAppointmentID(ctx, clinicA, apptNoCourse.ID)
		require.NoError(t, err)
		assert.Nil(t, got.CourseID)
		assert.Nil(t, got.Course, "CourseID が nil の場合 Course も Preload されず nil のまま")
	})

	// SetOptions は #212 修正前は既知のプロダクションバグで使用できなかったため（本ファイル冒頭コメント参照）、
	// 中間テーブル (appointment_trimming_options) へは直接 INSERT して Options Preload の非空分岐を検証する。
	// これは SetOptions の回避策ではなく FindByAppointmentID 側の Preload 挙動そのものを対象にしたテストであり、
	// preload_followup_clinic_isolation_test.go の
	// TestAppointmentTrimmingDetailRepository_FindByAppointmentID_MasterPreloadClinicIsolation
	// （L93）と同じ直接 INSERT 手法を用いる。
	// #212 修正済み（2026-07-13）: Options many2many タグに foreignKey:AppointmentID を追加したことで
	// 本サブテストは green になった。修正前は GORM Association()/Preload が joinForeignKey:AppointmentID を
	// 無視し model の primary key(detail.ID) で解決していたため、db.Debug() で確認した生成 SQL は
	// `WHERE appointment_trimming_options.appointment_id = <detail.ID>` であり、実予約ID
	// (apptWithOptions.ID = detail.AppointmentID) ではなく detail の主キー値でフィルタしていた
	// （エラーなしで常に空配列を返す静かな失敗）。
	// 既存の TestAppointmentTrimmingDetailRepository_FindByAppointmentID_MasterPreloadClinicIsolation
	// (preload_followup_clinic_isolation_test.go) はこのバグを検出できていなかった —
	// 別クリニックの Options が「混入しない」ことだけを assert.Empty で検証しており、同一クリニックの
	// Options が実際に返ることを検証する正常系ケースがこれまで存在しなかったため。
	t.Run("同一クリニックのOptionsはPreloadされる", func(t *testing.T) {
		apptWithOptions := makeReservation(t, db, clinicA)
		detailWithOptions := &model.AppointmentTrimmingDetail{
			ClinicID: clinicA, AppointmentID: apptWithOptions.ID, StyleRequest: "オプション付き",
		}
		require.NoError(t, db.WithContext(ctx).Create(detailWithOptions).Error)

		optA1 := &model.TrimmingOption{ClinicID: clinicA, Name: "耳掃除_Preload検証"}
		require.NoError(t, db.WithContext(ctx).Create(optA1).Error)
		optA2 := &model.TrimmingOption{ClinicID: clinicA, Name: "爪切り_Preload検証"}
		require.NoError(t, db.WithContext(ctx).Create(optA2).Error)
		require.NoError(t, db.WithContext(ctx).Create(&model.AppointmentTrimmingOption{
			AppointmentID: apptWithOptions.ID, OptionID: optA1.ID,
		}).Error)
		require.NoError(t, db.WithContext(ctx).Create(&model.AppointmentTrimmingOption{
			AppointmentID: apptWithOptions.ID, OptionID: optA2.ID,
		}).Error)

		got, err := repo.FindByAppointmentID(ctx, clinicA, apptWithOptions.ID)
		require.NoError(t, err)
		require.Len(t, got.Options, 2, "同一クリニックのオプションは Preload されるべき")
		ids := map[uint64]bool{}
		for _, o := range got.Options {
			ids[o.ID] = true
		}
		assert.True(t, ids[optA1.ID])
		assert.True(t, ids[optA2.ID])
	})
}

func TestAppointmentTrimmingDetailRepository_Create(t *testing.T) {
	db := setupAppointmentTrimmingDetailTestDB(t)
	repo := NewAppointmentTrimmingDetailRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	t.Run("正常系", func(t *testing.T) {
		appt := makeReservation(t, db, clinicA)
		detail := &model.AppointmentTrimmingDetail{
			ClinicID: clinicA, AppointmentID: appt.ID, StyleRequest: "テディベアカット",
		}
		require.NoError(t, repo.Create(ctx, detail))
		assert.NotZero(t, detail.ID)

		got, err := repo.FindByAppointmentID(ctx, clinicA, appt.ID)
		require.NoError(t, err)
		assert.Equal(t, "テディベアカット", got.StyleRequest)
	})

	t.Run("存在しないappointment_idはFK違反でInvalidInput", func(t *testing.T) {
		detail := &model.AppointmentTrimmingDetail{
			ClinicID: clinicA, AppointmentID: 999999, StyleRequest: "存在しない予約",
		}
		err := repo.Create(ctx, detail)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "appointment_id の外部キー違反は InvalidInput にマップされるべき（23503）")
	})

	t.Run("appointment_idの重複はAlreadyExists", func(t *testing.T) {
		appt := makeReservation(t, db, clinicA)
		first := &model.AppointmentTrimmingDetail{ClinicID: clinicA, AppointmentID: appt.ID, StyleRequest: "初回"}
		require.NoError(t, repo.Create(ctx, first))

		dup := &model.AppointmentTrimmingDetail{ClinicID: clinicA, AppointmentID: appt.ID, StyleRequest: "重複"}
		err := repo.Create(ctx, dup)
		require.Error(t, err)
		assert.True(t, apperrors.IsAlreadyExists(err), "appointment_id UNIQUE 制約違反は AlreadyExists にマップされるべき（23505）")
	})

	t.Run("存在しないcourse_idはFK違反でInvalidInput", func(t *testing.T) {
		// course_id は TrimmingCourse.gorm:"foreignKey:CourseID" の宣言済みリレーション経由で
		// AutoMigrate が実 FK 制約(REFERENCES trimming_courses(id))を作成する（001_init.sqlとも一致）。
		appt := makeReservation(t, db, clinicA)
		bogusCourseID := uint64(999999)
		detail := &model.AppointmentTrimmingDetail{
			ClinicID: clinicA, AppointmentID: appt.ID, CourseID: &bogusCourseID, StyleRequest: "存在しないコース",
		}
		err := repo.Create(ctx, detail)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "course_id の外部キー違反は InvalidInput にマップされるべき（23503）")
	})
}

func TestAppointmentTrimmingDetailRepository_Update(t *testing.T) {
	db := setupAppointmentTrimmingDetailTestDB(t)
	repo := NewAppointmentTrimmingDetailRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	appt := makeReservation(t, db, clinicA)
	detail := &model.AppointmentTrimmingDetail{
		ClinicID: clinicA, AppointmentID: appt.ID, StyleRequest: "旧リクエスト",
	}
	require.NoError(t, repo.Create(ctx, detail))

	t.Run("同一クリニックで更新できる（map による明示的な全フィールド更新）", func(t *testing.T) {
		updated := &model.AppointmentTrimmingDetail{
			ClinicID:      clinicA,
			AppointmentID: appt.ID,
			StyleRequest:  "新リクエスト",
			BWUnit:        model.BodyWeightUnitKg,
			UsedShampoo:   "オーガニックシャンプー",
			UsedRibbon:    "赤リボン",
			Remarks:       "毛玉なし",
		}
		require.NoError(t, repo.Update(ctx, updated))

		got, err := repo.FindByAppointmentID(ctx, clinicA, appt.ID)
		require.NoError(t, err)
		assert.Equal(t, "新リクエスト", got.StyleRequest)
		assert.Equal(t, "オーガニックシャンプー", got.UsedShampoo)
		assert.Equal(t, "赤リボン", got.UsedRibbon)
		assert.Equal(t, "毛玉なし", got.Remarks)
	})

	t.Run("別クリニックからの更新は NotFound", func(t *testing.T) {
		// bw_unit は Postgres ENUM(body_weight_unit) 型のため、Update() が常に全フィールドを
		// 上書きする実装（fields map に無条件で bw_unit を含む）である以上、空文字は不正値エラーに
		// なりゼロ件更新の NotFound と紛れる。有効な値を明示して純粋に clinic_id 隔離のみを検証する。
		updated := &model.AppointmentTrimmingDetail{ClinicID: clinicB, AppointmentID: appt.ID, StyleRequest: "乗っ取り", BWUnit: model.BodyWeightUnitKg}
		err := repo.Update(ctx, updated)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない appointment_id の更新は NotFound", func(t *testing.T) {
		updated := &model.AppointmentTrimmingDetail{ClinicID: clinicA, AppointmentID: 999999, StyleRequest: "x", BWUnit: model.BodyWeightUnitKg}
		err := repo.Update(ctx, updated)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("不正なbw_unitはInvalidInput", func(t *testing.T) {
		// bw_unit は Postgres ENUM(body_weight_unit) 型で、Update() は常に string(detail.BWUnit) を
		// 無条件で書き込む。ゼロ値の BWUnit ("") は該当 ENUM のいずれの値でもないため
		// invalid_text_representation (22P02) となり、result.Error 分岐（RowsAffected==0 の
		// NotFound とは別経路）を経由して FromGORM → InvalidInput にマップされる。
		updated := &model.AppointmentTrimmingDetail{ClinicID: clinicA, AppointmentID: appt.ID, StyleRequest: "不正リクエスト"}
		err := repo.Update(ctx, updated)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "bw_unit の ENUM 型違反は InvalidInput にマップされるべき（22P02）")
	})

	t.Run("存在しないcourse_idへの更新はFK違反でInvalidInput", func(t *testing.T) {
		// bw_unit は Update() が無条件で上書きするため、22P02（不正な bw_unit）と 23503（course_id FK違反）の
		// 両方とも FromGORM 経由で InvalidInput にマップされ変数が交絡する。course_id 違反のみを分離検証するため
		// 有効な BWUnit を明示し、appointment_id は既存の同一クリニックレコードを使う。
		bogusCourseID := uint64(999999)
		updated := &model.AppointmentTrimmingDetail{
			ClinicID: clinicA, AppointmentID: appt.ID, CourseID: &bogusCourseID,
			StyleRequest: "存在しないコースへ更新", BWUnit: model.BodyWeightUnitKg,
		}
		err := repo.Update(ctx, updated)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "course_id の外部キー違反は InvalidInput にマップされるべき（23503）")
	})
}

func TestAppointmentTrimmingDetailRepository_SetOptions(t *testing.T) {
	db := setupAppointmentTrimmingDetailTestDB(t)
	repo := NewAppointmentTrimmingDetailRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	appt := makeReservation(t, db, clinicA)
	detail := &model.AppointmentTrimmingDetail{ClinicID: clinicA, AppointmentID: appt.ID}
	require.NoError(t, repo.Create(ctx, detail))

	opt1 := &model.TrimmingOption{ClinicID: clinicA, Name: "耳掃除"}
	require.NoError(t, db.WithContext(ctx).Create(opt1).Error)
	opt2 := &model.TrimmingOption{ClinicID: clinicA, Name: "爪切り"}
	require.NoError(t, db.WithContext(ctx).Create(opt2).Error)
	opt3 := &model.TrimmingOption{ClinicID: clinicA, Name: "肛門腺絞り"}
	require.NoError(t, db.WithContext(ctx).Create(opt3).Error)

	// #212 修正済み（2026-07-13）: 旧タグでは `detail := &model.AppointmentTrimmingDetail{AppointmentID: appointmentID}`
	// （detail.ID は未設定）を `tx.Model(detail).Association("Options")...` に渡すと、Options の
	// many2many タグに foreignKey が未指定のため GORM が primaryKey(ID) をソース側結合キーとして解決し、
	// ゼロ値チェックで "primary key required" エラーとなり SetOptions は常に失敗していた。
	// model/trimming.go に foreignKey:AppointmentID を追加したことで解決した。
	t.Run("複数オプションを設定できる", func(t *testing.T) {
		require.NoError(t, repo.SetOptions(ctx, clinicA, appt.ID, []uint64{opt1.ID, opt2.ID}))

		got, err := repo.FindByAppointmentID(ctx, clinicA, appt.ID)
		require.NoError(t, err)
		require.Len(t, got.Options, 2)
		ids := map[uint64]bool{}
		for _, o := range got.Options {
			ids[o.ID] = true
		}
		assert.True(t, ids[opt1.ID])
		assert.True(t, ids[opt2.ID])
	})

	t.Run("再設定で Clear + Replace され全置換される", func(t *testing.T) {
		require.NoError(t, repo.SetOptions(ctx, clinicA, appt.ID, []uint64{opt3.ID}))

		got, err := repo.FindByAppointmentID(ctx, clinicA, appt.ID)
		require.NoError(t, err)
		require.Len(t, got.Options, 1, "旧オプション(opt1,opt2)は解除され opt3 のみ残る")
		assert.Equal(t, opt3.ID, got.Options[0].ID)
	})

	t.Run("空スライスで全解除される", func(t *testing.T) {
		require.NoError(t, repo.SetOptions(ctx, clinicA, appt.ID, []uint64{}))

		got, err := repo.FindByAppointmentID(ctx, clinicA, appt.ID)
		require.NoError(t, err)
		assert.Empty(t, got.Options)
	})
}

// TestAppointmentTrimmingDetailRepository_FindByAppointmentID_OptionsClinicScopeMixed は
// 同一 appointment に同一クリニックのオプションと別クリニックのオプションが両方リンクされている
// ケースを検証する（#212 修正の検証強化）。
//
// preload_followup_clinic_isolation_test.go の
// TestAppointmentTrimmingDetailRepository_FindByAppointmentID_MasterPreloadClinicIsolation は
// 別クリニックのオプションのみをリンクし「常に空」を assert.Empty するため、#212 のバグ
// （Preload が常に空を返す）があっても偶然パスしてしまい、clinic_id 述語が実際に機能していることを
// 証明できていなかった。本テストは同一 appointment に両クリニックのオプションを混在させ、
// 同一クリニック分のみが返ることを検証することで、Preload の clinic_id スコープが実際に効いている
// ことを fix-sensitive に証明する。
func TestAppointmentTrimmingDetailRepository_FindByAppointmentID_OptionsClinicScopeMixed(t *testing.T) {
	db := setupAppointmentTrimmingDetailTestDB(t)
	repo := NewAppointmentTrimmingDetailRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	appt := makeReservation(t, db, clinicA)
	detail := &model.AppointmentTrimmingDetail{ClinicID: clinicA, AppointmentID: appt.ID}
	require.NoError(t, repo.Create(ctx, detail))

	optA := &model.TrimmingOption{ClinicID: clinicA, Name: "医院Aのオプション"}
	require.NoError(t, db.WithContext(ctx).Create(optA).Error)
	optB := &model.TrimmingOption{ClinicID: clinicB, Name: "医院Bのオプション"}
	require.NoError(t, db.WithContext(ctx).Create(optB).Error)

	// 中間テーブルへは直接 INSERT する（SetOptions は clinic_id を検証しないサービス層の責務のため、
	// join テーブル自体が clinic をまたぐ行を持ちうる状態を意図的に作る）。
	require.NoError(t, db.WithContext(ctx).Create(&model.AppointmentTrimmingOption{
		AppointmentID: appt.ID, OptionID: optA.ID,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.AppointmentTrimmingOption{
		AppointmentID: appt.ID, OptionID: optB.ID,
	}).Error)

	got, err := repo.FindByAppointmentID(ctx, clinicA, appt.ID)
	require.NoError(t, err)
	require.Len(t, got.Options, 1, "同一クリニックのオプションのみ Preload されるべき")
	assert.Equal(t, optA.ID, got.Options[0].ID)
}

// TestAppointmentTrimmingDetailRepository_FindByAppointmentID_RetainsInactiveLinkedOption は
// #228（無効化された既存リンク済みオプションをカルテから消さない）の repository 層の前提条件を
// 実DBで検証する。trimming_service.go の validateTrimmingCourseAndOptions は detail.Options から
// existingOptionIDs を構築して既存リンクを免除するため、リンク済みオプションが後から
// is_active=false になっても Preload("Options") が引き続きそれを返すことが必須である
// （is_active でフィルタしていないことの回帰防止）。
func TestAppointmentTrimmingDetailRepository_FindByAppointmentID_RetainsInactiveLinkedOption(t *testing.T) {
	db := setupAppointmentTrimmingDetailTestDB(t)
	repo := NewAppointmentTrimmingDetailRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	appt := makeReservation(t, db, clinicA)
	detail := &model.AppointmentTrimmingDetail{ClinicID: clinicA, AppointmentID: appt.ID}
	require.NoError(t, repo.Create(ctx, detail))

	opt := &model.TrimmingOption{ClinicID: clinicA, Name: "後で無効化されるオプション", IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(opt).Error)
	require.NoError(t, repo.SetOptions(ctx, clinicA, appt.ID, []uint64{opt.ID}))

	// マスタ側で後から無効化（#228 が想定するシナリオ）。
	require.NoError(t, db.WithContext(ctx).Model(&model.TrimmingOption{}).
		Where("id = ?", opt.ID).Update("is_active", false).Error)

	got, err := repo.FindByAppointmentID(ctx, clinicA, appt.ID)
	require.NoError(t, err)
	require.Len(t, got.Options, 1, "無効化後も既存リンクは Preload で維持されるべき（service層の免除ロジックの前提）")
	assert.Equal(t, opt.ID, got.Options[0].ID)
	assert.False(t, got.Options[0].IsActive, "マスタ側の is_active=false が反映されていること")
}

// TestAppointmentTrimmingDetailRepository_SetOptions_ClinicIsolation は SetOptions の
// clinicID 引数（defense-in-depth、clinic-isolation-auditor 指摘）が実際に別クリニックの
// appointment_id への書き込みを拒否することを検証する。
// SetOptions の呼び出し元（trimming_service.go / liff_service_reservations.go）は
// appointmentID を事前に clinic 検証済みだが、repository 層でも fail-closed に再検証する
// （checkup_field_repository.go の ReplaceForCheckup と同型パターン）。
func TestAppointmentTrimmingDetailRepository_SetOptions_ClinicIsolation(t *testing.T) {
	db := setupAppointmentTrimmingDetailTestDB(t)
	repo := NewAppointmentTrimmingDetailRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	appt := makeReservation(t, db, clinicA)
	detail := &model.AppointmentTrimmingDetail{ClinicID: clinicA, AppointmentID: appt.ID}
	require.NoError(t, repo.Create(ctx, detail))

	opt := &model.TrimmingOption{ClinicID: clinicA, Name: "医院Aのオプション"}
	require.NoError(t, db.WithContext(ctx).Create(opt).Error)

	err := repo.SetOptions(ctx, clinicB, appt.ID, []uint64{opt.ID})
	require.Error(t, err, "別クリニックIDでは appointment_id が一致しても書き込めるべきではない")
	assert.True(t, apperrors.IsNotFound(err))

	got, err := repo.FindByAppointmentID(ctx, clinicA, appt.ID)
	require.NoError(t, err)
	assert.Empty(t, got.Options, "clinicB からの SetOptions は clinicA の appointment に影響してはならない")
}

// TestAppointmentTrimmingDetailRepository_FindByAppointmentID_CorrelatesAppointmentClinic
// SEC-SWEEP-02-TRIM-B1: appointment_trimming_details.appointment_id 読みは appointments.clinic_id と相関必須。
// 子 detail の clinic だけ一致して親 appointment が他院でも返ってしまう旧 failure mode を固定する。
func TestAppointmentTrimmingDetailRepository_FindByAppointmentID_CorrelatesAppointmentClinic(t *testing.T) {
	db := setupAppointmentTrimmingDetailTestDB(t)
	repo := NewAppointmentTrimmingDetailRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("rejects detail linked to foreign-clinic appointment", func(t *testing.T) {
		apptB := makeReservation(t, db, clinicB)
		// 子は clinicA を名乗りつつ親 appointment は clinicB — 旧実装は ClinicScope だけで返す。
		polluted := &model.AppointmentTrimmingDetail{
			ClinicID: clinicA, AppointmentID: apptB.ID, StyleRequest: "polluted-fk",
		}
		require.NoError(t, db.WithContext(ctx).Create(polluted).Error)

		_, err := repo.FindByAppointmentID(ctx, clinicA, apptB.ID)
		require.Error(t, err, "cross-tenant appointment parent must not yield a trimming detail")
		assert.True(t, apperrors.IsNotFound(err), "expected NotFound, got: %v", err)
	})

	t.Run("returns same-clinic appointment-linked detail", func(t *testing.T) {
		apptA := makeReservation(t, db, clinicA)
		valid := &model.AppointmentTrimmingDetail{
			ClinicID: clinicA, AppointmentID: apptA.ID, StyleRequest: "same-clinic",
		}
		require.NoError(t, db.WithContext(ctx).Create(valid).Error)

		got, err := repo.FindByAppointmentID(ctx, clinicA, apptA.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "same-clinic", got.StyleRequest)
		assert.Equal(t, apptA.ID, got.AppointmentID)
	})
}

// TestAppointmentTrimmingDetailRepository_SetOptions_CorrelatesAppointmentClinic
// SEC-SWEEP-02-TRIM-B1: SetOptions の対象行 lookup も appointments.clinic_id と相関必須。
// 書き込み列・置換意味は変えず、他院親に紐づく子行を読み出さず書き換えないことを固定する。
func TestAppointmentTrimmingDetailRepository_SetOptions_CorrelatesAppointmentClinic(t *testing.T) {
	db := setupAppointmentTrimmingDetailTestDB(t)
	repo := NewAppointmentTrimmingDetailRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("does not read or rewrite detail linked to foreign-clinic appointment", func(t *testing.T) {
		apptB := makeReservation(t, db, clinicB)
		polluted := &model.AppointmentTrimmingDetail{
			ClinicID: clinicA, AppointmentID: apptB.ID, StyleRequest: "must-not-change",
		}
		require.NoError(t, db.WithContext(ctx).Create(polluted).Error)
		pollutedID := polluted.ID

		opt := &model.TrimmingOption{ClinicID: clinicA, Name: "opt-A"}
		require.NoError(t, db.WithContext(ctx).Create(opt).Error)

		err := repo.SetOptions(ctx, clinicA, apptB.ID, []uint64{opt.ID})
		require.Error(t, err, "SetOptions must not target cross-tenant appointment parent edge")
		assert.True(t, apperrors.IsNotFound(err), "expected NotFound, got: %v", err)

		// 他院親に紐づく子行が書き換わっていないこと（StyleRequest 不変・Options 空）。
		var after model.AppointmentTrimmingDetail
		require.NoError(t, db.WithContext(ctx).First(&after, pollutedID).Error)
		assert.Equal(t, "must-not-change", after.StyleRequest)

		var optCount int64
		require.NoError(t, db.WithContext(ctx).Model(&model.AppointmentTrimmingOption{}).
			Where("appointment_id = ?", apptB.ID).Count(&optCount).Error)
		assert.Zero(t, optCount, "foreign-parent detail must not receive options")
	})

	t.Run("same-clinic SetOptions still replaces options", func(t *testing.T) {
		apptA := makeReservation(t, db, clinicA)
		valid := &model.AppointmentTrimmingDetail{ClinicID: clinicA, AppointmentID: apptA.ID}
		require.NoError(t, repo.Create(ctx, valid))
		opt := &model.TrimmingOption{ClinicID: clinicA, Name: "opt-same"}
		require.NoError(t, db.WithContext(ctx).Create(opt).Error)

		require.NoError(t, repo.SetOptions(ctx, clinicA, apptA.ID, []uint64{opt.ID}))
		got, err := repo.FindByAppointmentID(ctx, clinicA, apptA.ID)
		require.NoError(t, err)
		require.Len(t, got.Options, 1)
		assert.Equal(t, opt.ID, got.Options[0].ID)
	})
}
