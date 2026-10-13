# U05 — C14 依存・lockfile・CI 供給網完全性 判定レコード

Campaign: emr-332-security-defense / Unit: U05-FILES-SSRF-SUPPLY / Attempt: att-u05-1
Format: `docs/work/emr/emr-332-security/baseline/case-verdict-template.md` (12 fields)

## 判定レコード (12 fields)

| # | Field | 記録内容 |
|---|-------|----------|
| 1 | ケース ID | `C14` — dependency / lockfile / CI workflow / deploy provenance / supply-chain integrity |
| 2 | 対象リビジョン | worktree `feat/emr-332-security-defense-w05` HEAD `fc6acc6b8d267eb51050a4848069e1a1b569e7df` (read-only 静的レビュー対象) |
| 3 | 実施方式 | `static` — ファイル読取のみ。`go list`/`pnpm audit`/`npm install`/`go vuln`/Dependabot 照合・workflow 実行は一切未実施 |
| 4 | 正常条件 | 期待: 依存が pin され、lockfile が存在し、deploy が lockfile 検証済み node_modules 由来の wrangler を使い、environment/branch gate で production を制限。観測: 後述 Section A — go.mod pin+zero replace/exclude、go.sum 存在、pnpm-lock v9 (root+frontend)、wrangler 4.114.0 厳密 pin、frozen-lockfile→wrangler deploy、environment+branch gate 全て確認 |
| 5 | 拒否条件 | 期待: lockfile 不一致で install 拒否 (frozen-lockfile)、非 production ref からの production deploy 拒否、無許可権限の deny。観測: `pnpm install --frozen-lockfile` が deploy 前に置かれている (backend-deploy.yml:74,208, frontend-deploy.yml:70-71 ほか)。production 非許可 ref を fail する step あり (backend-deploy.yml:188-191, frontend-deploy.yml:44-47)。`permissions:` は概ね `contents: read` で狭く宣言 |
| 6 | 防御判定 | `PASS` (pinning/integrity/provenance 構成について static 確認) — ただし vulnerability 状態は `NOT ASSESSED` (Section C)。tag-pin された first-party actions と permissions 未宣言 workflow は残余リスクとして Section D に記録 |
| 7 | 副作用呼出回数 | `0` — 本ケースは read-only レビューのみで副作用呼出対象を持たない (install/execute/deploy 未実施) |
| 8 | 検出判定 | `UNKNOWN` — CI に gitleaks (ci.yml) と agentshield (security-scan.yml) が存在するが、本レビューでは workflow の実効実行を観測していない |
| 9 | 安全な証跡参照 | 本ファイル Section A/B の file:line 参照のみ (秘密値・lockfile 全体の転写なし) |
| 10 | 制約・未検証範囲 | 脆弱性スキャン (go vuln / pnpm audit / Dependabot / OSV 等) 未実行 — 依存の known-CVE 露出は **NOT ASSESSED**。lockfile 内容の hash 妥当性 (pnpm store 照合) 未実施。workflow の GitHub 側実効 (Environment protection rules の設定値) はリポジトリ外のため未検証 — yaml 宣言のみ確認。Wrangler/CI の実 deploy 実行はなし |
| 11 | 修復案 | 構成上の必須ギャップなし。残余改善候補として Section D に記録 (first-party action の SHA-pin 化・permissions 未宣言 workflow の最小権限明記・依存脆弱性スキャンの定期実行) — いずれも候補であり本 unit では修復しない |
| 12 | 再検証結果 | 未実施 — 修復なしのため。脆弱性状態の再検証は scanner 実行を伴う別 unit |

## Section A — pinning / lockfile / provenance 確認結果

