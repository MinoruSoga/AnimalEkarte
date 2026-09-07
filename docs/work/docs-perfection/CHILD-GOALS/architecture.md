---
id: "DOCS-PERFECT-ARCHITECTURE"
objective: "現行構造・型別owner表・機械allowlist・ADR履歴の参照先を一本化し、時点性を失わず技術説明を検証する。"
status: "COMPLETE"
scope: "docs本文保守の一カテゴリ。全対象MarkdownはINVENTORYの該当カテゴリ行から再列挙する。coordinator成果物は除外。"
allowlist: ["docs/architecture/**"]
done_condition: "対象Markdown全件に現行の根拠と判定があり、割当RQが修復済みまたは明示的な外部依存として正本へリンクされ、文書の未解消CRITICAL/HIGHが0。必須検証と独立レビューの直接証拠を最終報告に残す。外部依存の解消そのものを要求する項目は未取得ならBLOCKEDとし、COMPLETEにしない。"
verification: ["開始時に git status --short と git branch --list claim/DOCS-PERFECT-ARCHITECTURE を確認。claimがあれば編集せずBLOCKED。なければ git branch claim/DOCS-PERFECT-ARCHITECTURE を取得。claim削除はUSERのみ。", "bash scripts/check-docs-symbol-drift.sh", "git diff --check -- docs/architecture", "git diff --name-only と git status --short を開始時baselineに比較し、tracked/staged/untracked全ての新規差分がallowlist内であることを検査する。", "rg -n 'Decision ownership|型|owner|history|historical|履歴' docs/architecture/overview.md docs/architecture/model-write-owner-catalog.md docs/architecture/be9-2a-boundary-map.md", "ERD/auth/overviewの現行主張を backend/migrations/001_init.sql、backend/internal/model/permission.go、backend/internal/lintscan の該当定義と行単位で突合する。DDLを適用しない。", "全ADR差分を手動監査し、日付・訂正内容・理由・直接証拠が付かない本文の事実修正と、採択理由の静かな書換えが0件であることを確認する。"]
acceptance_tests: ["各INVENTORY対象行のclassificationと根拠を再照合する。開始時/終了時のファイル集合差分を説明し、スキップは理由と次のownerを記録する。", "修正した全claimにコード/機械検査/採択根拠の直接参照があること。古い測定SHAだけを理由に履歴を書換えない。", "driftが赤なら正確な失敗を報告し、担当外の修復に進まない。drift通過を本文全数一致やruntime合格の代用にしない。", "独立レビューでCRITICAL/HIGHを解消し、各要求をPASS/FAIL/BLOCKEDに照合。FAILを残して最終報告しない。"]
forbidden: ["edits outside allowlist", "git commit/push/PR/merge/force", "inventing product requirements", "アプリコード変更・migration/seed/DB操作・STG/PROD操作", "外部posting・credential変更・有料action", "フルtest/lint/build・make codegen", "他者WIPの破棄・claim branch削除", "docs/work/docs-perfection/** の変更（統合は親coordinatorが直列で行う）", "秘密/PHI/本番接続値の転記", "次カテゴリ/依存unitの自動開始"]
adr_policy: "dated-correction-default; in-place factual fix only with correction note; rationaleの静かな改変は禁止。ADRがallowlist外なら提案のみ。"
priority: "P1"
depends_on: ["DOCS-PERFECT-PRODUCT-PHILOSOPHY"]
repair_queue_refs: ["RQ-011"]
risk_level: "Local write; docs-only"
rollback_plan: "開始時baselineと自分のパッチを保存する。必要なら自分が変更した行だけを逆パッチし、他者差分を保持する。削除はdelete-zone/後継リンク/再作成条件を先に記録。全体restore/resetやclaim削除をしない。"
receipt_destination: "最終応答に変更パス・検証コマンド/exit/出力・各要求判定・外部依存を記録。親coordinatorがLEDGERへ統合する。"
---

# Child goal: DOCS-PERFECT-ARCHITECTURE

[Role map](../ROLE-MAP.md) / [Inventory](../INVENTORY.md) / [Repair queue](../REPAIR-QUEUE.md) / [Parent ledger](../LEDGER.md)

開始前に AGENTS.md → .claude/CLAUDE.md と対象の最寄りCLAUDE.mdを読む。親のcoordinator-completeだけを全docs完了とみなさない。以下の依存は順序を定めるもので、このファイルの作成が子実行を許可するものではない。新しい実行依頼でこの子だけを起動する。
