# マスタ設定ページ一覧

最終更新: 2026-03-18 (実装同期済)

## 凡例

| 項目 | 説明 |
|------|------|
| **フラットパターン** | 単一リスト。親子関係なし。行は `DataTableRow` の並列配置 |
| **階層パターン** | 親行（root）と子行（child）の2段構成。親の展開/折りたたみあり |
| **D&D** | `@dnd-kit` + `useSortableList` による行ドラッグ並び替え |

---

## 設定インデックス (`/settings`)

`MasterSettingsIndex.tsx`

テーブルなし。Notionスタイルのセクショングループ化されたリスト形式。

---

## 1. 医院マスタ (`/settings/clinic`)

**コンポーネント:** `ClinicMasterSettings.tsx`
**API:** `/api/v1/clinics`

テーブルなし。Notionプロパティ形式のフォーム入力。

---

## 2. 診療項目マスタ (`/settings/treatment-items`)

**コンポーネント:** `TreatmentPlanMaster.tsx`
**タブ切替:** `?tab=consultation|examination|procedure|vaccine|checkup`
**API:** `/api/v1/masters/consultations` 他

全タブ共通の `TreatmentTabContent` コンポーネントで描画。

| タブ | ラベル | パターン | テーブルコンポーネント | D&D |
|------|--------|---------|---------------------|-----|
| `consultation` | 診察 | 階層パターン | `DataTable` + `SortableDataTableRow` / `ChildTreatmentRow` | あり（root のみ） |
| `examination` | 検査 | 階層パターン | `DataTable` + `SortableDataTableRow` / `ChildTreatmentRow` | あり（root のみ） |
| `procedure` | 処置 | 階層パターン | `DataTable` + `SortableDataTableRow` / `ChildTreatmentRow` | あり（root のみ） |
| `vaccine` | 予防接種 | 階層パターン | `DataTable` + `SortableDataTableRow` / `ChildTreatmentRow` | あり（root のみ） |
| `checkup` | 定期健診 | 階層パターン | `DataTable` + `SortableDataTableRow` / `ChildTreatmentRow` | あり（root のみ） |

**テーブルカラム（全タブ共通）:**

| カラム | 幅 | 備考 |
|--------|----|------|
| D&Dハンドル | 32px | root行のみ |
| 名称 | flex-1 | 展開トグル付き（子ありの場合） |
| 単価(税込) | 120px | 右揃え |
| ステータス | 100px | 中央揃え、`NotionStatusPill` |
| 操作 | 80px | 右揃え、`RowActionButton` |

**階層構造:**
- root 行: `SortableDataTableRow`（D&D対象）
- child 行: `ChildTreatmentRow`（D&D対象外、インデント表示）

---

## 3. 診断マスタ (`/settings/diagnosis`)

**コンポーネント:** `DiagnosisSettings.tsx`
**タブ切替:** `?tab=diagnosis_category|diagnosis_name`
**API:** `/api/v1/masters/diagnosis-categories` / `/api/v1/masters/diagnosis-names`

### タブ1: 診断病名カテゴリ (`diagnosis_category`)

| 項目 | 内容 |
|------|------|
| パターン | フラットパターン |
| D&D | あり |

**カラム:** D&Dハンドル | カテゴリ名 | 備考 | ステータス | 操作

### タブ2: 診断病名 (`diagnosis_name`)

| 項目 | 内容 |
|------|------|
| パターン | フラットパターン（カテゴリ名を列表示） |
| D&D | あり |

**カラム:** D&Dハンドル | 所属カテゴリ | 診断病名 | ステータス | 操作

---

## 4. トリミングマスタ (`/settings/trimming`)

**コンポーネント:** `TrimmingSettings.tsx`
**タブ切替:** `course` | `option`
**API:** `/api/v1/masters/trimming-courses` / `/api/v1/masters/trimming-options`

### タブ1: トリミングコース

| 項目 | 内容 |
|------|------|
| パターン | フラットパターン |
| D&D | あり |

**カラム:** D&Dハンドル | コース名 | 対象サイズ | 所要時間 | 単価(税込) | ステータス | 操作

### タブ2: トリミングオプション

| 項目 | 内容 |
|------|------|
| パターン | フラットパターン |
| D&D | あり |

**カラム:** D&Dハンドル | オプション名 | 所要時間 | 組合せ可否 | 単価(税込) | ステータス | 操作

---

## 5. 薬剤マスタ (`/settings/medicine`)

**コンポーネント:** `MedicineSettings.tsx`
**API:** `/api/v1/masters/medicines`

タブなし。カテゴリ > 薬剤の2段階UI。独自 `Table` 実装。

---

## 6. 予約区分マスタ (`/settings/service-type`)

**コンポーネント:** `ServiceTypeSettings.tsx`
**API:** `/api/v1/masters/service-types`

| 項目 | 内容 |
|------|------|
| パターン | フラットパターン |
| D&D | あり |

**カラム:** D&Dハンドル | 名称(カラードット付) | 備考 | ステータス | 操作

---

## 7. スタッフマスタ (`/settings/staff`)

**コンポーネント:** `StaffSettings.tsx`
**API:** `/api/v1/masters/staffs`

| 項目 | 内容 |
|------|------|
| パターン | フラットパターン |
| D&D | あり |

**カラム:** D&Dハンドル | 氏名 | 職種 | ステータス | 操作

---

## 8. 入院マスタ (`/settings/hospitalization`)

**コンポーネント:** `HospitalizationSettings.tsx`
**API:** `/api/v1/masters/hospitalization-plans`

| 項目 | 内容 |
|------|------|
| パターン | フラットパターン |
| D&D | あり |

---

## 9. ケージマスタ (`/settings/cage`)

**コンポーネント:** `CageSettings.tsx`
**API:** `/api/v1/masters/cages`

| 項目 | 内容 |
|------|------|
| パターン | フラットパターン |
| D&D | あり |

---

## 10. 動物種マスタ (`/settings/animal-species`)

**コンポーネント:** `AnimalSpeciesSettings.tsx`
**API:** `/api/v1/masters/animal-species`

---

## 11. 職能マスタ (`/settings/job-title`)

**コンポーネント:** `JobTitleSettings.tsx`
**API:** `/api/v1/masters/job-titles`

---

## 12. 主訴カテゴリマスタ (`/settings/interview/chief-complaint`)

**コンポーネント:** `ChiefComplaintSettings.tsx`
**API:** `/api/v1/masters/chief-complaints`

---

## 13. 問診テンプレート (`/settings/interview/templates`)

**コンポーネント:** `InterviewTemplateSettings.tsx`
**API:** `/api/v1/masters/inquiry-templates`

---

## 14. 保険マスタ (`/settings/insurance`)

**コンポーネント:** `InsuranceSettings.tsx`
**API:** `/api/v1/masters/insurances`

---

## 共有コンポーネント早見表

| コンポーネント | パス |
|--------------|------|
| `DataTable` | `@/components/shared/DataTable` |
| `MasterSidePanel` | `@/components/shared/SidePeek` |
| `NotionStatusPill` | `@/components/shared/StatusPill` |
| `PropertyRow` | `@/components/shared/SidePeek` |
| `PropInput` | `@/components/shared/SidePeek` |
| `MoneyInput` | `@/components/shared/SidePeek` |
| `StatusToggleButton` | `@/components/shared/SidePeek` |
| `RowActionButton` | `@/components/shared/RowActionButton` |
