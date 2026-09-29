# DOCS-REFRESH-20260929 — work

対象: `docs/work/**`（`docs-perfection/**` は read-only 除外、本ディレクトリは本票のみ）。作業 tree `/private/tmp/ae-dr29-work`、branch `docs-refresh-20260929/work`、base `main` (`4722f4db7`)。

目的: 作業台帳・キャンペーン記録・調査証跡で「生きた SoT/状態」を装う陳腐化記述（特に Linear 現行扱い）へ日付付き注記を追加し、履歴本文は改竄しない。Linear は 2026-09-16 閉鎖・履歴参照のみ。現在の実行 SoT は Plane workspace `baritechllc` / Project `EMR`（hub `EMR-1`、case_id 旧 `BRT-4`）。

## 対象と判定

| 対象 | 判定 |
|---|---|
| `docs/work/README.md` | **更新** — 実行 SoT を Linear→Plane へ訂正、削除 zone の `bug.md` 記述を現行状態へ訂正、索引を現行ファイル集合へ同期 |
| `decisions/README.md` | **更新** — 正本 Linear→Plane、起票先 Linear→Plane（日付付き訂正） |
| `decisions/fable-po-recommendation.md` | **更新** — 冒頭に「2026-08-06 採択時点の記録」注記。本文不変更 |
| `phase2-deferred.md` | **更新** — 実行 SoT・再開手順・mermaid・legacy 再提案文を Plane へ。`PERF-AUDIT-TX P2` の Plane 状態を **UNKNOWN / 未起票** と記録 |
| `linear-f1-f6-mapping.md` | **更新** — 冒頭注記（2026-09-11 Linear 読取記録・F1–F6 対応は未特定のまま・「USER が Linear で行う手順」節は履歴） |
| `development-task-decisions.md` | **更新** — 冒頭注記（2026-09-11 時点の判断記録・`todo-issue.md` 参照は現行 `todo.md#issue-ledger` へ対応） |
| `stg-uat-clinic-feedback-q1-q4.md` | **更新** — 冒頭注記（状態列は 2026-09-15–17 時点・現行状態は Plane・主要 ID を 2026-09-29 読取結果で明記） |
| `emr-89/180/211` 証跡票 3 件 | **更新** — 冒頭注記（調査時点記録・Plane 2026-09-29 読取: いずれも Done） |
| `todo-campaign-20260918/` 7 件 | **更新** — 冒頭注記で Plane 正本 ID を明記（EMR-84/85 Done、MIG-15 Ready、EMR-106 Ready、EMR-107 Backlog、EMR-108、EMR-135 Cancelled）。`UAT-R2-CHART-FIT` は既存 Plane 注記済みで不変更 |
| `todo-campaign-20260919-ready17/` 9 件 | **更新** — 冒頭注記（EMR-86/87/101〜105/116/117）。CAMERA/MICROCHIP/OWNER-HEIGHT/VITALS は既存 Plane 注記済みで不変更 |
| `remaining-campaign-20260920/` 14 件 | **更新** — P6/P7 は既存注記済み。SLACK-* 13 件へ冒頭注記（EMR-88/90/92/94/97/99/109/110/111/113/114、MIG-17/18）。`SLACK-UAT-SCHEDULE` の誤った `EMR-200` 参照を `EMR-112` へ日付付き訂正 |
| `todo-ledger-20260922/` 5 件 | **更新** — 4 件へ冒頭注記（EMR-85 Done、EMR-125、MIG-16 Needs Human、MIG-19）。`DEV-V-OWNER-DB-RECEIPT` は既に Plane `EMR-125` 明記で不変更 |
| `linmig-campaign-20260919/` 12 件 | **不変更** — 全件が既に「Current task status is tracked in Plane `EMR-*` / migration receipt 参照」の冒頭注記を持つ |
| `emr-200-note2-coverage/NOTE2-SWEEP-COVERAGE.md` | **不変更** — 既に Plane `EMR-200` 注記済み |
| `plane-md-migration-20260923-receipt.md` | **不変更** — 移行 receipt 自体が crosswalk 正本。陳腐化記述なし |
| `plane-ready-triage-20260927/` | **不変更** — 2026-09-27 時点 verdict の日付付き記録（ready=0）。Linear 現行扱いなし |
| `skill-reeval-2026-09-06.md` | **不変更** — 2026-09-06 時点の資料選択記録。SoT/状態の陳腐化なし |
| `docs-perfection/**` | **不変更（read-only）** |

## 主要訂正（旧主張 → 新主張）

