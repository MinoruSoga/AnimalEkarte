# マスタ設定ページ一覧

最終更新: 2026-03-15

## 凡例

| 項目 | 説明 |
|------|------|
| **フラットパターン** | 単一リスト。親子関係なし。行は `DataTableRow` の並列配置 |
| **階層パターン** | 親行（root）と子行（child）の2段構成。親の展開/折りたたみあり |
| **D&D** | `@dnd-kit` + `useSortableList` による行ドラッグ並び替え |

---

## 設定インデックス (`/settings`)

`MasterSettingsIndex.tsx`

テーブルなし。カードリスト形式。

---

## 1. 医院マスタ (`/settings/clinic`)

**コンポーネント:** `ClinicMasterSettings.tsx`

テーブルなし。フォーム入力のみ（院名・住所・電話番号等）。

---

## 2. 診療項目マスタ (`/settings/treatment-items`)

**コンポーネント:** `TreatmentPlanMaster.tsx`
**タブ切替:** `?tab=consultation|examination|procedure|vaccine|checkup`
**タブコンポーネント:** `@radix-ui/react-tabs`（TabsPrimitive 直接）

全タブ共通の `TreatmentTabContent` コンポーネントで描画。

| タブ | ラベル | パターン | テーブルコンポーネント | D&D |
|------|--------|---------|---------------------|-----|
| `consultation` | 診察 | 階層パターン | `DataTable` + `SortableTreatmentRow` / `ChildTreatmentRow` | あり（root のみ） |
| `examination` | 検査 | 階層パターン | `DataTable` + `SortableTreatmentRow` / `ChildTreatmentRow` | あり（root のみ） |
| `procedure` | 処置 | 階層パターン | `DataTable` + `SortableTreatmentRow` / `ChildTreatmentRow` | あり（root のみ） |
| `vaccine` | 予防接種 | 階層パターン | `DataTable` + `SortableTreatmentRow` / `ChildTreatmentRow` | あり（root のみ） |
| `checkup` | 定期健診 | 階層パターン | `DataTable` + `SortableTreatmentRow` / `ChildTreatmentRow` | あり（root のみ） |

**テーブルカラム（全タブ共通）:**

| カラム | 幅 | 備考 |
|--------|----|------|
| D&Dハンドル | 32px | GripVertical |
| 名称 | flex-1 | 展開トグル付き（子ありの場合） |
| 単価(税込) | 120px | 右揃え |
| ステータス | 100px | 中央揃え、`NotionStatusPill`（inline定義） |
| 操作 | 80px | 右揃え、`RowActionButton` |

**階層構造:**
- `buildTree()` で flat array → `TreeItem[]`（root + children）に変換
- root 行: `SortableTreatmentRow`（D&D対象）
- child 行: `ChildTreatmentRow`（D&D対象外、インデント表示）
- 展開状態: `expandedIds: Set<string>` で管理

**サイドピーク（`TreatmentItemSidePanel`）フィールド:**

| ラベル | フィールド | 入力UI |
|--------|-----------|--------|
| ステータス | `isActive` | `NotionStatusPill` クリック切替（inline定義） |
| 単価(税込) | `price` | `<input type="number">` |
| 備考 | `description` | `PropInput`（inline定義） |

> **注意:** `checkup` タブは `interval`（健診間隔）・`target_age`（対象年齢）フィールドを表示しない。
> API には対応フィールドが存在するが、`TreatmentTabContent` が汎用設計のため未対応。

---

## 3. 診断マスタ (`/settings/diagnosis`)

**コンポーネント:** `DiagnosisSettings.tsx`
**タブ切替:** `?tab=diagnosis_category|diagnosis_name`
**タブコンポーネント:** `@/components/ui/tabs`（shadcn/ui）

### タブ1: 診断病名カテゴリ (`diagnosis_category`)

| 項目 | 内容 |
|------|------|
| パターン | フラットパターン |
| テーブルコンポーネント | `DataTable` + `SortableCategoryRow`（`DataTableRow` ラップ） |
| D&D | あり |

