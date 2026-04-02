# FE-080: Vercel React Best Practices 違反修正 (Medium Priority)

## 背景

全 `src/` フォルダの Vercel React Best Practices 監査で検出された、リリース前に修正すべき違反5件。
ネストされたコンポーネント定義と、useCallback deps へのオブジェクト混入が主な問題。

## 依存

- FE-079 完了後に着手（同一ファイルの重複修正を避けるため）

## 要件

### 1. `components/shared/Layout/Sidebar.tsx` — SidebarItem 抽出 + memo

**違反ルール**: `rerender-memo`

`SidebarItem`（約110行）がコンポーネント内にネスト定義されている。
モジュールレベルに抽出し `memo()` で包む。

```typescript
// before: Sidebar コンポーネント内にネスト
export function Sidebar() {
  const SidebarItem = ({ item, collapsed, level }: SidebarItemProps) => { ... };
  // ...
}

// after: モジュールレベルに抽出 + memo
const SidebarItem = memo(function SidebarItem({ item, collapsed = false, level = 0 }: SidebarItemProps) {
  // ... (既存ロジック維持)
});

export function Sidebar() {
  // SidebarItem を直接使用
}
```

### 2. `components/shared/ReservationFormModal/ReservationFormModal.tsx` — 子コンポーネント抽出

**違反ルール**: `rendering-hoist-jsx`, `rerender-memo`

`StepIndicator`（行40-53）と `SelectedPetChip`（行55-74）がコンポーネント内に定義されている。
モジュールレベルに巻き上げて `memo()` で包む。

```typescript
// before: ReservationFormModal 内に定義
export function ReservationFormModal({ ... }) {
  function StepIndicator({ ... }) { ... }
  function SelectedPetChip({ ... }) { ... }
  // ...
}

// after: モジュールレベルに抽出
const StepIndicator = memo(function StepIndicator({ ... }: StepIndicatorProps) { ... });
const SelectedPetChip = memo(function SelectedPetChip({ ... }: SelectedPetChipProps) { ... });

export function ReservationFormModal({ ... }) {
  // StepIndicator, SelectedPetChip を直接使用
}
```

### 3. `features/dashboard/hooks/use-dashboard-kanban.ts` — filteredColumns を deps から除外

**違反ルール**: `rerender-dependencies`

`moveCard` の deps に `filteredColumns`（オブジェクト配列）が含まれている。
`useRef` で最新値を保持し deps から除外する。

```typescript
// before (行119, 199)
const moveCard = useCallback((...) => {
  const sourceColFiltered = filteredColumns.find(...);
  // ...
}, [filteredColumns, updateStatusMutation]);

// after
const filteredColumnsRef = useRef(filteredColumns);
useEffect(() => { filteredColumnsRef.current = filteredColumns; }, [filteredColumns]);

const moveCard = useCallback((...) => {
  const sourceColFiltered = filteredColumnsRef.current.find(...);
  // ...
}, [updateStatusMutation]);
```

### 4. `features/shifts/components/ShiftCalendar/ShiftCalendar.tsx` — deps オブジェクト分解

**違反ルール**: `rerender-dependencies`

useCallback deps に `editShift`（オブジェクト）と `form`（オブジェクト）が含まれている。
プリミティブに分解するか、`useRef` パターンを適用する。

```typescript
// before (行122)
const handleSave = useCallback((...) => {
  // ...
}, [isEdit, editShift, form, staffId, date, updateShift, createShift, onClose, startSaveTransition]);

// after: useRef パターン
const editShiftRef = useRef(editShift);
useEffect(() => { editShiftRef.current = editShift; }, [editShift]);
const formRef = useRef(form);
useEffect(() => { formRef.current = form; }, [form]);

const handleSave = useCallback((...) => {
  const currentEditShift = editShiftRef.current;
  const currentForm = formRef.current;
  // ...
}, [isEdit, staffId, date, updateShift, createShift, onClose, startSaveTransition]);
```

## 受入条件

- [ ] `Sidebar.tsx` の `SidebarItem` がモジュールレベルに抽出され `memo()` で包まれている
- [ ] `ReservationFormModal.tsx` の `StepIndicator`, `SelectedPetChip` が同上
- [ ] `use-dashboard-kanban.ts` の `moveCard` deps にオブジェクト配列が含まれていない
- [ ] `ShiftCalendar.tsx` の `handleSave` deps にオブジェクトが含まれていない
- [ ] 各コンポーネントの動作に影響がないこと（UI の見た目・インタラクションが変わらない）
- [ ] `docker compose exec frontend npm run build` が成功
