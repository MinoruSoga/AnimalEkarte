# 診断・傷病名マスタ 仕様書 (Diagnosis Management)

## 概要
- **画面の目的**: 臨床現場で使用される標準的な診断名、および疾患カテゴリの体系的管理。
- **URLパターン**: `/settings/diagnosis`
- **アクセス権限**: 診療マスタ管理権限が必要（`ResourceMasterMedical`）

---

## 画面構成

### 1. 体系的カテゴリ・リスト
- **2階層構造**: 「カテゴリ（例：皮膚科）」と「診断名（例：アトピー性皮膚炎）」の親子関係で整理。
- **検索**: 疾患名、キーワードによる高速検索。

### 2. 詳細編集サイドパネル (`SidePeekPanel`)
- **名称**: 正式な診断名。
- **カテゴリ選択**: 所属する疾患グループの割り当て。
- **ソート順**: カルテ入力時に頻繁に使用する疾患を上位に表示するための重み付け設定。

---

## 主要な機能

### 1. カルテ入力の高速化
ここで定義されたマスタは、カルテ詳細画面の「診断」タブにおいて、サジェスト機能（Combobox）の基礎データとして使用されます。

### 2. 統計分析への活用
疾患カテゴリごとに来院件数を集計することで、病院が得意とする診療分野や季節性の疾患トレンドの分析が可能になります。

---

## 技術仕様

### 使用コンポーネント
- **`DiagnosisMatrix`**: カテゴリと傷病名の階層表示部品。
- **`PropInput`**: 各項目のインライン編集。

### API連携
| メソッド | エンドポイント | 用途 |
|:---|:---|:---|
| GET | `/api/v1/masters/diagnosis-types` | 診断カテゴリ一覧の取得。 |
| POST | `/api/v1/masters/diagnosis-types` | 診断カテゴリの作成。 |
| PATCH | `/api/v1/masters/diagnosis-types/:id` | 診断カテゴリ情報の更新。 |
| DELETE | `/api/v1/masters/diagnosis-types/:id` | 診断カテゴリの削除。 |
| PATCH | `/api/v1/masters/diagnosis-types/reorder` | 診断カテゴリ表示順の一括保存。 |
| GET | `/api/v1/masters/diagnosis-names` | 診断名一覧の取得。 |
| POST | `/api/v1/masters/diagnosis-names` | 診断名の作成。 |
| PATCH | `/api/v1/masters/diagnosis-names/:id` | 診断名情報の更新。 |
| DELETE | `/api/v1/masters/diagnosis-names/:id` | 診断名の削除。 |
| PATCH | `/api/v1/masters/diagnosis-names/reorder` | 診断名表示順の一括保存。 |

---