**カラム:**

| カラム | 幅 | 備考 |
|--------|----|------|
| D&Dハンドル | 32px | |
| カテゴリ名 | flex-1 | |
| 備考 | 240px | |
| ステータス | 100px | 中央揃え、`NotionStatusPill` |
| 操作 | 80px | 右揃え、`RowActionButton` |

**サイドピーク（`DiagnosisCategorySidePanel`）フィールド:**

| ラベル | フィールド | 入力UI |
|--------|-----------|--------|
| ステータス | `isActive` | `NotionStatusPill` クリック切替 |
| 備考 | `description` | `PropInput` |

### タブ2: 診断病名 (`diagnosis_name`)

| 項目 | 内容 |
|------|------|
| パターン | フラットパターン（カテゴリFKを列表示） |
| テーブルコンポーネント | `DataTable` + `SortableNameRow`（`DataTableRow` ラップ） |
| D&D | あり |

**カラム:**

| カラム | 幅 | 備考 |
|--------|----|------|
| D&Dハンドル | 32px | |
| 所属カテゴリ | 160px | カテゴリ名を表示（FK解決） |
| 診断病名 | flex-1 | |
| ステータス | 100px | 中央揃え、`NotionStatusPill` |
| 操作 | 80px | 右揃え、`RowActionButton` |

**サイドピーク（`DiagnosisNameSidePanel`）フィールド:**

| ラベル | フィールド | 入力UI |
|--------|-----------|--------|
| ステータス | `isActive` | `NotionStatusPill` クリック切替 |
| カテゴリ | `diagnosisCategoryId` | `Select`（shadcn/ui） |
| 備考 | `description` | `PropInput` |

---

## 4. トリミングマスタ (`/settings/trimming`)

**コンポーネント:** `TrimmingSettings.tsx`
**タブ切替:** `?tab=` なし（Tabs の value state 管理）
**タブコンポーネント:** `@/components/ui/tabs`（shadcn/ui）

### タブ1: トリミングコース

| 項目 | 内容 |
|------|------|
| パターン | フラットパターン |
| テーブルコンポーネント | `DataTable` + `DataTableRow` |
| D&D | なし |

**カラム:**

| カラム | 幅 | 備考 |
|--------|----|------|
| コース名 | flex-1 | |
| 対象サイズ | 120px | |
| 所要時間 | 100px | |
| 単価(税込) | 110px | 右揃え |
| ステータス | 90px | 右揃え、`NotionStatusPill` |

### タブ2: トリミングオプション

| 項目 | 内容 |
|------|------|
| パターン | フラットパターン |
| テーブルコンポーネント | `DataTable` + `DataTableRow` |
| D&D | なし |

**カラム:**

| カラム | 幅 | 備考 |
|--------|----|------|
| オプション名 | flex-1 | |
| 所要時間 | 100px | |
| 組合せ可否 | 110px | 中央揃え、`CombinablePill`（ローカル定義） |
| 単価(税込) | 110px | 右揃え |
| ステータス | 90px | 右揃え、`NotionStatusPill` |

---

## 5. 薬剤マスタ (`/settings/medicine`)

**コンポーネント:** `MedicineSettings.tsx`

タブなし。カテゴリ > 薬剤の2段階UI。

| 項目 | 内容 |
|------|------|
| パターン | **階層パターン**（カテゴリ行 + 薬剤行の折りたたみ） |
| テーブルコンポーネント | `Table` / `TableHeader` / `TableBody` / `TableRow` / `TableCell`（shadcn/ui 直接） + `DataTableRow`（薬剤行） |
| D&D | あり（`DragOverlay` 付き、`motion/react` アニメーション） |

> `DataTable`（共有コンポーネント）は**使用していない**。`Table` を直接構築する独自実装。

**カラム:**

