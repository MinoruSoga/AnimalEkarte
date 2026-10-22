package inventory

import (
	"encoding/json"
	"math/big"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
)

// merchandiseTaxRateDTO は物販 tax_rate の wire 契約（EMR-530 / 採択405・483の十進輸送）。
// api.yaml MerchandiseItem.tax_rate の定義に従い、比率を表す有限な非負 exact decimal 文字列のみ受理する。
// 10% は "0.10"（数値として0.1）で、百分率10へ変換しない。符号・指数表記・NaN/Infinity・
// JSON number は、旧clientの黙ったfloat変換を防ぐため 400 で拒否する。
// 入力範囲 0..1 は service 層で exact 比較により検査する。
//
// create: 省略/null → 既存既定 0.10。update: 省略/null → 保持。明示 "0" は 0 として保存し既定へ潰さない。
type merchandiseTaxRateDTO struct {
	set   bool
	value model.ExactNumeric
}

const (
	merchandiseTaxRateErrMsgNotString = "tax_rate には10進数の文字列を指定してください（数値型は指定できません）"
	merchandiseTaxRateErrMsgInvalid   = "tax_rate には有限な非負の10進数文字列を指定してください（例: \"0.10\"。符号・指数表記は使えません）"
	merchandiseTaxRateErrMsgRange     = "tax_rate は0以上1以下で指定してください"
	merchandiseTaxRateErrMsgScale     = "tax_rate の小数部が長すぎます（PostgreSQL numeric の scale 上限 16383 を超える値は保存できません）"
)

// maxTaxRateScale は PostgreSQL numeric の scale 上限（docs: CREATE TYPE numeric 精度制限）。
const maxTaxRateScale = 16383

// UnmarshalJSON は tax_rate の wire token を検査して ExactNumeric へ変換する。
func (d *merchandiseTaxRateDTO) UnmarshalJSON(data []byte) error {
	// null は「未指定」扱い（旧 *float64 wire で null が nil として扱われていた既定規則を維持）。
	if string(data) == "null" {
		*d = merchandiseTaxRateDTO{}
		return nil
	}
	if len(data) == 0 || data[0] != '"' {
		return apperrors.WrapInvalidInput(merchandiseTaxRateErrMsgNotString)
	}
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return apperrors.WrapInvalidInput(merchandiseTaxRateErrMsgNotString)
	}
	value, err := parseMerchandiseTaxRate(text)
	if err != nil {
		return err
	}
	*d = merchandiseTaxRateDTO{set: true, value: value}
	return nil
}

// optional は create 用。未指定の場合は nil（service 層の既定 0.10 適用）を返す。
func (d merchandiseTaxRateDTO) optional() *model.ExactNumeric {
	if !d.set {
		return nil
	}
	value := d.value
	return &value
}

// defaultMerchandiseTaxRate は既存既定税率 0.10 の exact 表現（sharedkernel.DefaultTaxRate の
// float64 定数は既存 consumer があるため触らず、本契約の入口だけ exact literal を使う）。
func defaultMerchandiseTaxRate() model.ExactNumeric {
	return model.ExactNumeric{pgtype.Numeric{Int: big.NewInt(10), Exp: -2, Valid: true}}
}

// parseMerchandiseTaxRate は有限な非負 exact decimal 文字列（api.yaml pattern
// ^(0|[1-9][0-9]*)(\.[0-9]+)?$）を無損失で ExactNumeric へ変換する。
// 税率列に固定 scale は無いため精度をそのまま保持するが、PostgreSQL numeric の
// scale 上限（16383）を超える字句は保存できないため入口で拒否する。
func parseMerchandiseTaxRate(text string) (model.ExactNumeric, error) {
	intPart, fracPart, found := strings.Cut(text, ".")
	if !validTaxRateDigits(intPart) || (found && (fracPart == "" || !allTaxRateDigits(fracPart))) {
		return model.ExactNumeric{}, apperrors.WrapInvalidInput(merchandiseTaxRateErrMsgInvalid)
	}
	if len(fracPart) > maxTaxRateScale {
		return model.ExactNumeric{}, apperrors.WrapInvalidInput(merchandiseTaxRateErrMsgScale)
	}
	digits := intPart + fracPart
	intValue, ok := new(big.Int).SetString(digits, 10)
	if !ok {
		return model.ExactNumeric{}, apperrors.WrapInvalidInput(merchandiseTaxRateErrMsgInvalid)
	}
	if intValue.Sign() == 0 {
		// 明示 0 は 0 として保存する（既定値へ潰さない）。
		return model.ExactNumeric{pgtype.Numeric{Int: big.NewInt(0), Exp: 0, Valid: true}}, nil
	}
	exp := -len(fracPart)
	return model.ExactNumeric{pgtype.Numeric{Int: intValue, Exp: int32(exp), Valid: true}}, nil
}

// validateMerchandiseTaxRate は tax_rate の入力範囲 0..1 を exact 比較で検査する（float 経由なし）。
// 無効値（NULL/NaN/無限大）も 0 として黙らず拒否する。
func validateMerchandiseTaxRate(value model.ExactNumeric) error {
	if !value.Valid || value.NaN || value.InfinityModifier != pgtype.Finite {
		return apperrors.WrapInvalidInput(merchandiseTaxRateErrMsgInvalid)
	}
	v := exactTaxRateRat(value)
	if v.Sign() < 0 || v.Cmp(big.NewRat(1, 1)) > 0 {
		return apperrors.WrapInvalidInput(merchandiseTaxRateErrMsgRange)
	}
	return nil
}

// validTaxRateDigits は api.yaml pattern の整数部（"0" または先頭0無しの正整数）を検査する。
func validTaxRateDigits(text string) bool {
	if text == "" || !allTaxRateDigits(text) {
		return false
	}
	return text == "0" || text[0] != '0'
}

func allTaxRateDigits(text string) bool {
	for i := 0; i < len(text); i++ {
		if text[i] < '0' || text[i] > '9' {
			return false
		}
	}
	return true
}

// exactTaxRateRat は ExactNumeric を big.Rat へ無損失変換する（比較用。丸め・float不使用）。
func exactTaxRateRat(n model.ExactNumeric) *big.Rat {
	if !n.Valid || n.Int == nil {
		return new(big.Rat)
	}
	num := new(big.Int).Set(n.Int)
	if n.Exp >= 0 {
		num.Mul(num, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n.Exp)), nil))
		return new(big.Rat).SetInt(num)
	}
	den := new(big.Int).Exp(big.NewInt(10), big.NewInt(-int64(n.Exp)), nil)
	return new(big.Rat).SetFrac(num, den)
}
