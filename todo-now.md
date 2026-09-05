# Astra品質監査 — 今回の改善作業

## このファイルの位置付け

ユーザーの明示依頼により作成した、今回の監査6件に限定した作業用記録。以下は会話で受領した報告の要約・着手用整理であり、監査全文の逐語転載ではない。
実行状態の正本は引き続きLinear（[作業台帳ルール](docs/work/README.md)）。Issueを確認・登録した際はこの表にIDを紐付け、二重管理を避ける。Linear照会・登録は未実施。F1〜F6はローカル検証後に`main`へ統合し、2026-09-05にfetch後の`origin/main`へpush済み。

- 受領報告: `Completion Report` / Run status `COMPLETE`（静的監査の完了）
- 監査日時: 2026-09-05 06:07–06:13 JST
- 監査対象: `main` / `c41ba8b1c1aef7f8150457dc310fe7f61fca1a75`
- モデル: ユーザーはAstraの報告として提示。報告内では正確なmodel IDはUNKNOWN。
- 転記時のcheckout: `main` / `b61ff163d2d23e4f9195a224d275a828167bdb31`
- 証拠の範囲: 代表的な高リスク経路の静的調査。全関数網羅、テスト実行、実DB再現、STG/PROD確認ではない。
- 指摘のseverity/confidenceは監査者の判定。転記時には6件の現コード再監査を行っていない。実装前にcurrent HEADで再確認する。
- 監査完了、局所テスト成功、実装完了、統合、release-readyは別の状態として扱う。
- push後確認: `main@75aa2b64c`のGitHub Actions CI run `33968954988`とSecurity Scan run `33968955004`がsuccess。CIではWorkflow Contracts、Frontend Build、Frontend Test 2 shards、Frontend集約・coverage ratchetが実行成功し、backend・migration・codegen系はpaths-filterでskip。E2E、fresh CI DB、Playwright、k6、STG/PROD、release readinessは引き続き未証明。

## 着手一覧

現在のLinear上の重複・対応状況は未照会。実装・統合状態は2026-09-05時点のGit履歴で更新した。

| 優先 | ID  | 指摘                                       | 分類                          | Severity / Confidence | この記録上の状態              | 次の一手                        | Linear |
| ---- | --- | ------------------------------------------ | ----------------------------- | --------------------- | ----------------------------- | ------------------------------- | ------ |
| 1    | F1  | 診察プラン更新・削除とカルテ確定の未直列化 | 製品defect指摘                | High / High           | COMPLETE・main/origin統合済み | 実ランタイム境界は別途確認      | 未照会 |
| 2    | F2  | 編集内容と保存versionの不一致              | 製品defect指摘                | High / High           | COMPLETE・main/origin統合済み | full E2E等は別途確認            | 未照会 |
| 3    | F3  | CI E2E・負荷テストの認証fixture不足        | verification gap              | Medium / High         | COMPLETE・main/origin統合済み | E2E/performance Actionsは未実行 | 未照会 |
| 4    | F4  | カルテ受入項目表とpayloadの不一致          | docs drift / verification gap | Medium / High         | COMPLETE・main/origin統合済み | 外部受入は別途                  | 未照会 |
| 5    | F5  | 診察所見3欄のlabel/id未接続                | a11y defect指摘               | Low / High            | COMPLETE・main/origin統合済み | 実読み上げは別途                | 未照会 |
| 6    | F6  | 検証skillに廃止packageの例示               | harness drift                 | Low / High            | COMPLETE・main/origin統合済み | runtime変更なし                 | 未照会 |

## 実施体制

- Sol: 指摘と現コードを照合し、1件ずつ着手プロンプトを作る。
- Terra: 再現、実装、局所検証、助言の反映を担当する。
- 別セッションのSol/Terra: 差分と元の事故シナリオを独立確認する。
- Astra: 難しい判断に限定したアドバイザー。F1のロック・transaction・競合テストのレビューを優先候補とする。
- `astra-advisor` スキルで相談資料を絞る。原則1タスク1回、自動起動なし。F2以降は重大な懸念や設計判断が残った場合に利用する。
- F1から順に着手する。独立性を確認して並行化する場合もworktree隔離とclaim protocolを守る。

## F1 — 診察プラン更新・削除とカルテ確定

### 監査の指摘

子行を更新したtransaction Aが監査・commit前で待機している間に、親カルテを確定するtransaction Bがcommitできる。後からAがcommitすると確定内容が変わり得る。Deleteも同様。

