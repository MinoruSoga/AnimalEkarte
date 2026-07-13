# BE Todo — バックエンド残タスク台帳

- **更新日**: 2026-07-13
- **本書の規約**: 今期着手可能な PERF/SEED 系の未対応のみを記載する。対応済みは残さない。詳細・手順の正本は git 履歴と `docs/tasks/closed/`。
- **別台帳**:
  - リファクタ系の今期着手可能残: `BE-refactor.md`
  - 次期送り・着手保留・任意検証: `BE-pending.md`
  - 本書と重複させない。

### 検証コマンド規約（Docker 必須・スコープ限定）

- 必ず `docker compose exec backend go test ./internal/<pkg>/ -run <Name> -count=1`。**フル `go test ./...`・`golangci-lint run ./...`・`gofmt -w ./...` は実行禁止**。
- 変更した Go ファイルは `docker compose exec backend gofmt -l ./internal/<dir>/` 無出力を確認してからコミット。
- `Co-Authored-By` なし。**push しない**（依頼があるまで）。

---

## 残タスク一覧

**エージェント実装可能な残タスクなし（2026-07-13 時点）。**

### 見送り（再開条件付き・今期着手しない）

| ID | 内容 | 再開条件 |
|----|------|----------|
| PERF-AUDIT-TX P2（outbox） | outbox パターン移行 | `audit_write_failed` が恒常的（目安: 月 1 件以上継続）に観測された場合、実測頻度 1 か月分を添えて再起案。正本: `docs/tasks/closed/perf/PERF-AUDIT-TX-UNIVERSAL-BEST-EFFORT.md` |

**スコープ外**: `docs/tasks/open/FEAT-searchable-select-targets.md` は FE 案件のため本台帳に含めない。

---

## 既知バグの skip テスト台帳（2026-07-13 第2次棚卸し）— 解消済み・0件

skip されたテストはバグが直っても・悪化しても fail しない。修正時は該当行の `t.Skip` を外して green を確認すること。件数は `rg -in "known production bug" backend/internal --glob '*.go'` の t.Skip 実測 11 件（+ 補足 2 件）。**旧台帳（小文字 `known production bug` 表記 9 件）は `ef4f8cf5` で根因修正・`5be86907` でクローズ済み**。本表はその後の #212 カバレッジ拡充（`e908f105`、2026-07-13）で新規発覚した未修正分であり、修正タスクは #236 に起票済み。

**追記（2026-07-13 #236 skip 解除・検証完了、commit bb2ad499）**: 下表の 11 件 + 補足 2 件、計 13 件すべて `t.Skip` を解除し green を実測確認した。実装側の修正（BUG#1: `reservation_staff_repository.go` Delete の EXISTS 相関サブクエリ化／BUG#3: `model/clinic_settings.go` への `gorm:"column:"`・`type:time` タグ付与／BUG#4: `medical_record_owner_visit_repository.go` の `apperrors.FromGORM` ラップ→`IsNotFound` 判定順序修正）はいずれも前回セッションで適用済みだったため、本セッションでは対応するテストファイルの `t.Skip` 呼び出しと陳腐化した「KNOWN PRODUCTION BUG」コメントブロックの除去のみを行った。補足の 2 件（`clinic_repository_test.go`）は BUG#3 と同一根因（`ClinicSettings` の AutoMigrate schema drift）であることを確認しており、BUG#3 の修正が適用済みだったことで副次的に解消していた（本セッションで実測発見・skip 解除）。検証コマンド: `docker compose exec backend go test ./internal/repository/ -run 'TestClinicRepository_CountBlockingReferencesByClinicID|TestReservationStaffRepository_Delete|TestClinicSettingsRepository_|TestMedicalRecordRepository_FindLatestByOwner' -count=1 -v` → 全 PASS。`rg -in "known production bug|known bug" backend/internal --glob '*.go'` → 0 件。`docker compose exec backend gofmt -l ./internal/model/ ./internal/repository/` → 無出力。上記修正・skip解除・本追記はすべて `bb2ad499`（fix(backend): クロステナントDelete・ClinicSettings列名不一致・FindLatestByOwner判定順序を修正 (#236)）としてコミット済み。push は未実施（本タスクの範囲外）。

### `t.Skip("KNOWN PRODUCTION BUG ...")` — 11 件 → 解消済み(2026-07-13, commit bb2ad499)

