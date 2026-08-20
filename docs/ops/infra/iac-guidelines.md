# IaC 運用規約（Cloudflare）

> **目的**: インフラをコードで安全に管理するための規約。**読者**: インフラ変更を行う全員。**タイミング**: `infra/` または `wrangler*.jsonc` に触れる前（MANDATORY）。

## 1. 2 層の境界（最重要）

| 層 | 管理対象 | 禁止事項 |
|---|---|---|
| **Terraform**（`infra/cloudflare/`） | ゾーン・DNS・R2・Hyperdrive・WAF/ruleset・通知 | Workers/Containers を Terraform で管理しない（デプロイと競合） |
| **Wrangler**（`backend/wrangler*.jsonc`） | Worker・Container・ルート・bindings・secrets | インフラの土台（DNS 等）を Wrangler 側に持たない |

## 2. state

- **リモート backend 必須**（R2 の S3 互換 backend。現状 local は Phase B で解消 — `reorg-plan.md`）
- **env ごとに state key を分離**（`stg/` と `prod/` を同じ state に入れない）
- tfstate を git 管理しない（属性が平文で入る）。state に secret を持つリソースを作らない

## 3. secrets / token

- Worker secrets は `wrangler secret put` か `worker-secret-sync.yml`（GH Secrets 経由）。`vars` は非機密のみ
- API トークンは**用途別・env 別・最小権限**で分割（Terraform 用 / CI deploy 用）。STG の統合トークン 1 本運用は暫定 — 本番では必ず分離
- 資格情報ローテーション（`pscale role reset-default` 等）は必ず「投入先の更新+再デプロイ」までをセットで実施

## 4. ドリフト

- **手動ダッシュボード操作は原則禁止**。やむを得ない場合は実施記録を残し `cf-terraforming import` で state に取り込む
  （実害例: 手動作成の CloudFront が AWS destroy を 409 でブロックした）
- 定期 `terraform plan` によるドリフト検知 CI を置く（docs/ops/infra/reorg-plan.md Phase D）

## 5. 変更フロー

- plan → 差分確認 → **明示承認** → apply。destroy・credential 変更は必ず停止して承認を得る
- 共有 env へのローカル apply は避け、CI 経由（PR で plan・merge で apply）へ寄せる（Phase D）
- provider / wrangler のバージョンを pin する（Cloudflare provider v4→v5 破壊的変更の前例）
