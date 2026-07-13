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

## 既知バグの skip テスト台帳（2026-07-13 第2次棚卸し）

skip されたテストはバグが直っても・悪化しても fail しない。修正時は該当行の `t.Skip` を外して green を確認すること。件数は `rg -in "known production bug" backend/internal --glob '*.go'` の t.Skip 実測 11 件（+ 補足 2 件）。**旧台帳（小文字 `known production bug` 表記 9 件）は `ef4f8cf5` で根因修正・`5be86907` でクローズ済み**。本表はその後の #212 カバレッジ拡充（`e908f105`、2026-07-13）で新規発覚した未修正分であり、修正タスクは #236 に起票済み。

### `t.Skip("KNOWN PRODUCTION BUG ...")` — 11 件

| ファイル:行 | テスト名 | skip メッセージ要約 |
|---|---|---|
| `internal/repository/reservation_staff_repository_test.go:166` | `TestReservationStaffRepository_Delete` / `別クリニックからの削除はNotFoundで行が残る` | **CRITICAL**: Delete の `Joins()+Delete()` が GORM 上で機能せずクロステナント削除が成立する |
| `internal/repository/medical_record_owner_visit_repository_test.go:205` | `TestMedicalRecordRepository_FindLatestByOwner` / `clinic_id隔離: 別クリニックのカルテは対象外` | `apperrors.IsNotFound` が生 gorm エラーに対し常に false |
| `internal/repository/medical_record_owner_visit_repository_test.go:221` | `TestMedicalRecordRepository_FindLatestByOwner` / `該当なし: カルテが存在しない飼い主は nil, nil` | 同上 |
| `internal/repository/clinic_settings_repository_test.go:109` | `TestClinicSettingsRepository_FindByClinicID` / `clinic_id隔離: 他院の設定は返らず自院はデフォルト値のまま` | **CRITICAL**: `repo.Save` が列名不一致（SQLSTATE 42703）で失敗 |
| `internal/repository/clinic_settings_repository_test.go:127` | `TestClinicSettingsRepository_FindByClinicID` / `行が存在すれば実際の値を返す` | 同上 |
| `internal/repository/clinic_settings_repository_test.go:141` | `TestClinicSettingsRepository_Save`（親テストごと skip） | 同上 |
| `internal/repository/clinic_settings_repository_test.go:172` | `TestClinicSettingsRepository_UpdateCPMVersion`（親テストごと skip） | **CRITICAL**: 内部 UPSERT が列名不一致（SQLSTATE 42703）で失敗 |
| `internal/repository/clinic_settings_repository_test.go:203` | `TestClinicSettingsRepository_UpdateDormantThresholds`（親テストごと skip） | 同上 |
| `internal/repository/clinic_settings_repository_test.go:239` | `TestClinicSettingsRepository_UpdateCPMV2Thresholds`（親テストごと skip） | 同上 |
| `internal/repository/clinic_settings_repository_test.go:260` | `TestClinicSettingsRepository_UpdateCPMV1Thresholds`（親テストごと skip） | 同上 |
| `internal/repository/clinic_settings_repository_test.go:286` | `TestClinicSettingsRepository_UpdateHealthPreventionThresholds`（親テストごと skip） | 同上 |

### 補足: `t.Skip("known bug ...")` — 2 件（テスト基盤起因、production バグではない）

| ファイル:行 | テスト名 | skip メッセージ要約 |
|---|---|---|
| `internal/repository/clinic_repository_test.go:284` | `TestClinicRepository_CountBlockingReferencesByClinicID`（親テストごと skip） | `model.ClinicSettings` の AutoMigrate schema drift がテーブル依存チェックを阻害 |
| `internal/repository/clinic_repository_test.go:337` | 同上 / `clinic_settingsはソフトデリート対象外テーブルとして検出される` | 同上 |

※ 旧台帳の代表パス（`hospitalization_repository_test.go` / `medicine_repository_test.go` / `trimming_repository_test.go` 等）は `ef4f8cf5` の根因修正でテスト復帰済みのため 0 件（2026-07-13 実測）。現存する skip は上記のみ。
