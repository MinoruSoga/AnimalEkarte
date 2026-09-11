# 認証・認可レビューの未完了TODO

> 作成日: 2026-09-08  
> 最終整理: 2026-09-11  
> 統合コミット: `a0e569a9a`（`main`）  
> 正規設計書: [docs/architecture/auth.md](docs/architecture/auth.md)  
> 元レビュー: [todo-check-auth.md](todo-check-auth.md)

本書は現在状態と実装証跡の入口だけを持つ。外部境界の検証・実行・判定条件は [todo-verification.md](todo-verification.md#認証認可の外部境界) を正本とする。

## 現在の状態

| 区分 | 状態 |
| --- | --- |
| ローカル実装 + disposable 実DB（D2 / D3 / D5 / D1 合成） | **完了**（`a0e569a9a`） |
| D1 本番付与・対象環境メール | **未完了**（対象環境未定・別承認） |
| Linear ライブ特定・投稿 | **BLOCKED**（workspaceの free issue limit exceeded） |

**LOCAL COMPLETE / EXTERNAL INCOMPLETE**。D1本番付与・メール、Linearのread/writeは統合検証TODOの順序・外部承認に従う。

## 実装・ローカル証跡

- 実装基準の起点は `a4292a241`。2026-09-10に disposable Postgresで D2 / D3 / D1合成 / D5を実行し、`a0e569a9a`に統合した。
- D3はGET/HEAD 178経路の実DB返却データ分離、D2は自己ロックアウト防止の並行検証、D5は独立2プロセスでの即時失効、D1合成はsuccess / admin conflict / email conflict / audit rollbackを対象とした。
- このローカル証跡は対象環境・本番・メール・Linear反映のPASSを示さない。詳細はGit履歴と統合検証TODOを参照する。

## 参照

- [統合検証TODO](todo-verification.md#認証認可の外部境界)
- [初回管理者手順](docs/ops/deploy/FIRST_SYSTEM_ADMIN.md)
- [D1 SQL fixture](backend/internal/auth/testdata/first_system_admin.sql)
