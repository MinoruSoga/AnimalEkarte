# ワクチン登録/編集 仕様書

## 概要
- **画面の目的**: ワクチン接種記録の新規登録・編集
- **URLパターン**:
  - 新規: `/vaccinations/new?petId=xxx`
  - 編集: `/vaccinations/:id`
- **アクセス権限**: 認証済ユーザー全員

## 表示項目

| フィールド名 | 型 | 必須 | 説明 | DB列 |
|------------|-----|------|------|------|
| ペット | lookup | ✅ | petIdからペット・飼主情報表示 | `pets.id` |
| ワクチン名 | string | ✅ | ワクチンの名称（マスタ選択 or 手動） | `vaccinations.vaccine_name` |
| ワクチンマスタ | lookup | | マスタからの選択 | `vaccinations.vaccine_master_id` |
| 接種日 | date | ✅ | 接種した日付 | `vaccinations.vaccination_date` |
| 次回接種予定日 | date | | 推奨次回接種日 | `vaccinations.next_date` |
| 接種医 | lookup | ✅ | 担当獣医師選択 | `staffs.id` |
| ロット番号 | string | | ワクチンのロット番号 | `vaccinations.lot_number` |
| 備考 | textarea | | メモ | `vaccinations.notes` |

## バリデーション
- **ワクチン名**: 必須
- **接種日**: 必須
- **接種医**: 必須

## API連携

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/vaccinations/:id` | ワクチン詳細取得 | 未実装 |
| POST | `/api/v1/vaccinations` | ワクチン記録作成 | 未実装 |
| PUT | `/api/v1/vaccinations/:id` | ワクチン記録更新 | 未実装 |
| DELETE | `/api/v1/vaccinations/:id` | ワクチン記録削除 | 未実装 |

## 実装状況
- フロントエンド(ui-sample): 実装済（モックデータ使用）
- バックエンドAPI: 未実装
