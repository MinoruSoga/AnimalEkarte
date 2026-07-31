# LINE residual R-01 / R-05 / R-06 / R-07 — PO 最終決裁

> **決裁日**: 2026-07-31
> **Unit**: `TODO-MD-LINE-RESIDUAL-PO-GPT56-20260731`
> **決裁者**: PO (gpt-5.6 Codex session 2026-07-31)
> **権限の前提**: 人間の個人名は入力されていないため、USER から本セッションへ委任された PO 権限を上記名義で記録する。
> **Runtime note**: セッションの実モデル名を独立に証明する runtime attestation は tool surface に無い。決裁者名は saved prompt が指定した委任 PO 名義であり、Parity Mode や benchmark manifest の主張ではない。
> **効力**: 本書は recommendation ではなく binding decision。R-01 / R-05 / R-06 / R-07 の再度の PO 確認なしに、下記 follow-up の実装計画へ進める。
> **Evidence boundary**: current HEAD `cf1e08113dbc0a9643a7ef64f94532cff707d957` の静的 evidence。R-06 / R-07 は landed configuration を確認したが、本 unit では Docker runtime test を実行していない。
> **非実施**: 製品コード、migration、seed、OpenAPI、FE/BE、DB、LINE 本番設定を変更しない。credential 値を読取・記録しない。

## 決裁サマリー

| ID | Chosen | Binding disposition | Follow-up |
|:---|:---|:---|:---|
| **R-01** | **B — CODE_TESTS_SOT + ARCHITECTURE_SUMMARY** | webhook event の executable SoT は code/tests。architecture は現行境界の要約と正本リンクを持つが、詳細契約表を二重管理しない | docs summary + event/signature contract test の明示 |
| **R-05** | **A-CI — SINGLE_SOT_CLINIC_INTEGRATIONS** | Channel Secret の唯一の SoT / write owner は `clinic_integrations`。`line_reservation_settings` から credential write/read を段階撤去し、恒久 dual-write を禁止する | High: mismatch-safe migration + verifier cutover +旧 surface 撤去 |
| **R-06** | **A — CLOSE_ORIGINAL_AS_HONESTY + EFFECTIVE_NAV_OPEN** | `deliveryMonitor` path と sidebar「配信監視」を追認し、intentional deep-link 方針へ戻さない。元の path / child item 欠落 residual は close | High: sidebar parent + `/lstep` route wrapper の権限 matrix 修正 |
| **R-07** | **A — CLOSE_ROUTE_AUTHORITY + EFFECTIVE_NAV_OPEN** | タグ管理の product RBAC 正本は route の `ResourceLstepAnalytics`。landed child resource を追認し、元の child mismatch は close | High: 親 gate + BE/API permission 突合 |
| **TASK-021 Phase2** | **PHASE2_START_APPROVED / CLEAN_GO_DROP_HOLD** | §6.1–§6.3 の consumer migration / BE / OpenAPI wave は別 packet で着手可 | DB・seed・DROP・migrate は別承認まで HOLD |

## 共通判断原則

- ①要件を疑う → ②削除 → ③簡素化 → ④cycle time → ⑤自動化、の順序を逆行しない（`docs/product-philosophy.md:13-24`）。
- 同じ business fact に二つの独立 write owner を置かない。二重管理を「同期」で恒久化しない（同 `:52-70`, `:155-165`）。
- 文書の product boundary と executable code/test の役割を分離し、同じ詳細表を複製しない。
- route authorization を nav 表示より弱くしない。sidebar parent または route wrapper が、許可済み child に未宣言の追加権限を課す状態を intentional policy としない。
- credential 欠落・不一致・復号失敗・DB failure を別値 fallback や全 clinic 試行で救済せず、fail-closed にする。
- landed source と runtime-green を同義にしない。deterministic test が未実行なら runtime 完了とは記録しない。

## Binding decisions

### R-01 — Webhook イベント契約の文書 SoT

