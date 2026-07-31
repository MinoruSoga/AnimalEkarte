本文書は推奨であり、裁定ではない。

# Decision recommendation pack — 2026-07-31

基準は `HEAD e1b51ceae0e8370146cdb13365a8324985728977` と、2026-07-31 に `gh issue list --state open` で実測した live open Issue 21 件である。批准者欄は役割だけを示し、個人名、資格情報値、臨床値、価格、契約条件は記入しない。

## 全項目表

| 現状分類 | 推奨処置 | 根拠 | 批准者の役割 | 解除条件 | AI裁定可否 |
|---|---|---|---|---|---|
| #89 — USER 専権・release blocker | 親 security gate として維持し、対象 credential 群の切替、旧値失効、必要 session 失効、非機密の調査結果を一つの bundle で完了してから close を推奨する。履歴対応は下記 A/B を批准する | `gh issue view 89` は OPEN。`docs/ops/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md:24-46` | Security owner、Release owner | 全旧値の revoke、必要 session の無効化、health、調査期間と非機密結論、scan、履歴方針の批准 | 不可 |
| #97 — USER 専権・release blocker | #89 の history/public-surface 子 gate として同じ bundle で完了し、表示 mask は失効後に行うことを推奨する | `gh issue view 97` は OPEN。`docs/ops/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md:24-29,38-46` | Security owner、Repository owner | #89 の解除条件に加え、history、fork、clone、artifact の影響評価と公開面の安全な mask | 不可 |
| #98 — USER 専権・repository slice 完了 | 新しい撤去実装は行わず、legacy RDS credential が #89/#97 の失効対象に含まれることを非機密に確認後、重複 Issue として close を推奨する | `test ! -e scripts/stg-db-tunnel.sh` は exit 0。`docs/ops/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md:7-9,51-53` | Security owner、Infrastructure owner | legacy 値の失効確認。script/docs 撤去は充足済み | 部分可。重複判定は可、失効操作は不可 |
| #99 — USER 専権・repository slice 完了 | ECS rollback や hot standby を再構築せず、provider 側の旧 ECS/IAM 経路が非稼働であることを一度確認して close を推奨する。本番整備は #253 に集約する | `test ! -e .github/workflows/backend-deploy-ecs.yml` は exit 0。`docs/ops/deploy/CI-CD-PIPELINE.md:7-23,28-38,114-127` | Infrastructure owner、Release owner | 旧経路の非稼働確認と現行 Cloudflare rollback 正本の確認 | 部分可。repository 判定は可、provider 操作は不可 |
| #201 — 判断待ち | **推奨案（未確定）**。現行の物理 block/fallback 設計を維持し、上限、warning 帯、欠落・lookup 障害時挙動の 3 行を一回で批准または修正要求することを推奨する | `backend/internal/medicalrecord/treatment_service.go:315-325,472-489`、`backend/internal/medicalrecord/treatment_dose_save.go:14-18,38-77`、`q&a.html:440-444,495-497` | 臨床責任者、Release acceptance owner（個人名は空欄） | 対象、単位、出典を備えた署名済み 3 行、対象環境の安全境界観測、close 批准 | 部分可。gate 設計は可、薬量・warning 数値と医学的妥当性は不可 |
| #211 — 判断待ち | **推奨案（未確定）**。新規コードや補助表を作らず、provisional seed を唯一の正本として行単位で一度だけ批准・訂正し、DB 適用とは別 gate にすることを推奨する | `q&a.html:452-456,499`、`backend/migrations/001_init.sql:3530-3652,3682-3723`、`backend/migrations/seeds/003_demo/checkup_type_fields.csv:1-54` | 臨床責任者、価格責任者、Environment owner（個人名は空欄） | canonical seed の行別 disposition、local fresh apply 証跡、対象環境の非破壊 schema 証跡 | 部分可。single-source と順序は可、臨床値・価格・DB 操作は不可 |
| #212 — 依存待ち | blanket な全 package coverage epic は実装せず、current artifact 取得後に clinic isolation、authorization、rollback、DB error の failure mode 別小 Issue へ分解することを推奨する | `docs/ops/coverage-policy.md:41-55,69-77`、`backend/.coverage-baseline:1-6`、`q&a.html:385-404`。`gh run view 30618388216` では Backend が skipped | Product Owner、Engineering Quality owner | current successful backend coverage artifact、除外方針の批准、未到達 branch の risk-prioritized inventory | 部分可。test 分解は可、target 批准と remote run は不可 |
| #235 — 依存待ち | DnD packet を作らず、利用者、頻度、現行操作数・時間、失敗率の実測を待つことを推奨する。価値が立証されなければ scope から削除する | `q&a.html:385-404`、`frontend/src/features/medical-records/components/ImageGalleryFilter.tsx:95-118`、`docs/product-philosophy.md:56-70,143-152` | Product Owner、現場業務責任者 | 既存 picker で不足することと、対象利用者、月次頻度、操作時間、失敗率、目標値の批准 | 部分可。再利用設計は可、需要・投資判断は不可 |
| #249 — 着手可能（technical）・判断待ち（clinical） | R-3 duplicate semantics と R-5 定性基準値表示を小 packet で進めることを推奨する。R-1 range は **推奨案（未確定）** とし、auto-commit は手動運用、停止、失敗通知、audit の実証まで開かない | `docs/spec/screens/settings/master-examinations.md:61-70`、`backend/internal/medicalrecord/lab_import_repository.go:196-217`、`frontend/src/features/examinations/components/ExamPivotTable.tsx:49-59`、`q&a.html:458-462,500` | Product Owner、臨床責任者、検査運用責任者（個人名は空欄） | R-3/R-5 scoped green、項目×動物種×測定系の署名済み range、手動 import と stop/notify/audit 証跡 | 部分可。technical residual は可、range と automation enable は不可 |
| #250 — 依存待ち | canonical consumer は完成と認め、完全な formal producer bundle 受領まで apply しないことを推奨する。移行自体の要否は下記 A/B を批准する | `docs/ops/deploy/CLINIC_CSV_IMPORT.md:5-18,48-58`、`Makefile:150-163`、`backend/internal/csvimport/cutover_contract_test.go:274-345` | Source data owner、Production cutover authority | 全 21 表と payment graph を含む complete/PASS bundle、業務照合、#253/#254/#255 の release gate | 部分可。consumer 検証は可、source 完全性・production apply は不可 |
| #252 — USER 専権 | 批准済み設定との差分だけを各院へ投入し、preview と過去履歴非再計算を確認することを推奨する。standard PATCH の境界 validation と audit 不在は別の agent-ready New Work 起票候補とし、この Issue の OPS scope と混ぜない | `gh api repos/MinoruSoga/AnimalEkarte/issues/252/comments` の最新 comment は `OPS_ONLY`。`backend/internal/clinic/closing_settings_request.go:3-7`、`backend/internal/clinic/closing_settings_service.go:141-165,350-364` | Clinic billing operations owner | 全院 read-only inventory、差分のみの USER apply、preview、過去履歴の非再計算。New Work は別 ID で acceptance を確定 | Issue の投入判断は不可。New Work の技術推奨は可 |
| #253 — USER 専権・USER blocker 付き hybrid | **推奨案（未確定）**。billing、GitHub production Environment、required reviewer を USER が先行し、その後に workflow 差分を実装・検証することを推奨する。ECS 経路は戻さない | `docs/ops/deploy/CI-CD-PIPELINE.md:26-38,87-95,171-182`、`docs/ops/infra/production/runbook.md:184-196` | Production release authority、Budget/contract owner（個人名は空欄） | billing 復旧、Environment/reviewer、必要設定、latest required CI green、restore と rollback rehearsal | 部分可。config 準備は可、課金・外部設定・production 操作は不可 |
| #254 — USER 専権 | 新しい agent packet を作らず、一つの isolated clean demo UAT で認証付き suite、5 フロー、DB/audit、実 LINE/LIFF、使い勝手 sign-off を束ねることを推奨する | `docs/ops/testing/scenarios/reports/2026-07-31-local-issue-254.md:10-12,58-89,199-211,234-250` | UAT acceptance owner、QA lead、LINE configuration owner | 安全な認証変数注入、認証付き suite、5 human flows、DB/audit、実 LINE/LIFF、全 FAIL disposition と sign-off | 部分可。suite と証跡整形は可、人の観測・waiver は不可 |
| #255 — 判断待ち（roster 受領済み）→ USER apply | roster の再待機はやめ、email 方針、clinic mapping、employment status、role-to-group を一回で批准してから、off-repo input を preflight し USER apply することを推奨する | `gh issue view 255 --json comments` の最新 comment は roster 54 名受領済みと PO 4 論点を記録。`docs/ops/deploy/STAFF_ACCOUNT_PROVISIONING.md:17-23,70-139` | Client identity and access owner、Product Owner | email 方針、各人の main/assigned clinic、employment inclusion、explicit permission groups、authorized actor、secure input | 部分可。preflight は可、本人性・業務権限・初期 credential 配布・apply は不可 |
| #256 — USER 専権・privacy blocker | clean demo 環境へ戻した後に `05-accounting`、`07-examinations`、`10-trimming` だけを再撮影し、10 画像を一回で sign-off することを推奨する。PII-bearing local history は publish 前に USER 承認の隔離・対応を行う | `reports/2026-07-31-task-024-manual-audit.md:164-190`、`todo.md:125-130` | Privacy/security owner、Documentation owner | history containment、clean-demo 3-image recapture、非 PII 確認、10/10 visual sign-off | 部分可。clean 環境後の capture/scan は可、reset・history 操作・sign-off は不可 |
| #257 — USER 専権 | **推奨案（未確定）**。過去日付の runbook を実行せず、全 gate green 後に新 window へ再スケジュールすることを推奨する。実施継続の要否は下記 A/B を批准する | `docs/delivery/GOLIVE_RUNBOOK.md:1-6,10-42,78-124` | Go-live authority、Client operations owner（個人名は空欄） | 全 prerequisite green、新日時、technical operator、client contact、support window、通知、rollback/restore 証跡 | 部分可。checklist は可、Go/No-Go と契約条件は不可 |
| #258 — 判断待ち | **推奨案（未確定）**。U1–U12 を正本に、client ownership と handover を推奨する。developer ownership を選ぶ場合は料金、期間、責任、終了時移管を明示契約する | `docs/delivery/DELIVERY_PACKAGE.md:3-16,215-240`、`q&a.html:507-521` | Contract owner、Client approver（個人名は空欄） | U1–U12 全行の出典・値・責任、#253 実測、approval envelope | 不可 |
| #259 — 依存待ち | Issue を外部 API enablement と STG acceptance に狭め、write/cron を再実装しないことを推奨する。契約条件が必要なら **推奨案（未確定）** のまま責任者へ戻す | `docs/ops/deploy/LSTEP_WRITE_API_PAUSE.md:7-25,34-44,57-78`、`backend/wrangler.jsonc:97-102`、`backend/worker/scheduled-jobs.ts:30-34` | Integration service owner、Product/contract owner（個人名は空欄） | provider API 有効化、安全な設定受領、clinic/master 整合、少数 STG 送信、通知・停止・audit | 部分可。source 判定は可、契約・外部 enable・実送信は不可 |
| #260 — USER 専権（close approval） | dated master-plan 本文の継続同期をやめ、この report と個別 Issue/view の突合後に historical close することを推奨する | live `gh issue view 260` は 7/27 計画。`3-session-agent.html:232-237`、`docs/product-philosophy.md:52-70` | Delivery owner、Project owner | 本 report が live open 集合を網羅し、個別 Issue と lightweight view で追跡可能と確認 | 推奨は可、GitHub close は不可 |
| #261 — USER 専権・既決と矛盾 | live Issue の未決分類要求と DEC-41/47 が競合するため、この run では推奨しない。下記「既決と矛盾」に隔離する | live latest comment は `PO_DECISION / PARTIAL hub`。`q&a.html:274-301,446-450` は 16 件を判定済みとし、残りを OPS-2/3/4 に限定 | Product Owner、Decision-record owner | live Issue を DEC-41/47 に整合するか、正式な override/reopen を別 run で記録 | 不可 |
| #284 — 依存待ち | source 変更を増やさず、試験環境と 3 実機の受領後に cold/warm/offline を含む一回の device QA を推奨する | `frontend/line-reserve/index.html:7-12`、`frontend/line-reserve/src/index.css:17-23`、`frontend/line-reserve/src/lib/liff-config.ts:6-14`、live `gh issue view 284` | QA lead、QA environment owner、Device custodian | 許可済み試験範囲、3 実機、Rendered Fonts、fallback、clip の evidence | 部分可。test design は可、実機観測は不可 |
| TASK-004 — ops event 待ち・削除候補 | 独立 open implementation task から外し、次の intentional land 時だけ発火する共通 checklist へ統合することを推奨する | `reports/2026-07-31-task-004-005-land-proc.md:39-47`、`docs/product-philosophy.md:52-59` | Repository integration owner | land 直前の porcelain、path-scoped stage、foreign WIP 非 stage | 可。境界照合は可、commit は USER |
| TASK-005 — ops event 待ち・削除候補 | TASK-004 と同じ land checklist へ統合し、独立 open item として反復追跡しないことを推奨する | `reports/2026-07-31-task-004-005-land-proc.md:31-46` | Verification/integration owner | staged 後の docs drift check と scoped tests | 可。検証は可、land 判断は USER |
| TASK-009 — USER 専権・破壊操作待ち | 下流 UAT または安全な再撮影に clean demo DB が必要な場合だけ、一度 USER runbook を実行することを推奨する | `reports/2026-07-31-task-009-reseed-ops.md:14-21,82-116,214-221` | Local environment/data owner | local 限定、必要データ退避、データ損失受容、USER 実行、post-check | 不可。静的事後確認のみ可 |
| TASK-010 — 依存待ち・hybrid runtime residual | 残数を人が全件目視せず、TASK-023 の一回 UAT 後に機械再 census し、客観 FAIL は BUG、臨床・実機・使い勝手だけを human exception に残すことを推奨する | `todo.md:45`、`docs/ops/testing/scenarios/reports/2026-07-31-local-issue-254.md:194-210`、`docs/product-philosophy.md:163-165` | QA lead、Scenario業務 owner、FAIL disposition owner | 必要 seed、認証環境、TASK-023 五フロー後の再 census | 部分可。客観検査・振分は可、人間観測は不可 |
| TASK-020 — credential 依存待ち | TASK-023 と同じ一回の認証変数注入 window で 93-test runtime を実行し、別 handoff を作らないことを推奨する | `reports/2026-07-31-task-020-env-forward.md:11-14,70-83,109-114`、`todo.md:46,98-102` | USER credential owner、QA/release owner | host への name-only `E2E_LOGIN_*` 注入、対象 DB 整合、workers=1 scoped run | 部分可。結果判定は可、credential 供給は不可 |
| TASK-021 source — 着手可能・既決・claim 中 | 新しい批准を求めず、consumer、BE、OpenAPI の非破壊 wave だけを順序実装することを推奨する | `reports/2026-07-31-line-residual-po-decisions-FINAL.md:129-149` | Verification owner。新規 product 批准は不要 | current claim owner 完了、known consumer zero、capability-only contract、scoped green | 可。source/test まで |
| TASK-021 DROP/apply — USER 判断待ち | HOLD を維持し、consumer zero と cleanup gates が全部 green の時点で一度だけ CLEAN-GO を求めることを推奨する | `reports/2026-07-31-task-021-stage-a-inventory.md:192-208`、`reports/2026-07-31-line-residual-po-decisions-FINAL.md:135-146` | Reservation/staff capability PO、DB/release operator | consumer zero、tests、OpenAPI、seed/export 整理、numbered migration review、破壊承認 | 不可。migration author/test は可、批准・apply は不可 |
| TASK-022 — local human residual | S13 手順、operational signer、application-role RLS を一つの release packet にし、それまでは Phase 2 を開かないことを推奨する | `reports/2026-07-31-task-022-identity-link-closeout.md:69-79`、`docs/ops/testing/scenarios/S13-identity-links-manual-correction.md:36-47` | Operational/clinical signer、DBA/security release owner | 非 PHI の 2 医院 fixture、必要権限 actor、全手順 PASS、署名、RLS runtime | 部分可。証跡検査は可、実施・署名は不可 |
| TASK-023 — local human residual | TASK-010/020 を束ねる一回の UAT umbrella とし、認証注入、五フロー、DB/audit、実 LINE/LIFF、sign-off を一 bundle で完了することを推奨する | `docs/ops/testing/scenarios/reports/2026-07-31-local-issue-254.md:199-211` | QA/UAT lead、PO/現場責任者、LINE configuration owner | 認証注入、seed 付き local、実機、DB observer、sign-off | 部分可。suite と証跡整理は可、人間観測は不可 |
| TASK-024 content — USER 依存＋human sign-off | clean demo 環境で必須 3 枚だけ再撮影し、全 10 枚を一回承認することを推奨する。任意 2 枚と 2×化は具体要件がない限り行わない | `reports/2026-07-31-task-024-manual-audit.md:164-180,184-190` | Documentation owner、Privacy reviewer | USER clean demo、認証 session、3 枚の非 PII 確認、10/10 sign-off | 部分可。capture/scan は可、reset/sign-off は不可 |
| TASK-024 history — USER/privacy 専権 | PII-bearing blob を含む local history を publish せず、privacy/repository owner 批准済みの隔離・履歴対応へ分離することを推奨する | `reports/2026-07-31-task-024-manual-audit.md:164-182`。`git rev-list --left-right --count origin/main...HEAD` は `0 15` | Privacy/security owner、Repository history owner | 影響範囲、批准済み手順または隔離方針、再 scan | 不可 |
| LINE R-01 — source landed・verification 待ち | current docs/test を scoped Docker gate で確認後、open follow-up から削除することを推奨する。再決裁しない | `docs/spec/line/architecture.md:41-53`、`backend/internal/lstep/line_link_service_test.go:312-351` | Verification/release owner | unsupported side-effect zero、follow/unfollow、signature-routing tests green | 可 |
| LINE R-02 — USER 専権 ops | agent queue から外し、本番 webhook/signature/provisioning release checklist としてだけ保持することを推奨する | `reports/2026-07-31-line-residual-po-decisions-FINAL.md:151-157`、`reports/2026-07-31-task-019-line-deep-audit.md:108` | LINE platform ops、Security/release owner | approved environment、destination/clinic provisioning、monitoring、rollback | 不可 |
| LINE R-03 — TASK-010 alias | standalone residual を削除し、TASK-010 だけで追跡することを推奨する | `reports/2026-07-31-line-residual-po-decisions-FINAL.md:153-155`、`docs/ops/testing/scenarios/reports/2026-07-31-local-issue-254.md:194-195` | Ledger owner | TASK-010 recensus に LINE marker を含める | 可 |
| LINE R-04 — USER 専権・external write | deploy gate を OFF のまま維持し、前提完了後に批准済み STG 少数送信を一回行うことを推奨する。契約値は **推奨案（未確定）** のまま残す | `docs/ops/deploy/LSTEP_WRITE_API_PAUSE.md:57-78,82-87` | Vendor/contract owner、LINE ops、Release owner（個人名は空欄） | channel/master 整合、read smoke、復旧手順、外部 API 批准、停止・通知 | 不可 |
| LINE R-05 source — 着手可能・既決・claim 中 | A-CI を再審理せず、inventory、canonical verifier、旧 request/read/write 撤去の順で single-SoT cutover を実装することを推奨する | `reports/2026-07-31-line-residual-po-decisions-FINAL.md:55-81` | Security reviewer。新規 product 批准は不要 | current claim owner 完了、consumer inventory、mismatch-safe tests | 可 |
| LINE R-05 live credential/migrate — USER 専権 | mismatch の winner を推測せず clinic を HOLD し、正規 UI で手動再設定後に USER が migration を適用することを推奨する | `reports/2026-07-31-line-residual-po-decisions-FINAL.md:64-78` | LINE/security owner、Clinic integration owner、DB/release operator | 値非出力 inventory、manual resolution、source inventory zero、批准済み apply | 不可 |
| LINE R-06 — source landed・verification 待ち | parent/sidebar/route matrix を targeted test で確認後、open row から削除することを推奨する | `frontend/src/app/routes/operations-routes.tsx:12-63`、`frontend/src/app/routes/operations-routes.lstep-delivery-monitor.test.tsx:42-84`、`frontend/src/components/shared/Layout/SidebarItems.test.tsx:100-181` | Frontend verification/release owner | Analytics-only、HospitalSettings-only、both、neither の matrix green | 可 |
| LINE R-07 — parent fix landed・BE/API residual は着手可能 | chosen resource を変えず、tag-management の BE read/write permission と clinic scope を inventory/test し、不一致だけ直すことを推奨する | `reports/2026-07-31-line-residual-po-decisions-FINAL.md:107-127`、`frontend/src/app/routes/settings-routes.tsx:304-313`、`backend/internal/lstep/routes.go:158-186` | Security reviewer、Release owner | action-level inventory、view による write 不可、clinic isolation、FE/BE resource 整合 | 可 |
| LINE R-08 — USER 専権 deploy ops | 新しい application SoT を増やさず、deploy 前に pet-health と line-reserve の LIFF ID 一致を blocker として検査することを推奨する | `reports/2026-07-31-line-residual-po-decisions-FINAL.md:151-157`、`reports/2026-07-31-task-019-line-deep-audit.md:114` | Deploy owner、LINE/LIFF configuration owner | 対象環境の両 ID 一致、実 LIFF smoke、値非出力 evidence | 不可 |

