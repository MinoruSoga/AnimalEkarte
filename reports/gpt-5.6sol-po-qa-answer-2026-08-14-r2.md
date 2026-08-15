### A. 読んだもの / 読めなかったもの

読んだもの:

- [依頼書](</Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/reports/gpt-5.6sol-po-qa-request-2026-08-14-r2.md>)
- [todo-po.md](/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/todo-po.md)
- [todo.md](/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/todo.md) §1〜3・§7
- [q&a.html](</Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/q&a.html>) の current authority、DR フォーム、OPS-1〜18、PO-008
- [product-philosophy.md](/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/docs/product-philosophy.md)
- [fable-po-recommendation.md](/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/docs/work/decisions/fable-po-recommendation.md)
- [UAT FINAL](/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/reports/uat-2026-08-14/FINAL.md)
- [DELIVERY_PACKAGE.md](/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/docs/delivery/DELIVERY_PACKAGE.md)
- staging runbook、residual closeout ledger、phase2、credential・Lステップ・#250 各 runbook
- `AGENTS.md`、`.claude/CLAUDE.md`、`.claude/rules/*`

読めなかったもの:

- `docs/CODEX-NAVIGATION-GUIDE.md` は checkout に存在しない。
- GitHub live state は操作禁止のため取得していない。#98/#99 は `todo.md` §7.1 と residual ledger の CLOSED 記録を採用した。
- PROD/STG への runtime 接続、DB、secret、外部サービスは確認していない。

依頼書の採番には矛盾がある。指定項目は、22 + 5 + 18 + 18 + 19 = **82件**だが、No は1〜81と指定されている。項目を落とさないため、指定 No.1〜81を維持し、採番から漏れた OPS-18 を **No.63a** として独立掲載する。

### B. 再審

再審なし。DEC-40〜68、Fable pack、`todo.md` §7.1〜7.2を維持する。

### C. 今日の最小セット（最大 3）

1. **PO-11 / #201 — DO_NOW**
   - なぜ今か: TASK-033、#261、通常保存 fail-closed cutover の共通解除条件だから。
   - 手順: §E-1を臨床責任者へ1通送る。
   - 完了条件: bundle全列が埋まり、approver role・発効日・opaque referenceがある。
   - やらないこと: 20%を医学的正本化しない。薬量・単位・適用対象を推測しない。TASK-033を先行しない。

2. **staging ← main preflight — DO_NOW**
   - なぜ今か: 実LINE・他環境migration・remote UATの前提だから。ローカルrefでは `staging...main = 4 / 1346`。
   - 手順: §E-7をUSER/release ownerが実施する。
   - 完了条件: staging-only差分、migration/checksum、backup/rollback、remote CI、merge方式が全て承認済み。
   - やらないこと: merge、push、squash、直接merge、STG reset、migration適用。

3. **#256 U13 disposition + close一行 — DO_NOW**
   - なぜ今か: visual/privacy dual sign-offは済み、残る判断セルが限定されているから。
   - 手順: U13を完了または未完として明示し、完了時だけ§E-5を別USER承認へ回す。
   - 完了条件: `U13_status=COMPLETED`、発効日、opaque ref、close承認が揃う。
   - やらないこと: history rewrite、画像・PII・参加者名の記録、未完状態でのclose。

### D. 全件回答表

