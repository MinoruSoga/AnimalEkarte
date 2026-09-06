# QA-FULL-CLINICAL-E2E — clinical / data-dependent E2E 設計

状態: **helper + allowlist 置換済み / `--clinical` 未実行 / full suite は別承認**  
範囲の正本: [TEST_ARCHITECTURE.md](./TEST_ARCHITECTURE.md) L3  
設計・局所 GREEN だけでは E2E を PASS にしない。auth smoke 成功を full suite coverage と扱わない。

## 対象 scenario / spec

auth smoke と **別 allowlist** にする。

| 区分 | ファイル | 今回の対象 |
|------|----------|------------|
| CI / 手動 workflow 既存 | `frontend/e2e/auth-flows.spec.ts` | 対象外。現行 `.github/workflows/e2e.yml` のまま |
| clinical / data-dependent | `clinical-flows.spec.ts`、`clinical-smoke.spec.ts`、`medical-records-*.spec.ts`、`examinations-flow.spec.ts`、`vaccinations-flow.spec.ts`、`checkups-flow.spec.ts`、`hospitalization-flow.spec.ts`、`estimates-flow.spec.ts` | 対象。退役 `003_demo` の固定氏名依存は合成 fixture 参照へ置換済み |
| 会計・予約・マスタ | `accounting-*.spec.ts`、`reservations-*.spec.ts`、`master-crud.spec.ts` 等 | 同一 fixture 契約が必要なら第 2 allowlist。初回実装には入れない |
| UI 監査 / LIFF | `ui-design-compliance-readonly.spec.ts`、`line-reservation-flow.spec.ts` | 対象外 |

L4（S01–S13 / V01–V05）の代替ではない。

## 使い捨て環境・clinic

| 項目 | 値 |
|------|-----|
| 環境 | ローカル compose、または承認済み disposable CI job。共有 STG 禁止 |
| `APP_ENV` | `test`（CI と同じ）。staging/production は拒否 |
| seed | `002_master` + `APP_ENV=test` の login seed のみ。`003_demo` / `004_staging` を復元しない |
| clinic | 合成 clinic 1 件。本番/STG clinic_id 1・2 を使わない |
| 認証 | 公開合成 login（現行 E2E と同じ系統）。値は証跡に書かない |
| LIFF | compose 既定の mock。実 LINE / 実 LSTEP は禁止 |

## 合成データ

`003_demo` の固定行を前提にしない。setup が新規作成する。

- owner / pet / 確定済みカルテ 1 件以上（一覧・検索・行遷移）
- 予約種別・入院ケージ等、allowlist が触るマスタは `002_master` または clinic スコープの合成行
- 会計完了時刻の操作は [S09-FIXTURE-DESIGN.md](./S09-FIXTURE-DESIGN.md) に任せ、本 suite では要求しない
- `medical-records-create.spec.ts` の synthetic interceptor は allowlist 外を遮断し、mount 時 POST をローカル fulfill する。この spec は DB への作成/永続化を証明しない。他の clinical spec 全体へ同じ interceptor が自動適用されるわけではない。

## 実装済みファイルと残る実行

- `frontend/e2e/helpers/` — disposable clinic の setup/teardown（合成 API または承認済み UAT helper）
- `frontend/e2e/clinical-*.spec.ts` / `medical-records-*.spec.ts` — デモ氏名ハードコードを fixture 参照へ置換
- `frontend/scripts/run-e2e.sh` — `--clinical` と `--auth-smoke` を分離
- `.github/workflows/e2e.yml` — **変更しない**（auth smoke のまま）。full suite job は別承認
- 本ファイルの実装検証節

## 許可操作 / 禁止操作

許可:

- 合成 clinic への書き込みと、その clinic に閉じた Playwright
- `synthetic-api` のローカル stub
- 失敗時の Playwright report（credential / cookie / idToken を含めない）

禁止:

- 共有 STG / PROD への `PLAYWRIGHT_TEST_BASE_URL`
- 実 LINE idToken、実 LSTEP write
- auth smoke 成功をもって clinical suite PASS
- 時計変更、既存医院データの改変
- workflow の push 自動実行化

## 環境誤指定時の停止

runner 先頭で fail-closed:

1. `APP_ENV` が `test` 以外
2. base URL が `http://localhost:3003` / compose frontend 以外
3. clinic が除外 ID、または fixture owner が未作成
4. teardown 手順が未登録

1 つでも欠けたら suite を開始しない。

## cleanup / 失敗時回収

- 通常の Playwright 終了時（spec 失敗を含む）は `run-e2e.sh` が fixture CLI の teardown を呼び、teardown 失敗ならその非ゼロ exit を返す。spec の `afterAll` は browser context を閉じる。clinic 削除は runner の責任。
- runner に EXIT/INT/TERM trap はなく、setup 後の中断・プロセス異常時の自動回収は保証しない。使い捨て DB と人手の回収手順を実行前に用意する。full suite 用 CI job / `down -v` の回収は未配線。
- ローカルは clinic sweep。共有 DB の手削除はしない
- teardown 失敗は UNREPORTED/BLOCKED。部分 PASS を full suite PASS にしない

## 実装検証（2026-09-06）

入れたもの:

- `backend/internal/clinicale2e` — APP_ENV=`test` + ローカル DB host のみ。clinic 1/2 拒否。合成 clinic / staff / owner / pet / 確定カルテ 21 件と allowlist 用の検査・予防接種・健診・入院・下書き見積を INSERT
- `backend/cmd/clinical-e2e-fixture` — `setup` / `teardown --clinic-id`。パスワードは `E2E_LOGIN_PASSWORD` からハッシュ。stdout は JSON のみ（秘密なし）
- `frontend/e2e/helpers/clinical-*.ts` — fail-closed（APP_ENV / base URL / teardown / fixture）
- allowlist spec の 003_demo 氏名（林 文明 / Iris）を fixture 参照へ置換
- `frontend/scripts/run-e2e.sh --clinical` / `--auth-smoke`
- `.github/workflows/e2e.yml` は未変更

未実施（このスライスでは PASS にしない）:

- ローカル `--clinical` 1 回
- `e2e.yml` への full suite job
- Linear Done

局所ユニット:

```bash
docker compose exec backend go test ./internal/clinicale2e/ -count=1
cd frontend && ./scripts/run-e2e.sh e2e/helpers/clinical-env.spec.ts e2e/helpers/clinical-fixture.spec.ts
```

`--clinical` の中身は allowlist のみ。`make e2e` 全件は別承認。`--clinical` は backend `APP_ENV=test` が無いと fail-closed。

## 承認が必要な実行操作

| 操作 | 承認 |
|------|------|
| 本設計 | source 定義済み。受入 sign-off とは別 |
| fixture helper と allowlist 置換 | このスライスで完了。E2E PASS ではない |
| ローカル `--clinical` 1 回 | 別承認 |
| `e2e.yml` への full suite job 追加 | USER。non-gating のまま |
| workflow_dispatch / push | USER |
| Linear Done / UAT PASS 転記 | USER。設計・局所 GREEN だけではしない |
