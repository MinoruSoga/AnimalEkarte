# 集計・締め

> **専用ドキュメント**: `docs/tasks/pending/accounting/FEAT-368_closing-aggregation.md`
>
> 仕様・DB設計・API設計・未確認事項はすべて上記を参照。

## 関連画面

| 画面 | URL | 説明 |
|------|-----|------|
| レジ締め | `/accounting/close` | AM/PM別締め・レジ金突合・印刷 |
| 締め履歴 | `/accounting/close/history` | 過去の締めレコード一覧 |
| 月次集計レポート | `/accounting/reports` | 経理ロールのみ。月間売上・CSV エクスポート |
| 締め時間設定 | `/settings/closing-time` | 境界時刻・特別期間の管理（管理者のみ） |

## 関連設定仕様

- `docs/screens/settings/closing-time-settings.md` — 締め時間設定画面の詳細
