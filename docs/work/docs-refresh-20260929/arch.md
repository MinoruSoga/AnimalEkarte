# docs-refresh-20260929 — arch partition report

- 担当: `arch` partition（`docs/architecture/**`）
- Worktree: `/private/tmp/ae-dr29-arch` / branch `docs-refresh-20260929/arch`（base `4722f4db7`）
- 実施日: 2026-09-29
- Doc commit: `e929d91ec` `docs(architecture): 2026-09-29 実装照合 — Linear→Plane 訂正・ERD 130 表・support domain 追加反映`

## 1. 対象ファイルと判定

| ファイル | 判定 | 内容 |
|:---|:---|:---|
| `docs/architecture/overview.md` | corrected | release gate 入口の Linear hub BRT-4 → Plane `EMR`/`EMR-1` へ日付付き訂正（2026-09-29） |
| `docs/architecture/data-flow.md` | corrected | 「再開条件の正本は GitHub #259 後続コメント」→ 実行状態の正本は Plane `EMR`（Issue は履歴証跡として保持）へ日付付き訂正。最新更新 → 2026-09-29 |
| `docs/architecture/auth.md` | corrected | (a) `NewCachedCurrentAccessResolver` を挟まない → env gate（`CURRENT_ACCESS_CACHE_TTL_SEC`>0）付き配線へ訂正。(b) 「`current_access_cache.go` は最終認可の入力に使わない」→ env gate 有効時のみ使用へ訂正。(c) GET/HEAD 台帳件数 203/201 → 207/205 へ訂正。v9.2 → v9.3（最新更新 2026-09-29） |
| `docs/architecture/erd.md` | corrected | 物理テーブル総数 128 → 130（`011_support_bug_reports` + `013_support_chat_messages`）。active inventory を `001`〜`013` の 13 本へ更新（`011` は 2 ファイル・`008` 欠番）。§4.4 に `005`/`006`/`007`/`009`/`010` の行を追加。すべて dated correction で履歴は保持 |
| `docs/architecture/be9-2a-boundary-map.md` | corrected | Linear hub BRT-4 → Plane（2 箇所）。`acceptedTopLevelPackages` 36/14 → 37/15 へ日付付き訂正 |
| `docs/architecture/exception-package-discipline.md` | corrected | 「Not a 15th domain」×2 → `support` が 15 番目の domain のため「domain でない」へ読み替え訂正。pin 値 36/14 → 37/15 |
| `docs/architecture/arch-a4-trigger-ledger.md` | corrected | 「Linear/BRT-4 is the execution SoT」→ Plane `EMR`/`EMR-1` へ日付付き訂正 |
| `docs/architecture/adr/006-backend-domain-package-boundaries.md` | corrected | (a) release gate の Linear 参照 → Plane。(b) `support/` を 15 番目の target domain として 2026-09-29 amendment 追加（許可 edge は `support → httpapi` のみ）。(c) 現行補足の pin 値 35/14 → 37/15 へ 2026-09-29 追補 |
| `docs/architecture/adr/007-lab-device-receive-and-commit.md` | corrected | `Relates to: Linear BRT-100/BRT-94` → Plane `EMR` / case `BRT-4`（BRT-n は履歴別名として保持）。採用時案の「Linear に出さない」規則を「issue tracker」へ一般化し日付付き訂正 |
| `docs/architecture/model-write-owner-catalog.md` | updated | `SupportBugReport` / `SupportChatMessage` → `support` の行を追加（`clinic_id`×`staff_id` スコープ、管理操作は `hospital-settings` 権限を流用） |
| `docs/architecture/fe-feature-be-domain-map.md` | updated | `support` feature → `support` domain の行を追加（認証済みスタッフ全員、一覧・更新のみ `hospital-settings`） |
| `docs/architecture/README.md` | kept | 索引・gate 注記ともに現行と整合 |
| `docs/architecture/composition-root-conventions.md` | kept | pin 対象の composition file 群（runtime/auth/staff/clinic/owner_pet/reservation/billing/medicalrecord）実在を確認。medicalrecord 構成 ~600 LOC（実測 623）も記述と整合 |
| `docs/architecture/cross-domain-orchestration-catalog.md` | kept | `support` は cross-domain write を持たないため行追加不要。既存 PATH-* の主張は実装と整合（`Complete`/`result.Created`、`SyncCPMStageTag`、`X-Scheduler-Token`/`_internal/scheduled-jobs` 等を実測確認） |
| `docs/architecture/delete-soft-delete-patterns.md` | kept | `DeleteSoftDeletedByClinicID`（permission_group_repository.go:306）・`ensureClinicCanBeDeleted`（clinic_service.go:407）・`persistence.DeleteScopedByID`（scope.go:112）を実在確認 |
| `docs/architecture/adr/README.md` | kept | ADR 索引・supersede 関係は現行と一致 |
| `docs/architecture/adr/001`〜`005`, `adr/008` | kept | Status 表記は実装状態と整合。ADR-008（user LaunchAgent・consumer token・PIMS opt-in・キュー上限）は `cmd/lab-device-agent` / `internal/labdeviceagent` / `docs/ops/deploy/LAB_DEVICE_*` と整合 |
| `docs/architecture/be9-2a-classification-manifest.csv` | kept | immutable snapshot（761 row、変更しない運用を各所が明記） |
| `docs/architecture/*.test.mjs`（4 本） | kept | marker 契約テスト。全 PASS |

