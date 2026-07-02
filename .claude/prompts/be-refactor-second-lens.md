## Objective
backend を初回リファクタ(DRY/dead-code/P1-P18)が深く見なかった4レンズ — error-handling 一貫性 / ファイルサイズ・凝集度(800行超) / 並行安全性 / テスト品質 — で verify-first に監査し、behavior-preserving に直せる confirmed 単位のみ修正する。behavior-changing な発見(error 契約変更・race・swallow)は直さず別バグ報告。計画先行・小コミット・テスト緑維持・main 直コミット(push なし)。

## Scope
- In scope:
  - Phase 0(計画先行・verify-first 監査): 4レンズで backend を静的監査し、各候補を「behavior-preserving-fix(修正)/ behavior-changing-report(報告のみ)/ clean(no-op)」に分類した優先順位付きバックログを出力してから着手。「コードは劣化している」と仮定しない。
  - ファイルサイズ・凝集度: >800行ファイルを責務単位で同一 package 内に分割(純機械的移動・build+テスト緑・挙動不変)。凝集して大きいだけのファイルは「分割不要」と根拠付き判定。
  - テスト品質(De-Sloppify 主戦場): 言語/framework/mock の挙動だけを検証するテスト・暗黙ゼロ値依存の脆いテストを除去/堅牢化。ビジネスロジック・回帰・境界検証テストは残す。実挙動のカバレッジを落とさない。
  - error-handling 一貫性のうち behavior-preserving なもの: 返却 error 契約(wire レスポンス)を変えない範囲の内部 wrapping/logging 整理のみ。
  - 並行安全性のうち behavior-preserving なもの: 意味等価な goroutine/同期パターンの抽出のみ(挙動・出力不変)。
  - security/clinic_id/auth コードを触る場合は characterization テスト緑が前提(初回 refactor と同条件)。
- Out of scope:
  - **behavior-changing な発見は本タスクで直さない・報告のみ**: error 契約変更・race 修正・swallow された error の挙動修正・concurrency バグ修正。これらは別バグ Issue。
  - 初回カバー済みレンズ(DRY/関数抽出・dead-code・P1-P18 レイヤリング)の再実施。
  - 機能追加・API/挙動/DBスキーマ変更・パフォーマンス再設計。
  - frontend/docs/CI/インフラ変更。リモート push/PR/マージ(コミットまで)。

## Success Criteria
- SC0(Phase 0 監査・verify-first): 4レンズ全てを監査し、各候補が fix/report/clean に分類された優先順位付きバックログがコード根拠(file:line)付きで出力されている。
- SC1(挙動不変・最重要): 修正した各 package のテストが before/after で同一に緑。挙動・公開API・error 契約・wire レスポンスに差分0。
- SC2(ファイルサイズ・凝集度): >800行ファイルが特定され、分割したものは build+package テスト緑かつ挙動不変、分割しなかったものは「凝集して妥当」の根拠あり。
- SC3(テスト品質): framework/mock 挙動のみ検証するテスト・脆い暗黙ゼロ値テストが特定され、除去/堅牢化済み。ビジネスロジック・回帰テストは温存、実挙動カバレッジ不低下。
- SC4(behavior-changing の報告): error 契約変更・race・swallow 等の behavior-changing 発見が、修正されず別バグとして file:line + 影響付きで報告されている(本タスクで直さない)。
- SC5(security/clinic_id 不変): 触れた auth/clinic_id コードに characterization テストが緑であり、#201 ガード(FindByID(ctx,clinicID,id) 等)が削除/弱体化されていない。
- SC6(限定変更・コミット規律): 変更が behavior-preserving に限定、API/契約差分0。各コミット独立緑・小粒・push なし。
- SC7(no-op 正直): clean と判定したレンズ/候補は「直さない」が正しく、根拠付きで報告。捏造 churn 0。

## Acceptance Checklist
Each checklist item states the expected behavior, the target surface, the verification method, and the PASS evidence. 実装前に各 Success Criteria を検証可能単位へ展開し、決定的検証手段が割り当たるまで着手しない。

