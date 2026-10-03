package medicalrecord

import (
	"context"
	"encoding/json"
	"math"
	"strings"

	"gorm.io/datatypes"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// checkup_field_service.go — 健診パッケージのフィールド定義マスタ（checkup_type_fields）の
// 書込側サービス（EMR-225）。読み取り側は CheckupFieldResultService.ListFields が担うため、
// 本サービスは Create/Update/Delete/Reorder のみを持つ。exam_type_field.go の
// examTypeService フィールド群と同型: WithTx 境界 + LockFieldByID 行ロック +
// clinic_id+checkup_type_id 複合スコープ（migration-013 複合 FK が enforcement model）。
//
// 履歴保全契約: checkup_field_results は field_name/field_type/unit/ref_min/ref_max を
// 非正規化保持し、checkup_type_field_id FK は ON DELETE SET NULL。本サービスは定義行のみを
// 操作し、結果行のスナップショット列は一切書き換えない（A2/A3）。

// CheckupFieldOptionInput は選択肢（options）1件分の入力 DTO。
type CheckupFieldOptionInput struct {
	Value string
	Label string
}

// CreateCheckupTypeFieldInput はフィールド定義作成の入力 DTO（明示 struct — mass assignment 防止）。
// IsProvisional / ImportNamespace / ImportKey はパッケージ import 専用のため受け付けない。
type CreateCheckupTypeFieldInput struct {
	Name      string
	FieldType string
	Unit      string
	MinValue  *float64
	MaxValue  *float64
	Options   []CheckupFieldOptionInput
	SortOrder int
}

// UpdateCheckupTypeFieldInput はフィールド定義部分更新の入力 DTO。ポインタ nil = 変更しない。
// Options は *[] で「キー省略（nil）= 維持」と「明示配列（空配列含む）= 上書き」を区別する。
// ClearMinValue / ClearMaxValue は基準値を NULL に戻す明示フラグ（clear_parent_id と同型）。
type UpdateCheckupTypeFieldInput struct {
	Name          *string
	FieldType     *string
	Unit          *string
	MinValue      *float64
	MaxValue      *float64
	Options       *[]CheckupFieldOptionInput
	SortOrder     *int
	ClearMinValue bool
	ClearMaxValue bool
}

// checkupTypeFinder は checkupTypeFieldService が親 checkup_type の存在/クリニック所有を
// 確認するための最小 view（service_deps.go の consumer-side view 規約に準拠）。
type checkupTypeFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error)
}

// CheckupTypeFieldService は健診パッケージのフィールド定義マスタの write 側を提供する。
type CheckupTypeFieldService interface {
	// CreateField は checkup_type 配下にフィールド定義を作成する（親不存在/他医院は NotFound）。
	CreateField(ctx context.Context, clinicID, checkupTypeID uint64, input *CreateCheckupTypeFieldInput) (*model.CheckupTypeField, error)
	// UpdateField はフィールド定義を部分更新する。field_type 変更も許可する
	// （checkup_field_results がスナップショットを持つため履歴は保全される）。
	UpdateField(ctx context.Context, clinicID, checkupTypeID, fieldID uint64, input *UpdateCheckupTypeFieldInput) (*model.CheckupTypeField, error)
	// DeleteField はフィールド定義をソフトデリートする（deleted_at）。定義行と既存結果行は残る。
	DeleteField(ctx context.Context, clinicID, checkupTypeID, fieldID uint64) error
	// ReorderFields は ids の並び順に sort_order を振り直す（単一 tx 内で原子的に実行）。
	// ids は対象スコープの生存フィールド id 集合と完全一致が必須（部分集合は InvalidInput、
	// スコープ外 id は NotFound — repository 側で検証）。
	ReorderFields(ctx context.Context, clinicID, checkupTypeID uint64, ids []uint64) error
}

type checkupTypeFieldService struct {
	fieldRepo    CheckupTypeFieldRepository
	checkupTypes checkupTypeFinder
	transactor   Transactor
}

// NewCheckupTypeFieldService は CheckupTypeFieldService の実装を返す。
func NewCheckupTypeFieldService(
	fieldRepo CheckupTypeFieldRepository,
	checkupTypes checkupTypeFinder,
	transactor Transactor,
) CheckupTypeFieldService {
	return &checkupTypeFieldService{
		fieldRepo:    fieldRepo,
		checkupTypes: checkupTypes,
		transactor:   transactor,
	}
}

// isCheckupSelectFieldType は options が必須の選択式フィールド型を返す
// （single_select / multi_select / checklist — 空 options だと FE 動的フォームが構築不能になる）。
func isCheckupSelectFieldType(t model.CheckupFieldType) bool {
	switch t {
	case model.CheckupFieldTypeSingleSelect, model.CheckupFieldTypeMultiSelect, model.CheckupFieldTypeChecklist:
		return true
	default:
		return false
	}
}