| No | ID | Verdict | POの答え（1文） | 次の人 | 次の一手 | 完成物（§Eの番号 or 本文） | 空欄に残すセル |
|---:|---|---|---|---|---|---|---|
| 1 | PO-11 / #201 | DO_NOW | canonical bundleを唯一の入力先として依頼し、現行20%は継続推奨に留めて医学的正本化しない（`q&a.html` DEC-48/65）。 | 臨床責任者 | §E-1へ全列を記入する。 | §E-1 | 対象、policy、単位、出典、approver role、発効日、opaque ref |
| 2 | staging←main preflight | DO_NOW | local refの4/1346差分を起点に、staging-only差分・migration・rollback・CIを先に判定する（`todo.md` §7.2）。 | リリース責任者 | §E-7を実施する。 | §E-7 | 実施者role、時刻、CI run、backup/rollback ref |
| 3 | staging←main merge | DO_NEXT | §E-7が全項目greenの場合だけmerge-commit PRで取り込み、squash・直接merge・STG resetは選ばない。 | リリース責任者 | preflight承認後にPRを作成・mergeする。 | `merge=PASS pr=[ ] merge_sha=[ ] ci=[ ] health=[ ]` | PR、merge SHA、CI、health ref |
| 4 | #256 U13 disposition | DO_NOW | U13未完ならopen維持、完了ならno-rewriteのまま別USER close承認へ進める（`todo-po.md`）。 | 納品責任者 | U13状態・発効日・opaque refを確定する。 | §E-5 | U13 status、発効日、close approval、opaque ref |
| 5 | 実LINE UAT（STG） | DO_NEXT | current mainのSTG deploy・health後にのみ、実通知・実token health・LIFF導線・auditを観測する。 | UAT責任者 | §E-8を実施する。 | §E-8 | 対象clinic role、実施時刻、各結果、opaque ref |
| 6 | PO-10 / LINE-R05 presence | DO_NEXT | clinic別presence件数を値非表示で取得し、presenceがゼロになるまでguard除去・DROPを禁止する。 | DB運用責任者 | §E-14を実施する。 | §E-14 | 環境、present別clinic件数、operator role、opaque ref |
| 7 | PO-12 / #249 range依頼 | DO_NEXT | 未承認rangeは「未判定」を維持し、項目×動物種×測定系の承認だけを依頼する。 | 臨床責任者 | §E-2を記入する。 | §E-2 | 項目、種別、測定系、値、単位、出典、承認者、発効日 |
| 8 | PO-13 / #211 | DO_NEXT | 臨床rowとclinic/environment/apply行を分離し、両方が揃うまでlocalを含む実row投入を禁止する。 | 臨床責任者 | §E-3の臨床行を記入し、OPS行を運用責任者へ回す。 | §E-3 | clinical row全欄、target env、operator、dry-run/apply/rollback結果 |
| 9 | PO-16 / #261 | HOLD | #201のopaque refと5項目のruntime結果が揃うまでopenを維持し、#201値は複製しない。 | PO | #201承認後に§E-4を記入する。 | §E-4 | clinical bundle ref、5結果enum、各opaque ref |
| 10 | #254 close | HOLD | local FAIL 0だけでは閉じず、実LINE・token・DB/audit・残件disposition・別承認が全てgreenの場合だけcloseする。 | QA責任者 | §E-6の全条件を実測する。 | §E-6 | window、operator、各結果、sign-off、opaque ref |
| 11 | #256 close | CLOSE_RECOMMEND | dual sign-offを維持し、U13完了・発効日・opaque ref・別USER承認が揃った時点でcloseする。 | PO | §E-5を承認後に使用する。 | §E-5 | U13 status、発効日、close approval、opaque ref |
| 12 | PO-17 他env migrate | DO_NEXT | 対象envを明示し、backup/rollback確認後にUSERが非破壊migrationを実行し、STG/PROD resetはしない。 | 環境運用責任者 | 必要envだけで手動実行し結果を記録する。 | `env=[ ] migration=PASS/FAIL rollback_ready=[ ] opaque_ref=[ ]` | env、結果、operator role、opaque ref |
| 13 | PO-18 / #89・#97 | DO_NEXT | 4系統をrotate・revokeし、必要session無効化・health・scanまで完了するまでIssueを閉じない。 | セキュリティ責任者 | §E-12を実施する。 | §E-12 | 対象区分、window、各結果、owner role、opaque ref |
| 14 | PO-19 / #253 | HOLD | PROD未構築のため、CI/CD・deploy・backup・restore・rollbackを完了扱いしない。 | 本番運用責任者 | U1〜U12とproduction authority確定後に構築計画を実行する。 | `production_env/CI/deploy/health/backup/restore/rollback=[各PASSまたはBLOCKED]` | 各結果、authority role、opaque ref |
| 15 | PO-20 / #257 | HOLD | 全gate green前にGo-live日付を設定せず、旧windowはNo-Goのまま維持する。 | リリース責任者 | #252を含む全gate green後に新windowを設定する。 | `gates=ALL_GREEN window=[ ] authority=[role] rollback_owner=[role] ref=[ ]` | window、各role、opaque ref |
| 16 | #250催促 | HOLD | complete/PASS producer bundle、completed_at、payment graph、crosswalkを受領するまでpartial applyを禁止する。 | 移行元責任者 | §E-9へ回答する。 | §E-9 | bundle ID、source owner role、complete判定、cutover authority |
| 17 | #259催促 | HOLD | 先方enableまで両gate default-offを維持し、enable後も少数STG送信・cron・停止/rollback実測までopenとする。 | 外部連携責任者 | §E-10へ回答する。 | §E-10 | contract owner、enable結果、各gate、acceptance、rollback ref |
| 18 | #284 | DEFER | 許可済み試験環境と3実機を受領した時点だけ再開し、それ以前は新unitを作らない。 | 実機QA責任者 | 受領後にcold/warm/offline matrixを1回実施する。 | `environment=READY devices=3 QA=[PASS/FAIL] disposition=[CLOSE/BUG] ref=[ ]` | environment、device custodian role、結果、ref |
| 19 | TASK-009 他env seed | DEFER | current mainのSTG health後かつ対象env・承認bundleが確定した場合だけ、必要envへ限定してseed適用を検討する。 | DB運用責任者 | bundle・checksum・rollbackを確認しUSER判断へ上げる。 | `env=[ ] bundle=[ ] checksum_gate=[ ] rollback=[ ] approval=[ ]` | env、bundle、結果、approval ref |
| 20 | TASK-021 C/D | DEFER | CはBのclient registryとaccess-log gate後、DはC・STG green・backup・rollback owner後に別々に承認する。 | PO | B→C→Dの順に解除判断する。 | `B evidence=[ ] C approval=[ ] STG=[ ] backup=[ ] rollback_owner=[ ] D approval=[ ]` | 各gate結果・ref |
| 21 | TASK-033 | HOLD | #201 bundle、構造化救急投薬記録、bounded DB plan、fresh DB reviewを一体でgreenにするまで着手しない。 | 臨床責任者 | §E-1を完成させる。 | §E-1 | #201全臨床欄、承認ref |
| 22 | POST-PULL | KEEP_OPEN | migration追加・変更を含むcommitをpullした場合だけ、利用前にUSERが`make migrate`を実行する。 | 開発環境利用者 | 該当pull後に手動実行し結果を残す。 | `pulled_commit=[ ] migration_changed=YES make_migrate=[PASS/FAIL]` | commit、env、結果 |
| 23 | DR-CLINICAL #201 | DO_NEXT | 上限・warningは現行継続を推奨するが、臨床責任者の出典付き承認までは未承認として扱う。 | 臨床責任者 | canonical bundleを一度だけ記入する。 | §E-1 | 全臨床値・policy・出典・承認欄 |
| 24 | DR-CLINICAL #261→#201 | HOLD | #201のopaque approval referenceだけを参照し、独立値が必要なら同じ列を持つ別行を追加する。 | 臨床責任者 | #201承認後に参照承認を記入する。 | `source=#201 opaque_ref=[ ] reference_approved=[YES/NO] independent_row_reason=[ ]` | ref、承認、独立行理由 |
| 25 | DR-CLINICAL #211 | HOLD | provisional packageは未承認のまま維持し、臨床行とOPS行が揃うまで実rowを投入しない。 | 臨床責任者 | §E-3の臨床行を記入する。 | §E-3 | stable row key/field、type/choice/range、unit、source、approval |
| 26 | DR-CLINICAL #249 | HOLD | 承認rangeがない項目は「未判定」を維持し、一般値・別測定系を流用しない。 | 臨床責任者 | §E-2を記入する。 | §E-2 | range全欄 |
| 27 | VACCINE-SPECIES | HOLD | master明示属性だけを権威とし、未承認mappingは候補をsilent非表示にせず要確認として扱う。 | 臨床責任者 | 適合性行を記入する。 | `master_row_ref=[ ] species=[dog/cat/other/both] alias=[ ] source=[ ] approver_role=[ ] effective_date=[ ]` | 全記入欄 |
| 28 | #258 A/B選択 | DO_NEXT | クライアント所有のAを推奨し、開発者保有Bは料金・期間・責任・終了時移管の明示契約が揃うまで選ばない。 | 契約責任者 | §E-11でA/BとU1〜U12を一度だけ記入する。 | §E-11 | A/B、契約責任者、client approver、発効日、opaque ref |
| 29 | U1 Cloudflare | DO_NEXT | Aを推奨し、クライアント名義・請求・移管方針を契約正本に一度だけ記入する。 | 契約責任者 | U1を記入する。 | §E-11 U1 | 契約名義、請求先、移管有無 |
| 30 | U2 PlanetScale | DO_NEXT | Aを推奨し、本番planは負荷・backup要件に基づき契約責任者が選び、価格や保持期間を発明しない。 | 契約責任者 | U2を記入する。 | §E-11 U2 | plan、backup頻度・保持、契約名義 |
| 31 | U3 Vercel | DO_NEXT | Aを推奨し、planとdomain registrar権限をクライアント管理下に置く。 | 契約責任者 | U3を記入する。 | §E-11 U3 | plan、契約名義、registrar権限 |
| 32 | U4 GitHub | DO_NEXT | Aを推奨し、client-controlled organization、最小権限、退任時revokeを運用方針にする。 | Repository責任者 | U4を記入する。 | §E-11 U4 | organization、role、Collaborator方針 |
| 33 | U5 LINE | DO_NEXT | Aを推奨し、各医院が本番channelを所有し、値はsecret管理へだけ投入する。 | 各医院LINE管理者 | U5の方針・結果だけを記入する。 | §E-11 U5 | owner role、設定結果、opaque ref。secret値は書かない |
| 34 | U6 Lステップ | DO_NEXT | Aを推奨し、各医院契約のAPI keyはsecret管理だけで扱い、文書へ複製しない。 | 各医院連携管理者 | U6の方針・結果だけを記入する。 | §E-11 U6 | owner role、設定結果、opaque ref。key値は書かない |
| 35 | U7 support窓口 | DO_NEXT | Aを推奨し、クライアント一次窓口と開発エスカレーション条件を契約で明確化する。 | サポート責任者 | U7を記入する。 | §E-11 U7 | 連絡手段、受付時間、一次対応role |
| 36 | U8 監視通知 | DO_NEXT | Aを推奨し、client-controlled distribution先を使い、実addressはrepo/chatへ書かない。 | 監視責任者 | U8を記入・Cloudflare側検証する。 | §E-11 U8 | address、verification結果、opaque ref |
| 37 | U9 本番backup実測 | HOLD | Aを推奨するが、Production構築後のrestore実測まで未完として維持する。 | 本番運用責任者 | 構築後にbackup・restoreを実測する。 | §E-11 U9 | 頻度、保持、手順、所要時間、結果ref |
| 38 | U10 R2 backup/versioning | DO_NEXT | Aを推奨し、採否・保持・復旧責任を契約責任者が確定する。 | ストレージ責任者 | U10を記入する。 | §E-11 U10 | 採否、保持、復旧方針、承認ref |
| 39 | U11 audit保持 | DO_NEXT | Aを推奨し、臨床・会計・法務要件に基づく保持年数と廃棄方針を先方が承認する。 | データガバナンス責任者 | U11を記入する。 | §E-11 U11 | 保持年数、廃棄方針、承認ref |
| 40 | U12 Production証跡 | HOLD | Aを推奨し、実構築・URL health・rollback証跡が揃うまで空欄・未構築を維持する。 | 本番運用責任者 | production構築後にU12を記入する。 | §E-11 U12 | setup結果、URL health、rollback、opaque ref |
| 41 | annual_visit_count | DO_NEXT | 現行の直近365日rollingを継続推奨とし、clientが承認するまで他指標と統一しない（`q&a.html` PO-008）。 | クライアント仕様責任者 | 現行継続または修正を回答する。 | `annual_visit_count=current_365d_rolling decision=[APPROVE/CORRECT] correction=[ ]` | decision、修正値、承認者、発効日 |
| 42 | annual_amount | DO_NEXT | 現行のFrom/To→Year→preset→未指定全期間という境界を継続推奨とし、365日に自動統一しない。 | クライアント仕様責任者 | 現行継続または修正を回答する。 | `annual_amount=current_priority_rules decision=[APPROVE/CORRECT] correction=[ ]` | decision、修正値、承認者、発効日 |
| 43 | CSV・last-visit/dormant | DEFER | 顧客集計CSVは追加せず、last_visit bucketとLステップdormant判定は別目的のまま維持する。 | クライアント仕様責任者 | CSVが必要なら責任者・目的・削除工程・頻度を新要件で提示する。 | `CSV=NOT_REQUIRED_BY_DEFAULT last_visit_dormant=KEEP_SEPARATE override_reason=[ ]` | override理由、責任者、指標 |
| 44 | L-step通常同期path | HOLD | deploy/clinic両gateを維持し、実2xx前に成功扱いせず、通常add/remove失敗は処理失敗として通知する。 | 外部連携責任者 | §E-10のenable条件を満たす。 | `normal_sync=FATAL_ON_WRITE_FAILURE success=REAL_2XX_ONLY approval=[ ]` | 承認、通知先、rollback ref |
| 45 | L-step cleanup/補助path | HOLD | cleanupは本体削除を止めないbest-effortを維持するが、失敗通知・監査を必須としsilent successにしない。 | 外部連携責任者 | path別の継続・通知・rollbackを承認する。 | `cleanup=CONTINUE_WITH_ALERT auxiliary=NON_FATAL_WITH_ALERT approval=[ ]` | 承認、通知、rollback ref |
| 46 | OPS-1 | DO_NEXT | #89/#97の4系統rotationをUSERが実施し、旧値失効までopenを維持する。 | セキュリティ責任者 | §E-12を実施する。 | §E-12 | 各非機密結果、window、owner role、opaque ref |
| 47 | OPS-2 | DO_NEXT | checksum mismatchのlocalだけ承認済みfresh適用、STG/PRODはresetせず通常migration証跡を取得する。 | DB運用責任者 | 対象envごとに手動実行する。 | `local_fresh=[ ] target_env_migration=[ ] startup=[ ] reset_used=NO ref=[ ]` | env、結果、ref |
| 48 | OPS-3 | DO_NEXT | current mainのSTG health後にSD-9 read-only判定を行い、0行ならclose、hitならbackfill課題へ分離する。 | DB運用責任者 | 承認済み環境で件数だけ取得する。 | `environment=[ ] zero_rule_groups=[count] assigned_staff_affected=[count] disposition=[CLOSE/BACKFILL_REQUIRED]` | env、件数、ref |
| 49 | OPS-4 | DO_NEXT | current mainのSTG health後に実LINE link E2Eを実測する。 | UAT責任者 | §E-8を実施する。 | §E-8 | 実施結果、opaque ref |
| 50 | OPS-5 | DO_NEXT | local PASSを再実施せず、残るDB/audit・実LINE・fixture・最終dispositionだけを補う。 | UAT責任者 | OPS-2/4後に未観測項目だけ実施する。 | `db_audit=[ ] real_line=[ ] fixtures=[ ] residual_disposition=[ ] signoff=[ ]` | 各結果、ref |
| 51 | OPS-6 | HOLD | Production未構築のため値を設定済み扱いせず、構築時に`false`を確認する。 | 本番運用責任者 | Production作成後に設定と表示結果を確認する。 | `environment=production demo_accounts=false result=[PASS/FAIL] ref=[ ]` | 結果、ref |
| 52 | OPS-7 | CLOSE_RECOMMEND | AWS IaCは退役済みとしてcloseを維持し、旧Terraform・ECS手順を再開しない。 | PO | 追加作業せずclosed状態を維持する。 | `OPS-7=CLOSED AWS_IAC=RETIRED DO_NOT_APPLY` | なし |
| 53 | OPS-8 | HOLD | R2公開domainを推測せず、storage ownerの確定後だけSTG/Productionへ設定する。 | ストレージ責任者 | domainと公開方針を承認する。 | `domain_policy=[ ] stg_result=[ ] production_result=[ ] ref=[ ]` | domain、各結果、ref |
| 54 | OPS-9 | DEFER | 非blocking目視としてrelease候補時に一度だけ確認し、NG時だけ個別bug化する。 | UI QA責任者 | 検索・scroll・選択・cascade・disabledを確認する。 | `search/scroll/select/cascade/disabled=[各PASS/FAIL] disposition=[CLOSE/BUG]` | 結果、bug ID |
| 55 | OPS-10 | CLOSE_RECOMMEND | 任意full type-checkを残件化せず、必要な変更時のscoped verificationへ委ねる。 | PO | OPS台帳から完了扱いにする。 | `OPS-10=OPTIONAL_NOT_REQUIRED` | なし |
| 56 | OPS-11 | CLOSE_RECOMMEND | repo外Notion目視を製品release blockerにせず、必要なら文書ownerが個別確認する。 | 文書責任者 | 正常ならclose、異常ならNotion側だけ修正する。 | `three_terms=[PASS/FAIL/NOT_CHECKED] disposition=[CLOSE/EXTERNAL_FIX]` | 結果 |
| 57 | OPS-12 | CLOSE_RECOMMEND | full PHI seedをdefault運用にせず、小さいdemo seedを維持して残件から外す。 | PO | 通常運用をsmall demoのまま確定する。 | `default_seed=SMALL_DEMO full_seed=EXCEPTION_ONLY` | 例外承認refのみ |
| 58 | OPS-13 | DO_NEXT | OPS-2と同一windowで実施してよいが、fresh適用・key coverage・rollbackを独立証跡として残す。 | DB運用責任者 | 対象envでUSERが実施する。 | `fresh_apply=[ ] migration_key_missing=0 checksum=[PASS/FAIL] rollback=[READY/NOT_READY]` | env、結果、ref |
| 59 | OPS-14 | DO_NEXT | staging PR後のremote CI greenとfull coverage artifactをrelease evidenceにする。 | CI責任者 | USERがremote CIを起動・確認する。 | `remote_ci=[PASS/FAIL] coverage_artifact=[ ] ratchet=[PASS/FAIL]` | run/ref、結果 |
| 60 | OPS-15 | HOLD | Production未構築のため実値設定・deployを行わず、U1〜U12と#253後まで待つ。 | 本番運用責任者 | 構築gate成立後に設定・deployする。 | `required_config=[PASS/BLOCKED] deploy=[ ] rollback=[ ]` | 各結果、ref |
| 61 | OPS-16 | HOLD | production相当環境ができるまでscheduler・observability・recovery rehearsalを完了扱いしない。 | 信頼性責任者 | 環境構築後にpause/resume/catchup/alert/recovery/rollbackを実測する。 | `scheduler/log/alert/recovery/rollback=[各PASS/FAIL]` | env、時刻、各結果、ref |
| 62 | OPS-17 | HOLD | LINE code deployとtest channel準備後だけredelivery/error statisticsを有効化し、順序・重複・rollbackを実測する。 | LINE運用責任者 | controlled non-2xxを含むrehearsalを実施する。 | `redelivery/error_stats/duplicate/order/rollback=[各PASS/FAIL]` | channel role、時刻、結果、ref |
| 63a | OPS-18 | DEFER | Production監視責任者が確定した時だけSentry free-onlyを検討し、自動paid upgradeは許可しない。 | 監視責任者 | free plan、PII抑止、停止手段を承認する。 | `plan=FREE_ONLY pii=OFF auto_upgrade=NO project=[ ] ref=[ ]` | project識別子、owner role、opaque ref。DSN値は書かない |
| 63 | #89 / #97 close | KEEP_OPEN | rotation・revoke・session無効化・health・scanが全て完了するまでIssueをopen維持する。 | セキュリティ責任者 | §E-12完了後にclose承認を求める。 | §E-12 | 各結果、close approval、opaque ref |
| 64 | #98 | CLOSE_RECOMMEND | `todo.md` §7.1とledger上はclosed済みなので再作業せず、liveでopenだった場合だけ§E-13を使う。 | PO | state差異があればUSERが確認する。 | §E-13 #98 | live state、必要時のclose ref |
| 65 | #99 | CLOSE_RECOMMEND | `todo.md` §7.1とledger上はclosed済みなので再作業せず、liveでopenだった場合だけ§E-13を使う。 | PO | state差異があればUSERが確認する。 | §E-13 #99 | live state、必要時のclose ref |
| 66 | #201 | KEEP_OPEN | canonical clinical bundleとTASK-033の安全な代替記録経路がgreenになるまでopenを維持する。 | 臨床責任者 | §E-1を完成させる。 | §E-1 | clinical bundle全欄 |
| 67 | #211 apply | KEEP_OPEN | clinical行とOPS行、dry-run、rollbackが揃うまで実applyをlocal含めHOLDする。 | データ移行責任者 | §E-3の両行完了後にdry-runする。 | §E-3 | clinic authorization、env、DB history、operator、各結果 |
| 68 | #249 外部import | DEFER | 外部format・停止・通知・audit・manual fallbackが承認されるまでmanual/default-offを維持する。 | 臨床プロダクト責任者 | 再開条件成立時に新Issueで起票する。 | `clinical_approval=[ ] format_owner=[ ] enable_stop_owner=[ ] audit=[ ] manual_fallback=[ ]` | 全欄 |
| 69 | #250 | KEEP_OPEN | formal producer bundle受領までGo-live gateとしてopen維持し、partial bundleを適用しない。 | 移行元責任者 | §E-9へ回答する。 | §E-9 | bundle ID、complete判定、authority |
| 70 | #252 | KEEP_OPEN | 批准済み値との差分だけをpreviewし、validation・lock/CAS・audit gateがgreen後にUSERが投入する。 | 会計運用責任者 | clinic別差分とwindowを確定する。 | `clinic_ref=[ ] diff=[ ] preview=[PASS/FAIL] operator_role=[ ] window=[ ] ref=[ ]` | 全欄 |
| 71 | #253 | KEEP_OPEN | Production未構築のため、CI・deploy・health・backup・restore・rollbackを実測するまでopen維持する。 | 本番運用責任者 | U1〜U12後に構築・復旧試験を行う。 | `billing/env/reviewer/CI/deploy/health/backup/restore/rollback=[各結果]` | 各結果、authority、ref |
| 72 | #254 | KEEP_OPEN | PASS 96・FAIL 0でもPARTIAL 4・BLOCKED 5があるため、§E-6完了までopen維持する。 | QA責任者 | 実LINE・audit・残件dispositionを補う。 | §E-6 | 各観測、sign-off、ref |
| 73 | #255 | KEEP_OPEN | identity・clinic・employment・permission groupを推測せず、repo外manifestの不明行だけHOLDする。 | identity運用責任者 | restricted manifestをpreflightしUSER applyする。 | `preflight=[PASS/FAIL] apply=[PASS/FAIL/NOT_RUN] distribution=[ ] owner_role=[ ] ref=[ ]` | 結果enum、role、opaque ref |
| 74 | #256 | CLOSE_RECOMMEND | no-rewriteを維持し、U13完了と別close承認が揃った時点でcloseする。 | PO | §E-5を使用する。 | §E-5 | U13、発効日、approval、ref |
| 75 | #257 | HOLD | #252を含む全gate greenとGo/No-Go・support・rollback role確定後だけ新windowを設定する。 | リリース責任者 | gate evidenceを一行ずつ収集する。 | `gates=ALL_GREEN authority/support/rollback=[各role] window=[ ] ref=[ ]` | role、window、ref |
| 76 | #258 | KEEP_OPEN | Aを推奨し、U1〜U12をDELIVERY_PACKAGEへ一度だけ記入するまでopen維持する。 | 契約責任者 | §E-11を完成させる。 | §E-11 | A/B、U1〜U12、承認、発効日、ref |
| 77 | #259 | KEEP_OPEN | 先方enableとSTG live send・cron・stop/rollback証跡が揃うまでopen維持する。 | 外部連携責任者 | §E-10へ回答する。 | §E-10 | 各gate・結果・ref |
| 78 | #260 | CLOSE_RECOMMEND | dated plan hubはhistoricalとしてcloseし、再開はobjective exit criteria付きの新Issue・新DECだけにする。 | PO | liveでopenならUSERがcloseする。 | `#260 close: disposition=HISTORICAL owner_role=[ ] opaque_ref=[ ]` | owner role、close ref |
| 79 | #261 | KEEP_OPEN | #201参照と5項目の結果enum・opaque refが揃うまでopen維持する。 | PO | §E-4を完成させる。 | §E-4 | #201 ref、5結果、close approval |
| 80 | #284 | DEFER | 試験環境と3実機受領を相対triggerとしてphase2に維持する。 | 実機QA責任者 | trigger成立時に新unitを起票する。 | `environment=READY devices=3 result=[ ] disposition=[ ] ref=[ ]` | 結果、ref |
| 81 | #212 / #235 | CLOSE_RECOMMEND | DEC-66/67どおりterminal closeを維持し、再開条件成立時だけ別Issueへ分解する。 | PO | liveでopenなら各Issueをcloseする。 | `#212 close=PHASE2_TRIGGERED_ONLY ref=[ ]; #235 close=VALUE_EVIDENCE_REQUIRED ref=[ ]` | 各close ref |

