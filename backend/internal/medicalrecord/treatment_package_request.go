package medicalrecord

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-521 package master DTO。wire 契約（api.yaml TreatmentPackage*）どおり
// ID/version は canonical signed-64 十進「文字列」、quantity は ExactNumeric 正数文字列、
// 円は非負整数文字列。JSON number や非 canonical 形は拒否し黙って丸めない。
// unknown key は拒否する（additionalProperties:false 相当を DTO 層で強制）。
var (
	treatmentPackageIDPattern = regexp.MustCompile(`^[1-9][0-9]*$`)
	// 正の canonical decimal: 先頭0なし整数部、末尾ゼロなし小数部、指数・符号・空白不可。
	treatmentPackageQuantityPattern = regexp.MustCompile(`^([1-9][0-9]*(\.[0-9]*[1-9])?|0\.[0-9]*[1-9])$`)
	// canonical decimal（tax_rate 用、負値可）。-0 は拒否。
	treatmentPackageDecimalPattern = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]*[1-9])?$`)
	// 非負整数円文字列。
	treatmentPackageYenPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)$`)
)

func parsePackageCanonicalInt64(raw string, field string) (int64, error) {
	if !treatmentPackageIDPattern.MatchString(raw) {
		return 0, apperrors.WrapInvalidInput(field + " は正の十進文字列で指定してください")
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, apperrors.WrapInvalidInput(field + " が範囲外です")
	}
	return v, nil
}

func parsePackageCanonicalID(raw json.RawMessage, field string) (uint64, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return 0, apperrors.WrapInvalidInput(field + " は canonical 十進文字列で指定してください")
	}
	v, err := parsePackageCanonicalInt64(s, field)
	if err != nil {
		return 0, err
	}
	return uint64(v), nil
}

// parsePackageQuantity は canonical 正数 decimal 文字列を ExactNumeric へ厳密変換する。
func parsePackageQuantity(raw json.RawMessage) (model.ExactNumeric, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return model.ExactNumeric{}, apperrors.WrapInvalidInput("quantity は正の decimal 文字列で指定してください")
	}
	var q model.ExactNumeric
	if !treatmentPackageQuantityPattern.MatchString(s) {
		return q, apperrors.WrapInvalidInput("quantity は正の canonical decimal 文字列で指定してください")
	}
	if err := q.Scan(s); err != nil {
		return q, apperrors.WrapInvalidInput("quantity の形式が不正です")
	}
	return q, nil
}

func parsePackageCanonicalDecimal(raw json.RawMessage, field string) (model.ExactNumeric, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return model.ExactNumeric{}, apperrors.WrapInvalidInput(field + " は decimal 文字列で指定してください")
	}
	var d model.ExactNumeric
	if s == "-0" || !treatmentPackageDecimalPattern.MatchString(s) {
		return d, apperrors.WrapInvalidInput(field + " は canonical decimal 文字列で指定してください")
	}
	if err := d.Scan(s); err != nil {
		return d, apperrors.WrapInvalidInput(field + " の形式が不正です")
	}
	return d, nil
}

func parsePackageCanonicalYen(raw json.RawMessage, field string) (int64, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return 0, apperrors.WrapInvalidInput(field + " は非負整数文字列で指定してください")
	}
	if !treatmentPackageYenPattern.MatchString(s) {
		return 0, apperrors.WrapInvalidInput(field + " は非負の整数文字列で指定してください")
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, apperrors.WrapInvalidInput(field + " が範囲外です")
	}
	return v, nil
}

// parsePackageUUID は canonical lowercase hyphenated non-nil UUID のみ受理する。
// 正規化せず非 canonical 形は拒否する。
func parsePackageUUID(raw json.RawMessage, field string) (uuid.UUID, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return uuid.Nil, apperrors.WrapInvalidInput(field + " は UUID 文字列で指定してください")
	}
	u, err := uuid.Parse(s)
	if err != nil || u == uuid.Nil || u.String() != s {
		return uuid.Nil, apperrors.WrapInvalidInput(field + " は canonical な小文字ハイフン区切り UUID で指定してください")
	}
	return u, nil
}

