package medicalrecord

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// UpsertCheckupFieldResultInput は健診結果値 1 件分の入力 DTO。
// status / is_abnormal はサーバ側で導出するため受け付けない（信頼境界はサーバ）。
// FieldName / FieldType / Unit / RefMin / RefMax はフィールド定義から解決するため受け付けない。
type UpsertCheckupFieldResultInput struct {
	CheckupTypeFieldID *uint64
	ValueNumber        *float64
	ValueText          string
	ValueBool          *bool
	ValueList          []string
}

// computeCheckupNumberStatus は number 型の値と基準値から status / is_abnormal を導出する
// （EXAM-001 の computeExamResultStatus と同方針。値は *float64）。
func computeCheckupNumberStatus(value, refMin, refMax *float64) (model.ExaminationResultStatus, bool) {
	if value == nil {
		return model.ExaminationResultStatusNormal, false
	}
	if refMin != nil && *value < *refMin {
		return model.ExaminationResultStatusLow, true
	}
	if refMax != nil && *value > *refMax {
		return model.ExaminationResultStatusHigh, true
	}
	return model.ExaminationResultStatusNormal, false
}

// CheckupFieldResultService は健診パッケージのフィールド定義取得・結果値の取得/置換を提供する。
type CheckupFieldResultService interface {
	// ListFields は指定 checkup_type のフィールド定義（FE 動的フォーム用）を返す。
	ListFields(ctx context.Context, clinicID, checkupTypeID uint64) ([]model.CheckupTypeField, error)
	// ListByCheckup は指定 checkup の結果値を返す（親 checkup 所有権を検証）。
	ListByCheckup(ctx context.Context, clinicID, medicalRecordID, checkupID uint64) ([]model.CheckupFieldResult, error)
	// ListByPet は pet 単位の健診結果を返す（飼い主レポート用）。
	ListByPet(ctx context.Context, clinicID, petID uint64) ([]model.CheckupFieldResult, error)
	// ReplaceForCheckup は健診結果値を一括置換する（PUT セマンティクス・#124 同型ガード付き）。
	// actorID は監査ログ用の操作スタッフ ID（nil = システム実行）。
	// inputs が nil（request で results 省略）は拒否する。意図的な全削除は明示的な空配列 [] を要求する（#211）。
	ReplaceForCheckup(ctx context.Context, clinicID, medicalRecordID, checkupID uint64, actorID *uint64, inputs []UpsertCheckupFieldResultInput) ([]model.CheckupFieldResult, error)
}

type checkupFieldResultService struct {
	checkupRepo       CheckupRepository
	medicalRecordRepo medicalRecordLocker
	fieldRepo         CheckupTypeFieldRepository
	resultRepo        CheckupFieldResultRepository
	auditTx           AuditTxLogger
	transactor        Transactor
}

// NewCheckupFieldResultService は CheckupFieldResultService の実装を返す。
// auditTx は tx 内監査（#211 fail-closed）の記録経路、transactor は「削除+挿入+監査」を
// 単一トランザクションで原子化するためのトランザクション境界。
func NewCheckupFieldResultService(
	checkupRepo CheckupRepository,
	medicalRecordRepo medicalRecordLocker,
	fieldRepo CheckupTypeFieldRepository,
	resultRepo CheckupFieldResultRepository,
	auditTx AuditTxLogger,
	transactor Transactor,
) CheckupFieldResultService {
	return &checkupFieldResultService{
		checkupRepo:       checkupRepo,
		medicalRecordRepo: medicalRecordRepo,
		fieldRepo:         fieldRepo,
		resultRepo:        resultRepo,
		auditTx:           auditTx,
		transactor:        transactor,
	}
}

func (s *checkupFieldResultService) ListFields(ctx context.Context, clinicID, checkupTypeID uint64) ([]model.CheckupTypeField, error) {
	fields, err := s.fieldRepo.FindByCheckupTypeID(ctx, clinicID, checkupTypeID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list checkup type fields")
	}
	return fields, nil
}

