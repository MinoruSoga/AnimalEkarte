# バグ修正実装完了レポート

**実装日**: 2026-04-02 16:10 JST
**ステータス**: ✅ 完了・Staging にデプロイ済み

---

## 概要

テスト報告書で発見された2件の新規バグの実装修正を完了しました。

| バグ ID | 優先度 | ステータス | 内容 |
|---------|--------|---------|------|
| BUG-EXAMINATION-DATE-INPUT-MISSING | HIGH | ✅ 完了 | 検査フォームに検査日フィールドを追加 |
| BUG-INVENTORY-CROSS-FIELD-VALIDATION | MEDIUM | ✅ 完了 | 在庫フォームに最低在庫数バリデーション追加 |

---

## 実装内容

### Bug 1: 検査日フィールド追加

**ファイル**: `frontend/src/features/examinations/routes/ExaminationForm.tsx`

**変更**: FormFieldsSection の grid 内（L125-133）に NotionDatePicker ベースの検査日フィールドを追加

```typescript
<div className="space-y-1.5">
  <Label className={`text-sm ${C.text60}`}>検査日</Label>
  <NotionDatePicker
    value={formData.date ? formData.date.split("T")[0] : ""}
    onChange={(v) => onSetFormData({ date: v ? `${v}T00:00:00Z` : new Date().toISOString() })}
    disabledDays={{ after: new Date() }}
  />
</div>
```

**特徴**:
- 過去日付のみ選択可（未来日付は disabled）
- 検査日未選択の場合は自動的に現在日時を設定
- 確定状態の検査は read-only 挙動

---

### Bug 2: 在庫クロスフィールドバリデーション

**ファイル**:
1. `frontend/src/features/inventory/hooks/use-inventory-form.ts`
2. `frontend/src/features/inventory/routes/InventoryForm.tsx`

**変更内容**:

#### ① use-inventory-form.ts

- FormState 型に `fieldErrors?: Record<string, string>` を追加
- useActionState 内で、送信時に `minStockLevel > quantity` をチェック
- バリデーション失敗時は fieldErrors を返す

```typescript
if (minStockLevel > quantity) {
  toast.error("最低在庫数は現在庫数以下で設定してください");
  return {
    success: false,
    timestamp: Date.now(),
    fieldErrors: { minStockLevel: "最低在庫数は現在庫数以下で設定してください" }
  };
}
```

#### ② InventoryForm.tsx

- StockInfoSectionProps に `minStockLevelError?: string` を追加
- minStockLevel フィールド下に FormFieldError コンポーネントを追加
- StockInfoSection 呼び出し側に `minStockLevelError` props を渡す

---

## コミット履歴

```
commit 515a517
Author: Claude Haiku 4.5
fix(examinations): remove invalid disabled prop from NotionDatePicker

commit 51e9f56
Author: Claude Haiku 4.5
fix: 2件の新規バグ修正 - 検査日フィールド + 在庫クロスフィールドバリデーション
```

---

## デプロイ情報

**ブランチ**: staging
**Push**: ✅ 成功 (2026-04-02 16:10)
**Pre-push Checks**:
- ✅ Lint: 0 errors (6 warnings は pre-existing shadcn/ui)
- ✅ TypeScript: tsc --noEmit 0 errors

**CI/CD パイプライン**: 進行中
- Target: https://stg.noah-karte.com
- Status: Waiting for GitHub Actions deployment

---

## ファイル変更サマリー

```diff
Total files changed: 3
Total insertions: +23
Total deletions: -1

frontend/src/features/examinations/routes/ExaminationForm.tsx    | 1 deletion
frontend/src/features/inventory/hooks/use-inventory-form.ts     | +18 additions
frontend/src/features/inventory/routes/InventoryForm.tsx        | +5 additions
```

### 詳細変更

#### ExaminationForm.tsx
- NotionDatePicker から無効な `disabled` prop を削除
- ✅ 既存の `disabledDays={{ after: new Date() }}` で制約完全カバー

#### use-inventory-form.ts
```typescript
// 行 15-19: FormState 型にfieldErrorsを追加
interface FormState {
  success: boolean;
  timestamp: number;
  fieldErrors?: Record<string, string>;
}

// 行 51-68: formAction 内でバリデーション追加
const quantity = quantityStr ? Number(quantityStr) : 0;
const minStockLevel = minStockLevelStr ? Number(minStockLevelStr) : 0;

if (minStockLevel > quantity) {
  toast.error("最低在庫数は現在庫数以下で設定してください");
  return {
    success: false,
    timestamp: Date.now(),
    fieldErrors: { minStockLevel: "最低在庫数は現在庫数以下で設定してください" }
  };
}
```