- [ ] SC0-a | Expected: 4レンズ監査+分類バックログ | Target surface: backend 全層 | Verification: lens ごとに候補を grep/codegraph/agent で抽出し fix/report/clean 分類(file:line) | PASS evidence: 優先順位付きバックログ
- [ ] SC1-a | Expected: 修正 package の挙動不変 | Target surface: 変更 package | Verification: 変更前 scoped テスト緑→修正→同一 scoped テスト緑(`docker compose exec backend go test ./internal/<pkg>/...`) | PASS evidence: before/after 同一緑出力
- [ ] SC2-a | Expected: >800行分割 or 妥当判定 | Target surface: >800行ファイル | Verification: `wc -l` で >800 列挙→分割は build+テスト緑、非分割は凝集根拠 | PASS evidence: 行数・build/テスト緑 or 非分割根拠
- [ ] SC3-a | Expected: 低価値テスト除去・脆さ堅牢化 | Target surface: 対象 *_test.go | Verification: framework/mock-only・暗黙ゼロ値特定+除去/堅牢化後 package 緑+カバレッジ不低下 | PASS evidence: 除去/変更一覧+緑+カバレッジ非退行
- [ ] SC4-a | Expected: behavior-changing を報告(直さない) | Target surface: error/concurrency 発見 | Verification: file:line・なぜ behavior-changing・影響を列挙、diff 非包含 | PASS evidence: 別バグ報告リスト+diff 非包含
- [ ] SC5-a | Expected: security/clinic_id 不変 | Target surface: 触れた auth/clinic_id | Verification: characterization 緑+#201 ガード grep 残存 | PASS evidence: 緑+ガード残存(非接触なら明記)
- [ ] SC6-a | Expected: 限定・契約不変・規律 | Target surface: git diff/log | Verification: behavior-preserving のみ・契約不変、小コミット独立緑・push なし | PASS evidence: diff レビュー+コミット一覧
- [ ] SC7-a | Expected: no-op 正直 | Target surface: clean 判定 | Verification: clean 根拠を file:line で提示 | PASS evidence: clean 根拠・捏造 churn 0

## Constraints
- verify-first。劣化を仮定しない。confirmed のみ修正、clean は no-op 報告。
- behavior-preserving のみ。error 契約変更・race 修正・swallow 挙動修正は直さず報告。
- 計画先行: Phase 0 バックログを出してから実装。
- security/clinic_id は characterization テスト緑が前提。#201 ガード温存。
- 型安全: `any`/`interface{}` 新規導入禁止。
- 初回カバー済(DRY/dead-code/P1-P18)を再実施しない。
- Docker 必須。scoped 検証のみ自動可。全 `go test ./...`/`golangci-lint run ./...`/`gofmt -w ./...` は禁止→手動依頼。
- テスト除去は慎重に: 回帰テスト・ビジネスロジックテストを消さない。
- 小コミットを main に積む。push しない。

## Risk Tier
- Local write(backend ローカル編集 + main ローカルコミット・reversible)。file-split は機械的・低リスク。test 変更は test のみ。security/clinic_id 接触時のみ高リスク(characterization 前提)。
- safety boundary: push/PR/マージ/外部公開は明示承認。コミットで停止。

## Truth Source Priority
1. 実行可能チェック: scoped go test/vet/build/lint・型・カバレッジ
2. 現行コード挙動(リファクタの不変対象)
3. 有効 ADR・P1-P18・coding-style(200-400行 typical/800 max)
4. CLAUDE.md/各層 CLAUDE.md
5. コメント・古いメモ・過去報告(コードで再検証、額面信用禁止)

## Verification Strategy
最狭の決定的検証面。file-split: 分割前後で package テスト緑 + build 緑。test-quality: 除去/変更後に package テスト緑 + カバレッジ不低下。error/concurrency: behavior-preserving 抽出は package テスト緑、behavior-changing は報告のみ(検証=diff 非包含)。security/clinic_id: characterization 緑 + #201 ガード grep。全スイート最終確認は禁止コマンドのため手動コマンド明記。

