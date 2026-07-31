# TASK-019 — `docs/spec/line/**` deep audit vs BE/FE

> **Date**: 2026-07-31  
> **Unit**: `TODO-MD-TASK-019-LINE-DEEP-AUDIT-20260731`  
> **Prior**: `reports/2026-07-31-task-019-line-audit.md`（high-level inventory only）  
> **Mode**: Deep residual disposition. No production LINE/Messaging API calls. No secrets.

---

## 1. Purpose / Non-actions

Deep-pass every material claim in `docs/spec/line/**` (and related LINE/LIFF/L-step screen specs) against current code. Disposition each residual as docs honesty / BUG / 要PO / ops / 残差なし.

**Non-actions (explicit):**
- No production webhook registration or live Messaging API / L-step write
- No secret / channel token / real LINE user ID inspection or paste
- No product feature invention (including `available-staffs` — WONTFILE)
- No migrate/seed apply, push/PR, claim delete

---

## 2. Documentation inventory

| Path | Role |
|:---|:---|
| `docs/spec/line/README.md` | Hub index + admin quick links |
| `docs/spec/line/architecture.md` | Inbound/Outbound split, identity, failure contracts |
| `docs/spec/line/reservation-spec.md` | Owner LIFF reservation product summary |
| `docs/spec/line/setup.md` | Console / Messaging / L-step ops setup |
| `docs/spec/line/lstep-integration.md` | CPM/VISIT + 15 delivery triggers + safety |
| `docs/spec/line/cost-analysis.md` | Cost notes (docs-only product surface) |
| `docs/spec/line/CLAUDE.md` | Agent-local notes (not product SoT) |

**Related screens:**

| Path | Role |
|:---|:---|
| `docs/spec/screens/28-line-reservation.md` | Clinic LINE reservation settings UI |
| `docs/spec/screens/37-line-reserve-owner-flow.md` | Owner multi-step reserve flow |
| `docs/spec/screens/38-liff-pet-health.md` | LIFF pet health + link |
| `docs/spec/screens/31-lstep-integration.md` | L-step admin settings/tags/sync |
| `docs/spec/screens/34-lstep-delivery-monitor.md` | Delivery trigger monitor |

**Ops cross-ref:** `docs/ops/deploy/LSTEP_WRITE_API_PAUSE.md`（Write dual-gate 正本）

---

## 3. Code surface map

### 3.1 Backend — `backend/internal/lstep/` (~260 Go files)

| Surface | Evidence | Notes |
|:---|:---|:---|
| Webhook | `POST /api/line/webhook` → `ReceiveLineWebhook` | `routes_snapshot_test.go` L117; public, signature routing + body limits + rate tests |
| Link token | `POST /api/v1/owners/:id/line/link-token` | Staff permission; raw once / SHA-256 base64url digest in DB |
| LIFF bind (injected) | `POST /api/liff/:clinicId/link` | Wired from reservation `RegisterLiffRoutes`; 10/min |
| Send / logs | `POST .../line/send`, `GET .../line/send-logs` (+ `/lstep/send*` aliases) | Permissioned |
| Line customers | `GET .../line-customers`, `PATCH .../link-owner` | Staff link path |
| L-step settings | `GET/PATCH .../lstep-settings`, `POST .../test-connection` | Dual LINE + L-step connection tests exist in service tests |
| Delivery monitor | `GET .../lstep/delivery-monitor/{logs,summary}` | Observes `lstep_delivery_trigger_log` only |
| Trigger priorities | `GET/PATCH .../lstep/trigger-priorities` | Defaults in `model/lstep_trigger_priority.go` |
| Checkup sync / tags / analytics / CSV | Additional `internal/lstep` handlers | Matches screen 31 |
| Credentials | AES-256-GCM via `INTEGRATION_ENCRYPTION_KEY` | `line_credentials.go` |
| Write dual-gate | `LSTEP_WRITE_API_ENABLED` + clinic `is_sync_enabled` | HTTP 0 + `ErrWriteDisabled` when deploy gate off (`infra/lstep`) |
| Delivery batch | 10:00 JST fixed | `lstep_batch_service.go` `deliveryTriggerHourJST = 10` |