### 「やらなくてよい」を優先する推奨

- #98 の tunnel/script 追加撤去、#99 の ECS rollback/hot standby 復元、#259 の write/cron 再実装、#260 の dated master-plan 詳細同期は行わない。
- #212 の blanket 数字稼ぎ、#235 の需要未実測 DnD、#261 の独立臨床値表、#284 の実機観測前の font source 変更は行わない。
- TASK-004/005 の独立 open tracking、TASK-010 の全件人手目視、TASK-020 の別 credential handoff、TASK-024 の根拠のない FAQ・任意再撮影は削除または統合する。
- LINE R-03 の重複追跡を削除し、gate green 後は LINE R-01/R-06 の open row も閉じる。

## 意見が割れる論点だけの A/B

### Credential history — #89/#97

- A（推奨）: rotation/revocation を実効対策とし、history rewrite は行わない。
- B: rotation 後にも必要性を Security owner と Repository owner が立証した場合だけ、fork/clone/deploy coordination 付きで rewrite する。
- 推奨理由: central history を変えても既存 clone は消えず、失効こそが再利用を止める。B は破壊的で coordination failure を増やす。
- 推奨が外れた場合の損失: A が不十分なら revoked 済みの過去値が取得可能なまま残る。B を誤採用すれば clone 分断、deploy/history 参照破損、偽の「消去済み」認識を生む。

