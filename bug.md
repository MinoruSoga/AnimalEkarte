# クローズ済み Issue 仕様適合監査レポート → 未対応 backlog

> **本書の役割**: 2026-07-02 監査レポートを土台に、**未対応・残タスクのみ**を運用する backlog。「判定サマリ」は監査時点の履歴。対応すべき項目は「残項目」以降のみ参照。

- **最終コード調査**: 2026-07-06（`git ls-files`、`backend-deploy.yml`、`backend-deploy-ecs.yml`、`performance-tests.yml`、`001_init.sql`、cash-register FE、ランブック Cloudflare 正系統整合）

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
| C-1 シークレットローテーション／#97 実値削除 | ⏳ **ユーザー実施要（credential-impacting）** | [ランブック §1](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) |
| #189 締め後訂正の可視化バッジ | 🚫 **BLOCKED（PO 未決）** | 下記 |

> **C-1 検証（2026-07-06）**: `.env.staging` の git 追跡解除（`git rm --cached`）はコード側で実施済み・コミット待ち。
> **未解決**: 平文値そのもののローテーション（PlanetScale DB パスワード・`JWT_SECRET`・`INTEGRATION_ENCRYPTION_KEY` の3点。
> `.env.staging` に LINE 系キーは含まれておらず C-1 のスコープ外 — ランブック §1.5 参照）は未実施。
> untrack は「今後の新規コミットに平文が乗らないこと」を保証するのみで、**過去 16 コミット分の履歴には現行の平文値がなお残存する** —
> 履歴を持つ全員にとって当該値は引き続き有効な認証情報として扱い、ローテーションをもって初めて C-1 は解消する。
> ランブック §1・§3 参照。また `backend-deploy-ecs.yml`（ECS ロールバック専用）は `.env.staging` を
> チェックアウト済みリポジトリから直接読む実装のため、untrack 後にこの経路を使う場合は別途復元手順が必要（ランブック §2 に記録）。

---

## 🔴 CRITICAL — C-1. [#97/#89] 平文シークレット露出

**完了（コード側）**: `.env.staging` の git 追跡解除（untrack コミット）。

**残タスク**: シークレットローテーション（PlanetScale DB パスワード／`JWT_SECRET`／`INTEGRATION_ENCRYPTION_KEY`）、Issue #97 本文の実値削除。[ランブック](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) §1・§3。**ユーザー承認必須（credential-impacting）**。

**PASS 条件**: 上記シークレットが新値へローテーション済み、旧資格情報でのアクセスが拒否されること。`.env.staging` untrack（コード側完了）はローテーション完了の代替にはならない。

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