### E. 完成物本文

#### E-1. PO-11 / #201 臨床責任者への依頼

件名: `#201 臨床承認 bundle 記入依頼（構造再選択なし・値は未記入）`

> DEC-48/65で、安全構造は確定済みです。通常dose保存のtechnical failure停止、missing-data時の最終fail-closed、構造化救急投薬記録、専用permission/default-deny、authenticated actor、transaction-bound audit、append-only correction、history/handoffは再選択しません。  
>   
> 次のcanonical bundle 1行だけを記入してください。現行20%は「現行継続を推奨」とできますが、出典付き承認までは医学的正本として扱いません。  
>   
> - 対象薬剤・動物種・master row reference: `[ ]`
> - 救急・既実施投薬の対象ケース・対象薬剤: `[ ]`
> - 上限policy: `[現行継続 / 修正]`
> - warning policy: `[現行継続 / 修正]`
> - 救急記録policy:
>   - medicine identity: `[ ]`
>   - route vocabulary: `[ ]`
>   - dose/strength/concentrationの単位・必須性: `[ ]`
>   - weight/species snapshot: `[ ]`
>   - reason taxonomyまたはbounded free-text: `[ ]`
>   - 訂正対象・rationale: `[ ]`
>   - create grant対象role/group: `[ ]`
> - 数値ごとの単位。非数値は「該当なし」: `[ ]`
> - guideline・添付文書・院内安全手順: `[ ]`
> - approver role: `[ ]`
> - opaque restricted approval reference: `[ ]`
> - 発効日: `[ ]`
>   
> 実氏名、患者情報、credential、承認資料本文はこの返信やrepoへ記載しないでください。全列が揃うまでTASK-033はHOLDです。

