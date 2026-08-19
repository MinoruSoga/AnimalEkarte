# 薬剤マスタ 仕様書 (Medicine Management)

## 概要
- **画面の目的**: 院内で使用・処方される医薬品の定義、単価設定、および在庫システムとの紐付け。
- **URLパターン**: `/settings/medicine`
- **アクセス権限**: 診療マスタ管理権限が必要（`ResourceMasterMedical`）

---

## 1. 画面構成

### 1.1 薬剤検索リスト
- **検索**: 薬品名による絞り込み。
- **ステータス**: 有効/無効のフィルタ切り替えに対応（フィルタはステータスのみ）。
- **親カテゴリ**: フィルタではなくグループ表示で扱い、カテゴリ単位の折りたたみ切り替えに対応。
- **カテゴリ判定**: 親なし・単価 0・剤形なし・単位なしの行だけをカテゴリ見出しとする（`isMedicineCategoryNode`）。親なしでも剤形または単位があれば未分類の薬剤として治療検索候補に出す。

### 1.2 詳細編集サイドパネル (`SidePeekPanel`)
- **基本情報**: 薬品名、親カテゴリ、剤形（錠剤、液剤、注射剤、外用剤、散剤）、単位。
- **処方設定**:
    - **単位**: 「1錠あたり」「1mlあたり」「1回あたり」「1gあたり」等の販売・投与単位の定義。
    - **単価(税込)**、**課税区分**（外税/内税/非課税）、**税率**（10%/8%）。
    - **保険対象外**: 対象/対象外の切り替え。
- **投与量自動計算（#201・実装済み）**:
    - `calculation_type = per_weight` の薬剤では、サイドパネルに `MedicineDoseParamsEditor` を表示し、動物種別（犬・猫）ごとの dose パラメータを編集できる（`MedicineSidePanelBody` 経由）。
    - 下限 > 上限はインラインエラーで拒否する。下限 = 上限は受理する。保存成功時はトーストを出す。HTML5 制約だけで無音ブロックしない。
- **在庫連携**:
    - バックエンドの `Medicine` モデルは在庫アイテム（`inventory_id`）を保持するが、本設定画面のサイドパネルに在庫アイテム紐付けUIは未実装（#201 とは別課題）。

---

## 2. 主要な機能

### 2.1 処方入力の自動化
ここで設定された剤形と単位は、カルテ詳細の「処方」タブにおける入力補助（数量選択や単位表示）として反映されます。体重・species からの数量自動プリフィルはカルテ側 `TreatmentRow`（`calculateDose` / `computeDoseGate`）が担う。上限超過時はインライン表示して送信せず、`ConfirmDialog` による解除経路は無い（#201。挙動詳細は `06-medical-records-form.md` §2.3）。

### 2.2 原価・在庫管理の統合（バックエンドのみ）
`PATCH /api/v1/masters/medicines/:id` は `inventory_id` を受け付け、薬品と在庫アイテムを紐付け可能だが、本設定画面のサイドパネルには対応する入力UIが存在しない（他画面または今後の実装課題）。

---

## 3. 技術仕様

### 3.1 構成コンポーネント
- **`MedicineSettings`**: メインページ。
- **`MedicineSidePanel`**: 薬価・剤形・単位・課税区分の直接編集。
- **`MedicineDoseParamsEditor`**: 動物種別 dose パラメータ（per_weight）の編集 UI（#201）。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/medicines` | 薬剤一覧の取得 | `master-medical` | `view` |
| GET | `/api/v1/masters/medicines/:id` | 特定の薬剤詳細の取得 | `master-medical` | `view` |
| POST | `/api/v1/masters/medicines` | 新規薬剤の登録 | `master-medical` | `create` |
| PATCH | `/api/v1/masters/medicines/:id` | 価格や在庫紐付け、剤形の更新 | `master-medical` | `edit` |
| DELETE | `/api/v1/masters/medicines/:id` | 薬剤の削除 | `master-medical` | `delete` |
| PATCH | `/api/v1/masters/medicines/reorder` | 表示順序の一括保存 | `master-medical` | `edit` |
| GET | `/api/v1/masters/medicines/:id/dose-params` | 薬剤の動物種別投与量計算パラメータ一覧取得 | `master-medical` | `view` |
| PUT | `/api/v1/masters/medicines/:id/dose-params/:species` | 動物種別投与量計算パラメータの登録・更新 | `master-medical` | `edit` |
| DELETE | `/api/v1/masters/medicines/:id/dose-params/:species` | 動物種別投与量計算パラメータの削除 | `master-medical` | `delete` |


---
