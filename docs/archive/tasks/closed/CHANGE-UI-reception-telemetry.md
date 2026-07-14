# CLOSED: 受付ヘッダー テレメトリ（change-ui）

- **完了日**: 2026-07-15（台帳同期時に実装済みを確認）
- **ルート台帳**: [`change-ui.md`](../../../../change-ui.md)（短い完了ポインタのみ）
- **アクティブ仕様**: [`docs/screens/01-reception.md`](../../../screens/01-reception.md) §3

## 実装サマリ

| Phase | 内容 | 状態 |
|-------|------|------|
| 1 | `ReceptionTelemetryStrip` で「本日受付 N件」（`columns` 全体値） | 実装済み |
| 2 | `appointments.checked_in_at` + 平均/最長待ち表示 | 実装済み（`RECEPTION_TELEMETRY_PHASE2_ENABLED = true`） |

## 受け入れ基準（コード裏取り）

| 基準 | 結果 |
|------|------|
| 本日受付がフィルタ非影響 | `Reception.tsx` は `columns` を使用 |
| 受付済 0 件で平均・最長が「—」 | Strip + 単体テストあり |
| 予約編集で待ち時間が不変 | BE: 非ステータス更新で `checked_in_at` 不変。FE 楽観更新で保持 |
| 再受付で待ちリセット | BE: 再 `checked_in` で上書き |
| migration は新規ファイルのみ | **逸脱記録**: 当初 `005` 追加後 `001_init.sql` に再統合（プロジェクト方針）。ERD に記載 |

## 未決（次期）

- `RECEPTION_TELEMETRY_PHASE2_ENABLED` キルスイッチ削除可否 → `FE-refactor.md`
