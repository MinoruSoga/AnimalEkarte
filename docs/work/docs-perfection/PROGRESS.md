# Docs perfection coordinator progress

本書はドキュメント保守ユニットの実行記録。製品作業の状態・担当・Done の正本は [work/README](../README.md) が指す Linear。親の目的・停止条件は [GOAL.yaml](GOAL.yaml) を参照。

## 実行前計画 — 2026-09-07

1. 保存 prompt の harness/receiver、6スキル hash、現行ルール、claim と既存 WIP を確認する。
2. 全 Markdown を列挙し、役割・Git更新日・構造・ローカル参照を機械で調べる。5カテゴリの独立読者は根拠付き矛盾候補を調査する。
3. 主エージェントだけが ROLE-MAP / INVENTORY / REPAIR-QUEUE / CHILD-GOALS / LEDGER を作成する。本文修復は子に残す。
4. inventory集合一致、子schema、drift実行、baselineとの差分、秘密の混入を検証する。
5. 同一rubricの独立2レビュー、指摘修正、全要件の再照合を経て coordinator-complete を記録する。親は active。

## 要求と検証の対応（編集前に固定）

| Success criterion | Acceptance | 判定方法 |
|---|---|---|
| SC1: 4役割・1 fact = 1 canonical owner | AC1 | 4見出しを rg。各ownerが実在し、重複のリンク方針を手動照合 |
| SC2: 全対象Markdown分類 | AC2 | find集合とinventory集合の双方向差分・重複・enum検査。各行の更新手がかりと確認範囲を検査 |
| SC3: Severity×Impactの修復キュー | AC3 | 点数順、証拠パス・行番号、推奨actionの存在を検査。矛盾の両側を読み直す |
| SC4: 最低6カテゴリ子goal | AC4 | ファイル名・必須キー・個別done_conditionとverification・DAG非循環を検査 |
| SC5: activeとcoordinator-complete | AC5 | YAML状態・checkpointと本書を照合。親completedを禁止 |
| SC6: drift実行の記録 | AC6 | bash scripts/check-docs-symbol-drift.sh のexitと出力を保存。緑であることは本ユニットの条件ではない |
| SC7: allowlist限定 | AC7, AC8 | baselineの既存docs hashと現在を比較、tracked/staged/untrackedを確認。カテゴリ本文の差分なし |
| SC8: 完了監査 | AC1–AC9 | LEDGERに全ACの期待・実際・判定・直接証拠を記録。弱い証拠はPASSにしない |
| 独立Review | AC9 | 同一入力・同一rubricの独立2レビュー。CRITICAL/HIGHを修正し両方再レビュー |

AC4 の prompt 記載 `rg -L` は ripgrep では follow symlinks を意味するため、各ファイルについて必須キーを個別に検査する。AC7 は開始時点で外部WIPがあるため、本ユニットが生んだ差分を baseline-relative に検査し、既存WIPを戻さない。

## Checkpoint: coordinator-planned

- budget: 3/24 actions。validation/hash/preflight → baseline/claim → plan。
- baseline: [EVIDENCE/BASELINE.json](EVIDENCE/BASELINE.json)。docsには開始時差分なし。主エージェント以外は読み取り専用で、共有台帳のwriterは一人。
- source prompt: `/Users/minoru/.claude/prompt-craft-runs/agent-docs-perfection-coordinator-codex-astra-v2.md`。
- promptSha256: `b1e9b0c3de8509ef8ce81dd0d638a6e4b9bb3fd2e3e64f47ec6ab2b6ec20a505`。
- harness=PASS / receiver=PASS / exit=0。権限は実セッションと `.codex/config.toml` とも danger-full-access / never。
- skills: ecc:living-docs-governance / ecc:continuous-agent-loop / ecc:ralphinho-rfc-pipeline / ecc:santa-method は native。ecc:source-command-update-docs / ecc:verification-loop は path-fallback（全6ファイルのhash一致を検証）。inline:goal は本書とGOAL.yamlで実行。
- Orchestration: inventory_architecture / inventory_spec / inventory_ops / inventory_delivery / inventory_work の5読者。writer-owned pathsは全員 none。保存promptの独立並列読み取り指定に従い、共有treeの現在証拠を読む。子goal実装や並列writerは起動しない。
- remaining: 全ファイルの棚卸し統合、子DAG作成、機械検証、独立レビュー。
- risks: driftの緑は本文全体の正確性を保証しない。外部状態は今回検証しない。
- next: inventory と SoT 参照候補を集約。

