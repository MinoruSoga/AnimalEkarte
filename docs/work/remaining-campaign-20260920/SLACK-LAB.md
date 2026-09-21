# SLACK-LAB — clinic × device correspondence

Campaign `remaining-ops-20260920` revision 1, unit `SLACK-LAB`, claim `claim/SLACK-LAB`.
Worktree `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-rem-slack-lab` on `feat/rem-slack-lab-20260920` at HEAD `873685b0b`. Sheet date: 2026-09-20.

Docs-only. This unit does **not** open serial ports, run `lab-device-agent`, send frames, mint consumer tokens, invent a コアグ protocol, or treat implemented receive code as hospital connected.

Binding: [todo-issue.md](../../../todo-issue.md) heading `### SLACK-LAB` (L282–286); [LAB_DEVICE_CONNECTIVITY.md](../../ops/deploy/LAB_DEVICE_CONNECTIVITY.md); [LINMIG-182.md](../linmig-campaign-20260919/LINMIG-182.md); [LabDeviceBoard.tsx](../../../frontend/src/features/lab-device/routes/LabDeviceBoard.tsx); [LAB_DEVICE_CLIENT_UAT.md](../../ops/testing/scenarios/LAB_DEVICE_CLIENT_UAT.md). Device-label photos are **not in this repo**.

This file is a campaign investigation sheet. Product modules do not import it. Callers are the campaign controller (`units.json` unit `SLACK-LAB` `owned_paths`; `unit-specs.json` in_scope) and operators reading `docs/work/remaining-campaign-20260920/`. Human pointers: [todo-issue.md](../../../todo-issue.md) L282–286 and [SLACK-INTAKE.md](../todo-campaign-20260919-ready17/SLACK-INTAKE.md) L50. Sibling contracts ([LAB_DEVICE_CONNECTIVITY.md](../../ops/deploy/LAB_DEVICE_CONNECTIVITY.md), [LINMIG-182.md](../linmig-campaign-20260919/LINMIG-182.md)) stay unedited. Manual urine is [SLACK-MANUAL-URINE](../todo-campaign-20260919-ready17/SLACK-MANUAL-URINE.md), not this sheet.

## 1. Why this sheet exists

[todo-issue.md](../../../todo-issue.md) L284–286: clinics differ by device; 敷島 alone mentioned 「コアグ」; 16 September asked whether/when receive would work. **Live receive is UNKNOWN.** First work is a clinic × device × send-format × support map. Separate 城東 existing adapters, decoder-only, unapproved daily ops, and unknown devices. 猫 / 八王子 model numbers and コアグ protocol stay undetermined without photos.

This sheet is that map. It is **not** a connection-success claim and **not** UAT PASS.

## 2. Status vocabulary (do not mix)

| Status | Means | Does not mean |
|---|---|---|
| **code-aligned** | Repo decoder / persist / slot / board path matches [LAB_DEVICE_CONNECTIVITY.md](../../ops/deploy/LAB_DEVICE_CONNECTIVITY.md) | That clinic's Mac is installed, that device is cabled, or UAT ran |
| **decoder-only** | Frames can be parsed; not a default agent slot; not daily ops support | Ready for hospital urine receive |
| **ops remainder** | Code exists; clinic hardware / LaunchAgent / token / UAT not observed | Connected |
| **unapproved daily** | Spec forbids treating this as always-on hospital use | Safe to enable `--pims-reply` on a hospital VetLab cable |
| **UNKNOWN** | Photos, model, protocol, or live receive not in repo / not observed | A guessed protocol or a promised date |
| **excluded** | Out of AnimalEkarte lab-device path | A hidden integration |

**Rule:** implemented receive code ≠ hospital connected. [LINMIG-182.md](../linmig-campaign-20260919/LINMIG-182.md) L50: clinic hardware, live LaunchAgent, and live token values are UNKNOWN without device I/O. [LAB_DEVICE_CLIENT_UAT.md](../../ops/testing/scenarios/LAB_DEVICE_CLIENT_UAT.md) rows 1–12 are all **未実施**.

## 3. Code facts (shared; not a clinic install)

