# Repository Layer — P2 / P3 / P4 / P9 / P16

## P2: deleted_at IS NULL in CountUsage (MANDATORY)

```go
// ✅
err := r.db.WithContext(ctx).Model(&model.Vaccination{}).
    Where("vaccine_id = ? AND deleted_at IS NULL", vaccineID).Count(&count).Error

// ❌
Where("vaccine_id = ?", vaccineID)  // deleted_at IS NULL なし
```

## P3: Preload with deleted_at IS NULL (MANDATORY)

ソフトデリート対象エンティティのすべての `Preload` に条件を付ける。

```go
// ✅
db.Preload("ReservationType", "deleted_at IS NULL").
   Preload("Doctor", "deleted_at IS NULL").Find(&reservations)

// ❌
db.Preload("ReservationType").Preload("Doctor").Find(&reservations)
```

**注意**: `Preload("Doctor")` は `Staff` モデルへのエイリアス — 対象。

対象 42 エンティティ（抜粋）: `Account`, `Billing`, `Cage`, `Checkup`, `Consultation`,
`Examination`, `Hospitalization`, `MedicalRecord`, `Medicine`, `Owner`, `Pet`,
`Reservation`, `Staff`, `Treatment`, `Vaccination`, `Vaccine` など。
完全なリストは `.claude/refs/gin-architecture-compliance.md` P3 を参照。

## P3.1: clinic-scoped マスタの Preload は clinic_id 述語必須 (MANDATORY — クロステナント read 漏洩防止)

clinic_id を持つ「マスタ/区分」を FK 値で `Preload` する場合、`deleted_at IS NULL` だけでなく
**`clinic_id` 述語を必ず付ける**。base クエリが clinic-scoped でも、FK 値（例: `vaccination.vaccine_id`）が
別クリニックのマスタを指すと（write 側の FK 検証漏れ・過去データ汚染 #124/#125）、clinic_id 述語の無い
Preload は別クリニックのマスタ名/価格を応答に混入させる（IDOR / read 漏洩）。

```go
// ✅ 単一クリニック
Preload("Vaccine", "clinic_id = ? AND deleted_at IS NULL", clinicID)
// ✅ 拠点横断 (#86: clinicIDs は handler で所属検証済み)
Preload("ReservationType", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs)

// ❌ clinic_id 述語なし → 別クリニックのマスタが混入する
Preload("Vaccine", "deleted_at IS NULL")
```

対象マスタ: `Vaccine` / `Medicine` / `Procedure` / `Consultation` / `ReservationType`(+`.Group`) /
`TrimmingCourse` / `TrimmingOption`(`TrimmingDetail.Course/Options`) / `Cage` / `Insurance` /
`ExaminationType` / `CheckupType` / `DiagnosisType`(+`Names`)/`DiagnosisName` など。
`AnimalSpecies`・`ManualArticle` はグローバルマスタ（clinic_id 無し）のため対象外。

**例外: Staff(`Doctor`/`EnteredByStaff`/`PaidByStaff`/`ClosedByStaff`/`CreatedByStaff` 等)**
staff は `staff_clinic_assignments` による**多医院所属**（`staffs.clinic_id` は主所属のみ）。
`staffs.clinic_id = ?` 単純スコープは共有スタッフを誤って隠す。さらに既往カルテ等の**履歴 preload**を
スコープすると、退職/再配属したスタッフの担当医名が過去記録から消える回帰を生む。
- **履歴系 preload（medical_record/vaccination/examination/hospitalization/checkup/billing/refund 等の Doctor/Staff）は意図的に scope しない**（漏洩は staff 名のみで低 severity・write 隔離 72e8887c で通常到達不可）。
- **現在/未来データの reservation の `Doctor`/`CreatedByStaff` のみ** assignment-EXISTS でスコープ可
  （`staffAssignedToClinicsCond`・`reservation_repository.go`）。`staffs.clinic_id` でなく
  `staff_clinic_assignments` の所属で判定し多医院所属を尊重する。

### P3.1 の機械強制（read 側）と write 側の決着

**read 側（マスタ Preload の clinic_id 述語）は機械強制済み**:
`preload_clinic_scope_lint_test.go`（`go:embed` + `go/ast`、`model/audit_taxonomy_exhaustiveness_test.go` と同枠組み）が
本パッケージの全 `Preload` を走査し、clinic-scoped マスタの述語に `clinic_id` が無いものを CI で fail させる
（`go test ./internal/repository/ -run TestPreloadClinicScope`）。対象マスタ集合・Staff/global allowlist・
site 例外（clinicID が scope 外の identity/junction lookup）は同ファイルに根拠付きで一覧化。新規 `Preload` は自動で対象に入る。

**write 側（request 由来 master FK の所有権検証）は静的機械強制を断念**:
「永続化前に `FindByID(clinicID,…)` 等で検証済みか」は handler→service→repository を跨ぐ手続き間データフロー（taint）
解析が必須で、`go/ast` 単体では値を呼び出し越しに追えず false-negative/positive が多発する。#124（f4e7b7a7）が反例:
親 `exam_type_id` は検証済みでもネストの `exam_type_field_id` が未検証だった——「検証済み親」と「未検証ネスト子」を
意味理解なしに構文だけで区別する信頼できる規則は無い。**偽の部分チェックは #124 を再発させた「検証した気にさせる」
失敗モードそのものなので作らない。**
- **正本ガード = 各サイトの runtime isolation test**（`*_clinic_isolation_test.go` /
  `cross_tenant_master_fk_write_test.go`: treatment/vaccination/examination/care_plan/clinical_plan/checkup ほか）。
  検証が実際に効くことを動作で証明する。

