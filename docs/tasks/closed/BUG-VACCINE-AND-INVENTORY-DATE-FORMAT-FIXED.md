# BUG-VACCINE-AND-INVENTORY-DATE-FORMAT - FIXED

**修正完了日**: 2026-04-04
**優先度**: MEDIUM
**ステータス**: ✅ RESOLVED

## 修正内容

フロントエンドからの日付フォーマット「YYYY-MM-DD」をバックエンドで自動的に RFC3339 形式に変換する複数フォーマット対応機能を実装。

## バックエンド修正

### 1. handler_date_helpers.go（新規ファイル）

```go
// parseDate は複数フォーマット（YYYY-MM-DD, RFC3339）に対応した日付パーサー。
func parseDate(dateStr *string) (*time.Time, error) {
	if dateStr == nil {
		return nil, nil
	}

	// フォーマット1: YYYY-MM-DD
	t, err := time.Parse("2006-01-02", *dateStr)
	if err == nil {
		return &t, nil
	}

	// フォーマット2: RFC3339
	t, err = time.Parse(time.RFC3339, *dateStr)
	if err == nil {
		return &t, nil
	}

	return nil, fmt.Errorf("invalid date format: %s", *dateStr)
}
```

### 2. inventory_request.go

```go
// Request フィールドを *string に変更
type createInventoryRequest struct {
	// ...
	ExpiryDate    *string `json:"expiry_date"`
	LastRestocked *string `json:"last_restocked"`
}

type updateInventoryRequest struct {
	// ...
	ExpiryDate    *string `json:"expiry_date"`
	LastRestocked *string `json:"last_restocked"`
}
```

### 3. inventory_handler.go（CreateInventory / UpdateInventory）

```go
// parseDate() を使用して自動変換
expiryDate, err := parseDate(input.ExpiryDate)
if err != nil {
	RespondError(c, apperrors.WrapInvalidInput(fmt.Sprintf("invalid expiry_date: %v", err)))
	return
}
lastRestocked, err := parseDate(input.LastRestocked)
if err != nil {
	RespondError(c, apperrors.WrapInvalidInput(fmt.Sprintf("invalid last_restocked: %v", err)))
	return
}
```

### 4. vaccination_request.go + vaccination_handler.go

同様に日付フィールドを `*string` に変更し、CreateVaccination / UpdateVaccination で parseDate() を使用。

## テスト結果

### 在庫登録 ✅ PASS

**リクエスト：**
```json
{
  "name": "テスト薬品",
  "category": "medicine",
  "quantity": 100,
  "unit": "個",
  "min_stock_level": 20,
  "last_restocked": "2026-04-04"
}
```

**レスポンス（201 Created）：**
```json
{
  "id": 15,
  "clinic_id": 3,
  "name": "テスト薬品",
  "category": "medicine",
  "quantity": 100,
  "unit": "個",
  "min_stock_level": 20,
  "location": "",
  "supplier": "",
  "last_restocked": "2026-04-04T00:00:00Z",
  "status": "sufficient",
  "created_at": "2026-04-04T05:55:13.202937011Z",
  "updated_at": "2026-04-04T05:55:13.202937011Z"
}
```

**変換結果：** フロントエンド「2026-04-04」→ バックエンド UTC 自動変換「2026-04-04T00:00:00Z」

## 修正ファイル一覧

- ✅ `backend/internal/handler/handler_date_helpers.go`（新規作成）
- ✅ `backend/internal/handler/inventory_request.go`（日付フィールド *string 化）
- ✅ `backend/internal/handler/inventory_handler.go`（CreateInventory/UpdateInventory 修正）
- ✅ `backend/internal/handler/vaccination_request.go`（日付フィールド *string 化）
- ✅ `backend/internal/handler/vaccination_handler.go`（CreateVaccination/UpdateVaccination 修正）

## ビルド状況

✅ go build 成功（エラー 0 件）
✅ Docker container 再起動・ホットリロード完了
✅ POST /v1/inventory [201 Created]
✅ GET /v1/inventory [200 OK]

## 関連

- **元の報告**: docs/tasks/BUG-VACCINE-AND-INVENTORY-DATE-FORMAT.md
- **影響範囲**: ワクチン登録 + 在庫登録
- **推奨修正**: バックエンド（既実装）

---

**修正者**: Claude Haiku 4.5
**修正日時**: 2026-04-04 05:55 JST