Daily path ([LAB_DEVICE_CONNECTIVITY.md](../../ops/deploy/LAB_DEVICE_CONNECTIVITY.md) L8–L9, L41–L42; [LINMIG-182.md](../linmig-campaign-20260919/LINMIG-182.md) L20–L35): one exam Mac LaunchAgent `lab-device-agent` owns wired serial; loopback `127.0.0.1:17654`; browser `/lab-device` (`LabDeviceBoard`) polls with an API-issued consumer token and posts frames. The board does not open serial.

| Code surface | Fact | Citation |
|---|---|---|
| Receive entry | `ReceiveFrames` decodes then persist-or-duplicate per clinic | [lab_device_receive_service.go](../../../backend/internal/medicalrecord/lab_device_receive_service.go) L54–L75 |
| Board | Permission `lab-import`; today-visit cards; day-grouped received list; no confirm dialog; no pet search | [LabDeviceBoard.tsx](../../../frontend/src/features/lab-device/routes/LabDeviceBoard.tsx) L42–L87; spec L41 |
| Default station slots | NX600, AU10V, VetLab. **PU-4010 is not in this JSON** | [lab_device_receive.go](../../../backend/internal/medicalrecord/lab_device_receive.go) L24 `labDeviceDefaultSlotsJSON` |
| Board UI listen | `isLabDeviceBoardSlotSupported` is **only** `fuji_nx600` and `fuji_au10v`. VetLab / PU-4010 / other → `unsupported` | [lab-device-board-model.ts](../../../frontend/src/features/lab-device/lib/lab-device-board-model.ts) L194–L196; [LabDeviceBoardPanels.tsx](../../../frontend/src/features/lab-device/routes/LabDeviceBoardPanels.tsx) L79 |
| Client UAT scope | NX600 and AU10V only. PU-4010 and IDEXX are not sent on that sheet | [LAB_DEVICE_CLIENT_UAT.md](../../ops/testing/scenarios/LAB_DEVICE_CLIENT_UAT.md) L5–L6, L49 |
| Drワン / MDB | Closed. `source_type=drwan` preview blocked; commit blocked | spec L6–L7, L141–L144 |

STG/PROD consumer-token Worker supply remains **blocked** until USER adds the secret + `envVars` ([LINMIG-182.md](../linmig-campaign-20260919/LINMIG-182.md) L69, L109). Do not call STG `/lab-device` connected.

## 4. Clinic × device map

Clinic labels below are the same catalog seed labels used on sibling sheets (ids 1–4). Live `clinic_id` occupancy on STG/PROD is **UNKNOWN**. Device rows that are not 城東-confirmed stay UNKNOWN. This table does **not** claim all clinics connected.

