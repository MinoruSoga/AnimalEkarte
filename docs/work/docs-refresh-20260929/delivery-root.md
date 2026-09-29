# DOCS-REFRESH-20260929 — delivery-root

対象ゾーン: `docs/delivery/**`, `docs/README.md`, `docs/product-philosophy.md`
作業 worktree: `/private/tmp/ae-dr29-delivery-root`（branch `docs-refresh-20260929/delivery-root`, base `main` @ `4722f4db7`）

## 対象と判定

| ファイル | 判定 | 主な根拠 |
|---|---|---|
| `docs/README.md` | corrected | 実行 SoT が「Linear」のまま 3 箇所。現在値の正本は Plane（`docs/work/plane-md-migration-20260923-receipt.md`、共有ブリーフ）。索引の 5 カテゴリ（architecture/spec/ops/delivery/work）とファイル構成は実ツリーと一致 |
| `docs/delivery/README.md` | corrected | 「実行状態は Linear」1 箇所 → Plane。受入条件セクション（#252/#254/#259）と U* 要約は現行と一致 |
| `docs/delivery/DELIVERY_PACKAGE.md` | corrected | 「Linear の現在値…未取得」→ Plane `EMR-39`/`EMR-150` への訂正ノート。在庫紐付けの前回訂正（薬=API のみ・UI 未実装、商品=フィールド非存在）は維持されコードとも一致 |
| `docs/delivery/GOLIVE_RUNBOOK.md` | corrected | (a) ヘッダの「Linear の現在値は未取得」→ Plane `EMR-40`、(b) `todo.md` P7→P8 前提の記述が陳腐化（P 番号構造は 2026-09-26 統合で廃止、`EMR-151`/`EMR-134` に移行）、(c) 「Linear の受容条件（todo.md P4）」→ Plane `EMR-131`。STAFF_ACCOUNT_PROVISIONING の前回訂正（受領履歴≠現行名簿）は維持 |
| `docs/delivery/OPERATION_MANUAL.md` | corrected | 「Linear BRT-47 の現在値は今回未取得」→ Plane `EMR-41`/`EMR-151` への訂正ノート。在庫紐付け§9・確定カルテ§12・ログイン 5req/min・パスワード規則は現行コードと一致 |
| `docs/delivery/trimming-history-copy-manual-text.md` | kept | `TrimmingRightColumn.tsx` / `use-trimming-history.ts` / `use-trimming-form-chrome.ts` / `manual/content/screens/10-trimming.md` と項目単位で一致（コピー対象・非対象・スタッフ確認ゲート・絞り込み仕様） |
| `docs/product-philosophy.md` | kept | 5 段階順序・「存在すべきでないものの最適化/自動化の禁止」・自動化の停止機構/監査は `.claude/CLAUDE.md` 圧縮版と `docs/spec/specification.md` §2.1（臨床安全優先・確定記録の不変性・追記訂正）と整合。事実・参照誤りなし |

## 実施した訂正（抜粋・重要度順）

- `docs/README.md:60`（訂正）: 「実行 SoT は Linear（hub BRT-4）・競合時の正本は Linear」→ Plane（workspace `baritechllc` / Project `EMR` / hub `EMR-1`、case_id `BRT-4` 不変。Linear は 2026-09-16 閉鎖）（根拠: `docs/work/plane-md-migration-20260923-receipt.md`）
- `docs/delivery/GOLIVE_RUNBOOK.md:47`（訂正）: 「現行 todo.md は P7→P8 を前提に含める・USER が Linear 上で確定」→ 現行 `todo.md` は P 番号構造を持たず、研修=Plane `EMR-151`・Go-live=Plane `EMR-134`、Linear は 2026-09-16 閉鎖。順序確定は USER が Plane 上で行う未解決差異として残す（根拠: 現行 `todo.md`、`docs/work/plane-ready-triage-20260927/VERDICTS.md:85,39`）
- `docs/delivery/GOLIVE_RUNBOOK.md:49`（訂正）: 「Linear の受容条件（`todo.md` P4）」→ Plane `EMR-131`（旧 `P4` / #254）（根拠: `VERDICTS.md:86`）
- `docs/delivery/GOLIVE_RUNBOOK.md:5`（訂正）: ヘッダの「Linear の現在値は未取得」→ Linear 閉鎖・正本は Plane `EMR-40`（根拠: 同上、`VERDICTS.md:96` で EMR-40=keep_blocked 確認）
- `docs/delivery/DELIVERY_PACKAGE.md:227`（訂正）: 「Linear の現在値」→ Plane `EMR-39`・U1–U12 入力待ち `EMR-150`（2026-09-27 時点も全行未記入=keep_needs_human）（根拠: `VERDICTS.md:40,71`）
- `docs/delivery/OPERATION_MANUAL.md:240`（訂正）: 「Linear BRT-47 の現在値は今回未取得」→ Plane `EMR-41`・U13 入力待ち `EMR-151`（2026-09-27 時点も未実施=keep_needs_human）（根拠: `VERDICTS.md:39,70`）
- `docs/delivery/README.md:35`（訂正）: 「実行状態は Linear」→ Plane
- `docs/README.md:49,65`（訂正）: カテゴリ表・フッタの SoT 記述を Plane へ。最新更新を 2026-09-29 に更新

