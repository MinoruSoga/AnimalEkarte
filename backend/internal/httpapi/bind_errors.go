package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// ParseBindError は validator.ValidationErrors を人間可読メッセージに変換する。
// ValidationErrors でない場合はそのまま err.Error() を返す。
func ParseBindError(err error) string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		msgs := make([]string, 0, len(ve))
		for _, fe := range ve {
			msgs = append(msgs, formatValidationError(fe))
		}
		return strings.Join(msgs, "; ")
	}
	// BUG-129: Go 内部エラーメッセージをサニタイズし、構造体名・フィールド型の漏洩を防止
	var unmarshalErr *json.UnmarshalTypeError
	if errors.As(err, &unmarshalErr) {
		return fmt.Sprintf("%s: 正しい形式で入力してください", unmarshalErr.Field)
	}
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return "リクエストの形式が正しくありません"
	}
	return "入力値が正しくありません"
}

func formatValidationError(fe validator.FieldError) string {
	// BUG-155: json タグに近い snake_case フィールド名を返す（API 仕様として公開情報）
	field := camelToSnake(fe.Field())
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s は必須です", field)
	case "min":
		return fmt.Sprintf("%s は %s 以上で入力してください", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s は %s 以下で入力してください", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s は次のいずれかで指定してください: %s", field, strings.ReplaceAll(fe.Param(), " ", ", "))
	default:
		return fmt.Sprintf("%s の値が正しくありません", field)
	}
}

// camelToSnake は CamelCase を snake_case に変換する。
// 連続した大文字（頭字語）は 1 単語として扱う。
// "OwnerName" → "owner_name"
// "IsDangerous" → "is_dangerous"
// "TypeID" → "type_id"     ← BUG-LINE-010: 以前は "type_i_d" になっていた
// "HTTPServer" → "http_server"
func camelToSnake(s string) string {
	var b strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) {
			prev := runes[i-1]
			// 直前が小文字/数字 → 単語境界として `_` を挿入
			// 直前が大文字で次が小文字 → 頭字語の末尾扱いで `_` を挿入 ("HTTPServer" → "http_server")
			// それ以外（連続大文字の途中）は `_` を挿入しない
			insertUnderscore := !unicode.IsUpper(prev)
			if !insertUnderscore && i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
				insertUnderscore = true
			}
			if insertUnderscore {
				b.WriteByte('_')
			}
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}
