# 今期外・見送り索引

> **目的**: 今フェーズでやらない判断の短い索引。
> **全文（2026-08-20 時点 HTML）**: CorpVault `50_Projects/ノア動物病院電子カルテ/evidence/2026-08-20-root-docs/phase2.html`
> **実行 SoT**: Linear（hub [BRT-4](https://linear.app/baritechllc/issue/BRT-4)）

再開時は担当者と再開条件を持つ Linear Issue を新規作成する。本ファイルへ作業結果を書き足さない。状態が競合した場合は Linear を正とする。

## 運用上残している再開条件

### PERF-AUDIT-TX P2（outbox パターン移行）— 見送り

- **再開条件**: `audit_write_failed` が恒常的（目安: 月 1 件以上継続）に観測された場合、実測頻度 1 か月分を添えて再起案する。
- **owner / action target**: Linear hub [BRT-4](https://linear.app/baritechllc/issue/BRT-4) から担当者付き Issue を作成し、本項へ直リンクを追加してから着手する。担当者未定のまま再開しない。
- **監視契約**: [STG-CONTINUOUS-OPERATIONS.md の `audit_write_failed`](../ops/deploy/STG-CONTINUOUS-OPERATIONS.md#24-監査書込失敗-audit_write_failed-監視perf-audit-tx-p1) を参照する。
- P0/P1 は repo 履歴上 2026-07-12 CLOSED。実行時の状態は Linear で再確認する。

## 削除した legacy 索引

旧 `todo.md §7` と `BUG-416-A/B`、`BUG-413-A/B`、`FEAT-VACCINE-SPECIES`、`DEC-3`、`DEC-46`、「第6期 / 第7期」等のラベルは、現行 repo 内に owner・直接参照・安定した定義がなく、再開可能な台帳ではなかったため本索引から削除した。必要な判断は上記 CorpVault snapshot または git 履歴を調査し、**直接 Linear URL・一行の再開条件・named owner** を持つ新 Issue として再起案する。legacy ラベルだけを復活させない。

次の「やらない」判断は現行ソース契約として維持する。変更する場合は通常の仕様変更として Linear に起案する。

- duration リテラル / `INTERVAL` SQL の一括定数化はしない
- `respondError` と handler `RespondError` の統一はしない（旧称 X-17 は検索 key に使わない）
- `isDog/isCatSpeciesName` と `doseSpeciesAliases` は契約が異なるため統合しない
- 会計 post-close 権限を service へ機械移動しない