## 維持確認（前回訂正）

- **在庫紐付け**: `DELIVERY_PACKAGE.md` Step5 / `OPERATION_MANUAL.md` §9 の「薬=API で `inventory_id` 受領可・設定 UI 未実装」「商品=在庫紐付けフィールド非存在」「在庫減算は明示 `InventoryID` 指定時のみ」を現行コードで再確認: `backend/internal/model/medicine.go`（`InventoryID *uint64`）/ `merchandise_item.go`（在庫フィールドなし）/ `medicalrecord/treatment_service_tx.go`（`InventoryID` 明示時のみ減算）/ `spec/screens/settings/master-medicine.md`・`master-merchandise.md` / `frontend/src/features/master/components/`（在庫 UI 参照なし）
- **スタッフ名簿受領履歴**: `GOLIVE_RUNBOOK.md` 前提#6 → `docs/ops/deploy/STAFF_ACCOUNT_PROVISIONING.md` の「受領履歴≠現行適用名簿・provisioning 完了でない」記述を維持。外部入力（現行名簿・clinic 対応・ロール・メール方針・環境・実行者権限・receipt）は未解消のまま

## 外部依存・BLOCKED（未解消のまま残すもの）

- **U1–U12**（DELIVERY_PACKAGE USER 入力待ち表）: 契約・本番 provider 事実・LINE/Lステップ秘密・通知窓口・バックアップ証跡・R2 方針・監査保持・本番構築 receipt — 全行未記入のまま（Plane `EMR-150` = keep_needs_human、VERDICTS.md:40）
- **U13**（操作説明会）: 日程・形式・対象・実施 receipt・別 close 承認 — 未完のまま（`EMR-151`/`EMR-41` = keep_needs_human、VERDICTS.md:39,70）
- **Go-live 実行**: 全 pre-window prerequisite + USER の新 window 記入まで HOLD（`EMR-40`/`EMR-134` = keep_blocked、VERDICTS.md:85,96）
- **本番運用状態**: `.github/workflows/backend-deploy.yml` に `deploy-production` + `environment: Production` + production branch/dispatch ゲートは存在するが、provider・secret・環境・実行 receipt は未取得。production 稼働済みとは記述しない（`docs/ops/deploy/README.md` の draft/UNKNOWN 方針を維持）
- **バックアップ**: `docs/ops/infra/production/runbook.md` §4 — 人間承認済み方法と証跡が揃うまで HOLD を維持

## 提案削除候補（理由・後継リンク）

なし（全対象ファイルは現行参照される正本/補助資料として存続が妥当）

## ゾーン外の発見（当該ゾーン担当へ引き継ぎ）

- `docs/work/README.md`: 「実行 SoT は Linear」記述が複数箇所残留（allowlist 外のため未編集）
- `docs/work/plane-ready-triage-20260927/VERDICTS.md`（EMR-134 周辺）: 「backend-deploy.yml に本番 trigger/environment ゲートなし」との記述は実ファイルと矛盾（`deploy-production` job + `environment: Production` + production branch/dispatch 条件を確認）
- `docs/spec/specification.md:22,61`・`docs/architecture/erd.md:157,161,332`: 宣言テーブル数 128 が実測 130 とドリフト（erd.md:3 は既に 130 表記で内部不整合もあり）
- `frontend/src/features/manual/content/screens/23-master-medicine.md`・`workflows/08-inventory-management.md`: アプリ内マニュアルが薬の在庫紐付け UI・商品の自動減算を実装済みのように記述。spec/コードと矛盾（frontend ゾーンまたは製品修正案件）

## 検証

| コマンド | 結果 |
|---|---|
| `git diff --check -- docs/delivery docs/README.md docs/product-philosophy.md` | exit 0（warning なし） |
| `bash scripts/check-docs-symbol-drift.sh` | exit 1 — FAIL 3 件は全て同一の「テーブル数: 宣言 128 / 実測 130」。宣言元は `specification.md`・`architecture/erd.md`（allowlist 外）で、base commit `4722f4db7` 時点からの既存ドリフト。本 diff はテーブル数記述に無関係 |
| ローカルリンク検査（7 ファイル） | 181 リンク走査、missing 0 |
| `git status --porcelain` | `M` 5 ファイル（`docs/README.md`, `docs/delivery/README.md`, `DELIVERY_PACKAGE.md`, `GOLIVE_RUNBOOK.md`, `OPERATION_MANUAL.md`）+ 本レポート新規。allowlist 外差分なし |
| コミット | `f1269a436 docs: delivery/README/索引の Linear SoT 記述を Plane へ日付付き訂正`（5 files, +9/-9） |