### 3.2 Backend — `backend/internal/reservation/` (LIFF + line reservation settings)

| Surface | Evidence |
|:---|:---|
| LIFF public API | `RegisterLiffRoutes`: `/api/liff/:clinicId/{settings,profile,courses,trimming-*,staffs,available-dates,available-times,reservations,my-reservations,health-card,link}` |
| Write owner | LIFF create uses reservation validators; `source=line` (`model.ReservationSourceLine`) |
| Settings | `GET/PUT .../line-reservation-settings` |
| No name+phone auto owner link | `liff_service_reservations.go` SEC-CS2-F02 comment + regression test |

### 3.3 Frontend

| Surface | Location | Admin/app path |
|:---|:---|:---|
| Clinic LINE reservation | `frontend/src/features/line-reservation` | `/line-reservation/{settings,page-editor,slots}` |
| L-step admin | `frontend/src/features/lstep` | `/settings/integrations/lstep`, `/settings/lstep/tags`, `/lstep/{checkup-sync,analytics,delivery-monitor}` |
| Owner reserve | `frontend/line-reserve/` | Multi-step SPA (Top→…→Complete / MyReservations) |
| Pet health / link | `frontend/liff/` | PetHealthPage / LiffLinkPage |
| Shared LIFF | `frontend/src/shared-liff/` | useLiff, Spinner, ErrorPage |
| Trigger labels | `frontend/src/config/lstep-trigger-types.ts` | Includes legacy history labels beyond current BE set |

### 3.4 Trigger codes (15) — doc ↔ code

Doc table (`lstep-integration.md` §3) matches `model.AllTriggerTypes()` + default priorities (`lstep_trigger_priority.go`) + batch registry (`lstep_batch_delivery.go`):

`first_visit_followup_3d/7d`, `next_visit_reminder`, `vaccine_deadline_60d/30d`, `birthday_message`, `dormant_prevention_180d/210d/240d/365d`, `filaria_alert`, `flea_tick_alert`, `food_refill_reminder`, `first_visit_welcome`, `checkup_followup`.

FE labels keep legacy `dormant_prevention_120d/220d` and `supp_refill_reminder` for **history display only** — not current BE `AllTriggerTypes`.

---

## 4. Residual disposition table

