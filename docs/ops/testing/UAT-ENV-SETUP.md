# 受入テスト環境セットアップ (UAT Environment Setup)

> **目的**: selected profile/browser について、stack、account、fixture、isolation を fail-closed で確認する。
> **最新更新**: 2026-09-06

## 1. profile と browser を先に選ぶ

- `profile=local`: disposable local DB。CSV seed は `002_master`、別 phase の合成 login seed は下記 §3.2。
- `profile=stg`: 承認済み dedicated UAT tenant。shared clinic や production は禁止。
- `browser=cdp`: Chrome remote debugging `:9222` と project Chrome DevTools MCP。
- `browser=playwright`: supported Docker runner `frontend/scripts/run-e2e.sh`。host `frontend/node_modules/playwright` は prerequisite ではない。

選択した一組について全項目が確認できない場合は **BLOCKED**。別 browser の INFO や frontend/backend HTTP 200 だけで ready としない。

> `scripts/check-uat-env.sh` はJSON encoding、stdin送信、mode 0600の一時ファイル、container health、browser routeを安全に確認する。ただしprofile/fixture/mutation/cleanup receiptをfail-closedに検証しないため、出力は **advisoryのみ**であり正式readiness gateではない。

## 2. 共通 readiness checklist

| 項目 | 合格条件 |
|:--|:--|
| target boundary | local disposable DB または承認済み専用 UAT tenant。対象 clinic ID と fixture owner を report に記録 |
| stack | frontend `:3003`、backend `:8080/health`、DB/container health が OK |
| migration | backend が migrate 後に healthy。migration 変更を pull した場合はユーザー実行の `make migrate` も完了 |
| seed | CSV `002_master` と別 phase `003_login` を区別する。臨床 row の存在をこれら seed の効果とみなさない |
| account | 明示 provisioning 済みの UAT identity が login 可能で、必要 permission group/capability を持つ |
| fixture sentinel | selected scenario が要求する clinic/owner/pet/staff/master が approved manifest/receipt と一致 |
| mutation | pre/post count、許可 clinic、deterministic teardown または idempotent fixture が定義済み |
| LIFF | local は effective BE/FE mock。STG は approved deploy/settings lane で mock 無効を人間が確認 |
| browser | selected `cdp` または `playwright` が実際に利用可能 |

## 3. local profile

### 3.1 stack startup（ユーザー操作）

エージェントは startup/reset/migration を自動実行しない。canonical startup は次だけである。

```bash
make up
```

`make up` は `.env.local`、orphan cleanup、`db backend frontend` の wait を含む。bare `docker compose up -d` は等価ではない。raw equivalent が必要な場合は、ユーザーが次を実行する。

```bash
docker compose --env-file .env.local up -d --wait --wait-timeout 1200 db backend frontend
```

Backend startup 自体も migrate を先に実行する。ただし migration を変更する commit を pull した後は project policy により、ユーザーが `make migrate` を実行する。これは demo/clinical fixture を作る手順ではない。

### 3.2 fixture と account

- CSV loader が読むのは `002_master` の医院骨格・参照 master だけ。臨床 demo は含まない。
- migrate は続いて `backend/cmd/migrate/login_seed.go` から `seedlogin` を実行する。`APP_ENV=development/local/dev/test/staging` では catalog の合成 account/staff/医院所属と「一般」権限を upsert し、`003_login` の履歴を記録する。production・空・未知の環境では skip する。これは CSV bundle でも、任意の UAT 操作が可能な専用 account でもない。
- login seed は再実行時も upsert する。catalog account の変更を恒久的な provisioning とせず、受入の許可 clinic・capability・cleanup は専用 fixture の receipt で確認する。
- local clinical data は [OLD_DB_HANDOFF_LOCAL.md](../deploy/OLD_DB_HANDOFF_LOCAL.md) と approved import contract に従う。データをこの文書へコピーしない。
- account/staff/permission は [STAFF_ACCOUNT_PROVISIONING.md](../deploy/STAFF_ACCOUNT_PROVISIONING.md) に従い、人間承認の receipt を確認する。
- destructive `make reset` は user-only。UAT agent が実行しない。

### 3.3 safe login check

承認済みの方法で `E2E_LOGIN_EMAIL` / `E2E_LOGIN_PASSWORD` を export 済みの shell で実行する。`.env.local` は shell として source しない。JSON は encoder を使って stdin から送り、response body は保存しない。credential 値を argv/log に置かない。

```bash
set -o pipefail
python3 -c 'import json,os,sys; json.dump({"email":os.environ["E2E_LOGIN_EMAIL"],"password":os.environ["E2E_LOGIN_PASSWORD"]},sys.stdout)'   | curl --silent --show-error --connect-timeout 5 --max-time 15 --output /dev/null --write-out '%{http_code}\n'       --request POST http://localhost:8080/api/v1/login       --header 'Content-Type: application/json'       --header 'X-Requested-With: XMLHttpRequest'       --header 'Origin: http://localhost:3003'       --data-binary @-
# expected: 200
```

### 3.4 browser readiness

For `browser=cdp`, Chrome `:9222` must listen and the project MCP must be connected. Use a UAT-only profile directory.

```bash
/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome   --remote-debugging-port=9222   --user-data-dir=/tmp/chrome-debug-uat   --no-first-run about:blank
```

For `browser=playwright`, Docker and `frontend/scripts/run-e2e.sh` must exist and the target fixture/account must be ready. The official Playwright image installs its own dependencies; do not check host `node_modules` as readiness evidence.

## 4. STG profile

- [STG-DEMO-DATA-LIFECYCLE.md](../deploy/STG-DEMO-DATA-LIFECYCLE.md) の承認済み UAT lane を使う。CSV は `002_master`、`APP_ENV=staging` では別 phase の合成 login seed も適用される。実環境への適用済み証拠や専用 UAT clinic の準備完了をこの配線だけで推定しない。
- staff/account は [STAFF_ACCOUNT_PROVISIONING.md](../deploy/STAFF_ACCOUNT_PROVISIONING.md) の preflight/approval/receipt で確認する。
- effective LIFF setting は approved deployment/settings runbook で人間が確認する。local compose defaults を STG evidence にしない。
- real LINE、external send、fixture apply、cleanup は人間の STG lane。欠落時は BLOCKED であり local PASS を代替にしない。

## 5. report と cleanup

`reports/uat-YYYY-MM-DD/` に profile、tenant/clinic、fixture receipt、pre/post counts、結果を残す。credential や個人情報は残さない。作成名の prefix だけを cleanup とみなさない。deterministic teardown がない場合は disposable DB を破棄する人間手順を選ぶ。

確認済み UAT FAIL は `bug.md` で dedupe/記録してから Linear で追跡する。environment/fixture shortage は BLOCKED とする。
