# Open Issue readiness dossier — 2026-08-01

基準は `main` の調査開始 HEAD `70da134ad2fc48aa40e8dfc0de88b30e7b90a9e8` と、2026-08-01 に read-only で取得した GitHub の open state・本文・コメントである。実行時 open 集合は `#89, #97, #98, #99, #201, #211, #212, #235, #249, #250, #252, #253, #254, #255, #256, #257, #258, #259, #260, #261, #284`。生成時観測との差分はない。

事実の優先順位は executable contract、current code、machine-enforced rule、現行 docs、Issue 本文・過去コメントの順。Docker、DB、migration、seed、実機、外部サービス操作は本 docs-only unit では実行していない。以下の「未実測」コマンドは次セッション用であり、本 dossier の静的確認を runtime green と読み替えない。個人名、患者情報、認証値、接続文字列は転記していない。

## Issue #89 — CRITICAL: 現行環境の露出済みcredentialをローテーションし旧値を無効化
- 分類: USER 専権
- 題名/本文の陳腐化: なし（現行本文は provider-neutral で、repository containment と外部失効未証明を分離している）

### 現状実測
Issue は OPEN。外部操作 runbook は `docs/ops/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md:24-46`、予防的な gitleaks job は `.github/workflows/ci.yml:71-90` にある。repository 側の containment は current tree で確認済みだが、tree 修正や scan 設定は旧値失効の証拠ではない。restricted incident history の path/hash は本 dossier に再掲しない。

### 残作業
Security owner が対象系統を安全な channel で rotate/revoke し、必要な session 無効化、health、調査期間、非機密の scan 結論を一つの evidence bundle にする。値は成果物へ出さない。外部状態は未実測で、再確認は `gh issue view 89 --comments` と runbook の completion checklist を用いる。

### 次に動くのは
USER（Security owner / Release owner）。外部認証状態の変更と失効確認は agent の権限外であり、repository の静的証拠だけでは close できない。

### 着手プラン
着手不能。解除入力は USER による対象区分、実施者、window、rollback、非機密結果と restricted evidence reference。非機密成果物は `reports/YYYY-MM-DD-security-external-closure.md` とし、受領後の agent 確認は `gh issue view 89 --comments --json state,title,updatedAt` と `rg -n 'status|owner|window|result|restricted evidence' reports/YYYY-MM-DD-security-external-closure.md`。history rewrite は既定で行わず、承認後も実行は USER-only。agent は rewrite/force-push を行わない。

### 回答起案
`q&a.html#decision-pack-issue-readiness-ops-20260801` の #89 行。推奨は rotate/revoke/session invalidation を同一 release gate に束ねること。履歴 rewrite は既定にせず、影響評価後に Repository owner が別途承認する。

## Issue #97 — 🚨 CRITICAL: git履歴・公開Issue由来のcredential露出 — ローテーションと旧値無効化
- 分類: USER 専権
- 題名/本文の陳腐化: なし（現行本文は公開面・repository 是正済みと外部失効未証明を分離している）

### 現状実測
Issue は OPEN。問題となった staging environment file は current HEAD に存在しないが、削除履歴は到達可能なため path/hash を本 dossier に再掲しない。`docs/ops/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md:24-46` は #89 と同じ release gate を定義する。外部失効は本 unit では未実測であり、公開面の mask だけを完了証拠にしない。

### 残作業
#89 と同じ USER bundle で旧値失効、必要 session 無効化、clone/fork/artifact 影響評価、非機密 scan 結論を取得する。実状態は未実測。確認は `gh issue view 97 --comments` と当該 runbook を使う。

### 次に動くのは
USER（Security owner / Repository owner）。履歴や公開面の編集は失効の代替ではなく、外部 state と影響範囲の判断を伴う。

### 着手プラン
着手不能。解除入力は #89 と同一 window の外部失効結果と clone/fork/artifact 影響評価。非機密成果物は `reports/YYYY-MM-DD-security-external-closure.md` に統合し、agent は `gh issue view 97 --comments --json state,title,updatedAt` と同 report の値非表示 checklist だけを確認する。

### 回答起案
`q&a.html#decision-pack-issue-readiness-ops-20260801` の #97 行。推奨は #89 の子 gate として処理し、表示修正だけを closure proof にしないこと。

## Issue #98 — 🔴 CRITICAL: 旧RDS credential履歴の残余リスクと廃止スクリプト撤去
- 分類: USER 専権
- 題名/本文の陳腐化: あり（廃止スクリプト撤去は完了済み）

### 現状実測
Issue は OPEN。retired tunnel/deploy artifacts は current tree に存在しない。restricted deletion history の path/hash は再掲しない。repository slice は完了しているが、旧 RDS 側の失効は repository から観測できない。

### 残作業
Infrastructure/Security owner が旧 RDS 系統を #89/#97 の対象に含めたことと provider 側失効を非機密に記録する。外部状態は未実測。`gh issue view 98 --comments` と provider evidence の照合が必要。

### 次に動くのは
USER。新しい撤去実装は不要で、残るのは provider state と重複 Issue の close 判断である。

