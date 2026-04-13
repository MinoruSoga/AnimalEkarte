# 問診関連マスタ設定 仕様書

## 概要
- **画面の目的**: カルテの問診タブで使用される主訴カテゴリおよび、問診票の入力テンプレートを管理する。
- **URLパターン**: 
  - 主訴カテゴリ: `/settings/interview/chief-complaint`
  - 問診テンプレート: `/settings/interview/templates`

## 1. 主訴カテゴリマスタ (`ChiefComplaintSettings`)

### 概要
問診入力時に選択する主訴の大まかな分類を定義する。

### 表示・フォーム項目（サイドパネル）
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| 名称 | `name` | `Input` | ✅ | タイトルエリア |
| ステータス | `isActive` | `NotionStatusPill` | - | |
| 説明 | `description`| `PropertyInput` | - | |

---

## 2. 問診テンプレートマスタ (`InterviewTemplateSettings`)

### 概要
主訴や経過の入力時に、定型文として呼び出せるテンプレートを定義する。

### 表示・フォーム項目（サイドパネル）
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| タイトル | `title` | `Input` | ✅ | タイトルエリア |
| カテゴリ | `category` | `PropertyInput` | ✅ | 検索用の分類ラベル |
| ステータス | `isActive` | `NotionStatusPill` | - | |
| 内容 | `content` | `Textarea` | ✅ | 挿入されるテキスト本文 |

## API連携
| メソッド | エンドポイント | 用途 |
|---------|--------------|------|
| GET | `/api/v1/masters/chief-complaint-categories` | 主訴カテゴリ一覧 |
| GET | `/api/v1/masters/inquiry-templates` | テンプレート一覧 |
| POST | `/api/v1/masters/chief-complaint-categories` | 主訴カテゴリ作成 |
| POST | `/api/v1/masters/inquiry-templates` | テンプレート作成 |