// validateCheckupFieldOptions は選択肢の各 value/label の非空と value 一意性を検証する。
func validateCheckupFieldOptions(options []CheckupFieldOptionInput) error {
	seen := make(map[string]struct{}, len(options))
	for _, opt := range options {
		if strings.TrimSpace(opt.Value) == "" {
			return apperrors.WrapInvalidInput("options の value は必須です")
		}
		if strings.TrimSpace(opt.Label) == "" {
			return apperrors.WrapInvalidInput("options の label は必須です")
		}
		if _, dup := seen[opt.Value]; dup {
			return apperrors.WrapInvalidInput("options の value は一意である必要があります")
		}
		seen[opt.Value] = struct{}{}
	}
	return nil
}

// validateCheckupFieldBounds は min_value / max_value の有限性と min<=max を検証する。
func validateCheckupFieldBounds(minValue, maxValue *float64) error {
	if minValue != nil && (math.IsNaN(*minValue) || math.IsInf(*minValue, 0)) {
		return apperrors.WrapInvalidInput("min_value は有限の数値で指定してください")
	}
	if maxValue != nil && (math.IsNaN(*maxValue) || math.IsInf(*maxValue, 0)) {
		return apperrors.WrapInvalidInput("max_value は有限の数値で指定してください")
	}
	if minValue != nil && maxValue != nil && *minValue > *maxValue {
		return apperrors.WrapInvalidInput("min_value は max_value 以下で指定してください")
	}
	return nil
}

// checkupFieldOptionJSON は options を options JSONB カラムへ書き込む際の wire 形。
// マニフェストで使う {value,label} 形に揃える。
type checkupFieldOptionJSON struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// marshalCheckupFieldOptionInputs は入力 options を JSONB 保存用にシリアライズする。
func marshalCheckupFieldOptionInputs(options []CheckupFieldOptionInput) (datatypes.JSON, error) {
	wire := make([]checkupFieldOptionJSON, 0, len(options))
	for _, opt := range options {
		wire = append(wire, checkupFieldOptionJSON(opt))
	}
	b, err := json.Marshal(wire)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("options を JSON に変換できません")
	}
	return datatypes.JSON(b), nil
}

// decodeCheckupFieldOptions は保存済み options JSONB を入力 DTO へ戻す（update 時の有効値判定用）。
func decodeCheckupFieldOptions(raw datatypes.JSON) ([]CheckupFieldOptionInput, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var wire []checkupFieldOptionJSON
	if err := json.Unmarshal(raw, &wire); err != nil {
		return nil, apperrors.WrapInternalServerError("保存済み options のデコードに失敗しました")
	}
	options := make([]CheckupFieldOptionInput, 0, len(wire))
	for _, opt := range wire {
		options = append(options, CheckupFieldOptionInput(opt))
	}
	return options, nil
}

func (s *checkupTypeFieldService) withTx(ctx context.Context, fn func(context.Context) error) error {
	// nil transactor は composition/wiring 失敗（サーバ内部）でありクライアント入力ではない → 500。
	if s.transactor == nil {
		return apperrors.WrapInternalServerError("checkup type field transaction dependency is required")
	}
	return s.transactor.WithTx(ctx, fn)
}

func (s *checkupTypeFieldService) CreateField(
	ctx context.Context,
	clinicID, checkupTypeID uint64,
	input *CreateCheckupTypeFieldInput,
) (*model.CheckupTypeField, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(errMsgInputNotNil)
	}
	fieldType := model.CheckupFieldType(input.FieldType)
	if !fieldType.IsValid() {
		return nil, apperrors.WrapInvalidInput("field_type は number / single_select / multi_select / boolean / checklist / text のいずれかで指定してください")
	}
	if err := validateRequiredName(input.Name); err != nil {
		return nil, err
	}
	if isCheckupSelectFieldType(fieldType) && len(input.Options) == 0 {
		return nil, apperrors.WrapInvalidInput("選択式フィールドには options を1件以上指定してください")
	}
	if !isCheckupSelectFieldType(fieldType) && len(input.Options) > 0 {
		return nil, apperrors.WrapInvalidInput("非選択式フィールドに options は指定できません")
	}
	if err := validateCheckupFieldOptions(input.Options); err != nil {
		return nil, err
	}
	if err := validateCheckupFieldBounds(input.MinValue, input.MaxValue); err != nil {
		return nil, err
	}
	optionsJSON, err := marshalCheckupFieldOptionInputs(input.Options)
	if err != nil {
		return nil, err
	}

	var field *model.CheckupTypeField
	err = s.withTx(ctx, func(txCtx context.Context) error {
		// 親 checkup_type の存在 + クリニック所有を tx 内で確認（他医院親へのフィールド作成は 404）。
		if _, err := s.checkupTypes.FindByID(txCtx, clinicID, checkupTypeID); err != nil {
			return err
		}
		created := &model.CheckupTypeField{
			ClinicID:      clinicID,
			CheckupTypeID: checkupTypeID,
			Name:          input.Name,
			FieldType:     fieldType,
			Unit:          input.Unit,
			MinValue:      input.MinValue,
			MaxValue:      input.MaxValue,
			Options:       optionsJSON,
			SortOrder:     input.SortOrder,
		}
		if err := s.fieldRepo.CreateField(txCtx, created); err != nil {
			return err
		}
		field = created
		return nil
	})
	return field, err
}