### 着手プラン
着手不能。解除入力は provider 側の非機密な失効結果と #89/#97 bundle の対象包含。成果物は `reports/YYYY-MM-DD-security-external-closure.md`、agent の確認は `gh issue view 98 --comments --json state,title,updatedAt` と current tree absence の `rg` に限定し、旧経路を再構築しない。

### 回答起案
`q&a.html#decision-pack-issue-readiness-ops-20260801` の #98 行。推奨は repository slice 完了を維持し、外部失効確認後に重複整理すること。

## Issue #99 — 🔴 HIGH: 廃止予定ECS deploy経路の撤去と現行rollback手順の一本化
- 分類: USER 専権
- 題名/本文の陳腐化: あり（repository 上の ECS deploy 撤去と Cloudflare rollback 正本化は完了済み）

### 現状実測
Issue は OPEN。旧 ECS workflow は存在せず、現行 workflow は `.github/workflows/backend-deploy.yml:1-67`。rollback は `docs/ops/deploy/CI-CD-PIPELINE.md:114-127` と production runbook にあり、関連 commit は `8e868e0d`, `10b23f82`。provider 側に実行可能な旧経路が無いことは未実測。

### 残作業
Infrastructure/Release owner が provider 側の旧 ECS/IAM 経路非稼働を一度確認し、現行 Cloudflare rollback rehearsal と #253 の production evidence を揃える。

### 次に動くのは
USER。provider 操作と production rehearsal は外部境界で、repository だけを見て close できない。

### 着手プラン
着手不能。解除入力は USER の provider inventory と Cloudflare rollback rehearsal。非機密成果物は `reports/YYYY-MM-DD-production-rollback-rehearsal.md` とし、agent は `gh issue view 99 --comments --json state,title,updatedAt`、current workflow absence、runbook checklist を照合する。ECS hot standby は再構築しない。

### 回答起案
`q&a.html#decision-pack-issue-readiness-ops-20260801` の #99 行。推奨は repository slice を完了扱いにし、provider 非稼働確認と #253 の rollback 証拠だけを残すこと。

## Issue #201 — [SAFETY] 薬量自動計算 — 上限超過の物理ブロックと例外統制
- 分類: 着手可能
- 題名/本文の陳腐化: あり（BE hard reject は着地済みだが、FE lookup 障害時の silent manual fallback が未記載）

### 現状実測
BE は repository/system error を保存中止へ伝播する一方、体重なし・species 名解決不能・parameter なしを同じ nil/no-error として上限検証から外す（`backend/internal/medicalrecord/treatment_dose_save.go:14-18,29-73`、特に `:29-42`; commit `fa119abb9`; vital fail-closed `b65cf69e`）。FE は dose parameter fetch error も catch して manual default へ落とす（`frontend/src/features/medical-records/components/TreatmentsTab/TreatmentsTab.tsx:239-277`、特に `:260-262`）。`TreatmentRow.tsx:80-105,172-187` では各欠落理由と取得障害を識別できない。

### 残作業
最終契約は FE/BE の technical failure、体重なし、species ID/名称の欠落・不正、該当 parameter なしを理由別にし、通常 dose 保存を全て fail-closed にする。ただし current addendum は finalized record 専用の自由記述であり、救急・既実施投薬の構造化事実記録を代替できない。TASK-025 は technical failure の visible error・保存停止・retry を先行実装する。TASK-033 は active/draft でも clinic/pet/medical-record 相関、薬剤、実投与量・単位、投与時刻、理由、authenticated actor、transaction-bound audit を持つ immutable な構造化救急投薬記録を通常治療履歴・handoff に表示できるようにし、その経路と臨床承認が green になった同一 cutover で欠落時の通常保存を fail-closed にする。上限値、warning 帯、記録対象ケースと権限条件は出典付き承認待ちで、現行 20% は実装観測にすぎず臨床推奨ではない。テストは未実測。

### 次に動くのは
先にエージェントが TASK-025 の technical failure slice を実装する。並行して臨床責任者が上限、warning、構造化救急投薬記録の対象ケース・権限・理由・訂正条件を承認し、その後エージェントが TASK-033 の記録経路と欠落時 cutover を一体で実装する。欠落時の最終 gate は設計判断だが、安全な代替記録経路なしに runtime だけを先行変更しない。

### 着手プラン
TASK-025 の RED は `TreatmentsTab.test.tsx` / `TreatmentRow.test.tsx` に lookup technical failure の visible error、通常保存不能、retry、upstream body 非表示を追加し、GREEN で error を manual default に変換しない typed state を row/save gate へ渡す。TASK-033 の RED は別 packet で、active/draft の構造化救急投薬記録、必須 field、wrong-clinic/pet/record、actor/audit failure rollback、通常履歴/handoff 表示、欠落時通常保存 zero-write を固定する。GREEN では dedicated event schema/API/UI を実装し、経路 green と臨床承認後にだけ欠落時 fail-closed を同時に有効化する。未実測コマンドは `docker compose exec -T frontend npx vitest run src/features/medical-records/components/TreatmentsTab/TreatmentsTab.test.tsx src/features/medical-records/components/TreatmentsTab/TreatmentRow.test.tsx`、`docker compose exec -T backend go test -p 1 ./internal/medicalrecord/... -run 'Test.*(EmergencyAdministration|DoseMissing|DoseSpecies|DoseParameter)' -count=1`、`docker compose exec -T backend go test -p 1 ./internal/apicontract ./internal/lintscan -count=1`。

