# TASK-129: 複数ハンドラ — POST 201 に Location ヘッダーなし

## 優先度

**Low** — REST 規約の不統一。機能には影響しない。

---

## 概要

以下のハンドラで POST 201 Created レスポンスに `Location` ヘッダーが設定されていない。
プロジェクト内の他の POST 201 ハンドラ（`cage_handler.go`, `vaccine_handler.go` 等）は
すべて `c.Header("Location", ...)` を設定しており、不統一になっている。

TASK-100（cash_register, closing_settings）・TASK-115（clinic_holiday）と同一パターン。

---

## 問題箇所

### `handler/staff_handler.go:110`

```go
// ❌ Location ヘッダーなし
c.JSON(http.StatusCreated, toStaffResponse(staff))
```

### `handler/permission_group_handler.go:105`

```go
// ❌ Location ヘッダーなし
c.JSON(http.StatusCreated, toPermissionGroupResponse(pg))
```

### `handler/inventory_handler.go` 付近（CreateInventory POST 201）

```go
// ❌ Location ヘッダーなし
c.JSON(http.StatusCreated, created)
```

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ handler/cage_handler.go (CreateCage)
c.Header("Location", fmt.Sprintf("/v1/cages/%d", cage.ID))
c.JSON(http.StatusCreated, toCageResponse(cage))

// ✅ handler/insurance_handler.go (CreateInsurance)
c.Header("Location", fmt.Sprintf("/v1/masters/insurances/%d", insurance.ID))
c.JSON(http.StatusCreated, toInsuranceResponse(insurance))
```

---

## 修正方針

各ハンドラに `c.Header("Location", ...)` を追加する。
エンドポイントパスは実際のルート登録に合わせること。

```go
// ✅ staff_handler.go
c.Header("Location", fmt.Sprintf("/v1/staffs/%d", staff.ID))
c.JSON(http.StatusCreated, toStaffResponse(staff))

// ✅ permission_group_handler.go
c.Header("Location", fmt.Sprintf("/v1/permission-groups/%d", pg.ID))
c.JSON(http.StatusCreated, toPermissionGroupResponse(pg))

// ✅ inventory_handler.go
c.Header("Location", fmt.Sprintf("/v1/inventory/%d", created.ID))
c.JSON(http.StatusCreated, toInventoryResponse(created))
```

---

## 影響範囲

| ファイル | 行 | 対象ハンドラ | 状態 |
|---------|---|------------|------|
| `handler/staff_handler.go:110` | CreateStaff | ❌ Location ヘッダーなし |
| `handler/permission_group_handler.go:105` | CreatePermissionGroup | ❌ Location ヘッダーなし |
| `handler/inventory_handler.go` | CreateInventory | ❌ Location ヘッダーなし |

---

## 準拠すべきプロジェクト規約

### プロジェクト内参照実装

- `handler/cage_handler.go` — `c.Header("Location", ...)` の正しいパターン
- 関連タスク: TASK-100, TASK-115（同一パターン先行チケット）
