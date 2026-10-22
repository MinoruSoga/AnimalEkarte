package medicalrecord

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-447 薬剤 base の default_quantity / tax_rate / strength の exact decimal wire。
// api.yaml Medicine と同じく、有限な非負 decimal 文字列だけを受け付ける。JSON number は旧 client の
// float 送信として 400 で拒否し、黙って変換しない。保存尺度で表せない値は丸めずに 400 にする。
// 省略と null はどちらも「値なし」（POST=既定、PATCH=保持）として扱い、null を新しいクリア手段にしない。
var medicineDecimalPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)(\.[0-9]+)?$`)

type medicineDecimalSpec struct {
	field         string
	maxIntDigits  int  // 整数部の最大桁数（0 = 制限なし）
	maxFracDigits int  // 末尾0を除いた小数部の最大桁数（-1 = 制限なし）
	positive      bool // 0 を拒否する
	atMostOne     bool // 1 を超える値を拒否する（比率）
}

var (
	// numeric(10,2)
	medicineDefaultQuantitySpec = medicineDecimalSpec{field: "default_quantity", maxIntDigits: 8, maxFracDigits: 2}
	// numeric(10,4)。既存 binding gt=0 を維持する。
	medicineStrengthSpec = medicineDecimalSpec{field: "strength", maxIntDigits: 6, maxFracDigits: 4, positive: true}
	// numeric。既存 binding min=0,max=1 を維持する（10% は "0.10"）。
	medicineTaxRateSpec = medicineDecimalSpec{field: "tax_rate", maxFracDigits: -1, atMostOne: true}
)

// parseMedicineDecimal は省略/null を nil、decimal 文字列を ExactNumeric へ厳密変換する（float を経由しない）。
func parseMedicineDecimal(raw json.RawMessage, spec medicineDecimalSpec) (*model.ExactNumeric, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil, nil
	}
	var s string
	if err := json.Unmarshal(trimmed, &s); err != nil {
		return nil, apperrors.WrapInvalidInput(spec.field + " は decimal 文字列で指定してください（JSON number は受け付けません）")
	}
	if !medicineDecimalPattern.MatchString(s) {
		return nil, apperrors.WrapInvalidInput(spec.field + " は符号・指数のない非負の decimal 文字列で指定してください")
	}
	intPart, frac, _ := strings.Cut(s, ".")
	frac = strings.TrimRight(frac, "0")
	if spec.maxIntDigits > 0 && len(intPart) > spec.maxIntDigits {
		return nil, apperrors.WrapInvalidInput(spec.field + " の整数部が保存できる桁数を超えています")
	}
	if spec.maxFracDigits >= 0 && len(frac) > spec.maxFracDigits {
		return nil, apperrors.WrapInvalidInput(spec.field + " は小数第" + strconv.Itoa(spec.maxFracDigits) + "位までで指定してください")
	}
	isZero := intPart == "0" && frac == ""
	if spec.positive && isZero {
		return nil, apperrors.WrapInvalidInput(spec.field + " は正の値である必要があります")
	}
	if spec.atMostOne && intPart != "0" && (intPart != "1" || frac != "") {
		return nil, apperrors.WrapInvalidInput(spec.field + " は 0 以上 1 以下で指定してください")
	}
	var n model.ExactNumeric
	if err := n.Scan(s); err != nil {
		return nil, apperrors.WrapInvalidInput(spec.field + " の形式が不正です")
	}
	return &n, nil
}

// exactNumericStringPtr は nullable な保存値を canonical decimal 文字列へ写す（NULL は nil）。
func exactNumericStringPtr(n *model.ExactNumeric) *string {
	if n == nil || !n.Valid {
		return nil
	}
	s := exactNumericCanonical(*n)
	return &s
}

// medicineStrengthFloat は製品含量を用量計算器・設定検証器の float64 入力へ明示変換する。
// 保存・HTTP は ExactNumeric のまま扱い、計算式（dose_calc）は本票で変更しない。
// 変換できない値は nil（含量なし）として返し、per_weight の検証・計算を fail-closed にする。
func medicineStrengthFloat(n *model.ExactNumeric) *float64 {
	if n == nil || !n.Valid {
		return nil
	}
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return nil
	}
	v := f.Float64
	return &v
}

// defaultMedicineTaxRate は POST 省略/null 時の既存既定税率（0.10）を ExactNumeric で返す。
func defaultMedicineTaxRate() (model.ExactNumeric, error) {
	var n model.ExactNumeric
	if err := n.Scan(strconv.FormatFloat(DefaultTaxRate, 'f', -1, 64)); err != nil {
		return n, apperrors.Wrap(err, "failed to build default tax rate")
	}
	return n, nil
}
