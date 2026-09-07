---
id: "DOCS-PERFECT-SPEC"
objective: "配信安全・bulk-read説明・画面ナビゲーションを既存正本へ整合させ、各画面仕様の実装との一致と意図的な未実装を区別する。"
status: "COMPLETE"
scope: "docs本文保守の一カテゴリ。全対象MarkdownはINVENTORYの該当カテゴリ行から再列挙する。coordinator成果物は除外。"
allowlist: ["docs/spec/**"]
done_condition: "対象Markdown全件に現行の根拠と判定があり、割当RQが修復済みまたは明示的な外部依存として正本へリンクされ、文書の未解消CRITICAL/HIGHが0。必須検証と独立レビューの直接証拠を最終報告に残す。外部依存の解消そのものを要求する項目は未取得ならBLOCKEDとし、COMPLETEにしない。"
verification: ["開始時に git status --short と git branch --list claim/DOCS-PERFECT-SPEC を確認。claimがあれば編集せずBLOCKED。なければ git branch claim/DOCS-PERFECT-SPEC を取得。claim削除はUSERのみ。", "bash scripts/check-docs-symbol-drift.sh", "git diff --check -- docs/spec", "git diff --name-only と git status --short を開始時baselineに比較し、tracked/staged/untracked全ての新規差分がallowlist内であることを検査する。", "rg -n 'HandlePetDeath|N\\+1|best-effort|safety gap|fallback' docs/spec/screens/34-lstep-delivery-monitor.md docs/spec/line/lstep-integration.md docs/spec/line/architecture.md", "RQ-001/003を backend/internal/lstep/lstep_delivery_trigger_state.go と lstep_delivery_trigger_batch.go に再照合し、現行制約を超える保証が0件であることを独立レビューする。アプリ修正や既存安全要件の撤回をしない。", "frontend/src/app/routes の該当lab-device routeとscreens/settings索引の対応を照合。新規の個別仕様を増やす前に既存契約へのリンクで足りるか確認する。", "全spec Markdownについて採択済み要件/現在実装/外部依存の区分と根拠を確認する。UI準拠表のruntime pendingを静的検査でPASSに変えない。"]
acceptance_tests: ["各INVENTORY対象行のclassificationと根拠を再照合する。開始時/終了時のファイル集合差分を説明し、スキップは理由と次のownerを記録する。", "修正した全claimにコード/機械検査/採択根拠の直接参照があること。古い測定SHAだけを理由に履歴を書換えない。", "driftが赤なら正確な失敗を報告し、担当外の修復に進まない。drift通過を本文全数一致やruntime合格の代用にしない。", "独立レビューでCRITICAL/HIGHを解消し、各要求をPASS/FAIL/BLOCKEDに照合。FAILを残して最終報告しない。"]
forbidden: ["edits outside allowlist", "git commit/push/PR/merge/force", "inventing product requirements", "アプリコード変更・migration/seed/DB操作・STG/PROD操作", "外部posting・credential変更・有料action", "フルtest/lint/build・make codegen", "他者WIPの破棄・claim branch削除", "docs/work/docs-perfection/** の変更（統合は親coordinatorが直列で行う）", "秘密/PHI/本番接続値の転記", "次カテゴリ/依存unitの自動開始"]
adr_policy: "dated-correction-default; in-place factual fix only with correction note; rationaleの静かな改変は禁止。ADRがallowlist外なら提案のみ。"
priority: "P1"
depends_on: ["DOCS-PERFECT-ARCHITECTURE"]
repair_queue_refs: ["RQ-001", "RQ-003", "RQ-007"]
risk_level: "Local write; docs-only"
rollback_plan: "開始時baselineと自分のパッチを保存する。必要なら自分が変更した行だけを逆パッチし、他者差分を保持する。削除はdelete-zone/後継リンク/再作成条件を先に記録。全体restore/resetやclaim削除をしない。"
receipt_destination: "最終応答に変更パス・検証コマンド/exit/出力・各要求判定・外部依存を記録。親coordinatorがLEDGERへ統合する。"
---

# Child goal: DOCS-PERFECT-SPEC

[Role map](../ROLE-MAP.md) / [Inventory](../INVENTORY.md) / [Repair queue](../REPAIR-QUEUE.md) / [Parent ledger](../LEDGER.md)

開始前に AGENTS.md → .claude/CLAUDE.md と対象の最寄りCLAUDE.mdを読む。親のcoordinator-completeだけを全docs完了とみなさない。以下の依存は順序を定めるもので、このファイルの作成が子実行を許可するものではない。新しい実行依頼でこの子だけを起動する。
