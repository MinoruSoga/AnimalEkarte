# TASK-127: `clinic_service.go` — 税率バリデーションが失敗しても無音でスキップ（エラー返却なし）

## 優先度

**High** — 不正な入力値がエラーにならず無音で無視される。ユーザーは更新が失敗したことに気づかない。

---

## 概要

`clinic_service.go` の `buildClinicUpdateFields` 関数（行 80-90）において、
`StandardTaxRate` および `ReducedTaxRate` のバリデーション（0 〜 1 の範囲チェック）が失敗した場合、
**エラーを返さずに該当フィールドを無音でスキップしている**。

クライアントが範囲外の値（例: `1.5`）を送信した場合、更新されずにレスポンスは 200 OK となる。
ユーザーはエラーを受け取らないまま更新が適用されていないことになる。

---

## 問題箇所

### `service/clinic_service.go:80-91`

```go
// ❌ 範囲外の値を無音でスキップ（エラー返却なし）
if input.StandardTaxRate != nil {
    r := *input.StandardTaxRate
    if r >= 0 && r <= 1 {
        fields["standard_tax_rate"] = r
    }
    // ← else: r が範囲外でも何もしない（エラーにならない）
}
if input.ReducedTaxRate != nil {
    r := *input.ReducedTaxRate
    if r >= 0 && r <= 1 {
        fields["reduced_tax_rate"] = r
    }
    // ← else: r が範囲外でも何もしない（エラーにならない）
}
```

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ service/billing_item_service.go — validateTaxType: 不正な値はエラーを返す
func validateTaxType(t string) error {
    switch model.TaxType(t) {
    case model.TaxTypeIncluded, model.TaxTypeExcluded:
        return nil
    }
    return apperrors.WrapInvalidInput("invalid tax_type: " + t)
}

// ✅ service/cage_service.go — validateCageType: 不正な値はエラーを返す
func validateCageType(t string) error {
    switch model.CageType(t) {
    case ...:
        return nil
    }
    return apperrors.WrapInvalidInput("invalid cage_type: " + t)
}
```

---

## 修正方針

### `service/clinic_service.go:80-91`

`buildClinicUpdateFields` を `error` 返却型に変更し、バリデーション失敗時はエラーを返す。

```go
// ✅ 修正後: buildClinicUpdateFields を error 返却型に変更
func buildClinicUpdateFields(input *UpdateClinicInput) (map[string]any, error) {
    fields := make(map[string]any)
    // ... (既存フィールド)

    if input.StandardTaxRate != nil {
        r := *input.StandardTaxRate
        if r < 0 || r > 1 {
            return nil, apperrors.WrapInvalidInput("standard_tax_rate must be between 0 and 1")
        }
        fields["standard_tax_rate"] = r
    }
    if input.ReducedTaxRate != nil {
        r := *input.ReducedTaxRate
        if r < 0 || r > 1 {
            return nil, apperrors.WrapInvalidInput("reduced_tax_rate must be between 0 and 1")
        }
        fields["reduced_tax_rate"] = r
    }
    return fields, nil
}
```

呼び出し側（`UpdateClinic` メソッド）も対応して変更:

```go
// ✅ 修正後の UpdateClinic
func (s *clinicService) UpdateClinic(ctx context.Context, id uint64, input *UpdateClinicInput) (*model.Clinic, error) {
    fields, err := buildClinicUpdateFields(input)  // error を受け取る
    if err != nil {
        return nil, err
    }
    if len(fields) == 0 {
        return s.GetClinicByID(ctx, id)  // 変更なし
    }
    ...
}
```

---

## 影響範囲

| ファイル | 行 | 状態 |
|---------|---|------|
| `service/clinic_service.go:80-91` | `buildClinicUpdateFields` 内の税率範囲チェック | ❌ 無音スキップ（エラーなし） |
| `service/clinic_service.go` | `UpdateClinic` 呼び出し元 | `buildClinicUpdateFields` の error 対応が必要 |

---

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — バックエンドアーキテクチャ規約

> handler → service → repository の軽量レイヤードを徹底

バリデーション失敗は必ず `apperrors.WrapInvalidInput(...)` でエラーを返し、
上位層でハンドリングできるようにする。

### プロジェクト内参照実装

- `service/billing_item_service.go` — `validateTaxType` でエラーを正しく返す
- `service/cage_service.go` — `validateCageType` でエラーを正しく返す
