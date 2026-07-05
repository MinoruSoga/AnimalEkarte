# クローズ済み Issue 仕様適合監査レポート → 未対応 backlog

> **本書の役割**: 2026-07-02 に実施したクローズ済み Issue 45件の仕様適合監査レポートを土台に、対応済み項目を随時削除して**未対応・残タスクのみの backlog**として運用している。以下の「調査日／対象／方法／判定サマリ」は監査実施時点の統計情報（履歴として保持）であり、実際に対応すべき項目は「対応状況」以降の各セクションのみを参照すること。

- **調査日**: 2026-07-02
- **対象**: 直近2週間（2026-06-18〜2026-07-02）にクローズされた GitHub Issue 45件（Search API と全件ローカルフィルタの2経路で突合し取りこぼしゼロを確認、全件 COMPLETED クローズ）
- **方法**: 各 Issue の本文・コメント・クローズ根拠を取得し、HEAD (main) の実コードと静的突合。CRITICAL/HIGH 全件＋主要 MEDIUM（M-4, M-6, M-8, M-9 および #181/#182/#188 の裏取り）は本体セッションで実コードを再検証済み。テスト実行・DB照会は未実施（静的検証のみ）。

## 判定サマリ（監査時点の統計・履歴参照用）

| 判定 | 件数 | 内訳 |
|---|---|---|
| MATCH（仕様充足） | 24 | #124 #125 #126 #127 #128 #151 #152 #153 #158 #159 #161 #179 #180 #182 #183 #184 #187 #188 #190 #191 #192 #198 #123 #193 |
| PARTIAL(一部未達) | 7 | #150 #155 #189 #194 #195 #196 #197 |
| MISMATCH(不一致) | 1 | #154 |
| CLOSED-AS-DECISION(実装なし・設計判断/リスク受容) | 13 | #156 #160 #178 #181 #185 #212 #89 #91 #96 #97 #98 #99 #109 |

## 対応状況（2026-07-05 更新 — 対応済み項目は本書から削除済み）

対応済み・削除済み: H-1/H-2 (`a45da439`) / H-3 (`4774666a`) / H-4 (`a620bdfc`) / H-6 (`aa9c0a5d`) / M-1/M-2 (`9d3df80c`) / M-8 (`5b0ac22d`) / M-9 (`ba8cecea`) / M-12 (`3f836cda`)、C-2（park 原本6 Issue #89 #91 #97 #98 #99 #109 の再オープンで代替解決）、仕様未充足 Issue 11件の再オープン、M-1 最終仕様化／M-3 越日EMG (`9ab95845`)／M-5 #178 保留クローズ／M-6 孤児API削除／M-7 税率 exact-match／M-10 `.gitignore` 明文化 (`830ee8e9`)／M-10 本番デモ非表示／H-7 カバレッジ ratchet＋baseline（#194 含む）／B-1 完全体（pagination・sort・E2E・service test mock）、H-5/C-1 **コード側**（SSM valueFrom workflow・gitleaks CI・外部 ops ランブック）。詳細は各コミット・`docs/coverage-policy.md`・`docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md` を参照。**全コミット push 未実施**。

**残っているのは外部操作（AWS/GitHub）と PO 判断待ち（#189）のみ。**

### 残項目

| 残項目 | 状態 | 手順書 |
|---|---|---|
| C-1 ローテーション／#97 実値削除／`.env.staging` 追跡解除 | ⏳ **ユーザー実施要** | [ランブック §1, §3](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) |
| H-5 SSM 実登録＋Terraform apply＋デプロイ検証 | ⏳ **ユーザー実施要** | [同 §2](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) |
| M-4 STG `db_reset=true` デプロイ | ⏳ **ユーザー承認要**（破壊的操作） | [同 §4](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) |
| M-11 GitHub Secrets 登録 | ⏳ **ユーザー実施要** | [同 §5](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) |
| #189 締め後訂正の可視化バッジ | 🚫 **BLOCKED（PO 未決）** | 下記参照 |

---

## 🚫 保留 (PO Decision Pending)

### #189. 締め後訂正の可視化バッジ

**事象**: レジ締め後に金額が訂正された場合、締め詳細画面・月次集計レポートにその事実を示す UI 表示（バッジ等）が存在しない。FE 側の警告表示自体は対応済み（`0dad744e`）だが、事後可視化の要否は PO 未決。

**方針**: **デフォルトでは実装しない**。PO が「バッジ要」と明示するまで着手しない。

**PO 承認後の参考**（未着手）: 締め詳細＋月次集計に最小バッジ（`audit_logs` の締め後書き込み検出）。

---

## 🔴 CRITICAL

### C-1. [#97/#89] 平文シークレット露出 — ローテーション未実施

**残タスク**: `.env.staging` 平文露出・Issue #97 本文実値の解消。[ランブック §1（ローテーション）・§3（追跡解除・#97 編集）](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md)。**ユーザー承認必須**。

**検証**: `git ls-files \| grep -E '\.env'` で `.env.staging` が消えること。旧資格情報でのアクセス拒否。

---

## 🟠 HIGH

### H-5. [#99] ECS secrets — SSM 実登録・デプロイ待ち

**残タスク**: [ランブック §2](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) — Terraform apply → SSM 新値登録（C-1 後）→ デプロイ → `aws ecs describe-task-definition` で平文シークレット不在を確認。**ユーザー承認必須**。

---

## 🟡 MEDIUM

### M-4. [#160 vs #211] 健診系統 — STG db_reset 待ち

**残タスク**: [ランブック §4](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md)。Checkup 系スキーマは `001_init.sql` に統合済み。STG 反映には `db_reset=true` 必須。**破壊的操作・ユーザー承認必須**。

**検証**: fresh DB apply 後、健診入力導線が Checkup 系1系統のみであること（UI）。

### M-11. [#109] CI テスト認証情報 — Secrets 登録待ち

**残タスク**: [ランブック §5](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) — `CI_TEST_EMAIL` / `CI_TEST_PASSWORD` を `gh secret set`。**ユーザー実施要**。

---

## 🟢 LOW（記録のみ・対応は任意）

| Issue | 内容 | 推奨対応 |
|---|---|---|
| #182 | クローズコメント「OPEN維持」と CLOSED 状態の矛盾 | 未投稿・承認待ち: `gh issue comment 182 --body "子 #188/#189 完了に伴いクローズ"` |
| #179 | クローズコメントの migration 番号（010/008）が `001_init.sql` 統合後と乖離 | 未投稿・承認待ち: `gh issue comment 179 --body "migration は 001_init.sql へ統合済み。010/008 表記は統合前の呼称"` |
| #184 | 印刷ビューの実ブラウザ視覚検証未実施 | 任意 follow-up（Playwright 基盤あり） |

---

## 本監査の対象外とした既知の横断事項

- **監査ログ書込が tx 外 best-effort**（`auditRepository.Create` が dbOrTx 非使用）: 各 Issue 判定には含めず、#211 系 follow-up で別途追跡。

---

## 推奨アクション（優先順・依存関係つき）

| # | 対応 | 対象 | 依存 |
|---|---|---|---|
| 1 | シークレット即時ローテーション＋Issue #97 実値削除 | C-1 | なし（最優先） |
| 2 | SSM 登録 → Terraform apply → デプロイ → `.env.staging` 追跡解除 | H-5 → C-1-3 | 1 の後 |
| 3 | STG `db_reset=true` デプロイ | M-4 | デプロイタイミング・ユーザー承認 |
| 4 | GitHub Secrets 登録 | M-11 | なし |
| 5 | #189 バッジ要否 | PO 判断 | なし |
