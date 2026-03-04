# トリミング登録/編集 仕様書

## 概要
- **画面の目的**: トリミング予約・施術記録の新規登録・編集
- **URLパターン**:
  - 新規: `/trimming/new?petId=xxx`
  - 編集: `/trimming/:id`
- **アクセス権限**: 認証済ユーザー全員

## 表示項目

| フィールド名 | 型 | 必須 | 説明 | DB列 |
|------------|-----|------|------|------|
| ペット | lookup | ✅ | petIdからペット・飼主情報表示 | `pets.id` |
| 予約日時 | datetime | ✅ | トリミング予約日時 | `trimmings.appointment_date` |
| コース | lookup | ✅ | トリミングコース選択（マスタ） | `trimmings.course` |
| オプション | multi-select | | 追加オプション選択（マスタ） | `trimmings.options` |
| スタイル要望 | textarea | | カットスタイルの要望 | `trimmings.style_request` |
| 担当者 | lookup | ✅ | 担当トリマー選択 | `staffs.id` |
| ステータス | enum | ✅ | 予約/進行中/完了 | `trimmings.status` |
| 合計金額 | calc | | コース+オプション料金の自動計算 | `trimmings.total_price` |
| 備考 | textarea | | メモ | `trimmings.notes` |

## バリデーション
- **予約日時**: 必須
- **コース**: 必須（マスタから選択）
- **担当者**: 必須
- **ステータス**: 必須、デフォルト「予約」

## API連携

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/trimmings/:id` | トリミング詳細取得 | 未実装 |
| POST | `/api/v1/trimmings` | トリミング作成 | 未実装 |
| PUT | `/api/v1/trimmings/:id` | トリミング更新 | 未実装 |
| DELETE | `/api/v1/trimmings/:id` | トリミング削除 | 未実装 |

## 実装状況
- フロントエンド(ui-sample): 実装済（モックデータ使用）
- バックエンドAPI: 未実装
