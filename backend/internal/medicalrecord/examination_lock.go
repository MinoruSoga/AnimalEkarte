package medicalrecord

import (
	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// examinationFullyLocked is the official seal: no parent or result mutation.
func examinationFullyLocked(exam *model.Examination) bool {
	return exam != nil && exam.Status == model.ExaminationStatusConfirmed
}

// examinationResultsLocked rejects result-row mutation and delete of sealed exams.
//
// - confirmed: always locked
// - completed without revision history: first-pass completion seal (BUG-033)
// - completed with revision history: post-unconfirm working copy — still editable
func examinationResultsLocked(exam *model.Examination) bool {
	if exam == nil {
		return false
	}
	switch exam.Status {
	case model.ExaminationStatusConfirmed:
		return true
	case model.ExaminationStatusCompleted:
		return exam.CurrentRevisionVersion == nil
	default:
		return false
	}
}

func errExaminationFullyLocked() error {
	return apperrors.WrapConflict("確定済みの検査は編集できません")
}

func errExaminationResultsLocked(exam *model.Examination) error {
	if exam != nil && exam.Status == model.ExaminationStatusConfirmed {
		return errExaminationFullyLocked()
	}
	return apperrors.WrapConflict("完了済みの検査結果は編集できません")
}

func errExaminationDeleteLocked(exam *model.Examination) error {
	if exam != nil && exam.Status == model.ExaminationStatusConfirmed {
		return apperrors.WrapConflict("確定済みの検査は削除できません")
	}
	return apperrors.WrapConflict("完了済みの検査は削除できません")
}
