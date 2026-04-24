---
description: Gin/GORM architecture compliance check (P1-P18) for handler/service/repository layers
alwaysApply: false
globs: ["backend/internal/**/*.go"]
---

# Gin Architecture Compliance Rules (P1–P18)

This project enforces 18 fixed patterns across handler / service / repository layers.
**When reviewing or implementing Go backend code, check ALL applicable patterns below.**

---

## Layer Responsibilities

```
Handler   → Bind request → call Service → toXxxResponse() → c.JSON()
Service   → Business logic → call Repository → apperrors.Wrap()
Repository → GORM queries → apperrors.FromGORM()
```

---

## P1: FindByID before Delete/Update (Service)

Delete/Update メソッドは先頭で必ず `FindByID` を呼ぶ。

```go
// ✅ Correct
func (s *vaccineService) Update(ctx context.Context, clinicID, id uint64, input UpdateVaccineInput) (*model.Vaccine, error) {
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {  // 先に存在確認
        return nil, apperrors.Wrap(err, "failed to find vaccine")
    }
    fields := buildVaccineUpdateFields(input)
    // ...
}

// ❌ Wrong: Update before FindByID
func (s *vaccineService) Update(...) (*model.Vaccine, error) {
    return s.repo.Update(ctx, id, fields)  // FindByID なし
}
```

---

## P2: CountUsage with deleted_at IS NULL (Repository)

CountBy*/CountUsage* の WHERE 句に `deleted_at IS NULL` を含める。

```go
// ✅ Correct
func (r *vaccineRepository) CountUsageByVaccineID(ctx context.Context, vaccineID uint64) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).Model(&model.Vaccination{}).
        Where("vaccine_id = ? AND deleted_at IS NULL", vaccineID).Count(&count).Error
    return count, err
}

// ❌ Wrong
Where("vaccine_id = ?", vaccineID)  // deleted_at IS NULL なし
```

---

## P3: Preload with deleted_at IS NULL (Repository)

`gorm.DeletedAt` を持つエンティティの Preload には条件を付ける。

ソフトデリート対象（42エンティティ）:
`Account`, `StaffClinicAssignment`, `Billing`, `BillingItem`, `Payment`,
`Cage`, `Checkup`, `CheckupType`, `ChiefComplaintType`, `ClinicalPlan`,
`Consultation`, `DiagnosisType`, `DiagnosisName`, `Estimate`, `EstimateItem`,
`Examination`, `ExaminationType`, `HospitalizationPlan`, `Hospitalization`, `TreatmentPlan`,
`InquiryTemplate`, `Insurance`, `InventoryItem`, `MedicalRecord`, `Medicine`,
`MerchandiseItem`, `Occupation`, `Owner`, `PaymentMethodMaster`, `PermissionGroup`,
`Pet`, `Procedure`, `Reservation`, `ReservationType`, `ReservationTypeGroup`,
`Staff`, `ShiftTemplate`, `Treatment`, `TrimmingCourse`, `TrimmingOption`,
`Vaccination`, `Vaccine`

**注意: `Preload("Doctor")` は `Staff` モデルへのエイリアス。必ず対象。**

```go
// ✅ Correct
db.Preload("ReservationType", "deleted_at IS NULL").
   Preload("Doctor", "deleted_at IS NULL").Find(&reservations)

// ❌ Wrong
db.Preload("ReservationType").Preload("Doctor").Find(&reservations)
```

---

## P4: clinicScope on Update/Upsert (Repository)

UPDATE/UPSERT クエリに `Scopes(clinicScope(clinicID))` を付ける。

**例外（clinicScope 不要）**: `clinic_repository.go`, `company_repository.go`, `account_repository.go`, `password_reset_token_repository.go`, `audit_repository.go`

```go
// ✅ Correct
func (r *vaccineRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccine, error) {
    err := r.db.WithContext(ctx).Model(&model.Vaccine{}).
        Scopes(clinicScope(clinicID)).
        Where("id = ?", id).Updates(fields).Error
    // ...
}

// ❌ Wrong
r.db.Model(&model.Vaccine{}).Where("id = ?", id).Updates(fields)  // clinicScope なし
```

