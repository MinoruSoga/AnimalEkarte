# FE-018: 全一覧ページを NotionFilter に移行

**Status**: Open
**Priority**: Medium
**Affects**: 全9一覧ページ
**Date Created**: 2026-03-17
**Related**: TASK-003, FE-017

## Summary

FE-017 で作成した `NotionFilter` 共有コンポーネントを使って、全9一覧ページの既存フィルタUI（SearchFilterBar + 個別 Select/DatePicker）を置き換える。各ページごとにフィルタプロパティ定義を作成し、NotionFilter に渡す。

## 現状のコード

各ページの現在のフィルタ実装:

| ページ | ファイル | 現在のフィルタUI | 行番号 |
|--------|---------|----------------|--------|
| 飼主一覧 | `features/owners/routes/OwnersList.tsx` | SearchFilterBar | 256-261 |
| カルテ一覧 | `features/medical-records/routes/MedicalRecords.tsx` | SearchFilterBar | 133-138 |
| 入院一覧 | `features/hospitalization/routes/HospitalizationList.tsx` | SearchFilterBar + Tabs + ToggleGroup | 50-79 |
| 健康診断一覧 | `features/examinations/routes/Examinations.tsx` | SearchFilterBar + DateRangePicker | 82-93 |
| 会計一覧 | `features/accounting/routes/Accounting.tsx` | SearchFilterBar + Select | 156-176 |
| 予防接種一覧 | `features/vaccinations/routes/VaccinationList.tsx` | SearchFilterBar + DateRangePicker | 80-91 |
| トリミング一覧 | `features/trimming/routes/TrimmingList.tsx` | SearchFilterBar + NotionDatePicker×2 | 203-233 |
| 在庫一覧 | `features/inventory/routes/InventoryList.tsx` | SearchFilterBar + Select×2 | 121-160 |
| 予約管理 | `features/reservations/routes/ReservationManagement.tsx` | Select（担当医） | 163-191 |

## 必要な変更

### 各ページのフィルタプロパティ定義

#### 1. 飼主一覧（owners）

```typescript
// features/owners/routes/OwnersList.tsx
const OWNER_FILTER_PROPERTIES: FilterProperty[] = [];
// テキスト検索のみ — 追加フィルタプロパティなし
// NotionFilter の searchTerm / onSearchChange のみ使用
```

#### 2. カルテ一覧（medical-records）

```typescript
const MEDICAL_RECORD_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "status",
    label: "ステータス",
    type: "select",
    icon: CircleDot,
    options: [
      { value: "in_progress", label: "診察中" },
      { value: "completed", label: "完了" },
      { value: "cancelled", label: "キャンセル" },
    ],
  },
];
```

#### 3. 入院一覧（hospitalization）

```typescript
const HOSPITALIZATION_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "status",
    label: "ステータス",
    type: "select",
    icon: CircleDot,
    options: [
      { value: "hospitalized", label: "入院中" },
      { value: "reserved", label: "予約" },
      { value: "discharged", label: "退院済" },
    ],
  },
];
// 表示モード切替（board/list）は NotionFilter の外に残す
```

#### 4. 健康診断一覧（examinations）

```typescript
const EXAMINATION_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "date",
    label: "日付",
    type: "date-range",
    icon: Calendar,
  },
  {
    key: "status",
    label: "ステータス",
    type: "select",
    icon: CircleDot,
    options: [
      { value: "pending", label: "依頼中" },
      { value: "in_progress", label: "検査中" },
      { value: "completed", label: "完了" },
    ],
  },
];
```

#### 5. 会計一覧（accounting）

```typescript
const ACCOUNTING_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "status",
    label: "ステータス",
    type: "select",
    icon: CircleDot,
    options: [
      { value: "waiting", label: "会計待ち" },
      { value: "completed", label: "会計済" },
      { value: "cancelled", label: "キャンセル" },
    ],
  },
  {
    key: "date",
    label: "日付",
    type: "date-range",
    icon: Calendar,
  },
];
```

#### 6. 予防接種一覧（vaccinations）

```typescript
const VACCINATION_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "date",
    label: "日付",
    type: "date-range",
    icon: Calendar,
  },
];
```

