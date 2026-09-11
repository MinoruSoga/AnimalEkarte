# 認証・認可レビューの未完了TODO

> 作成日: 2026-09-08  
> 最終整理: 2026-09-11  
> 正規設計書: [docs/architecture/auth.md](docs/architecture/auth.md)  
> 元レビュー: [todo-check-auth.md](todo-check-auth.md)

ローカル実装（D2 / D3 / D5 / D1 合成）は完了済み（Git 履歴）。外部境界の検証・実行条件は [todo-verification.md](todo-verification.md#認証認可の外部境界) を正本とする。

## 未完了

| 区分 | 状態 |
| --- | --- |
| D1 本番付与・対象環境メール | **未完了**（対象環境未定・別承認） |
| Linear ライブ特定・投稿 | **BLOCKED**（workspace free issue limit / 新規作成禁止。既存チケット更新のみ） |

**EXTERNAL INCOMPLETE**。順序・承認は統合検証TODOに従う。

## 参照

- [統合検証TODO](todo-verification.md#認証認可の外部境界)
- [初回管理者手順](docs/ops/deploy/FIRST_SYSTEM_ADMIN.md)
- [D1 SQL fixture](backend/internal/auth/testdata/first_system_admin.sql)