---

## P5: RequirePermission on write routes (Routes)

POST/PUT/PATCH/DELETE ルートには必ず `RequirePermission` を設定する。

**免除**: `/login`, `/logout`, `/auth/*`, `/me`, LIFF公開API

```go
// ✅ Correct (master routes)
masters.POST("/vaccines", RequirePermission("edit"), h.Create)
masters.DELETE("/vaccines/:id", RequirePermission("delete"), h.Delete)

// ✅ Correct (business routes with helper)
owners.POST("", h.RequirePermission(string(model.ResourceOwners), "create"), h.CreateOwner)

// ❌ Wrong
masters.POST("/vaccines", h.Create)  // パーミッションなし
```

---

## P6: DELETE routes use "delete" permission (Routes)

DELETE ルートには `"delete"` パーミッションを使う（`"edit"` は違反）。

```go
// ✅ Correct
masters.DELETE("/vaccines/:id", RequirePermission("delete"), h.Delete)

// ❌ Wrong
masters.DELETE("/vaccines/:id", RequirePermission("edit"), h.Delete)
```

---

## P7: toXxxResponse() conversion in handler (Handler)

handler で `c.JSON` にモデルを直接渡さない。必ず変換関数を経由する。

```go
// ✅ Correct
c.JSON(http.StatusOK, toVaccineResponse(vaccine))
c.JSON(http.StatusOK, toVaccineListResponse(vaccines))

// ❌ Wrong
c.JSON(http.StatusOK, vaccine)   // モデル直接返却
c.JSON(http.StatusOK, gin.H{"data": vaccine})
```

---

## P8: apperrors.Wrap in service (Service)

service 内の全エラーリターンを `apperrors.Wrap` で包む。

```go
// ✅ Correct
vaccine, err := s.repo.FindByID(ctx, clinicID, id)
if err != nil {
    return nil, apperrors.Wrap(err, "failed to find vaccine")
}

// ❌ Wrong
if err != nil {
    return nil, err  // Wrap なし
}
```

---

## P9: apperrors.FromGORM in repository (Repository)

GORM エラーを `apperrors.FromGORM` で変換する。

```go
// ✅ Correct
if err := r.db.WithContext(ctx).First(&vaccine, id).Error; err != nil {
    return nil, apperrors.FromGORM(err, "vaccine", fmt.Sprintf("%d", id))
}

// ❌ Wrong
if err := r.db.First(&vaccine, id).Error; err != nil {
    return nil, err  // FromGORM 未使用
}
```

---

## P10: FK dependency check before Delete (Service)

service の Delete で参照カウントを確認し、count > 0 なら `WrapConflict` を返す。

**注意**: 末端エンティティ（他から FK 参照されない）は依存チェック不要。

```go
// ✅ Correct
func (s *vaccineService) Delete(ctx context.Context, clinicID, id uint64) error {
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return apperrors.Wrap(err, "failed to find vaccine")
    }
    count, err := s.repo.CountUsageByVaccineID(ctx, id)
    if err != nil {
        return apperrors.Wrap(err, "failed to count usage")
    }
    if count > 0 {
        return apperrors.WrapConflict(fmt.Errorf("vaccine %d is in use", id))
    }
    return apperrors.Wrap(s.repo.Delete(ctx, clinicID, id), "failed to delete vaccine")
}
```

---

## P11: slog.ErrorContext on repository error paths (Service)

repo 呼び出しが返したエラー（DB障害・インフラ起因）のリターン前に `slog.ErrorContext` を呼ぶ。

**除外（ログ不要）**: `WrapInvalidInput` / NotFound 存在確認 / `WrapConflict`（ユーザー起因の正常フロー）

```go
// ✅ Correct
vaccines, err := s.repo.FindAll(ctx, clinicID)
if err != nil {
    slog.ErrorContext(ctx, "failed to find vaccines", "error", err)  // ← 必須
    return nil, apperrors.Wrap(err, "failed to find vaccines")
}

// ❌ Wrong: ログなし
if err != nil {
    return nil, apperrors.Wrap(err, "failed to find vaccines")
}
```

