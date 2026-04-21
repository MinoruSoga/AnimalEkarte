# 集計・締め

> **専用ドキュメント**: `docs/tasks/pending/accounting/FEAT-368_closing-aggregation.md`
>
> 仕様・DB設計・API設計・未確認事項はすべて上記を参照。

## 関連画面（実装済）

| 画面 | URL | 説明 |
|------|-----|------|
| レジ締め | `/accounting/close` | **実装済**。AM/PM別締め・レジ金突合・印刷（プレビュー） |
| 締め履歴 | `/accounting/close/history` | **実装済**。過去の締めレコード一覧 |
| 月次集計レポート | `/accounting/reports` | **実装済**。経理ロールのみ。月間売上・部門別集計 |
| 締め時間設定 | `/settings/closing-time` | **実装済**。境界時刻・休診日・特別期間の管理 |
| 支払方法マスタ | `/settings/payment-methods` | **実装済**。現金・カード等の支払区分管理 |

## 関連設定仕様

- `docs/screens/settings/closing-time-settings.md` — 締め時間設定画面の詳細
