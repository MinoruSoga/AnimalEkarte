package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ListItems は検査項目一覧を返す。clinic_id 隔離は repository の JOIN 条件で保証する。
// 親 exam の存在確認は FindByID で先行する（404 を返すため）。
func (s *examinationService) ListItems(ctx context.Context, clinicID, examID uint64) ([]model.ExamResult, error) {
	// Use repo.FindByID directly so GetByID's clinical-detail usage receipt is not double-written.
	exam, err := s.repo.FindByID(ctx, clinicID, examID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find examination")
	}
	if err := s.usage().RecordClinicalUse(ctx, clinicID, exam, model.LabImportUsageKindExaminationItems, nil); err != nil {
		return nil, err
	}
	items, err := s.repo.FindAllItemsByExamID(ctx, clinicID, examID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list examination items", "error", err, "exam_id", examID, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list examination items")
	}
	return items, nil
}

// ReplaceItems は検査項目を一括置換する（PUT セマンティクス）。
//
// 仕様:
//  1. 親 exam の存在を FindByID で確認（P1）
//  2. 親 exam が confirmed、または revision なし completed（BUG-033 初回完了シール）の場合は Conflict (409) で拒否
//  3. 各 input の inspection_value とサーバで解決した基準値から status / is_abnormal を導出
//  4. repository の ReplaceItemsByExamID（トランザクション内で全削除→一括挿入）に委譲
//  5. 実削除が発生した場合（deletedCount > 0）は同一 tx 内で監査ログを書き込む。監査書込が失敗したら
//     tx を rollback する（best-effort ではなく fail-closed。BE-refactor.md R1-2・#211 と同方針）。
func (s *examinationService) ReplaceItems(ctx context.Context, clinicID, examID uint64, actorID *uint64, inputs []UpsertExamItemInput) ([]model.ExamResult, error) {
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("examination write transaction dependency is required")
	}

	// #211/R1-2 tx 内監査による原子的置換: スナップショット読取→削除/挿入→削除監査 を単一トランザクションで
	// 実行する。監査書込が失敗したら tx 全体を rollback し、削除・挿入も巻き戻す（監査なしの検査結果削除を
	// 許さない＝fail-closed）。exam_results は hard-delete のため old_value が唯一の耐久記録であり、
	// 旧コードでは「置換 commit 後に監査を書く」経路自体が存在しなかった（audit_tx_inventory_lint_test.go
	// が発見した無監査ギャップ）。スナップショットも同一 tx 内で取得し TOCTOU 窓を作らない。
	var saved []model.ExamResult
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		locked, err := s.repo.LockByIDForUpdate(txCtx, clinicID, examID)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock examination")
		}
		if examinationResultsLocked(locked) {
			return errExaminationResultsLocked(locked)
		}
		before := *locked
		revisioned := locked.CurrentRevisionVersion != nil
		if revisioned {
			if s.revisionWorkflow == nil {
				return apperrors.WrapInternalServerError("examination revision workflow repository capability is required")
			}
			if err := s.validateParentMutationAudit(actorID); err != nil {
				return err
			}
		}
		if locked.MedicalRecordID != nil {
			if err := lockDraftMedicalRecord(
				txCtx,
				s.medRec,
				clinicID,
				*locked.MedicalRecordID,
				"failed to find medical record",
				"確定済みカルテの検査結果は編集できません",
			); err != nil {
				return err
			}
		}

		replaced, err := s.replaceItemsTx(txCtx, clinicID, locked, actorID, inputs)
		if err != nil {
			return err
		}
		saved = replaced
		if revisioned {
			if _, err = s.appendWorkingRevisionTx(
				txCtx,
				clinicID,
				actorID,
				&before,
				locked,
				examinationWorkingItemsReason,
			); err != nil {
				return err
			}
			return s.usage().RecordManualMutation(txCtx, clinicID, locked, actorID)
		}
		return s.usage().RecordManualMutation(txCtx, clinicID, locked, actorID)
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to replace examination items in transaction")
	}

	slog.InfoContext(ctx, "examination items replaced",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("examination_id", examID),
		slog.Int("item_count", len(saved)),
	)
	return saved, nil
}