| ファイル:行 | テスト名 | skip メッセージ要約 | 状態 |
|---|---|---|---|
| ~~`internal/repository/reservation_staff_repository_test.go:166`~~ | `TestReservationStaffRepository_Delete` / `別クリニックからの削除はNotFoundで行が残る` | **CRITICAL**: Delete の `Joins()+Delete()` が GORM 上で機能せずクロステナント削除が成立する | 解消済み(2026-07-13, commit bb2ad499) — Delete を EXISTS 相関サブクエリ化。green 実測済み |
| ~~`internal/repository/medical_record_owner_visit_repository_test.go:205`~~ | `TestMedicalRecordRepository_FindLatestByOwner` / `clinic_id隔離: 別クリニックのカルテは対象外` | `apperrors.IsNotFound` が生 gorm エラーに対し常に false | 解消済み(2026-07-13, commit bb2ad499) — `FromGORM` ラップ後に `IsNotFound` 判定する順序に修正済み。green 実測済み |
| ~~`internal/repository/medical_record_owner_visit_repository_test.go:221`~~ | `TestMedicalRecordRepository_FindLatestByOwner` / `該当なし: カルテが存在しない飼い主は nil, nil` | 同上 | 解消済み(2026-07-13, commit bb2ad499) — 同上。green 実測済み |
| ~~`internal/repository/clinic_settings_repository_test.go:109`~~ | `TestClinicSettingsRepository_FindByClinicID` / `clinic_id隔離: 他院の設定は返らず自院はデフォルト値のまま` | **CRITICAL**: `repo.Save` が列名不一致（SQLSTATE 42703）で失敗 | 解消済み(2026-07-13, commit bb2ad499) — `model/clinic_settings.go` 全フィールドに `gorm:"column:"` タグ付与済み。green 実測済み |
| ~~`internal/repository/clinic_settings_repository_test.go:127`~~ | `TestClinicSettingsRepository_FindByClinicID` / `行が存在すれば実際の値を返す` | 同上 | 解消済み(2026-07-13, commit bb2ad499) — 同上。green 実測済み |
| ~~`internal/repository/clinic_settings_repository_test.go:141`~~ | `TestClinicSettingsRepository_Save`（親テストごと skip） | 同上 | 解消済み(2026-07-13, commit bb2ad499) — 同上。green 実測済み |
| ~~`internal/repository/clinic_settings_repository_test.go:172`~~ | `TestClinicSettingsRepository_UpdateCPMVersion`（親テストごと skip） | **CRITICAL**: 内部 UPSERT が列名不一致（SQLSTATE 42703）で失敗 | 解消済み(2026-07-13, commit bb2ad499) — 同上。green 実測済み |
| ~~`internal/repository/clinic_settings_repository_test.go:203`~~ | `TestClinicSettingsRepository_UpdateDormantThresholds`（親テストごと skip） | 同上 | 解消済み(2026-07-13, commit bb2ad499) — 同上。green 実測済み |
| ~~`internal/repository/clinic_settings_repository_test.go:239`~~ | `TestClinicSettingsRepository_UpdateCPMV2Thresholds`（親テストごと skip） | 同上 | 解消済み(2026-07-13, commit bb2ad499) — 同上。green 実測済み |
| ~~`internal/repository/clinic_settings_repository_test.go:260`~~ | `TestClinicSettingsRepository_UpdateCPMV1Thresholds`（親テストごと skip） | 同上 | 解消済み(2026-07-13, commit bb2ad499) — 同上。green 実測済み |
| ~~`internal/repository/clinic_settings_repository_test.go:286`~~ | `TestClinicSettingsRepository_UpdateHealthPreventionThresholds`（親テストごと skip） | 同上 | 解消済み(2026-07-13, commit bb2ad499) — 同上。green 実測済み |

### 補足: `t.Skip("known bug ...")` — 2 件（テスト基盤起因、production バグではない）→ 解消済み(2026-07-13, commit bb2ad499)

| ファイル:行 | テスト名 | skip メッセージ要約 | 状態 |
|---|---|---|---|
| ~~`internal/repository/clinic_repository_test.go:284`~~ | `TestClinicRepository_CountBlockingReferencesByClinicID`（親テストごと skip） | `model.ClinicSettings` の AutoMigrate schema drift がテーブル依存チェックを阻害 | 解消済み(2026-07-13, commit bb2ad499) — BUG#3 修正（gorm column/type タグ付与）と同一根因のため副次的に解消。green 実測済み |
| ~~`internal/repository/clinic_repository_test.go:337`~~ | 同上 / `clinic_settingsはソフトデリート対象外テーブルとして検出される` | 同上 | 解消済み(2026-07-13, commit bb2ad499) — 同上。green 実測済み |

※ 旧台帳の代表パス（`hospitalization_repository_test.go` / `medicine_repository_test.go` / `trimming_repository_test.go` 等）は `ef4f8cf5` の根因修正でテスト復帰済みのため 0 件（2026-07-13 実測）。上記 13 件も解除済みのため、**production-bug 起因の skip は現在 0 件**（`rg -in "known production bug|known bug" backend/internal --glob '*.go'` で実測確認、s3_r2_live_test.go の env-gate skip と preload_master_model_reconciliation_test.go の条件付き skip を除く）。
