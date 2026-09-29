# docs-refresh-20260929 — docs/ 全区画 最新化キャンペーン INDEX

> **依頼（2026-09-29）**: 「docs/ 配下の内容を全て最新化。間違った情報・古い情報を入念にチェック。仕様はスプシ/Slack の先方やり取り文書（Obsidian 案件フォルダ）から判断。並列で着手」
> **実行形態**: 親オーケストレーター + 8 並列子エージェント（区画ごとに別 git worktree、claim ブランチで排他）
> **受入状態**: awaiting-human-verification（変更は main checkout の未コミット差分として配置）

## 正本の優先順位（本キャンペーンの判定基準）

1. クライアント確定ソース: `~/Vaults/CorpVault/50_Projects/顧客案件/ノア動物病院電子カルテ/`（`12_GitHub確定仕様_2026-08-20` / `13_UAT-R5確定仕様_2026-08-22` / `14_健診パッケージ_クライアント要件_2026-08-24` / `進行ボード/CURRENT.md` / `会話ログ/Slack_電子カルテ開発_曽我/`）と `…/ノア案件データ移行/`（マッピング台帳・投入スコープ）
2. 現行コード: `backend/` `frontend/` `infra/` `scripts/` `.github/workflows/` `backend/migrations/`
3. Plane（workspace `baritechllc` / Project `EMR` / hub `EMR-1`）= 実行状態 SoT（読み取りのみ、書き込みなし）

## 区画結果

| 区画 | レポート | branch | 変更ファイル数 | 主な訂正 |
|------|----------|--------|----------------|----------|
| architecture | [arch.md](arch.md) | `docs-refresh-20260929/arch` | 11 + レポート | erd テーブル数 128→130、package 数 36/14→37/15（support 追加）、auth キャッシュ resolver 配線済み訂正、Linear→Plane 8 箇所 |
| spec-core | [spec-core.md](spec-core.md) | `docs-refresh-20260929/spec-core` | 4 + レポート | ルート数 86→87（`/settings/bug-reports`）、E2E inventory 欠落 2 件、スキーマ 128→130、Linear→Plane |
| spec-detail | [spec-detail.md](spec-detail.md) | `docs-refresh-20260929/spec-detail` | 2 + レポート | lstep-tag-config はシステム管理者専用、LIFF 認証系レートリミット 60/分 |
| ops-deploy | [ops-deploy.md](ops-deploy.md) | `docs-refresh-20260929/ops-deploy` | 9 + レポート | Linear→Plane、CLINIC_CSV_IMPORT の billing_items 契約、STG 第4 cron（EMR-213）、coverage パス |
| ops-infra-testing | [ops-infra-testing.md](ops-infra-testing.md) | `docs-refresh-20260929/ops-infra-testing` | 14 + レポート | Container `sleepAfter` 10m→1h、e2e.yml `inputs.suite` 配線、シナリオ索引 S01-S39+V01-V05、S09 fixture 現状 |
| ops-scenarios | [ops-scenarios.md](ops-scenarios.md) | `docs-refresh-20260929/ops-scenarios` | 6 + レポート | S17 テスト件数 28→9、S19 sidebar 実体（useState+matchMedia）、S02 バッジ実体、Linear→Plane 5 箇所 |
| delivery-root | [delivery-root.md](delivery-root.md) | `docs-refresh-20260929/delivery-root` | 5 + レポート | 納品/索引の Linear SoT→Plane（EMR-39/40/41/131/134/150/151 明示）、GOLIVE の P 番号構造廃止 |
| work | [work.md](work.md) | `docs-refresh-20260929/work` | 45 + レポート | 台帳類の生きた状態記述に Plane-SoT 日付付き注記（EMR-84/85/89/180/211 Done 等）、`bug.md` 0byte 残存の訂正 |

合計: 96 文書訂正 + 8 レポート = 104 ファイル変更。全訂正は `(2026-09-29 訂正: …)` 形式で日付・根拠付き。履歴文書の rationale 改竄なし（dated-correction 原則）。

## 検証（統合後、main checkout で実施）

- `git diff --check -- docs/` → exit 0
- `bash scripts/check-docs-symbol-drift.sh` → **exit 0（ドリフトなし、検査トークン 609 件）**。前回・各子時点で残存していた「テーブル数 128 vs 実測 130」FAIL は `erd.md`（arch）と `specification.md`（spec-core）の訂正で解消
- docs 変更は runtime 検証不要（docs-only）
- 各子の `git status` allowlist 外差分 0 を確認済み。統合は `git diff main..<branch> | git apply` で全区画を直列適用

## BLOCKED / 外部依存（未解消・子が明示して残したもの）

- **Lステップ Write API 再開**: 外部 enable + STG live-send 実測待ち（`LSTEP_WRITE_API_ENABLED` 既定 OFF、Plane `EMR-132` Blocked）
- **健診パッケージ要件**: 臨床+PO 承認待ち（vault `14` は記録のみ）
- **STG 共有 DB 投入 / 本番 cutover**: 人間ゲート（formal COMPLETE bundle 未受領、八王子 `TBL_KNJO_DATA` 破損は継続 blocker）
- **U1–U12 / U13**: 未記入・未完（EMR-150/151、EMR-41/134）。完了扱いにしていない
- **E2E 実実行証跡**: clinical/v04・auth-smoke の Actions 実行証跡なし。scenario の runtime PASS は静的照合では記録していない
- **PERF-AUDIT-TX**: Plane work item 未起票 → UNKNOWN（再開時に起票要）
- **実機/外部ゲート**: LINE/LIFF 実機、検査機器実機（PU-4010/IDEXX）、S21 実並行、S23 PO 裁定、UAT-254 close 全外部ゲート — 未解消のまま明示

## 区画外の発見（今回の編集範囲外・引き継ぎ事項）

- `.github/workflows/backend-deploy.yml:126` に `sleepAfter 10m` の stale コメント（コード側・docs 範囲外）
- `frontend/src/features/manual/content/` 配下のアプリ内マニュアル（`23-master-medicine.md`/`08-inventory-management.md`）に在庫機能の誤記述（docs 範囲外）
- 各子レポートの「ゾーン外の発見」節に詳細

## claim / branch / worktree の状態

- 8 個の `claim/DOCS-REFRESH-20260929-*` ブランチと 8 個の `docs-refresh-20260929/*` ブランチ + `/private/tmp/ae-dr29-*` worktree が残存
- 人間の受入・統合判断後に claim 解放と branch/worktree 削除を行う（USER 判断。claim 削除は原則 USER のみ）