### Legacy data cutover — #250

- A（推奨）: complete/PASS の正式 bundle を作成し、rehearsal と批准後に移行する。
- B: legacy 移行を不要と正式決定し、retention 方針付き read-only 保管へ切り替える。
- 推奨理由: current consumer は complete bundle を fail-closed に検証でき、partial apply より clinical/accounting history の連続性を守れる。
- 推奨が外れた場合の損失: A が不要なら producer 完成と照合コストが無駄になる。B が誤りなら参照すべき履歴を新 system から失い、二重参照運用が残る。

### Go-live continuation — #257

- A（推奨）: 全 gate green 後に新しい window へ再スケジュールする。
- B: Go-live を正式に延期・中止し、暫定運用と責任期間を明文化する。
- 推奨理由: 過去日付の draft を再利用せず、実測済み prerequisite と rollback に基づく一回の切替にできる。
- 推奨が外れた場合の損失: A が早過ぎれば切替失敗と復旧不能を招く。B が不要なら二重運用と納品遅延が長期化する。

### Delivery ownership — #258

- A（推奨）: client が service account、契約、請求、backup、support、monitoring を所有し、handover を受ける。
- B: developer ownership を料金、期間、責任、終了時移管付きの明示契約にする。
- 推奨理由: A は shadow dependency と引渡し後の単一開発者依存を減らす。
- 推奨が外れた場合の損失: A で client の運用能力が足りなければ障害対応が遅れる。B を未整備で選べば継続責任、費用、移管の紛争になる。

