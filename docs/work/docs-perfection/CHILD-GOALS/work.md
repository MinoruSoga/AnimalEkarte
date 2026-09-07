---
id: "DOCS-PERFECT-WORK"
objective: "補助文書の時点と実行SoTを明示し、削除済み台帳を復活させず採択方針と外部状態を区別する。"
status: "COMPLETE"
scope: "docs本文保守の一カテゴリ。全対象MarkdownはINVENTORYの該当カテゴリ行から再列挙する。coordinator成果物は除外。"
allowlist: ["docs/work/README.md", "docs/work/phase2-deferred.md", "docs/work/linear-f1-f6-mapping.md", "docs/work/skill-reeval-2026-09-06.md", "docs/work/decisions/**"]
done_condition: "対象Markdown全件に現行の根拠と判定があり、割当RQが修復済みまたは明示的な外部依存として正本へリンクされ、文書の未解消CRITICAL/HIGHが0。必須検証と独立レビューの直接証拠を最終報告に残す。外部依存の解消そのものを要求する項目は未取得ならBLOCKEDとし、COMPLETEにしない。"
verification: ["開始時に git status --short と git branch --list claim/DOCS-PERFECT-WORK を確認。claimがあれば編集せずBLOCKED。なければ git branch claim/DOCS-PERFECT-WORK を取得。claim削除はUSERのみ。", "bash scripts/check-docs-symbol-drift.sh", "git diff --check -- docs/work/README.md docs/work/phase2-deferred.md docs/work/linear-f1-f6-mapping.md docs/work/skill-reeval-2026-09-06.md docs/work/decisions", "git diff --name-only と git status --short を開始時baselineに比較し、tracked/staged/untracked全ての新規差分がallowlist内であることを検査する。", "rg -n 'Needs Human|当時|UNKNOWN|Linear|削除済み' docs/work/README.md docs/work/decisions/fable-po-recommendation.md docs/work/linear-f1-f6-mapping.md", "既存6 Markdownの各現在状態表現を照合日/証跡に結び付ける。Linear未照会の状態を現在確認済みと書かない。", "docs/work/README.md の削除zoneの全行とcoordinator入口が維持され、旧STATUS.md/PO-todo.mdの新設が0であることを確認する。", "docs/work/docs-perfection/** は親coordinator専有として一切変更しない。子の最終報告を親が別セッションで直列統合する。"]
acceptance_tests: ["各INVENTORY対象行のclassificationと根拠を再照合する。開始時/終了時のファイル集合差分を説明し、スキップは理由と次のownerを記録する。", "修正した全claimにコード/機械検査/採択根拠の直接参照があること。古い測定SHAだけを理由に履歴を書換えない。", "driftが赤なら正確な失敗を報告し、担当外の修復に進まない。drift通過を本文全数一致やruntime合格の代用にしない。", "独立レビューでCRITICAL/HIGHを解消し、各要求をPASS/FAIL/BLOCKEDに照合。FAILを残して最終報告しない。"]
forbidden: ["edits outside allowlist", "git commit/push/PR/merge/force", "inventing product requirements", "アプリコード変更・migration/seed/DB操作・STG/PROD操作", "外部posting・credential変更・有料action", "フルtest/lint/build・make codegen", "他者WIPの破棄・claim branch削除", "docs/work/docs-perfection/** の変更（統合は親coordinatorが直列で行う）", "秘密/PHI/本番接続値の転記", "次カテゴリ/依存unitの自動開始"]
adr_policy: "dated-correction-default; in-place factual fix only with correction note; rationaleの静かな改変は禁止。ADRがallowlist外なら提案のみ。"
priority: "P2"
depends_on: ["DOCS-PERFECT-DELIVERY"]
repair_queue_refs: ["RQ-009"]
risk_level: "Local write; docs-only"
rollback_plan: "開始時baselineと自分のパッチを保存する。必要なら自分が変更した行だけを逆パッチし、他者差分を保持する。削除はdelete-zone/後継リンク/再作成条件を先に記録。全体restore/resetやclaim削除をしない。"
receipt_destination: "最終応答に変更パス・検証コマンド/exit/出力・各要求判定・外部依存を記録。親coordinatorがLEDGERへ統合する。"
---

# Child goal: DOCS-PERFECT-WORK

[Role map](../ROLE-MAP.md) / [Inventory](../INVENTORY.md) / [Repair queue](../REPAIR-QUEUE.md) / [Parent ledger](../LEDGER.md)

開始前に AGENTS.md → .claude/CLAUDE.md と対象の最寄りCLAUDE.mdを読む。親のcoordinator-completeだけを全docs完了とみなさない。以下の依存は順序を定めるもので、このファイルの作成が子実行を許可するものではない。新しい実行依頼でこの子だけを起動する。
