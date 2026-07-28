# Durable Scheduler Operations

> **目的**: Cloudflare scheduled eventで実行する定期jobを、安全に確認・停止・再開・catch-upする。
> **対象**: `animalekarte-scheduler-v1`。実装の正本は`backend/worker/scheduled-*`、`scheduler-*`、操作CLIの正本は`infra/scripts/cf-scheduler-ops.sh`。
> **release状態 (2026-07-24)**: code、Wrangler設定、local Worker testは実装済み。今回の変更をSTG/productionへdeployした事実、実環境cron発火、通知、停止・復旧rehearsalはまだない。実環境操作はrelease承認後だけ行う。

## Schedule contract

Cloudflare cronはUTCで解釈される。Workerは次の3式以外をfail-closedで拒否する。

| Cron (UTC) | JST | 実行順 |
|:---|:---|:---|
| `0 1 * * *` | 10:00 | `no_show` → `delivery` |
| `0 6,11 * * *` | 15:00 / 20:00 | `no_show` |
| `0 17 * * *` | 翌02:00 | `dormant` |

`no_show`は予約no-show処理、`delivery`は配信trigger、`dormant`は休眠判定を、Go Containerの内部endpointからdomain use caseとして実行する。1 jobの期限はGo 100秒、Worker fetch 110秒、coordinator lease 150秒である。leaseはWorker側ledgerの二重確定を防ぐが、開始済みのGo side effectを取り消さないため、domain側のidempotency/CASも安全境界に含まれる。

## Access and secrets

操作endpointは`/_internal/scheduler/*`で、公開APIへproxyしない。次のいずれかで認証する。

- automation: 32 UTF-8 byte以上の用途専用`SCHEDULER_OPS_SECRET`をBearer tokenとして使用
- human operator: Cloudflare Access JWT。team domainのJWKS、issuer、audience、署名、`exp`/`nbf`をWorker内で検証する

CLIはautomation secret方式を使う。JWT、migration secret、application JWT secretと共有しない。`openssl rand -base64 48`相当の暗号学的乱数で生成し、承認済みsecret storeへ直接保存する。32 UTF-8 byte未満の値はWorkerがfail-closedで拒否する。値をshell history、ログ、runbookへ記録しない。

AccessのJWKSはWorker isolate内で10分cacheする。未検証JWTの任意`kid`、JWKS upstream failure、同時requestが外部取得を増幅しないよう、unknown-key/failure後の再取得はteam domainごとに60秒cooldownし、同時取得は1本へ集約する。鍵rotation後の新しい`kid`はcooldown後に再取得して検証するため、最後の取得直後にrotationした場合は最大60秒fail-closedになる。

このcache/cooldownはdefense-in-depthであり、isolate再生成や複数PoPをまたぐ共有rate limitではない。Cloudflare Access policyで対象hostと`/_internal/scheduler/*`をWorker実行前に保護し、必要なedge rate limitも適用する。署名検証後のDurable Object rate limit（actorごと60 request/分）は、未認証trafficの代替防御として数えない。

```bash
export SCHEDULER_OPS_BASE_URL="https://<exact-worker-host>"
export SCHEDULER_OPS_ALLOWED_HOST="<exact-worker-host>"
export SCHEDULER_OPS_SECRET="<secret-from-approved-secret-store>"
```

CLIはHTTPS、host完全一致、redirect禁止、接続/全体timeout、4KiB以下の厳格JSON requestを強制する。`jq`が必要で、request ID省略時は`uuidgen`が必要である。

## Read-only status

変更操作の前後に必ずstatusを保存する。返却されるsecretやPIIはなく、control revision、active lease、直近run、操作履歴を確認できる。

```bash
pnpm cf:scheduler status
pnpm cf:scheduler status 50
```

確認項目:

- `control.paused`と`control.revision`
- `active`が残っていないか。lease期限前なら重複操作しない
- 対象slot/jobが`recentRuns`に既に存在しないか
- `partial`、`failed`、`busy`、`stale`、`fenced`とfailure code
- `recentOperations`のrequest ID、actor、reason、結果

run ledgerは35日、operator操作履歴は400日保持する。statusは最大50件を返すため、長期監査の唯一の保存先にはしない。

## Pause and resume

pause/resumeはproduction-impacting操作であり、対象環境、理由、実施者、復旧条件の承認を先に得る。現在の`control.revision`をstatusから取得し、compare-and-setで更新する。同じrevisionへの競合操作は`409 revision_conflict`となる。

```bash
pnpm cf:scheduler pause <current-revision> "incident-1234 scheduler containment"
pnpm cf:scheduler status

pnpm cf:scheduler resume <current-revision> "incident-1234 recovery approved"
pnpm cf:scheduler status
```

reasonは4〜200文字で制御文字を含めない。request IDを再利用する場合は、同一actor・同一payloadに限る。異なる意図で同じIDを使うと`request_id_conflict`になる。

