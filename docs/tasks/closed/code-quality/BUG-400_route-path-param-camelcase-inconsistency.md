# BUG-400: サブリソースのパスパラメータが camelCase で命名規則違反

## 概要
`staff_handler.go` のルート定義（488・492行目）でサブリソースの DELETE エンドポイントのパスパラメータが camelCase（`:unavailableTimeId`、`:occupationId`）になっている。プロジェクト全体の他のパスパラメータは全て `:id` のみを使用しており、Gin の慣例でも snake_case または単純な `:id` が標準。camelCase のパスパラメータはクライアント（API ドキュメント・SDK 生成）にも影響する。

## 再現手順
コードレビューで確認可能。

## 現状コード

### `backend/internal/handler/staff_handler.go:488,492`（ルート定義）
```go
masters.DELETE(
    "/reservation-types/:id/unavailable-times/:unavailableTimeId",  // ← camelCase
    perm(model.ResourceMasterReservationType, "edit"),
    h.DeleteUnavailableTime)

masters.DELETE(
    "/reservation-types/:id/occupations/:occupationId",  // ← camelCase
    perm(model.ResourceMasterReservationType, "edit"),
    h.UnlinkReservationTypeOccupation)
```

### `backend/internal/handler/reservation_type_handler.go:206,268`（パラメータ取得）
```go
unavailableTimeID, ok := parseIDParam(c, "unavailableTimeId")  // ← camelCase キー
occupationID, ok := parseIDParam(c, "occupationId")            // ← camelCase キー
```

### 比較: 正しい実装（プロジェクト内参照実装）
```go
// 全ての他のパスパラメータは :id のみ
masters.DELETE("/cages/:id", h.DeleteCage)
masters.DELETE("/medicines/:id", h.DeleteMedicine)
```

## 修正方針

### 1. ルート定義（staff_handler.go:488,492）を snake_case に変更
```go
masters.DELETE("/reservation-types/:id/unavailable-times/:unavailable_time_id", ..., h.DeleteUnavailableTime)
masters.DELETE("/reservation-types/:id/occupations/:occupation_id", ..., h.UnlinkReservationTypeOccupation)
```

### 2. ハンドラ（reservation_type_handler.go:206,268）のキー名も同期
```go
unavailableTimeID, ok := parseIDParam(c, "unavailable_time_id")
occupationID, ok := parseIDParam(c, "occupation_id")
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/naming-conventions.md` — API エンドポイント
> パス命名: **kebab-case**。パスパラメータは `:id` または `:snake_case_id` を使用。

## 優先度
**Low** — 機能上の問題なし。命名規則違反。API クライアント・ドキュメント生成への軽微な影響。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/handler/staff_handler.go:488,492` — ルート定義（修正対象）
- `backend/internal/handler/reservation_type_handler.go:206,268` — パラメータ取得（修正対象）