---

## P12: ShouldBindJSON unified error handling (Handler)

`ShouldBindJSON` エラーは統一形式で処理する。

```go
// ✅ Correct
var req CreateVaccineRequest
if err := c.ShouldBindJSON(&req); err != nil {
    RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
    return
}

// ❌ Wrong
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
}
```

---

## P13: const/buildFunc definition order (Service)

service ファイルでの定義順序を守る。

```go
// ✅ Correct order
const (
    colVaccineName = "name"  // 1. const
)

func buildVaccineUpdateFields(input UpdateVaccineInput) map[string]any { ... }  // 2. buildFunc

type VaccineService interface { ... }   // 3. interface
type vaccineService struct { ... }      // 4. struct
func NewVaccineService(...) { ... }     // 5. constructor
func (s *vaccineService) Create(...) { ... }  // 6. methods

// ❌ Wrong: interface が const より前
type VaccineService interface { ... }
const colVaccineName = "name"
```

---

## P14: Handler must not call Repository directly (Handler)

handler struct は `svc XxxService` のみを持つ。`repo` フィールドは禁止。

```go
// ✅ Correct
type VaccineHandler struct {
    svc service.VaccineService  // service 経由のみ
}

// ❌ Wrong
type VaccineHandler struct {
    svc  service.VaccineService
    repo repository.VaccineRepository  // 直接注入禁止
}
```

---

## P15: POST returns 201 + Location header (Handler)

Create 系ハンドラは `http.StatusCreated(201)` と `Location` ヘッダを返す。

```go
// ✅ Correct (master routes)
c.Header("Location", fmt.Sprintf("/v1/masters/vaccines/%d", vaccine.ID))
c.JSON(http.StatusCreated, toVaccineResponse(vaccine))

// ✅ Correct (business routes)
c.Header("Location", fmt.Sprintf("/api/v1/owners/%d", owner.ID))
c.JSON(http.StatusCreated, toOwnerResponse(owner))

// ❌ Wrong
c.JSON(http.StatusOK, toVaccineResponse(vaccine))    // 200
c.JSON(http.StatusCreated, toVaccineResponse(vaccine)) // Location なし
```

---

## P16: Repository method name conventions (Repository)

```
FindAll / FindByClinicID  ← 一覧取得（GetAll, List, Fetch は違反）
FindByID                  ← 単件取得（GetByID, Get, Find は違反）
Create                    ← 作成
Update                    ← 更新
Delete                    ← 削除
CountBy{Xxx}              ← カウント
CountUsageBy{Xxx}         ← 使用数カウント
```

```go
// ✅ Correct
type VaccineRepository interface {
    FindAll(ctx context.Context, clinicID uint64) ([]*model.Vaccine, error)
    FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccine, error)
    Create(ctx context.Context, vaccine *model.Vaccine) error
    Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccine, error)
    Delete(ctx context.Context, clinicID, id uint64) error
    CountUsageByVaccineID(ctx context.Context, vaccineID uint64) (int64, error)
}

// ❌ Wrong
GetAll(...)    // → FindAll
GetByID(...)   // → FindByID
List(...)      // → FindAll
```

---

## P17: Input struct naming convention (Service)

```go
// ✅ Correct
type CreateVaccineInput struct { ... }
type UpdateVaccineInput struct { ... }

// ❌ Wrong
type VaccineCreateRequest struct { ... }  // 順序逆
type CreateVaccineParams struct { ... }   // Params は違反
type VaccineInput struct { ... }          // Create/Update が不明
```

---

## P18: toXxxResponse function naming (Handler)

```go
// ✅ Correct
func toVaccineResponse(v *model.Vaccine) VaccineResponse { ... }
func toVaccineListResponse(vs []*model.Vaccine) []VaccineResponse { ... }

// ❌ Wrong
func convertToVaccine(...)     // convert プレフィックス
func buildVaccineResponse(...) // build プレフィックス
func mapVaccine(...)           // map プレフィックス
func newVaccineResponse(...)   // new プレフィックス
```