func rejectPackageNull(raw json.RawMessage, field string) error {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return apperrors.WrapInvalidInput(field + " に null は指定できません")
	}
	return nil
}

func rejectUnknownPackageKeys(raw map[string]json.RawMessage, allowed ...string) error {
	for k := range raw {
		found := false
		for _, a := range allowed {
			if k == a {
				found = true
				break
			}
		}
		if !found {
			return apperrors.WrapInvalidInput("未知のフィールドです: " + k)
		}
	}
	return nil
}

func strictPackageDecode(raw json.RawMessage, v any, field string) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return apperrors.WrapInvalidInput(field + " の形式が不正です")
	}
	return nil
}

// TreatmentPackageManualDefinition は manual_other の正規 manual schema。
// name/reason は trim 後非空白必須。欠落・無効 master への fallback には使わない。
type TreatmentPackageManualDefinition struct {
	Name           string
	Reason         string
	UnitLabel      *string
	UnitPrice      int64
	TaxType        string
	TaxRate        model.ExactNumeric
	TaxRateText    string
	IsNonInsurance bool
	// Raw は検証済み入力のそのままの JSON（manual_definition 保存用）。
	Raw json.RawMessage
}

func (m *TreatmentPackageManualDefinition) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return apperrors.WrapInvalidInput("manual は object で指定してください")
	}
	if err := rejectUnknownPackageKeys(raw,
		"name", "reason", "unit_label", "unit_price", "tax_type", "tax_rate", "is_non_insurance"); err != nil {
		return err
	}
	for _, k := range []string{"name", "reason", "unit_label", "unit_price", "tax_type", "tax_rate", "is_non_insurance"} {
		if _, ok := raw[k]; !ok {
			return apperrors.WrapInvalidInput("manual." + k + " は必須です")
		}
	}
	for _, k := range []string{"name", "reason", "unit_price", "tax_type", "tax_rate", "is_non_insurance"} {
		if err := rejectPackageNull(raw[k], "manual."+k); err != nil {
			return err
		}
	}
	var name, reason string
	if err := json.Unmarshal(raw["name"], &name); err != nil || strings.TrimSpace(name) == "" {
		return apperrors.WrapInvalidInput("manual.name は非空白文字列で指定してください")
	}
	if err := json.Unmarshal(raw["reason"], &reason); err != nil || strings.TrimSpace(reason) == "" {
		return apperrors.WrapInvalidInput("manual.reason は非空白文字列で指定してください")
	}
	var unitLabel *string
	if !bytes.Equal(bytes.TrimSpace(raw["unit_label"]), []byte("null")) {
		var s string
		if err := json.Unmarshal(raw["unit_label"], &s); err != nil {
			return apperrors.WrapInvalidInput("manual.unit_label の形式が不正です")
		}
		unitLabel = &s
	}
	price, err := parsePackageCanonicalYen(raw["unit_price"], "manual.unit_price")
	if err != nil {
		return err
	}
	var taxType string
	if err := json.Unmarshal(raw["tax_type"], &taxType); err != nil {
		return apperrors.WrapInvalidInput("manual.tax_type の形式が不正です")
	}
	if taxType != string(model.TaxTypeIncluded) && taxType != string(model.TaxTypeExcluded) && taxType != string(model.TaxTypeExempt) {
		return apperrors.WrapInvalidInput("manual.tax_type の値が不正です")
	}
	var taxRateText string
	if err := json.Unmarshal(raw["tax_rate"], &taxRateText); err != nil {
		return apperrors.WrapInvalidInput("manual.tax_rate は decimal 文字列で指定してください")
	}
	taxRate, err := parsePackageCanonicalDecimal(raw["tax_rate"], "manual.tax_rate")
	if err != nil {
		return err
	}
	var nonInsurance bool
	if err := json.Unmarshal(raw["is_non_insurance"], &nonInsurance); err != nil {
		return apperrors.WrapInvalidInput("manual.is_non_insurance は true または false で指定してください")
	}
	m.Name, m.Reason = name, reason
	m.UnitLabel, m.UnitPrice = unitLabel, price
	m.TaxType, m.TaxRate, m.TaxRateText = taxType, taxRate, taxRateText
	m.IsNonInsurance = nonInsurance
	m.Raw = append(json.RawMessage(nil), data...)
	return nil
}

