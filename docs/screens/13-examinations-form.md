# 検査入力/結果登録 仕様書

## 概要
検査の新規登録および結果入力は `/examinations/new` および `/examinations/:id` として実装されています。

- **コンポーネント**: `ExaminationForm`
- **アクセス権限**: 認証済ユーザー全員

## 画面構成（2カラム）
- **ヘッダー**: `PatientInfoCard` (ペット基本情報)、保存・キャンセル・削除ボタン
- **左カラム (FormFieldsSection)** [3/5幅]:
  - 検査種別選択 (`Select`) ＊必須
  - 担当医選択 (`Select`) ＊必須
  - 検査日 (`NotionDatePicker`) ＊必須
  - ステータス (`Select`): 依頼中 / 検査中 / 結果入力済み / 完了 / 確定
  - 備考・所見テキストエリア
- **右カラム (ExaminationHistory)** [2/5幅]:
  - 過去の検査履歴一覧（同一ペットの過去記録、`ExaminationCard` で表示）
  - `HistoryFilterPanel`（キーワード検索、期間フィルタ、昇降順ソート）

## 主要機能
- **React 19 アクション**: `useActionState` を使用した検査記録の保存。バリデーションエラー時には、検査種別（`testTypeId`）や担当医（`doctorId`）などの入力欄へ自動でスクロール・フォーカスするアクセシビリティ対応。
- **確定済みロック**: `status === "確定"` の場合、フォーム全体が `disabled` 状態となり、保存・削除アクションが制限される。
- **履歴管理**: 右カラムに同一ペットの過去の検査履歴を一覧表示。`HistoryFilterPanel` による絞り込みが可能。
- **未保存変更警告**: `NavigationBlocker`（`unstable_useBlocker`）により、保存前に離脱しようとするとアラートを表示。
- **臨床安全ガード**: `PatientInfoCard` において、死亡済みペットの場合はステータスが強調表示される。

## 実装状況・制約
| 機能 | 状態 | 備考 |
|------|------|------|
| 検査種別 / 担当医 / ステータス / 備考 | ✅ 実装済 | |
| 確定済みロック | ✅ 実装済 | status === "確定" で全フィールド disabled |
| 過去履歴（右カラム）| ✅ 実装済 | クライアントサイドフィルタ（petId一致） |
| 検査項目テーブル | ❌ 未実装 | バックエンド対応が必要（BE-EXAM-001） |
| 異常値判定 | ❌ 未実装 | バックエンド対応が必要（BE-EXAM-001） |
| 画像管理 | ❌ 未実装 | 設計未定 |

## API連携
| メソッド | エンドポイント | 用途 |
|---------|--------------|------|
| GET | `/api/v1/examinations/:id` | 検査詳細取得 |
| POST | `/api/v1/examinations` | 検査作成 |
| PATCH | `/api/v1/examinations/:id` | 検査更新 |
| DELETE | `/api/v1/examinations/:id` | 検査削除 |
