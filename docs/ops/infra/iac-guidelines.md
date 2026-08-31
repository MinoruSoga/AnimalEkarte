# IaC 運用規約（Cloudflare）

## 境界

| 層 | 管理対象 | 禁止 |
|---|---|---|
| Terraform (`infra/cloudflare/`) | zone、DNS、R2、provider が対応する WAF/ruleset/notification | Worker/Container、secret、Hyperdrive を管理しない |
| Wrangler (`backend/wrangler*.jsonc`) | Worker、Container、route、binding、secret name contract | DNS などの基盤を持たせない |

Hyperdrive の STG/PROD files は tombstone である。Containers では利用できず、credential が state に入るため再導入しない。

## State と drift

- tfstate を git 管理しない。secret を state に持つ resource を作らない。
- HEAD は local backend を使用する。env-separated remote backend と locking は **required future state / not implemented at reviewed baseline `70dc7405`**。実装前に backend/locking の設計と owner/issue を review する。
- 定期 `terraform plan` drift detection workflow も **not implemented**。現行の自動 gate と断定しない。
- dashboard change は原則避ける。やむを得ない場合は承認と記録を残す。provider/resource が support する場合だけ import する。support しない場合は revert、または reviewed state-migration plan で codify する。

## Secrets and tokens

- Worker secret は Cloudflare Secrets または承認済み GitHub Environment secret 経路で投入する。値は docs/state/log に残さない。
- token は用途別・environment 別・最小権限とする。
- credential rotation は投入先更新、再 deploy、verification を一体で行う。

## Change flow

plan → review → explicit human approval → apply。destroy、credential change、shared environment apply は agent が実行しない。provider/Wrangler version を pin する。
