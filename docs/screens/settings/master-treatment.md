# 診療項目マスタ 仕様書 (Treatment Items)

## 概要
- **画面 of Purpose**: 診察、処置、手術、検査等の臨床行為の定義、および標準価格・税率の設定。
- **URLパターン**: `/settings/treatment-items`
- **アクセス権限**: 診療マスタ管理権限が必要（`ResourceMasterMedical`）

---

## 1. 画面構成

### 1.1 項目一覧テーブル
- **カテゴリ分類**: 診察、検査、処置、予防、入院のタブ切り替えにより、膨大な項目を整理して表示。
- **検索**: 名称、略称、診療コードによる部分一致。

### 2. 詳細編集サイドパネル (`SidePeekPanel`)
- **基本属性**: 項目名、略称、診療区分コード。
- **価格・収益設定**:
    - **標準単価**: 税込の基本価格。
    - **変動価格フラグ**: オンにすると、実際のカルテ入力時に個別に価格を変更可能。
- **表示制御**:
    - **カルテで選択可能**: 臨床現場での入力を許可。
    - **見積専用項目**: 高額処置など、事前の概算提示にのみ使用。
- **税区分**: 標準税率 (10%) / 軽減税率 (8%) / 非課税。

---

## 2. 主要な機能

### 2.1 レジ精算への自動連動
カルテ（SOAPS）で選択された処置項目は、ここで設定された単価と税率を保持したまま会計セクションへ自動転送されます。

### 2.2 インボイス制度への対応
税率設定に基づき、領収書上での税率別の売上集計・消費税額が自動的に算出されます。

---

## 3. 技術仕様

### 3.1 構成コンポーネント
- **`TreatmentPlanMaster`**: メインページ。
- **`MasterCategoryTabs`**: 大分類ごとの高速なデータ切り替え。
- **`PropInput`**: 単価、名称、コードのボーダーレス編集。

### API連携
| メソッド | エンドポイント | 用途 |
|:---|:---|:---|
| GET | `/api/v1/masters/consultations` | 診察項目一覧の取得。 |
| POST | `/api/v1/masters/consultations` | 診察項目の作成。 |
| PATCH | `/api/v1/masters/consultations/:id` | 診察項目の更新。 |
| DELETE | `/api/v1/masters/consultations/:id` | 診察項目の削除。 |
| PATCH | `/api/v1/masters/consultations/reorder` | 診察項目の表示順一括保存。 |
| GET | `/api/v1/masters/examination-types` | 検査項目一覧の取得。 |
| POST | `/api/v1/masters/examination-types` | 検査項目の作成。 |
| PATCH | `/api/v1/masters/examination-types/:id` | 検査項目の更新。 |
| DELETE | `/api/v1/masters/examination-types/:id` | 検査項目の削除。 |
| PATCH | `/api/v1/masters/examination-types/reorder` | 検査項目の表示順一括保存。 |
| GET | `/api/v1/masters/procedures` | 処置項目一覧の取得。 |
| POST | `/api/v1/masters/procedures` | 処置項目の作成。 |
| PATCH | `/api/v1/masters/procedures/:id` | 処置項目の更新。 |
| DELETE | `/api/v1/masters/procedures/:id` | 処置項目の削除。 |
| PATCH | `/api/v1/masters/procedures/reorder` | 処置項目の表示順一括保存。 |
| GET | `/api/v1/masters/vaccines` | 予防接種項目一覧の取得。 |
| POST | `/api/v1/masters/vaccines` | 予防接種項目の作成。 |
| PATCH | `/api/v1/masters/vaccines/:id` | 予防接種項目の更新。 |
| DELETE | `/api/v1/masters/vaccines/:id` | 予防接種項目の削除。 |
| PATCH | `/api/v1/masters/vaccines/reorder` | 予防接種項目の表示順一括保存。 |
| GET | `/api/v1/masters/checkup-types` | 定期健診項目一覧の取得。 |
| POST | `/api/v1/masters/checkup-types` | 定期健診項目の作成。 |
| PATCH | `/api/v1/masters/checkup-types/:id` | 定期健診項目の更新。 |
| DELETE | `/api/v1/masters/checkup-types/:id` | 定期健診項目の削除。 |
| PATCH | `/api/v1/masters/checkup-types/reorder` | 定期健診項目の表示順一括保存。 |

---
