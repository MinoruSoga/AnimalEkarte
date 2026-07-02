## Objective
3監査クラスタ(read master-Preload 漏洩 / write master-FK 未検証 / write parent-FK 未検証)の再発を機械的に防ぐため、最も静的検出が確実な「clinic-scoped マスタの Preload は clinic_id 述語必須」を go/ast ベースのチェックとして実装し CI に組み込む。write 側 master-FK 検証は静的実現可能性を評価し、可能なら実装・不可能なら理由と代替を文書化する。

## Scope
- In scope:
  - **read Preload ルールの機械強制（P0 主軸）**: `backend/internal/repository/` 配下の GORM `Preload("<Assoc>", "<predicate>")` 呼び出しを go/ast で走査し、clinic-scoped マスタ関連の Preload で predicate 文字列に `clinic_id` を含まないものを**検出して fail** させるチェック。
  - **対象マスタ一覧の確定**: P3.1(`repository/CLAUDE.md`)＋実モデルから clinic-scoped マスタ関連名を列挙（Vaccine/Medicine/Procedure/Consultation/ReservationType(+Group)/TrimmingCourse/TrimmingOption(TrimmingDetail.Course/Options)/Cage/Insurance/ExaminationType/CheckupType/DiagnosisType(+Names)/DiagnosisName 等）。
  - **allowlist（誤検出防止・必須）**: ①global マスタ(AnimalSpecies/ManualArticle = clinic_id 無し)を除外。②**Staff/Doctor 履歴 preload の P3.1 例外**（EnteredByStaff/PaidByStaff/ClosedByStaff 等の意図的 unscoped・reservation の Doctor/CreatedByStaff のみ assignment-EXISTS）を allowlist。allowlist は一覧化し根拠コメントを付ける。
  - **bad/good フィクスチャ test**: (i)`Preload("Vaccine", "deleted_at IS NULL")` 等の既知 bad パターン（read 監査が修正したもの）を**検出して fail** を実証 (ii)現行(修正済)コードと allowlist 例外は**pass** を実証。
  - **CI 配線**: チェックを test/lint パイプラインで実行（既存の go/ast audit-taxonomy チェックと同経路）。
  - **既存の go/ast 先例の再利用**: コミット履歴の「go/ast exhaustiveness check for audit taxonomy」の仕組みを inspect し、同じ枠組み/配置で実装（新規フレームワーク導入しない）。
  - **write master-FK ルールの静的実現可能性評価**: request 由来 master FK が write 前に `FindByID(clinicID,…)` で検証されるかを静的に検出できるか評価。確実に実現可能なら実装＋fixtures、データフロー解析が要り信頼できる静的検出が不可能なら**その理由を文書化し代替**（convention/命名規約 test・reviewer gate・CLAUDE.md ルール補強）を提案する（偽装した部分カバレッジを PASS にしない）。
- Out of scope:
  - 既存 audit 修正の再実装(72e8887c/b3638d5e/03bf1cb5/f4e7b7a7 で完了)。
  - 残 MEDIUM/LOW master-FK write の修正(別タスク)。
  - golangci-lint カスタムプラグインのビルド（より単純な go/ast or CI スクリプトで充足するなら不要）。
  - parent-FK(72e8887c)の静的強制（read Preload を主軸とし、余力で言及）。
  - 新規 migration・FE・push(external write・別承認)。

## Success Criteria
- read Preload チェックが、clinic-scoped マスタの Preload で clinic_id 述語を欠くものを**全て検出して fail** する（bad フィクスチャで実証）。
- チェックが**現行の修正済コードと allowlist 例外に対し pass**（false positive ゼロ・good フィクスチャ＋実コードで実証）。
- **旧失敗モード非再発**: 既知 bad パターン（例 `Preload("Vaccine", "deleted_at IS NULL")`）を再導入するとチェックが fail することを test で固定。
- チェックが CI（test/lint 経路）で実行され、fail 時に CI が赤くなる。
- allowlist が一覧化され、各除外（global マスタ・Staff 履歴 P3.1 例外）に根拠が付く。
- write master-FK ルールは「機械強制を実装」または「静的不可能の理由＋代替提案」のいずれかで明示的に決着（曖昧な未対応を残さない）。

## Acceptance Checklist
Before implementation, expand every Success Criteria item into checklist items. Each item: Expected behavior / Target surface / Verification method / PASS evidence. Mark each PASS/FAIL/BLOCKED with evidence; do not produce the final answer while any item remains FAIL.

