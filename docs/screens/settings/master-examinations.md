# 検査項目定義マスタ 仕様書 (Examination Definitions)

## 概要
- **画面の目的**: 血液検査や生化学検査等における、各測定項目の名称、単位、および動物種ごとの正常値範囲（リファレンスレンジ）の定義。
- **URLパターン**: `/settings/treatment-items?tab=examination`
- **アクセス権限**: 診療マスタ管理権限が必要（`ResourceMasterMedical`）

---

## 画面構成

### 1. 検査グループ・項目一覧
- **グループ構造**: 「血液一般」「生化学」「尿検査」等のカテゴリで項目を整理。
- **項目リスト**: 項目名、略称（GOT, CRE等）、単位、現在の有効ステータス。

### 2. 詳細編集サイドパネル (`SidePeekPanel`)
- **基本属性**: 項目名、略称、測定単位（mg/dL, %, /μL 等）。
- **基準値設定 (Reference Range)**:
    - **下限値 (Min)**: これを下回ると「LOW」判定。
    - **上限値 (Max)**: これを上回ると「HIGH」判定。
    - **動物種別設定**: 犬、猫、あるいは共通の基準値を個別に設定可能。臨床現場のニーズに合わせ、動物種ごとに異なる「正常」を定義します。

---

## 主要な機能

### 1. 臨床判断の自動支援
カルテの検査結果入力画面（`/examinations`）において、ここで定義された基準値に基づき、異常値（HIGH/LOW）が即座にカラーハイライトされます。

### 2. 表示順の柔軟なカスタマイズ
外部の自動検査機からの出力順や、院内での転記フローに合わせ、ドラッグ&ドロップ（`reorder`）機能により項目の表示順を最適化できます。

---

## 技術仕様

### 使用コンポーネント
- **`ExaminationMatrix`**: グループと項目の階層構造を管理する専用 UI。
- **`PropInput`**: 数値（基準値）や単位のインライン編集。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/examination-types` | 定義済み項目の一覧取得 | `master-medical` | `view` |
| GET | `/api/v1/masters/examination-types/:id` | 特定の検査項目情報の取得 | `master-medical` | `view` |
| POST | `/api/v1/masters/examination-types` | 新規検査項目の登録 | `master-medical` | `create` |
| PATCH | `/api/v1/masters/examination-types/:id` | 基準値や属性の更新 | `master-medical` | `edit` |
| DELETE | `/api/v1/masters/examination-types/:id` | 検査項目の削除 | `master-medical` | `delete` |
| PATCH | `/api/v1/masters/examination-types/reorder` | 並び順の一括保存 | `master-medical` | `edit` |

---