func (s *checkupFieldResultService) ListByCheckup(ctx context.Context, clinicID, medicalRecordID, checkupID uint64) ([]model.CheckupFieldResult, error) {
	if _, err := s.verifyCheckup(ctx, clinicID, medicalRecordID, checkupID); err != nil {
		return nil, err
	}
	results, err := s.resultRepo.FindByCheckupID(ctx, clinicID, checkupID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list checkup field results")
	}
	return results, nil
}

func (s *checkupFieldResultService) ListByPet(ctx context.Context, clinicID, petID uint64) ([]model.CheckupFieldResult, error) {
	results, err := s.resultRepo.FindByPetID(ctx, clinicID, petID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list pet checkup field results")
	}
	return results, nil
}

// buildCheckupFieldResults は fieldByID マップ構築・per-input 検証・field_type に応じた
// value 列マッピングを行う純関数（BE-refactor.md E-5）。field_type に該当する value 列のみ
// 書き込む（非該当列はゼロ値のまま） — 無条件に全列を書くと、例えば boolean フィールドに
// 送られた value_text 等の異種値が未使用列へ残留する（migration コントラクト違反）。
func buildCheckupFieldResults(clinicID, checkupID uint64, fields []model.CheckupTypeField, inputs []UpsertCheckupFieldResultInput) ([]model.CheckupFieldResult, error) {
	fieldByID := make(map[uint64]model.CheckupTypeField, len(fields))
	for i := range fields {
		fieldByID[fields[i].ID] = fields[i]
	}

	results := make([]model.CheckupFieldResult, 0, len(inputs))
	for _, in := range inputs {
		if in.CheckupTypeFieldID == nil {
			return nil, apperrors.WrapInvalidInput("checkup_type_field_id は必須です")
		}
		field, ok := fieldByID[*in.CheckupTypeFieldID]
		if !ok {
			return nil, apperrors.WrapInvalidInput("checkup_type_field が当該健診パッケージに属していません（別クリニック/別パッケージの項目は紐付けできません）")
		}
		if err := validateCheckupFieldValue(&field, in); err != nil {
			return nil, err
		}
		result := model.CheckupFieldResult{
			ClinicID:           clinicID,
			CheckupID:          checkupID,
			CheckupTypeFieldID: in.CheckupTypeFieldID,
			FieldName:          field.Name,
			FieldType:          field.FieldType,
			Unit:               field.Unit,
			SortOrder:          field.SortOrder,
			Status:             model.ExaminationResultStatusNormal,
		}
		switch field.FieldType {
		case model.CheckupFieldTypeNumber:
			result.ValueNumber = in.ValueNumber
			result.RefMin = field.MinValue
			result.RefMax = field.MaxValue
			result.Status, result.IsAbnormal = computeCheckupNumberStatus(in.ValueNumber, field.MinValue, field.MaxValue)
		case model.CheckupFieldTypeBoolean:
			result.ValueBool = in.ValueBool
		case model.CheckupFieldTypeMultiSelect, model.CheckupFieldTypeChecklist:
			result.ValueList = in.ValueList
		case model.CheckupFieldTypeSingleSelect, model.CheckupFieldTypeText:
			result.ValueText = in.ValueText
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *checkupFieldResultService) ReplaceForCheckup(ctx context.Context, clinicID, medicalRecordID, checkupID uint64, actorID *uint64, inputs []UpsertCheckupFieldResultInput) ([]model.CheckupFieldResult, error) {
	// #211 偶発的全消去の防御: results 省略（nil）は拒否する。患者検診結果値の全削除は
	// 明示的な空配列 [] の送信を要求し、空ボディ/壊れた request での silent な全消去を遮断する。
	// 明示的な空配列での全削除は下流で監査される（PUT セマンティクスとしての意図的 clear は許容）。
	if inputs == nil {
		return nil, apperrors.WrapInvalidInput("results は必須です（全削除する場合も results: [] を明示送信してください）")
	}

	checkup, err := s.verifyCheckup(ctx, clinicID, medicalRecordID, checkupID)
	if err != nil {
		return nil, err
	}

	// #124 同型ガード: request の checkup_type_field_id が caller の clinic に属する
	// 当該 checkup_type のフィールドであることを検証する。別 clinic / 別パッケージの
	// フィールドを紐付けると、そのフィールドの定義（型・基準値・選択肢）が結果に誤適用される。
	//
	// X-11 の意図的な trade-off（go-reviewer 指摘）: 親カルテ確定済みチェックは
	// s.transactor.WithTx 内（この後段）に移動したため、本チェックより後に評価される。
	// 「確定済みカルテ」かつ「所有権不正な checkup_type_field_id」の両方に該当する request は、
	// 旧実装では WrapConflict（確定済み）を返したが、本実装では WrapInvalidInput（フィールド不正）
	// を返す（優先順位が入れ替わる）。いずれの分岐でも書込は拒否されるため安全性への影響はないが、
	// LockByIDForUpdate の行ロック保持時間を最小化する（DB に依存しない検証を先に済ませてからロックする）
	// ため意図的に許容した順序変更である。
	fields, err := s.fieldRepo.FindByCheckupTypeID(ctx, clinicID, checkup.CheckupTypeID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to load checkup type fields")
	}
	results, err := buildCheckupFieldResults(clinicID, checkupID, fields, inputs)
	if err != nil {
		return nil, err
	}

	// #211 tx 内監査による原子的置換: スナップショット読取→削除/挿入→削除監査 を単一トランザクションで
	// 実行する。監査書込が失敗したら tx 全体を rollback し、削除・挿入も巻き戻す（監査なしの患者検診結果
	// 削除を許さない＝fail-closed）。checkup_field_results は hard-delete のため old_value が唯一の耐久記録
	// であり、旧 best-effort では「置換 commit 後に監査書込が落ちると無記録削除が残る」窓があった
	// （healthcare review MEDIUM-1）。スナップショットも同一 tx 内で取得し、旧コードの
	// 「スナップショット↔削除」TOCTOU 窓も同時に解消する。
	var saved []model.CheckupFieldResult
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// 親カルテ確定済みなら編集拒否（checkup Create/Update と対称）。BE-refactor.md X-11:
		// LockByIDForUpdate の行ロックで finalize と直列化し、確定と同時の健診結果編集が確定済みカルテに
		// 混入する競合を防ぐ。
		if err := lockDraftMedicalRecord(txCtx, s.medicalRecordRepo, clinicID, medicalRecordID,
			"failed to find medical record", "確定済みカルテのため健診結果は編集できません"); err != nil {
			return err
		}

		existing, err := s.resultRepo.FindByCheckupID(txCtx, clinicID, checkupID)
		if err != nil {
			return apperrors.Wrap(err, "failed to load existing checkup field results")
		}

		replaced, deletedCount, err := s.resultRepo.ReplaceForCheckup(txCtx, clinicID, checkupID, results)
		if err != nil {
			return apperrors.Wrap(err, "failed to persist replaced checkup field results")
		}
		saved = replaced

		// 実際に削除が発生した場合のみ監査する（純粋な新規挿入は削除を伴わない）。ゲートはスナップショット
		// 件数でなく DELETE の実削除数（deletedCount）に基づく（#211 security MEDIUM-1: 並行 INSERT 競合下で
		// スナップショット 0 件でも実削除>0 を取りこぼさず、無監査 hard-delete を残さない）。
		// 監査書込失敗はエラーを返して tx を rollback する（best-effort ではなく fail-closed）。
		return logReplaceDeletionTx(txCtx, s.auditTx, clinicID, actorID, deletedCount,
			model.AuditActionCheckupFieldResultReplace, model.AuditResourceCheckupFieldResult, checkupID,
			extractCheckupFieldResultsAudit(existing), extractCheckupFieldResultsAudit(saved),
			map[string]any{
				"medical_record_id": medicalRecordID,
				"checkup_id":        checkupID,
				"deleted_count":     deletedCount,
				"new_count":         len(saved),
			},
			"audit log failed for checkup field results replace; rolling back deletion",
			"failed to write checkup field results deletion audit", "checkup_id")
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to replace checkup field results in transaction")
	}

	slog.InfoContext(ctx, "checkup field results replaced",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("checkup_id", checkupID),
		slog.Int("result_count", len(saved)),
	)
	return saved, nil
}

// verifyCheckup は checkup の clinic 所有権 + 親カルテ整合を検証する（404 を返すため）。
func (s *checkupFieldResultService) verifyCheckup(ctx context.Context, clinicID, medicalRecordID, checkupID uint64) (*model.Checkup, error) {
	checkup, err := s.checkupRepo.FindByID(ctx, clinicID, checkupID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find checkup")
	}
	if checkup.MedicalRecordID != medicalRecordID {
		return nil, apperrors.WrapNotFound("checkup", fmt.Sprintf("%d", checkupID))
	}
	return checkup, nil
}

// checkupFieldOption は checkup_type_fields.options（jsonb）の 1 要素。
type checkupFieldOption struct {
	Value string `json:"value"`
}

// validateCheckupFieldValue は request 値が field_type / options 定義に整合するか検証する
// （request 境界の untrusted 入力検証。型/スキーマで排除できないため残す）。
func validateCheckupFieldValue(field *model.CheckupTypeField, in UpsertCheckupFieldResultInput) error {
	switch field.FieldType {
	case model.CheckupFieldTypeSingleSelect:
		if in.ValueText == "" {
			return nil // 未入力は許容（任意項目）
		}
		allowed, err := parseCheckupOptionValues(field)
		if err != nil {
			return err
		}
		if _, ok := allowed[in.ValueText]; !ok {
			return apperrors.WrapInvalidInput(fmt.Sprintf("%s の選択値が選択肢に存在しません", field.Name))
		}
	case model.CheckupFieldTypeMultiSelect, model.CheckupFieldTypeChecklist:
		if len(in.ValueList) == 0 {
			return nil
		}
		allowed, err := parseCheckupOptionValues(field)
		if err != nil {
			return err
		}
		for _, v := range in.ValueList {
			if _, ok := allowed[v]; !ok {
				return apperrors.WrapInvalidInput(fmt.Sprintf("%s の選択値が選択肢に存在しません", field.Name))
			}
		}
	default:
		// number / boolean / text は値自体の制約なし（number は status をサーバ導出）。
	}
	return nil
}

func parseCheckupOptionValues(field *model.CheckupTypeField) (map[string]struct{}, error) {
	allowed := map[string]struct{}{}
	if len(field.Options) == 0 {
		return allowed, nil
	}
	var opts []checkupFieldOption
	if err := json.Unmarshal(field.Options, &opts); err != nil {
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("%s の選択肢定義が不正です", field.Name))
	}
	for _, o := range opts {
		allowed[o.Value] = struct{}{}
	}
	return allowed, nil
}

