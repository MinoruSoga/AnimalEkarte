# Docs perfection unit ledger

文書保守ユニット専用。製品タスクの状態・担当・Doneを代替しない（[ROLE-MAP](ROLE-MAP.md)）。先行ユニットの状態は各節の実行時点の記録。最新の追加依頼は末尾の `DOCS-REFRESH-ISSUE-RECONCILIATION` と [GOAL.yaml](GOAL.yaml) の `issue_refresh` を参照する。

## DOCS-PERFECT-COORDINATOR

- status: COMPLETE（DOCS-PERFECT-COORDINATORのみ。親はactive）
- baseline HEAD: `267a17e48ed8f0cc1944d25ad11900cdbd433226`
- claim: `claim/DOCS-PERFECT-COORDINATOR`（解放は統合/明示放棄後にUSERのみ。エージェントは削除しない）
- original docs: Markdown 187件、非Markdown 6件。非Markdownは本文inventory対象外、baselineには含む。
- existing WIP: `readiness-report.md`、`.ecc/`、`.prime/`、`backend/migrations.zip`。本ユニットが作成した差分ではない。
- changed files: `docs/work/docs-perfection/**`（成果物・証拠）、`docs/work/README.md`（保守role map入口1行）、`docs/ops/deploy/README.md`（Mac配布前提入口1行）。
- source prompt / validation: [PROGRESS](PROGRESS.md) / [delivery-validation.json](EVIDENCE/delivery-validation.json)。harness/receiver両方の結果とpromptSha256を保存。
- Deviation: AC7のglobal dirty-tree判定を開始時baselineとの差分判定にする。他者WIPを戻すことは禁止。AC4の `rg -L` は指定の意図（各ファイル両キー必須）を満たす個別schema検査に置換する。



## DOCS-PERFECT-PRODUCT-PHILOSOPHY

- status: COMPLETE
- integrated_at: 2026-09-07
- claim: `claim/DOCS-PERFECT-PRODUCT-PHILOSOPHY`（解放はUSERのみ）
- RQ-010: resolved
- Changed files: `docs/product-philosophy.md:41` (1 insertion, 1 deletion)
- docs/README.md: unchanged; Markdown links=11, missing=0; category set unchanged
- bash scripts/check-docs-symbol-drift.sh: exit 0 — OK docs-symbol-drift（検査トークン 565 件）
- git diff --check -- docs/product-philosophy.md docs/README.md: exit 0
- Independent reviews: product_santa_b PASS; product_santa_c PASS（実行側報告）。生成側 Mode 3 再照合でも文言整合を確認。
- Unresolved CRITICAL/HIGH: 0
- Parent goal: active
- Receipt source: Codex child Completion Report reconciled 2026-09-07

## 子ユニットDAG

PRODUCT-PHILOSOPHY 完了。残りは architecture → spec → ops → delivery → work。親が直列統合する。優先修復順はRQのScore、起動順は正本/参照先の依存を優先する。RQ-001/002は高優先度だが、今回の本文修復開始は認可されていない。

| ID | Definition | State | Depends on | 受入の焦点 |
|---|---|---|---|---|
| DOCS-PERFECT-PRODUCT-PHILOSOPHY | [product-philosophy.md](CHILD-GOALS/product-philosophy.md) | COMPLETE | DOCS-PERFECT-COORDINATOR | 原則と例、臨床安全の優先、索引の正本参照 |
| DOCS-PERFECT-ARCHITECTURE | [architecture.md](CHILD-GOALS/architecture.md) | COMPLETE | DOCS-PERFECT-PRODUCT-PHILOSOPHY | 所有表/機械検査/ADR履歴の分担 |
| DOCS-PERFECT-SPEC | [spec.md](CHILD-GOALS/spec.md) | COMPLETE | DOCS-PERFECT-ARCHITECTURE | 配信安全と劣化経路の説明・画面案内 |
| DOCS-PERFECT-OPS | [ops.md](CHILD-GOALS/ops.md) | COMPLETE | DOCS-PERFECT-SPEC | 証跡不在・保存先・配布ゲートの案内 |
| DOCS-PERFECT-DELIVERY | [delivery.md](CHILD-GOALS/delivery.md) | COMPLETE | DOCS-PERFECT-OPS | manual外部依存とU1–U13の正本境界 |
| DOCS-PERFECT-WORK | [work.md](CHILD-GOALS/work.md) | COMPLETE | DOCS-PERFECT-DELIVERY | 時点メモ・Linear SoT・削除zone維持 |

RQ-002のfrontend manual修復は、このDAGのdocs-only allowlistに含めない。別承認unitの依存として残す。子の実行結果は最終応答で受け取り、親coordinatorが直列に台帳を更新する。子同士がGOAL/LEDGERを共有編集しない。

## Orchestration record