#### E-2. PO-12 / #249 range承認依頼

件名: `#249 検査基準rangeの出典付き承認依頼`

> 未承認rangeは「未判定」を維持し、一般値や別測定系から推測しません。次を項目ごとに記入してください。
> 
> - 検査項目: `[ ]`
> - 動物種: `[ ]`
> - 測定系・機器・method: `[ ]`
> - 下限・上限、または定性規則: `[ ]`
> - 単位。非数値は「該当なし」: `[ ]`
> - 出典: `[ ]`
> - clinical approver role: `[ ]`
> - opaque approval reference: `[ ]`
> - 発効日: `[ ]`
> 
> 全列が揃うまで新unit、range import、外部auto-commitは開始しません。

#### E-3. PO-13 / #211 分離依頼

件名: `#211 provisional package — clinical行とOPS行の分離記入依頼`

> `003_demo`をshared環境へ昇格せず、localを含む実row投入は両行完成までHOLDします。
>
> Clinical行:
>
> - stable row key/field: `[ ]`
> - type/choice/range: `[ ]`
> - 単位: `[ ]`
> - 臨床出典: `[ ]`
> - clinical approver role: `[ ]`
> - opaque clinical approval reference: `[ ]`
> - 発効日: `[ ]`
>
> OPS行:
>
> - target clinic authorization結果: `[PASS / FAIL / BLOCKED]`
> - target environment: `[ ]`
> - DB history分類: `[FRESH / CURRENT_MIGRATIONS / CHECKSUM_BLOCKED / OTHER_REVIEW_REQUIRED]`
> - authorized operator role: `[ ]`
> - dry-run結果: `[PASS / FAIL / NOT_RUN]`
> - apply結果: `[PASS / FAIL / NOT_RUN]`
> - rollback結果: `[READY / NOT_READY / NOT_RUN]`
> - opaque restricted-evidence reference: `[ ]`
>
> clinic ID、staff ID、実manifest、receipt本文、接続情報はrepo/chatへ記載しないでください。

