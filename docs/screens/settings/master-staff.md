# スタッフマスタ設定 仕様書

## 概要
- **画面の目的**: システムを利用するスタッフの氏名、役割、資格番号、ログインアカウント情報を管理する。
- **URLパターン**: `/settings/staff`
- **コンポーネント**: `StaffSettings`
- **アクセス権限**: `ResourceMasterStaff` 権限が必要

## 画面構成
- **メインエリア**: `DataTable` によるスタッフ一覧。
  - カラム: 氏名、職種（`occupations`）、権限グループ（カラーバッジ形式で最大2件+個数表示）、ステータス、操作。
- **サイドパネル**: `StaffSidePanel` による詳細編集。
  - 5つのセクション（基本情報、LINE予約設定、対応不可コース、所属医院設定、権限グループ設定）で構成。

## 表示項目（サイドパネル）

### 基本情報セクション
| フィールド名 | 項目ID | 入力部品 | 必須 | 備考 |
|------------|--------|---------|------|------|
| 氏名 | `name` | `Input` | ✅ | 最大100文字 |
| ステータス | `isActive` | `StatusToggleButton` | - | |
| 職種 | `occupationId` | `Select` | - | 職種マスタ（`occupations`）連動 |
| 資格番号 | `licenseNumber` | `Input` | - | |
| メールアドレス | `email` | `Input` | ✅ | 新規登録時必須。編集時は表示のみ |
| パスワード | `password` | `Input` | ✅ | 新規登録時必須。編集時は「変更時のみ入力」 |

### LINE予約設定セクション
| フィールド名 | 項目ID | 入力部品 | 必須 | 備考 |
|------------|--------|---------|------|------|
| LINE表示名 | `reservationDisplayName` | `Input` | - | 空欄時は氏名を使用 |
| 予約ページに表示| `reservationVisible` | `Switch` | - | デフォルト: ON |
| スタッフ種別 | `staffType` | `Select` | - | 医師 / 看護師 / 設備 |
| LINE説明文 | `reservationComment` | `Input` | - | |
| 画像URL | `reservationImageUrl` | `Input` | - | |

### 対応不可コース（セクション）
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| 除外コース | `excludedIds` | `Checkbox` リスト | - | アクティブな予約区分（カラードット付）を表示。新規登録時は非表示 |

### 所属医院設定（セクション）
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| 所属医院 | `clinicIds` | `Checkbox` リスト | - | 全医院リストを表示。新規登録時は非表示 |

### 権限グループ設定（セクション）
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| 権限グループ | `groupIds` | `Checkbox` リスト | - | 全権限グループ（カラードット付）を表示。新規登録時は非表示 |

## 主要機能
- **権限バッジの動的生成**: 権限グループの色（`color`）に基づき、背景透過 10%・文字色 100% のスタイルを動的に生成してバッジ表示。
- **マルチセクション保存**: 基本情報の保存ボタン押下時に、権限・所属医院・除外コースの各更新 API を並行して呼び出し同期をとる（`useMasterSave` フック）。
- **職種フィルタ**: マスタから取得した職種一覧に基づき、`NotionFilter` の選択肢を動的に生成。

## API連携
| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/masters/staffs` | スタッフ一覧取得 | 実装済 |
| POST | `/api/v1/masters/staffs` | スタッフ作成（＋アカウント作成） | 実装済 |
| PATCH | `/api/v1/masters/staffs/:id` | スタッフ情報更新（LINE項目含む） | 実装済 |
| DELETE | `/api/v1/masters/staffs/:id` | スタッフ削除 | 実装済 |
| GET | `/api/v1/masters/staffs/:id/excluded-reservation-types` | 対応不可コース取得 | 実装済 |
| PUT | `/api/v1/masters/staffs/:id/excluded-reservation-types` | 対応不可コース設定 | 実装済 |
| GET | `/api/v1/masters/staffs/:id/clinics` | 所属医院リスト取得 | 実装済 |
| PUT | `/api/v1/masters/staffs/:id/clinics` | 所属医院リスト設定 | 実装済 |
| GET | `/api/v1/masters/staffs/:id/permission-groups` | 権限グループ取得 | 実装済 |
| PUT | `/api/v1/masters/staffs/:id/permission-groups` | 権限グループ設定 | 実装済 |
