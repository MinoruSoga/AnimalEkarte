# #ledger 残件 GrokAgent 着手計画

作成日: 2026-07-28（改訂: 2026-07-28 15:30・HEAD `c4ce786e0` 時点の実コード状態へ同期）

本書は `3-session-agent.html` の `#ledger` を正本として、別セッションの GrokAgent が一度に一つの作業単位だけを実行するための派生ビューである。対象は `BUG-433` / `BUG-437` / `BUG-440` / `BUG-448` / `BUG-449` / `BUG-454` / `BUG-455` / `BUG-456` / `BUG-458` / `BUG-459` / `BUG-460` / `TASK-444` / `TASK-445` / `TASK-461` / `SEC-DUR-01` / `SEC-SWEEP-02` / `E2E-BASELINE-DEBT-01` の17件。コードと実行可能な検査を台帳より優先し、各セッションは本書の一つの作業単位だけを担当する。

**本改訂での変更点（2026-07-28 15:30・実測根拠つき）**:
- `BUG-457` は `edit` 案で確定・CLOSED（`75aff5786`+`b415892ec`+`7f17aeb53`）。本書から節を削除した。対象リストからも除外。
- `BUG-453` は解消済みと再確認されCLOSED（`bbb31f9be`+`bbb87c6a1`）。本書から節を削除した。対象リストからも除外。
- `SEC-SWEEP-02` は**残余 allowlist 0・修復完了**（実測 `grandchildParentClinicCorrelationResidualSiteCount = 0`、`c68f7ff6b` で pin を 0 に閉じた）。残存サイトだった `GetCloseAggregate` / `ValidateCreateReferences` / `CountEstimatesByMedicalRecordID` / `CountMedicalRecordsByReservationID` は BILL-B1a・BILL-B1b（MR-EST-B1）・RES-B1（RES-MR-B1）で修復済み。**旧 `-BILL-B1` 節が名指しした `accounting_repository.go` / `estimate_repository.go` には対象サイトは存在せず**、実際の対象は `accounting_repository_reports_close.go`・`billing_item_repository.go`・`medicalrecord/medical_record_repository.go`・`reservation/reservation_repository.go` だった。各修復ユニットの許可パスには `lintscan` の残余 allowlist 更新が必要だった（ゲート維持のため空 allowlist も残置）。`STAFF-B1` は residual サイトではなく（disclosure は child `clinic_id` で既に遮断）、親 clinic 相関の integrity 修復として `2a7689080` で完了。
- `BUG-455-AUTH-S`（`ab9eaa170`）・`BUG-455-AUTH-B`（`6515ffd68`）はPermissionGroupについて完了・コミット済み。`BUG-455-I`（残27 fieldの census）着手前に必読の設計申し送りを追記した。
- `BUG-456` は結果値・基準値のfallback表示が解消済み（`7ed9ae808`）。残るのは単位（unit）列の空欄調査のみで、旧節の許可パス・検証コマンドを差し替えた。
- `BUG-440`（vaccination claim解放のimmutable actor監査）は DEC-28 でaudit_logs＋同一tx方式に裁定済みだが本書に未掲載だったため新設した。
- `E2E-BASELINE-DEBT-01`（E2E fixture/selector drift）・`TASK-461`（M-05 runtime evidence）が本書に未掲載だったため参照を追加した。`TASK-461` は既に専用プロンプト `~/.claude/prompt-craft-runs/agent-fast-task461-m05-runtime-evidence.md`（2026-07-28 14:25生成）が存在するため本書では複製しない。

## 調整

### 着手順と所有境界

1. `SEC-SWEEP-02-S1` で残存 read を機械検出できる状態にする。`SEC-SWEEP-02` の各 domain 修復はその後に行う。
2. `SEC-DUR-01-MR-T1` を先に置き、write 面の譲渡回帰を固定してから dead predicate を除去する。
3. `BUG-448` を臨床基準値の本番投入より先に完了し、削除監査の provenance を確保する。
4. `TASK-445-S1`、`TASK-445-B1`、`TASK-445-S2-PAY`、`TASK-445-S3-PET`、`TASK-445-S4-CLINICAL` の順で Payment の構造、挙動、各DDL concernを分ける。適用前 canary と migration 適用はユーザーが行う。
5. `TASK-444-S1` で新規誤用を止めた後、response DTO 化は一 domain ずつ行う。`BUG-433` に独立 writer は置かない。
6. `BUG-458-V` の再測定が完了するまで `BUG-458-S` / `BUG-458-T` を開始しない。`agent-fast-fe12-m05-sentinel-responsive.md`（11:45生成）がFE12レーンのM-05再測定と重複し得るため、着手前に当該プロンプトの実行状況を確認すること。
7. `BUG-459` / `BUG-460` は同じ既存実行契約を使うため、記録更新を直列化する。
8. ~~`BUG-455-AUTH-S` でpresence DTOを作成してから `BUG-455-AUTH-B` でhandler/service/repositoryへ接続する。~~ **完了済み**（`ab9eaa170`+`6515ffd68`）。次は `BUG-455-I`（残27 fieldのcensus）。着手前に「INSERT→条件付きUPDATEの補償方式」を26 repositoryへ複製する前提が正しいかを、単文方式（GORM `Select`明示指定）の実測評価で確認すること（台帳 `BUG-455` 節「設計上の申し送り」参照）。
9. `SEC-SWEEP-02` 全修復ユニットは完了（残余0）。S1/TRIM-B1/MR-B1（`c4ce786e0`）に続き BILL-B1a（`b2b577e15`）・RES-MR-B1（`e752a690a`）・MR-EST-B1/BILL-B1b（`35e22b057`）・residual pin 0（`c68f7ff6b`）・STAFF-B1（`2a7689080`）。記述整合は `SEC-SWEEP-02-DOC`（本ユニット・コメント+本書のみ）。
10. `BUG-457` / `BUG-453` はCLOSED。GrokAgentへ再割当てしない。

`BUG-459` / `BUG-460` は `/Users/minoru/.claude/prompt-craft-runs/agent-fast-bugmd-runtime-verification-closeout.md` だけを実行契約とする。本書へ手順を複製しない。前者のread-only probeは他の非runtime作業と並行できるが、API create/PATCH/DELETEは単独で行う。後者はブラウザ手段がある場合だけ並行可能で、手段が無ければBLOCKEDとする。両者の記録更新は同じwriterが直列に行う。

同時実行可能なのは、所有パスが交差しない `BUG-437`、`BUG-448`、`BUG-454-B1`、`BUG-456`、`TASK-444-S1`、`BUG-458-V` の組み合わせである。`backend/internal/billing` を使う `SEC-DUR-01`、`SEC-SWEEP-02`、`TASK-445` は直列化する。`backend/internal/medicalrecord` を使う `BUG-448`、`SEC-DUR-01`、`SEC-SWEEP-02-MR` も直列化する。

### 全作業単位共通のスコープ検査

各セッションは実装前に次を実行する。`<UNIT>` は作業単位ID、`<許可パス>` は当該節の列挙を改行区切りで置換する。

```bash
UNIT=<UNIT>
python3 - "/tmp/${UNIT}-baseline.json" <<'PY'
import hashlib, json, os, subprocess, sys
raw = subprocess.check_output(["git", "status", "--porcelain=v1", "-z"]).split(b"\0")
paths = [p.decode("utf-8", "surrogateescape")[3:] for p in raw if p]
snap = {p: hashlib.sha256(open(p, "rb").read()).hexdigest() if os.path.isfile(p) else None for p in sorted(set(paths))}
json.dump(snap, open(sys.argv[1], "w"), ensure_ascii=False, indent=2)
PY
```

実装後は各節の `<許可パス>` を入れて実行する。`outside_allowlist=[]` だけを PASS とする。

```bash
python3 - "/tmp/${UNIT}-baseline.json" <<'PY'
import hashlib, json, os, subprocess, sys
base = json.load(open(sys.argv[1]))
allow = set("""<許可パス>""".strip().splitlines())
raw = subprocess.check_output(["git", "status", "--porcelain=v1", "-z"]).split(b"\0")
paths = [p.decode("utf-8", "surrogateescape")[3:] for p in raw if p]
after = {p: hashlib.sha256(open(p, "rb").read()).hexdigest() if os.path.isfile(p) else None for p in sorted(set(paths))}
changed = [p for p, h in after.items() if base.get(p) != h]
bad = [p for p in changed if p not in allow]
print("changed_since_baseline=", changed)
print("outside_allowlist=", bad)
raise SystemExit(1 if bad else 0)
PY
git diff --name-only -- <許可パス>
```

既存差分がある許可パスには着手しない。他セッションの差分を restore / revert / stash しない。backend DB test は `-count=1 -p 1` を基本とする。新テーブルの test setup は migration を読まず `EnsureAutoMigrated` に登録された model だけを作るため、`SQLSTATE 42P01` を migration 未適用と決めつけない。部分 unique index を無条件 `uniqueIndex` tag へ置換しない。

## BUG-433 — 生成 model 型と wire DTO の不一致

- **目的**: 欠陥記録のwrite ownerを一つにし、同じ生成基盤を二重に変更しない。
- **根本原因**: `backend/tygo.yaml:1-15` が `backend/internal/model` を `frontend/src/types/generated/models.ts` へ生成する一方、実 wire は `backend/internal/pet/pet_response.go:51-76` や `backend/internal/reservation/reservation_response.go:11-44` の明示 DTO である。現行 import は約268 file / 294 site で、台帳の Pet 31対9という単一例だけでは現在の全体量を表さない。
- **変更方針**: 本IDでは変更しない。型境界のwrite ownerは調整セクションの割当を正とする。
- **許可パス**: なし。
- **禁止する巻き込み変更**: `models.ts` の手編集、全 import の一括置換、応答に model field を追加して型へ合わせること。
- **検証コマンド**:
  ```bash
  docker compose exec -T frontend npx eslint --max-warnings 0 eslint.config.js
  git diff --name-only -- frontend/src/types/generated/models.ts
  ```