監査時の根拠（行番号は監査対象commit基準）:

- `backend/internal/medicalrecord/clinical_plan_service.go:210`: 親行ロックを取らず、子UPDATEの条件式で競合を閉じるという説明。
- `backend/internal/medicalrecord/clinical_plan_repository.go:129,148`: 親のdraft条件を通常のsubqueryで参照。
- `backend/internal/sharedkernel/medical_record_lock.go:18`: 他の子記録向けの親行ロック。
- `backend/internal/medicalrecord/clinical_plan_repository_optimistic_lock_test.go:106`: version競合・既確定親・監査rollbackの既存防御。確定transactionとの重なりを制御するテストは不足との報告。

既存のclinic scope、version CAS、writeと監査の同一transactionを維持する。報告の工数概算は4–8h、専用の使い捨てPostgreSQLが前提。

### 着手案と受入条件

提案task ID: `AE-QA-CLINICAL-PLAN-FINALIZE-SERIALIZATION`
提案claim: `claim/AE-QA-CLINICAL-PLAN-FINALIZE-SERIALIZATION`

本ファイルへの転記はF1の取得・実装着手ではない。実装者は編集前にclaimを確認・取得する。releaseはユーザーのみ。

監査の提案allowlist（実行前に現コードとの適合を確認）:

```text
backend/internal/medicalrecord/clinical_plan_repository.go
backend/internal/medicalrecord/clinical_plan_service.go
backend/internal/medicalrecord/clinical_plan_repository_test.go
backend/internal/medicalrecord/clinical_plan_repository_optimistic_lock_test.go
backend/internal/medicalrecord/clinical_plan_diagnosis2_contract_test.go
backend/internal/medicalrecord/clinical_plan_audit_tx_test.go
backend/internal/medicalrecord/clinical_plan_finalize_concurrency_test.go
```

最後のファイルは新規候補。service変更は直列化に必要な範囲と競合説明の訂正に限定する。

- [x] current HEAD、caller、既存テスト、Linear重複、claim、test DBの安全性を確認する。
- [x] 独立した2transactionとchannel/barrierを使い、Update/DeleteそれぞれのREDを確認する。固定sleepだけに依存しない。
- [x] 子write先行時、子transaction終了前に親確定がcommitできないことを検証する。
- [x] 親確定先行時、子writeがConflictとなり、内容・version・監査が変更されないことを検証する。
- [x] 親→子のロック順序を守り、draft確認・子write・監査を同じtransactionに含め、commit/rollbackまでロックを保持する。
- [x] 単独repository呼出しを含むtransaction参加を確認し、transaction外でロックだけ実行しない。
- [x] version競合、他院拒否、404/409、監査rollback、診断2のset/change/clear、作成時subrecord経路を維持する。
- [x] 独立レビューとscoped検証を完了し、skipや環境不備をPASSにしない。

監査が提案した検証コマンド（未実行。実行者は対象テストが実際に選択されることも確認）:

```sh
docker compose exec backend go test ./internal/medicalrecord -run '^TestClinicalPlan' -count=1 -race -timeout=180s
docker compose exec backend go build ./internal/medicalrecord
docker compose exec backend go vet ./internal/medicalrecord
git diff --check
```

変更Goファイルにはscopedな`gofmt -l`も行う。DB helperはTRUNCATE等を行うため、専用test DBの確認なしに実行しない。migration/resetの権限は別途確認する。UI、他domain、schema、seed、codegen、API変更はF1の範囲外。

### F1 実行結果（2026-09-05）

