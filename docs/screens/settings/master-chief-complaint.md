# 主訴種別マスタ 仕様書 (Chief Complaint Types)

## 概要
- **画面の目的**: カルテ作成時に飼い主が訴える「主訴」を分類し、統計や入力補助に使用するための大分類（消化器、皮膚、眼科等）の管理。
- **URLパターン**: `/settings/interview/chief-complaint`
- **アクセス権限**: 診療マスタ管理権限が必要（`ResourceMasterMedical`）

---

## 1. 画面構成

### 1.1 主訴カテゴリ一覧
- **項目**: 名称、説明、有効ステータス。

### 1.2 詳細編集サイドパネル (`SidePeekPanel`)
- **カテゴリ名**: 例：「嘔吐・下痢」「痒み・赤み」「ワクチン・予防」。
- **説明**: カテゴリの補足説明を自由記述（テキストエリア）。
- **ステータス**: 有効/無効の切り替え。

---

## 2. 主要な機能

### 2.1 臨床入力の支援
カルテ入力画面（Tab1 問診）において、主訴のフリー入力と並行して「カテゴリ」を選択させることで、後の経営分析や症例研究での正確な母集団抽出を容易にします。

### 2.2 診療トレンドの可視化
顧客集計や経営レポートにおいて、「今月は皮膚科疾患が○%増加した」等のトレンドを把握するための基礎データとなります。

---

## 3. 技術仕様

### 使用コンポーネント
- **`ChiefComplaintSettings`**: メインページ。
- **`ChiefComplaintSidePanel`**: `MasterSidePanel` + `PropertyRow` によるカテゴリ名・説明の編集。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/chief-complaint-types` | 有効なカテゴリ一覧の取得 | `master-medical` | `view` |
| GET | `/api/v1/masters/chief-complaint-types/:id` | 特定のカテゴリ情報の取得 | `master-medical` | `view` |
| POST | `/api/v1/masters/chief-complaint-types` | 新規カテゴリの作成 | `master-medical` | `create` |
| PATCH | `/api/v1/masters/chief-complaint-types/:id` | 名称・説明・ステータスの更新 | `master-medical` | `edit` |
| DELETE | `/api/v1/masters/chief-complaint-types/:id` | カテゴリの削除 | `master-medical` | `delete` |
| PATCH | `/api/v1/masters/chief-complaint-types/reorder` | 表示順の一括保存（BE実装済みだが本画面からは未呼出） | `master-medical` | `edit` |

---