- [ ] Expected behavior: clinic-scoped マスタ一覧の確定 | Target surface: P3.1(repository/CLAUDE.md) + 実モデル | Verification method: P3.1 列挙と model の clinic_id 有無を突合し対象/除外を確定 | PASS evidence: 対象マスタ表 + 除外(global/Staff 例外)の根拠
- [ ] Expected behavior: Preload チェックが bad を検出 fail | Target surface: go/ast チェック + bad フィクスチャ | Verification method: clinic_id 述語を欠く Preload フィクスチャでチェックが非ゼロ exit/fail | PASS evidence: bad フィクスチャでの fail 出力
- [ ] Expected behavior: チェックが現行コード+allowlist で pass | Target surface: 実 repository コード + allowlist | Verification method: 現行コード(修正済)と Staff 例外/global で false positive ゼロ | PASS evidence: 実コードに対する pass 出力 + allowlist 一覧
- [ ] Expected behavior: 旧失敗モード非再発 | Target surface: 再導入 bad パターン | Verification method: `Preload("Vaccine","deleted_at IS NULL")` 再導入フィクスチャで fail を固定 test | PASS evidence: 再導入 fail test 出力
- [ ] Expected behavior: CI 配線 | Target surface: test/lint パイプライン(.github/workflows or 既存 go/ast 経路) | Verification method: チェックが CI step として実行される配線を確認 | PASS evidence: CI 設定 diff + ローカル実行コマンド出力
- [ ] Expected behavior: write master-FK ルールの決着 | Target surface: 静的解析可否評価 | Verification method: 実現可能なら fixtures 付き実装、不可能なら理由＋代替を文書化 | PASS evidence: 実装+test、または不可能理由+代替提案
- [ ] Expected behavior: 既存無回帰 | Target surface: 触れた package + CI | Verification method: スコープ test + go vet、CI が既存 step を壊さない | PASS evidence: scoped test 出力

Do not start implementation until every Success Criteria item has a checklist item with a deterministic verification method.

## Constraints
- Docker 必須・スコープ実行のみ。フルリポ test/lint/DB reset 自動禁止（`go test ./internal/repository/` フルは pool 枯渇）。
- **既存の go/ast 先例（audit-taxonomy exhaustiveness check）の枠組み・配置を再利用**し新規フレームワーク/プラグインを導入しない（KISS）。
- 対象マスタ一覧は P3.1＋実モデルの clinic_id 有無を真実とする。**Staff 履歴 preload の P3.1 例外と global マスタを必ず allowlist**（誤検出は CI を壊し開発を阻害する＝回帰）。
- チェックは false positive ゼロを優先（現行修正済コードを fail させない）。検出ロジックは Preload の predicate 文字列に `clinic_id` を含むか、で判定（過剰に賢くしない）。
- 既存 CI step・ビルドを壊さない。CI 配線変更時は `.github/workflows` env/with も確認。
- write master-FK の静的検出を「できるふり」で部分実装して PASS にしない。実現可能性を正直に評価し決着。
- 既存 migration 編集禁止。push 実行しない(別承認)。秘密を貼らない。

## Risk Tier
- 監査/設計=Read-only / 実装=Local write（go/ast チェック + test + CI 配線）。security-enforcement ツールだが runtime 挙動は変えない。
- 範囲内に External write なし（CI 設定はローカル編集・push は別承認）。誤検出による CI 阻害が主リスク → allowlist + false-positive ゼロ test で排除。

## Truth Source Priority
1. 実行可能チェック: 既存 go/ast audit-taxonomy チェック・go test・実モデルの clinic_id 列・修正済 repository コード(b3638d5e)
2. 現コード挙動(各 repository の Preload・allowlist 対象の Staff 例外)
3. P1-P18・P3/P3.1/P4 規則
4. `repository/CLAUDE.md`(P3.1 マスタ一覧・Staff 例外)・refs
5. コメント・旧サマリー
注: マスタが clinic-scoped か global かは実モデルの clinic_id 列が真実。Staff 例外は P3.1 と reservation_repository の実装が真実。

## Verification Strategy
- 静的チェック自体の検証: bad/good フィクスチャ（clinic_id 欠落 Preload を fail・修正済/allowlist を pass）+ 実 repository コードに対する実行（false positive ゼロ）。
- 旧失敗モード: 再導入 bad パターンで fail を固定。
- CI: チェックが pipeline step として走る配線 + ローカル再現コマンド。
- 触れた package の scoped test + go vet。
検証順序: ①bad/good フィクスチャ test → ②実コードに対する false-positive ゼロ確認 → ③CI 配線確認。
各 Acceptance Checklist 項目は上記を名指し。決定的検証未実行は PASS にしない。

## Harness Improvement Feedback
- 本タスク自体が P0 harness improvement の実装。完了後、write master-FK・parent-FK の機械強制が静的に困難なら、その分を P1（convention test/reviewer gate）として残課題に明記。
- スコープ外の追加改善は follow-up 報告。

