---
id: "DOCS-PERFECT-OPS"
objective: "手順・証跡の所在と現在確認の境界を明示し、ローカル不在証跡の誤読と一時ログ保存先の曖昧さを取り除く。"
status: "COMPLETE"
scope: "docs本文保守の一カテゴリ。全対象MarkdownはINVENTORYの該当カテゴリ行から再列挙する。coordinator成果物は除外。"
allowlist: ["docs/ops/**"]
done_condition: "対象Markdown全件に現行の根拠と判定があり、割当RQが修復済みまたは明示的な外部依存として正本へリンクされ、文書の未解消CRITICAL/HIGHが0。必須検証と独立レビューの直接証拠を最終報告に残す。外部依存の解消そのものを要求する項目は未取得ならBLOCKEDとし、COMPLETEにしない。"
verification: ["開始時に git status --short と git branch --list claim/DOCS-PERFECT-OPS を確認。claimがあれば編集せずBLOCKED。なければ git branch claim/DOCS-PERFECT-OPS を取得。claim削除はUSERのみ。", "bash scripts/check-docs-symbol-drift.sh", "git diff --check -- docs/ops", "git diff --name-only と git status --short を開始時baselineに比較し、tracked/staged/untracked全ての新規差分がallowlist内であることを検査する。", "test -f docs/ops/testing/scripts/check-uat-env.sh", "rg -n 'reports/uat-|gitignore|原証跡|保存先|デプロイ直後ログ' docs/ops/testing/UAT-DOMAIN-STATUS.md docs/ops/deploy/STG-CONTINUOUS-OPERATIONS.md", "全Markdownのローカルリンクを文書相対で解決し、ローカル不在は明示した非配布証跡識別子/承認済み保管先への案内へ変更する。原証跡は捏造・複製しない。", "Makefile、scripts、.github/workflows、backend/wrangler.jsonc とコマンド名/設定参照を静的照合する。runbookにあるコマンドは実行しない。", "RQ-005の入口リンクが残り、配布未完了ゲートを隠していないことを確認する。STG/PROD現在状態は未取得ならUNKNOWN。"]
acceptance_tests: ["各INVENTORY対象行のclassificationと根拠を再照合する。開始時/終了時のファイル集合差分を説明し、スキップは理由と次のownerを記録する。", "修正した全claimにコード/機械検査/採択根拠の直接参照があること。古い測定SHAだけを理由に履歴を書換えない。", "driftが赤なら正確な失敗を報告し、担当外の修復に進まない。drift通過を本文全数一致やruntime合格の代用にしない。", "独立レビューでCRITICAL/HIGHを解消し、各要求をPASS/FAIL/BLOCKEDに照合。FAILを残して最終報告しない。"]
forbidden: ["edits outside allowlist", "git commit/push/PR/merge/force", "inventing product requirements", "アプリコード変更・migration/seed/DB操作・STG/PROD操作", "外部posting・credential変更・有料action", "フルtest/lint/build・make codegen", "他者WIPの破棄・claim branch削除", "docs/work/docs-perfection/** の変更（統合は親coordinatorが直列で行う）", "秘密/PHI/本番接続値の転記", "次カテゴリ/依存unitの自動開始"]
adr_policy: "dated-correction-default; in-place factual fix only with correction note; rationaleの静かな改変は禁止。ADRがallowlist外なら提案のみ。"
priority: "P2"
depends_on: ["DOCS-PERFECT-SPEC"]
repair_queue_refs: ["RQ-004", "RQ-005", "RQ-006", "RQ-008"]
risk_level: "Local write; docs-only"
rollback_plan: "開始時baselineと自分のパッチを保存する。必要なら自分が変更した行だけを逆パッチし、他者差分を保持する。削除はdelete-zone/後継リンク/再作成条件を先に記録。全体restore/resetやclaim削除をしない。"
receipt_destination: "最終応答に変更パス・検証コマンド/exit/出力・各要求判定・外部依存を記録。親coordinatorがLEDGERへ統合する。"
---

# Child goal: DOCS-PERFECT-OPS

[Role map](../ROLE-MAP.md) / [Inventory](../INVENTORY.md) / [Repair queue](../REPAIR-QUEUE.md) / [Parent ledger](../LEDGER.md)

開始前に AGENTS.md → .claude/CLAUDE.md と対象の最寄りCLAUDE.mdを読む。親のcoordinator-completeだけを全docs完了とみなさない。以下の依存は順序を定めるもので、このファイルの作成が子実行を許可するものではない。新しい実行依頼でこの子だけを起動する。
