# BUG-245: buildXxxUpdateFields の price ポインタ未デリファレンス（8ファイル9箇所）

> **STATUS: FIXED** (2026-04-09) — 全9箇所修正済み、`go vet` パス確認

## 概要

`buildXxxUpdateFields` 関数群で `price` フィールドにポインタ (`*int64`) をデリファレンスせずに `map[string]any` に代入していた。
GORM の `Updates(map[string]any)` にポインタを渡した場合、SQL バインド時の動作が未定義であり、
PATCH で price を更新しようとした際に不正な値が DB に書き込まれるデータ破壊バグ。

監査時は6ファイルと報告されていたが、実際は8ファイル9箇所に同一バグが存在した
（hospitalization_plan, cage, consultation が追加）。

## 現状コード

### `backend/internal/service/vaccine_service.go:72`
```go
fields["price"] = input.Price  // *int64 ポインタをそのまま代入
```

### `backend/internal/service/procedure_service.go:96`
```go
fields["price"] = input.Price
```

### `backend/internal/service/checkup_type_service.go:90`
```go
fields["price"] = input.Price
```

### `backend/internal/service/exam_type_service.go:86`
```go
fields["price"] = input.Price
```

### `backend/internal/service/trimming_master_service.go:86`
```go
fields["price"] = input.Price  // TrimmingCourse
```

### `backend/internal/service/trimming_master_service.go:176`
```go
fields["price"] = input.Price  // TrimmingOption
```

### 比較: 正しい実装（同ファイル内の他フィールド）
```go
// vaccine_service.go:68-70 — name は正しくデリファレンスしている
if input.Name != nil {
    fields["name"] = *input.Name  // ✅ デリファレンス済み
}
```

## 影響範囲

| ファイル | 行 | エンティティ | 状態 |
|---------|-----|-------------|------|
| `backend/internal/service/vaccine_service.go` | 72 | Vaccine | 未修正 |
| `backend/internal/service/procedure_service.go` | 96 | Procedure | 未修正 |
| `backend/internal/service/checkup_type_service.go` | 90 | CheckupType | 未修正 |
| `backend/internal/service/exam_type_service.go` | 86 | ExamType | 未修正 |
| `backend/internal/service/trimming_master_service.go` | 86 | TrimmingCourse | 未修正 |
| `backend/internal/service/trimming_master_service.go` | 176 | TrimmingOption | 未修正 |

## 修正方針

全6箇所で `input.Price` → `*input.Price` に変更する。

### 修正例（全箇所共通）
```go
// 修正前
fields["price"] = input.Price

// 修正後
fields["price"] = *input.Price
```

### 安全性
`buildXxxUpdateFields` は `input.Price != nil` のガード内で呼ばれるため、
`*input.Price` のデリファレンスは安全（nil panic しない）。

## 準拠すべきプロジェクト規約

### `.claude/rules/go-language.md` — GORM PATCH パターン
> PATCH は ポインタ型 + buildXxxUpdateFields()
> ゼロ値問題を回避するために `*input.Field` でデリファレンスして map に格納する。

### プロジェクト内参照実装
`backend/internal/service/owner_service.go` の `buildOwnerUpdateFields` — 全フィールドが正しく `*input.Field` でデリファレンスされている。

## 優先度
**Critical** — PATCH API で price を更新した場合にデータ破壊が発生する。本番環境で価格変更操作が行われた時点で即影響。

## 関連チケット
- BUG-244: バックエンド Go コード規約準拠監査（親チケット）

## 関連ファイル
- `backend/internal/service/vaccine_service.go:72` — Vaccine price
- `backend/internal/service/procedure_service.go:96` — Procedure price
- `backend/internal/service/checkup_type_service.go:90` — CheckupType price
- `backend/internal/service/exam_type_service.go:86` — ExamType price
- `backend/internal/service/trimming_master_service.go:86` — TrimmingCourse price
- `backend/internal/service/trimming_master_service.go:176` — TrimmingOption price