| Agent ID / label | 責務・writer-owned paths | 状態 / 証拠 | 統合判断 |
|---|---|---|---|
| /root/inventory_architecture | architecture読者 / none | completed、22 Markdown | 型owner案内をRQ-011へ。日付付きADRの旧SHAを誤りとする提案は棄却（ADR-006:194,202） |
| /root/inventory_spec | spec読者 / none | completed、80 Markdown | RQ-001/003/007を採用。基準SHAの古さと明示済みruntime pendingを新規誤りとする提案は棄却 |
| /root/inventory_ops | ops読者 / none | completed、73 Markdown | RQ-004/005/006を採用。helper不存在は現物があるため棄却し、表記の曖昧さのみRQ-008へ。provider状態は未判定 |
| /root/inventory_delivery | delivery読者 / none | completed、4 Markdown | manual内の矛盾をRQ-002へ。docs外修復の自動実行は不採用。U13/本番前提のpendingを維持 |
| /root/inventory_work | work読者 / none | completed、既存6 Markdown | RQ-009を低優先の時点表現改善として採用。既に章分けされたmappingの再構造化/READMEへのADR訂正は棄却 |
| /root | product/index読者、成果物全ての唯一のwriter | RQ-010、187件の全件リンク/構造/Git日付検査 | [INVENTORY-SCAN.json](EVIDENCE/INVENTORY-SCAN.json)、[REPAIR-QUEUE](REPAIR-QUEUE.md)の両側根拠を直接再確認 |
| /root/santa_b, /root/santa_c | 独立レビューround 1 / none | 両者completed、PASS、CRITICAL/HIGH=0 | 引用行とHistory分類のLOW2件を修正。receiptはEVIDENCE/review-*-round1.json |
| /root/santa_b2, /root/santa_c2 | fresh独立レビューround 2 / none | 両者completed、PASS、CRITICAL/HIGH=0 | 生JSONは[review-B](EVIDENCE/review-B.json) / [review-C](EVIDENCE/review-C.json)。WIP比較の表現限定を採用。allowlistの自動交差検査追加は任意改善として保留（今回の全6件は両者が手動検証済み） |

保存promptの独立並列読み取り指示に従った。同じtreeを編集するsubagentは0。実セッションに多エージェント機能あり。MCPはprovider/Linearの現在状態を本ユニットで検証しないため利用せず、ローカルCLIの直接証拠を使用。

## 検証ログ

### 保存prompt

```text
node /Users/minoru/.claude/scripts/prompt-craft-delivery-validate.js --target agent /Users/minoru/.claude/prompt-craft-runs/agent-docs-perfection-coordinator-codex-astra-v2.md
```

exit=0。生JSONは [delivery-validation.json](EVIDENCE/delivery-validation.json)。6スキルのpath/hash一致を確認し、2スキルはpath-fallback。他4つはnative、goalはinline。選択はconstruction（役割・キュー・行動範囲を作成）/ rfc-dag（子の依存を固定）。停止条件は全ACの直接検証と独立レビュー、または真正なBLOCKED/予算/安全境界。

### docs-symbol-drift

```text
$ bash scripts/check-docs-symbol-drift.sh
OK    docs-symbol-drift: ドリフトなし（検査トークン 565 件）
exit=0
```

生出力は [docs-symbol-drift.txt](EVIDENCE/docs-symbol-drift.txt)。本文narrative、ADRの採択理由、全リンク、STG/PROD・UATを証明するゲートではない。

### Runtime checks

Build / Types / full Lint / Tests / coverage / DB / browser / deploy = **N/A**。ドキュメント保守のためruntime検証は不要。本ユニットは機械ゲートの静的実行だけを許可し、禁止のフルスイートは実行していない。

### Completion audit

機械照合の出力は [checks-draft.txt](EVIDENCE/checks-draft.txt)、最終完了検査は [checks-final.txt](EVIDENCE/checks-final.txt)。独立レビューのJSONはレビュー時点の行番号と最終metadata確定前の判定をそのまま保存した。下表の最終状態は、両者のPASS取得後にcheckpointを記録し、`--final`で検証する。

| Checklist item | Expected behavior | Actual behavior | Status | Verification method | Evidence |
|---|---|---|---|---|---|
| AC1 | 4役割とcanonical owner | 4見出し・owner表・リンク/ADR方針あり | PASS | ROLE-MAPのrg・リンク実在・手動owner照合 | checks-draft.txt: four role headings / coordinator local link targets exist |
| AC2 | 全Markdownに分類 | 198=198、187既存+11新規、差分は全行skip理由あり | PASS | find比較・集合一致・enum/更新/理由検査 | `PASS inventory=198 find=198 missing=0 extra=0 duplicate=0` |
| AC3 | Severity×Impact順、根拠と処置 | 11RQ、両側根拠、allowlist外依存を区別 | PASS | verify.pyのscore/action/引用検査、根拠本文の再読 | `PASS repair queue=11 Severity*Impact descending` |
| AC4 | 6子、各done_condition/verification | 6schema PASS、DAG非循環、全pending | PASS | 各frontmatterをキー別検査 | `PASS six required categories` / checks-draft.txtのchild schema全6件 |
| AC5 | 親active、coordinator完了checkpoint | active、coordinator-completeを記録 | PASS | GOAL.yamlとPROGRESS照合、verify.py --final | `PASS parent status active` / `PASS coordinator-complete checkpoint` |
| AC6 | driftのexitとverbatim出力 | 565トークン検査、exit=0を保存 | PASS | bash scripts/check-docs-symbol-drift.sh | 本書のdocs-symbol-drift節 / EVIDENCE/docs-symbol-drift.txt |
| AC7 | このunitの差分がallowlist内 | docs外tracked内容とindex不変、既存untrackedはstatus維持。README2件とcoordinatorのみ | PASS | status/tracked/staged/untracked/baseline hash | `PASS outside docs tracked content unchanged from baseline` / `PASS tracked/untracked status delta within allowlist`。既存untracked内容hashの比較は未実施 |
| AC8 | カテゴリ本文修復なし | ops/deploy READMEに案内1行のみ | PASS | git diff --name-only -- docs/architecture docs/spec docs/ops docs/delivery | `docs/ops/deploy/README.md`。役割表から配布前提ownerへ辿るため |
| AC9 | 独立2レビュー | fresh B2/C2ともPASS、CRITICAL/HIGH=0 | PASS | 同一元仕様・同一rubric・相互評価を見ないレビューとverify.py --final | EVIDENCE/review-B.json / review-C.json の `"verdict": "PASS"`, `"critical_issues": []` |

