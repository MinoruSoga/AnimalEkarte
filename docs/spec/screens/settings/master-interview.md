# 問診・定型文マスタ 仕様書 (Interview Templates)

## 概要
- **画面の目的**: カルテの主訴、所見、指導内容の入力を効率化するための定型文テンプレートの管理。
- **URLパターン**: `/settings/inquiry-templates`（サイドバー「問診設定」の導線。`/settings/interview/templates` も同一画面を表示する別ルートとして存在）
- **アクセス権限**: 診療マスタ管理権限が必要（`ResourceMasterMedical`）

---

## 画面構成

### 1. テンプレート一覧
- **カテゴリ分類**: `chief_complaint`(主訴)、`history`(既往歴)、`current_medications`(現在の投薬)、`notes`(メモ/備考) を日本語ラベルに変換して表示（`INQUIRY_CATEGORY_LABELS`、未知の値は生の文字列を表示）。カテゴリは自由入力のためこれ以外の値も登録可能。
- **検索**: タイトルおよびカテゴリ文字列による部分一致（本文内容は検索対象外）。
- **並び順**: `sort_order` 昇順→タイトル昇順（FEは作成時に `sort_order` を送信せず全件デフォルト0のため、実質タイトル順。ドラッグ並び替え・使用頻度によるソートは未実装）。

### 2. 詳細編集サイドパネル (`SidePeekPanel`)
- **タイトル**: 選択時に表示される短い名称。
- **カテゴリ**: 自由入力のテキストフィールド（`<input type="text">`）。事前定義された選択肢はない。
- **テンプレート内容**: プレーンテキストの `<textarea>`。Markdown記法の特別な解釈・レンダリングは行われない。

---

## 主要な機能

### 1. 現状（本設定画面のスコープ）
本画面は問診テンプレートのCRUD管理のみを提供する。`frontend/src/features/medical-records/` 配下のカルテ入力UIから `inquiry-templates` API・`useGetMasterItems("inquiryTemplate")` を呼び出す実装は現時点で存在せず、登録したテンプレートをカルテへ自動挿入する機能・臨床フェーズに応じた動的サジェスト機能は未実装（`frontend/src/hooks/use-master-items.ts` の `inquiryTemplate` キーはマップ定義のみで呼び出し元なし）。

---

## 技術仕様

### 使用コンポーネント
- **`InterviewTemplateSettings`**: メインページ。
- **`InterviewTemplateSidePanel`**: タイトル・カテゴリ（テキスト入力）・内容（`<textarea>`）の編集パネル。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/inquiry-templates` | 登録済みテンプレートの取得 | `master-medical` | `view` |
| GET | `/api/v1/masters/inquiry-templates/:id` | 特定のテンプレート情報の取得 | `master-medical` | `view` |
| POST | `/api/v1/masters/inquiry-templates` | 新規テンプレートの保存 | `master-medical` | `create` |
| PATCH | `/api/v1/masters/inquiry-templates/:id` | タイトル・カテゴリ・本文・ステータスの更新 | `master-medical` | `edit` |
| DELETE | `/api/v1/masters/inquiry-templates/:id` | テンプレートの削除 | `master-medical` | `delete` |
| PATCH | `/api/v1/masters/inquiry-templates/reorder` | 表示順の一括保存 | `master-medical` | `edit` |

**注記**: `reorder` エンドポイントはバックエンドに存在するが、`InterviewTemplateSettings` はドラッグ並び替えUIを持たず（1.「並び順」参照）、FE からは未呼出（対応する FE フックは FE-R3 のデッドコード削除で撤去済み）。

---