func (s *checkupTypeFieldService) UpdateField(
	ctx context.Context,
	clinicID, checkupTypeID, fieldID uint64,
	input *UpdateCheckupTypeFieldInput,
) (*model.CheckupTypeField, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(errMsgInputNotNil)
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, err
	}
	if input.FieldType != nil && !model.CheckupFieldType(*input.FieldType).IsValid() {
		return nil, apperrors.WrapInvalidInput("field_type は number / single_select / multi_select / boolean / checklist / text のいずれかで指定してください")
	}
	if input.Options != nil {
		if err := validateCheckupFieldOptions(*input.Options); err != nil {
			return nil, err
		}
	}

	var field *model.CheckupTypeField
	err := s.withTx(ctx, func(txCtx context.Context) error {
		existing, err := s.fieldRepo.LockFieldByID(txCtx, clinicID, checkupTypeID, fieldID)
		if err != nil {
			return err
		}

		// 有効値（既存値 ∪ 入力）で型依存の検証を行う。選択式への型変更は options の
		// 同時指定を要求し、定義なし選択式が残らないようにする。
		effectiveType := existing.FieldType
		if input.FieldType != nil {
			effectiveType = model.CheckupFieldType(*input.FieldType)
		}
		var effectiveOptions []CheckupFieldOptionInput
		if input.Options != nil {
			effectiveOptions = *input.Options
		} else {
			effectiveOptions, err = decodeCheckupFieldOptions(existing.Options)
			if err != nil {
				return err
			}
		}
		if isCheckupSelectFieldType(effectiveType) && len(effectiveOptions) == 0 {
			return apperrors.WrapInvalidInput("選択式フィールドには options を1件以上指定してください")
		}
		// 非選択式への options 指定は拒否（明示クリア options:[] は許可 — 型変更時の残存選択肢掃除用）。
		// input.Options==nil（未指定）時は既存値をそのまま維持する（有効値判定のみ）。
		if input.Options != nil && !isCheckupSelectFieldType(effectiveType) && len(*input.Options) > 0 {
			return apperrors.WrapInvalidInput("非選択式フィールドに options は指定できません")
		}

		effectiveMin := existing.MinValue
		if input.ClearMinValue {
			effectiveMin = nil
		} else if input.MinValue != nil {
			effectiveMin = input.MinValue
		}
		effectiveMax := existing.MaxValue
		if input.ClearMaxValue {
			effectiveMax = nil
		} else if input.MaxValue != nil {
			effectiveMax = input.MaxValue
		}
		if err := validateCheckupFieldBounds(effectiveMin, effectiveMax); err != nil {
			return err
		}

		fields := make(map[string]any)
		if input.Name != nil {
			fields["name"] = *input.Name
		}
		if input.FieldType != nil {
			fields["field_type"] = effectiveType
		}
		if input.Unit != nil {
			fields["unit"] = *input.Unit
		}
		if input.ClearMinValue {
			fields["min_value"] = nil
		} else if input.MinValue != nil {
			fields["min_value"] = *input.MinValue
		}
		if input.ClearMaxValue {
			fields["max_value"] = nil
		} else if input.MaxValue != nil {
			fields["max_value"] = *input.MaxValue
		}
		if input.Options != nil {
			optionsJSON, err := marshalCheckupFieldOptionInputs(*input.Options)
			if err != nil {
				return err
			}
			fields["options"] = optionsJSON
		}
		if input.SortOrder != nil {
			fields["sort_order"] = *input.SortOrder
		}
		if len(fields) == 0 {
			return apperrors.WrapInvalidInput(errMsgAtLeastOneField)
		}

		updated, err := s.fieldRepo.UpdateField(txCtx, clinicID, checkupTypeID, fieldID, fields)
		if err != nil {
			return err
		}
		field = updated
		return nil
	})
	return field, err
}

func (s *checkupTypeFieldService) DeleteField(ctx context.Context, clinicID, checkupTypeID, fieldID uint64) error {
	return s.withTx(ctx, func(txCtx context.Context) error {
		if _, err := s.fieldRepo.LockFieldByID(txCtx, clinicID, checkupTypeID, fieldID); err != nil {
			return err
		}
		// 定義行のみ soft delete。checkup_field_results（スナップショット保持 + FK ON DELETE
		// SET NULL）は一切触れない — 削除済みフィールドを参照する過去結果はそのまま残る。
		return s.fieldRepo.DeleteField(txCtx, clinicID, checkupTypeID, fieldID)
	})
}

func (s *checkupTypeFieldService) ReorderFields(
	ctx context.Context,
	clinicID, checkupTypeID uint64,
	ids []uint64,
) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(errMsgIDsNotEmpty)
	}
	seen := make(map[uint64]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			return apperrors.WrapInvalidInput("ids must be positive")
		}
		if _, ok := seen[id]; ok {
			return apperrors.WrapInvalidInput("ids must be unique")
		}
		seen[id] = struct{}{}
	}
	return s.withTx(ctx, func(txCtx context.Context) error {
		return s.fieldRepo.ReorderFields(txCtx, clinicID, checkupTypeID, ids)
	})
}