`propose-delete`: なし。

## 2. 重要訂正（優先度順）

1. **Linear → Plane 実行 SoT**（6 ファイル・8 箇所）: Linear は 2026-09-16 閉鎖。実行 SoT = Plane workspace `baritechllc` / project `EMR` / hub `EMR-1` / case `BRT-4`。`BRT-n` / `LINMIG-n` は履歴別名のみ。対象: `overview.md`, `data-flow.md`, `be9-2a-boundary-map.md`（×2）, `arch-a4-trigger-ledger.md`, `adr/006`, `adr/007`（×2）。いずれも `（2026-09-29 訂正: …）` 形式で旧主張を保持。
2. **ERD テーブル総数 128 → 130**: `001_init.sql` は 128 `CREATE TABLE` のままだが、`011_support_bug_reports.sql` / `013_support_chat_messages.sql` で +2。§4 判定表・§4.3 inventory・§4.4 を訂正し、`005`（子レコード version CAS）・`006`（accounts RLS ops bypass）・`007`（billings 複合部分 index、EMR-201）・`009`/`010`（seed bundle checksum reconcile、EMR-213/EMR-176）の行を補完。
3. **package boundary pin 36/14 → 37/15**: `backend/internal/lintscan/package_boundary_gate_test.go` が 37 top-level / 15 domain を pin（`requireSetSize`、行 111-112）。`support` は 2026-09-26 に追加された 15 番目の domain（`domainImportAllowlist` の許可 edge は `support → httpapi` のみ、domain_import_allowlist_lint_test.go:58）。
4. **auth.md の current-access cache 記述**: 「挟みません」「最終認可の入力には使いません」→ `composition_auth.go:159-184` の `currentAccessCacheTTL()`（`CURRENT_ACCESS_CACHE_TTL_SEC`>0 のみ有効。2026-09-22 `d96c4ba27` で追加・STG 合成データ向け）へ訂正。cache 有効時は権限/所属無効化の反映が最大 TTL 遅延することを明記。contract test `composition_auth_cache_contract_test.go` が無条件配線を静的に拒否。
5. **auth.md §4.7 GET/HEAD 台帳件数 203/201 → 207/205**: `get_head_permissions.json` の実測 207 行（`uploads` 2 行を除くと S3 モード 205）。support chat 追加（2026-09-28 `86fcec3d6`）による増加。
6. **write-owner / feature map に `support` 追加**: `model-write-owner-catalog.md`（`SupportBugReport`/`SupportChatMessage` → `support`）、`fe-feature-be-domain-map.md`（`features/support` → `support`、admin 操作は `hospital-settings`）。

## 3. 証拠（path / line）

| 主張 | 証拠 |
|:---|:---|
| 内部 package 数 37 / domain 15 | `backend/internal/lintscan/package_boundary_gate_test.go` 行 16, 56, 111-112（`requireSetSize` で 37/15 pin） |
| `support` の許可依存 | `domain_import_allowlist_lint_test.go:58` `"support": {"httpapi": {}}`; `backend/internal/support/*.go` の production import は `apperrors`/`httpapi`/`model` のみ |
| support route/権限 | `backend/internal/support/handler.go:54-77`（chat・bug-report 作成は認証済み全員、一覧/更新は `hospital-settings` view/edit、chat POST は rate limit） |
| current-access cache env gate | `backend/cmd/api/composition_auth.go:159-184`（`CURRENT_ACCESS_CACHE_TTL_SEC` 正の整数のみ有効、既定無効）、`composition_auth_cache_contract_test.go:23-48` |
| GET/HEAD 台帳 207 行 | `backend/cmd/api/testdata/get_head_permissions.json`（207 entries・`uploads` 2 行 → S3 205） |
| migration inventory 13 本 / 130 表 | `ls backend/migrations/*.sql`（001〜013・008 欠番・011 は 2 ファイル）、`001_init.sql` = 128 CREATE TABLE + support 2 表 |
| migration 内容 | `005`（version 列 CAS 4 表）、`006`（`tenant_accounts_isolation` の `bypass_rls` 再定義）、`007`（`idx_billings_clinic_owner_scheduled`）、`009`/`010`（checksum reconcile コメントに経緯）、`011×2`/`012`/`013` |
| seedlogin 許可 env | `backend/internal/seedlogin/env.go:24`（development/local/dev/test/staging）— auth.md §4.4 と一致 |
| `Complete`/`result.Created`/`SyncCPMStageTag` | `backend/internal/billing/accounting_complete.go:287`、`accounting_service_core.go:568`、`service_deps.go:94` |
| scheduler boundary | `backend/cmd/api/batch_scheduler.go:19`（`X-Scheduler-Token`）、`backend/internal/scheduler/handler.go:121`（`/_internal/scheduled-jobs/:jobAction`） |
| FE features 29 dir | `frontend/src/features/`（`support` を含む 29 機能 + CLAUDE.md）。map が全 dir を網羅 |
| Linear 閉鎖・Plane 移行 | `/tmp/ae-docs-refresh-20260929-briefs/SHARED.md` / `BRIEF-arch.md`（workspace `baritechllc`、project `EMR`、hub `EMR-1`、case `BRT-4`） |

