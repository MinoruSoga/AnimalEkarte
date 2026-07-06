# クローズ済み Issue 仕様適合監査レポート → 未対応 backlog

> **本書の役割**: 2026-07-02 監査レポートを土台に、**未対応・残タスクのみ**を運用する backlog。「判定サマリ」は監査時点の履歴。対応すべき項目は「残項目」以降のみ参照。

- **最終コード調査**: 2026-07-06（`git ls-files`、`backend-deploy.yml`、`performance-tests.yml`、`001_init.sql`、cash-register FE）

## 判定サマリ（監査時点の統計・履歴参照用）

| 判定 | 件数 | 内訳 |
|---|---|---|
| MATCH（仕様充足） | 24 | #124 #125 #126 #127 #128 #151 #152 #153 #158 #159 #161 #179 #180 #182 #183 #184 #187 #188 #190 #191 #192 #198 #123 #193 |
| PARTIAL(一部未達) | 7 | #150 #155 #189 #194 #195 #196 #197 |
| MISMATCH(不一致) | 1 | #154 |
| CLOSED-AS-DECISION(実装なし・設計判断/リスク受容) | 13 | #156 #160 #178 #181 #185 #212 #89 #91 #96 #97 #98 #99 #109 |

## 残項目

| 残項目 | 状態 | 手順書 |
|---|---|---|
| C-1 ローテーション／#97 実値削除／`.env.staging` 追跡解除 | ⏳ **ユーザー実施要** | [ランブック §1, §3](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) |
| #189 締め後訂正の可視化バッジ | 🚫 **BLOCKED（PO 未決）** | 下記 |

> **C-1 検証（2026-07-06）**: `git ls-files \| grep -E '\.env'` に `.env.staging` が残存。`.gitignore` 済みだが untrack 未実施。コード側（gitleaks CI・Cloudflare `wrangler secret`）は完了。

---

## 🔴 CRITICAL — C-1. [#97/#89] 平文シークレット露出

**残タスク**: シークレットローテーション、Issue #97 本文の実値削除、`.env.staging` の git 追跡解除。[ランブック](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) §1・§3。**ユーザー承認必須**。

**PASS 条件**: `git ls-files \| grep -E '\.env'` で `.env.staging` が消えること。旧資格情報でのアクセス拒否。

---

## 🚫 BLOCKED — #189. 締め後訂正の可視化バッジ

**事象**: 締め詳細・月次集計に「締め後訂正あり」バッジがない（`cash-register` FE に該当 UI なし）。

**対応済み（削除済み）**: 会計画面の締め後警告表示、BE 監査ログ（`post_close` メタデータ、`accounting_service_correction.go`）。

**方針**: PO が「バッジ要」と明示するまで実装しない。

---

## 🟢 LOW（任意・コードタスクではない）

| Issue | 内容 |
|---|---|
| #182 | クローズコメントと CLOSED 状態の矛盾 → `gh issue comment` で整理（未投稿） |
| #179 | migration 番号表記のドリフト → 同上 |
| #184 | 印刷ビューの実ブラウザ視覚検証（Playwright e2e 未作成） |