// treatmentPackageItemRef は discriminator "kind" 付き 8 分岐の typed 参照。
// 各 branch は該当 FK キー 1 個だけを持ち、他型 FK 混入・unknown key を拒否する。
// manual_other は FK 0 個 + 正規 manual schema 必須。
type treatmentPackageItemRef struct {
	Kind    string
	TypedID uint64 // manual_other のとき 0
	Manual  *TreatmentPackageManualDefinition
}

var packageItemKindFKField = map[string]string{
	model.TreatmentPackageItemKindConsultation:    "consultation_id",
	model.TreatmentPackageItemKindProcedure:       "procedure_id",
	model.TreatmentPackageItemKindMedicine:        "medicine_id",
	model.TreatmentPackageItemKindExaminationType: "examination_type_id",
	model.TreatmentPackageItemKindVaccine:         "vaccine_id",
	model.TreatmentPackageItemKindCheckupType:     "checkup_type_id",
	model.TreatmentPackageItemKindMerchandiseItem: "merchandise_item_id",
}

func (r *treatmentPackageItemRef) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return apperrors.WrapInvalidInput("item は object で指定してください")
	}
	kindRaw, ok := raw["kind"]
	if !ok {
		return apperrors.WrapInvalidInput("item.kind は必須です")
	}
	var kind string
	if err := json.Unmarshal(kindRaw, &kind); err != nil {
		return apperrors.WrapInvalidInput("item.kind の形式が不正です")
	}
	if kind == model.TreatmentPackageItemKindManualOther {
		if err := rejectUnknownPackageKeys(raw, "kind", "manual"); err != nil {
			return err
		}
		manualRaw, ok := raw["manual"]
		if !ok {
			return apperrors.WrapInvalidInput("manual_other には manual が必須です")
		}
		var manual TreatmentPackageManualDefinition
		if err := json.Unmarshal(manualRaw, &manual); err != nil {
			return err
		}
		r.Kind, r.TypedID, r.Manual = kind, 0, &manual
		return nil
	}
	fkField, ok := packageItemKindFKField[kind]
	if !ok {
		return apperrors.WrapInvalidInput("item.kind の値が不正です")
	}
	if err := rejectUnknownPackageKeys(raw, "kind", fkField); err != nil {
		return err
	}
	idRaw, ok := raw[fkField]
	if !ok {
		return apperrors.WrapInvalidInput("item." + fkField + " は必須です")
	}
	id, err := parsePackageCanonicalID(idRaw, "item."+fkField)
	if err != nil {
		return err
	}
	r.Kind, r.TypedID, r.Manual = kind, id, nil
	return nil
}

// TreatmentPackageComponentInput は service 層への構成入力。
// ComponentKey は client 供給の canonical UUID で編集/並替間で安定。
type TreatmentPackageComponentInput struct {
	ComponentKey uuid.UUID
	Kind         string
	TypedID      uint64
	Manual       *TreatmentPackageManualDefinition
	Quantity     model.ExactNumeric
	QuantityText string
}

type treatmentPackageComponentInputJSON struct {
	ComponentKey uuid.UUID
	Item         treatmentPackageItemRef
	Quantity     model.ExactNumeric
	QuantityText string
}