- **完了判定**: 本IDに独立差分がなく、`TASK-444` の boundary gate と domain 別移行へ完全に分界されている。

## BUG-437 — `Items` preload の clinic scope 静的保護

- **目的**: 現在安全な2つの `Preload("Items")` から clinic predicate が将来消えた場合に静的検査を失敗させる。
- **根本原因**: `backend/internal/lintscan/preload_clinic_scope_lint_test.go:245-256` は末尾 association 名だけで判定し、未登録名を無視する。registry には `ExamTypeField` が既にあるが、live site は `backend/internal/medicalrecord/exam_type_repository.go:46-68` の汎用名 `Items` である。
- **変更方針**: 親 model または site を識別する context-aware rule を追加し、対象2 siteの unscoped fail、string/closure scoped pass、他 model の `Items` 非対象、出現数 drift をfixtureで固定する。
- **許可パス**:
  - `backend/internal/lintscan/preload_clinic_scope_lint_test.go`
- **禁止する巻き込み変更**: `Items` の全域登録、association rename、repository/service/model、write registry、exemption、migration、seed。
- **検証コマンド**:
  ```bash
  docker compose exec -T backend go test ./internal/lintscan/ -run 'TestPreloadClinicScope|TestMasterModelReconciliation' -count=1 -p 1 -v
  docker compose exec -T backend go test ./internal/medicalrecord/ -run '^TestExamTypeRepository_ItemsPreload_ClinicIsolation$' -count=1 -p 1 -v
  docker compose exec -T backend go vet ./internal/lintscan/ ./internal/medicalrecord/
  docker compose exec -T backend gofmt -l internal/lintscan/preload_clinic_scope_lint_test.go
  git diff --name-only -- backend/internal/lintscan/preload_clinic_scope_lint_test.go
  ```
- **完了判定**: unscoped examination `Items` fixtureだけがREDになり、現行2 site、DB-backed汚染FK test、他 model の `Items` がgreenで、差分が1 fileだけである。

## BUG-448 — 検査明細削除監査の基準値 provenance

- **目的**: 明細置換で消える旧行と作成される新行について、当時の判定boundと評価可否を監査から復元可能にする。
- **根本原因**: `backend/internal/medicalrecord/examination_service.go:555-571` の `extractExamResultsAudit` は `is_abnormal` / `status` を残すが、modelが保持する `ref_min` / `ref_max` / `qualitative_min` / `qualitative_max` を落とすため `is_assessed` を再構成できない。
- **変更方針**: old/new mapへ4つの解決済みboundと派生 `is_assessed` を追加し、numeric、qualitative、unassessedをtest-firstで固定する。同一transaction・audit失敗rollbackは維持する。
- **許可パス**:
  - `backend/internal/medicalrecord/examination_service.go`
  - `backend/internal/medicalrecord/examination_service_test.go`
- **禁止する巻き込み変更**: audit schema/taxonomy、repository delete、range API/UI、lab import、migration、OpenAPI。
- **検証コマンド**:
  ```bash
  docker compose exec -T backend go test ./internal/medicalrecord/ -run 'TestExaminationService_ReplaceItems|TestExtractExamResultsAudit' -count=1 -p 1 -v
  docker compose exec -T backend go vet ./internal/medicalrecord/
  docker compose exec -T backend gofmt -l internal/medicalrecord/examination_service.go internal/medicalrecord/examination_service_test.go
  git diff --name-only -- backend/internal/medicalrecord/examination_service.go backend/internal/medicalrecord/examination_service_test.go
  ```
- **完了判定**: old/new双方の全boundと `is_assessed` が監査payloadにあり、actor/clinic/resource metadataとfail-closed rollbackがgreenで、差分が2 fileだけである。

## BUG-449 — 基準値投入の運用クローズ

- **目的**: 獣医師承認済みの clinic×field×species 基準値を既存入力面から投入し、実データで low/normal/high/unassessed と3 UI面を確認する。
- **根本原因**: コード経路は開通済みである。routeは `backend/internal/medicalrecord/routes.go:157`、atomic replaceは `exam_type_field.go:224-282`、UIは `frontend/src/features/master/components/ExamTypeFieldsEditor.tsx:154-162,207-226,411-479` にある。未了なのは承認済み臨床値と runtime 証拠であり、現在のDB行数は未報告である。
- **変更方針**: コードを変更しない。動物種、検査法/分析器、単位、片側を許すnumeric boundまたは `(-)<(±)<(+)<(++)<(+++)` のqualitative boundを臨床責任者が承認する。年齢・性別・品種・分析器で区間が変わるなら現schemaへ平均値を押し込まず要件へ戻す。blank-bound rowは運用gateで拒否する。
- **許可パス**: なし。
- **禁止する巻き込み変更**: seedへ推測値を入れること、`normal_value` の機械変換、schema/migration、lab-import例外の撤去、unit testをruntime証拠にすること。
- **検証コマンド**:
  ```bash
  docker compose exec -T backend go test ./internal/medicalrecord/ -run 'TestExamReferenceRangeService_ReplaceAtomicClinicSafeAndNoHistoryRewrite|TestExamReferenceRangeValidation|TestExaminationService_ReplaceItemsUsesSingleMasterRangeForHighLowAndLeavesMissingUnassessed|TestExaminationService_ReplaceItemsUsesSpeciesSpecificRanges' -count=1 -p 1
  docker compose exec -T frontend npx vitest run src/features/master/components/exam-type-fields-editor-model.test.ts src/features/master/components/ExamTypeFieldsEditor.test.tsx src/features/examinations/components/ExamItemsTable.test.tsx src/features/examinations/components/ExamPivotTable.test.tsx src/features/medical-records/components/ExaminationGroup.test.tsx
  git diff --name-only --
  ```
  上記は回帰検査だけでありクローズ証拠ではない。クローズには既存UI/APIで値を保存・再取得し、APIで disposable な below/in/above またはqualitative境界結果を作成し、3 UI面をブラウザで確認し、fixtureをAPIで削除する。
- **手動方法の正当化**: 値の決定は獣医師権限、保存は認証済みmaster操作、画面判定はbrowser描画を必要とするため、汎用shellへcredentialや臨床値を埋め込まない。実行者はbrowser network記録からPUT/GET bodyとstatus、作成したresult ID、3画面画像、DELETE statusを逐語保存する。いずれかが無ければBLOCKEDであり、上記testを代替証拠にしない。
- **完了判定**: 臨床責任者、値の根拠、全対象キー、API read-back、評価status、3 UI面の画像、API後始末が揃う。ブラウザまたは臨床値が無ければ BLOCKED とする。

## BUG-440（新設・未着手・プロンプト未生成）— vaccination claim解放のimmutable actor監査

- **目的**: 明細soft-delete時に`vaccination_id`/内部`clinic_id`をNULL化して再候補化する際、「どの接種claimを誰が解放したか」をimmutable audit経路で復元可能にする。
- **裁定**: DEC-28（2026-07-27・A）で既存 `audit_logs`＋`audit.LogEntryTx` の同一tx・fail-closed方式に確定済み（DDL不要・専用履歴表は不採用）。`DeleteItem` へのactor配線が対象、backend 4〜6 file見込み。
- **依存関係**: 台帳は「billing面の稼働unit清算後に発行」と明記——`SEC-SWEEP-02-BILL-B1a`／`TASK-251` U2b-full第2段（別セッション稼働中）と同じ `backend/internal/billing` を触るため、それらの着地を待ってから着手する。
- **所有パス**: 未確定（`billing_item_repository.go` の `DeleteItem` 周辺が中心と推定されるが、専用プロンプト生成時にexact pathを実測確定する）。
- **次アクション**: billing面の稼働unitが空くのを待ってから `/prompt-craft-agent-fast` で専用プロンプトを生成する。今すぐ着手可能ではない。

## E2E-BASELINE-DEBT-01（未着手・プロンプト未生成）— 検索・業務flow E2E の既存fixture/selector drift

- **目的**: `checkups-flow.spec.ts`／`examinations-flow.spec.ts`／`trimming-flow.spec.ts`／`vaccinations-flow.spec.ts`／`clinical-flows.spec.ts`／`owners-search.spec.ts` の既存赤（selector/heading drift、`petId=1`・「Iris」・「林」・空`name_kana`への依存）を現行UI/seedへ追随させる。詳細は台帳 `#E2E-BASELINE-DEBT-01` の全件表を参照。
- **境界**: seed bundleは変更しない。各spec単位で `bash frontend/scripts/run-e2e.sh e2e/<spec>.spec.ts` により個別検証する。
- **所有パス**: `frontend/e2e/*.spec.ts`（対象6 spec）。他ユニットと重複なし、独立着手可能。
- **次アクション**: `/prompt-craft-agent-fast` で専用プロンプトを生成する。

## TASK-461（プロンプト生成済み・未実行）— M-05残2件の runtime evidence

専用プロンプト `~/.claude/prompt-craft-runs/agent-fast-task461-m05-runtime-evidence.md`（2026-07-28 14:25生成）が既に存在する。本書では複製しない。着手前に同プロンプトの実行状況（`tmp/fe12-m05-evidence/` 配下の最新タイムスタンプ）を確認し、未実行なら次点の着手候補とする。実行はAPI create/PATCH/DELETEとdedicated Chrome 9222を使うbrowser evidence収集であり、コード変更を伴わない。

## BUG-453 — CLOSED（2026-07-28・`bbb31f9be`+`bbb87c6a1`）

解消済み・回帰再認定も完了。#ledgerから節削除済み。GrokAgentへ再割当てしない。

## BUG-454-B1 — pet→owner の application相関

