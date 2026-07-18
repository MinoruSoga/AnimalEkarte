package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// makeMedicineMaster / makeDoctor は clinic 隔離テスト群で共有するマスタ生成ヘルパー。
// （旧 medical_record_owner_medication_test.go に同居していたが、#158 エンドポイント削除に伴い分離）

func makeMedicineMaster(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Medicine {
	t.Helper()
	m := &model.Medicine{ClinicID: clinicID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(m).Error)
	return m
}

func makeDoctor(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Staff {
	t.Helper()
	s := &model.Staff{ClinicID: clinicID, Name: name, StaffType: model.StaffTypeDoctor}
	require.NoError(t, db.WithContext(context.Background()).Create(s).Error)
	return s
}

// makeShiftEntryWithType は shiftentry/repository_test.go の同名ヘルパーの最小限の複製
// （BE8-4 batch13: shift_entry_repository_test.go の移動先パッケージから本パッケージの
// reservation_type_occupation_repository_test.go が引き続き参照するため）。
func makeShiftEntryWithType(t *testing.T, db *gorm.DB, clinicID, staffID uint64, date time.Time, shiftType model.ShiftType) *model.ShiftEntry {
	t.Helper()
	e := &model.ShiftEntry{
		ClinicID:  clinicID,
		StaffID:   staffID,
		Date:      date,
		ShiftType: shiftType,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(e).Error)
	return e
}
