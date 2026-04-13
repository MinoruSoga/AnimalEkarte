# マスタ設定ページ一覧

最終更新: 2026-04-13 (最新実装同期済)

## 凡例

| 項目 | 説明 |
|------|------|
| **フラットパターン** | 単一リスト。親子関係なし。行は `DataTableRow` の並列配置 |
| **階層パターン** | 親行（root）と子行（child）の2段構成。親の展開/折りたたみあり |
| **D&D** | `@dnd-kit` + `useSortableList` による行ドラッグ並び替え |

---

## 設定インデックス (`/settings`)

`MasterSettingsIndex.tsx`

テーブルなし。Notionスタイルのセクショングループ化されたリスト形式。権限（`usePermission`）によりカードの表示/非表示を動的に制御。

---

## 1. 医院マスタ (`/settings/clinic`)

**コンポーネント:** `ClinicMasterSettings.tsx`
**API:** `/api/v1/clinics`

テーブルなし。Notionプロパティ形式のフォーム入力。

---

## 2. 権限グループマスタ (`/settings/permission-groups`)

**コンポーネント:** `PermissionGroupSettings.tsx`
**API:** `/api/v1/permission-groups`

権限グループの名称、色、および各リソースへのCRUD権限ルールを管理する。マルチテナント（`clinic_id`）分離。

---

## 3. スタッフマスタ (`/settings/staff`)

**コンポーネント:** `StaffSettings.tsx`
**API:** `/api/v1/masters/staffs`

スタッフの基本情報、職種（`occupations`）、所属クリニック、権限グループ、およびアカウント（Email/Pass）を管理する。
LINE予約用の表示設定（名称、画像、説明、表示フラグ）および対応不可予約区分の設定を含む。

---

## 4. LINE予約設定 (`/line-reservation/settings` / `/line-reservation/page-editor`)

**コンポーネント:** `LineReservationSettings.tsx`, `LineReservationPageEditor.tsx`
**API:** `/api/v1/clinics/:id/line-reservation-settings`

LINE公式アカウント連携および予約ページの表示内容・営業ルールを管理する。

---

## 5. 診療項目マスタ (`/settings/treatment-items`)

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

---

## 6. 診断マスタ (`/settings/diagnosis`)

**コンポーネント:** `DiagnosisSettings.tsx`
**タブ切替:** `?tab=diagnosis_category|diagnosis_name`
**API:** `/api/v1/masters/diagnosis-categories` / `/api/v1/masters/diagnosis-names`

### タブ1: 診断病名カテゴリ (`diagnosis_category`)
### タブ2: 診断病名 (`diagnosis_name`)

---

## 7. トリミングマスタ (`/settings/trimming`)

**コンポーネント:** `TrimmingSettings.tsx`
**タブ切替:** `course` | `option`
**API:** `/api/v1/masters/trimming-courses` / `/api/v1/masters/trimming-options`

---

## 8. 薬剤マスタ (`/settings/medicine`)

**コンポーネント:** `MedicineSettings.tsx`
**API:** `/api/v1/masters/medicines`

タブなし。カテゴリ > 薬剤の2段階UI。独自 `Table` 実装。

---

## 9. 予約区分マスタ (`/settings/reservation-type`)

**コンポーネント:** `ReservationTypeSettings.tsx`
**API:** `/api/v1/masters/reservation-types`

| 項目 | 内容 |
|------|------|
| パターン | フラットパターン |
| D&D | あり |

**カラム:** D&Dハンドル | 名称(カラードット付) | 備考 | ステータス | 操作

---

## 10. 入院マスタ (`/settings/hospitalization`)

**コンポーネント:** `HospitalizationSettings.tsx`
**API:** `/api/v1/masters/hospitalization-plans`

---

## 11. ケージマスタ (`/settings/cage`)

**コンポーネント:** `CageSettings.tsx`
**API:** `/api/v1/masters/cages`

---

## 12. 動物種マスタ (`/settings/animal-species`)

**コンポーネント:** `AnimalSpeciesSettings.tsx`
**API:** `/api/v1/masters/animal-species`

---

## 13. 職種マスタ (`/settings/occupations`)

**コンポーネント:** `OccupationSettings.tsx`
**API:** `/api/v1/masters/occupations`

---

## 14. 主訴カテゴリマスタ (`/settings/interview/chief-complaint`)

**コンポーネント:** `ChiefComplaintSettings.tsx`
**API:** `/api/v1/masters/chief-complaint-categories`

---

## 15. 問診テンプレート (`/settings/interview/templates`)

**コンポーネント:** `InterviewTemplateSettings.tsx`
**API:** `/api/v1/masters/inquiry-templates`

---

## 16. 保険マスタ (`/settings/insurance`)

**コンポーネント:** `InsuranceSettings.tsx`
**API:** `/api/v1/masters/insurances`

---

## 17. 商品マスタ (`/settings/merchandise-items`)

**コンポーネント:** `MerchandiseItemSettings.tsx`
**API:** `/api/v1/masters/merchandise-items`

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
| `StatusToggleButton` | `@/components/shared/StatusPill` |
| `RowActionButton` | `@/components/shared/RowActionButton` |