**review 網羅性は機械強制済み（correctness ではない）**:
`internal/service/master_fk_write_inventory_lint_test.go`（`go:embed` + `go/ast`、本 read lint と同枠組み）が
service の全 exported メソッドを走査し、**param が transitively に clinic-scoped master FK フィールド**
（`MedicineID`/`VaccineID`/`ExamTypeFieldID`/`ParentID`(self-ref master)/`GroupID`/… を含むネスト/slice/embedded DTO）
**を受けるもの**を列挙、保守された `masterFKWriteAllowlist`（各エントリに `guarded`/`known-unguarded`/`exempt` の
status + 含む master FK 集合を pin）と**双方向**突合する。新規未登録 write・stale エントリ・**既存 DTO への
master FK 追加（#124 のネスト子追加形）**で CI fail（`go test ./internal/service/ -run TestMasterFKWriteInventory`、
独立 job `master-fk-write-inventory`・DB 不要・全イベント無条件）。列挙トリガは**動詞ではなく master-FK 包含**
なので `ValidateAndCreate`/`Confirm`/`Close` 等の非 Create/Update 永続化経路も自動捕捉する。
- **この gate は correctness を検証しない**（FindByID ガードの有無は見ない＝上記 taint 断念のとおり）。
  「master-FK write を必ずレビュー（= isolation test 追加）の俎上に乗せる」名簿のみを担保する。status は人間のレビュー記録。
- **gate が cov[er] しない残存ギャップ**（同ファイル冒頭にも明記）:
  ①ガードの正しさ（runtime test が正本）②裸スカラ param の master FK（`medicineDoseParam.Upsert`・
  `staff.Set{Excluded,Capable}ReservationTypeIDs` 等は DTO field でないため対象外）③`model.`/stdlib 以外の
  cross-package struct param（現状ゼロ・`knownSafeParamQualifiers` で新出は fail-closed）。
- **新マスタ追加時**: write 側 `clinicScopedMasterFKField` を更新し、**そのマスタを clinic_id 述語なしで Preload する箇所が
  あれば** read 側 `clinicScopedMasterAssoc` も更新する。現在 `ChiefComplaintType`/`HospitalizationPlan`/`MerchandiseItem`/
  `PaymentMethodMaster`/`TrimmingCourseType`/`InventoryItem` は write 側のみに在る（read 側で Preload されていないため）。
  これらを将来 Preload する場合は **先に read 側 allowlist へ追加**しないと read gate を素通りする。

## P4: clinicScope on Update/Upsert (MANDATORY — 最重要)

UPDATE/UPSERT に `Scopes(clinicScope(clinicID))` を必ず付ける。

```go
// ✅
err := r.db.WithContext(ctx).Model(&model.Vaccine{}).
    Scopes(clinicScope(clinicID)).
    Where("id = ?", id).Updates(fields).Error

// ❌ クロスクリニックデータ更新リスク
r.db.Model(&model.Vaccine{}).Where("id = ?", id).Updates(fields)
```

**例外（clinicScope 不要）**: `clinic_repository.go`, `company_repository.go`,
`account_repository.go`, `password_reset_token_repository.go`, `audit_repository.go`

## P9: apperrors.FromGORM on GORM errors (MANDATORY)

```go
// ✅
if err := r.db.WithContext(ctx).First(&vaccine, id).Error; err != nil {
    return nil, apperrors.FromGORM(err, "vaccine", fmt.Sprintf("%d", id))
}

// ❌
if err := r.db.First(&vaccine, id).Error; err != nil {
    return nil, err
}
```

## P16: Method naming conventions (MANDATORY)

```
FindAll / FindByClinicID  ← 一覧（GetAll, List, Fetch は違反）
FindByID                  ← 単件（GetByID, Get, Find は違反）
Create / Update / Delete  ← 標準
CountBy{Xxx}              ← カウント
CountUsageBy{Xxx}         ← 使用数カウント
```

```go
// ✅
type VaccineRepository interface {
    FindAll(ctx context.Context, clinicID uint64) ([]*model.Vaccine, error)
    FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccine, error)
    Create(ctx context.Context, vaccine *model.Vaccine) error
    Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccine, error)
    Delete(ctx context.Context, clinicID, id uint64) error
}

// ❌
GetAll(...)  GetByID(...)  List(...)  Fetch(...)
```

## テストヘルパー: DROP+CREATE 系は setupIsolatedTestDB を使う (MANDATORY)

`*_test.go` の setup ヘルパーがテーブル/ENUM 型を `DROP TABLE`/`DROP TYPE` → 再作成する場合、
`setupTestDB`（全テストで共有する DB 接続プール）ではなく `setupIsolatedTestDB`（呼び出し毎の
使い捨て接続、`ltv_repository_test.go`）を使う。共有プール上で DROP+CREATE すると、他テストが
保持する古いテーブル/型 OID を参照したキャッシュ済み prepared statement が壊れる
（`cache lookup failed` SQLSTATE XX000 / `cached plan must not change result type` SQLSTATE 0A000）。

```go
// ✅ DROP+CREATE を伴う setup は使い捨て接続
func setupXxxTestDB(t *testing.T) *gorm.DB {
    db := setupIsolatedTestDB(t)
    db.Exec("DROP TABLE IF EXISTS xxx CASCADE")
    // ...
}

// ❌ 共有プールで DROP+CREATE すると他テストの prepared statement を破壊しうる
func setupXxxTestDB(t *testing.T) *gorm.DB {
    db := setupTestDB(t)
    db.Exec("DROP TABLE IF EXISTS xxx CASCADE")
    // ...
}
```

TRUNCATE のみ（テーブル構造を変えない）なら `setupTestDB` のままでよい。既存例:
`checkup_field_repository_test.go` の `setupCheckupFieldTestDB`。
