package billing

// repo_test_helpers_test.go — internal/repository の test helper の最小限の複製
// （package 跨ぎ import 不能のため・shiftentry/reservation 先例）。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

func makeDoctor(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Staff {
	t.Helper()
	s := &model.Staff{ClinicID: clinicID, Name: name, StaffType: model.StaffTypeDoctor}
	require.NoError(t, db.WithContext(context.Background()).Create(s).Error)
	return s
}

func ptrU64(v uint64) *uint64 { return &v }
