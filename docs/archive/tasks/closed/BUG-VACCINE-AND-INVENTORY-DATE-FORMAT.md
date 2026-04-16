# BUG-VACCINE-AND-INVENTORY-DATE-FORMAT

**報告日**: 2026-04-04
**優先度**: MEDIUM
**ステータス**: 要修正

## 概要

ワクチン登録フォーム・在庫登録フォームで、日付フィールド送信時に RFC3339 フォーマット不一致エラーが発生。フロントエンドが `YYYY-MM-DD` 形式で送信しているのに対し、バックエンドが `YYYY-MM-DDTHH:MM:SSZ07:00` 形式を期待している。

## 影響範囲

| 画面 | フィールド | エラーメッセージ |
|------|----------|-----------------|
| ワクチン登録Tab | 接種日 | `parsing time "2026-04-04" as "2006-01-02T15:04:05Z07:00"` |
| 在庫登録フォーム | 最終入荷日 | `parsing time "2026-04-04" as "2006-01-02T15:04:05Z07:00"` |

## 再現手順

### ワクチン登録
1. カルテ → Tab4（ワクチン）
2. ワクチン種別選択 → 日付入力（NotionDatePicker で「2026年4月4日」選択）
3. 「保存」ボタンクリック
4. **結果**: POST リクエスト 400 エラー

### 在庫登録
1. 在庫管理 → 新規登録
2. 品名、カテゴリ等を入力 → 最終入荷日に「2026年4月4日」選択
3. 「登録」ボタンクリック
4. **結果**: POST リクエスト 400 エラー

## 技術的詳細

### リクエストボディ例

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

### バックエンド期待値

バックエンド handler が以下のフォーマットを期待：
```go
time.RFC3339 // "2006-01-02T15:04:05Z07:00"
```

### エラーメッセージ

```
parsing time "2026-04-04" as "2006-01-02T15:04:05Z07:00": cannot parse "" as "T"
```

## 原因

NotionDatePicker コンポーネントが日付を `YYYY-MM-DD` 形式で返すが、バックエンド handler で `time.Parse(time.RFC3339, ...)` を使用しており、タイムゾーンと時刻コンポーネントを含まない形式が受け入れられない。

## 修正方法（オプション A: バックエンド修正）

```go
// handler/inventory_handler.go
func (h *handler) CreateInventoryItem(c *gin.Context) {
    var req CreateInventoryItemRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // ...
    }

    // ✅ 修正: 複数フォーマットをサポート
    var lastRestockedTime *time.Time
    if req.LastRestocked != "" {
        // 試行1: RFC3339
        t, err := time.Parse(time.RFC3339, req.LastRestocked)
        if err != nil {
            // 試行2: YYYY-MM-DD 形式
            t, err = time.Parse("2006-01-02", req.LastRestocked)
        }
        if err != nil {
            RespondError(c, apperrors.WrapInvalidInput(fmt.Sprintf("invalid date format: %v", err)))
            return
        }
        lastRestockedTime = &t
    }

    // service 呼び出し
    result, err := h.service.CreateInventoryItem(c.Request.Context(), &CreateInventoryItemInput{
        LastRestocked: lastRestockedTime,
        // ...
    })
    // ...
}
```

## 修正方法（オプション B: フロントエンド修正）

```typescript
// features/inventory/api/create-inventory.ts
export async function createInventoryItem(input: CreateInventoryItemRequest) {
  // last_restocked を RFC3339 に変換
  const payload = {
    ...input,
    last_restocked: input.last_restocked
      ? new Date(input.last_restocked).toISOString()  // ✅ RFC3339 に変換
      : undefined,
  };

  const response = await axios.post<{ id: number }>("/v1/inventory", payload);
  return response.data;
}
```

## テスト結果

- **ワクチン登録**: API 呼び出し成功（toast 表示）だが日付解析エラー
- **在庫登録**: API 呼び出し失敗（400 Bad Request）

## 推奨修正

**オプション A（バックエンド）** を推奨。複数フォーマット対応でフロントエンド側の柔軟性が確保される。

---

## 関連ファイル

- `backend/internal/handler/inventory_handler.go` — CreateInventoryItem ハンドラ
- `backend/internal/handler/vaccination_handler.go` — CreateVaccination ハンドラ
- `frontend/src/features/inventory/api/create-inventory.ts` — 在庫登録 API
- `frontend/src/features/medical-records/api/create-vaccination.ts` — ワクチン登録 API
- `frontend/src/components/shared/NotionDatePicker/NotionDatePicker.tsx` — 日付ピッカー