---

## Compliance Check Prompts

フェーズ・対象に応じて以下のプロンプトをそのままコピーして使用せよ。
詳細プロンプト例: `tmp/BE/be_master_examples.md` / `tmp/BE/be_non_master_examples.md`

| フェーズ | 対象 | 状況 | 使うプロンプト |
|---------|------|------|-------------|
| スキャン | マスタ系 | スキャン〜起票まで一気に | マスタ スキャン例A |
| スキャン | マスタ系 | 修正後に特定パターンのみ再確認 | マスタ スキャン例B |
| スキャン | 業務系 | スキャン〜起票まで一気に | 業務系 スキャン例A |
| スキャン | 業務系 | 修正後に特定パターンのみ再確認 | 業務系 スキャン例B |
| 実装 | どちらも | 多数タスクを並列処理 | 実装例A |
| 実装 | どちらも | 全タスクを優先度順に一括処理 | 実装例B |
| 単発 | 単一ファイル | 新規実装後の即時確認 | 単一ファイルチェック |

---

### マスタ系 スキャン例A: スキャン + 起票（フル実行）

```
【Step 0: 事前整合性チェック（必須・スキャン前に必ず実行）】
以下の Python スクリプトを Bash ツールで実行し、ドキュメントに記載された全ファイルが存在するか確認せよ。

python3 -c "
import re, os
with open('tmp/BE/be_master_check.md') as f:
    content = f.read()
files = re.findall(r'- (backend/internal/\S+\.go)', content)
missing = [f for f in files if not os.path.exists(f)]
print(f'登録ファイル数: {len(files)}')
print(f'欠損: {len(missing)}件')
for f in missing: print(f'  MISSING: {f}')
if not missing: print('全件OK — スキャン開始')
"

欠損ファイルが 1 件でもあった場合はスキャンを中止し、欠損一覧をユーザーに報告して終了せよ。
全件 OK の場合のみ Step 1 に進む。

【Step 1: スキャン + 起票実行】
tmp/BE/be_master_check.md を読み込み、完了条件3（起票）まで含めて全工程を実行せよ。

タスクファイルは docs/tasks/open/code-quality/TASK-{番号}-{kebab-case-title}.md に作成する。
タスク番号は docs/tasks/open/code-quality/ と docs/tasks/closed/code-quality/ の
既存ファイル名から最大番号を確認し、その +1 から採番する。

ultrathink

use AgentTeams with the following teams running in parallel:
- Team-Service: P1/P8/P10/P11/P13/P17 を tmp/BE/be_master_check.md の「Service」リストに対して検査
- Team-Repository-Master: P2/P3/P4/P9/P16 を tmp/BE/be_master_check.md の「Repository - マスタ系」リストに対して検査
- Team-Repository-Preload: P3 を tmp/BE/be_master_check.md の「Repository - 非マスタ系」リストに対して検査
- Team-Handler: P7/P12/P14/P15/P18 を tmp/BE/be_master_check.md の「Handler」リストに対して検査
- Team-Routes: P5/P6 を tmp/BE/be_master_check.md の「Routes」リストに対して検査

各チームは担当ファイルを全件 Read してから判定すること。
推測での OK/FAIL 出力は禁止。
全チーム完了後、違反サマリを集約して既存タスクとの重複チェックを行い、
未起票の違反のみを新規タスクとして起票する。
違反が 0 件だった場合は「違反なし」をユーザーに報告して終了せよ（起票不要）。
起票完了後、新規作成したタスクファイルのパス一覧をユーザーに報告して終了せよ。
```

---

### マスタ系 スキャン例B: 特定チームのみ再スキャン（修正後の確認）

