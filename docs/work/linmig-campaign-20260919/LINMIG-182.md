# LINMIG-182 — Lab-device spec versus acceptance, consumer-token supply

Campaign `linmig-ops-prep-20260919` revision 1. Claim id `LINMIG-182` (Linear issue not found; local packet only). This sheet maps [`docs/ops/deploy/LAB_DEVICE_CONNECTIVITY.md`](../../ops/deploy/LAB_DEVICE_CONNECTIVITY.md) and current lab-import / lab-device APIs to acceptance, names the consumer-token supply and browser-to-agent distribution path, and marks real-device UAT as remainder.

**This unit does not** edit application code, mint tokens, send payloads, open serial ports, or touch STG/PROD. No token values appear here.

Callers: campaign controller and operators reading `docs/work/linmig-campaign-20260919/` (same pattern as `docs/work/todo-campaign-20260918/*.md`). No application import. Net-new file; no prior `LINMIG-182.md`.

## Scope and exclusions

| In | Out |
| --- | --- |
| Spec versus current implementation contract | Application code changes |
| Consumer-token mint / env / bundle / LaunchAgent / browser distribution design | Real token generation or paste |
| Remaining clinic-hardware UAT gaps | Device I/O, STG/PROD writes |
| Local docs sheet only | Drワン (`source_type=drwan`), `mkan.mdb` |

Drワン remains excluded. `mkan.mdb` is not an input. COM4 / DRI-CHEM 7000V is not in scope.

## Daily path (implementation contract)

Spec (`LAB_DEVICE_CONNECTIVITY.md` L8–L28, L41–L42): file upload is not the daily path. One exam Mac owns wired serial via user LaunchAgent `lab-device-agent`. The agent binds loopback HTTP `127.0.0.1:17654`. Browser `/lab-device` (`LabDeviceBoard`) does not own serial; it polls the agent with a consumer token obtained from the authorized API, then posts frames to the API.

```
device serial --USB-Serial--> lab-device-agent (LaunchAgent)
                                    |
                                    | loopback HTTP 127.0.0.1:17654
                                    | claim / frames / ack|reject
                                    v
authenticated browser /lab-device (LabDeviceBoard)
                                    |
                                    | JWT to API only (not stored on agent)
                                    | POST /api/v1/lab-device/frames  device_hint=auto
                                    v
medicalrecord receive + board + attach/detach
```

Code citations:

- Agent listen address: `backend/internal/labdeviceagent/http.go` `ListenAddress = "127.0.0.1:17654"`
- Browser client: `frontend/src/features/lab-device/lib/lab-device-agent.ts` `LAB_DEVICE_AGENT_URL`
- Board: `frontend/src/features/lab-device/routes/LabDeviceBoard.tsx` (`useGetLabDeviceAgentConsumer` + `useLabDeviceAgentListen`)
- Drain: `drainLabDeviceAgentFrames` posts `deviceHint: "auto"` then ACK; HTTP 400 → reject; other failures retry
- Serial owner: `backend/cmd/lab-device-agent/main.go` (ports-file `/dev/cu.usbserial-*` only; empty token fails closed)

## Spec versus current implementation versus acceptance

Status values: **code-aligned** (repo matches spec), **ops remainder** (code exists, clinic/hardware or shared-env not observed), **blocked** (fail-closed until an ops path exists), **excluded**.

Clinic hardware, live LaunchAgent, and live token values are **UNKNOWN** in this session (no device I/O, no secret reads).