### 回答起案
technical failure と欠落時通常保存の最終契約は `q&a.html#dec-48` で fail-closed に代理裁定し、安全な cutover 順序を TASK-025/TASK-033 に分離する。上限値、warning 帯、構造化救急投薬記録の対象・理由・権限・actor/audit・訂正条件だけを `q&a.html#decision-pack-issue-readiness-clinical-20260801` の一行承認に残す。

## Issue #211 — [VERIFY] 健診パッケージ実装 — DB適用・provisional seed臨床確認
- 分類: 判断待ち
- 題名/本文の陳腐化: あり（「対象環境へ非破壊 migration」の一般化は統合前 checksum DB に適用できない）

### 現状実測
schema/seed/code は `backend/migrations/001_init.sql:3530-3723` と checkup service/test 群に着地済み。`checkup_field_results` の dependent-child CASCADE は `backend/migrations/CLAUDE.md:16-26` と `backend/internal/medicalrecord/checkup_field_cascade_test.go:241-261` が明示的に許容するため、欠陥扱いしない。provisional checkup rows は `backend/migrations/seeds/003_demo/checkup_type_fields.csv` にあり、staging/production は `backend/migrations/CLAUDE.md:58` のとおり `002_master` だけをロードするため、対象環境へ昇格する canonical path は現状ない。統合前 DB には no-reset upgrade path がない（同 `:54-63`）。

### 残作業
臨床責任者が provisional row を行単位で承認/修正する。次に seed owner が対象 bundle/environment と `seed-export` による生成・昇格方法を設計し、既適用 bundle の checksum を変更しない経路を別 packet で確定する。DB owner は対象 DB の履歴を判定し、fresh/current と統合前を分け、後者は承認済み再構築計画を選ぶ。DB 適用・runtime は未実測。

### 次に動くのは
臨床責任者と USER（DB/release owner）。値の妥当性と環境履歴は agent が推測できない。

### 着手プラン
着手不能。解除入力は「row ID・値/単位・出典・承認者・発効日」「対象 bundle/environment」「seed-export による生成・昇格方針」「対象 DB の migration 履歴分類」。これらが揃うまで USER apply を手順化しない。受領後は migration-seed-safety review と fresh DB/対象環境を分けた packet を起票し、適用は USER が行う。

### 回答起案
current authority は `q&a.html#dec-58` と `q&a.html#decision-pack-issue-readiness-clinical-20260801`。DEC-47 Q3 の旧 bundle-order 文と下部 P0 表の「残は DB 適用と臨床確認のみ」は superseded とし、未承認値や未設計の昇格経路を別 docs/runtime default で補完しない。

## Issue #212 — [TEST][DECISION] Repository integration coverage の重要ギャップを改善
- 分類: 依存待ち
- 題名/本文の陳腐化: なし（phase2 gate と current coverage artifact 待ちは継続）

### 現状実測
ratchet baseline は `backend/.coverage-baseline:1-6`、policy は `docs/ops/coverage-policy.md:54-55,100-108`、CI 集計は `.github/workflows/ci.yml:195-203`。read-only 実測コマンド `gh run view 30648745802 --json databaseId,workflowName,status,conclusion,url,headBranch,headSha,jobs --jq '{databaseId,workflowName,status,conclusion,url,headBranch,headSha,jobs:[.jobs[]|select(.name=="Backend")|{name,status,conclusion}]}'` の逐語出力は `{"conclusion":"failure","databaseId":30648745802,"headBranch":"main","headSha":"70da134ad2fc48aa40e8dfc0de88b30e7b90a9e8","jobs":[{"conclusion":"skipped","name":"Backend","status":"completed"}],"status":"completed","url":"https://github.com/MinoruSoga/AnimalEkarte/actions/runs/30648745802","workflowName":"CI"}`。current backend coverage artifact は無く距離を定量化できない。DEC-45（`q&a.html:385-404`）は blanket epic を開始しないとする。

### 残作業
USER/CI が成功 artifact を提供し、coverage exclusion と優先 failure mode を批准する。その後 clinic isolation、authorization、rollback、DB error ごとの小 task に分解する。

### 次に動くのは
USER/PO/CI owner。成功 artifact が無いまま追加 test の量や package を agent が決めると risk-based prioritization にならない。

### 着手プラン
着手不能。解除入力は current successful backend coverage artifact と exclusion 方針。受領後、`corepack pnpm` を project root から使う CI 再現経路と Docker scoped package tests を task ごとに定義する。

### 回答起案
`q&a.html#dec-45` を維持し、`q&a.html#decision-pack-issue-readiness-ops-20260801` の #212 行で artifact と一行批准を待つ。

## Issue #235 — [FEAT] カルテ画像・PDFのドラッグ&ドロップアップロード対応（Q&A No.30）
- 分類: 依存待ち
- 題名/本文の陳腐化: なし（価値実測後に再開する gate は有効）