## Harness Contract
Acceptance Checklist(4項目構造)・Harness Selection・Execution Loop Selection・Spec-Implementation Reconciliation Loop・Failure Signature log・Loop Control(2回で縮小・3回で BLOCKED)・De-Sloppify・Independent Review Gate・Subagent Orchestration・loop-status evidence・FAIL残存時の最終報告禁止・Saved Prompt Validation Gate・Eval Regression Capture を保持。
Validate saved prompts with `node ~/.claude/scripts/prompt-craft-harness-validate.js <file>`. The PreToolUse enforcement hook `node ~/.claude/scripts/prompt-craft-harness-enforce-hook.js` blocks invalid saves; self-test it with `node ~/.claude/scripts/prompt-craft-harness-enforce-hook.js --self-test`. 検証 exit 0 か逐項目手動確認時のみ「検証済み」と称する。

## Saved Prompt Validation Gate
This prompt must be saved to a file and validated before execution: save the prompt, then validate it with `/prompt-craft-validate <prompt-file>` (i.e. `node ~/.claude/scripts/prompt-craft-harness-validate.js <file>`) before execution. Chat-only output is not validated — if this is chat-only prompt output, note `prompt validation not run: chat-only output` in Deliverables. Local write・CI 配線変更・harness 強化に関わるため本ゲートは必須。

## Eval Regression Capture
取りこぼし・弱いループが出たら保存して `node ~/.claude/scripts/prompt-craft-eval-add.js <file> --invalid --name clinic-boundary-lint-enforcement` で eval corpus(eval fixture) 追加 → `--eval-dir` 再実行。

## Subagent Orchestration
Default routing: planner for planning/risk analysis; code-explorer/researcher for inventory; architect for design; code-reviewer (reviewer role) and go-reviewer for code quality; security-reviewer for vulnerability-class coverage; tdd-guide for bad/good fixtures; loop-operator (loop-status) for loop monitoring; harness-optimizer for harness review.
Task-specific selection:
- `architect`: チェック設計（go/ast vs CI スクリプト選定・既存 audit-taxonomy 先例の再利用・write 側の静的実現可能性評価）。
- `go-reviewer`: 静的解析コードの idiom・偽陽性/偽陰性・allowlist 設計。
- `security-reviewer`: チェックが脆弱性クラスを実際に網羅するか・bypass(述語に "clinic_id" を含むが別用途の文字列等)が無いか。
- `tdd-guide`: bad/good フィクスチャ。
- `code-explorer`: 全 Preload 呼び出し + 既存 go/ast チェックの配置の inventory。
read-only 調査/設計は並列、編集は main agent 単独。Treat subagent output as evidence, not authority: cross-check each finding against the Acceptance Checklist before acting. The main agent owns final integration and all edits. If a routed agent is missing or none were used, state the fallback and why in the report. 採否を記録。

## Harness Selection
- Chosen: tdd
- Why: 成果物は「既知 bad を検出し good を pass する静的チェック」＝bad/good フィクスチャで検証可能な決定的振る舞い。RED(bad 検出失敗)→GREEN(検出する) で駆動できる。`construction` はエージェント自身の action space 向けで本タスク(プロジェクト CI ツール)には不適。tdd-workflow 未導入時は手動 RED-GREEN にフォールバック。

## Execution Loop Selection
- Chosen: sequential(+ de-sloppify overlay)
- Why: 単一サーフェス（静的チェック + フィクスチャ + CI 配線）の一筆書き。並列バリアントや merge queue 不要。コード/テスト変更ありで de-sloppify を重ねる。
- The Spec-Implementation Reconciliation loop runs inside the selected loop (reconciliation stays the inner authority).
- Stop condition: 全 Acceptance Checklist PASS（チェックが bad 検出/good pass・CI 配線・write ルール決着）、または genuine blocker 文書化。

## Spec-Implementation Reconciliation Loop
各項目 expected vs actual を証拠(フィクスチャ test 出力・実コード実行・CI 配線 diff)で PASS/FAIL/BLOCKED。決定的検証未実行は PASS 禁止。FAIL は最小修正+Failure Signature 記録、新 root-cause でのみ再試行。FAIL 残存中は最終報告禁止。特に「実コードを false-positive で fail させる」FAIL を実コード実行で捕捉する。

## Loop Control
loop-status ツール(`loop-status.js` / ecc loop-status)があれば記録、無ければ「monitoring 不可」と明記し Failure Signature ログで継続。安全境界(push/merge・破壊・認証・CI 破壊・不可逆)で即停止。
Failure Signature: <checklist item> | <expected behavior> | <actual behavior> | <verification command/check> | <error signature> | <attempt number> | <attempted fix> | <result>
- If the same failure signature appears twice, reduce scope to the smallest failing unit.
- If the same failure signature appears three times, mark the item BLOCKED with the exact blocker and required input.

## De-Sloppify Pass
コード/テスト変更後: framework 挙動だけのテスト除去・過剰防御除去・デバッグログ/無関係リファクタ除去。bad/good フィクスチャと allowlist は維持。影響した最小検証を再実行。

