# BUG-378: マスタハンドラのレスポンス構造が不統一（配列直返し vs gin.H{"data": ...}）

## 概要
`reservation_type_handler.go` では、同一ハンドラファイル内でリスト系エンドポイントのレスポンス構造が混在している。`ListReservationTypes` は配列を直接返すが、`ListUnavailableTimes` と `ListReservationTypeOccupations` は `gin.H{"data": [...]}` ラッパーを使用している。フロントエンドのAPIクライアントが異なる構造を処理しなければならず、型安全性と保守性が低下する。

## 再現手順
1. `GET /v1/masters/reservation-types` → レスポンス: `[{...}, {...}]`（直接配列）
2. `GET /v1/masters/reservation-types/{id}/unavailable-times` → レスポンス: `{"data": [{...}]}`（ラッパーあり）
3. **結果**: 同一エンティティグループで異なるレスポンス形式

## 期待する動作
- 全リスト系エンドポイントで一貫したレスポンス構造を使用すること
- プロジェクト全体の他マスタハンドラ（直接配列）と統一すること

## 現状コード

### `backend/internal/handler/reservation_type_handler.go:46`（直接配列）
```go
c.JSON(http.StatusOK, mapSlice(reservationTypes, toReservationTypeResponse))
```

### `backend/internal/handler/reservation_type_handler.go:158`（ラッパーあり）
```go
c.JSON(http.StatusOK, gin.H{"data": mapSlice(items, toUnavailableTimeResponse)})
```

### `backend/internal/handler/reservation_type_handler.go:232`（ラッパーあり）
```go
c.JSON(http.StatusOK, gin.H{"data": mapSlice(items, toReservationTypeOccupationResponse)})
```

### 比較: 正しい実装（プロジェクト内参照実装）
```go
// backend/internal/handler/chief_complaint_handler.go:45
c.JSON(http.StatusOK, mapSlice(categories, toChiefComplaintResponse))

// backend/internal/handler/cage_handler.go
c.JSON(http.StatusOK, mapSlice(cages, toCageResponse))

// → 全マスタハンドラは直接配列を返す
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `backend/internal/handler/reservation_type_handler.go:158` | ListUnavailableTimes のレスポンス | 要修正 |
| `backend/internal/handler/reservation_type_handler.go:232` | ListReservationTypeOccupations のレスポンス | 要修正 |
| `frontend/src/features/reservations/` | API クライアントの response 型・transform | 確認・修正が必要 |

## 修正方針

### 1. `backend/internal/handler/reservation_type_handler.go:158`
```go
// 修正前
c.JSON(http.StatusOK, gin.H{"data": mapSlice(items, toUnavailableTimeResponse)})

// 修正後
c.JSON(http.StatusOK, mapSlice(items, toUnavailableTimeResponse))
```

### 2. `backend/internal/handler/reservation_type_handler.go:232`
```go
// 修正前
c.JSON(http.StatusOK, gin.H{"data": mapSlice(items, toReservationTypeOccupationResponse)})

// 修正後
c.JSON(http.StatusOK, mapSlice(items, toReservationTypeOccupationResponse))
```

### 3. フロントエンド確認
`frontend/src/features/` の予約種別関連 API クライアントで `response.data` を参照している箇所があれば、直接配列参照に修正する。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/api.md` — Error Response Format
> レスポンス形式は一貫性を持たせること

### プロジェクト内参照実装
`backend/internal/handler/chief_complaint_handler.go:45`、`cage_handler.go` — 全マスタハンドラで `mapSlice()` 直接返却が実装されている

## 優先度
**Medium** — API 契約の不整合。フロントエンドの transform が誤実装になるリスク。機能バグではないが放置すると技術的負債が拡大する。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/handler/reservation_type_handler.go:46,158,232` — 問題箇所
- `frontend/src/features/` 予約種別関連APIクライアント — 影響確認が必要