### 現状実測
既存 picker と image/PDF upload path は `frontend/src/features/medical-records/components/ImageGalleryFilter.tsx:95-118` にあるが、drop handler はない。DEC-45（`q&a.html:385-404`）は利用者、頻度、現行時間、失敗率を測るまで phase2 を維持する。

### 残作業
現場 owner が対象利用者、月次頻度、一件当たり現行操作時間、失敗率、目標値を測り、投資判断する。

### 次に動くのは
PO / 現場業務責任者。需要を agent の印象で補わない。

### 着手プラン
着手不能。解除入力は上記 metrics と再開批准。再開時は既存 upload mutation を再利用し、finalized chart の変更防止、drag keyboard alternative、file validation を test-first で packet 化する。

### 回答起案
`q&a.html#dec-45` を維持し、`q&a.html#decision-pack-issue-readiness-ops-20260801` の #235 行で A（価値実測後に再開）/B（scope 除外）の承認を待つ。

## Issue #249 — [FEAT] 検査機能 — Dr.ワン相当の院内検査結果管理の内製（旧システム Drive 資料に基づく仕様整理）
- 分類: 着手可能
- 題名/本文の陳腐化: あり（R-3 duplicate と R-5 定性基準表示は既に着地）

### 現状実測
R-3 は commit `3a6f1ccd8` と `backend/internal/medicalrecord/lab_import_repository.go:172-269`、R-5 は `239a8a736` と `frontend/src/features/examinations/components/ExamPivotTable.tsx:49-115` に着地。反面、初回 confirm が親を先に confirmed にして item replace を自己拒否する順序（`examination_service.go:239-299`）、confirmed delete の status guard 欠落（`:588-625`, `examination_repository.go:155-166`）、status conflict の 400 応答（service `:255-256,376-377`）が残る。parent examination の create/update/confirm/delete audit event と actor 伝播も無く、現行 audit は item 実削除に限られる（service `:506`、`backend/internal/model/audit_log.go:92`、handler `examination_handler.go:191`）。

### 残作業
TASK-026 で confirm transaction 順序、confirmed update/delete の 409、actor 伝播、parent create/update/confirm/delete の before/after audit と rollback を P0 修正する。TASK-027 は手動検査の edit と `confirmed -> completed` 確定解除、TASK-031 は保存済み snapshot の print、TASK-032 は lab import job の compensating revert とし、examination の確定解除と import job 取消を別経路にする。項目×動物種×測定系の range は臨床入力待ち、外部 file/crosswalk と `auto_commit` は実形式・停止・通知・audit・手動 fallback が揃うまで保留。テストは未実測。

### 次に動くのは
実装エージェントが TASK-026、次に依存しない小 slice として TASK-027/031/032。臨床責任者と外部連携 owner は並行して値と運用 contract を供給する。

### 着手プラン
TASK-026 は RED「初回 confirm 成功」「confirmed update/delete 409」「create/update/confirm/delete の actor + before/after audit」「audit 失敗時 rollback」から開始し、clinic-scoped 元 status lock→items/range 検証→replace→最後に confirmed→audit を同一 tx に置く。未実測コマンドは `docker compose exec -T backend go test -p 1 ./internal/medicalrecord/... -run 'Test.*Examination.*(Create|Update|Confirm|Delete|Audit)|Test.*ReplaceItems' -count=1`。既着地回帰は `TestLabImportExaminationService_PersistExam_SameDayDifferentContentNotDuplicate`, `TestLabImportExaminationService_PersistExam_FullIdenticalContentIsDuplicate`, `TestLabImportDuplicateCheckerDB_IsDuplicate_FullIdenticalOnly`, `transforms.test.ts`, `ExamPivotTable.test.tsx` を対象にする。

### 回答起案
技術 contract は `q&a.html#dec-53` と確定解除/import 取消を分ける `q&a.html#dec-57`。臨床 range と自動化 enable は `q&a.html#decision-pack-issue-readiness-clinical-20260801` に残し、値を推測しない。

## Issue #250 — [DELIVERY] 旧Accessデータ移行 — stage-import拡張・rehearsal・cutover
- 分類: 依存待ち
- 題名/本文の陳腐化: あり（`stage-import` は退役し canonical consumer は `csv-import`）

### 現状実測
consumer は `backend/cmd/csv-import/**`, `backend/internal/csvimport/**` と `docs/ops/deploy/CLINIC_CSV_IMPORT.md:3-18,48-58` に着地し、commit `377803248`, merge `3c1d2d650`, retirement `3de888ad0` は到達可能。formal producer bundle は全 table/payment graph/crosswalk が揃っておらず、current rehearsal は partial で formal handoff ではない。

### 残作業
Source data owner が complete/PASS bundle、payment graph、crosswalk、business reconciliation を提供する。old_db の責任範囲は CSV+manifest handoff までで、AnimalEkarte DB への import や partial CSV の seed 転記を行わない。

### 次に動くのは
先方 / Source data owner。完全 bundle が届くまで consumer の再実装や apply は不要。