#### E-4. PO-16 / #261 5項目テンプレ

```text
#261 runtime close pack
clinical_bundle_ref=[#201 opaque approval reference]

db_policy_result=[PASS/FAIL/BLOCKED/NOT_RUN] ref=[opaque ref]
authorization_audit_result=[PASS/FAIL/BLOCKED/NOT_RUN] ref=[opaque ref]
real_line_liff_result=[PASS/FAIL/BLOCKED/NOT_RUN] ref=[opaque ref]
target_environment_runtime_result=[PASS/FAIL/BLOCKED/NOT_RUN] ref=[opaque ref]
po_close_result=[APPROVED/REJECTED/PENDING] ref=[opaque ref]

rule=#201の値は複製しない。患者情報・credential・実principal・receipt本文は書かない。
```

#### E-5. #256 close一行

U13未完時は投稿せずopen維持する。close可能時だけ次を使う。

```text
#256 close approval: privacy=SIGNED_OFF repository=SIGNED_OFF policy=NO_REWRITE clean_demo=PASS visual_content_signoff=PASS U13_status=COMPLETED effective_date=[ ] close_approval=APPROVED opaque_ref=[ ]. No PII, roster, image/path/hash, credential, or receipt body is included.
```

#### E-6. #254 close条件一行

全項目がPASSになるまでcloseしない。

