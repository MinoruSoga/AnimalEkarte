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

## フォーム項目（VaccinationFormFields）

| フィールド | 項目ID | 入力部品 | 備考 |
|-----------|--------|----------|------|
| 予防接種名 | `vaccineId` | `Select` | `vaccine` マスタ連動 |
| 予防接種日 | `date` | `Input(date)` | |
| 補助説明 | `supplemental`| `Input` | |
| LOT1〜LOT4 | `lot1`〜`lot4` | `Input` × 4 | |
| 次回予定設定 | `nextScheduleType`| `RadioGroup` | 3週/4週/1年/以外 |
| 次回予定日 | `nextDate` | `Input(date)` | |
| 備考 | `remarks` | `Textarea` | |

## 履歴フィルタ項目（HistorySection）

| フィールド | 項目ID | 入力部品 | 備考 |
|-----------|--------|----------|------|
| 実施日(開始) | `filterStartDate`| `Input(date)` | |
| 実施日(終了) | `filterEndDate` | `Input(date)` | |
| 検索単語 | `historySearchTerm`| `Input` | |
| 並び順 | `sortOrder` | `Select` | 降順 / 昇順 |

## 機能詳細

### 1. 次回予定日の自動計算ロジック
- **間隔の引用**: 選択された予防接種名に基づき、ワクチンマスタ（`vaccines`）から標準的な接種間隔（1年、3週間など）を取得する。
- **カレンダー連動**: 「次回予定設定」のラジオボタン（3週/4週/1年）を選択すると、今回実施日にその期間を加算した日付が `nextDate` フィールドに自動セットされる。

### 2. LOT番号の複数管理
- **追跡可能性**: 同一ワクチンでも製造ロットが異なる場合に対応するため、最大4つまでのLOT番号（`lot1`〜`lot4`）を個別に記録し、後のトレーサビリティを確保する。

### 3. 履歴の時系列表示とフィルタ
- **動的フィルタ**: `HistorySection` において、特定のワクチン名や実施期間で絞り込みを行い、過去の反応や副反応のメモを即座に参照できる。

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/vaccinations/:id` | 予防接種詳細取得 | 実装済 |
| POST | `/api/v1/vaccinations` | 予防接種作成 | 実装済 |
| PATCH | `/api/v1/vaccinations/:id` | 予防接種更新 | 実装済 |

## 実装状況
- フロントエンド: 実装済（`features/vaccinations/routes/VaccinationForm.tsx`）
- バックエンドAPI: 実装済（`handler/vaccination_handler.go`）


## 備考

- 独立した `/vaccinations/new` や `/vaccinations/:id` ルートは存在しない
- 編集はカルテ内の予防接種タブを介して行う
- ワクチンマスタ（`vaccine` カテゴリ）の標準接種間隔から次回予定日を自動算出する設計