```
【Step 0: 事前整合性チェック（必須・スキャン前に必ず実行）】
以下の Python スクリプトを Bash ツールで実行し、ドキュメントに記載された全ファイルが存在するか確認せよ。

python3 -c "
import re, os
with open('tmp/BE/be_master_check.md') as f:
    content = f.read()
files = re.findall(r'- (backend/internal/\S+\.go)', content)
missing = [f for f in files if not os.path.exists(f)]
print(f'登録ファイル数: {len(files)}')
print(f'欠損: {len(missing)}件')
for f in missing: print(f'  MISSING: {f}')
if not missing: print('全件OK — スキャン開始')
"

欠損ファイルが 1 件でもあった場合はスキャンを中止し、欠損一覧をユーザーに報告して終了せよ。
全件 OK の場合のみ Step 1 に進む。

【Step 1: 再スキャン実行】
tmp/BE/be_master_check.md のチェックリストとファイルリストを参照し、
以下のチームの担当スコープのみ再スキャンせよ。

対象チーム: Team-Repository-Preload（P3のみ）
対象ファイル: tmp/BE/be_master_check.md の「Repository - 非マスタ系」リスト

完了条件: PASS/FAIL 表の出力のみ（起票不要）

ultrathink
```

---

### 業務系 スキャン例A: スキャン + 起票（フル実行）

```
【Step 0: 事前整合性チェック（必須・スキャン前に必ず実行）】
以下の Python スクリプトを Bash ツールで実行し、ドキュメントに記載された全ファイルが存在するか確認せよ。

python3 -c "
import re, os
with open('tmp/BE/be_non_master_check.md') as f:
    content = f.read()
files = re.findall(r'- (backend/internal/\S+\.go)', content)
missing = [f for f in files if not os.path.exists(f)]
print(f'登録ファイル数: {len(files)}')
print(f'欠損: {len(missing)}件')
for f in missing: print(f'  MISSING: {f}')
if not missing: print('全件OK — スキャン開始')
"

欠損ファイルが 1 件でもあった場合はスキャンを中止し、欠損一覧をユーザーに報告して終了せよ。
全件 OK の場合のみ Step 1 に進む。

【Step 1: スキャン + 起票実行】
tmp/BE/be_non_master_check.md を読み込み、完了条件3（起票）まで含めて全工程を実行せよ。

タスクファイルは docs/tasks/open/code-quality/TASK-{番号}-{kebab-case-title}.md に作成する。
タスク番号は docs/tasks/open/code-quality/ と docs/tasks/closed/code-quality/ の
既存ファイル名から最大番号を確認し、その +1 から採番する。

ultrathink

use AgentTeams with the following teams running in parallel:
- Team-Service: P1/P8/P10/P11/P13/P17 を tmp/BE/be_non_master_check.md の「Service」リストに対して検査
- Team-Repository: P2/P3/P4/P9/P16 を tmp/BE/be_non_master_check.md の「Repository」リストに対して検査
- Team-Handler: P7/P12/P14/P15/P18 を tmp/BE/be_non_master_check.md の「Handler」リストに対して検査
- Team-Routes: P5/P6 を tmp/BE/be_non_master_check.md の「Routes」リストに対して検査

各チームは担当ファイルを全件 Read してから判定すること。
推測での OK/FAIL 出力は禁止。
全チーム完了後、違反サマリを集約して既存タスクとの重複チェックを行い、
未起票の違反のみを新規タスクとして起票する。
違反が 0 件だった場合は「違反なし」をユーザーに報告して終了せよ（起票不要）。
起票完了後、新規作成したタスクファイルのパス一覧をユーザーに報告して終了せよ。
```

---

### 業務系 スキャン例B: 特定チームのみ再スキャン（修正後の確認）

```
【Step 0: 事前整合性チェック（必須・スキャン前に必ず実行）】
以下の Python スクリプトを Bash ツールで実行し、ドキュメントに記載された全ファイルが存在するか確認せよ。

python3 -c "
import re, os
with open('tmp/BE/be_non_master_check.md') as f:
    content = f.read()
files = re.findall(r'- (backend/internal/\S+\.go)', content)
missing = [f for f in files if not os.path.exists(f)]
print(f'登録ファイル数: {len(files)}')
print(f'欠損: {len(missing)}件')
for f in missing: print(f'  MISSING: {f}')
if not missing: print('全件OK — スキャン開始')
"

欠損ファイルが 1 件でもあった場合はスキャンを中止し、欠損一覧をユーザーに報告して終了せよ。
全件 OK の場合のみ Step 1 に進む。

【Step 1: 再スキャン実行】
tmp/BE/be_non_master_check.md のチェックリストとファイルリストを参照し、
以下のチームの担当スコープのみ再スキャンせよ。

対象チーム: Team-Repository（P2/P3/P4/P9/P16）
対象ファイル: tmp/BE/be_non_master_check.md の「Repository」リスト

完了条件: PASS/FAIL 表の出力のみ（起票不要）

ultrathink
```