## Harness Improvement Feedback
P0=破壊的/認証/外部書込み防止、P1=反復する正確性/回帰防止(例: >800行・framework-only テスト・swallow を機械検出する filelen/errcheck/go-ast lint)、P2=保守性。

## Harness Contract
本プロンプトは以下を保持する場合のみ有効:
- 各 Success Criteria に expected behavior / target surface / verification method / PASS evidence を備えた Acceptance Checklist
- harness type を1つ明記する Harness Selection(Chosen/Why)
- outer loop を1つ明記する Execution Loop Selection(Chosen/Why/Stop condition)、reconciliation loop はその内側
- checklist 単位 PASS/FAIL/BLOCKED の Spec-Implementation Reconciliation Loop
- 失敗リトライごとの Failure Signature ログ
- Loop Control: 同一signature 2回で最小単位へ縮小、3回で BLOCKED
- コード/テスト変更後の De-Sloppify Pass
- 高リスク/多ファイル/security 接触変更への Independent Review Gate
- task-to-agent routing と main-agent evidence integration を伴う Subagent Orchestration
- 長時間ループで loop-status が利用可能なら loop-status 証跡
- いずれかの checklist が FAIL の間は最終報告禁止
- Saved Prompt Validation Gate / Eval Regression Capture
保存検証: `/prompt-craft-validate <prompt-file>` または `node ~/.claude/scripts/prompt-craft-harness-validate.js <prompt-file>`。enforcement は PreToolUse の `prompt-craft-harness-enforce-hook.js` が保存時に強制する。When changing the harness, run `node ~/.claude/scripts/prompt-craft-harness-enforce-hook.js --self-test` to confirm the enforcement hook still passes.

## Saved Prompt Validation Gate
本プロンプトは Local write × 多ファイル。Save the prompt to a file and validate it with `/prompt-craft-validate <prompt-file>` or `node ~/.claude/scripts/prompt-craft-harness-validate.js <prompt-file>` before execution; chat-only output is not validated. 権限拒否時は enforce hook の保存許可(PreToolUse pass)を同値 evidence として記録する。

## Eval Regression Capture
本プロンプトが no-op を誤改変・behavior-changing を refactor に混入・回帰テストを巻き込んだ場合、`node ~/.claude/scripts/prompt-craft-eval-add.js <prompt-file> --invalid --name be-refactor-second-lens` で eval corpus へ追加し、`node ~/.claude/scripts/prompt-craft-harness-validate.js --eval-dir ~/.claude/evals/prompt-craft-harness` を再実行。

## Subagent Orchestration
Routing: use `planner` for planning and the Phase 0 backlog; `code-reviewer` and `go-reviewer` for code review; `security-reviewer` plus `healthcare-reviewer` for security/clinic_id surfaces; `loop-operator` / `loop-status` for long-running loop monitoring; `harness-optimizer` for harness review. The main agent must treat agent output as evidence, not authority, and reconcile every claim against the Acceptance Checklist; the main agent owns final integration. If a needed agent is missing or none were used, state the fallback (nearest command/skill or a manual fresh pass) and state why.

具体 routing(読み取り専用・並列): Phase 0 mapping=code-explorer/Explore(very thorough)。error-handling/swallow=silent-failure-hunter。テスト品質=pr-test-analyzer/test-strategist。並行安全性=go-reviewer。characterization=tdd-guide。ビルド=go-build-resolver。並列化は読み取り専用調査・レビューのみ。編集・コミットは並列化しない。main agent が「debt/race/低価値テスト」主張を自前 grep/codegraph/テストで裏取りしてから採否(教訓: 107 tables/naive 12 偽陽性/gorm:type:date 誤読/「3件 bug」過大評価)。各 agent の寄与・採否理由・未使用理由を Deliverables に記録。

