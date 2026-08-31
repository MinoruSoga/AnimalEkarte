# HISTORICAL — インフラ再編計画（DO NOT EXECUTE）

> 2026-07-20 に起案された履歴資料。現在の操作手順・architecture SSOT ではない。現行契約は [README.md](README.md)、[architecture.md](architecture.md)、[iac-guidelines.md](iac-guidelines.md)。外部 cloud/account 状態は未検証。

## Completion matrix at HEAD `70dc7405`

| Phase | Repository status | Notes |
|---|---|---|
| A: AWS-era code removal | completed in repository history | AWS live account/billing stateはこの文書では検証しない。旧 paid RDS snapshot の観測は superseded であり current resource として扱わない |
| B: modules/envs/remote state | **not implemented** | Terraform は flat STG + `production/`。remote state/locking も未実装 |
| C: docs consolidation | completed, then archives deleted | `_archive` は現在存在しない。2026-08-20 の判断で git history のみにした |
| D: drift CI/token split | **not implemented or external-verification-required** | drift workflow は未実装。token の外部状態は repo から確認不能 |

## Superseded target details

- 旧案の shared `hyperdrive/` module は撤回する。Hyperdrive を再導入しない。
- `migration-cloudflare.md` と AWS archive は物理ファイルではない。履歴調査時のみ `git show e0260d32f^:docs/ops/infra/_archive/` を使い、実行手順にしない。
- remote R2/S3 backend の lock は設計・実装・検証が終わるまで「有効」と表現しない。
- Phase A の zero-match grep は当時の検証記録であり、現在は history text を意図的に match するため gate として再利用しない。

## Remaining proposals

1. provider support と state migration plan を review して env separation/remote state を設計する。
2. owner/issue を定め、read-only plan の権限と通知を設計して drift CI を実装する。
3. token split は外部 secret operation のため人が実施・検証し、値を記録しない。

これらは proposals であり、current infrastructure contract ではない。
