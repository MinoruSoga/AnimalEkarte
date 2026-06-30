// Package service — 健診パッケージの型付き結果値（#211）のビジネスロジック。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
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
	ReplaceForCheckup(ctx context.Context, clinicID, medicalRecordID, checkupID uint64, inputs []UpsertCheckupFieldResultInput) ([]model.CheckupFieldResult, error)
}

type checkupFieldResultService struct {
	checkupRepo       repository.CheckupRepository
	medicalRecordRepo repository.MedicalRecordRepository
	fieldRepo         repository.CheckupTypeFieldRepository
	resultRepo        repository.CheckupFieldResultRepository
}

// NewCheckupFieldResultService は CheckupFieldResultService の実装を返す。
func NewCheckupFieldResultService(
	checkupRepo repository.CheckupRepository,
	medicalRecordRepo repository.MedicalRecordRepository,
	fieldRepo repository.CheckupTypeFieldRepository,
	resultRepo repository.CheckupFieldResultRepository,
) CheckupFieldResultService {
	return &checkupFieldResultService{
		checkupRepo:       checkupRepo,
		medicalRecordRepo: medicalRecordRepo,
		fieldRepo:         fieldRepo,
		resultRepo:        resultRepo,
	}
}

func (s *checkupFieldResultService) ListFields(ctx context.Context, clinicID, checkupTypeID uint64) ([]model.CheckupTypeField, error) {
	fields, err := s.fieldRepo.FindByCheckupTypeID(ctx, clinicID, checkupTypeID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list checkup type fields", "error", err, "checkup_type_id", checkupTypeID, "clinic_id", clinicID)
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
		slog.ErrorContext(ctx, "failed to list checkup field results", "error", err, "checkup_id", checkupID, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list checkup field results")
	}
	return results, nil
}

func (s *checkupFieldResultService) ListByPet(ctx context.Context, clinicID, petID uint64) ([]model.CheckupFieldResult, error) {
	results, err := s.resultRepo.FindByPetID(ctx, clinicID, petID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list pet checkup field results", "error", err, "pet_id", petID, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list pet checkup field results")
	}
	return results, nil
}

func (s *checkupFieldResultService) ReplaceForCheckup(ctx context.Context, clinicID, medicalRecordID, checkupID uint64, inputs []UpsertCheckupFieldResultInput) ([]model.CheckupFieldResult, error) {
	checkup, err := s.verifyCheckup(ctx, clinicID, medicalRecordID, checkupID)
	if err != nil {
		return nil, err
	}

	// 親カルテ確定済みなら編集拒否（checkup Create/Update と対称）。
	parent, err := s.medicalRecordRepo.FindByID(ctx, clinicID, medicalRecordID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find medical record", "error", err)
		return nil, apperrors.Wrap(err, "failed to find medical record")
	}
	if parent.Status == model.MedicalRecordStatusFinalized {
		return nil, apperrors.WrapConflict("確定済みカルテのため健診結果は編集できません")
	}

	// #124 同型ガード: request の checkup_type_field_id が caller の clinic に属する
	// 当該 checkup_type のフィールドであることを検証する。別 clinic / 別パッケージの
	// フィールドを紐付けると、そのフィールドの定義（型・基準値・選択肢）が結果に誤適用される。
	fields, err := s.fieldRepo.FindByCheckupTypeID(ctx, clinicID, checkup.CheckupTypeID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load checkup type fields", "error", err, "checkup_type_id", checkup.CheckupTypeID, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to load checkup type fields")
	}
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
		if err := validateCheckupFieldValue(field, in); err != nil {
			return nil, err
		}
		// field_type に該当する value 列のみ書き込む（非該当列はゼロ値のまま）。
		// 無条件に全列を書くと、例えば boolean フィールドに送られた value_text 等の
		// 異種値が未使用列へ残留する（migration コントラクト「該当列のみ書込」に違反）。
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

	saved, err := s.resultRepo.ReplaceForCheckup(ctx, clinicID, checkupID, results)
	if err != nil {
		slog.ErrorContext(ctx, "failed to replace checkup field results", "error", err, "checkup_id", checkupID, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to replace checkup field results")
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
		// NotFound は 404 正常系のためログ不要（P11）。DB 障害等のみ記録する。
		if !apperrors.IsNotFound(err) {
			slog.ErrorContext(ctx, "failed to find checkup", "error", err, "checkup_id", checkupID, "clinic_id", clinicID)
		}
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
func validateCheckupFieldValue(field model.CheckupTypeField, in UpsertCheckupFieldResultInput) error {
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

func parseCheckupOptionValues(field model.CheckupTypeField) (map[string]struct{}, error) {
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
