# CI/CD パイプライン構成書

> **目的**: 自動デプロイ、手動トリガー、障害時の確認経路を定義する。
> **読者**: 運用者・新規参加者。
> **最新更新**: 2026-07-23
>
> Backendの正系統はCloudflare Workers + Containers、FrontendはVercel。
> 旧AWS ECS/RDS経路と関連workflowは2026-07-20に廃止済み。

## 1. 全体フロー

| コンポーネント | 実行環境 | デプロイ方式 | 自動トリガー | Workflow |
|---|---|---|---|---|
| Backend API | Cloudflare Workers + Containers | `wrangler deploy` + migrate one-shot + `/health` polling | `staging`への対象path push | `.github/workflows/backend-deploy.yml` |
| Frontend | Vercel | Vercel CLI | `staging` / `production`への対象path push | `.github/workflows/frontend-deploy.yml` |

移行経緯は[`../infra/_archive/migration-cloudflare.md`](../infra/_archive/migration-cloudflare.md)を参照する。
現在のインフラ構成と運用手順は[`../infra/README.md`](../infra/README.md)を正本とする。

## 2. Backendパイプライン

### 2.1 実行ステップ

1. Checkout
2. pnpm / Node.js setupと依存関係の取得
3. `CLOUDFLARE_API_TOKEN`の存在確認と`wrangler whoami`
4. `backend/`をworking directoryとして`npx wrangler deploy`
5. `MIGRATE_RUN_SECRET`を使った`infra/scripts/cf-run-migrate.sh`
6. `/health`がHTTP 200かつ`status: ok`になるまでpolling
7. 認証情報が設定されている場合だけ`infra/scripts/cf-crud-smoke.sh`

deploy直後からmigration完了まで（最大`MIGRATE_TIMEOUT=150s`）新binaryが旧schemaへ到達し得る制約は、
現行Cloudflare構成の既知・受容済み制約である。workflowはmigrationをhealth checkより前へ置き、
この区間を最小化する。詳細は`.github/workflows/backend-deploy.yml`の契約コメントを正本とする。

### 2.2 手動実行と障害対応

```bash
gh workflow run backend-deploy.yml --ref staging
```

- 失敗したjobを成功扱いにせず、deploy、migration、health、smokeのどこで失敗したかを切り分ける
- DB reset、credential変更、production操作は別途明示承認を得る
- 停止・復旧・rollback判断は[`../infra/staging/runbook.md`](../infra/staging/runbook.md)に従う

## 3. Frontendパイプライン

`.github/workflows/frontend-deploy.yml`がVercel CLIをGitHub Actions上で実行する。
Vercelのnative Git連携hookは使用しない。

1. `staging` / `production`への対象path push、または`workflow_dispatch`を検知
2. `vercel pull`で対象environmentの情報を取得
3. `vercel build`でartifactを生成
4. `vercel deploy --prebuilt`で配布

手動実行時はworkflowの`environment` inputで`preview`または`production`を選ぶ。
production実行は外部変更として明示承認を必要とする。

## 4. Security

- Cloudflareのcredential、DB接続情報、migration secretはCloudflare SecretsまたはGitHub Encrypted Secretsで管理する
- Vercel token、org ID、project IDはGitHub Encrypted Secretsで管理する
- secretをworkflow、repository、logへ直接記載しない
- workflowの権限はjobに必要な最小権限へ限定する
