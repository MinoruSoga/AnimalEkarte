# 医院マスタ設定 仕様書

## 概要
- **画面の目的**: システムを利用する各医院（本院・分院等）の基本情報（名称・住所・連絡先等）を一元管理する。
- **URLパターン**: `/settings/clinic`
- **コンポーネント**: `[R] ClinicMasterSettings`
- **アクセス権限**: `hospital-settings` リソース — 管理者グループのみ作成・編集・削除可（RBAC実装済み）

## 画面構成
- **メインエリア (左)**: `DataTable` による医院一覧。
  - カラム: 院名、電話番号、メール、ステータス（ドット付）。
- **サイドパネル (右)**: `SidePeekPanel` による詳細編集。編集パネルが開いている間は、左側のリストも参照可能。

## 表示・フォーム項目

### 一覧テーブル
| フィールド | 説明 | 備考 |
|-----------|------|------|
| 院名 | 医院の名称 | `item.name` |
| 電話番号 | 代表電話番号 | `item.phoneNumber` (等幅) |
| メール | 代表メールアドレス | `item.email` |
| ステータス | 有効/無効 | `is_active` に応じて配色 |

### 編集項目（サイドパネル）
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| 院名 | `name` | `Input` | ✅ | タイトルエリア（大文字） |
| ステータス | `is_active` | `NotionStatusPill` | - | クリックでトグル |
| 郵便番号 | `postal_code` | `PropInput` | - | |
| 住所 | `address` | `PropInput` | - | |
| 電話番号 | `phone_number` | `PropInput` | - | |
| FAX番号 | `fax_number` | `PropInput` | - | |
| 登録番号 | `registration_number` | `PropInput` | - | インボイス登録番号 |
| 院長名 | `director_name` | `PropInput` | - | |
| 通常課税 | `standard_tax_rate`| `PropInput` (number) | - | 内部は 0.1 等の小数、UIは % 表示 |
| 軽減税率 | `reduced_tax_rate` | `PropInput` (number) | - | 内部は 0.08 等の小数、UIは % 表示 |

## 特徴的なUI・機能
- **React 19 アクション**: `useActionState` による保存処理。完了時に自動的にサイドパネルを閉じ、トースト通知を表示。
- **Notionスタイル**: `PropertyRow` と `PropInput` を使用したボーダーレスな入力体験。
- **離脱防止**: パネルが開いている間は `NavigationBlocker` によりSPA内遷移をガード。
- **マルチテナント**: 各医院のデータは JWT の `clinic_id` とは独立して、全医院（所属しているもの）を管理可能（システム管理者向け）。

## API連携
| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/clinics` | 医院一覧取得 | 実装済 |
| POST | `/api/v1/clinics` | 医院作成 | 実装済 |
| PATCH | `/api/v1/clinics/:id` | 医院更新 | 実装済 |
| DELETE | `/api/v1/clinics/:id` | 医院削除 | 実装済 |

## 実装状況
- フロントエンド: 実装済（`features/hospital-settings/routes/ClinicMasterSettings.tsx`）
- バックエンドAPI: 実装済（`handler/clinic_handler.go`）