func (c *treatmentPackageComponentInputJSON) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return apperrors.WrapInvalidInput("component は object で指定してください")
	}
	if err := rejectUnknownPackageKeys(raw, "component_key", "item", "quantity"); err != nil {
		return err
	}
	for _, k := range []string{"component_key", "item", "quantity"} {
		v, ok := raw[k]
		if !ok {
			return apperrors.WrapInvalidInput("component." + k + " は必須です")
		}
		if err := rejectPackageNull(v, "component."+k); err != nil {
			return err
		}
	}
	key, err := parsePackageUUID(raw["component_key"], "component.component_key")
	if err != nil {
		return err
	}
	var item treatmentPackageItemRef
	if err := json.Unmarshal(raw["item"], &item); err != nil {
		return err
	}
	var qtyText string
	if err := json.Unmarshal(raw["quantity"], &qtyText); err != nil {
		return apperrors.WrapInvalidInput("component.quantity は正の decimal 文字列で指定してください")
	}
	qty, err := parsePackageQuantity(raw["quantity"])
	if err != nil {
		return err
	}
	c.ComponentKey = key
	c.Item = item
	c.Quantity = qty
	c.QuantityText = qtyText
	return nil
}

type createTreatmentPackageRequest struct {
	Name        string
	Description string
	IsActive    bool
	Components  []TreatmentPackageComponentInput
}

// UnmarshalJSON は TreatmentPackageCreate の required/additionalProperties:false と
// 非 nullable field の明示 null 拒否を DTO 層で強制する。
func (r *createTreatmentPackageRequest) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return apperrors.WrapInvalidInput("request body は object で指定してください")
	}
	if err := rejectUnknownPackageKeys(raw, "name", "description", "is_active", "components"); err != nil {
		return err
	}
	for _, k := range []string{"name", "description", "is_active", "components"} {
		v, ok := raw[k]
		if !ok {
			return apperrors.WrapInvalidInput(k + " は必須です")
		}
		if err := rejectPackageNull(v, k); err != nil {
			return err
		}
	}
	if err := json.Unmarshal(raw["name"], &r.Name); err != nil || strings.TrimSpace(r.Name) == "" || len(r.Name) > 255 {
		return apperrors.WrapInvalidInput("name は非空白の255文字以内で指定してください")
	}
	if err := json.Unmarshal(raw["description"], &r.Description); err != nil {
		return apperrors.WrapInvalidInput("description の形式が不正です")
	}
	if err := json.Unmarshal(raw["is_active"], &r.IsActive); err != nil {
		return apperrors.WrapInvalidInput("is_active は true または false で指定してください")
	}
	components, err := decodePackageComponents(raw["components"])
	if err != nil {
		return err
	}
	r.Components = components
	return nil
}

func decodePackageComponents(raw json.RawMessage) ([]TreatmentPackageComponentInput, error) {
	var elements []json.RawMessage
	if err := json.Unmarshal(raw, &elements); err != nil {
		return nil, apperrors.WrapInvalidInput("components は array で指定してください")
	}
	if len(elements) < 1 {
		return nil, apperrors.WrapInvalidInput("components は1件以上必要です")
	}
	seen := make(map[uuid.UUID]struct{}, len(elements))
	out := make([]TreatmentPackageComponentInput, 0, len(elements))
	for _, el := range elements {
		var c treatmentPackageComponentInputJSON
		if err := json.Unmarshal(el, &c); err != nil {
			return nil, err
		}
		if _, dup := seen[c.ComponentKey]; dup {
			return nil, apperrors.WrapInvalidInput("component_key が重複しています")
		}
		seen[c.ComponentKey] = struct{}{}
		out = append(out, TreatmentPackageComponentInput{
			ComponentKey: c.ComponentKey,
			Kind:         c.Item.Kind,
			TypedID:      c.Item.TypedID,
			Manual:       c.Item.Manual,
			Quantity:     c.Quantity,
			QuantityText: c.QuantityText,
		})
	}
	return out, nil
}