## Failure Signature log

失敗時は AC・期待・実際・検証・署名・attempt・仮説/修正・結果を追記する。2回同署名なら最小単位に縮小、3回ならBLOCKED。ループ専用監視ツールは使わず、本書のaction数と検証結果で追跡する。

| AC / signature | 期待 / 実際 | 検証・attempt | 仮説 / 修正 / 結果 |
|---|---|---|---|
| AC2 / verifier-row-width | 全行の分類/更新日/理由を検査する / 検証補助が末尾区切りを余分な列として数えた | `python3 docs/work/docs-perfection/EVIDENCE/verify.py`、attempt 1、exit 1: `FAIL inventory classification/freshness/reason all rows` | 実際の1行をsplitして5フィールドと確認。補助の期待列数を6→5に訂正し再実行する。inventory本体の分類は変更しない |
| AC3 / citation-line | 引用が該当文を指す / RQ-010が160行を指すが補助的同意は162行 | Santa B/C round 1、LOW（verdictは両方PASS） | 元文書155–164行を再読し引用を162へ訂正。scope不変 |
| AC1/2 / historical-role | ROLE-MAPとinventoryの役割一致 / ARCH-A4の行だけStatus | Santa B/C round 1、LOW（verdictは両方PASS） | 原文3–4行の履歴宣言に従いinventoryとscanをHistoryへ統一。runtime/現行Statusへ昇格させない |
| AC1–9 / final-recorder-quote | 最終検証の出力を保存 / inline記録用Pythonのコマンド文字列に閉じquote不足 | attempt 1、exit 1: `SyntaxError: unterminated string literal (detected at line 14)`。検証自体の実行前 | コマンド配列末尾の引用符を修正。再実行で9コマンドすべてexit=0、`RESULT PASS phase=final`。成果物内容の修復は不要 |

## Checkpoint: coordinator-artifacts-verified

- budget: 12/24 actions。loop iteration 1。失敗signatureは検証補助の列数1件、最小修正後 `RESULT PASS phase=draft`（exit=0）。再発なし。
- evidence: [checks-draft.txt](EVIDENCE/checks-draft.txt)（12コマンドの生出力）、[docs-symbol-drift.txt](EVIDENCE/docs-symbol-drift.txt)。
- inventory: 187既存Markdown + 11新規Markdown = 198行、find=198、集合のmissing/extra/duplicate=0。新規11件は各行にskip-with-reasonがあり、別の全AC検証を受ける。非Markdownはfind対象外。
- integrated: RQ 11件、うちRQ-005は索引だけ修正。6子goalをpendingで作成。各allowlist・完了条件・検証・依存・rollbackを確認。
- De-Sloppify: 旧SHAだけを根拠にしたADR訂正、helper不存在の誤検出、既に分離されたwork記録の再分割を棄却。表現上の弱い懸念は低優先の候補として扱い、実装欠陥と断定しない。
- manual security/PHI pass: 新規本文はパス・保守方針・サニタイズ済みの静的検証のみ。患者/飼主実データ・秘密値・本番接続値を取り込んでいない。自動パターン検査は補助。
- scope: 開始時docs外tracked内容の集約hash、staging差分が不変。既存docs変更はREADME2件の各1リンクだけ。
- remaining: 独立B/Cレビュー、必要な指摘修正、全ACの最終照合、coordinator-complete checkpoint。
- next: 同じrubricと成果物を、相互の評価を見ない独立2読者へ渡す。

## Checkpoint: review-round-1

- budget: 15/24 actions。loop iteration 2。B/CともPASS、CRITICAL/HIGH=0。重複したLOW2件を修正。
- evidence: [review-B-round1.json](EVIDENCE/review-B-round1.json) / [review-C-round1.json](EVIDENCE/review-C-round1.json)（verdictと採用指摘のreceipt）。
- integration: 各レビューの引用行訂正とHistory分類統一を採用。全件inventoryの役割をROLE-MAPへ再照合する。本文カテゴリの編集は増えていない。
- remaining: 修正後の独立fresh B/C再確認、最終受入記録とcoordinator-complete。
- next: 既存レビューの評価を渡さず、同じ元仕様と修正後成果物をfresh reviewerへ渡す。

## Checkpoint: coordinator-complete