## 既決と矛盾

ここでは推奨を行わず、既決と current evidence の不一致だけを隔離する。

- #201 live Issue 本文は緊急例外の採否を未決扱いするが、DEC-47 Q1 は緊急例外を追加しない設計を固定している（`q&a.html:440-444`）。推奨なし。
- #249 live Issue 本文は空・parse 不能を `normal` と記すが、DEC-47 Q4 と current source は `is_assessed=false` の未判定を正とする（`q&a.html:458-462`、`backend/internal/medicalrecord/exam_result_assessment.go:53-65,146-148`）。推奨なし。
- #261 live latest comment は各 SD の PO 分類を未達とするが、DEC-41 は 13 件対象消失、2 件実装済み、1 件実機判定不能へ分類済みで、DEC-47 Q2 は新しい独立値表を禁止する（`q&a.html:274-301,446-450`）。live Issue と decision SoT の整合または正式な override が先であり、推奨なし。
- DEC-42 は旧 seed path と demo CSV の `ATOPY` 行を根拠にするが、current `backend/migrations/seeds/003_demo/pet_chronic_conditions.csv` は header のみで、引用 counterexample が存在しない（`q&a.html:308-321`）。裁定を変更せず evidence drift として推奨なし。
- DEC-46 は full-parent authorization proof を未報告と記すが、current source/test は着地済みである（`q&a.html:408-427`、`backend/internal/identitylink/service.go:481-486`、`backend/internal/identitylink/service_test.go:497-637`）。残る manual workflow と application-role RLS gate は維持し、推奨なし。