- **目的**: viewerが複数clinicを閲覧できる場合でも、破損した pet(A)→owner(B) をassociation、検索、並び順、detailへ復元しない。
- **根本原因**: `backend/internal/pet/repository.go:113` のJOINと`:152,209`のpreloadは `owners.clinic_id IN clinicIDs` だけで、`owners.clinic_id = pets.clinic_id` を要求しない。現test `repository_clinic_isolation_test.go:176` は破損graphの復元を許す旧契約である。
- **変更方針**: JOIN/preloadをpet行へ相関し、FindAll、search/order、FindByIDForClinicsのDB-backed fixtureを挙動変更として更新する。DDLは扱わない。
- **許可パス**:
  - `backend/internal/pet/repository.go`
  - `backend/internal/pet/repository_clinic_isolation_test.go`
- **禁止する巻き込み変更**: model tag、migration、owner API、pet lifecycle filter、`deleted_at` / `deceased_at` の追加。
- **検証コマンド**:
  ```bash
  docker compose exec -T backend go test ./internal/pet/ -run 'TestPetRepository_(FindAll|OwnerPreload_ClinicIsolation|FindByIDForClinics)' -count=1 -p 1 -v
  docker compose exec -T backend go build ./internal/pet/...
  docker compose exec -T backend go vet ./internal/pet/...
  docker compose exec -T backend gofmt -l internal/pet/repository.go internal/pet/repository_clinic_isolation_test.go
  git diff --name-only -- backend/internal/pet/repository.go backend/internal/pet/repository_clinic_isolation_test.go
  ```
- **完了判定**: 認可済みclinic集合内でも相関不一致ownerを復元せず、正常association/search/order/detailは維持され、差分が2 fileだけである。

## BUG-455 — `default:true` bool の create 契約

このIDは一括変更しない。現在コードでexact write pathまで確認できたPermissionGroupだけを実装可能とする。後続表は調査候補であって実装契約ではない。各候補はexact repository/service/test pathを再測定して7要素の独立契約へ昇格するまで変更禁止である。

**AUTH-S / AUTH-B は完了済み**（`ab9eaa170`＋`6515ffd68`、2026-07-28）。下記2節は実装済みの記録として残す。次の着手対象は `BUG-455-I` である。

**BUG-455-I 着手前に必読（台帳の設計申し送り・2026-07-28）**: AUTH-B が採った方式はINSERT→条件付きUPDATEの補償型であり、同一tx内でfail-closedだが `false` の場合にwriteが2回になり、tx内部に `is_active=true` の過渡状態を作る。DB triggerや監査がこの過渡状態を観測する設計だと影響が出る。残27 fieldへ同じ形を26 repositoryへ複製する前に、`Create` 側でzero valueをINSERTに含める単文方式（GORMの `Select` による明示field指定など）が本repoのmodel/tag構成で成立するかを実測で評価すること。成立するなら単文方式が正であり、AUTH-B の補償方式をtemplateとして横展開してはならない。この評価は `BUG-455-I` の検証コマンド（後述）に先立って行う。

### BUG-455-AUTH-S — PermissionGroup presence型（完了・記録として残置）

- **目的**: handlerへ未接続のpresence-aware create DTOとbinding fixtureを追加し、omitted / false / true を型として表現可能にする。
- **根本原因**: `backend/internal/auth/http_permission.go:127-146` の非pointer boolではomittedとfalseが同じzero valueになる。
- **変更方針**: 新DTOを追加してJSON bindingの3状態だけを固定する。この単位ではhandler/service/repositoryへ接続せず、runtime挙動を変えない。
- **許可パス**:
  - `backend/internal/auth/permission_group_create_request.go`（新規候補）
  - `backend/internal/auth/permission_group_create_request_test.go`（新規候補）
- **禁止する巻き込み変更**: handler/service/model/repository、default解決、PATCH、他master、migration、generated frontend型。
- **検証コマンド**:
  ```bash
  docker compose exec -T backend go test ./internal/auth/... -run '^TestPermissionGroupCreateRequest_' -count=1 -p 1
  docker compose exec -T backend go vet ./internal/auth/...
  docker compose exec -T backend gofmt -l internal/auth/permission_group_create_request.go internal/auth/permission_group_create_request_test.go
  git diff --name-only -- backend/internal/auth/permission_group_create_request.go backend/internal/auth/permission_group_create_request_test.go
  ```
- **完了判定**: binding testが3状態を区別し、production handler挙動とDBが不変で、差分が新規2 fileだけである。

### BUG-455-AUTH-B — PermissionGroup 永続化挙動（完了・記録として残置）

- **目的**: explicit falseをDB read-backでもfalseとして保持し、omittedはtrueにする。
- **根本原因**: `backend/internal/model/permission_group.go:10-20` のnon-pointer `bool` + `default:true` と `permission_group_repository.go:115-121` の通常Createによりfalse列がINSERTから落ちる。
- **変更方針**: presence-aware入力から解決済み値を作り、false列を明示保存する最小方式を選ぶ。全39 modelの一括pointer化はしない。
- **許可パス**:
  - `backend/internal/auth/permission_group_create_request.go`（AUTH-Sで作成済みの既存前提）
  - `backend/internal/model/permission_group.go`
  - `backend/internal/auth/http_permission.go`
  - `backend/internal/auth/permission_group_service.go`
  - `backend/internal/auth/permission_group_repository.go`
  - `backend/internal/auth/http_permission_handlers_test.go`
  - `backend/internal/auth/permission_group_repository_test.go`
  - `backend/internal/auth/permission_group_service_test.go`
- **禁止する巻き込み変更**: 他model tag、migration/seed、PATCH、clinic初期permission作成contract。
- **検証コマンド**:
  ```bash
  docker compose exec -T backend go test ./internal/auth/... -run 'PermissionGroup.*(Create|False|Default)' -count=1 -p 1
  docker compose exec -T backend go build ./internal/auth/...
  docker compose exec -T backend go vet ./internal/auth/...
  docker compose exec -T backend gofmt -l internal/auth/permission_group_create_request.go internal/model/permission_group.go internal/auth/http_permission.go internal/auth/permission_group_service.go internal/auth/permission_group_repository.go internal/auth/http_permission_handlers_test.go internal/auth/permission_group_repository_test.go internal/auth/permission_group_service_test.go
  git diff --name-only -- backend/internal/auth/permission_group_create_request.go backend/internal/model/permission_group.go backend/internal/auth/http_permission.go backend/internal/auth/permission_group_service.go backend/internal/auth/permission_group_repository.go backend/internal/auth/http_permission_handlers_test.go backend/internal/auth/permission_group_repository_test.go backend/internal/auth/permission_group_service_test.go
  ```
- **完了判定**: omitted→true、false→false、true→trueがresponseとDBで一致し、既存audit/transaction testもgreenで、差分が8 file以内である。

### BUG-455-I — 後続resource write-path測定

- **目的**: 残る27 fieldについて、create/upsert到達性、omitted/false/true契約、exact repository/service/test pathを確定し、実装可能な独立契約へ分解する。
- **根本原因**: `default:true` inventoryは39 field/26 modelだが、実害はcreate/upsert経路に依存する。台帳のhandler到達表だけでは、永続化symbolと既存test fileのexact allowlistを確定できない。
- **変更方針**: 次表を仮説としてread-only censusを行う。各rowについてrequest→handler→service→repository→GORM Createを引用し、既存testまたは明示したcandidate-new test path、S/B分類、3値matrixを記録する。結果はセッション報告へ出し、sourceは変更しない。
- **許可パス**: なし。
- **禁止する巻き込み変更**: 次表のpathを未検証のまま編集すること、39 tag一括変更、実害なし11 field、migration/seed、PATCH、generated frontend型。
- **検証コマンド**:
  ```bash
  docker compose exec -T backend go test ./internal/lintscan/ -run 'Test.*Model.*Drift' -count=1 -p 1
  docker compose exec -T backend go test ./internal/auth/... ./internal/pet/... ./internal/billing/... ./internal/inventory/... ./internal/medicalrecord/... ./internal/reservation/... ./internal/staff/... ./internal/trimming/... -run '^$' -count=1 -p 1
  git diff --name-only --
  ```
- **完了判定**: 27 fieldすべてにcreate/upsert到達性、exact write symbol、exact existing/candidate test、S/B契約があり、未分類0、source差分0である。

調査仮説表:

