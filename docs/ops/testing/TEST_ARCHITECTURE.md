# テストアーキテクチャ (Test Architecture)

> **目的**: 検証層、正本、実行手段、証跡、環境境界を一か所で定義する。
> **最新更新**: 2026-09-06

## 1. 原則

1. 同じ実行手順を複数文書へ複製しない。
2. L3 は自動回帰、L4 は [scenarios/](scenarios/README.md) による受入であり、相互に代替しない。
3. 結果は `reports/uat-YYYY-MM-DD/` に置き、scenario 本文へ書かない。
4. production、未承認の共有 STG clinic、実患者データに対して mutating test を行わない。使い捨て local DB または承認済みの専用 UAT tenant だけを使う。
5. 最終合否、shared STG、実 LINE、破壊操作は人間の承認レーンとする。

## 2. L0–L5 正本マップ

| 層  | 証明対象                             | 正本・実装                                         | 現在の実行経路                                                                                                                                                                    |
| :-- | :----------------------------------- | :------------------------------------------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| L0  | 正しい仕様                           | `docs/spec/`、product philosophy、ADR              | review                                                                                                                                                                            |
| L1  | 関数・component                      | code 隣の `*test.go` / `*.test.ts`                 | scoped test、CI                                                                                                                                                                   |
| L2  | HTTP、DB、認可、FK、clinic isolation | domain HTTP tests、inventory/guardrail             | path-filtered backend build/test shards + aggregate coverage ratchet。local `make ci` は inventory/guardrail checks                                                               |
| L3  | 実装済み画面回帰                     | `frontend/e2e/`、[E2E guide](E2E_TESTING_GUIDE.md) | local `make e2e` / `frontend/scripts/run-e2e.sh`。manual GitHub workflow は non-gating の auth smoke のまま。`--clinical` helper は実装済みだが未実行。auth smoke 成功を full suite coverage と扱わない |
| L4  | 業務・フォーム受入                   | [scenarios/](scenarios/README.md)                  | Chrome DevTools、scripted browser、人手                                                                                                                                           |
| L5  | focused exploratory / post-deploy    | [SECTION_14](SECTION_14_MANUAL_TEST_GUIDE.md)      | AI または人手                                                                                                                                                                     |

## 3. L4 の範囲

S01–S13 と V01–V05 が宣言済み受入範囲である。unique form総数はinventory再構築完了まで算定保留である。したがって次を区別する。

- **列挙済み項目**: [FIELD-LEVEL-PROTOCOL.md](scenarios/FIELD-LEVEL-PROTOCOL.md) の適用対象。
- **wildcard/実測待ち**: source と照合して列挙されるまで coverage gap。
- inventory-to-source drift gate は未実装。現時点で「全画面の全 field を機械的に網羅済み」と主張しない。

フォーム完了は、列挙済み全 field の適用可能な F check が PASS または理由付き N/A、かつ C1–C3 が PASS であること。wildcard gap はレポートで PARTIAL/BLOCKED として明示する。納品前の対象範囲は scenario index で決める。

## 4. 実行経路と安全境界

1. Project-configured Chrome DevTools MCP (`http://127.0.0.1:9222`)
2. `reports/uat-*/` に置く再現 script
3. Playwright MCP は user/global で明示設定されている場合だけ利用可能
4. 人手（実 LINE、承認が必要な操作、最終合否）

Mutating run の前に、対象 clinic ID、fixture owner、pre-count、期待 post-count、teardown/idempotency をレポートに記録する。teardown が保証できない suite は disposable DB だけで実行する。任意の `PLAYWRIGHT_TEST_BASE_URL` を安全とみなさない。

## 5. 環境プロファイル

| profile           | CSV seed / login phase                           | account / clinical fixture                                                                                                                         | LIFF                                   | 判定                                                                                                  |
| :---------------- | :--------------------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------- | :------------------------------------- | :---------------------------------------------------------------------------------------------------- |
| local UAT         | CSV `002_master` + 許可環境の login seed                        | [local handoff](../deploy/OLD_DB_HANDOFF_LOCAL.md) / approved import と [account provisioning](../deploy/STAFF_ACCOUNT_PROVISIONING.md) を明示実施 | local compose effective config で mock | disposable/local-only                                                                                 |
| STG UAT           | CSV `002_master` + staging login seed                        | [STG lifecycle](../deploy/STG-DEMO-DATA-LIFECYCLE.md) の承認済み skeleton/import/staff-account lane                                                | mock 禁止。実機は人間レーン            | dedicated UAT tenant only                                                                             |
| CI E2E auth smoke | `002_master` + `APP_ENV=test` login seed | public synthetic login fixture                                                                                                                     | mock intent                            | manual/non-gating `auth-flows.spec.ts` の配線を実装。Actions 実行・fresh DB 結果は UNREPORTED/UNKNOWN |
| CI E2E full suite | CSV `002_master` + 許可環境の login seed                        | `--clinical` helper は repo にある。実行と e2e.yml job は未                                                                                          | mock intent                            | BLOCKED。auth smoke の実装を full suite coverage と扱わない                                           |

CSV `002_master` は account/clinical rows を含まない。一方、migrate の別 phase `003_login` は `development/local/dev/test/staging` で catalog account/staff を upsert する（`backend/internal/seedlogin/env.go`、`backend/cmd/migrate/login_seed.go`）。「CSV に account がない」と「startup が account を作らない」を混同しない。production・空・未知の環境は login seed 対象外。

`make migrate`、`make seed`、通常の startup は退役済み `003_demo` / `004_staging` を復元しない。backend startup は migrate を実行してから healthy になる。一方、migration 変更を pull した後は project policy に従い、**ユーザーが** `make migrate` を実行する。

## 6. 記録と defect intake

| 判定    | 意味                          | 記録                                                  |
| :------ | :---------------------------- | :---------------------------------------------------- |
| PASS    | 期待どおり                    | report                                                |
| PARTIAL | 一部未確認                    | report。完了と呼ばない                                |
| BLOCKED | environment/spec/fixture 不足 | report または Linear Needs Human。`bug.md` へ書かない |
| FAIL    | 確認済み製品欠陥              | `bug.md` で dedupe/記録後、Linear で追跡              |

その他の新規製品 defect は通常の Linear intake に従う。証跡に credential、token、cookie、idToken、個人情報を含めない。

## 7. 関連文書

- [UAT environment](UAT-ENV-SETUP.md)
- [E2E guide](E2E_TESTING_GUIDE.md)
- [LIFF boundary](liff-verification.md)
- [Integration plan](INTEGRATION_TEST_PLAN.md)
- [Coverage policy](../coverage-policy.md)