### 着手プラン
着手不能。解除入力は formal complete manifest と全対象 CSV。受領後 agent は dry-run/verify を実施し、#253/#254/#255 gate 後に USER が cutover を承認する。

### 回答起案
`q&a.html#decision-pack-issue-readiness-ops-20260801` の #250 行。推奨は consumer 完了を維持し、partial rehearsal を current release readiness と扱わないこと。

## Issue #252 — [OPS] 各院の締め時間設定値の投入 — 全院を城東と同値で確定投入（PO裁定 2026-07-15）
- 分類: USER 専権
- 題名/本文の陳腐化: あり（値投入は有効だが、standard PATCH の validation/audit gap が未記載）

### 現状実測
Issue は OPS_ONLY。standard update は `backend/internal/clinic/closing_settings_service.go:141-165` の read-modify-save、repository は全設定列を upsert する（`backend/internal/clinic/clinic_settings_repository.go:50`）。special period 用 validation（service `:350-364`）に相当する standard boundary validation、application audit、actor/transactor 配線、row lock/CAS がなく、並行 partial PATCH は lost update を起こし得る。既存 test は `closing_settings_service_test.go:206-255`。これは投入作業と別の technical gap である。

### 残作業
USER が批准済み設定との差分だけを各院へ投入し、preview と過去履歴非再計算を確認する。別途 TASK-028 で standard PATCH の時刻関係 validation、clinic-scoped lock/CAS、actor 伝播、fail-closed な transaction-bound audit を追加する。

### 次に動くのは
Issue 本体は USER（Clinic billing operations owner）。TASK-028 はエージェントが並行着手可能で、実値投入を代行しない。

### 着手プラン
TASK-028 は invalid time order、並行 partial PATCH の lost-update 防止、actor 付き before/after audit、audit dependency/failure rollback、clinic scope を RED から固定する。service は clinic-scoped row/advisory lock または CAS の一方式を採り、update/audit を同一 transaction に置く。未実測コマンドは `docker compose exec -T backend go test -p 1 ./internal/clinic -run 'TestClosingSettingsService_UpdateStandard.*(Concurrent|Audit|Rollback|Validation)' -count=1`。USER apply は別 window/runbook。

### 回答起案
設計は `q&a.html#dec-54`、実投入は `q&a.html#decision-pack-issue-readiness-ops-20260801` の #252 行。technical fix を OPS completion と混ぜない。

## Issue #253 — [DELIVERY] 本番環境整備 — CI/CD・監視・DB backup/restore gate
- 分類: USER 専権
- 題名/本文の陳腐化: あり（本文が「最新」とする run `29490725904` は古い。未完 gate 自体は継続）

### 現状実測
Cloudflare workflow と production runbook は `.github/workflows/backend-deploy.yml`、`frontend-deploy.yml`、`docs/ops/infra/production/runbook.md:9-23,184-196` にある。production は未構築、GitHub billing/spending、Environment、required reviewer、required CI green、restore/rollback rehearsal が未証明。read-only 実測コマンド `gh run view 30648745802 --json databaseId,workflowName,status,conclusion,url,headBranch,headSha,jobs --jq '{databaseId,workflowName,status,conclusion,url,headBranch,headSha,jobs:[.jobs[]|select(.name=="Backend")|{name,status,conclusion}]}'` の逐語出力は `{"conclusion":"failure","databaseId":30648745802,"headBranch":"main","headSha":"70da134ad2fc48aa40e8dfc0de88b30e7b90a9e8","jobs":[{"conclusion":"skipped","name":"Backend","status":"completed"}],"status":"completed","url":"https://github.com/MinoruSoga/AnimalEkarte/actions/runs/30648745802","workflowName":"CI"}` で、green artifact ではない。

### 残作業
Budget/contract owner が billing、Production Environment、reviewer、必要設定を用意し、Release owner が required CI、deploy、health、backup restore、rollback を実測する。AWS 経路は戻さない。

### 次に動くのは
USER（Production release authority / Budget owner）。外部課金・設定・production 操作が解除条件。

### 着手プラン
着手不能。解除入力は billing 復旧と Environment/reviewer 完了。受領後 agent は workflow 差分と非機密証拠骨格を整え、実操作は USER 承認 window で行う。

### 回答起案
`q&a.html#decision-pack-issue-readiness-ops-20260801` の #253 行。推奨は USER prerequisite を先行し、latest required CI green と restore/rollback rehearsal を closure gate にすること。

## Issue #254 — [OPS] 納品前の全業務シナリオ通し確認 — 開発側デモ環境でのUAT代行（PO裁定 2026-07-15）
- 分類: USER 専権
- 題名/本文の陳腐化: あり（agent evidence skeleton は着地済みで、残りは human observation）

### 現状実測
統合 report は `docs/ops/testing/scenarios/reports/2026-07-31-local-issue-254.md:10-11,41-56,98-211,234-250`。commit `ffba58aec`, `deb707684`, `7fca84eef` は到達可能。認証用 environment names は unset、five flow、DB/audit、real LINE/LIFF、sign-off は PENDING。既存 TASK-023 が human residual を追跡する。