| 箇所 | 旧主張 | 新主張 | 根拠 |
|---|---|---|---|
| `README.md` 入口・競合ルール | Linear を実行 SoT・競合時の正 | Plane `EMR`（hub `EMR-1`）。Linear は 2026-09-16 閉鎖・履歴のみ | Plane live readback・migration receipt |
| `README.md` 削除 zone `bug.md` 行 | 「`bug.md`・`todo-now.md`…（2026-09-08）統合削除」のまま | 同様の禁止は維持しつつ、`bug.md` は Plane 移行期に再作成→2026-09-27 台帳整理で 0 バイト残存する経緯を日付付きで追記 | `git log -- bug.md`（`1b4b421bb` ほか）で現行ファイル実在を確認 |
| `decisions/README.md` | 正本 Linear・起票先 Linear | Plane `EMR`・起票先 Plane | 同上 |
| `phase2-deferred.md` | 実行 SoT Linear・再開は Linear Issue 作成 | Plane `EMR`・再開は Plane work item 作成。P2 の Plane 状態 UNKNOWN/未起票 | Plane `PERF-AUDIT-TX` 検索で該当 item なし（2026-09-29） |
| `stg-uat-clinic-feedback-q1-q4.md` | 状態列が現行に見える | 2026-09-15–17 時点記録と明記。EMR-84/85/120/121/122/123/124/181=Done、MIG-16=Needs Human、MIG-15=Ready、EMR-106=Ready、EMR-107=Backlog（2026-09-29 Plane 読取） | Plane live readback |
| `remaining-campaign-20260920/SLACK-UAT-SCHEDULE.md` | `Current task state: Plane EMR-200` | `EMR-200` は `NOTE2-SWEEP-COVERAGE` の ID で誤り → `EMR-112` へ訂正 | migration receipt crosswalk（SLACK-UAT-SCHEDULE → EMR-112）と Plane `EMR-200` 実体の照合 |
| `TODO-V-LINEAR.md` | Linear 読取照合の作業票として現行に見える | 2026-09-18 時点記録。Plane `EMR-135` = Cancelled（Linear 閉鎖で meta 照合終了） | Plane readback 2026-09-29 |

## 未解決の外部依存 / BLOCKED

- **`PERF-AUDIT-TX P2` の Plane 起票**: 該当 work item が見つからず UNKNOWN / 未起票。再開時に Plane 新規起票が必要（docs 側で代行しない）。
- **Plane 起票・状態遷移**: 本キャンペーンは Plane/Linear/GitHub/DB/STG への書込を一切行っていない。各票の Plane ID は readback と receipt crosswalk に基づく参照のみ。
- **`docs-symbol-drift` のテーブル数宣言値 128 vs 実測 130（3 件）**: base commit `4722f4db7` でも同一 FAIL を確認した既存ドリフト。宣言箇所は `docs/spec/`・`docs/architecture/` 側で本票の allowlist（`docs/work/**`）外のため **BLOCKED — 別 unit の対応が必要**。
- **45 キー → Plane ID 対応表（SLACK-INTAKE 内）**: 外部 Q&A シートの現行状態は repo から検証不能のまま UNKNOWN（原文記述を維持）。

## 削除候補

なし。本キャンペーンではファイル削除を実施・提案しない。README の削除 zone は復活防止の履歴台帳として維持し、現行ファイル集合と矛盾する記述（`bug.md`）のみ日付付き訂正した。

## 検証

| コマンド | 結果 |
|---|---|
| `git diff --check -- docs/work` | exit 0（空白・競合マーカーなし） |
| `bash scripts/check-docs-symbol-drift.sh` | exit 1 — 既存ドリフト（テーブル数 128/130 ×3）。**stash して base で再実行しても同一 FAIL** のため本変更由来ではない（既存 NG / allowlist 外） |
| `git status --porcelain` | 変更は `docs/work/` 配下 45 ファイル + 本票のみ。allowlist 外 diff なし、`docs-perfection/**`・`docs-refresh-20260929/**`（本票を除く）への変更なし |
| アプリ runtime / `go test` / `pnpm lint` / `pnpm build` / `make codegen` | SKIP（docs-only 変更のため不要・heavyweight 禁止に従い未実施） |

## その他

- 全訂正は `（2026-09-29 訂正: 旧主張 → 新主張）` または冒頭 `2026-09-29 注記` 形式で、本文の履歴記述・判断根拠・行番号参照は改竄していない。
- 秘密・患者・飼主・連絡先・実医院データの新規記述なし。Slack 原文の転記なし。