---

### 実装例A: パターン別に AgentTeams で並列実装

ファイル競合が起きないよう、担当ファイルが重複しないチーム編成にする。

```
docs/tasks/open/code-quality/ の全タスクファイルを Read し、
以下のチーム編成ルールに従って AgentTeams で並列実装せよ。

## 事前準備（実装前に必ず実施）
1. docs/tasks/open/code-quality/ の全ファイルを Read して対象タスクを把握する
2. フロントマターの `pattern:` フィールドが P1〜P18 のいずれかであるタスクのみを対象とする
   - pattern が P1〜P18 以外（例: BUG-xxx）または `status: partial` のタスクは除外する
3. 各タスクの「対象ファイル」を確認し、同一ファイルへの変更が重複するタスクを
   同一チームに割り当てる（チーム間のファイル競合を防ぐ）
4. 全タスクが以下のいずれかのチームに割り当てられたことを確認する。
   どのチームにも当てはまらないタスクがある場合はユーザーに報告して確認を取ってから進む

## チーム編成ルール（対象ファイルの種別で分割）

### Team-Routes
担当: P5/P6 — handler/*_handler.go の Register*Routes 関数（パーミッション設定）修正タスク全件

### Team-Repository-DeletedAt
担当: P2 — repository の CountUsage/CountBy* メソッドの deleted_at IS NULL 修正タスク全件

### Team-Repository-Preload
担当: P3 — repository の Preload deleted_at IS NULL 修正タスク全件

### Team-Repository-Scope
担当: P4 — repository の clinicScope 欠落修正タスク全件

### Team-Repository-Naming
担当: P9/P16 — repository の apperrors.FromGORM 未使用・メソッド名統一 修正タスク全件

### Team-Service-FindByID
担当: P1 — service の Delete/Update で FindByID 前置が必要なタスク全件

### Team-Service-Error
担当: P8/P10/P11 — service の apperrors.Wrap・FK依存チェック・slog.ErrorContext 修正タスク全件

### Team-Service-Naming
担当: P13/P17 — service の const/buildFunc 定義順序・Input 構造体命名統一 修正タスク全件

### Team-Handler
担当: P7/P12/P14/P15 — handler のレスポンス変換・ShouldBindJSON・repository直接呼出禁止・Location ヘッダ修正タスク全件

### Team-Handler-Naming
担当: P18 — handler の toXxxResponse 関数名統一 修正タスク全件

## 実装手順（各チーム共通）
1. 担当タスクファイルを全件 Read して「あるべき姿」を把握する
   - 担当タスクが 0 件の場合はその旨を報告して終了せよ（何もしない）
2. 対象ソースファイルを Read して現状を確認する
3. Edit ツールで修正する（あるべき姿以外の変更禁止）
4. 全タスク修正完了後、ユーザーに以下の手動実行を依頼する（自動実行禁止）:
   docker compose exec backend go test ./backend/internal/...
5. ユーザーからテスト結果を受け取ったら:
   - PASS → 担当タスク全件をまとめて 1 コミットし、タスクファイルを docs/tasks/closed/code-quality/ に移動する
   - FAIL → エラーログを確認して修正し、再度手動実行を依頼する。修正してもFAILが続く場合はユーザーに報告して止まる

## 禁止事項
- タスクファイルを読まずに実装しない
- 担当外のファイルを変更しない（チーム間の競合防止）
- あるべき姿以外の箇所を変更しない
- テスト結果未確認のままクローズしない
- テストFAILのままコミットしない

ultrathink

use AgentTeams.
```

---

