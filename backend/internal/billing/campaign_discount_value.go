package billing

import (
	"encoding/json"
	"math/big"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
)

// campaignDiscountValueDTO は campaign discount_value の wire 契約（EMR-530 / 採択483の十進輸送）。
// api.yaml Campaign.discount_value の定義に従い、有限な非負 exact decimal 文字列のみ受理する。
// 符号・指数表記・NaN/Infinity・JSON number は、旧clientの黙ったfloat変換を防ぐため 400 で拒否する。
// numeric(12,2) で表せない値（小数第3位以下・整数部10桁超）は丸めず 400 とする。
//
// create: 省略/null → 0（旧 float64 wire と同じ既定）。update: 省略/null → 保持（patch 既定規則は変更しない）。
type campaignDiscountValueDTO struct {
	set   bool
	value model.ExactNumeric
}

const (
	campaignDiscountErrMsgNotString = "discount_value には10進数の文字列を指定してください（数値型は指定できません）"
	campaignDiscountErrMsgInvalid   = "discount_value には有限な非負の10進数文字列を指定してください（例: \"10.5\"。符号・指数表記は使えません）"
	campaignDiscountErrMsgScale     = "discount_value は numeric(12,2) で保存するため、小数第3位以下と整数部11桁以上の値は指定できません"
)

// UnmarshalJSON は discount_value の wire token を検査して ExactNumeric へ変換する。
func (d *campaignDiscountValueDTO) UnmarshalJSON(data []byte) error {
	// null は「未指定」扱い（旧 float64 wire で null が 0 / nil として扱われていた既定規則を維持）。
	if string(data) == "null" {
		*d = campaignDiscountValueDTO{}
		return nil
	}
	if len(data) == 0 || data[0] != '"' {
		return apperrors.WrapInvalidInput(campaignDiscountErrMsgNotString)
	}
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return apperrors.WrapInvalidInput(campaignDiscountErrMsgNotString)
	}
	value, err := parseCampaignDiscountValue(text)
	if err != nil {
		return err
	}
	*d = campaignDiscountValueDTO{set: true, value: value}
	return nil
}

// orZero は create 用。未指定の場合も旧 wire と同じく 0 として保存する（列は NOT NULL default 0）。
func (d campaignDiscountValueDTO) orZero() model.ExactNumeric {
	if !d.set {
		return zeroExactNumeric()
	}
	return d.value
}

func zeroExactNumeric() model.ExactNumeric {
	return model.ExactNumeric{pgtype.Numeric{Int: big.NewInt(0), Exp: 0, Valid: true}}
}

// parseCampaignDiscountValue は有限な非負 exact decimal 文字列（api.yaml pattern
// ^(0|[1-9][0-9]*)(\.[0-9]+)?$）を numeric(12,2) 表現可能範囲で ExactNumeric へ変換する。
func parseCampaignDiscountValue(text string) (model.ExactNumeric, error) {
	intPart, fracPart, found := strings.Cut(text, ".")
	if !validDecimalDigits(intPart) || (found && !validFractionDigits(fracPart)) {
		return model.ExactNumeric{}, apperrors.WrapInvalidInput(campaignDiscountErrMsgInvalid)
	}
	if len(intPart) > 10 || len(fracPart) > 2 {
		return model.ExactNumeric{}, apperrors.WrapInvalidInput(campaignDiscountErrMsgScale)
	}
	digits := intPart + fracPart
	intValue, ok := new(big.Int).SetString(digits, 10)
	if !ok {
		return model.ExactNumeric{}, apperrors.WrapInvalidInput(campaignDiscountErrMsgInvalid)
	}
	exp := -len(fracPart)
	if intValue.Sign() == 0 {
		// "0" / "0.0" 等は明示0として保持する（既定値への潰しは行わない）。
		return zeroExactNumeric(), nil
	}
	return model.ExactNumeric{pgtype.Numeric{Int: intValue, Exp: int32(exp), Valid: true}}, nil
}

// validDecimalDigits は api.yaml pattern の整数部（"0" または先頭0無しの正整数）を検査する。
func validDecimalDigits(text string) bool {
	if text == "" {
		return false
	}
	if !allDigits(text) {
		return false
	}
	return text == "0" || text[0] != '0'
}

// validFractionDigits は api.yaml pattern の小数部（\.[0-9]+）を検査する。空は不許可。
func validFractionDigits(text string) bool {
	return text != "" && allDigits(text)
}

func allDigits(text string) bool {
	for i := 0; i < len(text); i++ {
		if text[i] < '0' || text[i] > '9' {
			return false
		}
	}
	return true
}

// exactNumericRat は ExactNumeric を big.Rat へ無損失変換する（比較用。丸め・float不使用）。
func exactNumericRat(n model.ExactNumeric) *big.Rat {
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
