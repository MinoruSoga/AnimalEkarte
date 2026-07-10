# 入院・宿泊プランマスタ 仕様書 (Hospitalization Plans)

## 概要
- **画面の目的**: 入院治療やペットホテルの基本料金体系、および初期ケアセットの管理。
- **URLパターン**: `/settings/hospitalization`
- **アクセス権限**: 設備マスタ管理権限が必要（`ResourceMasterHospitalization`）

---

## 1. 画面構成

### 1.1 プラン一覧
- **表示項目**: プラン名、対象体格、料金単位、単価、有効/無効ステータス。

### 1.2 詳細編集サイドパネル (`SidePeekPanel`)
- **プラン名称**: 例：「一般入院」「ICU管理」等（自由入力）。
- **対象体格（`bodySize`）**: 小型、中型、大型。
- **料金単位（`billingUnit`）**: 日単位（`per_day`）/ 泊単位（`per_night`）。
- **単価・課税区分・税率**、**備考**。
- **初期ケアセット**: このモデルにはケアタスクの自動展開機能は存在しない（`HospitalizationPlan` に紐づくタスクテンプレートのフィールドはない）。

---

## 主要な機能

### 1. 会計への自動集計
退院処理実行時、入院日数（または泊数）とこのマスタの単価を掛け合わせ、自動的に会計明細が作成されます。

### 2. ケアプランとの関連（別画面）
`CarePlanItem`（ケアプランタスク）は任意で `hospitalization_plan_id` を持てるが、これは入院登録時にタスクを自動展開する機能ではなく、本設定画面のサイドパネルからも編集できない。

---

## 技術仕様

### 3.1 構成コンポーネント
- **`HospitalizationSettings`**: メインページ。
- **`HospitalizationSidePanel`**: 対象体格・料金単位・単価・課税区分の編集。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/hospitalization-plans` | 有効なプラン一覧の取得 | `master-hospitalization` | `view` |
| GET | `/api/v1/masters/hospitalization-plans/:id` | 特定の入院プラン詳細の取得 | `master-hospitalization` | `view` |
| POST | `/api/v1/masters/hospitalization-plans` | 新規入院プランの登録 | `master-hospitalization` | `create` |
| PATCH | `/api/v1/masters/hospitalization-plans/:id` | 名称・体格・料金単位・単価・課税区分の更新 | `master-hospitalization` | `edit` |
| DELETE | `/api/v1/masters/hospitalization-plans/:id` | 入院プランの削除 | `master-hospitalization` | `delete` |
| PATCH | `/api/v1/masters/hospitalization-plans/reorder` | 表示順序の一括保存 | `master-hospitalization` | `edit` |

---

