# インフラ構成 — checked-in Cloudflare configuration

> この文書は HEAD の設定を説明する。Cloudflare、PlanetScale、Vercel、課金、証明書、DB 内容などの live state は証明しない。運用前に人が日付と証跡を付けて検証する。

## Topology

```text
STG user
  -> Vercel SPA (checked-in frontend/vercel.json rewrite is STG-specific)
  -> /api rewrite -> api.stg.noah-karte.com
  -> Cloudflare Worker -> Durable Object -> Container (Go/Gin)
  -> direct PlanetScale Postgres connection (sslmode=require)
  -> R2 for clinical images

PROD
  -> planned api.noah-karte.com topology
  -> production Wrangler/Terraform files remain drafts until external verification
```

Hyperdrive は Containers から利用できず、credential を Terraform state に載せるため再導入しない。

## Repository configuration state

| | STG config | PROD config |
|---|---|---|
| Worker | `animalekarte-stg-api` | `animalekarte-prod-api` |
| API route | `api.stg.noah-karte.com/*` | `api.noah-karte.com/*` planned |
| R2 | `animalekarte-stg-images` | `animalekarte-prod-images` planned |
| Container | `basic`, max 3, `sleepAfter = "10m"` | production draft |
| DB pool | direct connection | max open 10 / idle 5 in production draft |

“STG 稼働中”“PROD 未構築”、DB の存在や内容は runtime observation であり、repo config から断定しない。現行 migrate の seed bundle は全環境で `002_master` のみ。既存 STG データは別途確認する。

## Historical observations, not current guarantees

2026-07 の certificate、connection-slot、schema-owner の記述は当時の observation だった。価格、certificate coverage、schema ownership、恒久解は現在の公式 provider documentation と runtime evidence を人が再確認するまで運用判断に使わない。

## History

リポジトリ上の決定では AWS ECS/RDS は 2026-07-20 に廃止され、rollback target として使わない。AWS-era docs は 2026-08-20 に削除された。調査時のみ `git show e0260d32f^:docs/ops/infra/_archive/aws-legacy/` を参照し、手順として実行しない。これは repository history であり live AWS account state の検証ではない。
