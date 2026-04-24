# TASK-101: `err == gorm.ErrRecordNotFound` 直接比較 → `errors.Is()` に統一

## 優先度

**Medium** — エラーがラップされた場合に `==` 比較は失敗し、意図しないエラー伝播が起きる。

---

## 概要

FEAT-368 で追加した 3 repository ファイルで、`gorm.ErrRecordNotFound` をイコール演算子で
直接比較している。Go のエラーラッピング仕様上 `errors.Is()` を使わなければ、
`fmt.Errorf("...: %w", err)` でラップされたエラーを見落とす。

プロジェクトの他ファイル（`owner_repository.go`, `line_customer_repository.go` 等）は
すべて `errors.Is(err, gorm.ErrRecordNotFound)` を使用しており、今回の3ファイルが不統一。

---

## 問題箇所

### `repository/clinic_settings_repository.go:30`

```go
// ❌ 直接比較
if err == gorm.ErrRecordNotFound {
    return &model.ClinicSettings{...}, nil
}
```

### `repository/closing_special_period_repository.go:60`

```go
// ❌ 直接比較
if err == gorm.ErrRecordNotFound {
    return nil, nil
}
```

### `repository/cash_register_close_repository.go:81`

```go
// ❌ 直接比較
if err == gorm.ErrRecordNotFound {
    return nil, nil
}
```

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ repository/owner_repository.go
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, apperrors.WrapNotFound("owner", ...)
}

// ✅ repository/line_customer_repository.go:61
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, nil
}
```

---

## 修正方針

3 ファイルの `==` を `errors.Is()` に置き換える。`errors` パッケージの import を追加。

### 1. `repository/clinic_settings_repository.go:30`

```go
// ✅ 修正後
if errors.Is(err, gorm.ErrRecordNotFound) {
    return &model.ClinicSettings{
        ClinicID:            clinicID,
        ClosingAmPmBoundary: "14:00",
        ClosingWeekdayEnd:   "18:30",
        ClosingSundayEnd:    "17:30",
    }, nil
}
```

### 2. `repository/closing_special_period_repository.go:60`

```go
// ✅ 修正後
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, nil
}
```

### 3. `repository/cash_register_close_repository.go:81`

```go
// ✅ 修正後
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, nil
}
```

import ブロックに `"errors"` を追加すること。

---

## 影響範囲

| ファイル | 行 | 状態 |
|---------|---|------|
| `repository/clinic_settings_repository.go` | 30 | ❌ `==` 直接比較 |
| `repository/closing_special_period_repository.go` | 60 | ❌ `==` 直接比較 |
| `repository/cash_register_close_repository.go` | 81 | ❌ `==` 直接比較 |

---

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md` — Go エラーフロー

> Repository: GORM エラーは原則 `apperrors.FromGORM(err, "resource", id)` を使用。

`errors.Is()` を使わない直接比較は Go の標準エラーハンドリングパターンに違反する。

### `.claude/rules/go-language.md` — エラーハンドリング

> Sentinel エラー + `fmt.Errorf("...: %w", err)` で Wrap

`errors.Is()` はラッピングされたエラーを再帰的に unwrap して比較するため、
GORM が内部でエラーをラップしても正しく検出できる。

---

## 関連ファイル

- `repository/owner_repository.go:83,96` — 正しい実装の参照箇所
- `repository/line_customer_repository.go:61` — 正しい実装の参照箇所