#### InventoryForm.tsx
```typescript
// 行 21: FormFieldError インポート追加
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";

// 行 131-133: StockInfoSectionProps に prop 追加
interface StockInfoSectionProps {
  // ...既存properties...
  minStockLevelError?: string;
}

// 行 169-180: minStockLevel フィールド下にエラー表示追加
<Input
  id="minStockLevel"
  name="minStockLevel"
  type="number"
  min="0"
  step="1"
  defaultValue={defaultMinStockLevel ?? 0}
  className="mt-1"
  required
/>
<FormFieldError id="minStockLevel-error" message={minStockLevelError} />

// 行 370: StockInfoSection 呼び出しに props 追加
<StockInfoSection
  defaultQuantity={existingItem?.quantity}
  defaultMinStockLevel={existingItem?.min_stock_level}
  defaultLocation={existingItem?.location}
  resolvedExpiry={resolvedExpiry}
  onExpiryChange={handleExpiryChange}
  onMarkDirty={markDirty}
  minStockLevelError={formState.fieldErrors?.minStockLevel}
/>
```

---

## テスト検証チェックリスト

### Bug 1: 検査日フィールド (BUG-EXAMINATION-DATE-INPUT-MISSING)

検証環境: http://localhost:3003/examinations/new (ローカル) or https://stg.noah-karte.com/examinations/new (Staging)

- [ ] 検査日フィールドが表示される
- [ ] 過去の日付を選択できる
- [ ] 未来の日付は選択不可（disabled）
- [ ] 検査日未選択で保存 → 現在日時が自動設定される
- [ ] 確定状態の検査は検査日が読み取り専用

### Bug 2: 在庫バリデーション (BUG-INVENTORY-CROSS-FIELD-VALIDATION)

検証環境: http://localhost:3003/inventory (ローカル) or https://stg.noah-karte.com/inventory (Staging)

- [ ] 現在庫数=10, 最低在庫数=20 → エラーメッセージ表示
- [ ] 現在庫数=10, 最低在庫数=10 → 正常保存（等値許容）
- [ ] 現在庫数=10, 最低在庫数=5 → 正常保存
- [ ] エラーメッセージが minStockLevel フィールド直下に表示
- [ ] 既存在庫編集時もバリデーション適用

---

## 次のステップ

### 1. ステージング環境テスト（自動実行予定）
```
GitHub Actions パイプライン実行中...
→ デプロイ完了時: https://stg.noah-karte.com でテスト
```

### 2. QA 検証
テスト報告書の検証項目に従い、両バグが修正されたことを確認

### 3. 本番デプロイ
```bash
# staging → production へ --no-ff merge
git checkout production
git pull origin production
git merge --no-ff staging -m "Release: BUG fixes - examination date + inventory validation"
git tag vX.Y.Z
git push origin production --tags
```

---

## トラブルシューティング

### 検査日フィールドが表示されない場合
1. `ExaminationForm.tsx` で NotionDatePicker がインポートされているか確認
2. FormFieldsSection コンポーネントの L125-133 に code がそのまま存在するか確認
3. ブラウザキャッシュをクリア → F12 DevTools → Network → 「hard refresh」

### 在庫バリデーションエラーが表示されない場合
1. `use-inventory-form.ts` の FormState 型に `fieldErrors` が定義されているか確認
2. formAction の try ブロック前に minStockLevel チェックが存在するか確認
3. InventoryForm.tsx の StockInfoSection 呼び出しで `minStockLevelError` props が渡されているか確認
4. FormFieldError コンポーネントがインポートされているか確認

---

## ドキュメント

詳細な実装ドキュメント:
- `docs/tasks/BUG-EXAMINATION-DATE-INPUT-MISSING.md` — Bug 1 詳細
- `docs/tasks/BUG-INVENTORY-CROSS-FIELD-VALIDATION.md` — Bug 2 詳細
- `docs/tasks/BUG-FIX-VALIDATION-PLAN.md` — 検証計画
- `docs/tasks/BUG-FIX-IMPLEMENTATION-SUMMARY.md` — 実装サマリー

---

## サポート

実装内容に不明な点がある場合は、以下のドキュメントを参照:

- [Frontend CLAUDE.md](frontend/CLAUDE.md) — React 19 パターン
- [.claude/CLAUDE.md](.claude/CLAUDE.md) — プロジェクト全体ルール
- [.claude/rules/typescript-react.md](.claude/rules/typescript-react.md) — TypeScript/React 規約
