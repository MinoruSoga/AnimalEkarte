# BUG: 在庫管理 - 最低在庫数のクロスフィールドバリデーション未実装

**Issue Date**: 2026-04-02
**Severity**: Medium
**Status**: New
**Component**: Inventory Management (在庫管理)

## 概要

在庫登録フォームで、**最低在庫数（minStockLevel）> 現在庫数（quantity）**という不正な組み合わせを防ぐクロスフィールドバリデーションが実装されていません。

## 再現手順

1. `/inventory/new` に移動（または既存在庫を編集）
2. 以下のように入力：
   - **品名**: "テスト品"
   - **カテゴリ**: "医薬品"
   - **現在庫数**: "10"
   - **最低在庫数**: "20"  ← **現在庫数より大きい**
3. 「保存」ボタンをクリック
4. **期待**: バリデーションエラー「最低在庫数は現在庫数以下で設定してください」
5. **実際**: エラーなく登録される（不正状態で登録可能）

## 詳細

### 影響範囲

| ファイル | 行番号 | 内容 |
|---------|--------|------|
| `frontend/src/features/inventory/hooks/use-inventory-form.ts` | 48-96 | formAction コールバック内にバリデーション実装なし |
| `frontend/src/features/inventory/routes/InventoryForm.tsx` | 145-176 | minStockLevel フィールド定義（バリデーションなし） |
| `backend/internal/service/inventory_service.go` | - | サーバー側バリデーションも確認必要 |

### コード現状

**frontend/src/features/inventory/hooks/use-inventory-form.ts**

```typescript
// Line 48-96: formAction callback
const [formState, formAction, isPending] = useActionState(
  async (_prevState: FormState, formData: FormData): Promise<FormState> => {
    const quantityStr = formData.get("quantity") as string;
    const minStockLevelStr = formData.get("minStockLevel") as string;
    // ...
    // ❌ クロスフィールドバリデーション（minStockLevel <= quantity）なし
    try {
      if (isEdit && id) {
        // ...
      } else {
        const req: CreateInventoryItemRequest = {
          quantity: quantityStr ? Number(quantityStr) : 0,
          min_stock_level: minStockLevelStr ? Number(minStockLevelStr) : 0,
          // ...
        };
        await createMutation.mutateAsync(req);
      }
      return { success: true, timestamp: Date.now() };
    } catch (error) { ... }
  }
);
```

## 修正方法

### フロントエンド（必須）

**use-inventory-form.ts** の formAction callback 内に、以下のバリデーションを追加：

```typescript
const quantity = quantityStr ? Number(quantityStr) : 0;
const minStockLevel = minStockLevelStr ? Number(minStockLevelStr) : 0;

// ❌ 不正：最低在庫 > 現在庫
if (minStockLevel > quantity) {
  toast.error("最低在庫数は現在庫数以下で設定してください");
  return { success: false, timestamp: Date.now() };
}
```

### バックエンド（推奨）

**backend/internal/service/inventory_service.go** の Create/Update メソッドにも同様のチェック実装。

```go
if req.MinStockLevel > req.Quantity {
  return apperrors.WrapInvalidInput("min_stock_level must be <= quantity")
}
```

## テスト計画

- [ ] フロントエンド: minStockLevel > quantity で送信 → エラートースト表示確認
- [ ] フロントエンド: minStockLevel = quantity で送信 → 正常登録確認
- [ ] フロントエンド: minStockLevel < quantity で送信 → 正常登録確認
- [ ] バックエンド: API 直接呼び出しで minStockLevel > quantity → 400/422 エラー確認

## 関連バグ

- BUG-066: フロントエンド フォームバリデーション強化（本件と同カテゴリ）

---

**発見者**: Claude Code (2026-04-02 ローカル検証)
**最終更新**: 2026-04-02