```text
#254 remains OPEN until authenticated_uat=PASS business_flows=PASS db_audit=PASS real_line_liff=PASS token_health=PASS residual_disposition=APPROVED final_signoff=APPROVED opaque_ref=[ ]. Current local result PASS=96 PARTIAL=4 BLOCKED=5 FAIL=0 is not closure proof.
```

#### E-7. staging ← main preflightチェックリスト

- [ ] USER/release ownerが最新remote refsを取得した。
- [ ] foreign WIPを含まない隔離worktreeを用意した。
- [ ] `staging...main`の左右commit数を再測定した。現在のlocal ref参考値は `4 / 1346`。
- [ ] staging-only 4 commitsを一覧化し、維持・別移管・競合解消を1件ずつ決めた。
- [ ] main側差分を確認し、staging branchをresetしない方針を確認した。
- [ ] `main`から`staging`への方式をmerge-commit PRとした。squash・直接mergeは禁止。
- [ ] migration追加・変更・削除を一覧化した。
- [ ] migration key/checksum・既適用履歴・PlanetScale object ownership blockerを確認した。
- [ ] STG/PRODでreset・volume削除をしない。
- [ ] migrationはagentが実行せず、承認済みCIまたはUSER操作へ残した。
- [ ] deploy前backup、復元手順、rollback ownerを確認した。
- [ ] required remote CI、deploy、health、smokeの期待結果を定義した。
- [ ] failure時はmergeを止め、staging-only commitや履歴を破壊しない。
- [ ] 全項目の非機密結果enumとopaque evidence referenceを記録した。

