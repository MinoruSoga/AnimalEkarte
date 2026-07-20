package repository

// exam_type_repository_test.go — ExamTypeRepository's own test suite moved to
// internal/medicalrecord/exam_type_repository_test.go (BE9-2C —
// internal/repository/examtype/ was deleted, roll-up complete). makeExamTypeMaster/
// makeExaminationRec stay here because examination_repository_test.go and
// examination_repository_tx_atomicity_test.go (both still in internal/repository, out of
// this batch's scope — Examination is BE9-2D territory) construct ExaminationType/
// Examination fixtures via these helpers and need them to keep compiling.

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// makeExamTypeMaster はテスト用の ExaminationType を作成して返す（examination_repository_test.go からも再利用）。
func makeExamTypeMaster(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.ExaminationType {
	t.Helper()
	et := &model.ExaminationType{ClinicID: clinicID, Name: name, IsActive: true}
	require.NoError(t, db.WithContext(context.Background()).Create(et).Error)
	return et
}

// makeExaminationRec はテスト用の Examination を作成して返す（examination_repository_test.go からも再利用）。
func makeExaminationRec(t *testing.T, db *gorm.DB, ex *model.Examination) *model.Examination {
	t.Helper()
	require.NoError(t, db.WithContext(context.Background()).Create(ex).Error)
	return ex
}
