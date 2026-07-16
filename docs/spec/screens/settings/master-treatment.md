# 診療項目マスタ 仕様書 (Treatment Items)

## 概要
- **画面の目的**: 診察、処置、手術、検査等の臨床行為の定義、および標準価格・税率の設定。
- **URLパターン**: `/settings/treatment-items` (タブパラメータ: `?tab=consultation` (診察), `?tab=examination` (検査), `?tab=procedure` (処置), `?tab=vaccine` (予防接種), `?tab=checkup` (定期健診))
- **アクセス権限**: 診療マスタ管理権限が必要（`ResourceMasterMedical`）

---

## 1. 画面構成

### 1.1 項目一覧テーブル
- **カテゴリ分類**: 診察、検査、処置、予防接種、定期健診のタブ切り替えにより、膨大な項目を整理して表示。
- **検索**: 項目名（親・子とも）による部分一致。

### 2. 詳細編集サイドパネル (`SidePeekPanel`)
- **基本属性**: 項目名（タイトル編集）、有効/無効ステータス、親カテゴリ（子項目がある場合は変更不可）、備考。
- **価格設定**（`TreatmentItemSidePanel` は全タブ共通のため画面上は以下を一律表示するが、実際に保存されるフィールドはタブ（テーブル）ごとに異なる）:
    - **単価**: `MoneyInput` による金額入力。全タブで保存される。
    - **課税区分**（外税 / 内税 / 非課税）・**税率**（10% / 8%）: 診察・処置のみ保存される（`consultations`/`procedures` テーブルの `tax_type`/`tax_rate`）。検査・予防・定期健診では画面上操作できるが保存されない（バックエンドに対応する列が存在しない）。
    - **保険対象外**（対象 / 対象外）: 検査のみ保存される（`exam_types.is_non_insurance`）。診察・処置・予防・定期健診では保存されない。

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
- **`UnifiedTabs`**: 大分類ごとの高速なデータ切り替え。
- **`TreatmentItemSidePanel`**: `MasterSidePanel` による項目名・単価・課税区分・税率・保険対象外の編集。
- **`PropertyInput`**: 備考欄のボーダーレス編集。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/consultations` | 診察項目一覧の取得 | `master-medical` | `view` |
| GET | `/api/v1/masters/consultations/:id` | 特定の診察項目情報の取得 | `master-medical` | `view` |
| POST | `/api/v1/masters/consultations` | 診察項目の作成 | `master-medical` | `create` |
| PATCH | `/api/v1/masters/consultations/:id` | 診察項目の更新 | `master-medical` | `edit` |
| DELETE | `/api/v1/masters/consultations/:id` | 診察項目の削除 | `master-medical` | `delete` |
| PATCH | `/api/v1/masters/consultations/reorder` | 診察項目の表示順一括保存 | `master-medical` | `edit` |
| GET | `/api/v1/masters/examination-types` | 検査項目一覧の取得 | `master-medical` | `view` |
| GET | `/api/v1/masters/examination-types/:id` | 特定の検査項目情報の取得 | `master-medical` | `view` |
| POST | `/api/v1/masters/examination-types` | 検査項目の作成 | `master-medical` | `create` |
| PATCH | `/api/v1/masters/examination-types/:id` | 検査項目の更新 | `master-medical` | `edit` |
| DELETE | `/api/v1/masters/examination-types/:id` | 検査項目の削除 | `master-medical` | `delete` |
| PATCH | `/api/v1/masters/examination-types/reorder` | 検査項目の表示順一括保存 | `master-medical` | `edit` |
| GET | `/api/v1/masters/procedures` | 処置項目一覧の取得 | `master-medical` | `view` |
| GET | `/api/v1/masters/procedures/:id` | 特定の処置項目情報の取得 | `master-medical` | `view` |
| POST | `/api/v1/masters/procedures` | 処置項目の作成 | `master-medical` | `create` |
| PATCH | `/api/v1/masters/procedures/:id` | 処置項目の更新 | `master-medical` | `edit` |
| DELETE | `/api/v1/masters/procedures/:id` | 処置項目の削除 | `master-medical` | `delete` |
| PATCH | `/api/v1/masters/procedures/reorder` | 処置項目の表示順一括保存 | `master-medical` | `edit` |
| GET | `/api/v1/masters/vaccines` | 予防接種項目一覧の取得 | `master-medical` | `view` |
| GET | `/api/v1/masters/vaccines/:id` | 特定の予防接種項目情報の取得 | `master-medical` | `view` |
| POST | `/api/v1/masters/vaccines` | 予防接種項目の作成 | `master-medical` | `create` |
| PATCH | `/api/v1/masters/vaccines/:id` | 予防接種項目の更新 | `master-medical` | `edit` |
| DELETE | `/api/v1/masters/vaccines/:id` | 予防接種項目の削除 | `master-medical` | `delete` |
| PATCH | `/api/v1/masters/vaccines/reorder` | 予防接種項目の表示順一括保存 | `master-medical` | `edit` |
| GET | `/api/v1/masters/checkup-types` | 定期健診項目一覧の取得 | `checkups` | `view` |
| GET | `/api/v1/masters/checkup-types/:id` | 特定の定期健診項目情報の取得 | `checkups` | `view` |
| POST | `/api/v1/masters/checkup-types` | 定期健診項目の作成 | `checkups` | `create` |
| PATCH | `/api/v1/masters/checkup-types/:id` | 定期健診項目の更新 | `checkups` | `edit` |
| DELETE | `/api/v1/masters/checkup-types/:id` | 定期健診項目の削除 | `checkups` | `delete` |
| PATCH | `/api/v1/masters/checkup-types/reorder` | 定期健診項目の表示順一括保存 | `checkups` | `edit` |


---