完了一行:

```text
staging_preflight=PASS staging_only_disposition=APPROVED migration_checksum=PASS backup_restore=READY rollback_owner=[role] remote_ci_required=YES merge_method=MERGE_COMMIT opaque_ref=[ ]
```

#### E-8. 実LINE STG観測チェックリスト

- [ ] current mainがSTGへ反映済み。
- [ ] workers.dev直行と実URLのhealthが成功。
- [ ] 対象clinicとtest accountが承認済み。患者実データを使わない。
- [ ] 実LINE通知を受信。
- [ ] token/channelのhealth結果を確認。値自体は記録しない。
- [ ] 飼主フォームで紐付けURLを発行。
- [ ] LINEからURLを開きLIFFへ遷移。
- [ ] 対象clinicの正しい飼主へ紐付け成功。
- [ ] duplicate/cross-clinic誤紐付けがない。
- [ ] DB状態とaudit eventをrestricted環境で確認。
- [ ] failure時の停止・unlink・rollback手順を確認。
- [ ] 台帳へは結果enum、role、時刻、opaque refだけを記録。

```text
real_line_stg: notification=[PASS/FAIL] token_health=[PASS/FAIL] liff_link=[PASS/FAIL] isolation=[PASS/FAIL] db_audit=[PASS/FAIL] rollback=[READY/NOT_READY] opaque_ref=[ ]
```

#### E-9. #250 producerへの催促文

件名: `#250 formal producer bundle 完成版の提出依頼`

> consumer側preflightは準備済みですが、partial/rehearsal bundleは適用できません。次を満たすformal bundleを提出してください。
>
> - bundle ID: `[ ]`
> - source owner role: `[ ]`
> - producer status: `PASS`
> - completeness: `COMPLETE`
> - completed billingごとの`completed_at`: `COMPLETE`
> - payment/payment_splits graph: `COMPLETE`
> - source→target crosswalk: `COMPLETE`
> - 対象21表の非機密件数検証: `PASS`
> - formal handoff判定: `[APPROVED / REJECTED]`
> - cutover authority role: `[ ]`
> - opaque restricted bundle reference: `[ ]`
>
> PHI、CSV本文、manifest本文、credentialはIssue・repo・chatへ添付しないでください。不完全な場合は、不足項目と次回提出条件だけを返してください。

#### E-10. #259 先方enable催促文

件名: `#259 Lステップ外部enable確認のお願い`

> deploy gateとclinic gateは現在default-offです。実送信を開始する前に、次の非機密結果だけをご回答ください。
>
> - contract owner role: `[ ]`
> - vendor-side enable: `[CONFIRMED / NOT_ENABLED / BLOCKED]`
> - enable対象environment: `[ ]`
> - IP allowlist等の外部制約: `[READY / BLOCKED / NOT_APPLICABLE]`
> - settings owner role: `[ ]`
> - deploy gate予定: `[ON / KEEP_OFF]`
> - clinic gate予定: `[対象clinicのみON / KEEP_OFF]`
> - STG少数live-send acceptance owner: `[role]`
> - cron fire確認: `[PASS / FAIL / NOT_RUN]`
> - stop/rollback確認: `[READY / NOT_READY]`
> - opaque restricted-evidence reference: `[ ]`
>
> token、API key、LINE User ID、request/response本文は返信へ記載しないでください。enable未確認の間は両gateをOFFで維持します。

#### E-11. #258 U1〜U12記入依頼

件名: `#258 DELIVERY_PACKAGE U1〜U12 一括記入依頼`