### 残作業
USER が安全な channel で認証環境を注入し、QA が five flow、DB/audit、real LINE/LIFF、全 FAIL disposition、現場 sign-off を一回の isolated UAT で取得する。

### 次に動くのは
USER / QA / 現場責任者。agent は証拠骨格を完成済みで、human observation を代筆できない。

### 着手プラン
着手不能。解除入力は認証環境と named QA window。将来コマンドは `cd frontend && ./scripts/run-e2e.sh e2e/clinical-flows.spec.ts e2e/examinations-flow.spec.ts e2e/accounting-flow.spec.ts e2e/reservations-smoke.spec.ts e2e/trimming-flow.spec.ts e2e/line-reservation-flow.spec.ts`。実機/DB observation は同 report checklist を使う。

### 回答起案
`q&a.html#decision-pack-issue-readiness-ops-20260801` の #254 行。推奨は新規 packet を作らず TASK-023 の一回の UAT bundle に集約すること。

## Issue #255 — [OPS] スタッフアカウントの一括発行と役割別権限設定（スタッフ一覧の提供待ち・ブロック中）
- 分類: USER 専権
- 題名/本文の陳腐化: あり（roster は受領済みで「一覧待ち」は解消）

### 現状実測
Issue コメントは roster 受領済みを示す。preflight/apply は `backend/cmd/staff-provision/main.go:127-153` と `backend/internal/staff/staff_provisioning.go:534-635`、runbook は `docs/ops/deploy/STAFF_ACCOUNT_PROVISIONING.md:17-23,70-139`、実装 commit は `d16ad0c79`, `53a8e86f3`。preflight は zero-write、apply は transaction/advisory lock/receipt を持つ。

### 残作業
DEC-49〜52 の email、clinic mapping、leave/contractor、role→group 方針に従い、USER が repo 外 manifest と認証情報を作成する。具体的な本人性、所属、権限、配布、authorized actor、apply は USER-only。テストは未実測。

### 次に動くのは
代理裁定後は USER（Client identity/access owner）。agent は安全な input を受領した後の preflight と非個人 receipt review のみ支援可能。

### 着手プラン
着手不能。解除入力は explicit per-person clinic/group/status mapping と authorized actor。未実測コマンドは `docker compose run --rm --no-deps --entrypoint '' -T -v "$(pwd)/backend:/app" backend go test -p 1 ./internal/staff -run 'TestStaffProvisioning' -count=1`。apply は USER 承認後のみ。

### 回答起案
`q&a.html#dec-49`, `#dec-50`, `#dec-51`, `#dec-52` で委任可能な四論点を代理裁定。実 roster、本人性、認証情報、apply は `q&a.html#decision-pack-issue-readiness-ops-20260801` に残す。

## Issue #256 — [OPS] 操作マニュアル・手順書の整備 — 操作研修は納品後実施（PO裁定 2026-07-15）
- 分類: USER 専権
- 題名/本文の陳腐化: あり（後続 manual audit と画像差し戻し、reachable history assessment を未反映）

### 現状実測
主要業務 manual は `docs/delivery/OPERATION_MANUAL.md:19-30` と `frontend/src/features/manual/lib/manual-index.ts:51-61,86-87` に着地。`reports/2026-07-31-task-024-manual-audit.md:132-180` は FAQ no-add と、個人情報混入により置換画像を差し戻した経緯を記録する。unsafe/restore history は current refs から到達可能だが、回収導線となる path/hash は本 dossier に再掲しない。画像内容は本 unit で開いておらず、外部露出範囲は未実測。既存 TASK-024 が残差を追跡する。

### 残作業
Privacy/Repository owner が履歴到達性・露出範囲・対応方針を決める。clean-demo で差し戻し画像を再撮影し、個人情報非混入の受領検査、全対象画像の current UI/本文突合、named Documentation owner sign-off、manual flow E2E を完了する。

### 次に動くのは
USER（Privacy/Repository owner と clean-demo environment owner）。履歴修復は破壊的判断、visual sign-off は人の責任である。

### 着手プラン
着手不能。解除入力は履歴対応方針と clean-demo 環境。受領後は TASK-024 の手順で再撮影・privacy review・scoped Vitest/E2E を実施し、画像内容や識別子を evidence に転記しない。

### 回答起案
`q&a.html#decision-pack-issue-256-manual-privacy`。推奨は履歴方針承認、clean-demo 再撮影、privacy review、全対象 sign-off を close 条件にし、承認前は画像・履歴を変更しないこと。

## Issue #257 — [OPS] 本番切替（Go-live 2026-07-27）— 切替手順書・切り戻し基準・直後サポート体制
- 分類: USER 専権
- 題名/本文の陳腐化: あり（題名の日付と runbook 内 timeline は経過済み）

### 現状実測
`docs/delivery/GOLIVE_RUNBOOK.md:1-6,10-42,78-124` は過去 window と未確定 prerequisite/owner/contact を含む。`docs/ops/infra/production/runbook.md:9-23,184-196` は production 未構築と CI/billing/restore/rollback gate を示す。Issue は OPEN、実 cutover evidence はない。