- **Chosen**: **B — code/tests を executable SoT とし、architecture は要約 + link**
- **Owner**: PO (gpt-5.6 Codex session 2026-07-31) / 業務責任ロール: LINE 連携・運用 PO
- **Purpose**: 現在処理する webhook event と署名 failure boundary を、実装と同時に壊れる regression test で固定しつつ、architecture 読者が「全 event 対応」と誤認しない導線を一箇所で提供する。
- **Binding policy**:
  - 処理 event set、unsupported event の挙動、署名 routing、failure contract の executable SoT は `line_link_service.go` と、その contract test とする。
  - 現行 contract は `follow` / `unfollow` のみ business 処理し、他 event type は business side effect なしで skip する。新しい event type、message reply、配信方針を本決裁から発明しない。
  - 署名検証は event 処理より先に行い、webhook `destination` から対象 clinic を一意に絞る。全 clinic secret の走査、複数候補 HMAC、destination 欠落時の推測を禁止する。
  - `docs/spec/line/architecture.md` は上記を短く要約し、executable contract test と `setup` / ops へ link する。event ごとの詳細表は置かない。
  - `setup` / ops は provisioning・疎通・停止手順の正本であり、処理 event の product contract を独自に増減しない。
- **Risks accepted**:
  - architecture 単体では event ごとの全詳細を読めず、code/test link を辿る必要がある。
  - test 名や link が drift すると onboarding 性が落ちる。
  - unsupported event の成功 skip は「対応済み」と誤読され得るため、business side effect が無いことを明記する必要がある。
- **Follow-up**:
  - **Docs + contract-test TASK**: `architecture.md` に「処理対象は follow/unfollow、他 type は副作用なしで skip、署名不成立は fail-closed」という要約と executable test への相対 link を追加する。詳細な event 表は作らない。
  - 受け入れは、table-driven test が少なくとも `follow`、`unfollow`、unsupported type、destination 欠落、invalid signature、setting not found / DB failure を区別し、unsupported type の side effect がゼロであること、署名 path が `FindAll` や複数 secret 試行へ退行しないことを証明すること。event set を変更する PR は code・contract test・architecture summary を同時更新し、docs link check を通すこと。
- **Rationale (vs product philosophy)**: A の詳細表は code/test と同じ event list を二重管理する。B は不要な複製を削除したうえで、読者に必要な最小要約と正本導線を残す。
- **Overrides brief recommendation?**: **Yes**。brief の A（architecture 最小表）ではなく B を採用する。critic が指摘した duplicate contract drift を避けるため。
- **Evidence**: `reports/2026-07-31-line-residuals-R01-R07.md:38-62`; deep-audit residual table line 107; `backend/internal/lstep/line_link_service.go:269-307,365-407`。

### R-05 — LINE Channel Secret 二重ストア SoT

- **Chosen**: **A-CI — `clinic_integrations` を唯一の credential SoT とする**
- **Owner**: PO (gpt-5.6 Codex session 2026-07-31) / 業務責任ロール: LINE 連携・security PO
- **Purpose**: Channel Secret の更新・rotation を一つの write surface に集約し、二つの値が drift して webhook 署名が停止する failure mode を削除する。
- **Binding policy**:
  - Channel Secret の唯一の SoT / write owner は `clinic_integrations` の LINE/L-step integration credential とする。正規の更新 UI/API は L-step 連携設定である。
  - `line_reservation_settings` は予約設定と webhook destination routing metadata を保持してよいが、Channel Secret の request 受入、write、read、fallback owner にはしない。
  - Option C の恒久 atomic dual-write を採用しない。移行のための一時比較・backfill は、二つを同格 SoT とする契約ではない。
  - current `destination` → `line_bot_user_id` lookup で clinic を一件に絞った後、canonical credential を clinic scope で取得し、最大一回の decrypt / HMAC で検証する。全 clinic scan、旧値 fallback、複数 secret 試行を禁止する。
  - canonical credential が欠落する、不一致 inventory が未解決、DB/decrypt が失敗する場合は署名不成立として fail-closed にする。credential 値・可逆値・比較用 digest を report / log / error に出さない。
  - numbered migration が必要でも agent は apply しない。適用は USER の `make migrate` とする。
- **Risks accepted**:
  - 切替前 inventory で store 間の不一致 clinic が見つかり、手動再設定まで webhook を HOLD する可能性がある。
  - verifier は destination routing 後に canonical credential を取得する固定 cost が増える。
  - 旧 reservation settings API を直接使う未確認 consumer があれば、deprecation / migration が必要になる。
  - cutover 不備は webhook 全停止につながるため、段階移行と rollback 条件が必要である。
