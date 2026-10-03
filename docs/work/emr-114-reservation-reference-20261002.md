# EMR-114 (SLACK-RESERVATION-REFERENCE) — 予約「参照先が存在しません」の現行 build 再検証

- 日付: 2026-10-02
- 対象 unit: `UNIT-EMR-114`（campaign `emr114-sat-20261002e` / attempt `ATT-EMR-114-001`、claim ブランチ `claim/EMR-114`）
- 参照表の正本: [`docs/work/remaining-campaign-20260920/SLACK-RESERVATION-REFERENCE.md`](remaining-campaign-20260920/SLACK-RESERVATION-REFERENCE.md)
- 追加した pin: `backend/internal/reservation/reservation_staff_reference_test.go`
- 検証方法: backend Go テストのみ（Docker `--entrypoint go`）。フロントエンド変更・新エンドポイント・migration 追加・STG/共有環境操作は行っていない。

## 目的とスコープ

報告: 「スタッフ／シフト作成後も予約保存が『参照先が存在しません』で失敗した」。9/13 に対応報告済みだが、当時の修正と現行 build の同一性は別証跡であり、本ユニットは**現行コード+現行 DDL で同じ操作を再実行**して (a) save+reload 成功、(b) 非存在スタッフ参照の拒否、(c) 他医院スタッフ参照の拒否を pin する。再現した場合は失敗経路を特定して最小修正、再現しない場合は close 提案を出す。

## 実行した参照チェック経路（SLACK-RESERVATION-REFERENCE.md の A/B 表との対応）

| 層 | 経路（コード位置） | 本ユニットで実測した内容 |
|----|--------------------|---------------------------|
| A1 | `checkDoctorClinicAssignment`（`reservation_handler.go:32-46`）— staff_clinic_assignments で担当医の操作医院所属を確認 | handler 経路（`CreateReservation` + gin ctx `clinic_id`/`user_id` + 実 `StaffClinicAssignmentService`）で非存在 ID・他医院のみ所属 ID の両方が 400 `指定されたスタッフはこの医院に所属していません` で拒否されることを実測 |
| A2/A3 | `ValidateReservationStaffCapability` → `guard.FindByID` + `SupportsReservationType`（`reservation_staff_capability_validator.go:40-63`、`reservation_staff_repository.go:80-107,471-488`）— tx 内で staff/assignment/capability を FOR SHARE | service 経路（実 `ReservationRepository`/`ReservationStaffRepository`/`persistence.NewTransactor`）で非存在・他医院のみ所属が not-found、capability 保有の所属医は通過することを実測 |
| A7 | `assertReservationCreatedBy`（`reservation_created_by.go:19`） | creator として assignment 済み active スタッフを `user_id` から注入し通過を実測。兼務 creator（主所属≠操作医院、assignment あり）も実 DDL で保存成功を実測（003 修正済み経路） |
| B | 複合 FK 群 `fk_appointments_*`（`migrations/001_init.sql:5731-5768`）→ `23503` → `参照先が存在しません` | 共有テスト DB は AutoMigrate 由来で複合 doctor FK を持たないため、`*_test` DB 内スクラッチスキーマへ `001_init.sql`+`002`+`003` を適用した実 DDL で B 層を実測 |

## 実行レシート（verbatim）

共通シード（両スキーマ層で同一構成）: `clinic A` / `clinic B`、creator staff（clinic A 主所属 + assignment A + active）、`reservation_type A`、担当医 `doctorA`（主所属 A + assignment A + capability(A,doctorA,typeA) + `shift_entries` 当日 full）、`doctorShared`（主所属 B + assignment A/B + capability(A,doctorShared,typeA) + shift A）、`otherClinicStaff`（主所属 B + assignment B のみ）。payload は `{"start_time":"2027-06-01T10:00:00Z","end_time":"2027-06-01T10:30:00Z","reservation_type_id":<rtA>,"visit_type":"revisit","status":"pending","doctor_id":<id>}`（subtest 毎に時間枠をずらす）。

