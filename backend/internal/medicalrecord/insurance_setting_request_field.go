package medicalrecord

import (
	"encoding/json"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// EMR-454: 品目別保険設定 is_non_insurance の update DTO 用 wire メッセージ。
const (
	errMsgInsuranceSettingNullNotAllowed = "is_non_insurance に null は指定できません"
	errMsgInsuranceSettingInvalid        = "is_non_insurance は true または false で指定してください"
)

// boolUpdateField は PATCH DTO で「未送信」と「明示 true/false」を区別する。
// Go 標準の *bool だけでは encoding/json が null と未送信をどちらも nil にするため、
// set フラグでフィールド存在を保持し、service 層の *bool（nil = 変更なし）へ変換する。
//
// api.yaml の wire 契約（EMR-454）どおり、update の is_non_insurance は
// 省略=変更なし、null=400、true/false=明示更新として扱う。
// 旧欠落（DB NULL）を true/false へ読み替えないための区別であり、
// 明示 null を「変更なし」へ黙って落とさない。
type boolUpdateField struct {
	set   bool
	value bool
}

func (f *boolUpdateField) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return apperrors.WrapInvalidInput(errMsgInsuranceSettingNullNotAllowed)
	}
	var value bool
	if err := json.Unmarshal(data, &value); err != nil {
		return apperrors.WrapInvalidInput(errMsgInsuranceSettingInvalid)
	}
	f.set = true
	f.value = value
	return nil
}

// toServiceInput は未送信を nil（変更なし）に変換する。
func (f boolUpdateField) toServiceInput() *bool {
	if !f.set {
		return nil
	}
	value := f.value
	return &value
}