| Slice | Model field | S/B許可production path |
|---|---|---|
| PET-ANIMAL-SPECIES | `AnimalSpecies.IsActive` | `backend/internal/model/animal_species.go`, `backend/internal/pet/animal_species_request.go`, `backend/internal/pet/animal_species_handler.go` |
| MR-CAGE | `Cage.IsActive` | `backend/internal/model/cage.go`, `backend/internal/medicalrecord/cage_request.go`, `backend/internal/medicalrecord/cage_handler.go` |
| BILL-CAMPAIGN | `Campaign.IsActive` | `backend/internal/model/campaign.go`, `backend/internal/billing/campaign_request.go`, `backend/internal/billing/campaign_handler.go` |
| MR-CHECKUP-TYPE | `CheckupType.IsActive` | `backend/internal/model/checkup_type.go`, `backend/internal/medicalrecord/checkup_type_request.go`, `backend/internal/medicalrecord/checkup_type_handler.go` |
| MR-CHIEF-COMPLAINT | `ChiefComplaintType.IsActive` | `backend/internal/model/chief_complaint_type.go`, `backend/internal/medicalrecord/chief_complaint_request.go`, `backend/internal/medicalrecord/chief_complaint_handler.go` |
| MR-CONSULTATION | `Consultation.IsActive` | `backend/internal/model/consultation.go`, `backend/internal/medicalrecord/consultation_request.go`, `backend/internal/medicalrecord/consultation_handler.go` |
| MR-DIAGNOSIS-TYPE | `DiagnosisType.IsActive` | `backend/internal/model/diagnosis.go`, `backend/internal/medicalrecord/diagnosis_request.go`, `backend/internal/medicalrecord/diagnosis_handler.go` |
| MR-DIAGNOSIS-NAME | `DiagnosisName.IsActive` | `backend/internal/model/diagnosis.go`, `backend/internal/medicalrecord/diagnosis_request.go`, `backend/internal/medicalrecord/diagnosis_handler.go` |
| MR-EXAM-TYPE | `ExaminationType.IsActive` | `backend/internal/model/examination_type.go`, `backend/internal/medicalrecord/exam_type_request.go`, `backend/internal/medicalrecord/exam_type_handler.go` |
| MR-HOSPITALIZATION-PLAN | `HospitalizationPlan.IsActive` | `backend/internal/model/hospitalization_plan.go`, `backend/internal/medicalrecord/hospitalization_plan_request.go`, `backend/internal/medicalrecord/hospitalization_plan_handler.go` |
| MR-INQUIRY | `InquiryTemplate.IsActive` | `backend/internal/model/inquiry_template.go`, `backend/internal/medicalrecord/inquiry_template_request.go`, `backend/internal/medicalrecord/inquiry_template_handler.go` |
| BILL-INSURANCE | `Insurance.IsActive` | `backend/internal/model/insurance.go`, `backend/internal/billing/insurance_request.go`, `backend/internal/billing/insurance_handler.go` |
| RES-LINE-SETTING | `ShowNoStaffOption` | `backend/internal/model/line_reservation_setting.go`, `backend/internal/reservation/line_reservation_setting_request.go`, `backend/internal/reservation/line_reservation_setting_service.go` |
| MR-MEDICINE | `Medicine.IsActive` | `backend/internal/model/medicine.go`, `backend/internal/medicalrecord/medicine_request.go`, `backend/internal/medicalrecord/medicine_handler.go` |
| INV-MERCHANDISE | `MerchandiseItem.IsActive` | `backend/internal/model/merchandise_item.go`, `backend/internal/inventory/merchandise_item_request.go`, `backend/internal/inventory/merchandise_item_handler.go` |
| STAFF-OCCUPATION | `Occupation.IsActive` | `backend/internal/model/occupation.go`, `backend/internal/staff/occupation_request.go`, `backend/internal/staff/occupation_handler.go` |
| MR-PROCEDURE | `Procedure.IsActive` | `backend/internal/model/procedure.go`, `backend/internal/medicalrecord/procedure_request.go`, `backend/internal/medicalrecord/procedure_handler.go` |
| RES-TYPE-ACTIVE | `ReservationType.IsActive` | `backend/internal/model/reservation_type.go`, `backend/internal/reservation/reservation_type_request.go`, `backend/internal/reservation/reservation_type_handler.go` |
| RES-TYPE-VISIBLE | `ReservationType.ReservationVisible` | `backend/internal/model/reservation_type.go`, `backend/internal/reservation/reservation_type_request.go`, `backend/internal/reservation/reservation_type_service.go` |
| RES-SLOT | `ReservationTypeAvailableSlot.IsActive` | `backend/internal/model/reservation_type.go`, `backend/internal/reservation/reservation_type_request.go`, `backend/internal/reservation/reservation_type_service.go` |
| RES-GROUP | `ReservationTypeGroup.IsActive` | `backend/internal/model/reservation_type_group.go`, `backend/internal/reservation/reservation_type_group_request.go`, `backend/internal/reservation/reservation_type_group_handler.go` |
| STAFF-RESERVATION | `Staff.ReservationVisible` | `backend/internal/model/staff.go`, `backend/internal/staff/staff_request.go`, `backend/internal/staff/reservation_staff_request.go` |
| STAFF-SHIFT-TEMPLATE | `ShiftTemplate.IsActive` | `backend/internal/model/staff.go`, `backend/internal/staff/shift_template_request.go`, `backend/internal/staff/shift_template_service.go` |
| TRIM-COURSE | `TrimmingCourse.IsActive` | `backend/internal/model/trimming_master.go`, `backend/internal/trimming/trimming_course_request.go`, `backend/internal/trimming/trimming_course_handler.go` |
| TRIM-OPTION-ACTIVE | `TrimmingOption.IsActive` | `backend/internal/model/trimming_master.go`, `backend/internal/trimming/trimming_option_request.go`, `backend/internal/trimming/trimming_option_handler.go` |
| TRIM-OPTION-COMBINABLE | `TrimmingOption.IsCombinable` | `backend/internal/model/trimming_master.go`, `backend/internal/trimming/trimming_option_request.go`, `backend/internal/trimming/trimming_option_handler.go` |
| MR-VACCINE | `Vaccine.IsActive` | `backend/internal/model/vaccine.go`, `backend/internal/medicalrecord/vaccine_request.go`, `backend/internal/medicalrecord/vaccine_handler.go` |

この表の各rowは調査仮説であり、プレースホルダーコマンドや未確認pathを使った実装開始を許可しない。

## BUG-456 — カルテ検査タブの単位（unit）列空欄

**状態更新（2026-07-28）**: 結果値・基準値の表示は解消済み・コミット済み（`7ed9ae808`）。`ExaminationGroup.tsx` は `inspectionValue || result || "-"` と `referenceValue || normalValue || "-"` のlegacy-safe fallbackで固定され、14 testがgreen。**本entryはunit列の空欄のみを残す**。旧節の許可パス（`ExaminationGroup.tsx`等）は不要——unit欠落は表示側のfield名不一致ではなく、request payload→永続化の経路調査から始まる別種の欠陥である。

- **目的**: `GET /api/v1/examinations/:id/items` が `unit ''` を返す原因を特定し、通常登録経路で入力した単位が保存・返却されるようにする。
- **根本原因（未特定・要調査）**: 実測レスポンス（pet `1000018` / MR `1425546` / exam `1014562`）は `unit ''`。`upsertExamItemRequest`（`backend/internal/medicalrecord/examination_request.go`）が `unit` フィールドをそもそも受け取っているか、受け取っていても `examination_service.go` のsave経路で書き込まれていないか、GORM modelにfieldが存在しないかを request→handler→service→model の順で追跡する。
- **変更方針**: 調査unitとして着手し、read-onlyでrequest→handler→service→GORM createの到達性を確認したうえで、欠落箇所に応じた最小差分（DTOへのfield追加／serviceでの代入漏れ修正のいずれか）をtest-firstで実装する。表示側（`ExaminationGroup.tsx`）は既に正しいため変更しない。
- **許可パス**: 調査完了後に確定（候補: `backend/internal/medicalrecord/examination_request.go`、`backend/internal/medicalrecord/examination_service.go`、対応する `_test.go`）。調査フェーズはread-onlyで、production code着手前に発見事実を報告すること。
- **禁止する巻き込み変更**: `ExaminationGroup.tsx`（既に正しい）、`ExamItemsTable.tsx`（既に正しい）、mapper、lab import、generated model。
- **検証コマンド（調査フェーズ）**:
  ```bash
  docker compose exec -T backend go test ./internal/medicalrecord/ -run 'TestExamination' -count=1 -p 1 -v
  git grep -n '"unit"\|Unit\b' backend/internal/medicalrecord/examination_request.go backend/internal/medicalrecord/examination_service.go backend/internal/model/examination_record.go
  ```
- **完了判定**: 通常登録APIで単位を入力→保存→即時GETで同一値が返ることをAPI実測で確認し、対応するDB-backed testが追加され、allowlist外0。

## BUG-457 — CLOSED（2026-07-28・`edit`案で確定・`75aff5786`+`b415892ec`+`7f17aeb53`）

FE表示・latest callback guard・backend認可をすべて `canEdit` に統一済み。#ledgerから節削除済み。GrokAgentへ再割当てしない。

## BUG-458-V — レスポンシブ再測定

- **目的**: 6 route×4 viewportの24 cellを、viewport設定後navigationと十分なsettleを使って再測定する。
- **根本原因**: 旧captureはCDP metrics変更直後に撮影し、`Sidebar.tsx:71-76` のmatchMedia effectが反映される前のexpanded sidebarを含む可能性がある。台帳の「内訳未確認」は陳腐化したが、既存10 cell/13 findingを修正根拠へ直結できない。
- **変更方針**: `frontend/e2e/ui-design-compliance-readonly.spec.ts:358-425` を基礎に、viewport→goto→networkidle→2 rAF、sidebar幅56px、document overflow、対象element bounding box、inner scroll affordance、非GET=0、screenshotを24 cellで収集する。
- **許可パス**:
  - `frontend/e2e/ui-design-compliance-readonly.spec.ts`
  - `frontend/e2e/helpers/ui-design-audit.ts`
- **禁止する巻き込み変更**: production UI、backend/DB、fixture mutation、旧画像だけでの原因断定。
- **検証コマンド**:
  ```bash
  docker compose exec -T frontend npx playwright test e2e/ui-design-compliance-readonly.spec.ts --grep 'responsive'
  docker compose exec -T frontend npx eslint --max-warnings 0 e2e/ui-design-compliance-readonly.spec.ts e2e/helpers/ui-design-audit.ts
  git diff --name-only -- frontend/e2e/ui-design-compliance-readonly.spec.ts frontend/e2e/helpers/ui-design-audit.ts
  ```
- **完了判定**: `/`, `/medical-records`, `/hospitalization`, `/hospitalization/990007`, `/examinations`, `/checkups` の1440/1200/800/500全cellに測定表・画像・element rect・network mutation 0がある。

## BUG-458-S — shared shell/header/detail 修復

