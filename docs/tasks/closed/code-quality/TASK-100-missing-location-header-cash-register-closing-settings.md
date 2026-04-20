# TASK-100: POST ハンドラ — Location ヘッダー欠落（cash_register / closing_settings）

## 優先度

**Medium** — REST 規約違反。201 Created を返す全 POST エンドポイントで Location を返すのは必須。

---

## 概要

TASK-099 で他の Create ハンドラの Location ヘッダーを修正したが、FEAT-368 で追加した
`CloseCashRegister` と `CreateSpecialPeriod` の 2 エンドポイントが漏れていた。

---

## 問題箇所

### `handler/cash_register_handler.go:93`

```go
// ❌ Location ヘッダーなしで 201 返却
c.JSON(http.StatusCreated, record)
```

### `handler/closing_settings_handler.go:129`

```go
// ❌ Location ヘッダーなしで 201 返却
c.JSON(http.StatusCreated, period)
```

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ handler/payment_method_master_handler.go:64
c.Header("Location", fmt.Sprintf("/v1/payment-methods/%d", m.ID))
c.JSON(http.StatusCreated, m)
```

---

## 修正方針

### 1. `handler/cash_register_handler.go:93`

```go
// ✅ 修正後
c.Header("Location", fmt.Sprintf("/v1/cash-register/closes/%d", record.ID))
c.JSON(http.StatusCreated, record)
```

### 2. `handler/closing_settings_handler.go:129`

```go
// ✅ 修正後
c.Header("Location", fmt.Sprintf("/v1/closing-settings/special-periods/%d", period.ID))
c.JSON(http.StatusCreated, period)
```

---

## 影響範囲

| ファイル | エンドポイント | 状態 |
|---------|--------------|------|
| `handler/cash_register_handler.go:93` | POST /v1/cash-register/closes | ❌ Location なし |
| `handler/closing_settings_handler.go:129` | POST /v1/closing-settings/special-periods | ❌ Location なし |

---

## 準拠すべきプロジェクト規約

### `.claude/rules/api.md` — Endpoint Design

> Use proper HTTP status codes

RFC 7231 § 7.1.2: POST で 201 Created を返す場合、Location ヘッダーで新規リソースの URI を返すこと。

### プロジェクト内参照実装

- `handler/payment_method_master_handler.go:64-65` — Location ヘッダー付きの正しい実装
- `handler/exam_type_handler.go` — 同上

---

## 関連チケット

- TASK-099: animal_species / shift_template / reservation_type の同種問題（クローズ済み）
- TASK-073: その他11マスタの同種問題（クローズ済み）
