---
status: closed
closed_at: 2026-03-16
---

# [master] ステータストグルボタンパターンが全サイドパネルで重複定義

## 優先度
中

## 種別
冗長コード / DRY 原則違反

## 対象ファイル
- `frontend/src/features/master/routes/DiagnosisSettings.tsx`（DiagnosisCategorySidePanel, DiagnosisNameSidePanel）
- `frontend/src/features/master/routes/ServiceTypeSettings.tsx`（ServiceTypeSidePanel）
- `frontend/src/features/master/routes/TrimmingSettings.tsx`（TrimmingCourseSidePanel, TrimmingOptionSidePanel）
- その他マスタページのサイドパネル

## 問題

全マスタページのサイドパネルで、以下の「ステータストグルボタン」パターンが完全に同一のコードとして繰り返されている。

```tsx
// 各ファイルで重複している共通パターン
<PropertyRow label="ステータス">
  <button
    type="button"
    onClick={() => onChange({ ...item, isActive: !item.isActive })}
    className="cursor-pointer"
  >
    <NotionStatusPill isActive={item.isActive} />
  </button>
</PropertyRow>
```

現在 5〜6 箇所に同一コードが存在し、`NotionStatusPill` の props 変更時に全箇所を修正する必要がある。

## 修正方針

`@/components/shared/SidePeek/` に `StatusToggleButton` コンポーネントを作成し、各サイドパネルから使用する。

```tsx
// 新規作成: @/components/shared/SidePeek/StatusToggleButton.tsx
interface StatusToggleButtonProps {
  isActive: boolean;
  onToggle: () => void;
}

export function StatusToggleButton({ isActive, onToggle }: StatusToggleButtonProps) {
  return (
    <PropertyRow label="ステータス">
      <button type="button" onClick={onToggle} className="cursor-pointer">
        <NotionStatusPill isActive={isActive} />
      </button>
    </PropertyRow>
  );
}
```

## 影響範囲

上記の全サイドパネルファイルを修正する必要がある。