- **Follow-up**:
  - **High design / implementation TASK — single-SoT cutover**:
    1. clinic ごとに両 store の `empty / equal / mismatch` 状態だけを inventory し、値・hash・暗号文を artifact に残さない。
    2. canonical 側が空で旧側だけに値がある場合も自動採用を既定にせず、移行規則と actor/audit を設計する。mismatch は推測で winner を選ばず、当該 clinic を HOLD して正規 UI から再設定する。
    3. verifier を `destination → clinic → clinic_integrations credential → 最大1 HMAC` へ切り替え、旧 column fallback を入れない。
    4. reservation settings request/service/repository から Channel Secret write/read を撤去し、consumer inventory 後に旧 column を numbered migration で削除する。
  - 受け入れは、正規 UI の一回の rotation が canonical row だけを更新し、新値の正しい署名が通り旧値の署名が拒否されること、secret 欠落・mismatch・decrypt/DB failure が fail-closed であること、他 clinic の destination/credential を組み合わせられないこと、検証が全 clinic scan や複数 HMAC に退行しないこと、response・log・audit に credential 値が出ないことを unit/integration test で証明すること。旧 request field・repository write・verifier read が source inventory でゼロになってから旧 column 削除を提案できる。
- **Rationale (vs product philosophy)**: 短期 B の「役割を文書化して二重 store を残す」は二重管理を固定する。A を binding target とし、安全な段階移行を実装詳細として分離する。
- **Overrides brief recommendation?**: **Yes**。brief の短期 B → 中期 A ではなく、A を直ちに binding product law とする。実装は段階化するが、B を正状態とは呼ばない。
- **Evidence**: `reports/2026-07-31-line-residuals-R01-R07.md:66-91`; `backend/internal/lstep/lstep_settings_update.go:12-43`; `backend/internal/lstep/lstep_settings_service.go:295-328`; `backend/internal/reservation/line_reservation_setting_service.go:117-205`; `backend/internal/reservation/line_reservation_setting_repository.go:66-126`; `backend/internal/lstep/line_link_service.go:365-407`。

### R-06 — delivery-monitor navigation

- **Chosen**: **A — CLOSE_ORIGINAL_AS_HONESTY + EFFECTIVE_NAV_OPEN**
- **Owner**: PO (gpt-5.6 Codex session 2026-07-31) / 業務責任ロール: L-step 運用 PO
- **Purpose**: 配信 failure を監視する既存画面を、権限を持つ運用者が共有 URL に依存せず発見できる状態にする。
- **Binding policy**:
  - `paths.lstep.deliveryMonitor` と sidebar「配信監視」を保持する。intentional deep-link-only 方針へ戻さない。
  - delivery-monitor の chosen product permission は child route / child menu の `ResourceLstepAnalytics` 単独とする。
  - current route は `/lstep` wrapper の `ResourceHospitalSettings` と child の `ResourceLstepAnalytics` が重なり、実効的には両方を要求する。current behavior を chosen policy と同一視しない。
  - path / child item 欠落という original R-06 は landed configuration により close する。
  - ただし sidebar parent と `/lstep` route wrapper の二つの `ResourceHospitalSettings` gate が、Analytics-only の route-authorized child を隠す・拒否する現状を intentional としない。これは新しい effective-nav / route-auth follow-up として OPEN にする。
  - source landing を runtime-green と呼ばない。rendered permission matrix と page mount の deterministic test が PASS するまで end-to-end nav honesty は未完了である。
- **Risks accepted**:
  - 配信監視を analytics 権限者へ発見可能にするため、deep-link-only より露出範囲が広がる。
  - current parent gates により Analytics-only role は menu を見られず、`/lstep` wrapper でも route を拒否される。
  - landed report では Docker test 未実行で、実 browser behavior は本決裁の証拠に含まれない。
