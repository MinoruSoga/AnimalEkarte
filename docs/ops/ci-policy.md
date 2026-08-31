# CI ポリシー — ゲート分担と Actions ピン記法

> **目的**: リモート CI 必須ゲートとローカル必須チェックの分担、および GitHub Actions のピン記法方針(#195)を定義する。
> **読者**: CI ワークフローを変更する開発者、および PR 前に品質チェックを走らせる開発者。
> **タイミング**: `.github/workflows/*.yml` 編集時、または `make ci` / `make lint` の使い分け確認時。

## CI 必須 vs ローカル必須

リモート CI は **「STG に壊れたものを載せない」薄いゲート** に寄せる。  
手元 Docker で十分再現できる静的検査・inventory・lint は **`make ci` に集約**し、PR Checks の行数と実行時間を抑える。

| 区分 | 内容 | どこで担保するか |
|---|---|---|
| **リモート CI 必須** | path-filtered `build` / `test` / coverage ratchet、gitleaks（secret-scan）、codegen / migration 検証、schema drift（backend job 内）、AgentShield（**agent-config 変更時のみ fail** · 詳細 `README-security-scan.md`） | `.github/workflows/ci.yml` / `security-scan.yml` |
| **ローカル必須 (`make ci`)** | clinic/audit inventory・preload、reset-contract、ci-step-order、Go coverage merge contract、docs-symbol-drift、eslint-disable rationale、shellcheck、design CTA、golangci-lint、ESLint / type-check / knip、codegen 同期、backend/frontend build+test（手元再現） | `make ci`（実体: `scripts/run-local-ci.sh`） |
| **ローカル任意 (E2E)** | Playwright ブラウザ E2E（スタック起動が重く、PR Checks のノイズ・時間になりやすい） | `make e2e`（`frontend/scripts/run-e2e.sh`）。`.github/workflows/e2e.yml` は **workflow_dispatch のみ**（push/PR 自動実行なし） |
| **schedule のみ** | 負荷・プロファイリング（performance） | `.github/workflows/performance-tests.yml`（push トリガなし・`workflow_dispatch` 可） |

### リモート CI のジョブ（`ci.yml`）

| Job | 役割 |
|-----|------|
| Detect changes | paths-filter |
| Gitleaks Secret Scan | 秘密情報漏洩（リモート必須） |
| Backend Build | backend build（DB不要、test shards と並列） |
| Backend Test (matrix) | 独立PostgreSQLを持つ4 shard。各shard内は `-race -coverpkg=./internal/... -p 1` |
| Backend | 4 coverage profileの重複blockを統合し、ratchetを実行する集約check |
| Frontend Build | frozen-lock install + audit + build |
| Frontend Test (matrix) | Vitest 2 shard。blob reporterへcoverage payloadを保存 |
| Frontend | blobをVitestでnative mergeし、coverage ratchetを実行する集約check |
| Worker Tests | worker unit tests（paths 該当時） |
| Codegen Sync | Go model ↔ TS 型同期（paths 該当時） |
| Migration Verify | PR→main かつ migration 変更時のみ |

### PR / push の実行契約

- `main` 向けPRを含む、`main` / `staging` / `production` 向けPRで変更層のBuild/Testを省略しない。依存更新PRも同じゲートを通す。
- `main→staging` のopen PRがpush SHAと同じheadを既に検証する場合、push側の重いBuild/Test/Worker/Codegenだけを省く。GitHub API確認に失敗した場合はfail-openでpush CIを実行する。
- Backendは4つの独立DB shard、Frontendは2つのVitest shardで実行する。集約jobの表示名`Backend` / `Frontend`はbranch protectionとの互換性のため維持する。
- Frontend installは`pnpm install --frozen-lockfile`を必須とし、PRで検証したlock graphとCIの解決結果を一致させる。
- shardの完全ログはActions consoleへ無制限出力しない。consoleは失敗抜粋と末尾へ制限し、完全ログはgzip artifact（7日保持）で取得する。job timeoutで暴走を停止する。

### なぜ inventory / guardrail をリモートから外したか

- Docker 不要または `docker compose exec` で **ローカル完全再現**できる
- 独立ジョブにすると PR Checks が 20 行超になり、失敗/スキップのノイズが増える
- 再発防止の価値は **実行されること** にあり、実行場所はローカルで足りる
- リモートに残すのは **共有環境での信頼が必要なもの**（gitleaks、fresh DB 上の unit/integration test、coverage）
- E2E は Docker フルスタック + ブラウザで重いため **ローカル `make e2e`** に寄せる（自動 PR ゲートにしない）

### 開発者の使い方

```bash
make up          # コンテナ起動
make ci          # ローカル一括 CI（push / PR 前に実行）
make e2e         # Playwright E2E（任意・要 make up）
make lint        # Go lint のみ
make lint-front  # FE 静的のみ
```

`make help` の品質管理セクションにも同じ分担を要約してある。

（以下は #195 の積み残し「ピン記法ポリシー決定＋文書化」の決定記録）

## ピン記法の基準

| 対象 | 記法 | 理由 |
|---|---|---|
| GitHub 公式 actions（`actions/*`） | メジャータグ（例: `actions/setup-node@v6`） | 公式の改竄リスクは低く、パッチ追従の利便を優先 |
| ベンダー公式 actions（`aws-actions/*`、`golangci/*`、`pnpm/*` 等） | メジャータグまたは完全 semver（例: `@v6.1.0`） | 準公式。完全 semver 併用可（再現性優先の場合） |
| サードパーティ actions（個人・小規模組織） | **コミット SHA ピン**＋バージョンをコメント明記 | サプライチェーン対策（タグは付け替え可能、SHA は不変） |
| シェルからの外部スクリプト取得（curl 等） | **リモート pipe-to-shell 禁止**（下表）。Release artifact は **バージョン + SHA-256** で固定 | 未ピン / `main`/`master` 参照 / `curl\|sh` は改竄・非再現の両リスク |

### リモート pipe-to-shell 禁止（Dockerfiles / workflows）

次は **Dockerfile・GitHub Actions workflow で禁止**する（サプライチェーン）。

| 禁止パターン | 例 | 代替 |
|---|---|---|
| `curl … \| sh` / `wget … \| sh` | `curl -sSfL …/install.sh \| sh` | 公式 pin イメージ、または Release tarball + checksum |
| `bash <(curl …)` process substitution | `bash <(curl -Ls …/download-….bash)` | Release バイナリを SHA-256 検証してから実行 |
| `raw.githubusercontent.com/…/(master\|main\|HEAD)/…` の pipe install | golangci-lint `master/install.sh` | ローカルは `make lint` の `golangci/golangci-lint:<pin>` イメージ。CI は action または checksum 付き release |

機械的検査:

- `scripts/check-workflow-remote-exec-policy.sh [repo-root]` — 上記パターンを Dockerfiles / workflows から検出して fail
- `scripts/check-agent-security-policy.sh [repo-root]` — `.cursor/permissions.json` の `approvalMode: unrestricted` / `mcpAllowlist: ["*:*"]`、`.cursor/sandbox.json` の `networkPolicy.default: allow` を fail

actionlint の導入例（許可）: GitHub Releases の `actionlint_<ver>_linux_amd64.tar.gz` を `sha256sum -c` で検証してから展開（`.github/workflows/actionlint.yml`）。

## 運用ルール

1. **新規追加・既存変更時に本基準を適用する**（ratchet 方式）。既存の `uses:` を基準準拠のためだけに一括書き換えるスイープは行わず、該当ワークフローを触る際に合わせる。
2. 同一 action は全ワークフローで**単一バージョンに統一**する（#195 で達成済みの状態を維持）。ドリフト検出は `scripts/check-actions-version-drift.sh`（actionlint.yml の CI job 内で実行・混在を fail）で行う。**actionlint 自体はバージョンドリフトを検出しない**（構文・式チェックのみ — backend-deploy.yml 新設時に @v4 が混入し #195 が回帰した実績あり）。env/with/working-directory のドリフトはスクリプト対象外のため PR レビューで見る（既知の盲点）。
3. バージョン更新は四半期ごとに棚卸しする（actionlint の固定バージョン、checksum ピンの追従を含む）。
4. **新しい静的ゲートを増やすときは `make ci`（`scripts/run-local-ci.sh`）に足す。** リモート `ci.yml` に独立ジョブを増やすのは、fresh DB / 秘密スキャン / 共有ランナーでしか意味がない場合に限る。
5. **リモート pipe-to-shell を新規に入れない。** Dockerfile への `curl\|sh` や workflow の `bash <(curl …)` は PR で reject。golangci-lint は in-image install せず `make lint` の pin イメージを使う。

## 現状の準拠状況（2026-07-17 更新）

- `ci.yml` を軽量化: inventory / guardrail / shellcheck / design CTA をローカル `make ci` へ移管（PR Checks は Detect + Gitleaks + Backend/Frontend/Worker/Codegen/Migration 程度）
- 2026-07-05 の backend-deploy.yml 新設（Cloudflare Phase 5）で `setup-node@v4` / `pnpm/action-setup@v4` が混入し #195 が一時回帰 → 2026-07-07 に @v6 へ是正し、`check-actions-version-drift.sh` を CI 追加（再発は CI が fail させる）。

### 2026-07-02 時点の記録

- 統一済み: `actions/checkout@v7.0.0` / `actions/setup-node@v6` / `actions/setup-go@v6` / `actions/upload-artifact@v7` / `pnpm/action-setup@v6` / `aws-actions/configure-aws-credentials@v6.1.0` / `aws-actions/amazon-ecs-render-task-definition@v1.8.5` / `dorny/paths-filter@v4`
- 2026-07: `golangci/golangci-lint-action` は CI ゲートから外しローカル `make lint` / `make ci` に寄せた（ピン記法自体は再導入時に本基準を適用）
- SHA ピン済み: security-scan.yml の agentshield
- 既知の未収束: `dorny/paths-filter@v4`（サードパーティ）は本基準では SHA ピン対象 — 次回 paths-filter を触る PR で SHA 化する