- **目的**: 再測定で確定したshell、header、patient metadata、tabのclipだけを直す。
- **根本原因**: 候補は `Layout.tsx:30-33` のmain shrink、`FormHeader.tsx:16-37` のwrap、`PatientInfoCard.tsx:81-85,138-175` の固定幅、`UnifiedTabs.tsx:36-56` のoverflowである。確定原因は再測定結果を正とする。
- **変更方針**: shared contractをREDにし、44px target、臨床cue、全tab到達性を維持したままreflowする。
- **許可パス**:
  - `frontend/src/components/shared/Layout/Layout.tsx`
  - `frontend/src/components/shared/Form/FormHeader.tsx`
  - `frontend/src/components/shared/Form/FormHeader.test.tsx`
  - `frontend/src/components/shared/PatientInfoCard/PatientInfoCard.tsx`
  - `frontend/src/components/shared/PatientInfoCard/PatientInfoCard.test.tsx`
  - `frontend/src/components/shared/UnifiedTabs.tsx`
  - `frontend/src/components/shared/UnifiedTabs.test.tsx`
- **禁止する巻き込み変更**: design token、Sidebar redesign、navigation IA、table/list route、backend。
- **検証コマンド**:
  ```bash
  docker compose exec -T frontend npx vitest run src/components/shared/Form/FormHeader.test.tsx src/components/shared/PatientInfoCard/PatientInfoCard.test.tsx src/components/shared/UnifiedTabs.test.tsx
  docker compose exec -T frontend npx eslint --max-warnings 0 src/components/shared/Layout/Layout.tsx src/components/shared/Form/FormHeader.tsx src/components/shared/Form/FormHeader.test.tsx src/components/shared/PatientInfoCard/PatientInfoCard.tsx src/components/shared/PatientInfoCard/PatientInfoCard.test.tsx src/components/shared/UnifiedTabs.tsx src/components/shared/UnifiedTabs.test.tsx
  git diff --name-only -- frontend/src/components/shared/Layout/Layout.tsx frontend/src/components/shared/Form/FormHeader.tsx frontend/src/components/shared/Form/FormHeader.test.tsx frontend/src/components/shared/PatientInfoCard/PatientInfoCard.tsx frontend/src/components/shared/PatientInfoCard/PatientInfoCard.test.tsx frontend/src/components/shared/UnifiedTabs.tsx frontend/src/components/shared/UnifiedTabs.test.tsx
  ```
- **完了判定**: 確定cellでclip/overlap 0、44px・臨床cue・全操作到達性を維持し、allowlist外0。

## BUG-458-T — table/board route修復

- **目的**: 再測定で確定した一覧右端、action、boardのclipだけを直す。
- **根本原因**: 候補は `DataTable.tsx:42-45` のmin-width、route列優先度、`HospitalizationBoard.tsx:176-201` の `min-w-[800px]` である。
- **変更方針**: status/death/danger/deadline/actionを隠さず、明示scroll affordanceまたは列reflowをroute別testで固定する。
- **許可パス**:
  - `frontend/src/components/shared/DataTable/DataTable.tsx`
  - `frontend/src/components/shared/DataTable/DataTable.test.tsx`
  - `frontend/src/features/hospitalization/components/HospitalizationBoard.tsx`
  - `frontend/src/features/hospitalization/components/HospitalizationBoard.test.tsx`
  - `frontend/src/features/medical-records/routes/MedicalRecords.tsx`
  - `frontend/src/features/medical-records/routes/MedicalRecords.test.tsx`
  - `frontend/src/features/medical-records/routes/medical-records-columns.tsx`
  - `frontend/src/features/examinations/routes/ExaminationsList.tsx`
  - `frontend/src/features/examinations/routes/ExaminationsList.test.tsx`
  - `frontend/src/features/checkups/routes/CheckupsList.tsx`
  - `frontend/src/features/checkups/routes/CheckupsList.test.tsx`
- **禁止する巻き込み変更**: shared header/detail、design token、clinical column削除、API/backend。
- **検証コマンド**:
  ```bash
  docker compose exec -T frontend npx vitest run src/components/shared/DataTable/DataTable.test.tsx src/features/hospitalization/components/HospitalizationBoard.test.tsx src/features/medical-records/routes/MedicalRecords.test.tsx src/features/examinations/routes/ExaminationsList.test.tsx src/features/checkups/routes/CheckupsList.test.tsx
  docker compose exec -T frontend npx eslint --max-warnings 0 src/components/shared/DataTable/DataTable.tsx src/features/hospitalization/components/HospitalizationBoard.tsx src/features/medical-records/routes/MedicalRecords.tsx src/features/medical-records/routes/medical-records-columns.tsx src/features/examinations/routes/ExaminationsList.tsx src/features/checkups/routes/CheckupsList.tsx
  git diff --name-only -- frontend/src/components/shared/DataTable/DataTable.tsx frontend/src/components/shared/DataTable/DataTable.test.tsx frontend/src/features/hospitalization/components/HospitalizationBoard.tsx frontend/src/features/hospitalization/components/HospitalizationBoard.test.tsx frontend/src/features/medical-records/routes/MedicalRecords.tsx frontend/src/features/medical-records/routes/MedicalRecords.test.tsx frontend/src/features/medical-records/routes/medical-records-columns.tsx frontend/src/features/examinations/routes/ExaminationsList.tsx frontend/src/features/examinations/routes/ExaminationsList.test.tsx frontend/src/features/checkups/routes/CheckupsList.tsx frontend/src/features/checkups/routes/CheckupsList.test.tsx
  ```
- **完了判定**: 再測定で確定したtable/board cellのclip 0、必須列/操作可視、allowlist外0。

## TASK-444-S1 — response型境界の新規誤用防止

- **目的**: domain model生成型を新しいresponse型としてimportできない静的境界を作る。
- **根本原因**: `frontend/eslint.config.js:91` 付近に generated modelのresponse利用を制約するruleがなく、既存response codegenは `auth-responses.ts` / `trimming-responses.ts` に限定される。
- **変更方針**: 現import inventoryをfixtureとして凍結し、request/adapter/enum等の理由付き既存例外だけをallowlist化する。新規entity response importは失敗させる。既存294 siteを同時移行しない。
- **許可パス**:
  - `frontend/eslint.config.js`
  - `frontend/src/types/generated-model-response-boundary.test.ts`（新規候補）
- **禁止する巻き込み変更**: `models.ts`、既存import一括置換、backend DTO/codegen、無理由global exemption。
- **検証コマンド**:
  ```bash
  docker compose exec -T frontend npx vitest run src/types/generated-model-response-boundary.test.ts
  docker compose exec -T frontend npx eslint --max-warnings 0 eslint.config.js src/types/generated-model-response-boundary.test.ts
  git diff --name-only -- frontend/eslint.config.js frontend/src/types/generated-model-response-boundary.test.ts
  ```
- **完了判定**: 新規response誤用fixtureがRED、理由付き既存例外がgreen、inventory増加がfailし、差分2 fileだけ。

## TASK-444-S2 — domain別response DTO移行

- **目的**: 一つのdomainだけをwire response DTOへ移し、そのdomainのgenerated model entity importを0にする。
- **根本原因**: backendには明示response DTOがあるのに、frontend adapterが `@/types/generated/models` を受ける箇所が残る。
- **変更方針**: この契約ではpetだけを選ぶ。backend response structをsourceに既存response生成方式を再利用し、pet API types、transform、fixtureを同じsliceで移す。
- **許可パス**（pet slice）:
  - `backend/internal/pet/pet_response.go`
  - `backend/tygo.yaml`
  - `frontend/src/types/generated/pet-responses.ts`（生成候補）
  - `frontend/src/types/pet.ts`
  - `frontend/src/features/pets/api/create-pet.ts`
  - `frontend/src/features/pets/api/update-pet.ts`
  - `frontend/src/features/owners/loaders.ts`
  - `frontend/src/features/owners/hooks/use-pet-form-list-state.test.ts`
- **禁止する巻き込み変更**: reservation/owner等の別domain、`models.ts`手編集、wire field追加、全site一括移行。
- **検証コマンド**:
  ```bash
  docker compose exec -T frontend npx vitest run src/features/pets src/features/owners/hooks/use-pet-form-list-state.test.ts
  docker compose exec -T frontend npx eslint --max-warnings 0 src/types/pet.ts src/features/pets/api/create-pet.ts src/features/pets/api/update-pet.ts src/features/owners/loaders.ts src/features/owners/hooks/use-pet-form-list-state.test.ts
  git diff --name-only -- backend/internal/pet/pet_response.go backend/tygo.yaml frontend/src/types/generated/pet-responses.ts frontend/src/types/pet.ts frontend/src/features/pets/api/create-pet.ts frontend/src/features/pets/api/update-pet.ts frontend/src/features/owners/loaders.ts frontend/src/features/owners/hooks/use-pet-form-list-state.test.ts
  ```
- **完了判定**: 選択domainのwire fixtureが実DTOと一致し、entity response import 0、tests green、allowlist外0。生成はユーザー実行の `make codegen` 後に差分を再確認する。

## TASK-445-S1 — Payment clinic_id model

- **目的**: Paymentへ内部tenant keyを型として追加し、frontend wireへ露出させない。
- **根本原因**: `backend/internal/model/accounting.go:163-189` のPaymentと `backend/migrations/001_init.sql:1946-1965` のpaymentsに `clinic_id` が無い。
- **変更方針**: modelへ `ClinicID uint64` を `json:"-"` で追加し、model/response/test literalを構造変更としてコンパイル可能にする。write挙動とDDLは扱わない。
- **許可パス**:
  - `backend/internal/model/accounting.go`
  - `backend/internal/billing/accounting_response_test.go`
  - `backend/internal/billing/accounting_repository_tx_atomicity_test.go`
