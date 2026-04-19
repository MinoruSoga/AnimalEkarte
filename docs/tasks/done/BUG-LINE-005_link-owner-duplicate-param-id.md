# BUG-LINE-005: LINE予約管理 API 全域で :id 重複によりデータ破壊（CRITICAL SECURITY）

## 概要

`reservation_line_routes.go` のルートグループ全域で URL パスパラメータ `:id` が重複しており、Gin の `c.Param("id")` が **最初のマッチ（clinic ID）を返す** ため、全ての個別リソース操作が **意図しないレコードを操作する**。

**これは単なる 404 バグではない。他のレコードを誤って削除・更新する破壊的バグである。**

## 詳細

**ルート定義** (`backend/internal/handler/reservation_line_routes.go:11, 50, 52`):
```go
clinics := rg.Group("/clinics/:id")              // ← :id (1)
customers := clinics.Group("/line-customers")
customers.PATCH("/:id/link-owner", h.LinkOwnerToLineCustomer)  // ← :id (2)
```

結果的にフルパスが `/clinics/:id/line-customers/:id/link-owner` となる。

**ハンドラ** (`backend/internal/handler/line_customer_handler.go:31`):
```go
id, ok := parseIDParam(c, "id")  // ← Gin は最初にマッチした :id（clinic ID）を返す
```

**Gin の挙動**: `c.Params.ByName("id")` は `Params` スライスを先頭から走査し、**最初にマッチした値を返す**。よって customer の `:id` ではなく clinic の `:id` が取得される。

## 実測された被害（staging で確認済み）

### 検証1: `reservation-types`
- 実行: `PATCH /api/v1/clinics/3/reservation-types/1/status` body `{is_active: false}`
- 期待: reservation_type `id=1 (一般診察)` の is_active が false になる
- **実際: reservation_type `id=3 (ワクチン接種)` の is_active が false に書き換えられた**
- レスポンス: `{id: 3, name: "ワクチン接種", is_active: false}` ← clinic_id (=3) のレコードが返却される

### 検証2: `reservation-staffs`
- 実行: `PATCH /api/v1/clinics/3/reservation-staffs/1/status`
- 期待: staff `id=1 (林 文明)` の is_active が変更される
- **実際: staff `id=3 (三井 隆之)` の is_active が変更された**

### 検証3: `line-customers`
- 実行: `PATCH /api/v1/clinics/3/line-customers/4/link-owner`
- 期待: line_customer `id=4` に owner_id を紐付け
- 実際: customer `id=3` を探して見つからず 404

## 影響（CRITICAL）

### データ破壊系
- **PUT/PATCH/DELETE の全エンドポイントが clinic_id の ID を持つ別レコードを操作している**
- 削除系（DELETE /reservation-types/:id, /reservation-staffs/:id, /reservations/:id）は **誤削除** する
- 更新系（PUT, PATCH status, PATCH sort-order）は **誤更新** する

### 機能不動系
- clinic_id と同じ ID のレコードが存在しない場合は 404
- 例: customer/staff/type の個別取得やリンク付けが無差別に 404

## 影響を受けるエンドポイント（合計15本）

`reservation_line_routes.go` 配下で `:id` が重複する全ルート:

### reservation-types (7本)
- PUT `/clinics/:id/reservation-types/:id`
- DELETE `/clinics/:id/reservation-types/:id`
- PATCH `/clinics/:id/reservation-types/:id/status`
- PATCH `/clinics/:id/reservation-types/:id/sort-order`
- POST `/clinics/:id/reservation-types/:id/image`

### reservation-staffs (5本)
- PUT `/clinics/:id/reservation-staffs/:id`
- DELETE `/clinics/:id/reservation-staffs/:id`
- PATCH `/clinics/:id/reservation-staffs/:id/status`
- PATCH `/clinics/:id/reservation-staffs/:id/sort-order`
- POST `/clinics/:id/reservation-staffs/:id/image`

### schedules (3本) — さらに :date も重複する
- GET `/clinics/:id/reservation-staffs/:id/schedules`
- PUT `/clinics/:id/reservation-staffs/:id/schedules/:date`
- DELETE `/clinics/:id/reservation-staffs/:id/schedules/:date`

### reservations (1本)
- DELETE `/clinics/:id/reservations/:id`

### line-customers (1本)
- PATCH `/clinics/:id/line-customers/:id/link-owner`

## 修正案

### Option A: ルート定義でパラメータ名を分ける

`reservation_line_routes.go` の全てのネストしたルートで `:id` の重複を解消する:

```go
// Before
clinics := rg.Group("/clinics/:id")
customers := clinics.Group("/line-customers")
customers.PATCH("/:id/link-owner", h.LinkOwnerToLineCustomer)

// After
clinics := rg.Group("/clinics/:clinicId")
customers := clinics.Group("/line-customers")
customers.PATCH("/:customerId/link-owner", h.LinkOwnerToLineCustomer)
```

そしてハンドラ側も `parseIDParam(c, "customerId")` に変更。

### Option B: extractClinicID を URL param ベースに

`extractClinicID` は JWT ベースなので、URL の `:id` を customer ID 専用に使う場合は Option A が確実。

## 修正手順

1. `reservation_line_routes.go` で `/clinics/:id` → `/clinics/:clinicId` に変更
2. 配下の全リソース ID を固有名にリネーム:
   - `reservation-types/:id` → `/:typeId`
   - `reservation-staffs/:id` → `/:staffId`
   - `reservations/:id` → `/:reservationId`
   - `line-customers/:id` → `/:customerId`
3. 各ハンドラの `parseIDParam(c, "id")` を対応する固有名に変更
4. ハンドラは既に JWT 経由で clinic_id を取得する（`extractClinicID`）ので URL の clinic_id 部分は使わなくてよい。ただしマルチテナントの RBAC チェックには URL と JWT の一致検証が望ましい。

## テスト観点

- 既存テスト (`liff_handler_test.go`, `line_customer_handler_test.go`, etc.) がこのバグを検出できていない理由を調査
- 恐らくテストのルート登録が本番と違う (`liff_handler_test.go:132` のように `/types` 直接マウントなど)
- テスト側も本番のグループ階層を再現すべき

## 優先度

**CRITICAL / SECURITY** — 本番環境では他テナント・他レコードの意図しない改変・削除が起きる。Phase 7 結合テストのブロッカー。

## 再現手順（staging）

```javascript
// 検証1: /reservation-types/1/status にリクエストしているのに id=3 が変更される
fetch('/api/v1/clinics/3/reservation-types/1/status', {
  method: 'PATCH', credentials: 'include',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ is_active: false })
}).then(r => r.json()).then(console.log);
// → { id: 3, name: "ワクチン接種", is_active: false }  ← id=1 ではなく id=3 が返る
```

## 確認環境

- staging: `https://api.stg.noah-karte.com/api/v1/clinics/3/...`
- テスト実施日: 2026-04-14