## Closeout / Harness Improvement Feedback

- coordinator-complete。親のstatusはactive、全6子はpending。本文修復・UAT/本番検証・commit/push/PR/mergeは実行していない。
- Budget / loop health: 19/24 actions、1 session、2 review rounds。検証補助の列数誤りと最終ログ記録用コマンドの引用符誤りを最小修正して解決。2つのLOW指摘も修正し、fresh両レビューPASS。繰返し同署名はなし。詳細はPROGRESSのFailure Signature log。
- P1 prompt feedback: global `git status` に既存WIPがあれば「戻す」というAC7表現は、baseline-relativeなowned diff判定と他者WIP保持を明記する方がよい。今回この境界は上位project safetyに従って適用した。
- P2 prompt feedback: AC4の `rg -L "done_condition:|verification:"` はripgrepでは必須両キーの欠落検査にならない。各ファイルで両キーが非空であることをparseして検証すること。今回のverify.pyはその方法を採用。
- P2 harness pointer: AGENTSが参照する `docs/CODEX-NAVIGATION-GUIDE.md` は現treeに存在しない。現行のAGENTS/.claude規約と既存カテゴリ索引を読み、無関係な新ファイルは追加しなかった。
- Eval Regression Capture: 上記3件を生成側へのfeedbackとして残す。prompt/validator/skills自身は変更していない。
- Remaining: RQ-005の索引修正以外の修復キュー、子の全命題照合、外部未確認の証拠、RQ-002のdocs外manual別承認依存。現在のruntime健康・release-readyは本結果から導けない。
- Next-session handoff (historical coordinator closeout): product-philosophy は完了。継続は architecture 以降。
- Claim retained: `claim/DOCS-PERFECT-COORDINATOR`。統合/明示放棄後の解放はUSERのみ。


## Parent continuous loop (2026-09-07)

> 2026-09-07 訂正: この節のarchitecture以降のCOMPLETE・独立レビュー記述は生成側の先行下書き記録。Astraによる本ループの受入証拠としては未採用。旧記録を残し、現在の採否・根拠は末尾の `DOCS-PERFECT-ASTRA-LOOP` で確定する。coordinatorとproduct-philosophyの既存実行証拠とは区別する。

mode: continuous-agent-loop（親セッションが DAG を直列実行・台帳統合）

### DOCS-PERFECT-ARCHITECTURE — COMPLETE
- claim: `claim/DOCS-PERFECT-ARCHITECTURE`
- RQ-011 resolved: `docs/architecture/overview.md` Decision ownership に model-write-owner-catalog への行を追加
- drift: exit 0（親ループ終端で再実行）

### DOCS-PERFECT-SPEC — COMPLETE
- claim: `claim/DOCS-PERFECT-SPEC`
- RQ-001 resolved: `docs/spec/screens/34-lstep-delivery-monitor.md` 死亡除外の絶対表現を撤回し lstep-integration §4 へリンク
- RQ-003 resolved: 同ファイルの無条件 no-N+1 主張を撤回し lstep-integration §5.2 へリンク
- RQ-007 resolved: screens/README に ADR-007 案内、settings/README に検査機器正本リンク

### DOCS-PERFECT-OPS — COMPLETE
- claim: `claim/DOCS-PERFECT-OPS`
- RQ-004 resolved: UAT-DOMAIN-STATUS の欠落 reports リンク9件を非配布証跡識別子表記へ
- RQ-005: 既存 index-resolved を維持
- RQ-006 resolved: STG-CONTINUOUS-OPERATIONS の一時ログと正式証跡を分離
- RQ-008 resolved: UAT-ENV-SETUP の check-uat-env パスを文書相対で明示

### DOCS-PERFECT-DELIVERY — COMPLETE（docs allowlist）
- claim: `claim/DOCS-PERFECT-DELIVERY`
- RQ-002: frontend manual 本体は未修正のまま **external-scope**。OPERATION_MANUAL に依存明示を追加し「整合完了」主張を禁止

### DOCS-PERFECT-WORK — COMPLETE
- claim: `claim/DOCS-PERFECT-WORK`
- RQ-009 resolved: fable-po-recommendation の Needs Human を照合日時点＋外部未確認に明示結線

### Independent review (parent loop)
- Pass A: RQ 原文と差分の突合（過剰断定の除去・正本リンク・allowlist）
- Pass B: リンク実在・drift・「frontend manual を直したかのように読める表現がないこと」
- CRITICAL/HIGH: 0（RQ-002 は意図的に docs 外未修復）

### Parent goal status
- 修復キューの docs-owned 項目は resolved / external-noted
- 親 `done_condition` の「外部証跡待ちを完了扱いしない」を満たすため、RQ-002 の frontend 修復と Linear 現在状態の外部確認は **未完了の明示残件**
- したがって親 GOAL は `active` のまま（completed にしない）

