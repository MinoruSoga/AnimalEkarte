# Cloudflare 専用アカウント分離・noah-karte 改名計画

> EMR-298 / case BRT-4。2026-10-06 の静的棚卸しと計画。実環境の変更・改名実装は行っていない。
> 確定前提（保存された依頼文）: 専用アカウント化と全名称の noah-karte ファミリー化。移行先を新名で構築し、旧名を退役する。**共有アカウントへの本番構築は禁止**。
> 実行 SoT は Plane。後続 EMR-302/303/305/307 と本番チケットの割当は依頼文による。作成時点では Plane の読み取り API が使えず関係を UNKNOWN と記録したが、2026-10-06 に EMR-302/305 の blocked_by 解除・EMR-148/256/134/139 への禁止追記・EMR-302 の子課題分割を実施済み（後続実行者は最新状態を再照合する）。§4/§5 を後続手順の正とし、チケットに別仕様を複製しない。

## 1. 現状の棚卸し

### 確認範囲と証明の限界

ソースとローカル state の資源 type/name のみを確認した。state の属性、`.env*`、`sensitive-local/`、個人の資格ストア、患者・飼い主データは閲覧していない。任意の Wrangler 実環境確認は未実施。表中の資源名は設定名であり、存在・稼働・請求を証明しない。

| 層 | 確認した事実・対象 | 根拠 / 未確認 |
|---|---|---|
| STG | API と frontend の2 Worker。frontend の `API` binding は STG API を指す | `backend/wrangler.jsonc:24`、`frontend/wrangler.jsonc:12`、`frontend/wrangler.jsonc:35`。2026-10-03 live 記録は `SECURITY.md:91`。今回の live は UNKNOWN |
| 本番 | API と frontend の2 Worker の設定は draft | `backend/wrangler.production.jsonc:44`、`frontend/wrangler.production.jsonc:17`、`SECURITY.md:13`、`docs/ops/infra/production/setup.md:7`。本番未稼働を計画前提とするが、現在の live は UNKNOWN |
| R2 | STG/prod の画像 bucket、S3 と shared storage は環境内で同じ bucket | `backend/wrangler.jsonc:80`、`backend/wrangler.jsonc:82`、`backend/wrangler.production.jsonc:93`、`infra/cloudflare/r2.tf:22`、`infra/cloudflare/production/r2.tf:13` |
| DB | Container から PlanetScale へ `DB_*` 直結。DB は移管しない | `backend/wrangler.jsonc:4`、`backend/worker/index.ts:65`。接続先の値は記録しない |
| Hyperdrive | config 削除済み。state に `stg_planetscale` 残存 | `infra/cloudflare/hyperdrive.tf:1`、`infra/cloudflare/hyperdrive.tf:12`。remote 削除は UNKNOWN |
| Containers/DO | `AnimalEkarteApiContainer`、basic、max 3、APAC、default policy、SQLite class | `backend/wrangler.jsonc:131`、`backend/wrangler.jsonc:137`、`backend/wrangler.jsonc:138`、`backend/wrangler.jsonc:144`、`backend/wrangler.jsonc:154`、`backend/wrangler.jsonc:174` |
| DO named instance | scheduler と migrate-runner の2つ。通常 API 用 singleton も別に存在する設計 | `backend/worker/scheduled-jobs.ts:1`、`backend/worker/migrate-exec.ts:16`、`backend/worker/index.ts:316`。実際の instance 数は UNKNOWN |
| cron | STG 5式: `0 1 * * *` / `0 6,11 * * *` / `0 17 * * *` / `*/4 0-9 * * *` / `*/15 * * * *`。本番 draft は keep-alive を除く4式 | `backend/wrangler.jsonc:126`、`backend/wrangler.production.jsonc:126`。5式全てが keep-alive ではなく `*/4` のみ (`backend/worker/index.ts:470`) |
| ゾーン/DNS | `noah-karte.com`。STG/prod の route と DNS 定義がある | `infra/cloudflare/zone.tf:48`、`infra/cloudflare/production/zone.tf:18`。authoritative NS、registrar、DNSSEC、証明書、プランの現在値は UNKNOWN |
| Access | scheduler は team domain/audience を要求し署名を検証 | `backend/wrangler.jsonc:206`、`backend/wrangler.jsonc:207`、`backend/worker/scheduler-access-auth.ts:404`。既存 team/application/policy/service token の live 一覧は UNKNOWN |
| 通知 | edge-error policy 定義あり、state の15件には含まれない | `infra/cloudflare/notifications.tf:54`、`infra/cloudflare/notifications.tf:58`。state 不在は確認済み、手動作成・remote 未適用は UNKNOWN |
| D1/KV/Queues | 上記 Wrangler 設定と state type/name 一覧に定義なし | `backend/wrangler.jsonc:162`、`backend/wrangler.jsonc:178`。remote の不存在は UNKNOWN |

生成元の「prod-images は0件」「STG 約4.9万リクエスト/30日・Container 2稼働」「ゾーン active/Free/authoritative」は今回再計測していない。根拠は依頼文の dated observation に限り、現在値は UNKNOWN。本番データが見つかった場合は空 bucket 案を停止して再計画する。

### 実行前に固定する分類: Terraform state 15資源

検証コマンド（実行済み、exit 0、STG 11・本番4。属性は出力しない）:

```sh
python3 - <<'PY'
import json,pathlib
for p in ['infra/cloudflare/terraform.tfstate','infra/cloudflare/production/terraform.tfstate']:
 print(p)
 d=json.loads(pathlib.Path(p).read_text())
 for r in d.get('resources',[]): print(r['type'],r['name'])
PY
```

「移す」は DNS の論理設定を複製する意味で、ID/state の横流しではない。新アカウントでは新 ID を得て import または作成する。下表は計画案として1件ずつ固定し、G0 で承認する。旧 state のバックアップは資格を含む可能性があり保護する。

