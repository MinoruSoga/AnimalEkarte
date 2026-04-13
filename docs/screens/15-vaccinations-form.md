## 概要
予防接種（ワクチン、フィラリア予防等）の新規登録および編集を行う独立した画面です。

- **URLパターン**:
  - 新規: `/vaccinations/new?petId=xxx`
  - 編集: `/vaccinations/:id`
- **コンポーネント**: `VaccinationForm`
- **アクセス権限**: 認証済ユーザー全員（`ResourceVaccinations` 権限が必要）

## 画面構成（2カラム）
- **ヘッダー**: `PatientInfoCard` (ペット基本情報)、保存・キャンセル・削除ボタン
- **左カラム (入力フォーム)** [3/5幅]:
  - 予防接種日 (`NotionDatePicker`) ＊必須
  - ワクチン選択 (`Select`) ＊必須。マスタ (`vaccine`) と連動。
  - 補助説明 (`Input`)
  - LOT番号 1〜4（2列グリッド。全4項目保存可能）
  - 次回予定設定（`Select`: 3週後 / 4週後 / 1年後 / 以外（手動））
  - 次回予定日 (`NotionDatePicker`、選択肢に応じて自動計算）
  - 備考テキストエリア
- **右カラム (接種履歴)** [2/5幅]:
  - 同一ペットの過去の予防接種履歴一覧（`VaccinationCard`）
  - `HistoryFilterPanel`（キーワード検索、期間フィルタ、昇降順ソート）

## 主要機能
- **React 19 アクション**: `useActionState` を使用した記録の保存。バリデーションエラー時には、接種日（`vaccination-date`）やワクチン選択（`vaccine-select`）などの入力欄へ自動でスクロール・フォーカスするアクセシビリティ対応。
- **次回予定日の自動計算**: 「3週後/4週後/1年後」を選択すると、接種日に基づいて `nextDate` が自動計算される。
- **履歴管理**: 過去の接種記録を右カラムで確認しながら入力が可能。
- **未保存変更警告**: `NavigationBlocker`（`unstable_useBlocker`）により、意図しない離脱を防止。
- **臨床安全ガード**: `PatientInfoCard` において、死亡済みペットの場合はステータスが強調表示される。

## API連携
| メソッド | エンドポイント | 用途 |
|---------|--------------|------|
| GET | `/api/v1/vaccinations/:id` | 予防接種詳細取得 |
| POST | `/api/v1/vaccinations` | 予防接種作成 |
| PATCH | `/api/v1/vaccinations/:id` | 予防接種更新 |