## Claude Code Harness Inspection
着手前に確認: ルート/各層 CLAUDE.md、refs(go-language.md/error-handling.md/testing.md/code-style.md)、.claude/settings(.local).json、.claude フック(PostToolUse/Stop/PreToolUse enforce・formatter コメント分断罠)、利用可能 command/subagent/skill、permission 境界。並行セッションによる main への push を git status/reflog で監視。

## Session State
4レンズ・多ファイルの長期タスク。簡潔な進捗を維持: Phase 0 バックログ fix/report/clean 数 / 修正済み単位(lens 別)と before/after 緑 / behavior-changing 報告候補数 / characterization 状況 / 並行 push の remote 到達有無 / 次アクション。clean が多数なら「修正不要」が正しい完了。

## Loop Control and Failure Recovery
reconciliation は有界ループ。失敗単位ごとに Failure Signature を記録。If the same failure signature appears twice, reduce scope to the smallest failing unit (single file / single test) and inspect the code path. If the same failure signature appears three times, or verification is impossible due to missing external state, mark the item BLOCKED and state the required input. behavior-preserving が崩れたら即 revert(小コミット粒度)。safety boundary(push/PR/マージ/破壊的/認証/外部公開)で即停止。ECC loop 監視が使えるなら `ecc loop-status --json`(or `node scripts/loop-status.js --json --write-dir .ecc/loop-status`)を証跡に。無ければ「監視不可」と明記し手動 Failure Signature ログで継続。

### Failure Signature Log Format
Failure Signature: checklist item / expected behavior / actual behavior / verification command/check / error signature / attempt number / attempted fix / result.

## De-Sloppify Pass
test-quality レンズが本タスクの de-sloppify。framework/mock 挙動のみ検証するテスト除去、暗黙ゼロ値堅牢化、過剰防御整理(信頼境界は残す)、デバッグログ・コメントアウト・無関係 drive-by 除去。ビジネスロジック・旧 failure mode 回帰・境界検証テストは残す。実挙動カバレッジを落とさない。cleanup 後に影響範囲の最狭 scoped テスト再実行。

## Independent Review Gate
多ファイル(file-split)・テスト除去・security 接触のため reconciliation 後・最終報告前に fresh review: `/code-review` または code-reviewer/go-reviewer。security/clinic_id 接触ユニットは security-reviewer + healthcare-reviewer 追加。観点: 挙動変更混入・error 契約/API 差分・回帰テスト誤除去・カバレッジ退行・#201 退行・no-op 誤改変・behavior-changing 混入。CRITICAL/HIGH 修正後 reconciliation 再実行。

## Harness Selection
- Chosen: tdd
- Why: behavior-preserving リファクタは「変更前に緑なテストを変更後も緑に保つ」ことが品質の本質。file-split/test-cleanup を package テスト緑(+ build + カバレッジ非退行)で駆動できる。security ユニットは characterization を安全網に。backing skill は tdd-workflow + verification-loop、振る舞い保存 workflow は orch-refine-code。未提供環境では「テスト緑 before/after の手動構造化検証」へフォールバックし明記。

## Execution Loop Selection
- Chosen: sequential
- Why: Phase 0 バックログを1単位ずつ(audit→characterization 要時→restructure→scoped 緑→commit)main へ serial に積む。並列ブランチ/マージキュー(rfc-dag)も PR ゲート(continuous-pr)も不要。コード/テスト変更を伴うため de-sloppify をオーバーレイ。
- The reconciliation loop runs inside the selected loop and is the only authority for checklist verdicts.
- Stop condition: 全 Acceptance Checklist 項目が PASS、または genuine blocker 記録。clean のみなら no-op 完了。