| ID | Claim / drift | Doc path | Code evidence | Disposition |
|:---|:---|:---|:---|:---|
| D-01 | reservation-spec 飼主フローが「コース→スタッフ→日時→ペット→確定」の粗5段で、実装はお客様情報→コース(±トリミング)→スタッフ→日付→時間→要望→確認→完了 | `docs/spec/line/reservation-spec.md` | `frontend/line-reserve/src/App.tsx`, screen 37 | **docs fix done**（要約を実装/37準拠に更新） |
| D-02 | screen37「氏名+電話でオーナー自動紐付け」— コードは SEC-CS2-F02 で禁止 | `docs/spec/screens/37-line-reserve-owner-flow.md` | `liff_service_reservations.go:53-55`, `liff_service_reservations_test.go` SEC-CS2-F02 | **docs fix done** |
| D-03 | architecture §2「電話番号または飼主Noで名寄せ」— 主経路は staff link-token + LIFF `POST /link` / staff `link-owner`。予約時 name+phone 自動紐付けは無い | `docs/spec/line/architecture.md` | `line_link_handler.go`, `reservation/routes.go` L171, SEC-CS2-F02 | **docs fix done**（主経路を明記） |
| D-04 | architecture/lstep「Write API noop」表現 — 実装は silent nil 成功ではなく deploy gate で HTTP 0 + `ErrWriteDisabled`（ops 正本 `LSTEP_WRITE_API_PAUSE.md`） | architecture, lstep-integration | `infra/lstep/errors.go`, `client.go` | **docs fix done**（用語を dual-gate / ErrWriteDisabled に寄せる） |
| D-05 | README クイックリンクに delivery-monitor / cost-analysis 未掲載 | `docs/spec/line/README.md` | FE path `/lstep/delivery-monitor` | **docs fix done** |
| D-06 | setup が接続テスト 2 ボタン＝別 API と読めるが、FE は同一 `test-connection` を呼ぶ | `docs/spec/line/setup.md` | `use-lstep-settings.ts` L73–81; `postLstepTest` | **docs fix done** |
| R-01 | architecture に webhook イベント契約表が無い。実装は **follow/unfollow のみ**処理し他 type は skip。signature は destination→`line_bot_user_id` O(1) HMAC | architecture | `line_link_service.go` HandleWebhook; `line_link_signature_routing_test.go` | **要PO** — イベント契約を architecture に書くかコード+ops 正本のままか |
| R-02 | Webhook / 署名 / 本番チャネル疎通・`line_bot_user_id` プロビジョニング手順は docs だけでは完結しない | setup | FindByLineBotUserID path | **ops** — 本番操作は USER 専権（本パケット非対象） |
| R-03 | S04/S12/V05 等【要実測】観測マーク | prior audit + screens | runtime not in this packet | **TASK-010**（env/runtime residual） |
| R-04 | Write API 再有効化判断・実送信検証 | LSTEP_WRITE_API_PAUSE | deploy gate default off | **ops / USER** |
| R-05 | LINE Channel Secret の **二重ストア**: `clinic_integrations`（L-step settings 更新）と `line_reservation_settings`（webhook HMAC / LIFF settings）。SoT と dual-write 方針が未決 | setup / architecture / screen28 | `lstep_settings_update.go`; reservation line_reservation_setting; `verifySignatureAnyClinic` | **要PO** — 製品/運用の source-of-truth 決定が必要。本パケットでは発明しない |
| R-06 | `/lstep/delivery-monitor` は route 実装済だが `paths.lstep` とサイドバー Lステップ連携に未登録（deep-link のみ） | screen34 / README | `operations-routes.tsx`; `paths.ts`; `sidebar-menu.tsx` | **要PO** — intentional deep-link か nav 漏れか |
| R-07 | タグ管理: route guard は `ResourceLstepAnalytics`、sidebar は `ResourceHospitalSettings` で nav と到達可否が乖離し得る | screen31 | `settings-routes.tsx` vs `sidebar-menu.tsx` | **要PO** / 将来 BUG 候補（RBAC 整合。本 deep ではコード変更しない） |
| R-08 | pet-health LIFF は `VITE_LIFF_ID`、line-reserve は clinic `settings.liff_id` — 二重 bootstrap | screen38 | `frontend/liff/src/lib/liff-config.ts`; line-reserve App.tsx | **ops** — deploy で ID 一致を保証 |
| V-01 | 15 配信トリガー + 既定優先度 | lstep-integration §3 | `AllTriggerTypes`, priority defaults, batch list | **残差なし** |
| V-02 | CPM V2 閾値コメント（0–1…13+）とタグ名 | lstep-integration §2 | `lstep_tag_sync_service.go` CPMStageV2* | **残差なし** |
| V-03 | VISIT_120/180/220/240 | lstep-integration | same package constants + tests | **残差なし** |
| V-04 | 配信バッチ 10:00 JST | architecture §4.1 | `deliveryTriggerHourJST = 10` | **残差なし** |
| V-05 | 配信監視 URL `/lstep/delivery-monitor` と trigger-log 専用 | lstep-integration / screen34 | FE routes + BE delivery-monitor routes | **残差なし** |
| V-06 | LIFF API 表（settings…my-reservations） | screen 37 | `RegisterLiffRoutes` + liff_handler_test | **残差なし** |
| V-07 | source=line | reservation-spec | `ReservationSourceLine` | **残差なし** |
| V-08 | 変更機能・前日リマインド未実装 | reservation-spec §5 | line-reserve に reschedule 無し; cancel/create のみ | **残差なし** |
| V-09 | AES-256-GCM credentials | architecture §5 | `line_credentials.go` | **残差なし** |
| V-10 | is_sync_enabled + Write dual-gate 停止手段 | architecture §4.2 | LSTEP_WRITE_API_PAUSE + buildClient patterns | **残差なし**（用語は D-04 で修正） |
| V-11 | cost-analysis 製品コード drift | cost-analysis | n/a product | **残差なし**（docs-only） |
| V-12 | available-staffs | n/a | `available_staffs_ban_test.go` WONTFILE | **残差なし**（発明禁止） |
| V-13 | Admin quick links 主要4本 | README | `frontend/src/config/paths.ts` | **残差なし**（monitor は D-05 で追加） |
| V-14 | screen28 Channel Secret を line-reservation UI で扱わない | screen28 | form design + setup は L-step 連携設定側 | **残差なし**（画面分担） |
| V-15 | link token raw 一度 / digest 保存 / force 拒否 | screen31 | line_link_service tests | **残差なし**（静的） |
| V-16 | Pet health + link branch | screen38 | `frontend/liff` App.tsx, health-card API | **残差なし**（静的） |

