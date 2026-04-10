# BUG-286: デッドコード（第2回監査）— 安全に削除可能な項目一覧

## 概要

BUG-279 の対応後に再監査を実施。全項目は grep で呼び出し元ゼロを確認済み。

## カテゴリ A: 確実にデッドコード（即削除可）

### 1. SanitizeNullBytes ミドルウェアの二重登録

`handler/handler.go:67` — `protected.Use(middleware.SanitizeNullBytes())` は不要。
`cmd/api/main.go:100` で `r.Use(middleware.SanitizeNullBytes())` がグローバル登録済みのため、
protected グループを含む全ルートに既に適用されている。

### 2. handler 層の重複 validateTreatmentItemType

`handler/validators.go:43-51` — `validateTreatmentItemType(v string) error`

- `service/treatment_service.go:330` に正規版が存在（`model.TreatmentItemType` 型引数）
- handler 版は `CreateTreatment` でのみ呼ばれ、`UpdateTreatment` では呼ばれない（不整合）
- service 版が Create/Update 両方でバリデーションするため、handler 版は冗長

### 3. RefreshToken のヘルパーメソッド（3メソッド）

`model/auth.go:19-30` — `IsExpired()`, `IsRevoked()`, `IsValid()`

- `auth_handler.go` の RefreshToken 処理はフィールドを直接参照しており、これらのメソッドを一切呼んでいない
- インターフェースにも含まれていない

### 4. AuditActionUserPermissionGroupSet 定数

`model/audit_log.go:32` — `AuditActionUserPermissionGroupSet = "user_permission_group.set"`

- 定義のみ。`permission_group_handler.go` の SetStaffGroups 操作は監査ログを記録していない
- 他の `AuditAction*` 定数は BUG-278 で組み込み済みだが、この定数だけ未使用

### 5. AccountRepository.Delete

`repository/account_repository.go:74-78` — `Delete(ctx context.Context, id uint64) error`

- インターフェース定義 + 実装の両方が存在
- `AccountService` にも `auth_handler` にも Delete メソッドが存在しない
- アカウント削除機能は未実装

### 6. BillingItemService.GetByID

`service/billing_item_service.go:73` — インターフェース定義
`service/billing_item_service.go:88-94` — 実装

- handler は `CreateItem`, `UpdateItem`, `DeleteItem` のみ呼び出し
- 単一 BillingItem 取得のルート（GET /billing-items/:id）は存在しない

### 7. ReservationCourseService.GetByID

`service/reservation_course_service.go:15` — インターフェース定義
`service/reservation_course_service.go:71-77` — 実装

- handler は `List`, `Create`, `Update`, `Delete`, `PatchStatus`, `PatchSortOrder` を呼び出し
- 単一コース取得のルートは存在しない

### 8. StaffClinicAssignment の未使用メソッド（前回 BUG-279 カテゴリ B から昇格）

**Repository 層** — `staff_clinic_assignment_repository.go`:
- `FindByClinicID` / `Update` / `Delete`

**Service 層** — `staff_clinic_assignment_service.go:14-17` + 実装（36-72行）:
- `FindByClinicID` / `Update` / `Delete`

現在使用されているのは `FindByStaffID` と `Create` のみ。

## カテゴリ B: 未使用モデル定数（DB enum 値）

以下の定数はコード上の参照ゼロだが、DB の enum 値やフロントエンドのフォーム入力として有効な値である。
削除してもコンパイルは通るが、将来の enum バリデーション強化時に再定義が必要になる。

| ファイル | 定数 | 参照 |
|---------|------|------|
| `model/hospitalization.go:76-78` | `PlanTimingMorning`, `PlanTimingNoon`, `PlanTimingNight` | ゼロ（サービスは文字列リテラルを使用） |
| `model/vital.go:9` | `BodyWeightUnitG` | ゼロ（`BodyWeightUnitKg` のみ使用） |
| `model/accounting.go:22-23` | `BillingStatusCancelled`, `BillingStatusPending` | ゼロ（`Waiting`/`Completed` のみ使用） |
| `model/accounting.go:29-31` | `PaymentMethodCash`, `PaymentMethodCreditCard`, `PaymentMethodElectronicMoney` | ゼロ |

## カテゴリ C: テスト専用（本番コード参照ゼロ）

以下はテストコードでのみ使用。テストのアサーション用として有用のため削除は非推奨。

| ファイル | シンボル | テストでの使用箇所 |
|---------|---------|------------------|
| `model/staff.go:48-50` | `ShiftTypeFull`, `ShiftTypeMorning`, `ShiftTypeAfternoon` | `shift_entry_service_test.go`（8箇所） |
| `errors/errors.go:69` | `IsInvalidInput(err) bool` | 11テストファイルで使用 |
| `errors/errors.go:83` | `IsAlreadyExists(err) bool` | `staff_service_test.go` で使用 |

## 削除手順

1. カテゴリ A の 1〜7 を削除
2. `go vet ./...` で確認
3. カテゴリ A の 8（StaffClinicAssignment）を削除 + インターフェース縮小
4. `go test ./...` で確認
5. カテゴリ B は製品判断後に対応（enum バリデーション導入時に再利用可能）
6. カテゴリ C は削除しない（テストアサーション用として有用）

## 優先度

**Medium** — 機能的な影響はないが、コードベースの可読性と保守性に寄与。

## 関連チケット

- BUG-285: デッドコード監査第2回 親チケット
- BUG-277〜279: デッドコード監査第1回（対応済み）