### 実装例B: 全 open タスクを優先度順に一括処理

```
docs/tasks/open/code-quality/ の全タスクを優先度順（Critical → High → Medium → Low）に
すべて実装してクローズせよ。

## 実装手順
1. docs/tasks/open/code-quality/ のファイル一覧を取得する
2. 各タスクファイルを Read して「優先度」「対象ファイル」「あるべき姿」を把握する
   - フロントマターの `pattern:` が P1〜P18 のタスクのみを対象とする
   - `status: partial` のタスクは除外する（partial_note を確認し未実装箇所をユーザーに確認する）
3. 対象タスクが 0 件の場合はその旨をユーザーに報告して終了せよ
4. 優先度でグルーピングし、Critical から順番に実装する
5. 同一優先度のタスクは担当ファイルが重複しないものを AgentTeams で並列実装する
6. 各グループの全タスク修正完了後、ユーザーに以下の手動実行を依頼する（自動実行禁止）:
   docker compose exec backend go test ./backend/internal/...
7. ユーザーからテスト結果を受け取ったら:
   - PASS → そのグループのタスクをまとめて 1 コミットし、docs/tasks/closed/code-quality/ に移動してから次のグループへ進む
   - FAIL → エラーログを確認して修正し、再度手動実行を依頼する。修正してもFAILが続く場合はユーザーに報告して止まる
8. 全グループ完了後、docs/tasks/open/code-quality/ に残っている P1〜P18 タスクがないことを確認し、ユーザーに完了報告する

## 並列実行の制約
- 同一ファイルへの変更が発生するタスクは順番に実行する（競合防止）
- 異なるファイルへの変更は並列実行してよい

## 禁止事項
- タスクに書かれた「あるべき姿」以外の修正を行わない
- テスト結果未確認のままタスクをクローズしない
- テストFAILのままコミットしない
- `status: partial` のタスクをフルクローズしない（partial_note を必ず読むこと）
- 既に closed/ にあるタスクを再実装しない

ultrathink

use AgentTeams.
```

---

### 単一ファイルチェック（新規実装後の即時確認）

```
backend/internal/{layer}/{name}.go を読み込み、
該当レイヤーの適用パターン（下表）を全て検査し PASS/FAIL を報告せよ。

| レイヤー     | 適用パターン                        |
|------------|-----------------------------------|
| handler    | P7, P12, P14, P15, P18            |
| service    | P1, P8, P10, P11, P13, P17        |
| repository | P2, P3, P4, P9, P16               |
| routes     | P5, P6                            |

各パターンの判定基準は .claude/refs/gin-architecture-compliance.md を参照せよ。
推測での OK/FAIL 出力は禁止。コードを Read してから判定すること。
```

---

## Quick Reference

| Pattern | Layer | What to Check |
|---------|-------|---------------|
| P1 | Service | FindByID before Delete/Update |
| P2 | Repository | `deleted_at IS NULL` in CountUsage |
| P3 | Repository | `"deleted_at IS NULL"` in Preload |
| P4 | Repository | `clinicScope` on Update/Upsert |
| P5 | Routes | RequirePermission on POST/PUT/PATCH/DELETE |
| P6 | Routes | "delete" permission on DELETE routes |
| P7 | Handler | `toXxxResponse()` wrapping in c.JSON |
| P8 | Service | `apperrors.Wrap` on all error returns |
| P9 | Repository | `apperrors.FromGORM` on all GORM errors |
| P10 | Service | FK dependency check before Delete |
| P11 | Service | `slog.ErrorContext` before repo error return |
| P12 | Handler | `RespondError(c, apperrors.WrapInvalidInput(...))` for bind errors |
| P13 | Service | const → buildFunc → interface → struct → New → methods |
| P14 | Handler | No direct repository injection in handler |
| P15 | Handler | 201 + Location header on Create |
| P16 | Repository | FindAll/FindByID/Create/Update/Delete/CountBy |
| P17 | Service | `CreateXxxInput` / `UpdateXxxInput` naming |
| P18 | Handler | `toXxxResponse` / `toXxxListResponse` naming |