## Independent Review Gate
security-sensitive(脆弱性クラスの網羅性)・CI に関わるため、reconciliation 後・報告前に fresh review(`/code-review` or security/architect/go)。チェックが脆弱性クラスを実際にカバーするか・bypass の有無・false positive ゼロ・allowlist の妥当性・CI 配線を Acceptance Checklist と突合。CRITICAL/HIGH 修正後に再 reconciliation。Do not produce the final answer while any item remains FAIL.

## Execution Flow
1. Risk Tier 分類(実装=Local write・CI 配線・push 範囲外)。
2. 既存 go/ast audit-taxonomy チェック・全 Preload 呼び出し・P3.1 マスタ一覧/Staff 例外・修正済コードを inspect。
3. Success Criteria を Acceptance Checklist へ展開。
4. Subagent 計画→architect(設計)・code-explorer(inventory)・security/go を早期起動。
5. Harness Selection=tdd。
6. Execution Loop=sequential+de-sloppify。
7. /plan 出力後、確認待ちせず進む。
8. /tdd: bad フィクスチャで RED(検出されない)→チェック実装で GREEN(検出する)→good/allowlist で false-positive ゼロ確認。
9. De-Sloppify→最小再検証。
10. Spec-Implementation Reconciliation Loop(Loop Control 準拠)。
11. /code-review or Independent Review Gate(security/architect/go)。CRITICAL/HIGH 修正→再 reconciliation。
12. /harness-audit。
13. write master-FK/parent-FK の機械強制が静的困難なら P1 残課題として報告。
14. 全 Acceptance Checklist PASS or blocker 文書化時のみ完了。push 未実行(commit+scoped GREEN+CI 配線まで)。

## Session State
多フェーズ(マスタ一覧確定→チェック実装→allowlist→CI 配線→write ルール評価)。完了ステップ/現ブロッカー/残項目/実行済みフィクスチャ結果/次アクションを表で維持。

## Execution Rules
- in-scope を end-to-end 実行。mid-task 確認しない(安全境界承認時のみ)。
- 安全境界(push/merge/deploy・破壊・認証・CI 破壊・不可逆)で停止。push 範囲外。commit 前に HEAD 確認(並行衝突警戒)。
- 最小差分・既存 CI/ビルドを壊さない。false positive は回帰として排除。write ルールを「できるふり」で PASS にしない。
- 秘密を貼らない。

## Context
- 本タスクは3監査クラスタ(72e8887c parent-FK / b3638d5e read master-Preload / 03bf1cb5+f4e7b7a7 write master-FK)の**再発防止 P0**。同じ欠陥クラスを手作業で3回潰したため機械強制する。
- 既存先例: コミット履歴「go/ast exhaustiveness check for audit taxonomy validator maps」(audit taxonomy の go/ast チェック)。同じ枠組みを再利用。
- ルール出典: `repository/CLAUDE.md` P3.1（clinic-scoped マスタ Preload は clinic_id 述語必須・対象マスタ一覧・Staff 履歴例外・global マスタ除外）。
- 関連: cross_tenant_read_idor_audit_20260629(b3638d5e の Preload 修正パターン)・cross_tenant_master_fk_write_audit_20260629・base.go(clinicScope)・各 `*_repository.go` の Preload。
- 検出対象パターン: `Preload("<ClinicScopedMaster>", "<predicate without clinic_id>")`。allowlist: global マスタ・Staff 履歴 preload。

## Deliverables
- Root-cause summary(なぜ機械強制が3監査クラスタの根本予防か)
- Minimal patch (or patch plan)(go/ast チェック + フィクスチャ + CI 配線)
- Acceptance checklist with PASS/FAIL/BLOCKED status and evidence
- Saved Prompt Validation Gate result, or `prompt validation not run: chat-only output`
- Eval Regression Capture result, or "not needed"
- Failure Signature log, or "none"
- Verification strategy used for each checklist item
- Risk Tier and safety boundary assessment
- Environment-specific harness inspection result(既存 go/ast チェック・CI workflows・CLAUDE.md P3.1)
- Subagent Orchestration result(使用 agent・採否・理由)
- Harness Selection result(tdd・理由)
- Execution Loop Selection result(sequential+de-sloppify・理由・stop condition)
- Loop control result(iteration 数・loop-status or 不可明記)
- De-sloppify pass result
- Independent review result(security/architect/go)
- Session State(多フェーズ進捗)
- Harness improvement feedback(本タスクが P0 実装・write/parent-FK 静的困難なら P1 代替)
- Verification result (command outputs)
- Regression checks completed(既存 CI step・ビルド無回帰・false positive ゼロ)
- Remaining risks and follow-ups(write master-FK/parent-FK の機械強制が静的困難な分・残 MEDIUM/LOW)
