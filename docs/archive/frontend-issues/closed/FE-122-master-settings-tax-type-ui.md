# FE-122: マスタ設定 - 課税区分・税率 UI（診察・処置・薬剤・物販）

**Status**: Closed
**Priority**: High
**Affects**: features/master-settings または 各マスタ feature
**Date Created**: 2026-03-25
**Related**: TASK-029, BE-060（先行必須）

## Summary

診察項目・処置項目・薬剤・物販（merchandise_items）の各マスタ設定フォームに「課税区分」「税率」の選択フィールドを追加する。
入院プランも同様に対応する。

## 現状のコード

```typescript
// 各マスタフォームの現状（consultations を例に）
// 現在: name, price, is_active 等のフィールドのみ
// tax_type, tax_rate フィールドなし

// merchandise_items の場合
// 現在: name, unit_price, tax_rate（数値入力）, is_active
// tax_type なし
```

## 必要な変更

### 1. 共有 UI コンポーネントの作成

```typescript
// frontend/src/components/shared/TaxTypeSelector.tsx
// 課税区分選択 Select コンポーネント（全マスタで再利用）

interface TaxTypeSelectorProps {
  value: TaxType;
  onChange: (value: TaxType) => void;
  disabled?: boolean;
}

export function TaxTypeSelector({ value, onChange, disabled }: TaxTypeSelectorProps) {
  return (
    <Select
      value={value}
      onValueChange={(v) => onChange(v as TaxType)}
      disabled={disabled}
    >
      <SelectTrigger>
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="excluded">外税</SelectItem>
        <SelectItem value="included">内税</SelectItem>
        <SelectItem value="exempt">非課税</SelectItem>
      </SelectContent>
    </Select>
  );
}
```

```typescript
// frontend/src/components/shared/TaxRateSelector.tsx
// 税率選択 Select コンポーネント（全マスタで再利用）

interface TaxRateSelectorProps {
  value: number; // 0.10 or 0.08
  onChange: (value: number) => void;
  disabled?: boolean;
}

// 選択肢
// ┌──────────────┐
// │ 10%（通常課税）│
// │  8%（軽減税率）│
// └──────────────┘

export function TaxRateSelector({ value, onChange, disabled }: TaxRateSelectorProps) {
  return (
    <Select
      value={String(value)}
      onValueChange={(v) => onChange(Number(v))}
      disabled={disabled}
    >
      {/* ... */}
    </Select>
  );
}
```

### 2. 型定義の更新（各マスタの api/types.ts）

```typescript
// BE-060 + make codegen 後に models.ts から TaxType が追加される
import type { TaxType } from "@/types/generated/models";

// 各マスタの Create/Update リクエスト型に追加
export interface CreateConsultationRequest {
  name: string;
  price?: number | null;
  tax_type: TaxType;  // 新規追加
  tax_rate: number;   // 新規追加 (0.10 or 0.08)
  is_active: boolean;
  // ...
}

export interface UpdateConsultationRequest {
  name?: string;
  price?: number | null;
  tax_type?: TaxType;  // 新規追加
  tax_rate?: number;   // 新規追加
  // ...
}
```

### 3. 各マスタフォームへの追加

各マスタフォームのコンポーネントに `TaxTypeSelector` + `TaxRateSelector` を追加する。

```typescript
// フォーム表示例（Figmaデザインなし - 既存マスタフォームのスタイルに合わせる）
// ┌────────────────────────────────────────┐
// │ 名前   [______________________________] │
// │ 価格   [______________]                 │
// │ 課税区分  [外税 ▼]                      │
// │ 税率      [10%（通常課税） ▼]           │
// │ 有効       [○] 有効                    │
// │                          [保存] [削除]  │
// └────────────────────────────────────────┘
```

#### 対象フォーム

- 診察項目マスタフォーム
- 処置項目マスタフォーム
- 薬剤マスタフォーム
- 物販・フードマスタフォーム（`merchandise_items`）
- 入院プランマスタフォーム

### 4. merchandise_items の既存 tax_rate 入力を TaxRateSelector に置き換え

```typescript
// 現在: tax_rate を数値入力（<Input type="number">）で受け付けている可能性
// 変更後: TaxRateSelector で 10%/8% を選択式に変更
```

## UI 操作フロー

1. ユーザーがマスタ設定 > 診察項目（など）フォームを開く
2. 「課税区分」セレクト（外税/内税/非課税）が表示される
3. 「税率」セレクト（10%/8%）が表示される
4. 値を選択して保存
5. 一覧に戻ると保存された値が表示される

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし（TaxTypeSelector を直接 import）
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理
- [ ] 型は `models.ts` から導出（`TaxType` 型）
- [ ] 共有コンポーネントは `components/shared/` に配置

## 依存関係

- BE-060 が先に完了している必要がある（各マスタ API に tax_type が追加済みであること）
- `make codegen` で `models.ts` に `TaxType` が含まれていること

## 完了条件

- [ ] 診察・処置・薬剤・物販・入院プランの各マスタフォームに課税区分・税率フィールドが表示される
- [ ] 選択した課税区分・税率が保存・読み込みできる
- [ ] `TaxTypeSelector`, `TaxRateSelector` が共有コンポーネントとして実装されている
- [ ] 非課税を選択しても税率フィールドは表示されている（税額計算は Backend が 0 を返す）
- [ ] `pnpm build` 型エラーなし
- [ ] `pnpm lint` エラーなし