### 層 A: 共有テスト DB（AutoMigrate スキーマ）— `TestReservationStaffReference_SharedSchema`

```
--- PASS: TestReservationStaffReference_SharedSchema (0.53s)
    --- PASS: .../assigned_doctor_saves_and_reloads_via_handler_path (0.01s)
    --- PASS: .../nonexistent_doctor_is_rejected_before_any_write (0.00s)
    --- PASS: .../other-clinic-only_staff_reference_is_rejected_at_both_layers (0.00s)
    --- PASS: .../assigned_multi-clinic_doctor_saves_on_automigrate_schema (0.00s)
    --- PASS: .../doctor_id_zero_normalizes_to_unset (0.00s)
```

| 操作 | 結果 |
|------|------|
| POST `doctor_id=doctorA`（報告操作の再現） | HTTP 201、応答 `"doctor_id":<doctorA>`、`repo.FindByID` reload で `DoctorID==doctorA.ID` かつ `Doctor` preload が名前付きで解決 |
| POST `doctor_id=<max(staffs.id)+900001>` | handler 400 `指定されたスタッフはこの医院に所属していません`（A1）。service 直叩きは `apperrors.IsNotFound`（A3）。0 行 |
| POST `doctor_id=otherClinicStaff` | handler 400 同上（A1）、service not-found（A3）。0 行 |
| POST `doctor_id=doctorShared`（兼務） | **201 成功** — AutoMigrate スキーマには複合 `(doctor_id, clinic_id)` FK が無く `staffs.id` の存在しか見ないため保存される（層 B との差分が要点） |
| POST `doctor_id=0` | 201 成功、reload で `DoctorID == nil`（`normalizeCreateDoctorID` の 0→NULL 正規化が現行コードで有効） |

### 層 B: 実 DDL スクラッチスキーマ（001+002+003 を `ekarte_db_test` 内専用スキーマへ適用）— `TestReservationStaffReference_RealDDL`

```
--- PASS: TestReservationStaffReference_RealDDL (1.33s)
    --- PASS: .../same-primary-clinic_doctor_saves_and_reloads_on_real_DDL (0.01s)
    --- PASS: .../handler_create_with_assigned_doctor_returns_201 (0.00s)
    --- PASS: .../multi-clinic_creator_saves_on_post-003_created_by_FK (0.01s)
    --- PASS: .../assigned_multi-clinic_doctor_on_real_DDL_records_current_constraint_behavior (0.00s)
    --- PASS: .../nonexistent_doctor_stays_rejected_on_real_DDL (0.00s)
    --- PASS: .../other-clinic-only_staff_stays_rejected_on_real_DDL (0.00s)
```

| 操作 | 結果 |
|------|------|
| svc.Create `doctor_id=doctorA`（主所属=操作医院） | 成功、FindByID で `Doctor` reload |
| POST `doctor_id=doctorA`（handler 経路） | 201 + reload 一致 |
| svc.Create `created_by=兼務 creator`、`doctor_id=nil` | 成功（003 の単一カラム `fk_appointments_created_by` が正しく機能） |
| POST `doctor_id=doctorShared`（兼務担当医） | **400 `参照先が存在しません`**、capture した PgError: `Code=23503`, `ConstraintName=fk_appointments_doctor_clinic`。appointments 行 0（部分書き込みなし） |
| svc.Create `doctor_id=90000042`（非存在） | `IsNotFound`（A3 が INSERT 前に拒否） |
| svc.Create `doctor_id=<B のみ所属>` | `IsNotFound`（A3） |

パッケージ全体の bound コマンド（`test -count=1 -p 1 ./internal/reservation/...`）:

```
ok  	github.com/animal-ekarte/backend/internal/reservation	21.172s   （EXIT=0）
```

## 判定

**部分再現（reproduced）— ただし残存経路は兼務スタッフ担当医に限定される。**