| Spec / acceptance item | Spec source | Current implementation | Acceptance now |
| --- | --- | --- | --- |
| Daily path is LaunchAgent + loopback, not Web Serial / file upload | `LAB_DEVICE_CONNECTIVITY.md` L8–L9, L41–L42; ADR-008 | `LabDeviceBoard` + `useLabDeviceAgentListen`; no serial APIs in the board | **code-aligned**. Clinic install **UNKNOWN** |
| Browser does not open serial | same | `lab-device-agent.ts` talks only to `127.0.0.1:17654` | **code-aligned** |
| One `/lab-device` board; permission `lab-import`; no confirm dialog; no pet search; today-visit cards; day-grouped received list | `LAB_DEVICE_CONNECTIVITY.md` L41; `docs/spec/ui-design-compliance.md` | `LabDeviceBoard.tsx` `ResourceLabImport`; `labDeviceSelectableTodayVisits`; `groupLabDeviceCardsByDay` | **code-aligned** |
| Wait is a selected today-visit, not a TTL numeric UI | spec L42 | `PUT/DELETE /api/v1/lab-device/wait`; board shows wait, no TTL editor | **code-aligned** |
| Attach from unlinked is `{ pet_id }` only; values not edited | spec L43; ADR-007 | `POST /api/v1/lab-imports/:job_id/attach` `labDeviceAttachRequest.PetID`; FE `useAttachLabDeviceJob` | **code-aligned** |
| Device frames go to `/lab-device/frames`, not `POST /lab-imports` | spec L47–L50; `lab_import_service.go` PreviewBatch device sources | `ReceiveLabDeviceFrames` `POST /api/v1/lab-device/frames`; preview of device `source_type` blocked with “use /lab-device/frames” | **code-aligned** |
| `POST /lab-imports` fixture only; other source_types 400 | spec L50; handler comment Phase 3 | `lab_result_import_service.go` commit allowlist `fixture` only | **code-aligned** (fixture saga, not daily device path) |
| `POST /lab-imports/preview` write-less; `drwan` 200 + `blocked_reasons` | spec L49 | `PreviewLabImport` always 200; drwan/manual/device populate `blocked_reasons` | **code-aligned** |
| `GET /lab-imports/:job_id` refuses `drwan` | spec L46 | `GetJob` `source_type=drwan cannot be opened` | **code-aligned** / **excluded** |
| Attach / detach / revert / events | spec L51–L55 | `routes_lab.go` attach/detach `lab-import:edit`; revert fixture compensating path | **code-aligned**. Device undo must not use fixture revert (ADR-007) |
| Default slots NX600 / AU10V / VetLab; PU-4010 decoder-only, not default agent slot | spec L67–L71, L82 | `labDeviceDefaultSlotsJSON` in `lab_device_receive.go`; FE `isLabDeviceBoardSlotSupported` only `fuji_nx600` and `fuji_au10v` | **code-aligned** for daily NX600/AU10V UI. VetLab slot exists in backend JSON but UI treats non-NX/AU as unsupported |
| Jouto NX600 / AU10V AE-LAB-0–4 done in spec narrative | spec L8 | Decoders and persist exist in medicalrecord | **code-aligned** for decoder/persist. **ops remainder**: real-device UAT sheet all 未実施 |
| PU-4010 7E1/8E1 unreviewed; agent default slot / ops support out | spec L69 | Decoder-only; agent must not guess 2400 8E1 (ADR-008) | **ops remainder** / not daily UAT |
| `idexx_vetlab` decoder/persist/default slot implemented; hospital always-on unapproved; `--pims-reply` opt-in only | spec L15, L70, L125–L129 | `lab-device-agent --pims-reply`; default read-only | **ops remainder**. Hospital VetLab cable **must not** enable `--pims-reply` |
| Consumer token shared between API env and agent; browser fetches then presents header; JWT not stored on agent | ADR-008 L25; `LAB_DEVICE_AGENT_MACOS.md` L15–L18 | See token section below | **code-aligned** locally if env set. **blocked** for STG/PROD Worker supply |
| Cloudflare / STG/PROD must not be called connected until token env path exists | `LAB_DEVICE_AGENT_MACOS.md` L18 | `backend/wrangler.jsonc` `secrets.required` has no `LAB_DEVICE_AGENT_CONSUMER_TOKEN`; `backend/worker/index.ts` `envVars` does not inject it; empty API env → GET agent-consumer **503** | **blocked** until USER implements/approves Worker secret + envVars allowlist |
| Developer ID sign / notarize / Gatekeeper | `LAB_DEVICE_AGENT_MACOS.md` L11 | Current bundle is client-UAT unsigned | **ops remainder** (release gate, not this sheet) |
| Real-device client UAT NX600/AU10V | `docs/ops/testing/scenarios/LAB_DEVICE_CLIENT_UAT.md` items 1–12 | Sheet records 未実施 for every row | **remainder**. This session does not execute it |
| Drワン / `mkan.mdb` | spec L6–L7, L141–L144 | preview blocked; commit blocked; GetJob 400 | **excluded** |

## Lab-import and lab-device API contract (current)

From `backend/internal/medicalrecord/routes_lab.go` and handlers. Daily device traffic uses the `/lab-device/*` group plus attach/detach. Fixture preview/commit remains a separate saga.

