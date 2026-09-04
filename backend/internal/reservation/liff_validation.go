package reservation

import (
	"encoding/json"
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// BUG-LINE-012: LIFF 予約作成入力のサイズ制限定数。
// 数値は業務上の妥当性と DoS 防御のバランスで設定する。
const (
	// maxCustomerFieldsSize は customer_fields JSON 全体の最大バイト数。
	maxCustomerFieldsSize = 10 * 1024 // 10KB
	// maxCustomerFieldsKeys は customer_fields の最大キー数（トップレベル）。
	maxCustomerFieldsKeys = 20
	// maxCustomerFieldValueLength は各 string 値の最大文字数。
	maxCustomerFieldValueLength = 500
	// maxRequestTextLength は request_text の最大文字数。
	maxRequestTextLength = 1000
)

// validateLiffReservationInput は LIFF 予約作成リクエストのサイズ制限を検証する。
//   - customer_fields の JSON 全体サイズ
//   - customer_fields のトップレベルキー数
//   - customer_fields の各 string 値の長さ（ネスト配列内の string も対象）
//   - request_text の長さ
func validateLiffReservationInput(req *liffCreateReservationRequest) error {
	// request_text 長さ
	if l := len([]rune(req.RequestText)); l > maxRequestTextLength {
		return apperrors.WrapInvalidInput(fmt.Sprintf("request_text は %d 文字以内で入力してください", maxRequestTextLength))
	}

	// customer_fields 全体サイズ
	if len(req.CustomerFields) == 0 {
		return nil
	}
	if len(req.CustomerFields) > maxCustomerFieldsSize {
		return apperrors.WrapInvalidInput(fmt.Sprintf("customer_fields のサイズが大きすぎます（上限 %d バイト）", maxCustomerFieldsSize))
	}

	// customer_fields のパースと検証
	var fields map[string]any
	if err := json.Unmarshal(req.CustomerFields, &fields); err != nil {
		return apperrors.WrapInvalidInput("customer_fields は JSON オブジェクトで指定してください")
	}
	if len(fields) > maxCustomerFieldsKeys {
		return apperrors.WrapInvalidInput(fmt.Sprintf("customer_fields のキー数が多すぎます（上限 %d 個）", maxCustomerFieldsKeys))
	}
	for key, value := range fields {
		if err := validateCustomerFieldValue(key, value); err != nil {
			return err
		}
	}
	if phone, ok := fields["phone"].(string); ok && phone != "" && !isValidLiffPhone(phone) {
		return apperrors.WrapInvalidInput("電話番号の形式が正しくありません")
	}
	return nil
}

func isValidLiffPhone(phone string) bool {
	digits := 0
	for _, r := range phone {
		switch {
		case r >= '0' && r <= '9':
			digits++
		case r == '-' || r == '+' || r == ' ' || r == '(' || r == ')':
			continue
		default:
			return false
		}
	}
	return digits >= 10
}

// validateCustomerFieldValue は customer_fields の値を再帰的に検証する。
// string は長さ制限、配列・オブジェクトは要素を再帰的にチェックする。
func validateCustomerFieldValue(key string, value any) error {
	switch v := value.(type) {
	case string:
		if l := len([]rune(v)); l > maxCustomerFieldValueLength {
			return apperrors.WrapInvalidInput(fmt.Sprintf("customer_fields.%s は %d 文字以内で入力してください", key, maxCustomerFieldValueLength))
		}
	case []any:
		for i, item := range v {
			if err := validateCustomerFieldValue(fmt.Sprintf("%s[%d]", key, i), item); err != nil {
				return err
			}
		}
	case map[string]any:
		for subKey, subVal := range v {
			if err := validateCustomerFieldValue(key+"."+subKey, subVal); err != nil {
				return err
			}
		}
	}
	// number / bool / null はサイズ制限なし
	return nil
}
