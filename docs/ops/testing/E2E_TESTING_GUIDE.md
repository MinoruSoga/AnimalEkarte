# E2E・システムテスト実行ガイド (End-to-End Testing)

> **目的**: 現在実装済みの Playwright coverage と supported runner を定義する。
> **最新更新**: 2026-09-06

## 1. 現在の coverage

`frontend/e2e/*.spec.ts` は次の回帰を含む。

- auth、route/navigation、business/clinical/accounting/settings/operations smoke
- owner/patient/reservation/medical-record search、pagination、selected create/update flows
- selected accounting、estimate、hospitalization、inventory/master/settings CRUD
- checkup、examination、vaccination、shift、trimming、LINE reservation、L-step の selected flows
- read-only UI design check

これは予約から会計まで、入院 daily care から請求まで、LIFF から calendar/L-step tag までの一連の journey を end-to-end で保証していない。これらは target/manual acceptance journey であり、L4 の [scenarios/](scenarios/README.md) を正本とする。

## 2. 安全な実行条件

- disposable local DB または承認済み isolated UAT tenant だけを使う。production と未承認 shared STG clinic は禁止。
- migration seed `002_master` は account/clinical fixture を含まない。[UAT setup](UAT-ENV-SETUP.md) に従い明示 provisioning する。
- suite は write を行い、全 spec が teardown を保証するわけではない。対象 clinic、pre/post counts、deterministic cleanup/idempotency を先に定義する。保証できなければ disposable DB だけで実行する。
- `PLAYWRIGHT_TEST_BASE_URL` が設定可能であることは、その target が安全という意味ではない。

## 3. runner

Supported full-suite route は `make e2e` または `frontend/scripts/run-e2e.sh`。runner は official Playwright Docker image を使う。

- `playwright.config.ts` の direct default: `http://localhost:3003`
- Docker runner の effective default: `http://host.docker.internal:3003`
- authentication: `E2E_LOGIN_EMAIL` / `E2E_LOGIN_PASSWORD`
- auth state: `E2E_AUTH_STATE_PATH` で上書き可能

```bash
# repository root
make e2e

# scoped spec
cd frontend && ./scripts/run-e2e.sh e2e/accounting-flow.spec.ts
```

Current wrapper は headless-only と扱う。DISPLAY/Wayland/X11/VNC を接続しないため `--headed` は supported procedure ではない。

## 4. GitHub workflow と artifact の現状

`.github/workflows/e2e.yml` は `workflow_dispatch` の optional manual workflow であり、PR/push gate ではない。

- workflow が実行するのは `auth-flows.spec.ts` のみ。`APP_ENV=test` と合成 `E2E_LOGIN_*` を渡し、migrate の login seed を利用する配線は実装済み。
- `--auth-smoke` は同 spec の runner alias。`--clinical` は別の 10 spec allowlist と disposable clinic setup/teardown を持つ（[CLINICAL-E2E-DESIGN.md](CLINICAL-E2E-DESIGN.md)）。workflow に clinical/full suite job はない。
- auth smoke の成功、fresh DB、`--clinical` の実行結果は、このソース照合では確認していない。全 suite には退役 demo fixture の固定氏名/ID に依存する spec が残るため、login seed だけで実行準備完了とはしない。
- runner は `--reporter=list` の console output を使い、host-mounted HTML report を生成しない。workflow の `frontend/playwright-report/` upload target と runner output の不一致は残る。

workflow の配線、実行成功、artifact の存在を区別する。`--clinical` の環境チェックと通常終了時 teardown は、全 suite の isolation/cleanup を保証しない。

## 5. pass/report contract

- 実行した spec は 100% PASS。skip/retry は明示する。
- allowed clinic 以外へ mutation がないことを pre/post evidence で確認する。
- own-clinic row も deterministic teardown または disposable DB disposal で処理する。
- 現行 runner の retained output は console のみ。UAT evidence は `reports/uat-YYYY-MM-DD/` に別途記録し、credential/cookie/idToken を含めない。

## 6. LIFF boundary

Playwright は実 LINE app 内 SDK を保証しない。mock と実機の security/config boundary は [liff-verification.md](liff-verification.md)、acceptance steps は [S04](scenarios/S04-liff-reservation-journey.md) と [S12](scenarios/S12-liff-pet-health.md) を正本とする。本書には手順を複製しない。