#### 7. トリミング一覧（trimming）

```typescript
const TRIMMING_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "date",
    label: "日付",
    type: "date-range",
    icon: Calendar,
  },
];
```

#### 8. 在庫一覧（inventory）

```typescript
const INVENTORY_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "category",
    label: "カテゴリ",
    type: "select",
    icon: Package,
    options: [
      { value: "medicine", label: "医薬品" },
      { value: "consumable", label: "消耗品" },
      { value: "food", label: "フード" },
      { value: "other", label: "その他" },
    ],
  },
  {
    key: "status",
    label: "ステータス",
    type: "select",
    icon: CircleDot,
    options: [
      { value: "sufficient", label: "十分" },
      { value: "low", label: "残少" },
      { value: "out", label: "在庫切れ" },
    ],
  },
];
```

#### 9. 予約管理（reservations）

```typescript
const RESERVATION_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "doctor",
    label: "担当医",
    type: "select",
    icon: User,
    options: [], // 動的生成（API から取得した医師リスト）
  },
];
// カレンダーUIのナビゲーション（月/週切替、日付移動）は NotionFilter の外に残す
```

### 移行パターン（各ページ共通）

```typescript
// Before:
const [searchTerm, setSearchTerm] = useState("");
const [statusFilter, setStatusFilter] = useState("all");
const deferredSearch = useDeferredValue(searchTerm);

<SearchFilterBar
  searchTerm={searchTerm}
  onSearchChange={setSearchTerm}
  placeholder="飼主名、ペット名..."
  count={filteredRecords.length}
>
  <Select value={statusFilter} onValueChange={setStatusFilter}>
    ...
  </Select>
</SearchFilterBar>

// After:
const [searchTerm, setSearchTerm] = useState("");
const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
const deferredSearch = useDeferredValue(searchTerm);

<NotionFilter
  properties={PAGE_FILTER_PROPERTIES}
  activeFilters={activeFilters}
  onFilterChange={setActiveFilters}
  searchTerm={searchTerm}
  onSearchChange={setSearchTerm}
  searchPlaceholder="飼主名、ペット名..."
  count={filteredRecords.length}
/>
```

### activeFilters → API/クライアントフィルタへの変換

```typescript
// ヘルパー hook を作成して activeFilters を各ページのフィルタ形式に変換
// 例: 会計一覧
const statusFilter = activeFilters.find((f) => f.key === "status")?.value as string | undefined;
const dateFilter = activeFilters.find((f) => f.key === "date")?.value as { from?: string; to?: string } | undefined;
```

## UI 操作フロー

1. ユーザーが一覧画面を開く
2. テキスト検索欄に入力 → 即時フィルタ（useDeferredValue）
3. 「+ フィルタを追加」をクリック → Popover が開く
4. 「ステータス」を選択 → 値選択 Popover に遷移
5. 「会計待ち」を選択 → ピルが表示される `[ステータス: 会計待ち ×]`
6. さらに「+ フィルタを追加」→「日付」→ 日付範囲を選択
7. 2つのフィルタが AND で適用される
8. ピルの `×` をクリック → そのフィルタが削除される

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useDeferredValue` でテキスト検索遅延（既存パターン維持）
- [ ] `useMemo` でフィルタプロパティ定義を安定化（モジュール定数）
- [ ] SearchFilterBar の import を全ページから削除

## 依存関係

- **FE-017** が先に完了している必要がある（NotionFilter コンポーネントが必要）
- Backend 変更不要（API フィルタパラメータは対応済み）

## 完了条件

- [ ] 全9一覧ページで NotionFilter を使用
- [ ] 各ページに FilterProperty 定義を配置
- [ ] SearchFilterBar の import が全ページから削除されている
- [ ] テキスト検索が全ページで動作
- [ ] ステータスフィルタがピルで表示・削除可能
- [ ] 日付範囲フィルタがピルで表示・削除可能
- [ ] 複数フィルタの AND 適用が動作
- [ ] 入院の表示モード切替、予約のカレンダーUIは NotionFilter 外に残存
- [ ] 型エラーなし（`docker compose exec frontend npm run build` パス）
- [ ] ESLint エラーなし（`docker compose exec frontend npm run lint` パス）