## DOCS-PERFECT-ASTRA-LOOP

- status: COMPLETE（本ループのA/B/Cと独立レビュー。親GOALはactive）
- source prompt: `/Users/minoru/.claude/prompt-craft-runs/agent-docs-perfection-astra-investigate-then-repair-loop.md`
- validation: `node /Users/minoru/.claude/scripts/prompt-craft-delivery-validate.js --target agent /Users/minoru/.claude/prompt-craft-runs/agent-docs-perfection-astra-investigate-then-repair-loop.md`; exit=0; ok/harness.ok/receiver.ok=true
- promptSha256: `aa7383f95b6770c08e97c6a14cb82ab34b69d0346852e24340095c72694a8ed1`
- risk: Local write; no commit/push/PR/app/manual/migration; parent remains active
- claim: `claim/DOCS-PERFECT-ASTRA-LOOP` created exit=0。既存7 claimは今回の保存promptの明示指示に従い再利用し、削除しない。子契約の既存claim停止・次子禁止・親統合別セッションの規則は、今回明示された親の直列ループ認可の範囲でのみ上書きする。allowlist/安全/検証条件は維持。
- baseline: [astra-loop-baseline.json](EVIDENCE/astra-loop-baseline.json)、[先行docs差分](EVIDENCE/astra-loop-draft.patch)。docs外のbackend等のWIPは対象外。
- harness: construction / continuous-pr（品質ゲートの反復のみ、PR作成なし）。living-docs-governance・continuous-agent-loop・santa-methodはnative読込、verification-loopはpath-fallback。4 hash一致、goalはinline。実セッションとproject configはdanger-full-access / never。

### Phase A — 調査正本の配置確認

`work/README → docs-perfection/README → ROLE-MAP / REPAIR-QUEUE / INVENTORY / CHILD-GOALS / EVIDENCE` の入口を確認。新規カテゴリ・外部SoTは作らない。

```text
Phase A entry links missing=0 []
Current docs Markdown=199
```

README追加がinventoryに未反映だったため1行を追加。coordinatorの198件は履歴として保持。カテゴリ本文の採用/修復はこの確認後に開始する。

| 読了対象 | 役割・確認内容 |
|---|---|
| README.md / docs/work/README.md | 調査入口とLinearの製品実行SoTを分離。入口リンクを存在確認 |
| ROLE-MAP.md | Constitution/Map/Status/Historyと各fact ownerの参照表 |
| INVENTORY.md | 187既存Markdownの一次調査＋保守成果物。currentは全命題の一致証明ではない |
| REPAIR-QUEUE.md | 両側根拠付き11RQ。RQ-002のdocs外依存を維持 |
| CHILD-GOALS/{product-philosophy,architecture,spec,ops,delivery,work}.md | 6件のallowlist、done_condition、verification、禁止、依存順を確認 |
| GOAL.yaml / LEDGER.md / PROGRESS.md | 親の目的・状態・履歴。先行COMPLETEは再監査で根拠を置く |
| EVIDENCE/BASELINE.json | coordinator開始時のhash/status。今回の開始状態ではない |
| EVIDENCE/INVENTORY-SCAN.json | 187既存文書の構造・ローカルリンク・Git手がかりの調査データ |
| EVIDENCE/delivery-validation.json | coordinator保存promptのharness/receiver結果 |
| EVIDENCE/checks-draft.txt / checks-final.txt | coordinator時点のコマンド原文とexit。後続修復の検証に転用しない |
| EVIDENCE/docs-symbol-drift.txt | coordinator時点の565トークン静的ゲート |
| EVIDENCE/review-{B,C}-round1.json / review-{B,C}.json | coordinator独立レビューの履歴。今回のsantaとは別 |
| EVIDENCE/verify.py | coordinator範囲・全子pendingを前提にした再生スクリプト。今回のscope/statusには適用しない |

### 実装前の受入対応

| AC | Expected | Verification |
|---|---|---|
| A1 | 調査入口確定 | READMEリンクの存在と経路検査 |
| A2 | 調査正本読了 | 上の読了表と各artifactの役割確認 |
| A3 | 各先行差分の採否 | baseline patchの全docsファイルとRQの対応表 |
| A4 | docs-owned RQ resolved/BLOCKED、RQ-002 external | RQと本文・指定code/spec参照を再照合、子契約の検証 |
| A5 | drift実行 | bash scripts/check-docs-symbol-drift.sh の原文出力・exit |
| A6 | whitespace健全 | git diff --check -- docs |
| A7 | 独立2レビュー | 同一入力・rubric、相互評価を見ない2 agentでCRITICAL/HIGH 0 |
| A8 | 親偽完了なし | rg -n '^status:' docs/work/docs-perfection/GOAL.yaml がactive |

### Phase B — 先行本文差分の採否（Phase A確認後）

全12ファイルを開始時 `git diff -- docs` と [保存patch](EVIDENCE/astra-loop-draft.patch) で照合。adoptは既存下書きをそのまま受け入れ、amendは下記の限定修正後に受け入れる。破棄する下書きは0。既存のRQ resolved表記だけを証拠にはせず、Phase Cと独立レビューで再判定する。