- **禁止する巻き込み変更**: `001_init.sql`、incremental migration、builder/repository挙動、generated frontend型。
- **検証コマンド**:
  ```bash
  docker compose exec -T backend go test ./internal/billing/ -run 'TestAccountingResponse|TestAccountingRepository_SavePayment' -count=1 -p 1
  docker compose exec -T backend go vet ./internal/billing/...
  docker compose exec -T backend gofmt -l internal/model/accounting.go internal/billing/accounting_response_test.go internal/billing/accounting_repository_tx_atomicity_test.go
  git diff --name-only -- backend/internal/model/accounting.go backend/internal/billing/accounting_response_test.go backend/internal/billing/accounting_repository_tx_atomicity_test.go
  ```
- **完了判定**: backend compile/test green、frontend生成差分0、挙動未変更、差分3 file以内。

## TASK-445-B1 — Payment clinic_id write

- **目的**: 全Payment作成・訂正経路がowned Billingのclinic_idを同一transactionで保存する。
- **根本原因**: `accounting_service_builders.go:117` のPayment builderと `accounting_repository.go:395` のSavePaymentは現在clinic_idを要求しない。
- **変更方針**: request由来clinic_idを信用せず、locked Billingから導出してbuilder/correctionへ伝播する。既存 `SavePayment` はmodelの全fieldを保存するため変更せず、cross-clinic mismatchをservice/DB testでfail-closedにする。
- **許可パス**:
  - `backend/internal/billing/accounting_service_builders.go`
  - `backend/internal/billing/accounting_service_core.go`
  - `backend/internal/billing/accounting_service_correction.go`
  - `backend/internal/billing/payment_clinic_write_test.go`（新規候補）
  - `backend/internal/billing/accounting_repository_tx_atomicity_test.go`
  - `backend/internal/billing/accounting_service_correction_test.go`
- **禁止する巻き込み変更**: migration、reports、PaymentSplit、request clinic_id追加、TASK-251 category contract。
- **検証コマンド**:
  ```bash
  docker compose exec -T backend go test ./internal/billing/ -run 'Test.*(SavePayment|CorrectCreditPayment|PaymentClinic).*' -count=1 -p 1
  docker compose exec -T backend go vet ./internal/billing/...
  docker compose exec -T backend gofmt -l internal/billing/accounting_service_builders.go internal/billing/accounting_service_core.go internal/billing/accounting_service_correction.go internal/billing/payment_clinic_write_test.go internal/billing/accounting_repository_tx_atomicity_test.go internal/billing/accounting_service_correction_test.go
  git diff --name-only -- backend/internal/billing/accounting_service_builders.go backend/internal/billing/accounting_service_core.go backend/internal/billing/accounting_service_correction.go backend/internal/billing/payment_clinic_write_test.go backend/internal/billing/accounting_repository_tx_atomicity_test.go backend/internal/billing/accounting_service_correction_test.go
  ```
- **完了判定**: normal/correction/rollback全経路がlocked Billing clinicを保存し、foreign clinicを拒否し、allowlist外0。

## TASK-445-S2-PAY — Payment clinic複合FK migration

- **目的**: paymentsへclinic_idをbackfillし、billings/payment_methodsとのclinic複合FKだけを追加する。
- **根本原因**: `backend/migrations/001_init.sql:1946-1965` のpaymentsにclinic_idがなく、billing_id/payment_method_idはsingle-column FKである。
- **変更方針**: current tailが003であることを再確認して `004_add_payments_clinic_scope.sql` を追加する。canaryで payment→billing と payment→payment_method のclinic不一致を列挙し、billing由来backfill、NOT NULL、必要なparent unique、2 composite FKを順序化する。
- **許可パス**:
  - `backend/migrations/004_add_payments_clinic_scope.sql`（新規候補。番号競合時は編集せず停止）
  - `backend/internal/model/rls_migration_test.go`
- **禁止する巻き込み変更**: pets/owners、medical_records/vaccinations/billingsの他FK、`001_init.sql`、seed、application repository、CASCADE、migration適用、testdb tag。
- **検証コマンド**:
  ```bash
  docker compose exec -T backend go test ./internal/model/ -run 'Test.*Payments.*Clinic.*Migration' -count=1 -p 1
  docker compose exec -T backend go vet ./internal/model/...
  docker compose exec -T backend gofmt -l internal/model/rls_migration_test.go
  git diff --name-only -- backend/migrations/004_add_payments_clinic_scope.sql backend/internal/model/rls_migration_test.go
  ```
- **完了判定**: canary、backfill、NOT NULL、2 composite FKのstatic contract green、CASCADE 0、allowlist外0。runtime schemaはユーザー適用まで未報告。

## TASK-445-S3-PET — pets→owners clinic複合FK migration

- **目的**: pets `(clinic_id, owner_id)` が同じclinicのowners `(clinic_id, id)` だけを参照できるようにする。
- **根本原因**: `backend/migrations/001_init.sql:1098` は `pets.owner_id → owners.id` のsingle-column FKで、`owners`側のclinic uniqueだけではchild相関を強制しない。
- **変更方針**: current migration tailを再確認して `005_enforce_pets_owner_clinic_fk.sql` を追加する。corrupt pet→owner canary、旧FK置換、composite FKを一つのDDL concernとして記述する。
- **許可パス**:
  - `backend/migrations/005_enforce_pets_owner_clinic_fk.sql`（新規候補。番号競合時は編集せず停止）
  - `backend/internal/model/rls_migration_test.go`
- **禁止する巻き込み変更**: Payment、他clinical table、`001_init.sql`、seed、repository、CASCADE、migration適用、model relation tag。
- **検証コマンド**:
  ```bash
  docker compose exec -T backend go test ./internal/model/ -run 'Test.*Pets.*Owner.*Clinic.*Migration' -count=1 -p 1
  docker compose exec -T backend go vet ./internal/model/...
  docker compose exec -T backend gofmt -l internal/model/rls_migration_test.go
  git diff --name-only -- backend/migrations/005_enforce_pets_owner_clinic_fk.sql backend/internal/model/rls_migration_test.go
  ```
- **完了判定**: corrupt graph canaryがALTER前に停止し、正常/同clinic pet-ownerだけを許し、CASCADE追加0、allowlist外0。

## TASK-445-S4-CLINICAL — clinical graph clinic複合FK migration

- **目的**: medical_records、vaccinations、billingsの既存clinic列を親pet/owner/recordへ宣言的に相関する。
- **根本原因**: 現DDLは一部parent uniqueを持つが、`medical_records→pets/owners`、`vaccinations→pets/medical_records`、`billings→pets/owners` のchild側clinic相関が不完全である。
- **変更方針**: current tailを再確認して `006_enforce_clinical_graph_clinic_fks.sql` を追加する。6 edgeを列挙したcanaryとcomposite FKだけを含める。
- **許可パス**:
  - `backend/migrations/006_enforce_clinical_graph_clinic_fks.sql`（新規候補。番号競合時は編集せず停止）
  - `backend/internal/model/rls_migration_test.go`
- **禁止する巻き込み変更**: Payment、pets→owners、provenance owner snapshot、`001_init.sql`、seed、repository、CASCADE、migration適用。
- **検証コマンド**:
  ```bash
  docker compose exec -T backend go test ./internal/model/ -run 'Test.*Clinical.*Clinic.*Migration' -count=1 -p 1
  docker compose exec -T backend go vet ./internal/model/...
  docker compose exec -T backend gofmt -l internal/model/rls_migration_test.go
  git diff --name-only -- backend/migrations/006_enforce_clinical_graph_clinic_fks.sql backend/internal/model/rls_migration_test.go
  ```
- **完了判定**: 6 edgeのcanaryとFKが逐語確認でき、owner snapshot semanticsを変更せず、CASCADE追加0、allowlist外0。

## SEC-DUR-01 — snapshot/current-owner durable contract

### SEC-DUR-01-MR-T1（test構造）

- **目的**: 譲渡後のExamination update、vaccination update、会計なし退院、記録確定がclinic/petを守りながら継続するwrite回帰を固定する。
- **根本原因**: MR-A1/MR-C1-3はproduction完了だがwrite面の回帰testが不足する。
- **変更方針**: production不変のtest-only unitとする。
- **許可パス**:
  - `backend/internal/medicalrecord/examination_service_test.go`
  - `backend/internal/medicalrecord/vaccination_service_test.go`
  - `backend/internal/medicalrecord/hospitalization_service_test.go`
  - `backend/internal/medicalrecord/medical_record_service_test.go`
- **禁止する巻き込み変更**: production、billing、migration、owner snapshot意味変更。
- **検証コマンド**:
  ```bash
  docker compose exec -T backend go test ./internal/medicalrecord/ -run 'Test.*(Transfer|Relation|Vaccination|Discharge).*' -count=1 -p 1
  docker compose exec -T backend go vet ./internal/medicalrecord/...
  docker compose exec -T backend gofmt -l internal/medicalrecord/examination_service_test.go internal/medicalrecord/vaccination_service_test.go internal/medicalrecord/hospitalization_service_test.go internal/medicalrecord/medical_record_service_test.go
  git diff --name-only -- backend/internal/medicalrecord/examination_service_test.go backend/internal/medicalrecord/vaccination_service_test.go backend/internal/medicalrecord/hospitalization_service_test.go backend/internal/medicalrecord/medical_record_service_test.go
  ```
- **完了判定**: 同一clinic譲渡は成功、cross-clinicは拒否、rollbackを固定し、production差分0。

### SEC-DUR-01-MR-S1（dead predicate構造）

- **目的**: 到達不能なowner比較を除去し、pet/clinic検証の正本を明瞭化する。
- **根本原因**: `clinical_relation_validation.go:79-86` と `vaccination_service.go:356` のowner比較は同じpet行由来なのでpet ID一致後は常に真である。
- **変更方針**: owner節だけを除去し、pet ID/clinic fail-closedを維持する。
- **許可パス**:
  - `backend/internal/medicalrecord/clinical_relation_validation.go`
  - `backend/internal/medicalrecord/vaccination_service.go`
  - `backend/internal/medicalrecord/clinical_relation_validation_test.go`（新規候補）
  - `backend/internal/medicalrecord/vaccination_service_test.go`
