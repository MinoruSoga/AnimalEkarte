# LIFF 検証経路（local mock / STG 実 LINE）

> **目的**: mock と実 LINE の保証範囲、設定 source、秘密管理境界を定義する。acceptance steps は [S04](scenarios/S04-liff-reservation-journey.md) と [S12](scenarios/S12-liff-pet-health.md) を正本とする。
> **最新更新**: 2026-09-01

## 1. guarantee boundary

| route | guarantees | does not guarantee |
|:--|:--|:--|
| local mock | shared LIFF hooks、mock-token API/UI paths | real SDK、idToken signature、LINE app、channel settings |
| remote CI intent | mock-only token scope | manual E2E auth smoke の実行成功、実 LINE、clinical/full suite の保証。合成 login 配線はあるが実行証跡とは別 |
| STG real LINE | approved dedicated UAT clinic での real SDK/idToken/in-client behavior | local gate の代替ではない。human lane only |

Local mock PASS を real LINE PASS とみなさない。STG prerequisite が欠けた場合は BLOCKED/Needs Human とする。

## 2. local mock

- Backend `LIFF_MOCK=true`: `backend/internal/middleware/liff_auth.go` の mock path。
- Frontend `VITE_LIFF_MOCK=true`: `frontend/src/shared-liff/use-liff.ts` と app config。
- release mode は backend の `LIFF_MOCK=true` を拒否する。
- mode flags の名前と nonsecret value `true/false` は診断出力に使える。token、channel secret、cookie、idToken は出力しない。
- mock link screen が API を呼ばず success 表示だけになる場合がある。link establishment の証明には backend mock API または STG real LINE が必要。

Scoped regression commands は project Docker policy に従う。

```bash
docker compose exec frontend pnpm exec vitest run   src/shared-liff/use-liff.test.ts   liff/src/hooks/use-liff-link.test.ts

docker compose exec backend go test ./internal/middleware/ -count=1 -run 'LiffAuth_Mock'
```

Full Playwright route は `make e2e` / `frontend/scripts/run-e2e.sh` だけを supported とする。local `pnpm --dir frontend test:e2e` は案内しない。手動 acceptance は S04/S12 を使い、結果を `reports/uat-YYYY-MM-DD/` に置く。

## 3. configuration and secret storage

LIFF/LINE configuration は単一の global env model ではない。

| component | source |
|:--|:--|
| health-card/link SPA | build-time `VITE_LIFF_ID` |
| line-reserve SPA | clinic settings API の per-clinic `liff_id` |
| LINE channel credentials | per-clinic encrypted settings / `clinic_integrations` |
| platform bootstrap/encryption keys | approved platform secret manager |

`LINE_CHANNEL_ACCESS_TOKEN` / `LINE_CHANNEL_SECRET` を runtime の必須 global environment variables とみなさない。per-clinic credentials は approved encrypted settings flow に置く。secret value は repo、docs、PR、chat、logs に書かない。詳細は [LINE setup](../../spec/line/setup.md)、[LINE architecture](../../spec/line/architecture.md)、[STG runbook](../infra/staging/runbook.md) に従う。

## 4. STG real LINE

- approved dedicated UAT clinic と explicitly provisioned account/fixture を使う。
- effective deployment/settings で mock disabled、correct per-clinic LIFF ID/credentials を人間が確認する。local compose defaults は evidence にならない。
- real device で S04/S12 を実行する。external message/file send は human lane が明示承認した場合だけ。
- evidence は date、device/OS、LINE version、result に限定し、token、secret、idToken、personal data を残さない。

## 5. prohibited

- release/STG で mock を有効化する
- real token/channel secret を取得、表示、記録する
- mock UI success だけで account link を PASS にする
- manual/non-gating E2E workflow を working CI gate と表現する
- production または unapproved shared clinic で real send/mutation を行う
