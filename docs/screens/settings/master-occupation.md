# 職種マスタ 仕様書 (Occupations Management)

## 概要
- **画面の目的**: クリニック内で勤務するスタッフの職種（獣医師、看護師、トリマー、受付等）の定義。
- **URLパターン**: `/settings/occupations`
- **アクセス権限**: スタッフ管理者権限が必要（`ResourceMasterStaff`）

---

## 1. 画面構成

### 1.1 職種一覧
- **項目**: 職種名、説明、有効ステータス。
- **並び替え**: シフト管理やスタッフ選択時の優先順位を管理するため、ドラッグ操作による表示順の変更に対応。

### 1.2 詳細編集サイドパネル (`SidePeekPanel`)
- **職種名**: 例：「院長」「チーフトリマー」「AHT（動物看護師）」。
- **説明**: 職種の補足説明を自由記述で入力する任意項目。
- **有効/無効**: `StatusToggleButton` によるステータス切り替え。

---

## 2. 主要な機能

### 2.1 スタッフ登録との連動
スタッフ管理画面 (`/settings/staff`) において、ここで定義された職種を割り当てます。

### 2.2 権限プリセットとの親和性
「獣医師」などの職種に基づき、推奨される権限グループを紐付ける際の論理的な基準として機能します。

---

## 3. 技術仕様

### 使用コンポーネント
- **`OccupationSettings`**: メインコンテナ。
- **`OccupationSidePanel`**: `MasterSidePanel` による職種名・説明の編集。
- **`PropertyInput`**: 説明欄のインライン編集。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/occupations` | 定義済み職種の一覧取得 | `master-staff` | `view` |
| GET | `/api/v1/masters/occupations/:id` | 特定の職種詳細の取得 | `master-staff` | `view` |
| POST | `/api/v1/masters/occupations` | 新規職種の追加 | `master-staff` | `create` |
| PATCH | `/api/v1/masters/occupations/:id` | 属性やカテゴリの更新 | `master-staff` | `edit` |
| DELETE | `/api/v1/masters/occupations/:id` | 職種の削除 | `master-staff` | `delete` |
| PATCH | `/api/v1/masters/occupations/reorder` | 表示順序の一括保存 | `master-staff` | `edit` |

---

