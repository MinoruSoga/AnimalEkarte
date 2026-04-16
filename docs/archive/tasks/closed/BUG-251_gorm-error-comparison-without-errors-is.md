# BUG-251: `gorm.ErrRecordNotFound` を `==` で比較（`errors.Is` 未使用）

> **STATUS: FIXED** (2026-04-09) — 全3箇所修正済み、`go vet` パス確認

## 概要

`reservation_schedule_repository.go` と `reservation_customer_repository.go` で
`gorm.ErrRecordNotFound` を `==` 演算子で直接比較していた。
`errors.Is` に置換し、import に `"errors"` を追加した。

## 現状コード

### `backend/internal/repository/reservation_schedule_repository.go:95,99`
```go
if err != nil && err != gorm.ErrRecordNotFound {  // ← == で比較
    return nil, apperrors.Wrap(err, "find schedule entry")
}
if err == gorm.ErrRecordNotFound {  // ← == で比較
    // ...
}
```

### `backend/internal/repository/reservation_customer_repository.go:57`
```go
if err != nil && err != gorm.ErrRecordNotFound {  // ← == で比較
```

### 比較: 正しい実装
```go
// errors パッケージの Is を使用
if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, apperrors.FromGORM(err, "schedule_entry", "")
}
if errors.Is(err, gorm.ErrRecordNotFound) {
    // ...
}
```

## 影響範囲

| ファイル | 行 | 状態 |
|---------|-----|------|
| `backend/internal/repository/reservation_schedule_repository.go` | 95 | 未修正 |
| `backend/internal/repository/reservation_schedule_repository.go` | 99 | 未修正 |
| `backend/internal/repository/reservation_customer_repository.go` | 57 | 未修正 |

## 修正方針

```go
// 修正前
if err != nil && err != gorm.ErrRecordNotFound {
if err == gorm.ErrRecordNotFound {

// 修正後
if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
if errors.Is(err, gorm.ErrRecordNotFound) {
```

import に `"errors"` を追加すること。

## 準拠すべきプロジェクト規約

### `.claude/rules/go-language.md` — エラーハンドリング
> `errors.Is(err, ErrConflict)` で Sentinel エラーを判定する。直接比較は禁止。

## 優先度
**High** — GORM がエラーをラップした場合、`ErrRecordNotFound` を見逃して不正な 500 エラーを返す。

## 関連チケット
- BUG-244: バックエンド Go コード規約準拠監査（親チケット）

## 関連ファイル
- `backend/internal/repository/reservation_schedule_repository.go:95,99`
- `backend/internal/repository/reservation_customer_repository.go:57`