### 残作業
#250/#253/#254/#255 と security/closing-setting gates を完了し、新 window、Go/No-Go authority、technical operator、client contact、support window、notification、rollback/restore evidence を確定する。

### 次に動くのは
USER（Go-live authority / Client operations owner）。過去日付を agent が新日付へ推測変更せず、production 実操作も行わない。

### 着手プラン
着手不能。解除入力は全 prerequisite green と新 window/owner/contact。承認後に runbook を current state へ更新し、production checks は USER window で実測する。

### 回答起案
`q&a.html#decision-pack-issue-readiness-ops-20260801` の #257 行。推奨 A は全 gate green 後に新 window へ再 schedule、対案 B は正式延期/中止と暫定責任の明文化。

## Issue #258 — [OPS] 納品ドキュメントの整備 — 管理者設定手順・運用手順・システム構成概要
- 分類: 判断待ち
- 題名/本文の陳腐化: あり（repository 由来の三領域は同期済み、契約 owner 入力だけが残る）

### 現状実測
`docs/delivery/DELIVERY_PACKAGE.md:3-16,217-240` は Production 未構築、repo-derived 同期済み、U1〜U13 の USER 入力を分離する。#258 contract pack の正本は既存 `q&a.html:507-521` の U1〜U12。commit `b79749095`, `c30480a5d`, `1ee09266c` は到達可能。

### 残作業
Contract owner/Client/#253 operations owner が U1〜U12 に出典、値、責任者、発効日を記入し、production/backup/support/monitoring evidence を揃える。料金、契約、認証情報、production 事実は agent が補完しない。

### 次に動くのは
契約責任者 / クライアント / USER。repository docs の追加実装では解除できない。

### 着手プラン
着手不能。解除入力は U1〜U12 の全欄と一行承認。入力後 agent は値を複製せず delivery package と #253 evidence の整合だけを review する。

### 回答起案
既存 `q&a.html#decision-pack-delivery` を使用。推奨 A は client ownership/handover、代案 B は developer ownership の料金・期間・責任・終了時移管を明示契約すること。

## Issue #259 — [FEAT] Lステップ連携の再開 — Write API再有効化＋cron配線（納品後対応・先方API有効化待ち）
- 分類: 依存待ち
- 題名/本文の陳腐化: あり（Write dual gate と cron/private scheduler は配線済み）

### 現状実測
deploy gate は `backend/internal/infra/lstep/client.go:22-25,72-85` で OFF 時に `ErrWriteDisabled` + HTTP zero、clinic gate は `backend/internal/lstep/lstep_delivery_trigger_service_test.go:838-841,895-897` で `is_sync_enabled=false` を意図した skip/noop とする別契約である。cron は `backend/wrangler.jsonc:97-102`、plan は `backend/worker/scheduled-jobs.ts:30-34`、private routes は `backend/cmd/api/batch_scheduler.go:49-63`。commits `4ea46a23b`, `9fe15c303`, `c03c792f1`, `ebfc3755e` は到達可能。三 docs はこの二 gate を混同している。

### 残作業
先方 enablement、USER deploy/clinic gate、clinic/master 整合、STG 少数 live send、cron fire、停止/rollback evidence が closure gate。TASK-029 は docs で deploy OFF の error/HTTP-zero と clinic OFF の intentional skip/noop を分離するだけで、外部 enablement や close を代替しない。テストは未実測。

### 次に動くのは
Issue 本体は先方 / USER / Product-contract owner。agent は TASK-029 を並行実施可能。

### 着手プラン
TASK-029 で三 docs を「deploy gate OFF = disabled error + HTTP zero」「clinic gate OFF = intentional skip/noop」と別記し、runbook と source/test link を固定する。将来の未実測回帰は `docker compose exec -T backend go test -p 1 ./internal/infra/lstep ./internal/lstep ./cmd/api` と Worker scoped tests。実送信は行わない。

### 回答起案
scope は `q&a.html#dec-55`、外部/契約/実送信は `q&a.html#decision-pack-issue-readiness-ops-20260801` の #259 行。write/cron を再実装しない。

## Issue #260 — [PLAN] 3セッション開発計画（7/27納品）— 正本
- 分類: USER 専権
- 題名/本文の陳腐化: あり（日付、基準 HEAD、三セッション、closed Issue、退役経路、migration 在庫が historical）

### 現状実測
Issue は OPEN。本文の複数 Issue は既に CLOSED、`stage-import` は `3de888ad0` で退役し現行は `csv-import`。`3-session-agent.html:64-76,241-263` は commit `cf1e08113` で lightweight Issue view と正本境界へ再構築済み。live open 集合と view 集合は調査時に一致した。

### 残作業
本 dossier と lightweight view が個別 Issue を追跡できることを批准し、USER が historical close する。継続する場合だけ、日付や session 数ではない客観的 exit 条件を新たに提示する。

### 次に動くのは
USER（Delivery/Project owner）。docs reconciliation は本 unit で完了するが、GitHub close は外部書込み境界。

### 着手プラン
本 unit で `q&a.html#dec-56` と view action を同期する。USER close 後、`gh issue view 260 --json state,title,updatedAt` と live/view set diff を再実測し、close された行を view から削除する。

