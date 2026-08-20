# 今期外・見送り索引

> **目的**: 今フェーズでやらない判断の短い索引。  
> **全文（2026-08-20 時点 HTML）**: CorpVault `50_Projects/ノア動物病院電子カルテ/evidence/2026-08-20-root-docs/phase2.html`  
> **実行 SoT**: Linear（hub [BRT-4](https://linear.app/baritechllc/issue/BRT-4)）

再開するときは Linear に新 Issue を切る。本ファイルへ作業結果を書き足さない。

## 運用上残している再開条件

### PERF-AUDIT-TX P2（outbox パターン移行）— 見送り

- **再開条件**: `audit_write_failed` が恒常的（目安: 月 1 件以上継続）に観測された場合、実測頻度 1 か月分を添えて再起案。
- P0/P1 は完了済み（2026-07-12 CLOSED）。詳細経緯は git 履歴。

運用監視の参照元: [STG-CONTINUOUS-OPERATIONS.md](../ops/deploy/STG-CONTINUOUS-OPERATIONS.md)

## 索引（HTML から圧縮）

| 区分 | 内容 |
|------|------|
| PO 決裁済み「やらない」 | 旧 todo.md §7。正本は Linear |
| 条件付き再開 | BUG-416-A/B、BUG-413-A/B、FEAT-VACCINE-SPECIES（DEC-3）、#239 Phase 2（DEC-46） |
| BE 見送り | PERF-AUDIT-TX P2 |
| BE 次期 | 第7期監査引き継ぎ・層違反疑い・god-function 走査 |
| FE 次期 | 第6期監査引き継ぎ |
| インフラ保留 | HTML 本文 |

第7期で確定した「やらない」（次期でも踏襲）:

- duration リテラル / `INTERVAL` SQL の一括定数化はしない
- `respondError` と handler `RespondError` の統一はしない（X-17）
- `isDog/isCatSpeciesName` と `doseSpeciesAliases` は契約が異なるため統合禁止
- 会計 post-close 権限を service へ機械移動しない