| 先行dirty docsファイル | RQ / purpose | 判定 | 根拠・必要な限定修正 |
|---|---|---|---|
| docs/architecture/overview.md | RQ-011 | adopt | :112 の型別catalog案内。model-write-owner-catalog.md:3,14とADR-006の所有原則を分離 |
| docs/product-philosophy.md | RQ-010 | adopt | :41は:162,172とCLAUDE:19,22 / specification:22に整合。既存Astra修復を保持 |
| docs/spec/screens/34-lstep-delivery-monitor.md | RQ-001/003 | amend | :48の死亡配信保証限定は採用。:60は通常経路bulk-readの「必須」を残し、実装fallbackの既知gapを別記する |
| docs/spec/screens/README.md | RQ-007 | adopt | :24は実在の検査受信routeをADR-007へ案内。ADR-007冒頭のADR-008へのsupersedes参照も確認 |
| docs/spec/screens/settings/README.md | RQ-007 | amend | :28の機器受信案内を項目マッピング画面と混同しない。settings-routes.tsx:462のitem-mastersとADR-007 §7を直接案内する |
| docs/ops/deploy/README.md | RQ-005 | adopt | :47にMac配布前提ownerの入口。署名/token/STG配布完了を主張しない |
| docs/ops/deploy/STG-CONTINUOUS-OPERATIONS.md | RQ-006 | amend | :235–236の一時採取/正式証跡分離は採用。週次・月次も正式保存対象と明示し、生成指示由来のwiki注釈を除く |
| docs/ops/testing/UAT-DOMAIN-STATUS.md | RQ-004 | amend | 9箇所の欠落リンクを識別子にする方針を採用。未確認の保管者の所持を断定せず、同じ注意を9回複製しない |
| docs/ops/testing/UAT-ENV-SETUP.md | RQ-008 | adopt | :15で実在helperの文書相対/ルート相対を区別しadvisoryを保持 |
| docs/delivery/OPERATION_MANUAL.md | RQ-002 | amend | :34の未修復依存明示は採用。読者向けに会計作成の誤読箇所を示し、保守allowlist説明をRQへ委譲。既存D-254表に依存を結線 |
| docs/work/README.md | 調査入口 | adopt | :21–23の3リンク実在。Linearと削除zoneを維持 |
| docs/work/decisions/fable-po-recommendation.md | RQ-009 | adopt | :11を:5の照合日と:6の未照会に結線。現在のLinearを推定しない |

Phase Cはarchitecture → spec → ops → delivery → workの5ユニット。変更前に読取laneの根拠を統合し、書込は主エージェントのみ。

### Phase C.1 — architecture

- RQ-011: adopt。overview:112から型別catalogへ到達。ADRの採択理由/歴史SHAは変更なし。
- checks: [architecture-checks](EVIDENCE/astra-loop-architecture-checks.json)。diff-check exit0、ADR差分出力なし。
- read-only laneの直接照合: category/inventory=22/22、ERD DDL/listed=128/128、auth AllResources/listed=37/37、lintscan=35 top-level /14 domain。全件receiptと独立gateで最終判定する。
- state: static-verified / independent-review-pending。本文追加編集0。units 1/5。

### Phase C.2 — spec

- RQ-001: 死亡除外の既知gapと是正ownerを案内する下書きをadopt。コードの欠陥を修正したとはしない。
- RQ-003: 通常経路bulk-read必須を保持し、実装のper-owner fallback未是正を分けた。同じ説明を持つ `docs/spec/screens/31-lstep-integration.md:69` も是正ownerへの参照を追加。これは先行dirty12件には無かった追加1ファイルで、元のspec子allowlist内。新製品要件なし。
- RQ-007: screens索引をadopt、settings索引は `/settings/lab-device-item-masters` と20-master-settings / ADR-007 §7へ限定。受信ボードと項目設定を区別。
- checks: [spec-checks](EVIDENCE/astra-loop-spec-checks.json)。diff-check exit0、安全/fallback表現と実在routeを照合。
- state: static-verified / independent-review-pending。units 2/5。

### Phase C.3 — ops

- RQ-004: 非配布証跡9リンクは識別子として保持。注意を冒頭:27へ一本化し、内容・現在の所在は未確認と明記。保管者の所持を断定する下書きは採用しない。原証跡を取得/捏造せず、過去UATの状態も変更しない。
- RQ-005: Mac配布前提リンクはindex-resolvedのまま。配布/署名/token問題の解消ではない。
- RQ-006: 一時採取と正式保存を分離。デプロイ直後・週次・月次すべてに正式な保存先を明示。
- RQ-008: 実在helperへの文書相対リンクとadvisory境界をadopt。helper本体は実行しない。
- checks: [ops-checks](EVIDENCE/astra-loop-ops-checks.json)。diff-check/helper存在exit0。read-only laneは73/73 inventory、全local Markdownリンク欠落0、helper bash -n exit0。
- state: static-verified / independent-review-pending。units 3/5。

### Phase C.4 — delivery

- 全4文書を本文・local参照・U1–U13/受入欄の境界で確認。OPERATION_MANUALのみ限定修正。
- RQ-002: external-scope (docs-noted)を維持。workflow:111–112と会計画面manual:68 / docs99:41の矛盾を再確認。納品読者に未修復を案内し、§11 D-254に別承認依存を結線。frontend manual本体の変更0。
- DELIVERY_PACKAGEのU1–U12、OPERATION_MANUALのU13所有を維持。未記入の値・署名・provider/UAT receiptは補わない。GOLIVEのHOLD/旧window失効/研修順序未決定を維持。
- checks: [delivery-checks](EVIDENCE/astra-loop-delivery-checks.json)、diff-check exit0。units 4/5。