- 状態: COMPLETE。実装commit `9f80acbed`（`fix: serialize clinical plan finalization`）は`main`およびローカル追跡ref `origin/main`へ統合済み。専用worktree・work branch・`claim/AE-QA-CLINICAL-PLAN-FINALIZE-SERIALIZATION`は削除済み。実装基底は `fb04c3ab7b501b404e3e0d8e320086a2de9062aa`。release readinessは別判定。
- 根本原因と修正: Update/Delete が親 `medical_records` を transaction 内で `FOR UPDATE` せずに child DML と監査へ進んでいた。両 mutation は transaction 先頭で clinic-scoped draft parent を lock し、child DML・reload・fail-closed audit・commit/rollback まで lock を保持する。Delete は既存 clinical_plan NotFound 契約を保つ read-only preflight 後に、lock 後の tx 内再読込を行う。
- RED/GREEN: 修正前の実 PostgreSQL 競合試験は child-first Update/Delete で「finalize completed before waiting for the parent row lock」と失敗し、parent-first だけが通過した。修正後は Update/Delete × child-first/finalize-first の4経路が PASS、`-count=5 -race` でも PASS。finalizer の PostgreSQL backend PID を固定して `pg_stat_activity` の actual `Lock` wait を観測し、goroutine順序だけで判定していない。
- 変更パス: `backend/internal/medicalrecord/clinical_plan_service.go`、`backend/internal/medicalrecord/clinical_plan_audit_tx_test.go`、新規 `backend/internal/medicalrecord/clinical_plan_finalize_concurrency_test.go`、この F1 結果節のみ。
- 検証: one-off Docker runner（専用 worktree bind mount、`${DB_NAME}_test`、`-p 1`）で `TestClinicalPlanFinalizeConcurrency -count=5 -race`、`^TestClinicalPlan -race`、`TestMedicalRecord.*(Finalize|SubRecord|Create) -race`、`go build ./internal/medicalrecord`、`go vet ./internal/medicalrecord`、golangci-lint `./internal/medicalrecord/...` を PASS。詳細は当該実行の Completion Report を参照。
- 独立レビュー: clinic-isolation review は CRITICAL/HIGH/MEDIUM なしで Approve。Go review の HIGH（Delete の missing-plan 時 404/409 drift）と MEDIUM（共有DBの他待機を拾う可能性）は修正し、上記検証を再実行済み。
- Astra: `astra-advisor` 向け相談資料は準備済み。自動起動はせず、手動レビューは未実施。質問候補は親lock順序、Delete preflight/re-read、PID固定の実DB競合証明である。

## F2 — 編集内容と保存versionの不一致

### 監査の指摘

Aがv1を表示し、Bが所見を変更してv2にした後、Aの同一recordのqueryだけが再取得されると、入力内容がv1のまま保存versionだけv2になる。CASを通過してBの変更を上書きし得る。

- `frontend/src/features/medical-records/hooks/use-apply-clinical-plan.ts:59`: hydrateはrecord IDが変わった初回のみ。
- `frontend/src/features/medical-records/hooks/use-medical-record-form.ts:75`: 保存versionは現在のquery dataから取得。
- `frontend/src/features/medical-records/hooks/use-medical-record-save-action.ts:212`: PATCHは所見3欄を常時送信。
- `frontend/src/features/medical-records/hooks/use-medical-record-form.action.test.ts:449`: 既存テストは取得versionの転送を確認。

version未取得時の拒否、409表示、backend CASは既存防御。window focus再取得は無効であり、focusだけを再現条件にしない。工数概算4–8h。

- [x] 同一record IDのqueryをv1→v2へ差し替え、編集中のpayloadと基準versionを確認するREDテストを書く。
- [x] 編集内容と基準versionを同じsnapshotとして管理し、他者更新でversionだけ進めない。
- [x] 409または明示的な競合状態を維持し、他者の所見を古い内容で黙って上書きしない。
- [x] 自身の保存成功、再取得、record切替をhook結合テストで確認する。
- [x] scoped Vitestと独立レビューを実施する。具体的な編集allowlistはSolが現コードから定義する。

### F2 execution result — 2026-09-05

- 状態: COMPLETE。実装commit `30a0bac89`（`fix: keep clinical plan edit snapshot consistent`）は`main`およびローカル追跡ref `origin/main`へ統合済み。専用worktree・work branch・`claim/AE-QA-CLINICAL-PLAN-EDIT-SNAPSHOT`は削除済み。実装基底は `9f80acbed67e980adaeba22013f764335d2e4c2d`。
- 根本原因と修正: queryの最新versionをフォームの旧入力へ直接結合していた。clinical-planの3欄・診断1/2 ID・baseline versionを一つのsnapshotにし、dirtyなremote更新は採用しない。clean更新と完全なown PATCH応答だけをbaselineへ採用し、古いinvalidate refetch、in-flight入力、record切替後の遅延応答を拒否する。
- 変更パス: `use-apply-clinical-plan.ts`、新規`use-apply-clinical-plan.test.ts`、`use-medical-record-form-helpers.ts`、`use-medical-record-form.ts`、`use-medical-record-save-action.ts`、`use-medical-record-save-action.test.ts`、`use-medical-record-form.action.test.ts`、このF2結果節のみ。
- RED/GREEN: v1→編集→同一record v2 rerenderで旧実装は`A の編集値`にversion `2`を送信してRED。修正後、one-off candidate mountで4 Vitest files・80 tests PASS、scoped ESLint PASS、Prettier PASS。
- 独立review: React reviewとTypeScript/state-machine reviewの各HIGH（完全応答、in-flight入力、record switch）を修正し、対応回帰を追加して再検証済み。CRITICALなし、MEDIUMは古いrefetch回帰を追加して解消。Astraはnot invoked: tests/reviewsで未解決のsnapshot/React concurrency判断は残らない。
- 境界: 実装時の証拠はlocal scoped testsのみ。full type-check/build/E2Eとrelease readinessは未実施・未証明。commit・main/origin統合は後続作業で完了した。

