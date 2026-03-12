# 予防接種フォーム 仕様書

## 概要

- **画面の目的**: 予防接種記録の新規登録・編集（カルテ内タブとして提供）
- **アクセス経路**:
  - 一覧の「新規登録」ボタン → `/medical-records/select-pet`（state: `{ activeTab: "予防接種" }`）
  - 一覧の行クリック / 「カルテを開く」 → `/medical-records/{id}`（state: `{ activeTab: "予防接種" }`）
- **アクセス権限**: 認証済ユーザー全員

## 重要仕様

**予防接種フォームはカルテ（`/medical-records/:id`）内の「予防接種」タブとして実装されている。**
独立したフォームページ（`/vaccinations/new` 等）は存在しない。

カルテの「予防接種」タブからインライン入力し、カルテ保存と同時に予防接種記録も保存する設計となっている。

## カルテ内 予防接種タブ 表示項目

| フィールド名 | 入力部品 | 必須 | 説明 | DB列 |
|------------|---------|------|------|------|
| ワクチン名 | `MasterSelectModal`（vaccine マスタ連動）/ 手動入力 | 必須 | ワクチンの名称 | `vaccinations.vaccine_id` |
| 接種日 | `NotionDatePicker` | 必須 | 接種した日付 | `vaccinations.date` |
| 次回接種予定日 | `NotionDatePicker` | - | 推奨次回接種日（マスタ標準間隔から自動計算） | `vaccinations.next_date` |
| 担当医 | カルテの担当医を引き継ぎ | - | 接種担当獣医師 | `staffs.id` |
| ロット番号 | `Input` | - | ワクチンロット番号 | `vaccinations.lot1` |
| 備考 | `Textarea` | - | メモ | `vaccinations.remarks` |

## バリデーション

- **ワクチン名**: 必須（マスタ選択または手動入力）
- **接種日**: 必須

## 遷移フロー

```
予防接種一覧「新規登録」
  → /medical-records/select-pet
      → ペット選択
          → /medical-records/new?petId=xxx  (カルテ新規作成)
              → 「予防接種」タブ で入力
                  → カルテ保存と同時に予防接種記録保存
```

## API連携

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/vaccinations/:id` | ワクチン詳細取得 | 未実装 |
| POST | `/api/v1/vaccinations` | ワクチン記録作成 | 未実装 |
| PUT | `/api/v1/vaccinations/:id` | ワクチン記録更新 | 未実装 |
| DELETE | `/api/v1/vaccinations/:id` | ワクチン記録削除 | 未実装 |
| GET | `/api/v1/master-items?category=vaccine` | ワクチンマスタ取得 | 実装済 |

## 実装状況

- フロントエンド(ui-sample): 実装済（カルテ内タブ、モックデータ使用）
- バックエンドAPI: 未実装

## 備考

- 独立した `/vaccinations/new` や `/vaccinations/:id` ルートは存在しない
- 編集はカルテ内の予防接種タブを介して行う
- ワクチンマスタ（`vaccine` カテゴリ）の標準接種間隔から次回予定日を自動算出する設計