type updateTreatmentPackageRequest struct {
	Version     int64
	Name        *string
	Description *string
	IsActive    *bool
	Components  *[]TreatmentPackageComponentInput
}

// UnmarshalJSON は省略=維持と非 nullable field の明示 null 拒否を区別する。
// version は必須の canonical version 文字列。
func (r *updateTreatmentPackageRequest) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return apperrors.WrapInvalidInput("request body は object で指定してください")
	}
	if err := rejectUnknownPackageKeys(raw, "version", "name", "description", "is_active", "components"); err != nil {
		return err
	}
	vRaw, ok := raw["version"]
	if !ok {
		return apperrors.WrapInvalidInput("version は必須です")
	}
	var vText string
	if err := json.Unmarshal(vRaw, &vText); err != nil {
		return apperrors.WrapInvalidInput("version は canonical 十進文字列で指定してください")
	}
	version, err := parsePackageCanonicalInt64(vText, "version")
	if err != nil {
		return err
	}
	r.Version = version
	if v, ok := raw["name"]; ok {
		if err := rejectPackageNull(v, "name"); err != nil {
			return err
		}
		var s string
		if err := json.Unmarshal(v, &s); err != nil || strings.TrimSpace(s) == "" || len(s) > 255 {
			return apperrors.WrapInvalidInput("name は非空白の255文字以内で指定してください")
		}
		r.Name = &s
	}
	if v, ok := raw["description"]; ok {
		if err := rejectPackageNull(v, "description"); err != nil {
			return err
		}
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			return apperrors.WrapInvalidInput("description の形式が不正です")
		}
		r.Description = &s
	}
	if v, ok := raw["is_active"]; ok {
		if err := rejectPackageNull(v, "is_active"); err != nil {
			return err
		}
		var b bool
		if err := json.Unmarshal(v, &b); err != nil {
			return apperrors.WrapInvalidInput("is_active は true または false で指定してください")
		}
		r.IsActive = &b
	}
	if v, ok := raw["components"]; ok {
		if err := rejectPackageNull(v, "components"); err != nil {
			return err
		}
		components, err := decodePackageComponents(v)
		if err != nil {
			return err
		}
		r.Components = &components
	}
	return nil
}

type reorderTreatmentPackagesRequest struct {
	OrderVersion int64
	IDs          []uint64
}

// UnmarshalJSON は order_version 必須・ids の canonical ID 文字列配列・
// 重複拒否（uniqueItems 相当）を DTO 層で強制する。
func (r *reorderTreatmentPackagesRequest) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return apperrors.WrapInvalidInput("request body は object で指定してください")
	}
	if err := rejectUnknownPackageKeys(raw, "order_version", "ids"); err != nil {
		return err
	}
	for _, k := range []string{"order_version", "ids"} {
		v, ok := raw[k]
		if !ok {
			return apperrors.WrapInvalidInput(k + " は必須です")
		}
		if err := rejectPackageNull(v, k); err != nil {
			return err
		}
	}
	var ovText string
	if err := json.Unmarshal(raw["order_version"], &ovText); err != nil {
		return apperrors.WrapInvalidInput("order_version は canonical 十進文字列で指定してください")
	}
	ov, err := parsePackageCanonicalInt64(ovText, "order_version")
	if err != nil {
		return err
	}
	var elements []json.RawMessage
	if err := json.Unmarshal(raw["ids"], &elements); err != nil {
		return apperrors.WrapInvalidInput("ids は array で指定してください")
	}
	seen := make(map[uint64]struct{}, len(elements))
	ids := make([]uint64, 0, len(elements))
	for _, el := range elements {
		id, err := parsePackageCanonicalID(el, "ids[]")
		if err != nil {
			return err
		}
		if _, dup := seen[id]; dup {
			return nil, apperrors.WrapInvalidInput("ids に重複があります")
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	r.OrderVersion = ov
	r.IDs = ids
	return nil
}