### F2 scope closeout — 2026-09-05

- 承認履歴: 元のF2保存promptは7パスのみをallowlistし、`use-medical-record-save-action.test.ts` は当初範囲外だった。後続のユーザー明示承認でこの1パスを8番目として事後承認した。これは元から許可されていたとの遡及的な記載ではない。
- 再照合: 8パスのcandidate union、empty index、retained claim/worktree、candidate-mount SHA-256を確認した。4 Vitest files・80 tests、追加testを含むscoped ESLint、Prettier、diff/new-file whitespaceはPASS。
- fresh closeout review: CRITICAL/HIGHなし。MEDIUMの「不完全な成功応答を直接test化」は、完全応答のみ通知する現行guardの追加防御候補として記録するが、今回のcloseout修正条件ではなく変更しない。Astraはnot invoked。
- closeout後の更新: candidateはcommit `30a0bac89`としてmain/originへ統合済みで、worktree・work branch・claimも削除済み。full type-check/build/E2Eとrelease readinessは引き続き未実施・未証明。

## F3 — CI認証fixtureの準備経路

fresh CI DBの通常seedにはaccountがなく、E2E/負荷テストが認証できない構成との指摘。

- `frontend/e2e/helpers/auth.ts:9`: E2E認証情報必須。
- `.github/workflows/e2e.yml:88`: stack待機後の専用account provisioning不足。
- `.github/workflows/performance-tests.yml:54,90`: local backendに認証する負荷テスト。
- `backend/internal/seedbundle/manifest.go:25`: runtime seedは002_masterのみ。
- `docs/ops/testing/TEST_ARCHITECTURE.md:50`: 認証済みCI E2Eは既知BLOCKED。

認証失敗を成功扱いしない防御はある。監査時点の最新Actions結果はUNKNOWN。工数概算1–2日。

- [ ] Linear上の既知gap・重複Issueを確認する。
- [x] E2Eと負荷テストの必要データ・認可範囲を確定する。
- [x] fresh CI専用の合成tenant/accountと必要最小限のfixtureを準備する経路を実装する。
- [x] 認証smoke・対象API・再実行・同一runner cleanupの配線とfail-closed契約をscoped contractで検証する。
- [x] 実STG account・実患者データを流用せず、workflow変更と外部CI実行の証拠を区別する。

### F3 candidate result — 2026-09-05

- 範囲履歴: 元の契約は10パスだった。ユーザーは実行後に `load-tests/README.md` を第11パスとして将来向けに承認し、local API／spike手順の stale `STG_DEMO_*` 指示を `LOAD_TEST_LOGIN_*` へ修正した。
- 状態: COMPLETE。実装commit `861192357`（`ci: provision synthetic auth fixtures`）をmerge commit `67f63da03`で`main`へ統合済みで、ローカル追跡ref `origin/main`にも含まれる。専用worktree・work branch・`claim/AE-QA-CI-AUTH-FIXTURE`は削除済み。`APP_ENV=test` のsynthetic login seedをE2E auth smokeとlocal k6 jobに明示し、E2Eは`auth-flows.spec.ts`のみを実行する。
- 検証: workflow contract 10/10、candidate mount SHA-256一致、Compose renderと非secret APP_ENV assertion、scoped seedlogin/migrate Go tests、4 MarkdownのPrettier、`git diff --check`をPASS。fresh CI/security reviewはCRITICAL/HIGH/MEDIUMなし。Astraは未使用。
- 証拠境界: full clinical/data-dependent E2E fixtureは未準備で別途BLOCKED。F3固有のE2E/performance Actions workflow、fresh CI DB、Playwright、k6の実行結果はUNREPORTED/UNKNOWNであり、main/origin統合と通常CIのsuccessはそれらの実行成功やrelease readinessを証明しない。

