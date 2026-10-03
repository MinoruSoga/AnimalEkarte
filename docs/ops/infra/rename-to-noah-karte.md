# AnimalEkarte → noah-karte リネーム計画

> 作成: 2026-10-03。対象: リポジトリ・Cloudflare・DB・CI/CD・外部連携・コード・ドキュメントの全命名。
> 目的: 製品名 `noah-karte.com` への統一。MECE 分類 A〜G + フェーズ実行順 + リスク台帳。

## 0. 前提: 命名規則の確定

現状は3系統の表記が混在している。新名は2形態に統一する。

| 現行 | 用途 | 新名 |
|---|---|---|
| `animalekarte-*` | CF リソース・リポジトリ | `noah-karte-*` |
| `animal-ekarte` | Go module・npm パッケージ | `noah-karte`（スラグ） |
| `AnimalEkarte` | リポジトリ名・docs・UI 表記 | `Noah Karte` / `noah-karte` |

- リソース名はドット不可のため `noah-karte` スラグを使用（例: `noah-karte-stg-api`）。
- ドメイン `noah-karte.com` / `api.noah-karte.com` は **prod の確定 URL** であり wrangler vars で既に採用済み。
  本計画で変わるのは裏側のリソース名のみで、prod の公開 URL は一切変更しない。
- 環境サフィックスは現行と同じ `-stg-` / `-prod-` を維持する。
- 【選択肢】STG にも `stg.noah-karte.com` / `api-stg.noah-karte.com` 系の custom domain を切る案。
  workers.dev サブドメインに依存しなくなり、将来リネームが再びあっても公開 URL が不変になる
  （webhook・frontend の WORKER_URL 設定が安定する）。実施するなら B1/G1 と同一作業に混ぜる。

### スコープ外（変更しない）

- **`baritechllc`（会社名）**: Plane workspace slug・Cloudflare アカウントの workers.dev
  サブドメイン（`*.baritech-soga.workers.dev`）等、組織名に属する命名は全て据え置き。
  本計画は電カル**プロジェクト**の命名のみを対象とする。
- GitHub アカウント名（`MinoruSoga`）等の個人/組織アカウント識別子。

## A. アイデンティティ層（表示名・再作成不要なもの）

| # | 対象 | 現状 | 作業 | コスト |
|---|---|---|---|---|
| A1 | GitHub リポジトリ `MinoruSoga/AnimalEkarte` | `gh repo rename` | 低 — 旧 URL は自動リダイレクト、保護・secrets・Actions 履歴は維持 |
| A2 | ローカルフォルダ `~/Dev/Case/AnimalHospital/AnimalEkarte` + `/tmp/ae-*` worktrees | `git worktree` 再 clone | 低 |
| A3 | Go module `github.com/animal-ekarte/backend`（**1671 ファイル**が import で参照） | 機械的置換 or **据え置き選択肢あり** | 高 churn — 外部から import されない内部 module のため据え置きでも機能上問題なし。やるなら専用 PR 1 本で完結 |
| A4 | npm 名 `animal-ekarte-frontend`、UI 内製品表記・ログイン画面・LINE/メール文面 | 置換 | 低〜中 |
| A5 | Plane project 表示名（workspace `baritechllc` は**社名で据え置き** / identifier `EMR` / UUID `6a009324-…`） | Plane UI で project 名のみ rename | 低 — UUID・identifier 不変なら API 連携無影響 |

## B. Cloudflare リソース（再作成型 = 破壊的変更の中核）

| # | 対象 | 現状 | 作業 | 注意 |
|---|---|---|---|---|
| B1 | Workers `animalekarte-stg-api` / `animalekarte-stg-frontend` / （未作成）prod 2 本 | 新規作成 + traffic 切替 | Worker 名は変更不可 = リネームではなく**新 Worker 作成**。workers.dev サブドメインが変わる |
| B2 | R2 `animalekarte-stg-images` / （未作成）`animalekarte-prod-images` | 新 bucket + オブジェクト移行 + S3 key 再発行 | bucket 名変更不可。バグ報告スクショ・医療画像のコピー期間中は dual-read または移行窓が必要 |
| B3 | DO class `AnimalEkarteApiContainer` + containers app + `animalekarte-migrate-runner-v1` | 新名で定義 | **DO storage は class_name に紐付く** — class rename は新 namespace 扱いで既存 instance データへ到達不可。API container が stateless であることを先に検証する |
| B4 | Worker secrets 一式 | `wrangler secret put` を新 Worker に全件再投入 | `MIGRATE_RUN_SECRET` / `JWT_SECRET` / `DB_*` / `PLANE_API_KEY` / `SCHEDULER_*` / `INTEGRATION_ENCRYPTION_KEY` / LINE・LSTEP 系。**投入漏れ = 2026-10-03 の障害と同型**。`infra/cloudflare/.env` を source of truth にして一括投入スクリプト化する |
| B5 | workers.dev サブドメイン・routes | 新 Worker 名で新 URL | prod の `api.noah-karte.com` / `noah-karte.com` custom domain は**新名で済んでおり変更不要**（workers.dev 参照は STG のみ） |
| B6 | Cloudflare Access ポリシー（scheduler ops の JWT audience） | Access アプリ側で audience 更新 or 新規 | `/_internal/*` 系保護に使っている Access 設定を新 URL に追随 |