// extractCheckupFieldResultsAudit は監査ログ用に結果値の PII フリーなスナップショットを構築する。
// 飼い主/患者の識別情報は含まず、行 ID・フィールド定義（field_name/field_type）と入力値のみを記録する
// （LogVitalChange / LogMedicalRecordChange が臨床値を old/new に残すのと同方針）。
// checkup_type_field_id はフィールド定義 hard-delete 時に SET NULL されうるため（migration 010）、
// nil 安全に *uint64 のまま格納する（json.Marshal が null として出力する）。
func extractCheckupFieldResultsAudit(results []model.CheckupFieldResult) []map[string]any {
	if len(results) == 0 {
		return nil
	}
	out := make([]map[string]any, 0, len(results))
	for i := range results {
		r := results[i]
		entry := map[string]any{
			"id":                    r.ID,
			"checkup_type_field_id": r.CheckupTypeFieldID,
			"field_name":            r.FieldName,
			"field_type":            string(r.FieldType),
			"is_abnormal":           r.IsAbnormal,
		}
		switch r.FieldType {
		case model.CheckupFieldTypeNumber:
			entry["value_number"] = r.ValueNumber
		case model.CheckupFieldTypeBoolean:
			entry["value_bool"] = r.ValueBool
		case model.CheckupFieldTypeMultiSelect, model.CheckupFieldTypeChecklist:
			entry["value_list"] = []string(r.ValueList)
		case model.CheckupFieldTypeSingleSelect, model.CheckupFieldTypeText:
			entry["value_text"] = r.ValueText
		}
		out = append(out, entry)
	}
	return out
}