## F4 — カルテ受入項目表のdrift

- `docs/ops/testing/scenarios/FORM-FIELD-INVENTORY.md:18`: `soap_s/o/a/p`、`diagnosis3_type/name`等を列挙。
- `backend/internal/medicalrecord/clinical_plan_request.go:12`: requestは`physical_exam`、`diagnosis_details`、`treatment_policy`と診断1・2。

存在しない欄・wire keyを受入対象にし、実際の保存契約との対応を誤る懸念。inventoryは再構築中と明示され、inventory-to-source gateは未実装との報告。工数概算2–4h。

- [x] カルテ1フォームのUI state→request key→APIを照合する。
- [x] 列挙済み項目を正し、未確認範囲を明示する。診断3の新規実装へ拡張しない。
- [x] 対応表の検査を当該フォームに限定して追加する。

### F4 execution result — 2026-09-05

- 状態: COMPLETE。commit `52d26ecdc`（`test: guard clinical plan field inventory`）を`main`と`origin/main`へ統合済み。専用worktree・work branch・`claim/AE-QA-FORM-FIELD-INVENTORY`は削除済み。release readinessは未証明。
- 変更: `FORM-FIELD-INVENTORY.md`でclinical-planの8項目をUI state→PATCH key→API contractへ対応付け、`check-workflow-contracts.test.mjs`へ当該section限定のdrift gateを追加した。診断3の新規実装には拡張していない。
- 検証: contract test 11/11、candidate mount SHA-256一致、Prettier 3.9.6、`git diff --check`をPASS。独立docs/CI/security reviewで未解消CRITICAL/HIGHなし。push後の通常CIではWorkflow Contractsがsuccess。E2E Actions・Playwright/UATはUNREPORTED。

## F5 — 診察所見3欄のlabel/id

- `frontend/src/features/medical-records/components/ClinicalPlanSection/ClinicalPlanSection.tsx:72`: labelのhtmlForとtextareaのidが未接続。
- `frontend/src/components/shared/CharCountTextarea/CharCountTextarea.tsx:39`: wrapperは受領idを転送する。
- `frontend/src/features/medical-records/components/ClinicalPlanSection/ClinicalPlanSection.test.tsx:56`: 既存テストはgetByDisplayValueで値・disabledを確認。

ラベルクリックによるfocusとプログラム上の関連付けが成立しないとの指摘。読み上げ実挙動は未確認。工数概算1–2h。

- [x] 3欄をgetByLabelTextで取得できることを確認するテストを書く。
- [x] 一意なidとhtmlForを接続し、ラベルクリックのfocusを検証する。
- [x] 値表示・disabledの既存挙動を維持する。

### F5 execution result — 2026-09-05

- 状態: COMPLETE。commit `134d294ef`（`fix: associate clinical plan textarea labels`）を`main`と`origin/main`へ統合済み。専用worktree・work branch・`claim/AE-QA-CLINICAL-PLAN-TEXTAREA-LABELS`は削除済み。release readinessは未証明。
- 変更: 身体検査所見・診断詳細・治療方針の3 textareaへ安定した`id`を設定し、可視labelの`htmlFor`と接続した。
- 検証: deterministic RED後、対象Vitest 3/3、scoped ESLint、Prettier、candidate mount hash、`git diff --check`をPASS。独立React/a11y reviewでCRITICAL/HIGHなし。スクリーンリーダー実機確認は未実施。
- 後続候補（未起票・未着手）: 同componentの「診断カテゴリ」「診断病名」labelも`SearchableSelect` triggerと未接続。F5のtextarea限定scopeには含めず、別タスクとして現HEADで再確認してからID・claimを割り当てる。

## F6 — 検証skillの廃止package例示

`.claude/skills/scoped-verification-gates/SKILL.md:19`に存在しない`./internal/service/...`のlint/vet/test例が残るとの指摘。監査時点では`.agents`側とcanonical側はcmp一致し、単なる同期遅れではない。工数概算1h程度。

- [x] 現時点のcanonical所有元と同期手順を確認する。
- [x] 例示を存在するdomain packageへ修正し、既存同期手順で反映する。
- [x] 参照先存在と差分・同期結果を確認する。製品runtime testは不要。

### F6 execution result — 2026-09-05

