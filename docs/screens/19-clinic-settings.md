# 医院マスタ設定 仕様書

## 概要
- **画面の目的**: システムを利用する各医院（本院・分院等）の基本情報（名称・住所・連絡先等）を一元管理する。
- **URLパターン**: `/settings/clinic`
- **コンポーネント**: `[R] ClinicMasterSettings`
- **アクセス権限**: `hospital-settings` リソース — 管理者グループのみ作成・編集・削除可（RBAC実装済み）

## 画面構成
- **メインエリア**: `DataTable` による医院一覧。
- **サイドパネル**: スライドイン形式の編集パネル。

## 表示・フォーム項目

### 一覧テーブル
| フィールド | 説明 | 備考 |
|-----------|------|------|
| 院名 | 医院の名称 | |
| 電話番号 | 医院の代表電話番号 | |
| メール | 医院の代表メールアドレス | |
| ステータス | 有効/無効のステータスバッジ | |
| 操作 | 詳細・編集ボタン | |

### 編集項目（サイドパネル）
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| 院名 | `name` | `Input` | ✅ | タイトルエリア |
| ステータス | `is_active` | `StatusToggleButton` | - | |
| 郵便番号 | `postal_code` | `PropInput` | - | |
| 住所 | `address` | `PropInput` | - | |
| 電話番号 | `phone_number` | `PropInput` | - | |
| FAX番号 | `fax_number` | `PropInput` | - | |
| 登録番号 | `registration_number` | `PropInput` | - | インボイス登録番号等 |
| 院長名 | `director_name` | `PropInput` | - | |
| メールアドレス | `email` | `PropInput` | - | |
| Webサイト | `website` | `PropInput` | - | |
| 通常税率 | `standard_tax_rate`| `Input(number)` | - | デフォルト 10% |
| 軽減税率 | `reduced_tax_rate` | `Input(number)` | - | デフォルト 8% |

## 特徴的なUI・機能
- **NotionスタイルUI**: `PropertyRow` と `NotionStatusPill` を使用したクリーンな属性表示。
- **インライン編集**: サイドパネル内での直接編集と、`useActionState` による非同期保存。
- **ナビゲーション保護**: 編集パネルが開いている間は、`NavigationBlocker` により不意の離脱を防止。

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