| 項目 | 証跡 | 結果 |
|------|------|------|
| `backend/go.mod` go directive | `go 1.26.0` (go.mod:3) | pinned toolchain 宣言 |
| `backend/go.mod` replace/exclude | `grep -c 'replace\|exclude'` = **0** | なし — 迂回・除外指令は存在しない |
| `backend/go.sum` | 存在 (184 行) | integrity hash 記録あり |
| root `pnpm-lock.yaml` | `lockfileVersion: '9.0'` | v9 lockfile |
| `frontend/pnpm-lock.yaml` | `lockfileVersion: '9.0'` | v9 lockfile (frontend workspace) |
| wrangler pin | `package.json:26` `"wrangler": "4.114.0"` — 厳密 version (range なし) | pinned |
| pnpm バージョン | `frontend/package.json` `packageManager: "pnpm@10.15.0"` | pinned |
| セキュリティ overrides | `frontend/package.json:102-131` `pnpm.overrides` — undici `^7.29.1`, nanoid `^3.3.18`, path-to-regexp (複数 `8.4.2`/`6.3.0`), tar `7.5.22`, ajv `8.20.0`, minimatch+brace-expansion 系, vite>postcss `8.5.25`, browserslist `4.28.7`, js-yaml `4.3.2`, smol-toml `1.9.0`, source-map-js `^1.2.2`, ts-morph `28.0.0`, fast-uri `^3.1.8` ほか + `onlyBuiltDependencies` (swc/esbuild/msw) | transitive pin による既知脆弱性対応が存在 |
| deploy 手順 install | `pnpm install --frozen-lockfile` — backend-deploy.yml:74 (STG), :208 (PROD), frontend-deploy.yml:70-71, chromatic.yml, storybook.yml, performance-tests.yml, worker-secret-sync.yml | lockfile 検証済み node_modules 由来 |
| wrangler deploy 呼出 | backend STG `npx wrangler deploy` (backend-deploy.yml:94), backend PROD `npx wrangler deploy -c wrangler.production.jsonc` (:236), frontend `npx wrangler deploy -c "$CONFIG"` (frontend-deploy.yml) — wrangler binary は lockfile-verified node_modules から。`worker-secret-sync.yml:35-38` に SEC-CS3-O7 として `npx -y wrangler@x.y.z` unpinned fetch を排除した旨の明記あり | provenance: Actions → frozen-lockfile install → pinned wrangler |
| wrangler config | `backend/wrangler.jsonc`, `backend/wrangler.production.jsonc`, `frontend/wrangler.jsonc`, `frontend/wrangler.production.jsonc` — PROD は `-c` 明示必須 (backend-deploy.yml:232-236 に config 誤配の実害記録コメント) | STG/PROD config 分離 |
| Environment gate | backend-deploy.yml `environment: staging` (:45) / `environment: Production` (:178); frontend-deploy.yml `environment: ${{ ... && 'Production' || 'staging' }}` (:37) | environment 保護 gate |
| Branch gate | backend STG `refs/heads/staging` or dispatch 非 production (:40), PROD `refs/heads/production` or dispatch target=production (:169) + 非 production ref で fail (:188-191); frontend `:44-47` 同旨 | production = production branch のみ |
| token 分離 | backend-deploy PROD job で `PROD_*` secret env (frontend-deploy.yml の `PROD_CLOUDFLARE_API_TOKEN` も同様) | STG/PROD credential 分離 |
| workflow `permissions` | backend-deploy.yml:29, frontend-deploy.yml:38-40 (`contents: read` + `deployments: write`), worker-secret-sync `contents: read`, security-scan `contents: read` (+ 個別 job scope), plane-merge-state `contents: read`+`pull-requests: read`, ci gitleaks job `contents: read`+`pull-requests: read`, chromatic `contents: read`, storybook `contents: write` (gh-pages) | 概ね最小権限で宣言 |
| 残余 (documented) | backend-deploy.yml:96-118 — wrangler deploy は即時切替 (段階 rollout なし, DO binding 制約)。deploy→migrate ≤150s window が user-accepted residual | known-issue 明記済み |

## Section B — action pin 状況

| 区分 | actions |
|------|---------|
| SHA pin (immutable) | `dorny/paths-filter@ceb8a2b8…`, `affaan-m/agentshield@b089130…`, `chromaui/action@4784f91…`, `peaceiris/actions-gh-pages@84c30a8…` |
| tag pin (mutable — 残余リスク) | `actions/checkout@v7.0.1`, `actions/setup-node@v7`, `actions/setup-go@v7`, `actions/upload-artifact@v7`, `actions/github-script@v9`, `pnpm/action-setup@v6`, `gitleaks/gitleaks-action@v3` |
| `permissions` 未宣言 | `e2e.yml`, `performance-tests.yml`, `stg-migrate-direct.yml`, `stg-smoke.yml`, `actionlint.yml` — repo default token に依存 (repo 設定はリポジトリ外で確認不可) |

## Section C — vulnerability status (pinning 判定と分離)

- **`NOT ASSESSED`** — `go vuln`, `pnpm audit`, Dependabot データ, OSV/GHSA 照合, `gosec` を本 unit で実行していない。依存の既知 CVE 露出は本レコードの結論対象外。
- pnpm overrides はセキュリティ対応として存在する (Section A) が、各 override が現在の脆弱性を完全に解消しているかの照合は network/data 依存のため未実施。

## Section D — remediation candidates (修復せず記録のみ)

| ID | 対象 | 候補 |
|----|------|------|
| RC-1 | first-party actions の tag pin | `actions/*`・`pnpm/action-setup`・`gitleaks-action` を SHA pin に寄せる候補 (mutable tag の差替え耐性向上) |
| RC-2 | `permissions` 未宣言 workflow | `e2e.yml` 等に明示的 `permissions:` を宣言する候補 (repo default 依存の解消) |
| RC-3 | 依存脆弱性スキャン | `go vuln` / `pnpm audit` / Dependabot の定期実行を CI または手動手順に組み込む候補 — 本 unit では network 制約により実施不能 |
| RC-4 | performance-tests.yml `github-script` comment step | 現行 trigger (schedule+dispatch のみ) では未到達の dead config — 削除または pull_request trigger 追加の候補 |
