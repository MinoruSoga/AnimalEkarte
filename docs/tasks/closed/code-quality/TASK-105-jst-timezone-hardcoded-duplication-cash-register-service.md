# TASK-105: JST タイムゾーンのハードコード重複（cash_register_service.go）

## 優先度

**Low** — 機能は正常。コードの一貫性・メンテナンス性の改善。

---

## 概要

`cash_register_service.go` の `GetPreview` と `Close` の両メソッドで
`time.FixedZone("Asia/Tokyo", 9*60*60)` が同一内容でハードコードされている。
タイムゾーンオフセット（9時間）を変更する際に2箇所修正が必要になり、
変更漏れのリスクがある。

---

## 問題箇所

### `service/cash_register_service.go:74`

```go
// GetPreview 内
jst := time.FixedZone("Asia/Tokyo", 9*60*60)
dateJST := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, jst)
```

### `service/cash_register_service.go:123`

```go
// Close 内（同一記述）
jst := time.FixedZone("Asia/Tokyo", 9*60*60)
dateJST := time.Date(input.Date.Year(), input.Date.Month(), input.Date.Day(), 0, 0, 0, 0, jst)
```

---

## 修正方針

パッケージレベルの変数として定義し、両メソッドから参照する。

```go
// ✅ ファイル先頭（package 宣言直下）に追加
var jstLocation = time.FixedZone("Asia/Tokyo", 9*60*60)

// GetPreview 内
dateJST := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, jstLocation)

// Close 内
dateJST := time.Date(input.Date.Year(), input.Date.Month(), input.Date.Day(), 0, 0, 0, 0, jstLocation)

// resolvePeriodRange のシグネチャからも jst パラメータを削除できる
// （現在: resolvePeriodRange(dateJST, period, schedule, jst)）
// （修正: resolvePeriodRange(dateJST, period, schedule)）
```

`resolvePeriodRange` の第4引数 `jst *time.Location` も `jstLocation` 参照に変更することで
引数を1つ削減できる。

---

## 影響範囲

| ファイル | 行 | 問題 |
|---------|---|------|
| `service/cash_register_service.go` | 74, 123 | JST ハードコード重複 |
| `service/cash_register_service.go:186` | `resolvePeriodRange` のシグネチャ | jst 引数の削除候補 |

---

## 準拠すべきプロジェクト規約

### `.claude/rules/go-language.md`

> Global mutable state 禁止

`var jstLocation` はイミュータブルな `time.Location` ポインタなので問題なし（read-only）。
パッケージレベルの定数的な値は許容される。