> 推奨はA「クライアントがservice account・契約・請求・backup・support・monitoringを所有」です。B「開発者保有」は、料金・期間・責任・終了時移管の明示契約が揃うまで選びません。値の正本は`DELIVERY_PACKAGE.md` U1〜U12だけとし、他の台帳（`todo-po.md` 等）へ複製しません。
>
> - Ownership選択: `[A / B]`
> - U1 Cloudflare: 契約名義 `[ ]`、請求先 `[ ]`、移管有無 `[ ]`
> - U2 PlanetScale: plan `[ ]`、backup頻度 `[ ]`、保持 `[ ]`、契約名義 `[ ]`
> - U3 Vercel: plan `[ ]`、契約名義 `[ ]`、domain registrar権限 `[ ]`
> - U4 GitHub: organization `[ ]`、権限方針 `[ ]`、Collaborator方針 `[ ]`
> - U5 LINE: owner role `[ ]`、本番設定結果 `[ ]`、opaque ref `[ ]`。値は書かない
> - U6 Lステップ: owner role `[ ]`、本番設定結果 `[ ]`、opaque ref `[ ]`。API keyは書かない
> - U7 support: 連絡手段 `[ ]`、受付時間 `[ ]`、一次対応role `[ ]`
> - U8 monitoring: 送信先 `[restricted]`、Cloudflare検証結果 `[ ]`、opaque ref `[ ]`
> - U9 production backup: 頻度 `[ ]`、保持 `[ ]`、restore手順 `[ ]`、実測時間 `[ ]`
> - U10 R2: backup/versioning採否 `[ ]`、復旧方針 `[ ]`
> - U11 audit log: 保持年数 `[ ]`、廃棄方針 `[ ]`
> - U12 Production: setup結果 `[ ]`、URL health `[ ]`、rollback確認 `[ ]`
> - Contract owner role: `[ ]`
> - Client approver role: `[ ]`
> - 発効日: `[ ]`
> - opaque restricted-evidence reference: `[ ]`
>
> U9・U12は Production 構築後（#253）に開発側が記入します。今回は空欄のままで構いません。
>
> 金額、secret、実email、実identity、Go-live日付、receipt本文はこの依頼への返信やrepoへ記載しないでください。U13は#256所有であり、#258へ含めません。

#### E-12. #89 / #97 非機密rotationチェックリスト

- [ ] 対象4区分を、値を表示せずinventory化した。
  - PlanetScale DB credential
  - Cloudflare API/Worker secrets
  - LINE channel secret/access token
  - JWT/INTEGRATION_ENCRYPTION_KEY等
- [ ] 各providerで新credentialを発行した。
- [ ] 承認済みsecret storeへ投入した。
- [ ] 依存serviceを再デプロイまたは再接続した。
- [ ] 新credentialでhealthが成功した。
- [ ] 旧credentialをrevokeした。
- [ ] 適用可能なsessionを無効化した。
- [ ] 旧credentialでのアクセス拒否を確認した。
- [ ] repository・Issue・logの非機密scan結論を記録した。
- [ ] history方針はno-rewriteを維持した。
- [ ] 台帳には値・ciphertext・digest・token・接続文字列を残していない。
- [ ] 各結果enumとopaque restricted referenceだけを保存した。

```text
#89/#97 rotation: db=[PASS/FAIL] cloudflare=[PASS/FAIL] line=[PASS/FAIL] app_keys=[PASS/FAIL] old_values_revoked=[YES/NO] sessions_invalidated=[YES/NO/NA] health=[PASS/FAIL] scan=[PASS/FAIL] history=NO_REWRITE owner_role=[ ] opaque_ref=[ ]
```

#### E-13. #98 / #99 がopenだった場合のclose一文

#98:

```text
#98 close: provider_revocation=CONFIRMED residual_history_risk=ACCEPTED policy=NO_REWRITE rotation_execution_owned_by=#89/#97 owner_role=[ ] opaque_ref=[ ]. No credential value or restricted evidence body is included.
```

#99:

```text
#99 close: legacy_ecs_workflow=ABSENT executable_old_deploy_path=ZERO rollback_sot=#253 owner_role=[ ] opaque_ref=[ ]. The retired AWS path will not be restored.
```

#### E-14. PO-10 presence件数取得手順

1. DB運用責任者が承認済みread-only windowと対象environmentを指定する。
2. 値を表示・exportしない接続経路で、次の集計だけを実行する。
3. raw row、clinic ID、secret値、ciphertext、digest、query result CSVを保存しない。
4. `legacy_present`ごとのclinic件数だけを記録する。

```sql
WITH per_clinic AS (
    SELECT
        clinic_id,
        NULLIF(BTRIM(line_channel_secret), '') IS NOT NULL AS legacy_present
    FROM line_reservation_settings
)
SELECT
    legacy_present,
    COUNT(*) AS clinic_count
FROM per_clinic
GROUP BY legacy_present
ORDER BY legacy_present;
```

5. `legacy_present=true`が1件以上なら、LINE-R05 presence guard除去・column DROPはHOLD継続。
6. ゼロなら次の非機密一行だけを記録する。

```text
LINE-R05 presence inventory: environment=[ ] clinics_with_legacy_presence=0 clinics_without_legacy_presence=[count] result=ZERO_PRESENT operator_role=[ ] opaque_ref=[ ]. No secret, clinic ID, ciphertext, or digest retained.
```

### F. カバレッジ自己点検

- [x] YES — todo-po未了22件を個別に書いた
- [x] YES — DR-CLINICAL 5行を個別に書いた
- [x] YES — U1〜U12を個別に書いた
- [x] YES — PO-008の5行を個別に書いた
- [x] YES — OPS-1〜18を個別に書いた。採番漏れのOPS-18はNo.63a
- [x] YES — open Issueを実行行とclose行に分けて書いた
- [x] YES — §6の14本文を省略せず書いた
- [x] YES — 臨床数値・契約金額・secret・実identity・Go-live日付を発明していない
- [x] YES — agent製品unitを増やしていない

### G. agent に渡す unit

NONE

### H. 発明しなかったもの

薬用量・warning閾値・reference range・ワクチン適合性、契約plan・金額・保持期間、credential・token・DSN、実氏名・email・clinic/staff/patient ID、Go-live日付、STG/PRODの未取得証跡は発明していない。コード変更、GitHub操作、migration/seed適用、secret設定、外部送信は実施しておらず、作業ツリーの既存WIPにも変更を加えていない。