- 報告された標準操作（操作医院を主所属とするスタッフ＋シフト作成後にそのスタッフを担当医に指定して保存）は**現行 build で成功・再読込できる**（層 A・層 B 双方で実測）。非存在 ID・他医院のみ所属 ID はアプリ層（A1/A3）が専用メッセージ/ not-found で先に拒否し、総称 23503 には到達しない。`doctor_id=0` の正規化と `created_by` 兼務パス（003 修正）も実測で有効。
- **ただし同一エラーを出す残存経路が確定的に再現した**: `staff_clinic_assignments` で clinic A へ正当に配属され、capability・当日シフトも揃った兼務スタッフ（`staffs.clinic_id` が別医院）を担当医にすると、A1/A2/A3/A7 をすべて通過した末に `fk_appointments_doctor_clinic` が `(doctor_id, clinic_id) → staffs(id, clinic_id)` の主所属一致を要求して `23503 → 400 参照先が存在しません` となる。created_by 側は migration 003 で単一カラム化済みだが、doctor 側の複合 FK は 001 時点のまま残っており、アプリ層の assignment ベース所属モデル（`FindByID`・preload の `staffAssignedToClinicsCond`・`checkDoctorClinicAssignment`）と DB 制約が矛盾している。
- 共有テスト DB（AutoMigrate）ではこの操作が成功するため、**従来のテストハーネスでは構造的に検出不可能**だった点も併せて実測で示した。

## 最小修正（→ 2026-10-02 追記: `017_multiclinic_staff_fk_fix.sql` で適用済み）

- 必要な修正は `003_appointments_created_by_staff_fk.sql` と同型の新規 migration: `ALTER TABLE appointments DROP CONSTRAINT fk_appointments_doctor_clinic` + `ADD CONSTRAINT fk_appointments_doctor FOREIGN KEY (doctor_id) REFERENCES staffs (id) ON DELETE SET NULL`（SET NULL 動作は現行制約の `ON DELETE SET NULL (doctor_id)` を維持）。アプリ層の検証（assignment+capability）は既に完全であり、コード変更は不要。
- ~~本ユニットの forbidden_ops に「migrate apply / 新規 migration ファイル」が含まれるため、上記修正の実施は BLOCKED（スコープ外権限）~~ → **controller 側フォローアップとして `017_multiclinic_staff_fk_fix.sql` で実施済み**。001 内の同一欠陥クラス全6サイト（`appointments.doctor_id`・`hospitalizations.doctor_id`・`medical_records.doctor_id`・`cash_register_close_adjustments.actor_id`・`medical_record_image_upload_quota.staff_id`・`lab_device_waits.staff_id`）を一括で単一カラム化。`RealDDL` スクラッチスキーマは 017 を適用し、兼務 subtest は成功 pin + 制約形状 pin（単一カラム 6件存在・旧複合 6件不在）に反転済み。

## close 提案の可否

- 「対応報告後の確認済み」として閉じる条件（対象 build の同じ操作で保存・再読込でき、無効/他院参照が拒否される receipt）は、**標準パスについては満たしている**。
- 兼務スタッフ経路の残存不具合は `017_multiclinic_staff_fk_fix.sql` で修正済み（実 DDL ピン留め反転で検証）。**各環境で `make migrate` 適用後に close 可能**。適用前の環境では 001 由来の複合 FK が残るため、兼務スタッフ経路は引き続き `参照先が存在しません` を返すことに注意。

## ハーネス覚書

- `TestReservationStaffReference_RealDDL` は `ekarte_db_test`（`DB_NAME`+`_test` 派生、`TEST_DATABASE_URL` は読まない）内部の `rsv_ref_<nonce>` スキーマに実 DDL を適用し、cleanup で `DROP SCHEMA ... CASCADE` する。共有 `public` スキーマのオブジェクトは読み取りのみで変更しない。
- `search_path` は `<schema>,public` とする（`pg_trgm` 等の extension は `public` に既インストールで `CREATE EXTENSION IF NOT EXISTS` が skip されるため、演算子クラス解決に public が必要）。
- `internal/lintscan` の inventory 変更なし（新規 `DBOrTx` call site は test ファイル内のみで production code 非該当）。