- **Follow-up**:
  - **High FE RBAC TASK — L-step sidebar + route parent honesty**: sidebar parent を独立した強い authorization gate とせず、許可された child が一件でもあれば group を表示する。親の default click が未許可 settings route を開かない構造にする。
  - `/lstep` route wrapper の一律 `ResourceHospitalSettings` guard を child-specific guard へ分解する。`checkup-sync` は `ResourceHospitalSettings` を維持し、`delivery-monitor` と `analytics` は `ResourceLstepAnalytics` を要求する。wrapper 自体から authorization を外しても、未許可 sibling を露出させてはならない。
  - 受け入れは、Analytics-only / HospitalSettings-only / both / neither の四 matrix を実 `SidebarItemWithPermission` render と route test の両方で検証すること。Analytics-only は「配信監視」を表示して delivery-monitor page に到達できるが checkup-sync を拒否されること、HospitalSettings-only は checkup-sync に到達できるが delivery-monitor / analytics を拒否されること、both は各許可 child に到達できること、neither は group 非表示かつ全 child route を拒否されることを証明する。static menu-data assertion だけでは完了としない。
- **Rationale (vs product philosophy)**: 既存の監視工程へ到達するための URL 転記・共有を削除する。一方、親 gate による隠れた追加権限は簡素化せず残さない。
- **Overrides brief recommendation?**: **No, A を採用**。ただし「path と child がある」ことを end-to-end nav 完了とみなす過剰 close は採用しない。
- **Evidence**: `reports/2026-07-31-line-residuals-R01-R07.md:95-122`; `reports/2026-07-31-line-r06-r07-nav-honesty.md:22-33,69-80`; `frontend/src/config/paths.ts:259-262`; `frontend/src/components/shared/Layout/sidebar-menu.tsx:114-126`; `frontend/src/app/routes/operations-routes.tsx:12-59`; `frontend/src/components/shared/Layout/SidebarItems.tsx:127-138`。

### R-07 — タグ管理 RBAC

- **Chosen**: **A — CLOSE_ROUTE_AUTHORITY + EFFECTIVE_NAV_OPEN**
- **Owner**: PO (gpt-5.6 Codex session 2026-07-31) / 業務責任ロール: L-step 権限・security PO
- **Purpose**: sidebar が見える権限と route に入れる権限を一つにし、「見えるが入れない / 入れるが見えない」を許可モデルとして固定しない。
- **Binding policy**:
  - タグ管理の product permission は existing route の `ResourceLstepAnalytics` を正とする。sidebar child を `ResourceHospitalSettings` へ戻さない。
  - landed child resource alignment を追認し、original R-07 の child-level mismatch は close する。
  - 専用 resource、両 resource 必須、frontend-only 例外を本決裁から追加しない。
  - parent `ResourceHospitalSettings` が analytics child を隠す実効 AND 条件は R-06 と同じ High follow-up で除去する。
  - BE/API の write authorization が route policy と一致する証拠は本 unit に無い。backend permission を推測で緩和せず、inventory と security test 後に整合する。
- **Risks accepted**:
  - HospitalSettings-only role はタグ管理導線を失う。これは route に入れない現行 product boundary と一致する。
  - `ResourceLstepAnalytics` という名称でタグの管理操作を守るため、resource 名と業務動詞の意味が直感的でない。
  - parent gate と backend authorization が未検証のため、child field の一致だけでは end-to-end RBAC 完了ではない。
- **Follow-up**:
  - R-06 の parent-container matrix test をタグ管理にも適用し、Analytics-only で表示・route 到達、HospitalSettings-only で非表示・route 拒否を証明する。
  - BE/API のタグ read/write endpoint と action-level permission を inventory し、frontend route より弱い authorization、clinic scope 欠落、view 権限による write 許可が無いことを security test で固定する。不一致があれば chosen resource のまま action policy を整合し、別 resource の発明は再度の具体的要件がある場合だけ行う。
- **Rationale (vs product philosophy)**: 既存 route を authority とし、sidebar だけの第二 RBAC truth を削除する。新 resource や二重条件を追加せず、現在の一つの境界を end-to-end に通す。
- **Overrides brief recommendation?**: **No, A を採用**。ただし current implementation を runtime-complete とは扱わない。
- **Evidence**: `reports/2026-07-31-line-residuals-R01-R07.md:126-153`; `reports/2026-07-31-line-r06-r07-nav-honesty.md:34-47,69-80`; `frontend/src/components/shared/Layout/sidebar-menu.tsx:114-126`; `frontend/src/app/routes/settings-routes.tsx:304-313`; `frontend/src/components/shared/Layout/SidebarItems.tsx:127-138`。