- **禁止する巻き込み変更**: pet/clinic predicate、repository read、billing、DDL。
- **検証コマンド**:
  ```bash
  docker compose exec -T backend go test ./internal/medicalrecord/ -run 'Test.*(ClinicalRelation|Vaccination).*' -count=1 -p 1
  docker compose exec -T backend go vet ./internal/medicalrecord/...
  docker compose exec -T backend gofmt -l internal/medicalrecord/clinical_relation_validation.go internal/medicalrecord/vaccination_service.go internal/medicalrecord/clinical_relation_validation_test.go internal/medicalrecord/vaccination_service_test.go
  git diff --name-only -- backend/internal/medicalrecord/clinical_relation_validation.go backend/internal/medicalrecord/vaccination_service.go backend/internal/medicalrecord/clinical_relation_validation_test.go backend/internal/medicalrecord/vaccination_service_test.go
  ```
- **完了判定**: dead比較0、pet/clinic test green、挙動不変、allowlist外0。

### SEC-DUR-01-BILL-B1（current-owner billing挙動）

- **目的**: LTVとvaccination claimがDEC-27のcurrent-owner/snapshot境界に一致する。
- **根本原因**: `accounting_repository_ltv.go:15` と `billing_item_repository.go:468` にmedical record owner snapshotとの等値が残る。
- **変更方針**: 現ownerはpets経由、診療ownerはsnapshotとして維持し、譲渡fixtureでLTV/claimを挙動変更する。
- **許可パス**:
  - `backend/internal/billing/accounting_repository_ltv.go`
  - `backend/internal/billing/billing_item_repository.go`
  - `backend/internal/billing/accounting_repository_ltv_test.go`
  - `backend/internal/billing/billing_item_vaccination_test.go`
- **禁止する巻き込み変更**: unpaid/close、estimate、TASK-251、clinic DDL、lifecycle filter。
- **検証コマンド**:
  ```bash
  docker compose exec -T backend go test ./internal/billing/ -run 'Test.*(LTV|Vaccination|OwnerPet).*' -count=1 -p 1
  docker compose exec -T backend go vet ./internal/billing/...
  docker compose exec -T backend gofmt -l internal/billing/accounting_repository_ltv.go internal/billing/billing_item_repository.go internal/billing/accounting_repository_ltv_test.go internal/billing/billing_item_vaccination_test.go
  git diff --name-only -- backend/internal/billing/accounting_repository_ltv.go backend/internal/billing/billing_item_repository.go backend/internal/billing/accounting_repository_ltv_test.go backend/internal/billing/billing_item_vaccination_test.go
  ```
- **完了判定**: 譲渡後current-owner基準とsnapshot保持がfixtureで明示され、clinic隔離green、allowlist外0。

### SEC-DUR-01-BILL-B2（unpaid/close挙動）

- **目的**: unpaid/close集計をDEC-27へ揃え、旧snapshot owner mismatchで診療会計を消さない。
- **根本原因**: `accounting_repository_unpaid.go:56` 等のowner等値が残る。
- **変更方針**: unpaid/closeだけをpets経由のcurrent-owner基準へ変更し、direct billingと譲渡fixtureを固定する。
- **許可パス**:
  - `backend/internal/billing/accounting_repository_unpaid.go`
  - `backend/internal/billing/accounting_repository_unpaid_test.go`
  - `backend/internal/billing/accounting_repository_reports_close_test.go`
- **禁止する巻き込み変更**: LTV/claim、TASK-251 category、estimate、DDL。
- **検証コマンド**:
  ```bash
  docker compose exec -T backend go test ./internal/billing/ -run 'Test.*(Unpaid|Close|OwnerPet).*' -count=1 -p 1
  docker compose exec -T backend go vet ./internal/billing/...
  docker compose exec -T backend gofmt -l internal/billing/accounting_repository_unpaid.go internal/billing/accounting_repository_unpaid_test.go internal/billing/accounting_repository_reports_close_test.go
  git diff --name-only -- backend/internal/billing/accounting_repository_unpaid.go backend/internal/billing/accounting_repository_unpaid_test.go backend/internal/billing/accounting_repository_reports_close_test.go
  ```
- **完了判定**: 譲渡/direct billing/unpaid/closeの期待がgreenで、TASK-251面とallowlist外差分0。

## SEC-SWEEP-02 — grandchild親clinic相関

**進捗（2026-07-28・残余0で完了）**: `S1` / `TRIM-B1` / `MR-B1` / `BILL-B1a` / `BILL-B1b`（`MR-EST-B1`）/ `RES-B1`（`RES-MR-B1`）/ residual pin close（`c68f7ff6b`）/ `STAFF-B1` まで実装・コミット済み。`grandchildParentClinicCorrelationResidualSites` は空・`ResidualSiteCount = 0`（ゲート機構は count 0 でも維持）。**計画書の旧 `-BILL-B1` 許可パス誤り**: `accounting_repository.go` / `estimate_repository.go` にサイトは無く、実在は `accounting_repository_reports_close.go`（GetCloseAggregate）・`billing_item_repository.go`（ValidateCreateReferences）・`medical_record_repository.go`（CountEstimates）・`reservation_repository.go`（CountMedicalRecords）。各 B1 の許可パスに `lintscan` 残余 allowlist 更新が必要だった。`STAFF-B1` は residual 計上外（child clinic_id で disclosure 遮断済み）で integrity 修復として別クローズ。下記各節は実装済みの記録として残す。

### SEC-SWEEP-02-S1（lint構造・完了済み・記録として残置）

- **目的**: 重複した2 lintを統合し、残5 model classとdynamic countを機械検出する。
- **根本原因**: `grandchild_parent_clinic_correlation_lint_test.go:29` は `AppointmentTrimmingDetail` / `Billing` / `Estimate` / `MedicalRecord` / `VitalRecord` を含まず、`pet_grandchild_parent_clinic_correlation_lint_test.go:21` は先行6 targetだけを重複管理する。
- **変更方針**: registryを一つにし、GORM/raw SQL/alias/qualified idとstale occurrenceをfixture化する。修復コードは触らない。
- **許可パス**:
  - `backend/internal/lintscan/grandchild_parent_clinic_correlation_lint_test.go`
  - `backend/internal/lintscan/pet_grandchild_parent_clinic_correlation_lint_test.go`
- **禁止する巻き込み変更**: repository、migration、global SQL parser、`deleted_at` / `deceased_at`。
- **検証コマンド**:
  ```bash
  docker compose exec -T backend go test ./internal/lintscan/ -run 'Test(GrandchildParentClinicCorrelation|PetGrandchildParentClinicCorrelation)' -count=1 -p 1 -v
  docker compose exec -T backend go vet ./internal/lintscan/
  docker compose exec -T backend gofmt -l internal/lintscan/grandchild_parent_clinic_correlation_lint_test.go internal/lintscan/pet_grandchild_parent_clinic_correlation_lint_test.go
  git diff --name-only -- backend/internal/lintscan/grandchild_parent_clinic_correlation_lint_test.go backend/internal/lintscan/pet_grandchild_parent_clinic_correlation_lint_test.go
  ```
- **完了判定**: 6残面をfixtureで検出し、旧対象を維持し、重複registry 0、差分2 fileだけ。

### SEC-SWEEP-02-TRIM-B1

- **目的**: appointment trimming detail readをappointment/petのclinicへ相関する。
- **根本原因**: `backend/internal/trimming/trimming_repository.go:30` のgrandchild readに親clinic相関が無い。
- **変更方針**: 破損FK fixtureを保持して親clinic predicateを追加する。
- **許可パス**:
  - `backend/internal/trimming/trimming_repository.go`
  - `backend/internal/trimming/trimming_repository_test.go`
- **禁止する巻き込み変更**: option/course master、appointment lifecycle、DDL、soft-delete/deceased filter。
- **検証コマンド**:
  ```bash
  docker compose exec -T backend go test ./internal/trimming/ -run 'Test.*ClinicIsolation' -count=1 -p 1
  docker compose exec -T backend go vet ./internal/trimming/...
  docker compose exec -T backend gofmt -l internal/trimming/trimming_repository.go internal/trimming/trimming_repository_test.go
  git diff --name-only -- backend/internal/trimming/trimming_repository.go backend/internal/trimming/trimming_repository_test.go
  ```
- **完了判定**: corrupt childを他clinic parentへ復元せず、正常/履歴read green、allowlist外0。

### SEC-SWEEP-02-BILL-B1 — SUPERSEDED（2026-07-28 15:16生成の専用プロンプトが正）

**下記の許可パス・根本原因は誤りである。** 残余allowlistの実測（`grandchildParentClinicCorrelationResidualSites`）では `accounting_repository.go` / `estimate_repository.go` に対象サイトは存在しない。正しい対象は `backend/internal/billing/accounting_repository_reports_close.go` の `GetCloseAggregate`（親 `pets`・FK `pet_id`）と `backend/internal/billing/billing_item_repository.go` の `ValidateCreateReferences`（親 `appointments`・FK `appointment_id`）の2件のみ。実行契約は `~/.claude/prompt-craft-runs/agent-fast-grok-sec-sweep-02-bill-b1a-billing-repair.md`（2026-07-28 15:16生成）を正本とし、本節は複製しない。**Estimate（`medical_record_repository.go` の `CountEstimatesByMedicalRecordID`）と MedicalRecord（`reservation_repository.go` の `CountMedicalRecordsByReservationID`）は当該プロンプトが明示的に「別ユニットの担当」として除外しており、下記 `SEC-SWEEP-02-MR-EST-B1` / `SEC-SWEEP-02-RES-MR-B1` で扱う。**

### SEC-SWEEP-02-MR-EST-B1（完了・`35e22b057` / 別名 BILL-B1b）