pauseは新しいrun claimを止めるが、既にContainerへ送信したside effectを取り消さない。`active`がある場合はstatusとlogsで完了/timeoutを確認してから復旧判断する。

## Missing-slot catch-up

catch-upもproduction-impacting操作である。次をすべて満たした場合だけ実行する。

1. statusとWorkers Logsから「そのslotが未記録」であることを確認した。
2. 業務ownerが再実行時のside effectと対象件数を確認した。
3. schedulerがresume済みで、active runがない。
4. slotは現在より過去、35日以内、UTCの正時で、jobに対応するcron slotである。
5. 専用のUUID request IDと、incident/change IDを含むreasonを用意した。

Unix epoch millisecondsは、UTC時刻を明示して生成する。変換結果を別担当者と照合する。

```bash
slot_ms="<verified-utc-slot-epoch-ms>"
request_id="$(uuidgen | tr '[:upper:]' '[:lower:]')"

pnpm cf:scheduler run no_show "${slot_ms}" "incident-1234 approved catch-up" "${request_id}"
pnpm cf:scheduler status 50
```

01:00 UTCで`no_show`と`delivery`の両方が欠けた場合も、jobごとに別request IDで1件ずつ実行する。既にledgerがあるslotは`slot_already_recorded`、pause中は`scheduler_paused`、別run中は`scheduler_busy`で拒否される。HTTP `202`はpendingであり成功ではないため、statusでterminal resultまで確認する。

## Logs and alerts

Cloudflare Workers Logsで次のstructured eventを検索する。

- `scheduler_job_failed`
- `scheduler_invocation_failed`
- `scheduler_ops_request_failed`
- `scheduler_alert_not_configured`
- `scheduler_alert_delivery_failed`
- lease/fence関連のfailure code

失敗通知は`SCHEDULER_ALERT_WEBHOOK_URL`へHTTPS POSTし、`SCHEDULER_ALERT_ALLOWED_HOST`とのhost完全一致、専用Bearer secret、`Idempotency-Key`を使う。redirect、非2xx、timeout、host不一致、未設定は成功扱いにしない。少なくとも以下をrelease前に実値で確認する。

- `SCHEDULER_ENVIRONMENT`
- `SCHEDULER_ALERT_ALLOWED_HOST`
- `SCHEDULER_ALERT_WEBHOOK_URL`
- `SCHEDULER_ALERT_WEBHOOK_SECRET`
- `SCHEDULER_ACCESS_TEAM_DOMAIN`
- `SCHEDULER_ACCESS_AUDIENCE`
- `SCHEDULER_OPS_SECRET`

secretの存在名だけでは通知成立の証拠にならない。STGでAccess JWT、ops secret、通知の受信、失敗時のWorkers Logsを実測し、時刻とrun IDをrelease evidenceへ残す。

## Incident recovery

1. `status 50`とWorkers Logsを保存し、影響したjob/slot/clinic範囲を特定する。
2. 継続実行が危険なら、承認を得てrevision付きpauseを行う。
3. Go API `/health`、Container、PlanetScale、外部連携を切り分ける。AWSは退役済みで切り戻し先ではない。
4. active lease中は二重runしない。timeout後もdomain side effectの有無をDB/auditから確認する。
5. 原因を修正し、通知とstatusが正常であることを確認する。
6. 必要なmissing slotだけを承認付きcatch-upする。
7. revision付きresume後、次の自然発火と通知を確認する。
8. incident記録へstatus、run ID、request ID、reason、logs、件数、承認者、復旧時刻を残す。secret値は残さない。

## Release qualification

次を満たすまで「scheduler運用開始」と判定しない。

- Worker test、typecheck、Wrangler dry-run、dependency auditがgreen
- STGへexact versionをdeployし、3 cron bindingが反映された
- no-opまたは安全なfixtureで各jobの自然発火を確認した
- pause/resume、revision conflict、missing-slot catch-up、重複拒否をrehearsalした
- Accessとops secretの両認証、rate limit、alert受信/失敗時ログを確認した
- forged/unknown-key JWTの連続送信でJWKS取得が60秒に1回以下へ抑制され、Access policyが未認証requestをedgeで拒否することを確認した
- Container restart/rolling update後もcoordinator stateとduplicate防止を確認した
- production placeholder、32 UTF-8 byte以上の専用ops secret、allowed host、通知先が実値化・review済み
- production実施者とrollback/recovery判断者の承認がある

参考:

- [Cloudflare Cron Triggers](https://developers.cloudflare.com/workers/configuration/cron-triggers/)
- [Scheduled handler](https://developers.cloudflare.com/workers/runtime-apis/handlers/scheduled/)
- [Workers Logs](https://developers.cloudflare.com/workers/observability/logs/workers-logs/)
- [Tail Workers](https://developers.cloudflare.com/workers/observability/logs/tail-workers/)