### 回答起案
`q&a.html#dec-56` で historical retirement を代理裁定し、実 close は `q&a.html#decision-pack-issue-readiness-ops-20260801` の一行 USER 承認に残す。

## Issue #261 — [TRIAGE][DELIVERY] 臨床安全・画面仕様ギャップのPO決裁
- 分類: USER 専権
- 題名/本文の陳腐化: あり（blanket 未決表と「q&a 分類未達」は DEC-41/47 後の実態と矛盾）

### 現状実測
DEC-41 は `q&a.html:274-301`、DEC-47 Q2 は `:446-450` にあり、新しい臨床値正本を作らない。SD-4 は `backend/internal/inventory/inventory_request.go:34-83` と commit `5d38e3e48`、SD-19 は `frontend/src/hooks/use-pet-vaccinations.ts:13-45` と `4208f12e4`、SD-14 wiring は `backend/internal/lstep/line_link_service.go:180-220` と `2e4808b55`。死亡 guard helper は各領域へ導入済みだが、trimming の detail create/update は request に `pet_id` がある場合だけ検証し、算出済み `finalPetID` を常時検証しない（`trimming_service.go:490,646`）。pet_id 省略時に既存予約の死亡ペットが通り得て、経路別回帰もない。

### 残作業
TASK-030 で trimming detail create/update の `finalPetID` を常に死亡確認し、pet_id 省略・明示置換の両経路で zero write/audit を固定する。trimming create の既存 guard 回帰と stale `phase2.html:206` 同期も行う。USER は DEC-41 批准、対象 DB 履歴に応じた OPS-2、0-rule audit、real LINE/LIFF、対象環境 runtime を一 bundle にする。static source/test existence は runtime evidence ではない。

### 次に動くのは
Issue 全体は USER/PO/DB/QA。agent は TASK-030 を並行着手可能だが、それだけで #261 完了を宣言しない。

### 着手プラン
TASK-030 は detail create/update で request の `pet_id` が nil でも予約由来 `finalPetID` が死亡なら拒否する RED、明示 pet replacement、通常 create の RED を追加する。GREEN では `finalPetID` を business write 前に必ず検証し、拒否時 repository write/audit zero を確認する。未実測コマンドは `docker compose exec -T backend go test -p 1 ./internal/trimming -run 'TestTrimmingService_.*Deceased' -count=1`。OPS-2 は `backend/migrations/CLAUDE.md:54-63` に従い、統合前 DB へ一般的な no-reset upgrade を指示しない。

### 回答起案
既存 `q&a.html#dec-41` と `#dec-47` を維持し、`q&a.html#decision-pack-issue-readiness-ops-20260801` の #261 行で批准/override と OPS evidence を一行入力に残す。

## Issue #284 — [QA] line-reserve（LIFF予約）Noto Sans JP 実機フォント確認 — 3実機（試験環境・実機の受け渡し待ち）
- 分類: 依存待ち
- 題名/本文の陳腐化: あり（環境・実機待ちは有効だが、`FE-refactor.md` 退役前提は既に完了）

### 現状実測
font source は `frontend/line-reserve/index.html:7-12`, `src/main.tsx:5`, `src/index.css:17-23` にあり、commit `861f2363e` は到達可能。`FE-refactor.md` は `a4c17ea9` で削除済み。`frontend/line-reserve/src/App.test.tsx:118-232` は遷移を検証するが rendered font/FOIT/fallback を証明しない。実機結果は未実測。

### 残作業
QA environment owner と device custodian が許可済み試験環境と iPhone/Android/iPad を引き渡し、QA が cold/warm/offline、font file HTTP、computed family、Rendered Fonts、clip、fallback operation を確認する。artifact は個人情報と認証値を除去する。

### 次に動くのは
USER 側 environment/device owner、次に実機 QA。source 修正を推測で増やさない。

### 着手プラン
着手不能。解除入力は非秘密の試験範囲と三実機。将来の source regression は `docker compose exec -T frontend npx vitest run line-reserve/src/App.test.tsx` だが、実機証拠の代替にはしない。不具合再現時だけ agent task を起票する。

### 回答起案
`q&a.html#decision-pack-issue-readiness-ops-20260801` の #284 行。推奨は一回の device matrix QA、代案は環境未準備のまま open 維持であり、source の speculative fix はしない。

## Reconciliation summary

- 即着手 packet: TASK-025〜TASK-032。TASK-027/031/032 は TASK-026 後の独立 slice。TASK-033 は臨床承認と migration review 後に、構造化救急投薬記録と #201 欠落時 fail-closed cutover を一体で行う。既存 TASK-023/#254 と TASK-024/#256 は再利用する。
- 代理裁定: DEC-48〜DEC-58。臨床値、認証値、課金/契約、外部 enablement、production/DB/実機操作は確定していない。
- 反証で棄却: #211 の pure dependent child CASCADE は現行 migration rule と専用 test が許容するため、新規修正 task にしない。
- 未実測境界: 本 unit は docs-only。Docker、DB、migration、seed、browser、実機、外部操作を実行しておらず、将来コマンドは readiness plan としてのみ記録した。