| カラム | 幅 | 備考 |
|--------|----|------|
| D&Dハンドル | — | カテゴリ行のみ |
| 名称 | flex-1 | カテゴリは太字、薬剤はインデント |
| 剤形 | — | 薬剤行のみ |
| 単価(税込) | — | 右揃え |
| 操作 | — | `DropdownMenu`（MoreHorizontal + Maximize2） |

---

## 6. 予約区分マスタ (`/settings/service-type`)

**コンポーネント:** `ServiceTypeSettings.tsx`

| 項目 | 内容 |
|------|------|
| パターン | フラットパターン |
| テーブルコンポーネント | `DataTable` + `SortableServiceTypeRow`（`DataTableRow` ラップ） |
| D&D | あり |

**カラム:**

| カラム | 幅 | 備考 |
|--------|----|------|
| D&Dハンドル | 32px | |
| 名称 | flex-1 | カラードット付き |
| 備考 | 240px | |
| ステータス | 100px | 中央揃え、`NotionStatusPill` |
| 操作 | 80px | 右揃え、`RowActionButton` |

**サイドピーク（`ServiceTypeSidePanel`）フィールド:**

| ラベル | フィールド | 入力UI |
|--------|-----------|--------|
| ステータス | `isActive` | `NotionStatusPill` クリック切替 |
| カラー | `color` | `<input type="color">` + `PropInput` |
| 備考 | `description` | `PropInput` |

---

## 7. スタッフマスタ (`/settings/staff`)

**コンポーネント:** `StaffSettings.tsx`

| 項目 | 内容 |
|------|------|
| パターン | フラットパターン |
| テーブルコンポーネント | `DataTable` + `DataTableRow` |
| D&D | なし |

**カラム:**

| カラム | 幅 | 備考 |
|--------|----|------|
| 氏名 | flex-1 | |
| 職種 | — | |
| ステータス | — | `NotionStatusPill` |
| 操作 | 80px | `RowActionButton` |

---

## 8〜11. 汎用 Settings.tsx 経由ページ

**コンポーネント:** `Settings.tsx`（`category` prop で動作を切り替え）

| # | ページ | パス | category |
|---|--------|------|----------|
| 8 | 入院マスタ | `/settings/hospitalization` | `hospitalization` |
| 9 | ケージマスタ | `/settings/cage` | `cage` |
| 10 | 職種マスタ | `/settings/job-title` | `job_title` |
| 11 | 保険マスタ | `/settings/insurance` | `insurance` |

全ページ共通:

| 項目 | 内容 |
|------|------|
| パターン | フラットパターン |
| テーブルコンポーネント | `DataTable` + `DataTableRow` |
| D&D | なし |
| データ取得 | `useMasterItems(category)` hook |

表示カラムは `CATEGORY_CONFIG[category]` の `showPrice` / `showCode` / `showCategory` フラグで動的に切り替わる。

---

## 共有コンポーネント早見表

| コンポーネント | パス | 使用ページ |
|--------------|------|-----------|
| `DataTable` | `@/components/shared/DataTable` | treatment-items, diagnosis, service-type, staff, Settings.tsx |
| `DataTableRow` | `@/components/shared/DataTable` | 上記すべて + MedicineSettings（薬剤行のみ） |
| `NotionStatusPill` | `@/components/shared/StatusPill/NotionStatusPill` | diagnosis, service-type, staff, Settings.tsx |
| `PropertyRow` | `@/components/shared/SidePeek/PropertyRow` | diagnosis, service-type |
| `PropInput` | `@/components/shared/SidePeek/PropInput` | diagnosis, service-type |
| `RowActionButton` | `@/components/shared/RowActionButton` | treatment-items, diagnosis, service-type |
| `useSortableList` | `@/hooks/useSortableList` | treatment-items, diagnosis, service-type, MedicineSettings |

> **注意:** `TreatmentPlanMaster.tsx` は `NotionStatusPill`, `PropertyRow`, `PropInput` をファイル内に inline 定義しており、shared コンポーネントを使用していない。共有版への移行が必要。
