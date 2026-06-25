package repository

// line_reservation_setting_clinic_isolation_test.go — #196 clinic_id テナント隔離回帰テスト
//
// テスト対象: LineReservationSettingRepository の clinic_id 境界
// 保護する不変条件: clinic A の LINE 設定を clinic B のスコープで
//   取得できない。
//
// clinicScope を line_reservation_setting_repository.go から削除すると必ず失敗するよう設計されている。
// 注意: FindAll は Webhook 署名検証用の設計で clinic_id フィルタなし（全件返却が正常動作）。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// setupLineSettingIsolationTestDB は LINE 予約設定 clinic_id 隔離テスト用の DB を整備する。
// line_reservation_settings は TRUNCATE で初期化する。
func setupLineSettingIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.LineReservationSetting{}))
	db.Exec("TRUNCATE TABLE line_reservation_settings CASCADE")
	return db
}

// makeLineReservationSetting はテスト用 LINE 予約設定を1件作成する。
// not null DEFAULT なし列（additional_fields）を明示的に設定する。
func makeLineReservationSetting(t *testing.T, db *gorm.DB, clinicID uint64) *model.LineReservationSetting {
	t.Helper()
	setting := &model.LineReservationSetting{
		ClinicID:         clinicID,
		ClosedWeekdays:   []byte(`[]`),
		ClosedDates:      []byte(`[]`),
		BusinessHours:    []byte(`{"start":"0900","end":"1900"}`),
		BreakHours:       []byte(`[{"start":"1200","end":"1300"}]`),
		AdditionalFields: []byte(`{}`),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(setting).Error)
	return setting
}

// TestLineReservationSettingRepository_FindByClinicID_ClinicIsolation は
// clinic A の設定を clinic B の clinicID で取得できないことを検証する。
// clinicScope を削除すると「別クリニックIDでは取得できない」が失敗する。
func TestLineReservationSettingRepository_FindByClinicID_ClinicIsolation(t *testing.T) {
	db := setupLineSettingIsolationTestDB(t)
	repo := NewLineReservationSettingRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	makeLineReservationSetting(t, db, clinicA)

	t.Run("同一クリニックIDでは取得できる", func(t *testing.T) {
		got, err := repo.FindByClinicID(ctx, clinicA)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, clinicA, got.ClinicID)
	})

	t.Run("別クリニックIDでは取得できない（clinic_id 隔離）", func(t *testing.T) {
		// clinic B に設定が存在しない → clinicScope が正しく機能すれば NotFound を返す。
		got, err := repo.FindByClinicID(ctx, clinicB)
		assert.Error(t, err, "clinic B は clinic A の LINE 設定を参照できてはならない")
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})
}
