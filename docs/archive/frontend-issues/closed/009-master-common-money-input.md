---
status: closed
closed_at: 2026-03-16
---

# [master] 単価入力フィールドが 6 ファイルで重複定義（MoneyInput 共通化）

## 優先度
中

## 種別
冗長コード / DRY 原則違反

## 対象ファイル
- `frontend/src/features/master/routes/TrimmingSettings.tsx`（TrimmingCourseSidePanel, TrimmingOptionSidePanel）
- `frontend/src/features/master/routes/HospitalizationSettings.tsx`
- `frontend/src/features/master/routes/CageSettings.tsx`
- `frontend/src/features/master/routes/TreatmentPlanMaster.tsx`
- `frontend/src/features/master/routes/MedicineSettings.tsx`
- `frontend/src/features/master/routes/Settings.tsx`

## 問題

`¥` プレフィックス付きの金額入力フィールドが 6 ファイルに重複して定義されている。

```tsx
// 全ファイルで重複しているパターン
<PropertyRow label="単価(税込)">
  <div className="flex items-center gap-1">
    <span className="text-sm text-muted-foreground">¥</span>
    <input
      type="number"
      min="0"
      value={formData.price === 0 ? "" : String(formData.price)}
      onChange={(e) => setFormData({ ...formData, price: Number(e.target.value) || 0 })}
      className="w-24 rounded border px-2 py-1 text-right text-sm"
      placeholder="0"
    />
  </div>
</PropertyRow>
```

さらに数値/文字列の扱いがファイルによって異なる（`price: number` vs `price: string`）。

## 修正方針

`@/components/shared/SidePeek/` に `MoneyInput` コンポーネントを作成する。

```tsx
// 新規作成: @/components/shared/SidePeek/MoneyInput.tsx
interface MoneyInputProps {
  label?: string;
  value: number;
  onChange: (value: number) => void;
  placeholder?: string;
}

export function MoneyInput({ label = "単価(税込)", value, onChange, placeholder = "0" }: MoneyInputProps) {
  return (
    <PropertyRow label={label}>
      <div className="flex items-center gap-1">
        <span className="text-sm text-muted-foreground">¥</span>
        <input
          type="number"
          min="0"
          value={value === 0 ? "" : String(value)}
          onChange={(e) => onChange(Number(e.target.value) || 0)}
          className="w-24 rounded border px-2 py-1 text-right text-sm"
          placeholder={placeholder}
        />
      </div>
    </PropertyRow>
  );
}
```

`value` は `number` に統一する（`string` で管理しているファイルは変換ロジックを `MoneyInput` 内に封じ込める）。