| Method | Path | Grant | Daily device path? | Notes |
| --- | --- | --- | --- | --- |
| `POST` | `/api/v1/lab-device/frames` | `lab-import:create` | yes | `{ payload_base64, device_hint }`; FE always `auto` |
| `PUT` / `DELETE` | `/api/v1/lab-device/wait` | create | yes | selected today-visit wait |
| `GET` | `/api/v1/lab-device/board` | create | yes | wait, unlinked, saved, received, today_visits, station |
| `GET` | `/api/v1/lab-device/agent-consumer` | create | yes | `{ agent_consumer_token }` from env; empty → 503 |
| `GET` | `/api/v1/lab-device/unlinked` | view | exam banner | not the board primary poll |
| `GET` / `PUT` | `/api/v1/lab-device/station` | view / edit | setup | slots JSON |
| `POST` | `/api/v1/lab-imports/:job_id/attach` | edit | yes | `{ pet_id }` only |
| `POST` | `/api/v1/lab-imports/:job_id/detach` | edit | yes | undo attach; not fixture revert |
| `POST` | `/api/v1/lab-imports/preview` | create | no | write-less; drwan 200 + blocked |
| `POST` | `/api/v1/lab-imports` | create | no | `source_type=fixture` only |
| `GET` | `/api/v1/lab-imports/:job_id` | view | inspect | drwan cannot be opened |
| `GET` | `/api/v1/lab-imports/:job_id/events` | view | audit | |
| `POST` | `/api/v1/lab-imports/:job_id/revert` | edit | fixture only | terminal persisted → reverted |

Agent loopback (not the Gin API): `GET /health` unauthenticated; `POST /claim`, `GET /frames`, `POST /frames/:id/ack|reject` require `X-Lab-Device-Consumer-Token` (claim also clinic match; frames/ack/reject also owner lease + `X-Clinic-ID` + `X-Lab-Device-Owner`). Host must be `127.0.0.1:17654` or `localhost:17654`. CORS/PNA for the bundled allowed origin.

## Consumer-token supply and distribution

Token is an operator-supplied shared secret. The API does not mint it. Bundlers do not generate it. Empty is fail-closed. **Do not put values in git, command history, logs, chat, or this sheet.**

### Supply (mint / register)

1. **USER** generates the secret on an approved path (password manager or clinic secret store). This sheet does not specify a generator command and does not run one.
2. **Same value** is registered in two places that never share a git file:
   - API process environment `LAB_DEVICE_AGENT_CONSUMER_TOKEN` (local Compose: backend `.env.local` — presence/value **UNKNOWN** here).
   - Agent install: 4th argument to `scripts/build-lab-device-agent-bundle.sh` **or** the same env name at bundle time (`scripts/build-lab-device-agent-bundle.sh` L19–L27). Empty → exit 2.
3. Bundle writes clinic id, allowed origin, token as three lines of `lab-device-agent.conf` mode `0600` (`scripts/build-lab-device-agent-bundle.sh` L52–L54). Install copies that into LaunchAgent `--consumer-token` (`packaging/macos/configure-lab-device-agent-plist.sh`, plist placeholders `__CONSUMER_TOKEN__`).
4. `backend/cmd/lab-device-agent/main.go` requires `--consumer-token` non-empty (exit 2).
5. **STG/PROD Cloudflare:** `backend/wrangler.jsonc` `secrets.required` (L143–L180) does not list `LAB_DEVICE_AGENT_CONSUMER_TOKEN`. `backend/worker/index.ts` `envVars` (L53–L111) does not inject it. Until USER adds secret + Worker allowlist + verification, GET agent-consumer stays 503. Do not treat STG `/lab-device` as connected.

### Browser-to-agent distribution