| 環境 | state の type.name | 分類 | 新側の扱い |
|---|---|---|---|
| STG | cloudflare_zone.noah_karte | 作り直す | 同じドメインを新アカウントへ追加、新 zone ID を STG state の単一管理に import |
| STG | cloudflare_dns_record.stg_frontend | 移す | 新 STG frontend route の受信確認後、同じ hostname に複製 |
| STG | cloudflare_dns_record.api_stg_backend | 移す | route 用の論理設定を維持。旧 AWS CNAME の placeholder は後続で整理 |
| STG | cloudflare_dns_record.acm_validation_stg | 捨てる | 廃止 AWS の検証用。現在の利用なしを人が再確認できるまで削除しない |
| STG | cloudflare_dns_record.domainconnect | 移す | registrar 再確認まで維持 |
| STG | cloudflare_dns_record.caa_amazon | 捨てる | 廃止 AWS 用。新証明書の発行要件に反しない確認が前提 |
| STG | cloudflare_dns_record.caa_pki_goog | 移す | 新 TLS 発行 CA と照合 |
| STG | cloudflare_dns_record.caa_sectigo | 移す | 同上 |
| STG | cloudflare_dns_record.caa_letsencrypt | 移す | 同上 |
| STG | cloudflare_hyperdrive_config.stg_planetscale | 捨てる | remote 残存を人が照会・必要なら承認後廃止。**新側 import/再作成前に旧 state から除外**。再作成しない |
| STG | cloudflare_r2_bucket.stg_images | 作り直す | 新名 bucket を作りデータだけコピー |
| 本番 | cloudflare_zone.noah_karte | 作り直す | 新 zone の参照へ再取得。type/name だけでは managed/data を判定しない。prod は data source とする |
| 本番 | cloudflare_dns_record.api_prod_backend | 移す | 新 prod API 殻の受信・本番承認後に有効化 |
| 本番 | cloudflare_dns_record.frontend_www | 移す | 新 prod frontend 殻の受信・本番承認後に有効化 |
| 本番 | cloudflare_r2_bucket.prod_images | 作り直す | 空を再確認し、空の新名 bucket のみ作成 |

DNS の用途根拠: `infra/cloudflare/zone.tf:98`、`infra/cloudflare/zone.tf:121`、`infra/cloudflare/zone.tf:133`、`infra/cloudflare/zone.tf:147`、`infra/cloudflare/zone.tf:161`、`infra/cloudflare/zone.tf:175`。AWS を rollback に使わない決定は `docs/ops/infra/architecture.md:69`。production は zone を二重管理しない (`infra/cloudflare/production/zone.tf:1`)。通知は15資源の外で、新側への適用を EMR-303 に分離する。

### 改名対応表（現行名→新名）

下表の新名は計画候補。CF は `noah-karte-*`、class は `NoahKarte*`、DB/docker は `noah_karte*`、表示は `Noah Karte`。repo/dir/module の3文字列は §5 決定結果（2026-10-06 確定）の太字の値。同一アカウント内の rename 可否に依存せず、今回は別アカウントへ新規構築する。

| 層 | 現行名 | 新名 / 方針 | 根拠 |
|---|---|---|---|
| Worker STG API | animalekarte-stg-api | noah-karte-stg-api | `backend/wrangler.jsonc:24` |
| Worker STG frontend | animalekarte-stg-frontend | noah-karte-stg-frontend | `frontend/wrangler.jsonc:12` |
| Worker prod API | animalekarte-prod-api | noah-karte-prod-api | `backend/wrangler.production.jsonc:44` |
| Worker prod frontend | animalekarte-prod-frontend | noah-karte-prod-frontend | `frontend/wrangler.production.jsonc:17` |
| R2 STG / prod | animalekarte-stg-images / animalekarte-prod-images | noah-karte-stg-images / noah-karte-prod-images | `backend/wrangler.jsonc:182`、`backend/wrangler.production.jsonc:164` |
| DO class / Container app class | AnimalEkarteApiContainer | NoahKarteApiContainer。app/image は新側で初回作成。生成 app 名は実行時 readback で固定 | `backend/worker/index.ts:47`、`backend/wrangler.jsonc:131` |
| DO binding | API_CONTAINER | API_CONTAINER 維持（製品名を含まない API 契約） | `backend/wrangler.jsonc:165` |
| DO scheduler instance | animalekarte-scheduler-v1 | noah-karte-scheduler-v1 | `backend/worker/scheduled-jobs.ts:1` |
| DO migrate instance | animalekarte-migrate-runner-v1 | noah-karte-migrate-runner-v1 | `backend/worker/migrate-exec.ts:16` |
| 通知 policy | animalekarte-stg-worker-edge-error-rate | noah-karte-stg-worker-edge-error-rate（prod を追加する場合も同規則） | `infra/cloudflare/notifications.tf:56` |
| workers.dev URL | animalekarte-{stg,prod}-{api,frontend}.baritech-soga.workers.dev（prod は候補で live UNKNOWN） | noah-karte-{stg,prod}-{api,frontend}.<NEW_SUBDOMAIN>.workers.dev。運用 URL は既存ゾーンを推奨 | `SECURITY.md:91`、下記公式 workers.dev 資料 |
| Go module | github.com/animal-ekarte/backend | **github.com/MinoruSoga/noah-karte/backend** | `backend/go.mod:1`。remote 読み取り結果は `git@github.com:MinoruSoga/AnimalEkarte.git` |
| GitHub repo slug | AnimalEkarte | **noah-karte**（既存 repo の in-place 改名） | `git remote get-url origin` の上記実出力 |
| ローカル dir | AnimalEkarte | **noah-karte**（親階層は維持） | 実行 cwd。worktree 参照も後続で更新 |
| docker project/container/network | dir 由来 animalekarte / animalekarte-db-1 / animalekarte_ekarte-network | project noah_karte / noah_karte-db-1 / noah_karte_noah_karte_network | `docker-compose.yml:219`。実際の既存 container 名は UNKNOWN |
| docker volumes / image | ekarte-postgres-data 等 / ekarte-frontend:latest | noah_karte_postgres_data / noah_karte_frontend_node_modules / noah_karte_go_mod_cache / noah_karte_go_build_cache / noah_karte_frontend:latest | `docker-compose.yml:93`、`docker-compose.yml:207`、`docker-compose.yml:209`、`docker-compose.yml:211`、`docker-compose.yml:213` |
| ローカル/テスト DB | animal_ekarte / animal_ekarte_test | noah_karte / noah_karte_test。ホスト済み PlanetScale DB の改名ではない | `.github/workflows/performance-tests.yml:25`、`backend/internal/reservation/reservation_created_by_fk_test.go:72` |
| macOS label/plist | com.animalekarte.lab-device-agent | com.noah-karte.lab-device-agent / 同名.plist | `packaging/macos/com.animalekarte.lab-device-agent.plist:6` |
| macOS install dir / build temp | Application Support/AnimalEkarte / animalekarte-lab-device-agent-* | Application Support/Noah Karte / noah-karte-lab-device-agent-* | `scripts/lab-device/install-lab-device-agent.sh:27`、`scripts/lab-device/build-lab-device-agent-bundle.sh:36` |
| npm package / docs/UI | animal-ekarte-frontend / AnimalEkarte 等 | noah-karte-frontend / Noah Karte（URL は noah-karte.com） | `frontend/package.json:2`、`SECURITY.md:1`、`docs/ops/infra/rename-to-noah-karte.md:38` |

