# FE-107: NumberInput 共有コンポーネント作成

**Status**: Closed
**Priority**: Low
**Affects**: accounting, inventory, estimates, medical-records, hospitalization, master（18ファイル）
**Date Created**: 2026-03-24
**Related**: TASK-024

## Summary

数値入力（単価・数量・体重・在庫数）の `<input type="number">` + 単位サフィックス（円・kg・ml）が 18 ファイルで独立実装されており、スタイルも不統一。共有 `NumberInput` コンポーネントを作成して統一する。

## 現状のコード

```typescript
// frontend/src/features/accounting/routes/AccountingDetail.tsx:410-417
// 金額入力（右揃え・大フォント・"円"サフィックス）
<Input
  type="number"
  className="h-14 text-xl font-bold text-right pr-10"
  value={receivedAmount}
  onChange={(e) => onReceivedAmountChange(e.target.value)}
/>
<span className="absolute right-3 top-4 text-gray-400">円</span>
```

```typescript
// frontend/src/features/hospitalization/components/DailyRecord/VitalDialog.tsx:80
// バイタル入力（step="0.1"・単位サフィックスあり）
<Input type="number" step="0.1" value={form.temperature} onChange={...} />
// （単位ラベルは別要素でハードコード）
```

```typescript
// frontend/src/features/inventory/routes/InventoryForm.tsx
// 在庫数量（基本的な type="number"）
<Input type="number" value={quantity} onChange={...} />
```

```typescript
// frontend/src/features/estimates/routes/EstimateForm.tsx
// 見積数量・単価（type="number"）
<Input type="number" value={quantity} onChange={...} />
```

```typescript
// components/shared/SidePeek/MoneyInput.tsx（既存・金額専用）
// ¥ 記号付きの金額入力（SidePeek 内でのみ使用）
```

## 必要な変更

### 1. NumberInput コンポーネント作成

```typescript
// frontend/src/components/shared/NumberInput/NumberInput.tsx（新規作成）

interface NumberInputProps {
  value: number | string;
  onChange: (value: string) => void;
  placeholder?: string;
  step?: number;
  min?: number;
  max?: number;
  suffix?: string;       // "円", "kg", "ml", "錠" など
  align?: "left" | "right";
  className?: string;
  disabled?: boolean;
  id?: string;
}

export function NumberInput({
  value,
  onChange,
  placeholder = "0",
  step = 1,
  min,
  max,
  suffix,
  align = "left",
  className,
  disabled,
  id,
}: NumberInputProps) {
  return (
    <div className="relative">
      <Input
        id={id}
        type="number"
        step={step}
        min={min}
        max={max}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        disabled={disabled}
        className={`${suffix ? "pr-10" : ""} ${align === "right" ? "text-right" : ""} ${className ?? ""}`}
      />
      {suffix ? (
        <span className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-sm text-[#37352F]/60">
          {suffix}
        </span>
      ) : null}
    </div>
  );
}
```

```typescript
// frontend/src/components/shared/NumberInput/index.ts（新規作成）
export { NumberInput } from "./NumberInput";
```

### 2. 置き換え対象（優先度の高いもの）

```typescript
// accounting/routes/AccountingDetail.tsx:410-417
// Before: Input + 絶対配置 span
// After:
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
<NumberInput value={receivedAmount} onChange={onReceivedAmountChange} suffix="円" align="right" />
```

```typescript
// hospitalization/components/DailyRecord/VitalDialog.tsx
// Before: Input type="number" step="0.1" + 別要素の単位ラベル
// After:
<NumberInput value={form.temperature} onChange={(v) => setForm(...)} step={0.1} suffix="℃" />
```

## 注意事項

- `SidePeek/MoneyInput.tsx` は既存のまま残す（SidePeek 専用ラッパー）。内部で `NumberInput` を使う形に変更してもよい
- `onChange` は `string` を返す（`e.target.value` がstring のため）。呼び出し側で `Number()` 変換は従来通り

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] barrel index 経由 import なし（`NumberInput/NumberInput` を直接 import）
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `FC` / `forwardRef` なし

## 依存関係

- Backend 変更なし。他の FE イシューとも独立。

## 完了条件

- [ ] `frontend/src/components/shared/NumberInput/NumberInput.tsx` が作成されている
- [ ] accounting の受取金額入力が `NumberInput` を使用している
- [ ] hospitalization のバイタル数値入力が `NumberInput` を使用している
- [ ] 数値入力の見た目・動作が変化なし
- [ ] `pnpm lint` エラーなし
- [ ] `pnpm build` エラーなし