- budget: 19/24 actions、1 session。loop iteration 2を完了。coordinatorユニットだけの完了であり、親GOAL.yamlのstatusはactive。
- completed: ROLE-MAP、198行のINVENTORY、11件のREPAIR-QUEUE、6子goal、GOAL/LEDGER、2索引リンク、静的gate/範囲監査、独立2レビュー×2round、要件別完了監査。
- evidence: [LEDGERのCompletion audit](LEDGER.md#completion-audit)、[checks-final.txt](EVIDENCE/checks-final.txt)、[review-B.json](EVIDENCE/review-B.json)、[review-C.json](EVIDENCE/review-C.json)。
- review integration: fresh B2/C2ともPASS、CRITICAL/HIGH=0。B2の比較範囲の表現限定を採用（docs外tracked/indexの不変とuntrackedのstatus維持。untracked内容hashは未比較）。C2のallowlist交差の自動検査追加は、全6件の手動検証が済んでいるため任意改善として保留。
- Failure Signatures: verifier-row-widthとfinal-recorder-quoteは修正後exit=0。citation-line / historical-roleはLOW訂正後、fresh両者PASS。同じ失敗署名の再発なし。
- regression boundary: アプリ/runtimeは未変更でN/A。本文の既知gap、UAT/外部状態、docs外manual修復をPASSへ昇格させていない。
- remaining: 子カテゴリ本文の修復と全命題の照合。RQ-002の別承認依存。親goalはこれらを消さず継続する。
- next: [product-philosophy child](CHILD-GOALS/product-philosophy.md) を新しい実行依頼で起動。claimは保持し、USER以外は解放しない。


## Checkpoint: child-product-philosophy-complete (2026-09-07)

- done: RQ-010 wording fix integrated; inventory/repair-queue/ledger/child status updated by parent loop
- evidence: docs/product-philosophy.md:41; Mode 3 reconciliation PASS
- remaining: architecture → spec → ops → delivery → work
- next: DOCS-PERFECT-ARCHITECTURE (RQ-011)


## Checkpoint: parent-continuous-loop-children-complete (2026-09-07)

- done: architecture/spec/ops/delivery/work RQ 修復を親 continuous loop で直列実行し台帳統合
- evidence: git diff on listed docs; drift exit 0; LEDGER Parent continuous loop section
- remaining: RQ-002 frontend manual（別承認）; claim 解放は USER; 親 GOAL は active
- next: USER が claim 解放と RQ-002 別 unit を判断

## Checkpoint: astra-loop-phase-a — 2026-09-07

- plan: Phase Aの調査配置 → Phase Bの全先行docs差分採否 → Phase Cのarchitecture/spec/ops/delivery/work直列受入 → scoped検証 → 独立2レビュー → 親台帳統合。
- done: 保存prompt検証exit0、4スキルhash一致、入口リンク欠落0。READMEのinventory漏れ1件を最小補完。本文の再監査前に配置を確認した。
- baseline: EVIDENCE/astra-loop-baseline.json（tracked4602、untracked27）、生成側の先行本文diffをastra-loop-draft.patchへ保存。
- loop health: iteration A、category units 0/5、santa fix cycles 0/2。専用loop監視は未使用、ここで段階・残予算・gate結果を記録。
- remaining: 先行本文下書きの採否、子契約の現行根拠照合、独立レビュー。
- deviations: 新promptが明示する既存claim再利用と親の連続実行を優先。旧coordinator予算/子自動開始禁止は当時の契約として扱う。今回の上限は5 category units、santa修正2回。
- Failure Signature: A1/inventory-entry-missing、期待=全Markdown列挙、実際=199実在に対し198行、検査=集合比較、attempt1、修正=README行追加、結果=再検証待ち。

## Checkpoint: astra-loop-phase-b — 2026-09-07

- done: 全12先行dirty docsをadopt7/amend5/revert0へ分類しLEDGERに根拠を記載。
- evidence: Phase A再検証は `PASS Phase A inventory=199 actual=199 missing=0 extra=0 duplicate=0`、`PASS Phase A investigation entry reachable; product Linear SoT unchanged`。A1/inventory-entry-missingはattempt1で解消。
- loop health: iteration B、category units 0/5、santa fix cycles 0/2。次はarchitectureから直列受入。
- orchestration: 新規read-only loop_architecture / loop_specを起動。追加laneのspawnは `agent thread limit reached` で失敗したため既存inventory_opsを再利用。全員writer-owned paths none。主のみ書込。
- source findings: RQ-003下書きの必須要件弱化とsettings受信/item-master混同を修正対象にした。範囲は既存子allowlist内。

## Checkpoint: astra-loop-categories-verified — 2026-09-07

- done: 全5カテゴリを順序どおり静的照合。先行本文を採用/限定修正し、追加の31-lstep仕様は同じRQ003説明の参照漏れだけ補完。
- evidence: category checks JSON、全199ファイルscan、architecture22/spec80/ops73の本文分類reader receipt、delivery4/work6/PP2の主照合表。
- verification: drift exit0 `OK    docs-symbol-drift: ドリフトなし（検査トークン 565 件）`。`verify-astra-loop.py` は draft PASS。
- loop health: 5/5 category units、0/2 santa fix cycles。子statusはIN_REVIEW、親active。
- failure recovery: 新規spec laneのusage-limitを既存read-only laneで補完。H1-only不足は80件本文の追読で解消。RQ004の未確認所持断定HIGHは修正・読者再確認済み。
- scope: tracked本文の自編集6件、親packageのみ追加更新。docs外2 lintscan WIPの並行変化を観測したが編集/復元せず別記。index/HEAD不変。
- remaining: 同一rubric独立2レビュー、必要時修正、final metadataと全ACの検証。

## Checkpoint: astra-loop-concurrent-wip-observed — 2026-09-07

- Failure Signature: scope/ADR-global-diff-empty。期待=自分のADR変更0、実際=後発foreign ADR差分3行。draft記録はassert exit1、最小verify再実行は `FAIL ADR history diff empty`。
- attempt2の調査: ADR/関連2文書の差分とpackage pinを直接確認。生成patchではなく作業中の共有WIPと分離し、baselineを保持したままownership receiptで自差分検証へ限定。
- evidence: EVIDENCE/astra-loop-ownership.jsonのprior/current observations、LEDGERの後発共有WIP節。自己編集6件に後発3件を吸収しない。
- loop health: 5/5 category units、0/2 santa fix cycles。gate再検証と独立レビュー待ち。

## Checkpoint: astra-loop-complete — 2026-09-07

- done: Phase A/B/C、独立B/C PASSを親に統合。A1–A8 PASS、全6子は文書受入COMPLETE。親GOALはactive。
- evidence: EVIDENCE/astra-loop-review-{B,C}.json、astra-loop-checks-final.txt、astra-loop-drift.txt。原文出力/exitは生ログ参照。
- loop health: category units 5/5、Santa fix cycles 0/2。3回同一失敗は0。今回のloopを終了し追加campaignは開始しない。
- integration: 主writerのみ。readersは完了、loop_specはusage-limit後にinventory_specで補完。独立2レビュー以後は最終metadata/証拠のみ更新。
- remaining: RQ-002 frontend manualの別承認修復、未取得の外部状態/原証跡。runtime/UAT/本番受入を静的PASSへ昇格しない。
- scope: 自tracked本文6件＋親package更新、先行adopt7件。後発foreign docs3件・docs外5件は時点別観測として保持。claim8件を保持し、統合/明示放棄後の解放はUSER。


## Checkpoint: rq-002-manual-resolved (2026-09-07)

- done: frontend manual の「カルテ確定→会計自動作成」誤記を実装整合。退院は create_accounting 選択時のみ。
- remaining: claim 解放は USER。D-254 FAQ/スクショ、外部証跡の再取得は別件。

## 2026-09-07 — GitHub Issueを含む全件照合

追加依頼の現在記録は [Issue照合レポート](ISSUE-RECONCILIATION.md)。開始時のマニュアル修復差分とGOAL completedは既存WIPとして保存し、過去の未修復/active記録は当時の履歴として読む。

独立担当が受入仕様と検証器を先に固定し、読者は別worktreeで並列調査、rootだけが文書を変更した。GitHubの本文・コメントを照合し、IssueのCLOSEDを実行成功へ読み替えない。Linearは取得経路を試したが現在値を取得できずUNKNOWN。

初期の構造・見出しだけの証拠は本文検証として採用せず、直接のコード/契約/履歴根拠へ補強した。飼主検索の指摘は実際のloaderが `/pets` を使うため撤回。トリミング画像はrequest builderの項目省略を確認し、既存の説明を保持。GitHubコメント一括取得はCLIオプション非互換を分離して再実行し、957件取得した。受入条件は緩和していない。

修復後の機械照合・独立レビューは候補の文書全体hashに結び付ける。共有treeへ反映する直前に開始時hashを再照合し、他者差分が増えたファイルは上書きしない。今回の人間受入状態はawaiting-human-verification。