## C. データストア

| # | 対象 | 作業 |
|---|---|---|
| C1 | PlanetScale database 名（現在 `postgres`）・branch 名 | **rename 対象確定** — プロジェクト命名の全リソースを `noah-karte` 系に統一。PlanetScale の rename 可否は仕様上 rename または新規+データ移行のどちらかで実施。**branch 名を変えると `user.branch` 形式のクレデンシャルが変わる**ため `STG_DB_USER`/`STG_DB_PASSWORD`（GitHub secret + 新 Worker への secret put）の同時更新が必須。org が会社単位の共有 org なら org 名はスコープ外（§0 スコープ外と同規則） |
| C2 | R2 オブジェクト（B2 と同一線） | `rclone`/`aws s3 sync --endpoint` で旧→新コピー → アプリの `S3_BUCKET` vars 切替 → 旧削除 |
| C3 | ローカル dev: docker compose project・DB 名・`.env*` の参照 | `COMPOSE_PROJECT_NAME`・接続文字列・`make local-db-*` の grep 置換 |

## D. CI/CD・GitHub 設定

| # | 対象 | 箇所 |
|---|---|---|
| D1 | workflows 内のハードコード URL/Worker 名 | `backend-deploy.yml` の `WORKER_URL`、`frontend-deploy.yml` STG URL、`stg-smoke.yml` default、`worker-secret-sync.yml` の `--name animalekarte-stg-api`、`stg-migrate-direct.yml` |
| D2 | GitHub secrets | 値自体は新 Worker へ移すだけ。`Production` Environment は**未構築**なので最初から新名 Worker を対象に作成する |
| D3 | branch protection・CODEOWNERS・Environments | repo rename では維持される。`CODEOWNERS` 内の repo 参照と Environment の deployment branch 設定を確認 |

## E. 外部連携

| # | 対象 | 作業 |
|---|---|---|
| E1 | Plane | workspace `baritechllc` は社名で据え置き。project 表示名のみ rename（A5）。`PLANE_*` env 値は不変 |
| E2 | LINE/LSTEP webhook callback URL | STG workers.dev URL を参照していれば新 URL に差替。prod は custom domain 前提なら無影響。**クリニックごとに DB 管理の token と webhook は両方確認** |
| E3 | Chromatic `CHROMATIC_PROJECT_TOKEN` | プロジェクト名のみ変更なら token 不変。リポジトリ rename に伴う再リンク要確認 |
| E4 | LLM プロバイダ（`SUPPORT_LLM_BASE_URL` 等） | 名前依存なし。確認のみ |
| E5 | 監視/アラート（存在すれば: Sentry, BetterStack, Cloudflare notifications 等） | Worker 名ベースのルールは張り替え |
| E6 | codex-security policy（`SECURITY.md` 生成物） | **rename 後に再生成**で新名を正確に記録。再生成までの間は F1/F3 の機械置換対象（タイトル等に `AnimalEkarte` を含む）として暫定追従する |

## F. コード内参照（機械的置換 — A〜E の付随作業）

| # | パターン | 規模 | 対象 |
|---|---|---|---|
| F1 | `animalekarte` | 約 200 ファイル | wrangler 設定・`infra/scripts/cf-run-migrate.sh`・Makefile・`.env.example`・workflow・`.claude/` 配下・tests |
| F2 | `animal-ekarte` | 約 1671 ファイル | **ほぼ Go import path**（A3 と同一作業。据え置きなら対象外） |
| F3 | `AnimalEkarte`（title case） | docs・`.claude/rules`・skills・agents 表記 | F1 と同一の置換対象（多くは同一ファイル） |
| F4 | `animal_ekarte` | 2 ファイル | 同上 |
| F5 | リポ外: `~/Vaults/CorpVault`・Bitwarden エントリ名・`infra/cloudflare/.env` のコメント | docs/ops 手順と整合させて手動更新 | — |