## Execution Flow
1. Risk Tier と safety boundary を分類(push 境界で停止)。
2. ルート/各層 CLAUDE.md・refs・既存挙動・利用可能ハーネスを調査。初回 refactor の結論(well-maintained)を前提に重複回避。
3. 各 Success Criteria を決定的検証手段つき Acceptance Checklist へ展開。
4. Subagent Orchestration: 読み取り専用調査 agent を早期並列起動。主張は main agent が裏取り。
5. Harness Selection(tdd)。6. Execution Loop Selection(sequential)。
7. Phase 0 の4レンズ優先順位付きバックログを出力し続行。
8. 各 fix 単位: security/clinic_id なら characterization 先行→behavior-preserving に restructure→scoped 緑→小コミット。behavior-changing は report。
9. De-Sloppify Pass→影響範囲の最狭 scoped テスト再実行。
10. Spec-Implementation Reconciliation Loop を Loop Control 付きで実行。
11. Independent Review Gate(security ユニットは dual)。CRITICAL/HIGH 修正後 reconciliation 再実行。
12. 全 Acceptance Checklist が PASS or genuine blocker 記録時のみ完了。

## Spec-Implementation Reconciliation Loop
各 checklist 項目: 期待挙動→実装挙動→使用 evidence(scoped テスト/build/カバレッジ/grep/git)→PASS/FAIL/BLOCKED。決定的検証が存在し未実行なら PASS にしない。挙動変更/契約変更/カバレッジ退行/回帰テスト誤除去の疑いがあれば PASS にしない。FAIL: 当該単位のみ patch(または revert)→最狭 scoped 検証を再実行→全 checklist 再確認→Failure Signature 更新。BLOCKED: 正確なブロッカー・必要入力を明示。Do not produce the final answer while any item remains FAIL.

## Execution Rules
- 本プロンプトを全 in-scope 作業の end-to-end 実行 authorization として扱う。
- 実装開始後は mid-task の確認/承認を求めない。質問は safety boundary のみ。
- safety boundary: 破壊的/認証情報/外部 posting・pushing・merging/課金/不可逆な第三者変更(本タスクでは push が該当)。
- 劣化を仮定して no-op を誤改変しない・behavior を変える「改善」をしない。behavior-changing は報告のみ。public API・error 契約を保つ。
- 機密/API キー/トークンを直接貼らない。

## Context
- リポジトリ: AnimalEkarte(動物病院 電子カルテ)。BE: Go 1.25/Gin/GORM/PostgreSQL 18。handler→service→repository、clinic_id マルチテナント、P1-P18。
- 重要前提: 初回の系統的 BE リファクタ(DRY/dead-code/P1-P18)は実行済みで「well-maintained・sweeping 不要」と結論。本タスクは別レンズ(error-handling/file-size/concurrency/test-quality)で初回カバー分は再実施しない。
- 教訓: agent/過去報告の「debt」主張は額面で信じない。verb/コードで裏取りしてから採用。no-op は正しい結論たり得る。
- 規約: coding-style(200-400行 typical/800 max)・error-handling refs・testing refs。#201 ガード温存。
- 並行: 複数セッションが同一 main にコミット・bundled push 実績(git status/reflog 監視)。
- 納品: main に小コミットで直接積む(push なし)。

## Deliverables
- Phase 0 の4レンズ優先順位付きバックログ(候補・lens・fix/report/clean 分類・file:line)
- lens 別 fix/report/clean サマリー
- 変更ファイル一覧・before→after・コミットハッシュ(clean は不変=正しい根拠)
- behavior-changing 別バグ報告リスト(file:line・なぜ・影響)
- Acceptance checklist PASS/FAIL/BLOCKED + evidence
- Saved Prompt Validation Gate 結果(validator exit 0 / enforce hook pass 同値 / chat-only output is not validated)
- Eval Regression Capture 結果、または "not needed"
- Failure Signature ログ、または "none"
- Risk Tier・safety boundary 評価(並行 push 監視結果含む)
- Subagent Orchestration 結果(使用agent/evidence/採否理由)
- Harness Selection(tdd) / Execution Loop Selection(sequential・stop condition)
- De-sloppify pass 結果 / Independent review 結果
- 検証結果(scoped go test before/after, build, wc -l, カバレッジ, grep, git log)。全スイート手動コマンド明記
