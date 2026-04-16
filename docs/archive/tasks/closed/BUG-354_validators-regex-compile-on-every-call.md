# BUG-354: service/validators.go の validatePhoneFormat が毎呼び出し時に regexp.MustCompile を実行

## 概要

`service/validators.go:145` の `validatePhoneFormat` が呼び出しのたびに `regexp.MustCompile` で正規表現を再コンパイルしている。
同ファイル内の `emailPattern`（L13）はパッケージレベル変数として正しく定義されているのに対し、一貫性がない。

## 違反箇所

```go
// backend/internal/service/validators.go:138-150
func validatePhoneFormat(phone string) error {
    if phone == "" {
        return nil
    }
    // ❌ 毎回コンパイル
    phonePattern := regexp.MustCompile(`^0\d{1,4}-?\d{1,4}-?\d{4}$`)
    if !phonePattern.MatchString(phone) {
        return apperrors.WrapInvalidInput("...")
    }
    return nil
}
```

## 修正内容

```go
// パッケージレベル変数に昇格
var phonePattern = regexp.MustCompile(`^0\d{1,4}-?\d{1,4}-?\d{4}$`)

func validatePhoneFormat(phone string) error {
    if phone == "" {
        return nil
    }
    if !phonePattern.MatchString(phone) {
        return apperrors.WrapInvalidInput("...")
    }
    return nil
}
```

同様に `validatePostalCodeFormat`（L158）の `postalCodePattern` も同じ問題がある。

## 優先度

**LOW** — 正規表現コンパイルは `~1μs` 程度で、API レイテンシへの影響は限定的だが、不要な GC プレッシャーと一貫性の問題がある。

## 関連ファイル

- `backend/internal/service/validators.go:145,158`