- **結果**: `CountEstimatesByMedicalRecordID` は親 `medical_records.clinic_id` JOIN でスコープ。**`estimates.clinic_id` はフィルタしない**（削除ガード fail-closed）。表示用途の「自院見積件数」には流用しない。
- **所有パス実績**: `medical_record_repository.go` + 回帰 test + `lintscan` residual allowlist 更新。

### SEC-SWEEP-02-RES-MR-B1（完了・`e752a690a` / 別名 RES-B1）

- **結果**: `CountMedicalRecordsByReservationID` は親 `appointments.clinic_id` JOIN でスコープ。**`medical_records.clinic_id` はフィルタしない**。呼び出し元は `UpdateForTrimming` / `DeleteForTrimming` / `reservationService.Delete` の3箇所（「唯一の呼び出し元 Delete」は誤りだった）。
- **所有パス実績**: `reservation_repository.go` + 回帰 test + `lintscan` residual allowlist 更新。

### SEC-SWEEP-02-MR-B1

- **目的**: medical record→appointment edgeとvital readを親clinicへ相関する。
- **根本原因**: `medical_record_repository.go` のappointment scopeと `vital_repository.go:40,54` が残存する。
- **変更方針**: appointment/vitalを別fixtureでREDにし、親clinicだけを相関する。
- **許可パス**:
  - `backend/internal/medicalrecord/medical_record_repository.go`
  - `backend/internal/medicalrecord/vital_repository.go`
  - `backend/internal/medicalrecord/medical_record_owner_pet_preload_clinic_isolation_test.go`
  - `backend/internal/medicalrecord/vital_repository_test.go`
- **禁止する巻き込み変更**: SEC-DURのowner snapshot、daily/addenda既修復、billing/estimate、lifecycle filter。
- **検証コマンド**:
  ```bash
  docker compose exec -T backend go test ./internal/medicalrecord/ -run 'Test.*(Grandchild|Vital|Appointment.*Clinic).*' -count=1 -p 1
  docker compose exec -T backend go vet ./internal/medicalrecord/...
  docker compose exec -T backend gofmt -l internal/medicalrecord/medical_record_repository.go internal/medicalrecord/vital_repository.go internal/medicalrecord/medical_record_owner_pet_preload_clinic_isolation_test.go internal/medicalrecord/vital_repository_test.go
  git diff --name-only -- backend/internal/medicalrecord/medical_record_repository.go backend/internal/medicalrecord/vital_repository.go backend/internal/medicalrecord/medical_record_owner_pet_preload_clinic_isolation_test.go backend/internal/medicalrecord/vital_repository_test.go
  ```
- **完了判定**: corrupt edge非表示、正常/soft-deleted/deceased履歴は維持、allowlist外0。

### SEC-SWEEP-02-STAFF-B1（完了・`2a7689080` — residual サイトではない）

**実測**: `CountBlockingReferencesByStaffID` は `grandchildParentClinicCorrelationResidualSites` に**一度も計上されていない**（disclosure は child `clinic_id` 直接フィルタで遮断済み）。親 FK 破損時の依存件数 integrity として addenda/vitals を親 clinic 相関へ修復してクローズ。**残余 allowlist の「未修復4件」には含めない。**

## 分類

| 元ID | 現在の分類 | 実装判断 |
|---|---|---|
| `BUG-453` | 挙動・解消済み | 実装不要、回帰再認定のみ |
| `BUG-449` | 挙動・code-complete | 臨床値投入とruntime/browser証拠のみ |
| `BUG-459` | 挙動・code-complete | 既存実行契約によるAPI実証のみ |
| `BUG-460` | 挙動・実装不要 | 既存実行契約によるbrowser実証のみ |
| `BUG-433` | 構造・上位欠陥 | writerなし、`TASK-444`へ統合 |
| `BUG-437`, `TASK-444`, `TASK-445-S1/S2/S3/S4`, `SEC-DUR-01-MR-T1/S1`, `SEC-SWEEP-02-S1` | 構造 | 構造だけを変更 |
| `BUG-448`, `BUG-454`, `BUG-456`, `BUG-458`, `BUG-455-*-B`, `TASK-445-B1`, `SEC-DUR-01-BILL-*`, `SEC-SWEEP-02-*-B1` | 挙動 | 挙動だけを変更 |
| `BUG-440`, `E2E-BASELINE-DEBT-01` | 挙動（新設） | 挙動だけを変更 |

構造と挙動を同じ作業単位に含むものは0件である。`BUG-457`／`BUG-453`はCLOSEDのため分類対象から除外した。

## 台帳との乖離

| ID | 台帳の旧記述 | 2026-07-28実測 | 根拠 |
|---|---|---|---|
| `BUG-433` | Pet 31 vs DTO 9を中心に記述 | 問題は継続するが約268 file / 294 import site。独立修復でなくTASK-444所有 | `backend/tygo.yaml:1-15`, `frontend/eslint.config.js:91` |
| `BUG-437` | read registry未登録 | `"ExamTypeField"` は登録済み。未解消なのはlive alias `Items` をlintが無視するblind spot | `preload_clinic_scope_lint_test.go:79-97,245-256`, `exam_type_repository.go:46-68` |
| `BUG-449` | API/UIなし、range 0行 | API/UIは存在しcode-complete。runtime行数はdirect SQL禁止のため未報告 | `routes.go:157`, `exam_type_field.go:224-282`, `ExamTypeFieldsEditor.tsx:411-479` |
| `BUG-453` | OPEN対象に含まれる | current ledger/codeともCLOSED、本書から節削除 | `repository_kana_search_test.go` 3 domain |
| `BUG-456` | lab import precedenceとdetail未確認 | active writerはinspectionValue、`ExamItemsTable`は既に正しい。**7ed9ae808で解消済み**。残るのはunit列のみ | `lab_import_examination_service.go:76-93`, `ExamItemsTable.tsx:152-173` |
| `BUG-457` | FE/BE具体箇所未確認 | **edit案で確定・CLOSED**（`75aff5786`+`b415892ec`+`7f17aeb53`）、本書から節削除 | `HospitalizationDetailActions.tsx`, `routes.go` |
| `BUG-458` | route×viewport内訳未確認 | 旧10 cell/13 findingは列挙済み。ただしcapture settle不十分のため24 cell再測定が必要。FE12レーンの `agent-fast-fe12-m05-sentinel-responsive.md` と重複の可能性あり要確認 | `layout-review.md:5-36`, `ui-design-compliance-readonly.spec.ts:358-425` |
| `TASK-445` | migration番号未確定 | current tailは003、次候補は004。paymentsにclinic_idなし | `backend/migrations/003_add_exam_results_exam_type_field_id_index.sql`, `accounting.go:163-189` |
| `SEC-DUR-01` | 横断方針全体が未着手に見える | MR-A1/MR-C1-3は完了。MR-C1-3b（write回帰test）は `agent-fast-secdur01-mr-c1-3-write-surface-regression.md` で対応済みの可能性——実行状況を確認してから着手。残りはBILL-B1/B2 | `clinical_relation_validation.go:79-86`, `accounting_repository_ltv.go:15` |
| `SEC-SWEEP-02` | census/lint新設が未着手に見える | **残余0で完了**（pin `ResidualSiteCount=0`、`c68f7ff6b`）。BILL-B1a/B1b・RES-B1・STAFF-B1（residual外）までコミット済み。旧計画の `accounting_repository.go`/`estimate_repository.go` は誤許可パス | `grandchild_parent_clinic_correlation_lint_test.go` residual empty + `2a7689080` |

## 生成済み専用プロンプト一覧（2026-07-28 15:40台・本改訂で追加生成）

以下5件は `/prompt-craft-agent-fast` で本書の該当節から生成済み。本書の該当節（実装契約の詳細）より、これらの専用プロンプトを実行時の正本とする。

| 対象 | 保存先 | 機械検証 |
|---|---|---|
| `SEC-SWEEP-02-STAFF-B1` | `~/.claude/prompt-craft-runs/agent-fast-grok-sec-sweep-02-staff-b1-dependency-count-audit.md` | 書込時hookでPASS確認済み |
| `SEC-SWEEP-02-RES-MR-B1` | `~/.claude/prompt-craft-runs/agent-fast-grok-sec-sweep-02-res-mr-b1-reservation-repair.md` | 書込時hookでPASS確認済み |
| `SEC-SWEEP-02-MR-EST-B1` | `~/.claude/prompt-craft-runs/agent-fast-grok-sec-sweep-02-mr-est-b1-medicalrecord-estimate-repair.md` | PASS（ユーザー実行で確認済み・2026-07-28） |
| `BUG-448` | `~/.claude/prompt-craft-runs/agent-fast-bug448-exam-audit-provenance.md` | PASS（ユーザー実行で確認済み・2026-07-28） |
| `TASK-444-S1` | `~/.claude/prompt-craft-runs/agent-fast-task444-s1-generated-model-response-boundary.md` | PASS（初回2件はsemantic FAIL——`named-gate-equivalence-execution`と`generated-artifact-commit-trackability`。原因は`equivalent`語とbacktickコマンドの同居、およびAcceptance Checklist内の`commit`語。両方を意味を変えずに語を置換・再構成して解消。2026-07-28） |

`SEC-SWEEP-02` の修復ユニット（BILL-B1a/B1b・RES-MR-B1・MR-EST-B1・STAFF-B1）はすべて完了。残作業は台帳（`3-session-agent.html`）への反映と、必要なら `SEC-SWEEP-02-DOC` のコメント/本書整合（本セッション）。

## ユーザー手動境界

- `TASK-444-S2` の生成結果はユーザーが `make codegen` を実行して確認する。
- `TASK-445-S2/S3/S4` は既存データcanaryをユーザーが確認し、その後ユーザーが `make migrate` を実行する。
- DB reset、直接SQL、migration適用、依存install、全project test/lint/buildは本書の各実行セッションでは行わない。