| Clinic (catalog id / label) | Device / send form | Code | Board daily UI | Hospital install / live receive |
|---|---|---|---|---|
| 2 / 城東センター病院 | 富士 DRI-CHEM NX600 · COM6 · 9600 8N1 · `fuji_nx600` · STX…ETX | **code-aligned** decoder + persist + default slot. Spec: AE-LAB-0〜4 for 城東 3種 narrative ([LAB_DEVICE_CONNECTIVITY.md](../../ops/deploy/LAB_DEVICE_CONNECTIVITY.md) L8, L67) | **supported** (`fuji_nx600`) | **UNKNOWN / ops remainder.** UAT 1–12 **未実施**. Mac + two USB-serial + bundle **not observed** |
| 2 / 城東センター病院 | 富士 DRI-CHEM IMMUNO AU10V · COM7 · 9600 8N1 · `fuji_au10v` | **code-aligned** decoder + persist + default slot (spec L68) | **supported** (`fuji_au10v`) | **UNKNOWN / ops remainder.** UAT **未実施**. Same as NX600: code ≠ cabled |
| 2 / 城東センター病院 | アークレイ PU-4010 · COM3 · 現場 2400 8E1 (`mdcon2` 2400 7E1) · `arkray_pu4010` | **decoder-only.** Not in `labDeviceDefaultSlotsJSON`. 7E1/8E1 unreviewed (spec L69, L82) | **unsupported** | **UNKNOWN.** Agent default slot / ops support **out**. Not daily UAT. Do not PASS this sheet for urine receive |
| 2 / 城東センター病院 | IDEXX VetLab Station PIMS (ProCyte / Catalyst behind it) · COM5 · 9600 8N1 · `idexx_vetlab` | **code-aligned** decoder + persist + default slot. Hospital always-on **unapproved**. Receive-only drops PIMS (spec L15, L70, L125–L129) | **unsupported** (not NX/AU) | **unapproved daily.** Do not enable `--pims-reply` on a hospital VetLab cable. Ethernet to analyzer bodies **forbidden** |
| 2 / 城東センター病院 | 富士 DRI-CHEM 7000V · COM4 | **excluded** — 「触らない」 (spec L63, L71, L145) | not a slot | Do not connect |
| 2 / 城東センター病院 | Drワン / `mkan.mdb` | **excluded** | n/a | Not an AnimalEkarte input |
| 1 / 八王子病院 | 尿は機器測定 (Slack 212–294 via [SLACK-MANUAL-URINE](../todo-campaign-20260919-ready17/SLACK-MANUAL-URINE.md) L15) | No 八王子 model in [LAB_DEVICE_CONNECTIVITY.md](../../ops/deploy/LAB_DEVICE_CONNECTIVITY.md) | n/a until model known | **UNKNOWN.** Photos not in repo. Do not assume PU-4010 |
| 1 / 八王子病院 | Other analyzers | none in spec | n/a | **UNKNOWN** |
| 3 / ノア動物病院　敷島病院 | 「コアグ」 (Slack; 敷島のみ) | **no protocol in repo** | n/a | **UNKNOWN.** Do not invent コアグ framing, baud, or `source_type` |
| 3 / ノア動物病院　敷島病院 | 尿試験紙は目視 | not lab-device | n/a | Hand-entry path only ([SLACK-MANUAL-URINE](../todo-campaign-20260919-ready17/SLACK-MANUAL-URINE.md)). Not this unit |
| 4 / ノア動物病院　Hako bu neco（猫） | 尿試験紙は目視 | not lab-device | n/a | Hand-entry. Device models **UNKNOWN** (todo-issue L285) |
| 4 / ノア動物病院　Hako bu neco（猫） | Other analyzers | none in spec | n/a | **UNKNOWN.** Photos not in repo |
| Any clinic **not** in 1–4 | unknown devices | none | n/a | **UNKNOWN.** Do not mark connected |

城東 NX600/AU10V are the **only** rows that are both code-aligned **and** board-supported. That still does not make 城東 connected: [LINMIG-182.md](../linmig-campaign-20260919/LINMIG-182.md) L65–L66, L133–L146.

## 5. コアグ (UNKNOWN — stop)

[todo-issue.md](../../../todo-issue.md) L284: 敷島のみの「コアグ」. No source_type, baud, frame shape, or vendor string exists in [LAB_DEVICE_CONNECTIVITY.md](../../ops/deploy/LAB_DEVICE_CONNECTIVITY.md), [LINMIG-182.md](../linmig-campaign-20260919/LINMIG-182.md), or `LabDeviceBoard` slot support.

This sheet **does not** invent a コアグ protocol, map it onto `fuji_*` / `arkray_pu4010` / `idexx_vetlab`, or promise a receive date. Required input: photos / vendor docs + PO whether it is in-scope for `/lab-device`. Until then: **BLOCKED** for 敷島 コアグ.

## 6. Stop conditions

- No serial capture, no device I/O, no token values in this sheet.
- Do not answer 16 September 「いつ受信できるか」 from decoder existence.
- PU-4010 7E1/8E1 and IDEXX hospital always-on stay out of daily UAT.
- 猫 / 八王子 model numbers remain UNKNOWN without photos.
- Manual urine stays on SLACK-MANUAL-URINE.
- Missing Worker `LAB_DEVICE_AGENT_CONSUMER_TOKEN` → STG/PROD board is not connected ([LINMIG-182.md](../linmig-campaign-20260919/LINMIG-182.md) L109).

## 7. Follow-ups (not this unit)

1. USER: clinic exam-Mac bundle + matching API env, then [LAB_DEVICE_CLIENT_UAT.md](../../ops/testing/scenarios/LAB_DEVICE_CLIENT_UAT.md) for 城東 NX600/AU10V only.
2. USER: Worker consumer-token secret + `envVars` before any STG connection claim.
3. Photos of 八王子 / 猫 / 敷島 コアグ labels; then a **new** packet if a protocol is identified. Do not extend this sheet with invented framing.
4. Separate reviewed PU-4010 serial profile; keep `--pims-reply` off the hospital VetLab cable.