### TASK-021 Stage A Phase2 — 着手可否

- **Chosen**: **PHASE2_START_APPROVED / CLEAN_GO_DROP_HOLD**
- **Owner**: PO (gpt-5.6 Codex session 2026-07-31) / 業務責任ロール: 予約・staff capability PO
- **Purpose**: exclusion compatibility consumer を順序立てて capabilities-only contract へ移し、最終的な二重 SoT 撤去を可能にする。table drop を先行して予約機能を壊さない。
- **Binding policy**:
  - 別 claim / 別 implementation packet で、inventory §6.1–§6.3 の consumer proof、positive capability contract、known-client migration、BE facade/port consumer 撤去、OpenAPI/docs 整合へ着手してよい。
  - 最初に FE / known client の exclusion field 非利用と、legacy endpoint の external use zero または deprecation 終了を証明する。証明前に public contract を削除しない。
  - §6 の順序を守り、`capable_courses` / capable path を正式 contract にした後で exclusion response/path を撤去する。
  - 本決裁は `staff_reservation_exclusions` table、seed、seed-export、RLS、numbered DROP migration の削除承認ではない。CLEAN-GO / DROP / `make migrate` は HOLD。
  - `available-staffs` は WONTFILE のまま。既存 `reservation-staffs`、`on-duty-staffs`、`available-times`、capabilities surface を削除しない。
- **Risks accepted**:
  - deprecation / external-use evidence が取れない場合、Phase2 は contract removal 前で HOLD する。
  - BE/OpenAPI を段階移行する期間は compatibility surface の保守コストが残る。
  - generated types と tests の更新順を誤ると client build を壊すため、contract-first の小さい slice が必要になる。
- **Follow-up**:
  - §6.1 の consumer proof → capable contract → internal consumer migration → BE/OpenAPI exclusion surface removal → inventory 再実行、の順で packet 化する。
  - Phase2 受け入れは、known FE/client が exclusion fields/endpoints を送受信しないこと、`capable_courses` が OpenAPI と generated types に存在すること、production BE が exclusion facade を呼ばないこと、clinic isolation と candidate selection が capabilities-only で fail-closed なこと、available-staffs ban が残ること。ここまで PASS しても CLEAN-GO ではなく、§6.4–§6.6 と別の破壊変更承認が必要である。
- **Rationale (vs product philosophy)**: consumer を先に capabilities-only へ移してから obsolete surface を削除する。即時 DROP も恒久 facade も採らない。
- **Overrides brief recommendation?**: **N/A**。LINE brief の対象外。prior staged B→A policy は再審理せず、inventory §6 の Phase2 start だけを承認する。
- **Evidence**: TASK-021 Stage A inventory §5–§6, lines 150–208。

## Out-of-scope residuals

- **R-02**: production webhook / signature / provisioning は ops / USER のまま。
- **R-03**: runtime observation は TASK-010 のまま。
- **R-04**: L-step Write 再有効化・実送信は ops / USER のまま。
- **R-08**: LIFF ID 一致は deploy/ops residual のまま。
- 上記四件、本番 LINE 操作、credential rotation 実行、DB/migration/seed 適用を本書は承認しない。

## 決裁後の実装優先度と close boundary

1. **High**: R-05 single-SoT inventory / cutover design。mismatch を自動 winner 選択しない。
2. **High**: R-06 / R-07 共通の parent-container RBAC honesty + rendered permission matrix。
3. **Medium**: R-01 architecture summary + executable webhook contract test。
4. **Medium**: TASK-021 Phase2 を既存 TASK-021 claim と競合しない別 packet で順序実装。

- **PO decision closed**: R-01 / R-05 / R-06 / R-07 / TASK-021 Phase2 start。再決裁不要。
- **Original residual closed**: R-06 の path / child item 欠落、R-07 の child resource mismatch。
- **Implementation open**: R-01 docs/test、R-05 migration/cutover、R-06/R-07 effective parent gate、TASK-021 Phase2。
- **未証明**: R-06/R-07 runtime-green、R-05 live credential consistency、TASK-021 CLEAN-GO/DROP。

follow-up が実装済みであるとの主張は、各 deterministic verification が PASS してから行う。