## AI が決めてはならない

- #201: 薬剤・種別・master row ごとの上限と warning 帯。必要形式は target、数値、単位、臨床出典、発効日。source-owner role は臨床責任者。未承認時は DEC-47 の既存 fail-safe を維持する。
- #201 の欠落/lookup matrix: weight、species、param、lookup failure ごとに現行挙動の批准または正式な再審要求が必要。source-owner role は臨床責任者。#261 は批准済み row ID を参照し、値を複製しない。
- #211: seed bundle、row ID、field ごとの type、choice、range、unit、出典。source-owner role は臨床責任者、価格を変更する行だけ価格責任者。未承認 row は適用しない。
- #249: 検査項目×動物種×測定系の上下限または定性規則、unit、検証済み出典。source-owner role は臨床責任者。未承認時は推測せず未判定を維持する。
- Vaccine species: master row ごとの dog/cat/other/both と alias、出典。source-owner role は臨床責任者。unknown は silent 非表示にしない。
- #258: U1–U12 の契約名義、plan、請求、権限、backup、support、monitoring、保持。必要形式は選択/値、通貨・期間・時刻または該当なし、契約・service spec・runbook evidence。source-owner role は Contract owner、Client approver、各医院管理者、#253 operations owner。
- PO-008: annual metrics の期間境界、CSV 要否、L-step path ごとの継続、通知、rollback。必要形式は current source value と approve/correct。source-owner role は Client/contract owner。未承認時は current semantics を統一せず、write は default-off と実 2xx 成功判定を維持する。
- Sentry paid plan: free cap 超過時だけ plan、budget cap、payer、renewal/cancel authority、通貨・期間、vendor quote。source-owner role は Budget/contract owner。未承認時は free-only のまま auto-upgrade しない。
- 外部 write、credential rotation、DB reset/migration、production/go-live、実機、目視、sign-off は値ではなく行為そのものが USER/人間境界であり、AI は実施済みと裁定しない。