### Phase C.5 — work

- 全6既存Markdownを照合。RQ-009のNeeds Humanを2026-08-20時点に結び付ける下書きをadopt。Linear現在値の外部照会は実施しない。
- docs/work/READMEの削除zoneと製品SoTを保持。STATUS.md/PO-todo.mdの新設0。親調査パッケージの変更は本loopの親writerだけが担当し、work子成果と区別。
- checks: [work-checks](EVIDENCE/astra-loop-work-checks.json)、diff-check exit0。units 5/5。
- 全5カテゴリはstatic-verified / independent-review-pending。これ以上のカテゴリunitは開始しない。最終santaと全ACの証拠照合を行う。

### 全件再照合の証拠と検証範囲

- [astra-loop-scan.json](EVIDENCE/astra-loop-scan.json): 全199 Markdownのpath/role/classification/hash、ローカルリンク、文中の時点・証拠・契約表現。機械抽出と本文の意味検証を区別する。
- [astra-loop-readers.json](EVIDENCE/astra-loop-readers.json): architecture全22件の内容根拠、spec/opsのRQ直接照合と補足読了記録。
- docs-owned RQ解消に伴いInventoryの8行をcurrentへ再分類。元の一次調査値は旧INVENTORY-SCAN.json/検査ログを保持し、日付だけを理由に歴史SHAを更新しない。
- 原証跡/provider/Linearの外部内容は本loopで未取得。UAT/本番未実施を静的ゲートでPASSにしない。
- [astra-loop-ownership.json](EVIDENCE/astra-loop-ownership.json): 今回のtracked本文編集6件。作業中に観測したdocs外の `backend/internal/lintscan/dbortx_inventory_lint_test.go` と `master_fk_write_inventory_lint_test.go` の変化を別記。主エージェントによるdocs外writeは0、producer identityは未確認。全tree不変とは主張せず、他者WIPを戻さない。indexとHEADは開始時と一致。

### 要件別照合（最終）

| Checklist item | Expected behavior | Actual behavior | Status | Verification method | Evidence |
|---|---|---|---|---|---|
| A1 | 調査入口確定 | workから調査README/RQ/ROLE-MAPへ到達 | PASS | verify-astra-loop.py | `PASS work entry README.md` / `PASS work entry REPAIR-QUEUE.md` / `PASS work entry ROLE-MAP.md` |
| A2 | 調査正本読了 | 役割別読了表と既存証拠の時点を記録 | PASS | 本書Phase A表、artifact本文・schema確認 | 本書「Phase A — 調査正本の配置確認」 |
| A3 | 全先行差分の採否 | 12件adopt7/amend5/revert0 | PASS | baseline patchと現在diff、RQ照合 | 本書Phase Bの12行 |
| A4 | docs-owned RQ解消/依存明示 | 9件resolved、RQ005 index-only、RQ002 external | PASS | RQ/code/spec/子検証 | REPAIR-QUEUE.md:9–19 / EVIDENCE/astra-loop-{architecture,spec,ops,delivery,work}-checks.json |
| A5 | drift終端実行 | exit0、565トークン検査 | PASS | bash scripts/check-docs-symbol-drift.sh | `OK    docs-symbol-drift: ドリフトなし（検査トークン 565 件）` / EVIDENCE/astra-loop-drift.txt |
| A6 | whitespace健全 | exit0 | PASS | git diff --check -- docs | `PASS git diff --check -- docs` |
| A7 | 独立2レビュー | B/CともPASS、CRITICAL/HIGH0 | PASS | santa dual review | EVIDENCE/astra-loop-review-B.json:2 / astra-loop-review-C.json:2、各critical_issuesは空 |
| A8 | 親偽完了なし | active、RQ002未修復 | PASS | rg -n '^status:' GOAL.yaml | `status: active`、GOAL.yaml:5 |

### 主エージェントのdelivery/work/product再照合

| 文書 | 内容根拠・判定 |
|---|---|
| docs/delivery/README.md | :6のprovider時点、:14–16の納品正本分担、:27–36のU1–U13、:39の外部backup依存を維持。Map/current |
| docs/delivery/DELIVERY_PACKAGE.md | :6–7のrepo/外部境界、:225以降U1–U12未記入、U13はOPERATION_MANUAL。採取/構築は未実行。Map/current contract |
| docs/delivery/GOLIVE_RUNBOOK.md | :3–8のHOLDと旧window履歴、:26の研修順序未決定、:76のbackup gate。日付・named owner・receiptを埋めない。Map/current contract |
| docs/delivery/OPERATION_MANUAL.md | :34でRQ002未修復、§11 D-254へ接続、§10 U13未完。Map/current with external dependency |
| docs/work/README.md | :7,13 Linear優先、:21–23調査入口、:25以降削除zone8行維持。Map/current |
| docs/work/decisions/README.md | :3–4採択ポインタ・Linear優先、:12以降規律。Map/current |
| docs/work/decisions/fable-po-recommendation.md | :5–6時点/外部未確認と:11 Needs Humanを結線。History/current |
| docs/work/phase2-deferred.md | :3–7時点と再開owner、:13–16測定待ちの再開条件、:24以降やらない判断。Map/current contract |
| docs/work/linear-f1-f6-mapping.md | :3–6照合日とUNKNOWN、:19以降repo履歴、:51以降USER手順を分離。Status/current as dated record |
| docs/work/skill-reeval-2026-09-06.md | :3照合日、:8–14代表タスク停止条件、:18以降契約メモ。History/current |
| docs/product-philosophy.md / docs/README.md | PP:41を:162,172、CLAUDE:19,22 / specification:22へ再照合。README local .md11件、カテゴリ5+PP不変。既存修復を保持 |

