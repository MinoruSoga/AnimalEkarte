# [master] SidePeek の組み立てボイラープレートを MasterSidePanel に共通化せよ

## 優先度
中

## 種別
冗長コード / DRY 原則違反

## 対象ファイル（11箇所で重複）
- `frontend/src/features/master/routes/DiagnosisSettings.tsx`（×2: カテゴリ + 病名）
- `frontend/src/features/master/routes/TrimmingSettings.tsx`（×2: コース + オプション）
- `frontend/src/features/master/routes/StaffSettings.tsx`
- `frontend/src/features/master/routes/CageSettings.tsx`
- `frontend/src/features/master/routes/HospitalizationSettings.tsx`
- `frontend/src/features/master/routes/Settings.tsx`
- `frontend/src/features/master/routes/ServiceTypeSettings.tsx`
- `frontend/src/features/master/routes/MedicineSettings.tsx`
- `frontend/src/features/master/routes/TreatmentPlanMaster.tsx`

## 問題

`@/components/shared/SidePeek/` に `SidePeekPanel`・`SidePeekToolbar`・`SidePeekBody`・`SidePeekTitleInput`・`SidePeekFooter` のビルディングブロックは存在するが、
それらを組み合わせるラッパーコンポーネントが存在しない。

以下の**同一ボイラープレートが 11 箇所に重複**している：

```tsx
// DiagnosisSettings, TrimmingSettings, Settings, ... 全ファイルで同一
<SidePeekPanel>
  <SidePeekToolbar onClose={onClose} onDelete={...} />
  <SidePeekBody>
    <SidePeekTitleInput value={formData.name} onChange={...} />
    {/* ← ここだけ各ページ固有 */}
  </SidePeekBody>
  <SidePeekFooter onCancel={onClose} onSave={() => onSave(formData)} />
</SidePeekPanel>
```

さらに各ファイルで `SidePeekPanel` 〜 `SidePeekFooter` の 5 コンポーネントを個別にインポートする行が繰り返されており、
1 コンポーネントの props 変更（例: `SidePeekToolbar` に新 prop 追加）が全 11 ファイルの修正を要求する。

## 修正方針

### 1. `MasterSidePanel` を新規作成

`@/components/shared/SidePeek/MasterSidePanel.tsx` を作成し、共通構造をカプセル化する。

```tsx
interface MasterSidePanelProps {
  title: string;
  onTitleChange: (value: string) => void;
  onClose: () => void;
  onSave: () => void;
  onDelete?: () => void;
  isPending?: boolean;
  children: ReactNode; // ← ページ固有フィールド
}

export function MasterSidePanel({
  title,
  onTitleChange,
  onClose,
  onSave,
  onDelete,
  isPending = false,
  children,
}: MasterSidePanelProps) {
  return (
    <SidePeekPanel>
      <SidePeekToolbar onClose={onClose} onDelete={onDelete} />
      <SidePeekBody>
        <SidePeekTitleInput value={title} onChange={onTitleChange} />
        {children}
      </SidePeekBody>
      <SidePeekFooter onCancel={onClose} onSave={onSave} isPending={isPending} />
    </SidePeekPanel>
  );
}
```

### 2. 各マスタページの Side Panel を置き換え

```tsx
// 変更前（各ページで繰り返し）
import { SidePeekPanel } from "@/components/shared/SidePeek/SidePeekPanel";
import { SidePeekToolbar } from "@/components/shared/SidePeek/SidePeekToolbar";
import { SidePeekBody } from "@/components/shared/SidePeek/SidePeekBody";
import { SidePeekTitleInput } from "@/components/shared/SidePeek/SidePeekTitleInput";
import { SidePeekFooter } from "@/components/shared/SidePeek/SidePeekFooter";

<SidePeekPanel>
  <SidePeekToolbar onClose={onClose} onDelete={onDelete} />
  <SidePeekBody>
    <SidePeekTitleInput value={formData.name} onChange={...} />
    {/* fields */}
  </SidePeekBody>
  <SidePeekFooter onCancel={onClose} onSave={() => onSave(formData)} />
</SidePeekPanel>

// 変更後
import { MasterSidePanel } from "@/components/shared/SidePeek/MasterSidePanel";

<MasterSidePanel
  title={formData.name}
  onTitleChange={(v) => setFormData({ ...formData, name: v })}
  onClose={onClose}
  onSave={() => onSave(formData)}
  onDelete={onDelete}
>
  {/* ページ固有フィールドのみ */}
</MasterSidePanel>
```

## 確認事項

- `SidePeekFooter` が現状 `isPending` prop を受け取っていない場合は追加する（`useTransition` 対応のため）
- 各ページで `SidePeekToolbar` に渡している props（`onDelete` 以外に何があるか）を確認してから interface を確定する
- `MasterSidePanel` は `features/master/` 専用ではなく `components/shared/SidePeek/` に配置してよい（他 feature でも同パターンが使われる可能性がある）