1. Staff with selected clinic and `lab-import:create` opens `/lab-device`.
2. `useGetLabDeviceAgentConsumer` (`frontend/src/features/lab-device/api/lab-device.ts`) `GET /v1/lab-device/agent-consumer` over the normal authenticated API (JWT stays on the API hop).
3. Handler `GetLabDeviceAgentConsumer` (`lab_device_agent_consumer_handler.go`) re-checks clinic + create grant, reads `os.Getenv("LAB_DEVICE_AGENT_CONSUMER_TOKEN")`, 503 if empty, else JSON `agent_consumer_token`.
4. `useLabDeviceAgentListen` builds `createLabDeviceAgentClient(token)` only in the browser process. Token is query-cached (`staleTime: Infinity`); it is not written to the agent disk.
5. Client `POST http://127.0.0.1:17654/claim` with `X-Lab-Device-Consumer-Token` and `{ clinic_id }`. Agent compares SHA-256 of the header to the LaunchAgent flag (`authorizeConsumerToken`). Clinic must match the installed `--clinic-id`. Single consumer lease; other tab/clinic → 409.
6. Subsequent `GET /frames` and ACK/reject send token + `X-Clinic-ID` + `X-Lab-Device-Owner`. `/health` has no token (diagnostics).
7. On API success, browser ACKs the agent frame; on 400, reject (raw stays in agent reject queue). Other API failures leave the frame pending for retry. Agent JWT storage is forbidden (ADR-008).

### Trust boundary (do not overclaim)

ADR-008 local trust: any process of the logged-in Mac user can read `/dev/cu.usbserial-*` without the HTTP token. The token prevents **wrong clinic tab / second consumer** mis-delivery, not same-OS-user compromise. OS account lock and malware controls remain operator duties.

### Distribution checklist for later USER ops (not executed here)

- Mint once; copy to local API env and to bundle/install; never commit.
- Allowed origin = exact deployed frontend origin (canonical https contract in `LAB_DEVICE_AGENT_MACOS.md`).
- Verify GET agent-consumer is 200 for a create-granted user **without** printing the body in tickets.
- Verify unsigned UAT bundle SHA-256 over a second channel before `install.sh`.
- STG/PROD: implement Worker secret + `envVars` before any connection-success claim.
- Rotate by replacing both sides together; mismatched sides fail closed.

## Remaining UAT and release gaps

Source of truth for client hardware checks: [`docs/ops/testing/scenarios/LAB_DEVICE_CLIENT_UAT.md`](../../ops/testing/scenarios/LAB_DEVICE_CLIENT_UAT.md). All twelve rows are **未実施**. This campaign unit does not change that.

| Gap | Status | Required input |
| --- | --- | --- |
| NX600/AU10V two-port LaunchAgent on the exam Mac (UAT 1, 8, 9) | remainder / clinic hardware **UNKNOWN** | Clinic Mac, two dedicated USB-serial adapters, bundle install |
| Live NX600 send matches board values (UAT 2, 12) | remainder | Device + known specimen; no raw payload in chat |
| Live AU10V send (UAT 3, 12) | remainder | same |
| Unlinked vs wait-pet attach (UAT 4–5) | remainder | Today-visit cards on that clinic |
| Duplicate fingerprint (UAT 6) | remainder | Device resend |
| Queue while page closed (UAT 7) | remainder | Memory queue; reboot loses pending (ADR-008) |
| Mixed NX600+AU10V (UAT 10) | remainder | both devices |
| `diagnose.sh` no raw/USB/patient leak (UAT 11) | remainder | run on clinic Mac |
| PU-4010 7E1/8E1 | not on this UAT; do not PASS | separate reviewed serial profile |
| IDEXX hospital always-on / `--pims-reply` | unapproved | do not enable on hospital VetLab cable |
| Cloudflare token supply | **blocked** | USER Worker secret + envVars + verification |
| Codesign / notarization | remainder | release gate |
| Local Compose token currently set? | **UNKNOWN** | do not read `.env.local` into this sheet |
| Linear LINMIG-182 | not found | claim_id only |

## Verification notes for this sheet

- Citations above are from the repo at HEAD `aac697645` (`feat/linmig-182-ops-prep`).
- No Docker app tests, no token mint, no serial, no Linear/GitHub writes.
- Coverage ratchet (`docs/ops/coverage-policy.md`) is a CI gate for code changes; this unit is docs-only and does not claim coverage PASS.

## Follow-ups (not this unit)

1. USER: Worker `LAB_DEVICE_AGENT_CONSUMER_TOKEN` secret + `envVars` allowlist, then STG read-only 503/200 check without logging the token.
2. USER: local or clinic UAT using `LAB_DEVICE_CLIENT_UAT.md` after a protected bundle + matching API env.
3. Separate packet for PU-4010 profile review and IDEXX hospital policy; keep `--pims-reply` off the hospital cable.