## G. カットオーバー手順（環境別）

### G1. STG（リハーサル — prod 命名手順の検証台）

1. 新名リソース作成: `noah-karte-stg-api` Worker、`noah-karte-stg-frontend`、R2 `noah-karte-stg-images`、新 DO class
2. secrets を `infra/cloudflare/.env` から一括 re-put（投入スクリプト + 投入後の `secret list` 照合）
3. データ移行: R2 オブジェクト sync、DO storage は stateless 確認後に切捨て、PlanetScale は rename or 新 branch + dump/restore
4. URL 切替: `S3_BUCKET`・`WORKER_URL`・webhook 類を新名へ。`/_internal/migrate` + `/health` + `/health/db` + smoke で検証
5. 旧リソースは**検証期間（目安 1 週間）経過後に退役** — 即削除しない

### G2. Prod（未構築 — 最初から新名で作成）

- `docs/ops/infra/production/setup.md` のチェックリスト実行時に `noah-karte-prod-*` で作成するだけ（rename 作業なし＝最安値）
- `animalekarte-prod-*` の設定記述（wrangler.production.jsonc）は G2 実施時に新名へ書換

### G3. 検証・ロールバック

- 検証: `/health` `/health/db` `/owners` SPA fallback・ログイン・migrate endpoint・Plane 起票・スクショ upload/download・LINE webhook
- ロールバック: 旧 Worker/bucket を残しておけば traffic 戻しで即復旧（G1-5 の退役期間が保険）

## 推奨実行順序

| Phase | 内容 | ブロッカー |
|---|---|---|
| 0 | 本計画の承認 + A3 module rename の可否決定 + `infra/cloudflare/.env` への secrets 網羅確認 | — |
| 1 | F1/F3/F4 の機械的置換（設定値を新名に先立って変数化）+ codex-security policy 再生成 | 0 |
| 2 | G1: STG 全面カットオーバー（B+C+D+E の STG 分） | 1 |
| 3 | G2: prod 構築を新名で実施 | go-live 日程 |
| 4 | A1 repo rename + A2 フォルダ + A4 表示名 + A5 Plane | 2（STG 安定確認後） |
| 5 | A3 Go module rename（実施する場合のみ・専用 PR） | 4 |
| 6 | 旧リソース退役・CorpVault/docs 最終更新 | 2 の検証期間後 |

## リスク台帳

| リスク | 影響 | 緩和 |
|---|---|---|
| secrets re-put 漏れ | 起動/migrate/LLM 全滅（10-03 障害と同型） | `.env` 駆動の一括投入 + `wrangler secret list` 照合手順化 |
| DO class rename で storage 孤立 | container instance データ到達不可 | 事前に stateless 検証。stateful なら移行手順を別途設計 |
| webhook URL 更新漏れ | LINE/予約連携の silent 断 | E2 を G1-4 の検証リストに含める |
| R2 移行中の画像不整合 | スクショ/医療画像の 404 | sync → 切替 → 検証 → 旧削除の順を厳守 |
| repo rename 後の stale remote | ローカル・CI・外部ツール | git リダイレクトで緩衝、`git remote set-url` 周知 |
| 1671 ファイル churn | review 不能な巨大差分 | A3 は分離・専用 PR。据え置き選択肢を Phase 0 で決定 |
| `main` は unprotected | rename 作業中の直接 push 混入 | 作業前に一時的に保護をかけるか、作業窓を決めて実施 |

## 未決事項（Phase 0 で回答必要）

1. Go module path を rename するか据え置くか（推奨: **据え置き** — 内部 module で外部解決不要）
2. DO `AnimalEkarteApiContainer` の stateful/stateless 確認
3. STG の新名確定: `noah-karte-stg-*` でよいか
4. STG 公開 URL を `stg.noah-karte.com` / `api-stg.noah-karte.com` の custom domain にするか
   （workers.dev のままか — §0 の選択肢参照）
5. PlanetScale: database 名の新名（例: `noah-karte`）・branch 命名・org が会社共有かどうか