---

## 5. Docs honesty edits applied (this packet)

Allowlisted paths only:

1. `docs/spec/line/reservation-spec.md` — 飼主フローを実装/screen37 準拠の要約に更新  
2. `docs/spec/screens/37-line-reserve-owner-flow.md` — name+phone 自動オーナー紐付け記述を削除し、SEC-CS2-F02 を反映  
3. `docs/spec/line/architecture.md` — Identity Mapping 主経路 + Write dual-gate 用語  
4. `docs/spec/line/lstep-integration.md` — Write pause 用語を dual-gate に整合  
5. `docs/spec/line/README.md` — delivery-monitor / cost-analysis を索引へ  
6. `docs/spec/line/setup.md` — 接続テスト 2 ボタンが同一 API である旨を明記  

No product code. No new features.

---

## 6. Sample claim verification (rg / file evidence)

1. **Webhook route** — `backend/internal/lstep/routes_snapshot_test.go`: `POST /api/line/webhook ReceiveLineWebhook`  
2. **15 triggers** — `backend/internal/model/lstep_delivery_trigger_log.go` `AllTriggerTypes()` length 15  
3. **No name+phone auto-link** — `backend/internal/reservation/liff_service_reservations.go` SEC-CS2-F02  
4. **10:00 JST** — `backend/internal/lstep/lstep_batch_service.go` `deliveryTriggerHourJST = 10`  
5. **LIFF staffs (not available-staffs)** — `GET /api/liff/:clinicId/staffs` in `liff_handler_test.go`; WONTFILE ban in `staff/available_staffs_ban_test.go`

---

## 7. Orchestration notes

- Mode: native Workflow `task-019-line-deep-audit-ro` (parallel RO: ro-docs / ro-be / ro-fe) + parent single writer  
- Workflow agents: `ro-docs` / `ro-be` / `ro-fe` all **done** (joined); parent synthesized report + honesty edits  
- Parent also performed independent code mapping (routes snapshot, model triggers, line-reserve App, FE paths) and reconciled probe residuals R-05..R-08  
- Secrets scan: report contains no tokens/channel secrets; env var **names** only (`LSTEP_WRITE_API_ENABLED`, `INTEGRATION_ENCRYPTION_KEY`, `VITE_LIFF_ID`)

---

## 8. Outcome for TASK-019

| Acceptance | Result |
|:---|:---|
| Deep report filed with disposition table | **Yes** (this file) |
| Material claims verified or dispositioned | **Yes** |
| Silent open drifts | **None** — residuals have IDs |
| Optional honesty docs | **Applied** (allowlist) |
| SPEC-TOP-LINE-AUDIT | **Closed for deep pass** — residuals only R-01..R-08（要PO/ops/TASK-010）。verified 残差なし V-01..V-16；docs honesty D-01..D-06 done |

---

## 9. USER next (out of this packet)

- TASK-009 seed apply (USER)  
- TASK-010 / TASK-020 runtime (env)  
- TASK-021 Stage A destructive deletion (inventory + approval)  
- R-01 要PO if architecture should host webhook event contract table  
- R-02/R-04 ops for live webhook/write re-enable  
- Commit dirty tree when ready (agent does not push)