### Loop health / failure recovery

- category units: 5/5、santa fix cycles: 0/2。専用loop監視は未使用、PROGRESSのcheckpointで追跡。
- A2/body-evidence-depth: 初回spec再照合がH1/Inventory中心だったため、本文契約/未実装/時点/外部依存を全80件で再読する補足laneを実施。結果はreaders.jsonのfile_receipts。attempt1で解消。全API/UI細部のsource auditとはしない。
- RQ004/keeper-possession: read-only laneのHIGH指摘を採用、内容/現在所在を未確認と明示。再確認でCRITICAL/HIGH0。独立santa前の修正。
- orchestration/capability: 新規lane追加のthread limitは既存idle lane再利用で回復。loop_specのusage-limitエラー後は部分証拠だけ採用し、inventory_specがRQコードと全80本文を再確認。未完了laneを完成証拠にしない。
- De-Sloppify: 9回の同文注意を1箇所へ統合。必須bulk要件を復元し、既存是正ownerへリンク。納品本文から保守allowlistの実装用語を除く。ADR歴史・UAT未確認・RQ002外部依存を保持。

### 後発の共有WIP観測（独立レビュー前）

最初のarchitecture照合後、別作業の変更が `adr/006-backend-domain-package-boundaries.md`、`be9-2a-boundary-map.md`、`exception-package-discipline.md` に加わった。これらはPhase B開始時12ファイルに含まれず、今回の修復としてadopt/amend/revertしない。変更せず保持する。

- 原文差分を読み取り確認: ADR-006:55に2026-09-07 amendment、boundary-map:11 / exception-package:53の36 top-level /14 domainは現在のpackage_boundary_gate_test.go:109–110と一致。機能実装や安全性の受入は当該作業のownerに属する。
- 先のarchitecture readerの35/14・ADR差分なしは観測時点の生証拠として保持する。現時点の全tree状態へ流用しない。自分のADR編集は0。
- scope gate失敗 `FAIL ADR history diff empty`（draft記録のassert失敗後、最小gateを再実行して原因特定、2観測）は、共有tree全差分を自差分扱いした検証条件が原因。3文書の差分と自分のwrite receiptを照合し、owned docs 6件と後発foreign docs 3件を区別する検証へ修正する。baselineは取り直さず保持。
- docs外WIPの追加変化もownership.jsonの時点別観測に記録。自分のwriteはdocsのみ。未分類の新差分はgateで停止し、他者WIPを戻して通さない。

### 独立レビュー・最終統合

- Changed files: 開始baselineからの自編集6本文・親package5既存ファイル・新規証拠18ファイルは [astra-loop-changed-files.json](EVIDENCE/astra-loop-changed-files.json) に全パスを列挙。子statusは先行下書きのCOMPLETEを一度IN_REVIEWに戻して再受入したため、開始時からの最終内容差分は0。
- [Review B](EVIDENCE/astra-loop-review-B.json) / [Review C](EVIDENCE/astra-loop-review-C.json): 同一rubric・他方の評価を共有しない独立2パス。両者 `"verdict": "PASS"`、`"critical_issues": []`。各自がdraft verifierを直接実行しexit0。本文修正要求なし。
- LOW提案を採用: READMEの次アクションと本unit冒頭を最終状態へ揃え、A7と6子のstatusを更新。以降はmetadata/生証拠保存のみで本文変更なし。Santa fix cycles 0/2。
- phase結果: A配置・199件整合PASS、B採否12件（adopt7/amend5/revert0）、C順序どおり5/5完了。先行PP修復を維持。親GOALはRQ-002 external-unfixedにつきactive。
- 最終コマンドの原文とexitは [astra-loop-checks-final.txt](EVIDENCE/astra-loop-checks-final.txt)、終端driftは [astra-loop-drift.txt](EVIDENCE/astra-loop-drift.txt) に保存。docs-onlyのためruntimeテスト不要。全UI/APIの実装証明・UAT・release証明は含まない。

| Agent ID / label | 責務・証拠 | writer-owned paths | 終了・統合 |
|---|---|---|---|
| /root/loop_architecture | architecture22本文・DDL/RBAC/package参照、readers.json | none | completed。分類と当時35/14の直接観測を採用。後発36/14とは時点を分離 |
| /root/loop_spec | RQ001/003/007初回再確認 | none | usage-limitでerrored。部分指摘のみ採用、完成証拠には使わず下記で補完 |
| /root/inventory_spec | spec80本文再読・RQコード再照合、readers.json | none | completed。H1-only判定を本文receiptで置換。全API/UI達成認定には不採用 |
| /root/inventory_ops | ops73本文・参照/コマンド静的照合、readers.json | none | completed。keeper-possession HIGHを修正、再確認0件を採用 |
| /root/product_santa_b | 今回の独立受入レビューB、review-B.json | none | completed/PASS。LOWの最終metadata整合を採用 |
| /root/product_santa_c | 今回の独立受入レビューC、review-C.json | none | completed/PASS。LOWの最終metadata整合を採用 |

