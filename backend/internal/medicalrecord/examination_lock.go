package medicalrecord

import (
	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// examinationFullyLocked is the official seal: no parent or result mutation.
//
// status は NULL 許容列。未記録（nil）は封印済みかどうかを判定できないため、編集可能側へ倒さず
// sharedkernel.RequireStored のエラーで拒否する（EMR-549 C3・臨床安全）。
func examinationFullyLocked(exam *model.Examination) (bool, error) {
	if exam == nil {
		return false, nil
	}
	status, err := sharedkernel.RequireStored(exam.Status, "exams.status")
	if err != nil {
		return false, err
	}
	return status == model.ExaminationStatusConfirmed, nil
}

// examinationResultsLocked rejects result-row mutation and delete of sealed exams.
//
// - confirmed: always locked
// - completed without revision history: first-pass completion seal (BUG-033)
// - completed with revision history: post-unconfirm working copy — still editable
// - status 未記録（nil）: 判定不能のためエラー（編集可能側へは倒さない。EMR-549 C3）
func examinationResultsLocked(exam *model.Examination) (bool, error) {
	if exam == nil {
		return false, nil
	}
	status, err := sharedkernel.RequireStored(exam.Status, "exams.status")
	if err != nil {
		return false, err
	}
	switch status {
	case model.ExaminationStatusConfirmed:
		return true, nil
	case model.ExaminationStatusCompleted:
		return exam.CurrentRevisionVersion == nil, nil
	default:
		return false, nil
	}
}

func errExaminationFullyLocked() error {
	return apperrors.WrapConflict("確定済みの検査は編集できません")
}

// errExaminationResultsLocked / errExaminationDeleteLocked は文言の選択だけを行う。
// 拒否そのものは呼び出し元の lock 判定（nil はエラー）で確定済みのため、PtrEquals の false でも編集は許可されない。
func errExaminationResultsLocked(exam *model.Examination) error {
	if exam != nil && sharedkernel.PtrEquals(exam.Status, model.ExaminationStatusConfirmed) {
		return errExaminationFullyLocked()
	}
	return apperrors.WrapConflict("完了済みの検査結果は編集できません")
}

func errExaminationDeleteLocked(exam *model.Examination) error {
	if exam != nil && sharedkernel.PtrEquals(exam.Status, model.ExaminationStatusConfirmed) {
		return apperrors.WrapConflict("確定済みの検査は削除できません")
	}
	return apperrors.WrapConflict("完了済みの検査は削除できません")
}
