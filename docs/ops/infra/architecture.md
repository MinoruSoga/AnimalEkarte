# インフラ構成（現行・Cloudflare）

> **目的**: 現行インフラの全体像。**読者**: 全開発者。**タイミング**: インフラに関わる変更・調査の最初。
> AWS 時代の構成は git 履歴参照（2026-07-20 に全廃止・課金停止済み。2026-08-20 に記録も削除 — `git show e0260d32f^:docs/ops/infra/_archive/aws-legacy/` 配下）。

## 全体像

```
利用者
  ↓ HTTPS
Vercel (Frontend SPA: stg.noah-karte.com / 本番ドメイン)
  ↓ /api rewrite
Cloudflare zone (noah-karte.com · NS/DNS/SSL/proxy)
  ↓ Workers ルート (api.stg.noah-karte.com/*)
Worker (薄いプロキシ + /_internal/migrate)
  ↓ Durable Object binding
Container (Go/Gin API · Dockerfile.production)
  ↓ 直結 (Hyperdrive 不可 · sslmode=require)
PlanetScale Postgres (ap-northeast/東京)         R2 (臨床画像 · S3互換)
```

## 環境

| | STG | PROD |
|---|---|---|
| 状態 | **稼働中**（2026-07-17 切替完了） | **未構築**（#253・[production/setup.md](production/setup.md)） |
| API | api.stg.noah-karte.com（proxied） | api.noah-karte.com（予定） |
| Worker | `animalekarte-stg-api` | `animalekarte-prod-api`（予定） |
| DB | PlanetScale `animalekarte-stg`（フルデモ投入済み） | 未作成 |
| R2 | `animalekarte-stg-images` | `animalekarte-prod-images`（予定） |
| デプロイ | staging push → `backend-deploy.yml` 自動 | 未整備 |

## 主要設定値と根拠

- **SSL**: Full (strict)。Universal SSL は 1 階層まで — `api.stg.〜`（2 階層）は **ACM($10/月) の `*.stg` 証明書**でカバー（本番 `api.〜` は無料枠で足りる）
- **Container**: `instance_type: basic`(1/4 vCPU/1GiB)・`max_instances: 3`・`sleepAfter: 10m`（scale-to-zero）
- **DB 接続プール**: `DB_MAX_OPEN_CONNS=10 / DB_MAX_IDLE_CONNS=5`（wrangler vars で注入）。**直結・プーラー無しのため低値必須** — 2026-07-17 のスロット枯渇全断の恒久対処。本番も同方針
- **secrets**: `wrangler secret put` または GitHub Actions `worker-secret-sync.yml`（GH Secrets → Worker）。`vars` は非機密のみ

## 既知の制約（運用上の注意）

1. **Containers は Hyperdrive 不可**（公式 issue #97）→ DB 直結。プール×インスタンス数が PlanetScale 接続上限を超えないこと
2. **ローリング更新は非同期** — デプロイ直後のリクエストは旧イメージに当たりうる。イメージ更新を伴う検証は 15 分静置（sleepAfter 超過）後に行う
3. 障害記録・教訓の一次資料: [_archive/migration-cloudflare.md](_archive/migration-cloudflare.md) の P7-2 観測 #1/#2