DB/docker の underscore と container 自動 suffix の hyphen、macOS reverse-domain label は層の仕様による例外。会社名・Plane identifier `EMR` は製品名ではなく維持する。通常 API singleton は製品名を含まないため今回の2 named instance の改名とは区別する。

改名表記の再実測（編集前、git が列挙する tracked/untracked テキスト。禁止パスと state を除外、重複ファイル計上あり）: `AnimalEkarte` 112ファイル、`animalekarte` 144、`animal-ekarte` 1687（backend 1674）、`animal_ekarte` 6。以前の約1700という記載は概数で、backend 1674全てが import とは断定しない。正確なコマンド/出力は handoff。生成型、テスト、運用スクリプト、docs/UI、CI を含めた後続置換が必要。履歴記録は注記・対応表で追跡し、既存 Plane 履歴の一括置換は行わない。

## 2. アカウント固有の参照箇所

### 資格・設定の投入と失効（値は記録しない）

| 対象 | 参照根拠 | 新側投入→確認→旧側失効の順序 |
|---|---|---|
| account/zone | `infra/cloudflare/variables.tf:1`、`infra/cloudflare/zone.tf:50`、`infra/cloudflare/production/r2.tf:14` | 新 account `<ACCOUNT_ID>` を人が確認し、STG/prod 別 state と新 zone 参照を結線。単に account_id を置換した旧 state を適用しない |
| AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY | `backend/wrangler.jsonc:191`、`backend/wrangler.jsonc:192` | 新側 bucket 限定キーを環境別に発行・投入、copy/照合/切替/訓練/soak 後に当該旧キーのみ失効 |
| S3_ENDPOINT / S3_BUCKET / S3_SHARED_BUCKET | `backend/wrangler.jsonc:84`、`backend/wrangler.production.jsonc:98` | 新 account endpoint と新名 bucket を同時更新。旧 endpoint/キーは退役まで rollback 用に保管 |
| SCHEDULER_ACCESS_TEAM_DOMAIN / SCHEDULER_ACCESS_AUDIENCE | `backend/wrangler.jsonc:206`、`backend/worker/index.ts:596`、`backend/worker/index.ts:597` | 新 team/application/policy/audience、必要な service token を再作成→scheduler 更新→正規 operator 許可・誤 issuer/audience 拒否→soak 後に当該旧 app/資格を撤去 |
| CLOUDFLARE_API_TOKEN | `.github/workflows/backend-deploy.yml:78`、`.github/workflows/frontend-deploy.yml:75`、`.github/workflows/worker-secret-sync.yml:42` | STG 予行ではこれだけ更新。repo-level と staging Environment の shadow を値を表示せず確認し、実効 secret が新 account を指すことを照合。旧 token は退役まで保持 |
| PROD_CLOUDFLARE_API_TOKEN | `.github/workflows/backend-deploy.yml:175`、`.github/workflows/backend-deploy.yml:211`、`.github/workflows/frontend-deploy.yml:106` | STG 合格後、本番殻作成時に **Production** Environment に結線。STG 訓練の戻し対象には含めない。旧資格は利用先確認後の退役で失効 |
| PlanetScale DB_HOST / DB_USER / DB_PASSWORD（CI の STG_DB_USER / STG_DB_PASSWORD） | `backend/wrangler.jsonc:188`、`backend/worker/index.ts:66`、`.github/workflows/worker-secret-sync.yml:43` | 環境別の同じホスト済み DB 接続情報を保護された正本から新 Worker に引き継ぐ。DB 自体を移管・改名しない。Hyperdrive 専用資格が存在すれば人が用途確認後ローテーション |
| その他の必須・任意 Worker secret | `backend/wrangler.jsonc:195`、`backend/wrangler.jsonc:212`、`backend/wrangler.jsonc:234`、`backend/wrangler.production.jsonc:170` | MIGRATE_RUN_SECRET、SCHEDULER_OPS_SECRET、SCHEDULER_INTERNAL_TOKEN、SCHEDULER_ALERT_*、JWT_SECRET、INTEGRATION_ENCRYPTION_KEY、SMTP_*、必要な SUPPORT_LLM_API_KEY/PLANE_API_KEY/SEEDLOGIN_* を名前ベースで照合。環境間流用はしない。既存暗号化データ用 key を勝手に変更しない |

旧 Worker の secret **名前一覧と投入先**を退役前に人が照合し、DB 接続情報、不要な Hyperdrive origin 資格、任意 secret の残置がないことを確認する。値をログ・文書・Terraform state に複製しない。全アカウント共通 token なら他案件の用途を切り分けるまで失効禁止。新側投入、読み取り/疎通、切替、STG 訓練、soak、旧側失効・削除の順序を固定する。

### workers.dev 参照10ファイルの置換方針