// replaceItemsTx validates and replaces examination results inside the caller-owned transaction.
// Create, Update, and the split PUT endpoint all share this tail so result persistence and
// deletion audit use the same transaction as the parent mutation.
func (s *examinationService) replaceItemsTx(
	ctx context.Context,
	clinicID uint64,
	exam *model.Examination,
	actorID *uint64,
	inputs []UpsertExamItemInput,
) ([]model.ExamResult, error) {
	fieldIDs := make([]uint64, 0, len(inputs))
	fieldIDSet := make(map[uint64]struct{}, len(inputs))
	for _, in := range inputs {
		if in.ExamTypeFieldID != nil {
			if _, exists := fieldIDSet[*in.ExamTypeFieldID]; !exists {
				fieldIDSet[*in.ExamTypeFieldID] = struct{}{}
				fieldIDs = append(fieldIDs, *in.ExamTypeFieldID)
			}
		}
	}

	// #124 防止: request の exam_type_field が caller の clinic に属する、ロック済み検査の
	// 検査種別フィールドであることを同じ transaction 内で検証する。
	if len(fieldIDs) > 0 {
		examType, err := s.examTypeRepo.FindByID(ctx, clinicID, exam.ExamTypeID)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to verify exam type ownership")
		}
		validFieldIDs := make(map[uint64]struct{}, len(examType.Items))
		for i := range examType.Items {
			validFieldIDs[examType.Items[i].ID] = struct{}{}
		}
		for _, in := range inputs {
			if in.ExamTypeFieldID != nil {
				if _, ok := validFieldIDs[*in.ExamTypeFieldID]; !ok {
					return nil, apperrors.WrapInvalidInput("exam_type_field が当該検査種別に属していません（別クリニック/別種別の項目は紐付けできません）")
				}
			}
		}
	}

	resolvedRanges := make(map[uint64]model.ExamReferenceRange, len(fieldIDs))
	if len(fieldIDs) > 0 {
		if exam.PetID == nil {
			return nil, apperrors.WrapInvalidInput("基準値を解決するには検査対象のペットが必要です")
		}
		if s.referenceRanges == nil {
			return nil, apperrors.WrapInternalServerError("examination reference range resolver is required")
		}
		animalSpeciesID, err := s.referenceRanges.FindAnimalSpeciesID(ctx, clinicID, exam.ID)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to resolve examination animal species")
		}
		resolvedRanges, err = s.referenceRanges.ResolveByFieldIDs(ctx, clinicID, animalSpeciesID, fieldIDs)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to resolve examination reference ranges")
		}
	}

	items := make([]model.ExamResult, 0, len(inputs))
	for _, in := range inputs {
		var refMin, refMax *float64
		var qualitativeMin, qualitativeMax *string
		if in.ExamTypeFieldID != nil {
			if referenceRange, ok := resolvedRanges[*in.ExamTypeFieldID]; ok {
				refMin = cloneOptionalFloat64(referenceRange.RefMin)
				refMax = cloneOptionalFloat64(referenceRange.RefMax)
				qualitativeMin = cloneOptionalString(referenceRange.QualitativeMin)
				qualitativeMax = cloneOptionalString(referenceRange.QualitativeMax)
			}
		}
		assessment := assessExamResult(
			in.InspectionValue,
			refMin,
			refMax,
			qualitativeMin,
			qualitativeMax,
		)
		items = append(items, model.ExamResult{
			ExamID:          exam.ID,
			ExamTypeItemID:  in.ExamTypeFieldID,
			Name:            in.Name,
			InspectionValue: in.InspectionValue,
			NormalValue:     in.NormalValue,
			Result:          in.Result,
			Unit:            in.Unit,
			ReferenceValue:  in.ReferenceValue,
			RefMin:          refMin,
			RefMax:          refMax,
			QualitativeMin:  qualitativeMin,
			QualitativeMax:  qualitativeMax,
			IsAbnormal:      assessment.isAbnormal,
			Status:          assessment.status,
			SortOrder:       in.SortOrder,
		})
	}

	before, err := s.repo.FindAllItemsByExamID(ctx, clinicID, exam.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to snapshot existing examination items before replace", "error", err, "exam_id", exam.ID, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to load existing examination items")
	}

	saved, deletedCount, err := s.repo.ReplaceItemsByExamID(ctx, clinicID, exam.ID, items)
	if err != nil {
		slog.ErrorContext(ctx, "failed to replace examination items", "error", err, "exam_id", exam.ID, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to replace examination items")
	}

	// 実際に削除が発生した場合のみ監査する（純粋な新規挿入は削除を伴わない）。ゲートはスナップショット
	// 件数でなく DELETE の実削除数（deletedCount）に基づく（#211 security MEDIUM-1 と同方針: 並行 INSERT
	// 競合下でスナップショット 0 件でも実削除>0 を取りこぼさない）。監査書込失敗は tx を rollback する。
	if err := logReplaceDeletionTx(ctx, s.auditTx, clinicID, actorID, deletedCount,
		model.AuditActionExamResultReplace, model.AuditResourceExamResult, exam.ID,
		extractExamResultsAudit(before), extractExamResultsAudit(saved),
		map[string]any{
			"exam_id":       exam.ID,
			"deleted_count": deletedCount,
			"new_count":     len(saved),
		},
		"audit log failed for examination items replace; rolling back deletion",
		"failed to write examination items deletion audit", "exam_id"); err != nil {
		return nil, err
	}
	return saved, nil
}

func cloneOptionalFloat64(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

// extractExamResultsAudit は監査ログの old_value/new_value に格納する検査結果値のスナップショットを構築する。
// 飼主/患者の識別情報は含まず、行 ID・フィールド定義参照・検査値のみを記録する
// （extractCheckupFieldResultsAudit と同方針）。
func extractExamResultsAudit(results []model.ExamResult) []map[string]any {
	if len(results) == 0 {
		return nil
	}
	out := make([]map[string]any, 0, len(results))
	for i := range results {
		r := results[i]
		// is_assessed is derived, not stored — reuse assessExamResult so audit
		// provenance matches API assessment rules (bounds, fail-closed cases).
		assessment := assessExamResult(
			r.InspectionValue,
			r.RefMin,
			r.RefMax,
			r.QualitativeMin,
			r.QualitativeMax,
		)
		out = append(out, map[string]any{
			"id":                 r.ID,
			"exam_type_field_id": r.ExamTypeItemID,
			"name":               r.Name,
			"inspection_value":   r.InspectionValue,
			"ref_min":            r.RefMin,
			"ref_max":            r.RefMax,
			"qualitative_min":    r.QualitativeMin,
			"qualitative_max":    r.QualitativeMax,
			"is_assessed":        assessment.isAssessed,
			"is_abnormal":        r.IsAbnormal,
			"status":             string(r.Status),
		})
	}
	return out
}
