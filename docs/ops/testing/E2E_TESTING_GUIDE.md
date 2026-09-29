# E2E・システムテスト実行ガイド (End-to-End Testing)

> **目的**: 現在実装済みの Playwright coverage と supported runner を定義する。
> **最新更新**: 2026-09-29

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

`.github/workflows/e2e.yml` は `workflow_dispatch` の optional manual workflow であり、PR/push gate ではない（2026-09-29 訂正: 旧記述は「auth smoke のみ実行・clinical/full suite job なし」。2026-09-28・EMR-128 で `inputs.suite` 振り分けが配線された）。

- `inputs.suite`（required、default `auth-smoke`）が `auth-smoke` / `clinical` / `v04` を `run-e2e.sh` の同名モードへ振り分ける。`auth-smoke` は `auth-flows.spec.ts` を実行し、`APP_ENV=test` と合成 `E2E_LOGIN_*` を渡して migrate の login seed を利用する。`clinical` / `v04` は job の使い捨て `APP_ENV=test` compose stack 上で fixture clinic を setup/teardown し、`E2E_RESULTS_DIR` の test-results を常時 artifact 化する。push/PR 自動実行・full suite job はない。
- `--auth-smoke` は auth spec の runner alias。`--clinical` は別の 10 spec allowlist と disposable clinic setup/teardown を持つ（[CLINICAL-E2E-DESIGN.md](CLINICAL-E2E-DESIGN.md)）。`--v04` は同じ fixture ゲートを共有する。
- auth smoke の Actions 実行・fresh DB 結果はこの照合では確認していない。`--clinical` のローカル実行は 2026-09-23 に 1 回（33 PASS / 7 FAIL = spec 側 drift、green 未達。証跡は gitignore 対象の `reports/uat-2026-09-23/clinical-e2e-emr128/`）。workflow_dispatch での clinical / v04 実行は未。全 suite には退役 demo fixture の固定氏名/ID に依存する spec が残るため、login seed だけで実行準備完了とはしない。
- runner は `--reporter=list` の console output を使い、host-mounted HTML report を生成しない。workflow の `frontend/playwright-report/` upload target（failure 時のみ）と runner output の不一致は残る。一方、`E2E_RESULTS_DIR`（runner が `…/test-results` へ mount）の upload は clinical / v04 で `if: always()` 化済み。

workflow の配線、実行成功、artifact の存在を区別する。`--clinical` の環境チェックと通常終了時 teardown は、全 suite の isolation/cleanup を保証しない。

```mermaid
flowchart TB
    W["e2e.yml workflow_dispatch<br>manual・non-gating<br>inputs.suite"] --> A[auth-smoke<br>auth-flows.spec]
    W --> CW["clinical / v04<br>disposable APP_ENV=test stack<br>+ fixture setup/teardown"]
    R["local runner<br>make e2e / run-e2e.sh"] --> A
    R --> C["--clinical / --v04<br>別 allowlist + disposable clinic fixture"]
    R --> F[full suite 全件]
    F -. e2e.yml job なし・別承認 .-> J[full suite job]
```

## 5. pass/report contract

- 実行した spec は 100% PASS。skip/retry は明示する。
- allowed clinic 以外へ mutation がないことを pre/post evidence で確認する。
- own-clinic row も deterministic teardown または disposable DB disposal で処理する。
- 現行 runner の retained output は console のみ。UAT evidence は `reports/uat-YYYY-MM-DD/` に別途記録し、credential/cookie/idToken を含めない。

## 6. LIFF boundary

Playwright は実 LINE app 内 SDK を保証しない。mock と実機の security/config boundary は [liff-verification.md](liff-verification.md)、acceptance steps は [S04](scenarios/S04-liff-reservation-journey.md) と [S12](scenarios/S12-liff-pet-health.md) を正本とする。本書には手順を複製しない。