- 状態: COMPLETE。commit `d6f2e8ec1`（`docs: align scoped verification examples`）を`main`と`origin/main`へfast-forward統合済み。専用worktree・work branch・`claim/AE-QA-SCOPED-VERIFICATION-SKILL-DRIFT`は削除済み。release readinessは未証明。
- 変更: canonical `.claude/skills/scoped-verification-gates/SKILL.md`から廃止済み`internal/service`例を除去し、Makefileと同じ`golangci/golangci-lint:v2.11.4`および実在するdomain例へ更新した。検証対象を`VERIFY_GO_PACKAGE` / `VERIFY_GO_FILE`として一度確定し、lint/build/vet/test/gofmtで同じ対象を使うfail-closed手順にした。
- 検証: 旧参照0件、対象package/file存在、canonical/generated mirrorの`cmp`・SHA-256一致、sync script、`git diff --check`をPASS。独立pre-commit reviewのMEDIUMを上記target変数方式で解消後、CRITICAL/HIGH/MEDIUMなし。製品runtime testはdocs/instruction-only変更のため不要。

補足: `docs/CODEX-NAVIGATION-GUIDE.md`は監査時に存在しなかったが、実AGENTS.mdには参照もなかった。会話に入力されたECC補足との違いであり、repoの壊れたリンクとして起票しない。

## 監査で確認した防御と検証限界

| 品質軸           | 報告された防御・確認範囲                                                         | gap / 限界                                         |
| ---------------- | -------------------------------------------------------------------------------- | -------------------------------------------------- |
| 分離・認可       | JWT後の現在所属・権限再評価、owner/pet/record/billing scope、Preload・破損FK確認 | 調査経路で新規越境なし。全経路の安全認定ではない   |
| 臨床・会計       | 更新監査rollback、締め後会計の監査欠落rollback、カルテ削除と見積作成の親行lock   | F1                                                 |
| Go/Gin           | route権限、Context、binding、error mapping、5xx記録、commit前再取得              | F1以外の独立findingなし                            |
| React/TypeScript | 編集state・Action・PATCH、権限ref、死亡/確定guard、失敗toast、fieldset           | F2、F5                                             |
| PostgreSQL/GORM  | 親scope、診断master Preload、FK、unique/index、migration checksum拒否            | F1。実DB isolation・query plan・index利用はUNKNOWN |
| テスト・CI       | backend DB shards、race、coverage集約、FE shards、L0–L5分離                      | F3、F4                                             |
| 性能・観測・回復 | 集計clinic条件、N+1静的gate、k6 API、timeout、request ID、shutdown               | F3。latency・通知到達・復旧実績はUNKNOWN           |
| 文書・harness    | 仕様/payload/項目表、coverage/workflow、AGENTS/skill照合                         | F4、F6                                             |

代表証拠（監査報告の参照）:

- `backend/internal/middleware/auth.go:86`
- `backend/internal/pet/repository.go:535`
- `backend/internal/medicalrecord/medical_record_repository.go:197`
- `backend/internal/billing/accounting_service_update.go:59`
- `backend/internal/billing/accounting_repository_tx_atomicity_test.go:157`
- `backend/internal/medicalrecord/medical_record_delete_estimate_concurrency_test.go:57`
- `backend/cmd/api/main.go:214`、`backend/cmd/api/server_runner.go:72`

coverage ratchetは実装・配線ありとの報告（`.github/workflows/ci.yml:365,549`）。backend 87.3、frontend 43.78は設定baselineで、監査時の実測値ではない。patch coverage、domain/capability gate、inventory-to-source gateはproposal/未実装と報告された。readiness scoreを製品品質へ換算しない。

## 次回開始時の確認

- [ ] LinearでF1〜F6の重複Issue・進行中作業を確認し、必要なIDだけ対応付ける。
- [ ] F3固有のE2E/performance Actions workflow、fresh CI DB、Playwright auth smoke、k6を外部実行し、UNREPORTED境界を更新する。
- [ ] 「診断カテゴリ」「診断病名」labelと`SearchableSelect` triggerの未接続を現HEADで再確認し、別タスクとしてscope・ID・claimを確定する。
- [x] F4〜F6を含むローカル`main`をreview後にpushし、CIの実行成功とpaths-filterによる未実行範囲を分離して記録する。

F1〜F6のCOMPLETEは上記commit、ローカル検証、`main`/`origin/main`統合、および明記したCI実行範囲に限る。Linear照会、skipされたbackend系CI、E2E、fresh CI DB、Playwright、k6、STG/PROD、release readinessを意味しない。今後の更新にも対象commit、実行した検証と結果、独立レビュー、未確認事項を添える。