## 4. BLOCKED / 外部依存 / UNKNOWN

| 項目 | 状態 | 詳細 |
|:---|:---|:---|
| `scripts/check-docs-symbol-drift.sh` の FAIL×3 | **BLOCKED（allowlist 外）** | 原因は `docs/spec/specification.md` 行 22（「全 128 テーブル」）と行 61（「スキーマは 128 テーブル」）。行 22 がスクリプトの検査対象集合に重複して数えられ 3 件と表示される。実測 130 と不一致。`docs/spec/**` は arch partition の allowlist 外のため本 partition では修正不可。`spec` 担当 partition への引き継ぎ要 |
| Plane の実際の実行状態 | **UNKNOWN** | Plane は読み取り専用ルール。本文書では SoT 参照のみを記述し、個別 work item の存在・状態は未検証（例: ARCH-A4 の active issue なし、は ledger の記述どおり「no active issue verified」） |
| migration の稼働 DB 適用状態 | **UNKNOWN / 外部依存** | repo DDL の静的照合のみ。STG/PROD への適用・checksum 状態は本作業の検証対象外（erd.md の既存注記どおり） |
| `009`/`010` checksum reconcile の STG 効果 | **外部依存** | migration コメントに経緯のみ確認（2026-09-24 STG crash loop 修復）。実適用結果は未検証 |
| lab-device 実機 UAT / 医院 rollout | **外部依存（human gate）** | ADR-007/008 の「code complete・UAT/release gate 未了」記述は維持。実機・医院展開の完了証跡なし |
| LSTEP 運用再開（#259 関連） | **外部依存** | コード配線は存在するが先方 enable・STG 実配信・cron 実測は別途必要（data-flow.md の記述を維持しつつ SoT 参照先のみ訂正） |

## 5. 削除候補・後継リンク

- **削除候補なし**。`arch-a4-trigger-ledger.md` と `be9-2a-boundary-map.md` は historical record として保持（各ファイル冒頭の status 注記どおり。後継の現行裁定は `adr/006`・`internal/lintscan` gate）。
- `docs/spec/specification.md` の 128 表記 2 箇所 → 後継修正は spec partition（`docs/spec/**` の担当）へ。

## 6. 検証コマンド結果

| コマンド | 終了コード | 結果 |
|:---|:---:|:---|
| `git diff --check -- docs/architecture` | 0 | clean（whitespace error なし） |
| `bash scripts/check-docs-symbol-drift.sh` | **1** | FAIL×3「テーブル数: docs の宣言値 128 が実装の実測値 130 と不一致」— すべて `docs/spec/specification.md`（行 22・61）由来。`docs/architecture` 内の token（605 件検査）は ERD 130 宣言を含め不一致なし。**allowlist 外のため本 partition では未解決** |
| `node docs/architecture/model-write-owner-catalog.test.mjs` | 0 | pass |
| `node docs/architecture/fe-feature-be-domain-map.test.mjs` | 0 | pass |
| `node docs/architecture/cross-domain-orchestration-catalog.test.mjs` | 0 | pass |
| `node docs/architecture/cross-domain-orchestration-catalog-automation.test.mjs` | 0 | pass |
| `git status --porcelain` | — | 変更は `docs/architecture/` 11 ファイル + 本レポートのみ |

注記: catalog test は marker 契約の確認であり実装の完全正当性を証明しない。symbol-drift は narrative/ADR status を検証しない。静的照合であり STG/PROD・runtime・migration 適用・Plane 実行状態の証明ではない。

## 7. allowlist 遵守確認

変更ファイルは `docs/architecture/` 配下 11 件（commit `e929d91ec`）+ 本レポートのみ。`docs/spec`、`docs/ops`、`backend/`、`frontend/`、`migrations/`、`scripts/`、他 partition の `docs/work/` ファイルへの差分なし。main repo（`/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte`）への変更なし。