新規thread上限により既存idle agentを再利用したが、今回の本文writerとは分離して実際に並列読取を実施。主エージェントだけが書込・最終統合。loop_spec以外の起動済み役割は完了し、実行中の子は0。旧coordinatorのinventory_architectureは旧証拠で、今回の独立santaに数えない。専用loop監視facilityは使用せずPROGRESSで上限/各反復gateを追跡した。

### 最終失敗ログ・harness改善

| Checklist / signature | 期待 / 実際・検査 | attempt / 修正・結果 |
|---|---|---|
| A1/inventory-entry-missing | 全件列挙 / 集合比較199対198 | 1: package README行追加、missing/extra/duplicate0 |
| A1/final-log-creation-order | 最終ログへのリンク実在 / 初回final gate時にログ保存前で2リンク欠落 | 1: 失敗出力を同ログへ保存後、追記方式で再実行。ログ内容やgateを緩めず実在順序を是正 |
| A2/body-evidence-depth | 本文に基づく判定 / 初回H1中心のreceiptを読取審査 | 1: spec80本文の補足読了で解消。全UI/API証明とはしない |
| A4/RQ004-keeper-possession | 原証跡の未知を保持 / ops reader HIGH | 1: 所持断定を除去、注意1箇所へ統合、再確認CRITICAL/HIGH0 |
| A6/scope-ADR-global-diff-empty | 自ADR編集0 / 共有diffを0と要求しassert/最小gateで2失敗 | 2: 最小3文書を直接確認しowned/foreign分類へ限定、再検証PASS。baselineは保持 |
| orchestration/capability | 並列読取 / thread-limitとloop_spec usage-limit | 既存idle laneで必要証拠を補完。未完了laneをPASSにしない |

- P1生成側feedback: 先行本文修復とCOMPLETE/独立レビュー記述は下書きとして区別し、実行側の直接受入前に完了扱いしない。
- P1契約整合feedback: 新promptの既存claim再利用・親直列統合と子契約の停止条件が競合する。今回は明示された新指示を限定適用し記録。生成側で次版の子契約にも整合を付ける。
- P2能力feedback: thread/usage上限時のidle lane再利用を実行契約に明記するとよい。独立性と全lane joinを維持する。ハーネス本体の変更はしない。
- Safety boundary events: none。Local write内のみ。commit/push/PR/merge/アプリ/マイグレーション/外部状態変更/claim削除なし。
- retained claims: `claim/DOCS-PERFECT-{COORDINATOR,PRODUCT-PHILOSOPHY,ARCHITECTURE,SPEC,OPS,DELIVERY,WORK,ASTRA-LOOP}`。統合または明示的放棄後の削除はUSERのみ。


## RQ-002 resolution (2026-09-07)

- status: COMPLETE（frontend manual 本文）
- evidence: handleFinalize は `status: finalized` のみ（`use-medical-record-quick-patch-actions.ts`）。billing Create なし。会計は会計画面で作成し未請求を pull（`05-accounting.md` / `99-medical-record-flow.md:41`）。退院は `discharge-with-billing` の `create_accounting` 選択時のみ会計作成。
- changed manuals: `01-new-owner-first-visit`, `02-existing-owner-visit`, `15-multiple-pets-visit`, `16-service-only-visit`, `26-efficiency-tips`, `00-overview`, `40-automation-rules`, `03-hospitalization-flow`, `09-hospitalization`
- delivery OPERATION_MANUAL RQ-002 注記を更新
- parent GOAL: completed（修復キュー docs+manual の RQ は解消。D-254 FAQ/スクショ・外部 UAT 原証跡は別残件として LEDGER に残す）

## DOCS-REFRESH-ISSUE-RECONCILIATION

- 依頼: `ecc:loop-design-check` に従い、GitHub Issue等の仕様を理解したうえで `docs/` 全体を最新化（2026-09-07）。先行子契約とは別のdocs-only追加依頼。
- 受入状態: **awaiting-human-verification**。先行GOALのcompletedは開始時からある値で、今回の人間受入に転用しない。
- 範囲と根拠: [Issue照合レポート](ISSUE-RECONCILIATION.md)。既存199 Markdown + GOAL.yaml、35補助ファイルはバイト列を保持。
- 入力: HEAD `267a17e48ed8f0cc1944d25ad11900cdbd433226` と開始時WIP、GitHub全144 Issue・コメント957件の取得スナップショット。コメントにはPR分も含み、個々の文書では関連Issueのコメントだけを根拠にする。
- 分担: architecture / spec / ops / workの独立read-only読者、rootがdelivery・governance確認と唯一のwriter。受入仕様・検証器は独立担当が編集前に固定。最終判定は独立レビューへ渡す。
- loop設計: 週次反復需要の証拠がないため常設ループ・cronは作らない。一回の手動実行をplan → build → judgeへ分離し、同一失敗は最大3回。人間受入は自動化しない。
- 今回の修復: 薬品/物販在庫UI案内、#255の受領履歴と現在版/apply待ちの分離、ADR-006の36/14補足、RQ-002と先行状態の古い管理説明。
- 検証境界: docs-onlyのためruntime検証不要。UAT・provider・Linear現在値は未確認。先行EVIDENCEは再生成せず保存する。
- claim: `claim/DOCS-REFRESH-ISSUE-RECONCILIATION`。統合/明示放棄後の解放はUSERのみ。