推奨: 定常 API は `https://api.stg.noah-karte.com`、定常 frontend は `https://stg.noah-karte.com`。新 account の workers.dev は NS/TLS 切替前の一時予行用。新 subdomain は `<NEW_SUBDOMAIN>`（候補 noah-karte、取得可否 UNKNOWN）。新側 STG `workers_dev=true` を一時許可し、正規 hostname の疎通後は両方 false。本番は false を維持（`backend/wrangler.production.jsonc:58`、`frontend/wrangler.production.jsonc:23`）。無効化後に CI/スクリプトが一時 URL を使い続けないことを確認する。

| 参照ファイル | 置換先/扱い |
|---|---|
| `.github/workflows/backend-deploy.yml:34` | API 正規 hostname（一時予行時だけ新 workers.dev を明示） |
| `.github/workflows/frontend-deploy.yml:86` | frontend 正規 hostname |
| `.github/workflows/stg-smoke.yml:18` | API 正規 hostname |
| `infra/scripts/cf-run-migrate.sh:16` | API 正規 hostname。ただし移行中の migration dispatch は停止 |
| `infra/scripts/cf-crud-smoke.sh:23` | API 正規 hostname |
| `scripts/tools/manual-sync.py:55` | API 正規 hostname |
| `load-tests/k6-cf-stg-sustained.js:56` | API 正規 hostname |
| `SECURITY.md:91` | dated live 記録を保全し、新 hostname の検証日付き記録を追加。値の機械置換で履歴を改ざんしない |
| `docs/ops/infra/rename-to-noah-karte.md:27` | 旧方針撤回と新 account/正規 hostname の対応を追記する後続タスク |
| `docs/work/linmig-campaign-20260919/LINMIG-232.md:23` | 履歴 URL は保全し、現行手順への参照注記を後続タスクで追加 |

再調査で11件目 `docs/work/infra-cf-rename-task-coverage-review-20261006.md:30` に歴史的記載が見つかった。元の10件とは分けて保全し、新規の実行参照漏れだけを退役ゲートで確認する。secret-sync の Worker 名も更新対象 (`.github/workflows/worker-secret-sync.yml:53`)。LINE/LIFF コンソール、CORS、端末 allowed origin、外部 webhook は repo 外もあるため一覧・担当は UNKNOWN を埋めてから切替する。

## 3. 移行方式の選択肢

| 方式 | 評価（計画案） |
|---|---|
| A: 新 account に新名で再作成、R2 データ複製、ゾーン設定再構築 | **推奨**。本番未稼働の前提に合う。旧側を期限付きで残して STG 訓練し、依存を照合して退役 |
| B: 共有 account の権限だけ絞る | 分離・改名の確定方針を満たさないため採用しない。実行条件未決なら現状の STG を維持して実行を待つ |
| C: 旧 state を丸ごとコピーして account_id だけ変更 | ID、資格残存、同じ zone の二重管理を招くため採用しない。新 state に実在する新資源を import/作成する |

公式確認（2026-10-06 閲覧。以下は方式の根拠で、実行許可ではない）:

- ゾーンは新 account に同じドメインを追加し DNS を export/import、registrar で NS を変更する。DNSSEC/add-on と証明書の再発行を検討する。Pending 中は proxy 不可。旧 zone は Moved Away 後に自動削除へ進むため、旧 zone を永続 rollback と見なさない。[ゾーン移行](https://developers.cloudflare.com/fundamentals/manage-domains/move-domain/)
- API token は account/zone を選択可能。現在は個別 Worker に role を限定できるが、未作成 Worker には指定できず初回作成は product-level Admin が必要。初回用の期限付き資格と、作成後の環境別最小権限 CI を分離する。Route の Zone 権限と Containers/R2/DO の必要権限は別途検証し、不必要な KV/Hyperdrive/WAF 権限を旧設定からコピーしない。[token の作成](https://developers.cloudflare.com/fundamentals/api/get-started/create-token/)、[Workers 権限](https://developers.cloudflare.com/workers/authorization/workers/)
- R2 の endpoint は account 固有で、object 権限は bucket に絞れる。新 bucket の作成と S3-compatible copy を採る。ネイティブな account 間 bucket 移管の有無は UNKNOWN。[R2 認証](https://developers.cloudflare.com/r2/api/tokens/)、[Rclone](https://developers.cloudflare.com/r2/examples/rclone/)
- Containers は Dockerfile からビルドし、新 account の Cloudflare registry に格納して app を初回作成する。既存 app/registry の account 間移管 API は UNKNOWN のため用いない。[image 管理](https://developers.cloudflare.com/containers/guides/image-management/)、[初回構築](https://developers.cloudflare.com/containers/get-started/)
- Hyperdrive は account 配下の構成。本件は未使用のため移さず、新側に作らない。account 間移管は UNKNOWN。[Hyperdrive API](https://developers.cloudflare.com/api/resources/hyperdrive/)
- `renamed_classes` は同じ Worker 内で class storage を引き継ぐ仕組みで、account 間コピーでも named instance 名の変更でもない。新 account の DO は空。後続で既存 `v1` 履歴から `v2` の `AnimalEkarteApiContainer`→`NoahKarteApiContainer` を設計し、最初の新側リリースで class/bindings/2 named instance 名を一体変更する。旧名で一度 deploy してから改名する二度手順は禁止。現行 legacy migrations を保ち、exports 方式への変更を本件に混ぜない。[legacy DO migrations](https://developers.cloudflare.com/durable-objects/reference/durable-object-class-migrations-legacy/)
- Access の team domain は account 固有。新しい専用 team を採り、application/policy/audience を再作成する。旧共有 team の名前を移すと他案件に影響するため移さない。[Zero Trust FAQ](https://developers.cloudflare.com/cloudflare-one/faq/getting-started-faq/)
- service binding の相手 Worker は同じ account に必要。API/frontend は環境ごとに同一の承認済みリリース単位で構築する（2つの deploy の実行は API→frontend の順でよく、原子操作を意味しない）。[service binding](https://developers.cloudflare.com/workers/runtime-apis/bindings/service-bindings/)
- workers.dev は account subdomain と Worker 名から URL が決まる。旧会社 subdomain の解放は範囲外。[workers.dev](https://developers.cloudflare.com/workers/configuration/routing/workers-dev/)

所有者が自社の場合: 自社の共有会社メール・請求管理下で当該製品専用 account を作り、顧客の閲覧/引継ぎ権限と終了時の移管責務を契約で決める。顧客所有の場合: 顧客の共有メールと請求管理下に account を置き、自社を役割限定の管理者として招待する。退職・契約終了時の自社メンバー/CI資格撤去と顧客への引継ぎが必要。ログイン共有は避け、復旧担当と承認者を人が固定する。同じログインで別 account 作成できるかは今回 UNKNOWN。

## 4. 推奨する移行手順（案）

### 番号付きの順序制約・承認ゲート

以下は **後続 EMR-302 の計画**。今回実行しない。移行前後で環境ごとの PlanetScale は同じ DB であるため、cron/DO/migrate の二重実行は安全でなく、Containers 費用も二重化する。

1. **G0 人間承認**: §5 の所有者、費用主体、メール、プラン費用、時期、3文字列、DO 空起動、LINE/LIFF 担当を記録する。通知先は所有者メールと同一で決定済み（§5 決定結果）。state 分類と URL 方針を固定し、Plane の現行条件を読み直す。共有 account への本番構築は禁止。移行中の自動 deploy、手動 scheduler、migration/direct migration を凍結する。切り戻し: 未着手に戻す。停止時間: 準備はなし。
2. **旧側の前提確認**: 人が資源・secret 名前・bucket 件数・DB/Access・authoritative NS/registrar/DNSSEC/TLS を再照会する。削除済み Hyperdrive は remote の残存・用途を確認し、承認後廃止と旧 state 除外を行う。**除外完了までは新 account への Terraform import/再作成をしない**。旧 state/バックアップの資格残存を保護・整理する。切り戻し: 保護した旧構成を維持、Hyperdrive は未使用なので再導入しない。停止時間: なし（利用発見時は停止・再計画）。
3. **新 account 準備・G1 人間承認**: account/メンバー/有料プラン、zone と DNS 論理設定、別 state、STG bucket、専用 Access team/application/policy/audience、R2資格を用意する。ゾーンの authoritative 化・NS 変更はまだ承認窓まで実施しない。新 state には新 ID だけを import/作成し、prod zone は参照だけにする。切り戻し: 旧側維持、新側を休止。停止時間: なし。
4. **STG の2 Worker 初回構築**: STG 予行で更新する CI は `CLOUDFLARE_API_TOKEN` のみ。api と frontend を同一 STG リリースにまとめ、新名と binding を整合させる。旧名での初回 deploy は行わず、DO class の renamed_classes 設計・named instance 改名を初回リリースに含める。ただし renamed_classes は同一 Worker 内の storage 継承の仕組みであり、storage を引き継がない新 account の新 Worker では新 class 宣言だけで足りる可能性がある。段階4の着手前に公式の挙動（新 Worker で v1 旧 class→v2 rename を同時適用する際に旧 class の export を要求されるか等）を確認し、不要なら新規宣言に簡略化する。Dockerfile を再ビルドし registry/image/container app を新 account に作る。`DB_*` は同じ STG PlanetScale に引き継ぐ。新 cron は **全停止**、新 scheduler/DO の自動/手動実行を禁止し、migrate-runner は停止確認後の排他テスト以外で呼ばない。切り戻し: CI を保管した旧 STG 値/設定へ戻し新側休止。停止時間: 準備はなし。
5. **R2 コピーと予行**: 旧 `animalekarte-stg-images`→新 `noah-karte-stg-images` を copy、object key 集合・件数/bytes と可能ならチェックサム・metadata を照合する。キー名・内容は保護証跡だけに保存する。書込みを止めて最終差分を copy/再照合してから新 bucket へ切り替える。ETag が checksum と等価かは UNKNOWN（multipart 等を一律同一視しない）。旧キーは失効しない。新 S3_ENDPOINT、AWS_*、bucket/binding を整合させる。DO の空状態・再入安全性の試験は旧実行停止→新だけ試験→新停止→必要なら旧再開の排他順序で行う。切り戻し: 新への書込みを止め差分を保全・旧 bucket に照合して戻す。停止時間: copy は並行、最終差分の書込み停止時間は容量・速度 UNKNOWN のため予行で実測。
6. **G2 STG 切替承認**: 新 Worker の同一 hostname の route/権限/証明書を事前検証する。Pending zone で proxy 動作を証明できないため、新 workers.dev/隔離予行と設定照合、NS 後の受信確認・戻し条件を分ける。同一 hostname を受けられることの確認計画が成立しない場合は DNS を変更せず停止する。切替時は旧 cron/DO/手動 scheduler/migrate 実行を止め、進行中処理・lease が残らない確認後に新側だけ武装する。新旧同時武装の窓を作らない。DNS/NS、LINE webhook/LIFF endpoint、CORS と端末 allowed origin の更新を同じ窓で行う。旧 **資源削除**はまだしない。切り戻し: 新実行停止・drain→STG hostname/LINE/LIFF を旧側へ→旧 STG cron のみ再武装。停止時間: NS/TLS の伝播と処理 drain 次第で UNKNOWN、ゼロを約束しない。
7. **STG 切り戻し訓練1回→再切替→短い soak**: 旧 R2 キーと旧 Worker を保持したまま上記 STG の戻しを実演し、R2 の新規書込み差分を欠落なく戻せることを確認する。必要な STG CI の復元後、新側へ再切替して新 cron のみ武装する。戻すのは STG hostname、LINE/LIFF、旧 STG cron とその実行に必要な STG 設定だけ。後段の `PROD_CLOUDFLARE_API_TOKEN` と prod Worker 殻は対象外。「hostname を旧側へ戻す」は zone の NS/移管を旧 account へ戻すことを意味する（別 account の workers.dev への proxied CNAME は 1014 エラーになりえるため、実質 NS 戻し）。soak は案として24時間、少なくとも5式それぞれの実行周期を含める。切り戻し: 同じ排他手順。停止時間: 2回の窓を予行で計測。旧 zone の自動削除期限より前に訓練する。
8. **G3 本番殻作成承認（EMR-302）**: STG 合格・訓練・単一武装後に `PROD_CLOUDFLARE_API_TOKEN` を Production Environment に結線。prod の api/frontend 2つを別の同一リリース単位で作り、cron は **停止したまま**。空を再確認した新 `noah-karte-prod-images` を作る。データコピーはしない。prod DB/資格の確認なしに Container 起動・migration は行わない。切り戻し: 新本番殻を休止し本番公開を延期する。旧共有 account へ本番を作る戻しは禁止。停止時間: 本番未稼働前提では利用者影響なし、違えば再計画。
9. **本番初回構築・G4（別チケット）**: EMR-302 の成果物境界は「空 prod-images、cron 停止の prod Worker 殻、本番トークン結線」まで。PROD-SETUP=EMR-148、10/20 切替=EMR-256、GOLIVE=EMR-134、初回管理者付与=EMR-139 は、その殻の上の初回本番構築を持つ。DNS/route の有効化、DB/migration・backup/restore、初回管理者、本番公開と **本番 cron 武装は EMR-256** 側の承認へ割り当てる（読み替えは人間決定）。同じ hostname の受信確認後、DNS・LINE/LIFF・CORS の変更を同窓で行う。本番ゲート不合格の切り戻しは未公開のまま延期。停止時間: 初回構築の所要を測定し、人が10/20窓を決定する。
10. **共有 account 側の退役・G5 人間承認（独立段階）**: stg-images 照合、STG 訓練1回、soak 合格、保護証跡と残存依存ゼロを確認した後に限り、対象旧 API token/R2キーを失効し、旧 Worker×4（実在確認できたもの）、R2×2、DO named instance 2つ/関連 namespace、停止済み cron 5式、当該 Access app/資格、Terraform の当該 state 資源を撤去する。旧 DB secret 等の残置ゼロを確認する。通常 API singleton/container と registry image も当該 app の実測依存だけ整理する。他案件の共有資格を失効しない。account 閉鎖、会社 workers.dev subdomain 解放、共有 Access team 削除は禁止。切り戻し: **失効/削除後は訓練済みの即時 rollback 不可**。新側の復旧/保護バックアップからの再構築が必要。停止時間: 新側安定なら予定なし、復旧時間 UNKNOWN。

### 確認項目（後続実行時の証跡を要求。合格ゲート別）

EMR-302 の「**予行合格**」は下表の **G2 列の記載のみ**を指す。G3（本番殻）・G4/G5（本番構築・退役）の記載は各段階の合格条件であり、予行合格を待たず後段のゲートで要求する — 全項目を予行合格に含めると段階8以降を予行が待つ循環になる。

| 面 | G2（STG 予行合格） | G3（本番殻） | G4/G5（本番・退役） |
|---|---|---|---|
| Worker | STG 2つの実受信・deploy account/version（段階4〜7）。4つを STG token の予行で作らない | prod 2つの殻/停止確認（段階8） | — |
| binding | 環境内 frontend→API、同一 account、`/api/*` の実疎通。別 account の旧 API を参照しない | — | — |
| R2 | copy 後 stg-images の集合/件数またはチェックサム照合、read/write | prod-images の空確認 | — |
| DO instance 2つ | 新 class/新名・空状態、scheduler pause/lease/fence/run ledger の空起動と業務側再入安全性、migrate-runner の排他。旧 state を継承したと解釈しない | — | — |
| cron | STG 5式の登録・時間・単一実行・失敗通知・keep-alive の停止/再武装 | prod draft 4式が殻で全停止 | 本番 cron 武装は EMR-256（G4） |
| Access | 新 team/application/audience、正規 operator、誤 issuer/audience/期限の拒否、旧資格の効力確認 | — | — |
| LINE/LIFF・CORS | console の before/after 照合、受信・返信、LIFF endpoint、許可 origin。実データを証跡に貼らない | — | — |
| authoritative zone | registrar の NS と authoritative 応答、新 zone Active、TLS chain、CAA/DNSSEC、STG hostname の受信 | — | — |
| CI | STG token の実効投入先 account、Production token 未変更 | 本番 token が新 account を指す確認。Environment 保護の live は別証跡 | — |
| rollback/soak | STG 訓練1回、逆差分 R2 照合、新旧実行の同時武装なし | — | soak と退役許可の記録 |

Container 起動自体が `/app/migrate` を実行する (`backend/Dockerfile.production:78`)。cron を止めるだけでは migration を止められない。段階4〜5の稼働確認前に旧 migration を停止・完了確認し、人が新側の起動/migration を1回だけ承認する。既存 CI も deploy 後 migration を呼ぶ (`.github/workflows/backend-deploy.yml:119`) ため、排他設計を後続実装の前提とする。

### Cloudflare 外の改名は後続実装タスク

EMR-305（Go module 単位②は子 EMR-307）で行う。①CF runtime 名は EMR-302 初回 deploy に同梱。②Go import・module と repo の in-place 改名は同じ専用窓、④packaging/導入済み端末、⑤docs/UI/npm/生成型はそれぞれ独立窓、**③dir/compose/local test DB は改名単位の最後**にする。repo/dir/Go module の改名を DNS 切替窓に重ねず、②と③も分ける。**②（EMR-307）は STG 予行（段階4〜5）の開始より前に完了するか、GOLIVE 後に行う** — 予行合格後〜本番構築前の窓で②をマージすると、新 account の STG で soak していないコードが本番リリースに載る。EMR-302 完了や EMR-303 開始を②〜⑤でブロックしない。

local DB の値は今回未閲覧。dir/project 変更だけでは固定 volume が再利用されるため、DB 再作成が必須とは断定しない (`docker-compose.yml:207`)。後続では project/volume の実効名・backup を確認し、保全した dump/restore または意図した再利用を選び、旧 volume を勝手に削除しない。テスト DB 名は hosted PlanetScale の改名と混同しない。

lab device agent は label/plist/install path と build bundle を更新し、導入済み各端末で旧サービスを unload（bootout）→保護した設定/port/consumer token と origin を引継ぎ→新 label を再登録（bootstrap）→起動/診療機器受信を確認する。旧・新を同時起動しない。失敗時は新を unload、旧 label/設定を戻す。既存 installer は旧 label を対象に bootout/bootstrap するため、新 installer だけでは旧 label を撤去できない (`packaging/macos/install-lab-device-agent.sh:11`、`packaging/macos/install-lab-device-agent.sh:117`、`packaging/macos/install-lab-device-agent.sh:118`)。端末一覧・実施者・停止時間は UNKNOWN、後続人間作業。

## 5. 人間の決定事項

専用アカウント化・全名称改名は確定前提であり、再決定しない。下表は推奨案と決定根拠。**全項目 2026-10-06 に人間が決定済み**（承認記録は下表の後）。§4 の着手に必要な決定を G0 で記録する。

| 決定 | 選択肢 | 推奨 / 承認位置 |
|---|---|---|
| 所有者 | 自社専用 account / 顧客所有 account に自社招待 | 長期の顧客データ・終了時引継ぎを優先して顧客所有。自社所有なら契約で引継ぎ責務を固定。G0 |
| 費用主体 | 顧客直接払い / 自社払い・契約精算 | 所有者と一致する顧客直接払い。代払なら上限・精算を明文化。G0 |
| 所有者メール | 所有会社の共有メール / 個人メール | 会社共有メール、復旧担当と MFA を個人別に管理。実アドレスは人が保護ストアで指定。G0 |
| 有料プランの費用承認 | Workers Paid と必要な従量分を承認 / 延期 | 公式見積と上限・通知先・承認者を記録して承認。zone プランと Workers Paid を区別。未承認なら延期。G0/G1 |
| 移行の実施時期 | STG 予行→soak→本番殻の窓を確保 / 延期 | STG 予行と短い soak を先行、NS/TLS 戻し時間込みで窓を確保。G0/G2 |
| 改名実施者（repo/dir/module） | 人間 / 承認済み後続 EMR-305/307 担当 | repo 権限操作は人間、module/dir は後続担当＋所有者レビュー。切替窓と分離。G0 |
| DO named instance 空起動の受け入れ | 再入安全性確認後に空起動を承認 / 保留して復旧設計 | scheduler/migrate-runner の二重実行・履歴欠落を評価してから承認。`renamed_classes` に account 間 storage 継承を期待しない。G0/G2 |
| 10/20 との関係 | 10/20 を新 account 初回本番構築とする / 条件が STG 予行に間に合わず日程変更 | 所有者・費用主体・有料プランと STG 合格が揃う場合だけ10/20採用。揃わなければ延期。**共有 account への本番構築は選択肢にせず禁止**。本番チケットの読み替えを人が承認。G0/G4 |
| LINE/LIFF console 更新担当 | 顧客管理者 / 自社の委任管理者 | 権限を持つ人間1名と代行者を固定し同窓で更新、before/after を保護記録。担当名は UNKNOWN。G0/G2 |
| Go module path（1文字列） | `github.com/MinoruSoga/noah-karte/backend` / 人が別の有効 owner path を指定 | 表の1文字列を推奨。remote owner と module owner を整合し、人間が承認。別案なら実行前にこの表を1文字列へ更新。G0 |
| GitHub repo slug（1文字列） | `noah-karte` / 人が別 slug を指定 | `noah-karte` を既存 repo の in-place 改名で採用。新規 repo は secret/CI 履歴継続を失うため避ける（旧文書の根拠 `docs/ops/infra/rename-to-noah-karte.md:35`）。G0 |
| ローカル dir 名（1文字列） | `noah-karte` / 人が別 dir 名を指定 | `noah-karte`。worktree/launcher の参照を後続で検証し、repo/module と実施窓を分ける。G0 |
| notification_email | 所有者メールと同一 / 別の共有運用受信先 | 所有会社の共有運用受信先を推奨。同一でもよいが送信専用 SMTP_FROM を転用しない (`infra/cloudflare/variables.tf:28`)。EMR-303 の適用前 |
| 定常 URL / 一時 subdomain | 既存 zone の hostname / 新 workers.dev を継続 | 既存 zone を定常、新 workers.dev は予行のみ。新 subdomain 取得可否と flag 無効化を人が確認。G0/G2 |

### 決定結果（2026-10-06 人間承認）

| 決定 | 確定値 |
|---|---|
| 所有者 | 自社専用 account。終了時引継ぎ責務は契約で固定する（顧客所有案の保護を代替手段として維持） |
| 費用主体 | STG=自社、PROD=顧客。CF 請求は account 単位のため按分は契約・精算で行い、根拠データは EMR-303 の定期使用量レポートを使う |
| 所有者メール | 人が指定したメール（値は EMR-298 の決定コメントと保護ストア。文書には書かない）。復旧担当と MFA は個人別に管理する |
| 有料プラン | Workers Paid 承認（上限・通知先・承認者を G1 で記録） |
| 実施時期 | 直近開始（〜10/中旬）。STG 予行→切り戻し訓練→soak→本番殻の窓を確保 |
| 改名実施者 | repo の in-place 改名は人間。module/dir は EMR-305/307 担当＋所有者レビュー |
| DO 空起動 | 受け入れ。scheduler/migrate-runner の再入安全性確認を実施条件として残す |
| 10/20 | 新 account 初回本番構築。前提（所有者・費用主体・有料プラン・STG 合格）が間に合わなければ日程を動かす。共有 account 構築は禁止 |
| LINE/LIFF 担当 | 人間（Minoru）が console 更新を実施。before/after を保護記録。代行者は未指名（必要なら別途決定） |
| 改名3文字列 | module=`github.com/MinoruSoga/noah-karte/backend`、repo slug=`noah-karte`、local dir=`noah-karte` |
| notification_email | 所有者メールと同一（値は同上） |
| 定常 URL | 既存 zone の hostname を定常に。新 workers.dev は予行確認のみ |
| 終了時引継ぎ方式 | **未決（契約項目・G4 の前提に追加）**: 自社所有＋PROD 顧客負担の構成では、契約終了時に PROD を顧客へ渡すには「account ごと譲渡（自社 STG・製品ドメインも渡る）」か「今回と同様の再移行作業」のどちらかになる。G4 本番構築の前提として契約への記載が必要 |

## 6. リスクと未確認

| リスク / UNKNOWN | 扱い・再開条件 |
|---|---|
| 古い NS コメントと authoritative 稼働という依頼文の不一致 | `backend/wrangler.jsonc:10`、レビュー `docs/work/infra-cf-rename-task-coverage-review-20261006.md:28` は live 検証にならない。registrar/NS/DNSSEC/TLS を人が日付付き再照会するまで移管しない |
| 旧改名文書との矛盾 | `docs/ops/infra/rename-to-noah-karte.md:27` の組織命名据え置き、`:37` の module 据え置き案は今回の確定方針で撤回。`:56` の hosted DB 改名案は本件では採用せず PlanetScale を移管しない。`:47` の class rename=状態喪失という一般化は公式 rename migration と区別。旧文書は今回編集せず後続で改訂 |
| state の資格・旧 Hyperdrive | type/name 以外未閲覧。旧 state の安全な保管/廃棄と残存資格の用途調査は人間。設定削除だけで remote 廃止済みとしない |
| DO 空状態の ledger/lease/fence/pause 喪失 | `backend/worker/index.ts:269`。新 account に跨る lease は無い。停止/drain/単一武装と空起動の再入試験が必要。保存状態のコピー・引継ぎ期待は範囲外 |
| R2 copy 中と rollback 中の書込み | STG 書込み停止/差分 copy/照合と逆差分の保全を訓練。件数一致だけで内容完全一致と断定しない。容量・copy 速度 UNKNOWN |
| 新 account の Containers 能力/料金/権限 | APAC/default/max 3 と Dockerfile の新側初回作成を予行で確認。旧 account の能力制約を新側で実証するまで同一と断定しない |
| 本番が既に稼働、prod bucket 非空 | 現在 UNKNOWN。見つかったら空資源前提を撤回して実行停止、別のデータ/停止計画を承認する。本番データコピーは今回の範囲外 |
| zone Pending・TLS・旧 zone 自動削除 | 新 hostname の受信を DNS 前後に分け、NS/TLS の fail 条件と rollback 期限を実行者が固定。既存 zone を indefinitely 保持できるとはしない |
| CI・外部連携・端末の live 設定 | Environment/reviewer、token 実効 account、LINE/LIFF、CORS、導入端末は UNKNOWN。設定の存在だけでは合格にしない |
| Plane 最新状態・関係 | 本書作成時は読み取り API なしで UNKNOWN と記録。その後 2026-10-06 に EMR-302/305 の blocked_by 解除、EMR-148/256/134/139 への禁止・着手条件追記、EMR-302 の子課題分割（STG 予行/本番殻/退役）を実施。後続実行者は最新状態を再照合する |

移行の範囲外: account 閉鎖、共有 workers.dev subdomain 解放、共有 Zero Trust team 移管/削除、旧 Plane 履歴の一括置換、DO storage の account 間継承、本番データコピー、Cloudflare 外の改名実装。通知 policy の未適用は移行のブロッカーにしない。

## 7. 費用と通知の方針

費用主体・支払手段・上限は §5 の人間決定（確定: STG=自社、PROD=顧客、按分は契約・精算側）。実請求額と現在の使用量は UNKNOWN。専用 account は請求・管理を分離するが、STG/prod が同居するその account 内の請求は環境別に自動分離されない — Workers Paid 基本料・zone・registry 等の共通費は環境別に測れない。按分ルールの案: **共通費は固定比率、従量は Worker 名別の Analytics 比率で配分、請求の正は CF 請求書**。これを契約側で固定し、EMR-303 の定期集計は STG/PROD 別に出せることを受け入れ条件に含める。

Containers は Workers Paid の対象で、稼働時間に応じた memory/disk、active CPU、network、Workers/DO/logs 等を確認する。公式には Workers Paid の基本月額 USD 5 が示されるが、総費用・為替・税・支払は未算定。zone Free と有料 Workers は別契約である。[Containers 料金](https://developers.cloudflare.com/containers/platform/pricing/)

現行は basic/max 3 と sleepAfter 1h、営業時間 keep-alive (`backend/wrangler.jsonc:137`、`backend/worker/index.ts:54`、`backend/worker/index.ts:470`)。稼働時間が費用に影響するという推論は公式料金に基づく。旧新の同時 cron/DO 武装を禁止し、旧実行停止と資源削除を分ける。新旧検証 Container の短期稼働や R2 二重保持/copy/照合の操作費は別に上限承認する。

`billing_usage_alert` の Workers/Containers 対象外・account 全体 USD 支出アラートなしという旧調査は `infra/cloudflare/notifications.tf:30` に記録されている。今回の[公式 alerts](https://developers.cloudflare.com/notifications/notification-available/)では製品別閾値と Argo/Load Balancing 例は確認できたが、Workers/Containers 対象外・支出通知の不存在を網羅的に再確認できず **現在の可否は UNKNOWN**。新 account の利用可能 alert types を人が確認するまで、この通知で予算監視できると扱わない。

代替案は EMR-303 の GraphQL Analytics による定期確認。開始条件は「新 account STG 稼働＋新 token で照会可」で、EMR-302 退役完了や EMR-305/307 を待たない。初期は毎日、安定後は週次案とし、Worker 別リクエスト、Container 稼働/CPU/memory/disk、R2量/操作、DO/logs を取得可能な指標だけ記録する。GraphQL の正確な dataset/field・retention・token 権限は UNKNOWN、後続が schema 確認と請求画面との照合を行う。公式にも Analytics は請求用の使用量測定として使わないと明記され、集計値を実請求額と同一視しない。閾値超過時の担当通知・承認済み停止手段を runbook 化する。[GraphQL Analytics](https://developers.cloudflare.com/analytics/graphql-api/)

edge-error policy は使用量/金額通知の代替にならず、zone 全体指標である (`infra/cloudflare/notifications.tf:18`)。notification_email の所属・検証を人が確定後に EMR-303 が適用し、実受信を別途確認する。値未決・policy 未適用を移行完了や10/20のブロッカーにしない。
