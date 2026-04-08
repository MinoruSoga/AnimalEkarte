# BUG-122: バックエンド API の書き込み操作に権限チェックミドルウェアが未適用

## 概要
バックエンド API の POST/PATCH/DELETE/PUT エンドポイント（32箇所）に `RequirePermission` ミドルウェアが
適用されていない。`permission-groups` の書き込み操作のみ適用済みだが、それ以外の全リソースは
JWT 認証チェックのみで、RBAC 権限チェック（authorization）がない。

フロントエンドの `RequirePermission` ガードをバイパスして直接 API を叩けば、
一般権限ユーザーでも任意のリソースを作成・更新・削除可能。

## 脆弱性分類
- **CWE-862**: Missing Authorization
- **OWASP A01:2021**: Broken Access Control
- **影響**: 認証済みユーザーが権限のないリソースを改ざん・削除可能

## 再現手順
1. `vet@example.com` / `password`（一般権限）でログイン
2. ブラウザ DevTools の Application タブで Cookie から JWT トークンを取得
3. 以下の curl コマンドを実行:
```bash
# 一般権限ユーザーでも会計データを作成可能（本来は create 権限なし）
curl -X POST http://localhost:8080/api/v1/owners \
  -H "Cookie: auth_token=<JWT>" \
  -H "Content-Type: application/json" \
  -d '{"name": "不正作成テスト", "phone": "000-0000-0000"}'
```
4. **結果**: 権限チェックなしでリソースが作成される（200 OK）

## 期待する動作
- 各エンドポイントの POST/PATCH/DELETE に対応するリソース・アクションの権限チェックを実施
- 権限がない場合は `403 Forbidden` + `{"error": "forbidden"}` を返す

## 現状の RequirePermission 実装

### `backend/internal/handler/response.go:177-188`
```go
func (h *Handler) RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !h.hasPermission(c, resource, action) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}
```

### `backend/internal/handler/clinic_handler.go:50-83` — hasPermission 実装
```go
func (h *Handler) hasPermission(c *gin.Context, resource, action string) bool {
	isSystemAdmin, ok := extractIsSystemAdmin(c)
	if ok && isSystemAdmin { return true }  // system_admin は全権限バイパス
	// permission_group_rules テーブルから権限判定
	staffID, ok := extractStaffID(c)
	if !ok { return false }
	rules, err := h.repos.PermissionGroup.GetEffectivePermissionsByStaffID(ctx, staffID)
	// ... resource + action マッチング
}
```

## 正しい実装例（唯一の適用済みエンドポイント）

### `backend/internal/handler/staff_handler.go:399-409`
```go
// Permission Group — GET は誰でも可能
masters.GET("/permission-groups", h.ListPermissionGroups)
masters.GET("/permission-groups/:id", h.GetPermissionGroup)

// ✅ 書き込み操作には RequirePermission ミドルウェアを適用
pgWrite := masters.Group("/permission-groups")
pgWrite.Use(h.RequirePermission(string(model.ResourceMasterPermission), "edit"))
pgWrite.POST("", h.CreatePermissionGroup)
pgWrite.PATCH("", h.ReorderPermissionGroups)
pgWrite.PATCH("/:id", h.UpdatePermissionGroup)
pgWrite.DELETE("/:id", h.DeletePermissionGroup)
pgWrite.PUT("/:id/rules", h.SetPermissionGroupRules)
```

## 権限チェック未適用のエンドポイント一覧（32箇所）

### メインルート登録（`backend/internal/handler/handler.go:46-83`）

```go
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	protected := api.Group("")
	protected.Use(middleware.Auth(h.cfg.JWTSecret))          // ← 認証のみ
	protected.Use(middleware.SanitizeNullBytes())

	h.registerOwnerRoutesWithAuth(protected)                 // write に RequirePermission なし
	h.RegisterPetRoutes(protected)                           // write に RequirePermission なし
	h.RegisterReservationRoutes(protected)                   // write に RequirePermission なし
	h.registerMedicalRecordRoutesWithAuth(protected)         // write に RequirePermission なし
	h.registerHospitalizationRoutesWithAuth(protected)       // write に RequirePermission なし
	h.registerAccountingRoutesWithAuth(protected)            // write に RequirePermission なし
	h.registerTrimmingRoutesWithAuth(protected)              // write に RequirePermission なし
	h.registerExaminationRoutesWithAuth(protected)           // write に RequirePermission なし
	h.registerVaccinationRoutesWithAuth(protected)           // write に RequirePermission なし
	h.registerInventoryRoutesWithAuth(protected)             // write に RequirePermission なし
	h.registerMasterRoutesWithAuth(protected)                // permission-groups 以外は RequirePermission なし
	h.RegisterClinicRoutes(protected)                        // write に RequirePermission なし
	h.registerEstimateRoutesWithAuth(protected)              // write に RequirePermission なし
	h.RegisterShiftRoutes(protected)                         // write に RequirePermission なし
	h.RegisterCompanyRoutes(protected)                       // write に RequirePermission なし
	h.RegisterGlobalCheckupRoutes(protected)                 // write に RequirePermission なし
	h.RegisterBillingItemRoutes(protected)                   // write に RequirePermission なし
}
```

### 1. Pet Routes — `pet_handler.go:178-185`
```go
func (h *Handler) RegisterPetRoutes(rg *gin.RouterGroup) {
	pets := rg.Group("/pets")
	pets.GET("", h.ListPets)
	pets.POST("", h.CreatePet)           // ❌ RequirePermission なし
	pets.GET("/:id", h.GetPet)
	pets.PATCH("/:id", h.UpdatePet)      // ❌ RequirePermission なし
	pets.DELETE("/:id", h.DeletePet)     // ❌ RequirePermission なし
}
```
**リソース**: `owners`（ペットは飼主の子リソース）

### 2. Reservation Routes — `reservation_handler.go:296-303`
```go
func (h *Handler) RegisterReservationRoutes(rg *gin.RouterGroup) {
	reservations := rg.Group("/reservations")
	reservations.GET("", h.ListReservations)
	reservations.POST("", h.CreateReservation)      // ❌
	reservations.GET("/:id", h.GetReservation)
	reservations.PATCH("/:id", h.UpdateReservation) // ❌
	reservations.DELETE("/:id", h.DeleteReservation) // ❌
}
```
**リソース**: `reservations`

### 3. Clinic Routes — `clinic_handler.go:228-235`
```go
func (h *Handler) RegisterClinicRoutes(rg *gin.RouterGroup) {
	clinics := rg.Group("/clinics")
	clinics.GET("", h.ListClinics)
	clinics.POST("", h.CreateClinic)        // ❌
	clinics.GET("/:id", h.GetClinic)
	clinics.PATCH("/:id", h.UpdateClinic)   // ❌
	clinics.DELETE("/:id", h.DeleteClinic)  // ❌
}
```
**リソース**: `hospital-settings`（`ResourceHospitalSettings`）

### 4. Company Routes — `company_handler.go:50-53`
```go
func (h *Handler) RegisterCompanyRoutes(rg *gin.RouterGroup) {
	rg.GET("/company", h.GetCompany)
	rg.PATCH("/company", h.UpdateCompany)   // ❌
}
```
**リソース**: `hospital-settings`

### 5. Shift Routes — `shift_handler.go:141-147`
```go
func (h *Handler) RegisterShiftRoutes(rg *gin.RouterGroup) {
	shifts := rg.Group("/shifts")
	shifts.GET("", h.ListShiftEntries)
	shifts.POST("", h.CreateShiftEntry)             // ❌
	shifts.PATCH("/:id", h.UpdateShiftEntry)        // ❌
	shifts.DELETE("/:id", h.DeleteShiftEntry)       // ❌
}
```
**リソース**: `shifts`

### 6. Billing Item Routes — `billing_item_handler.go:168-173`
```go
func (h *Handler) RegisterBillingItemRoutes(rg *gin.RouterGroup) {
	items := rg.Group("/billing-items")
	items.POST("", h.CreateBillingItem)      // ❌
	items.PATCH("/:id", h.UpdateBillingItem) // ❌
	items.DELETE("/:id", h.DeleteBillingItem) // ❌
}
```
**リソース**: `accounting`

### 7. Checkup Type Routes — `checkup_handler.go:187-199`
```go
func (h *Handler) RegisterGlobalCheckupRoutes(rg *gin.RouterGroup) {
	checkups := rg.Group("/checkup-types")
	checkups.GET("", h.ListCheckupTypes)
	checkups.POST("", h.CreateCheckupType)           // ❌
	checkups.GET("/:id", h.GetCheckupType)
	checkups.PATCH("/:id", h.UpdateCheckupType)      // ❌
	checkups.DELETE("/:id", h.DeleteCheckupType)     // ❌
}
```
**リソース**: `checkups`

### 8. Vital Routes — `vital_handler.go:177-182`
```go
func (h *Handler) RegisterVitalRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/vitals", h.ListVitals)
	rg.POST("/:id/vitals", h.CreateVital)                  // ❌
	rg.PATCH("/:id/vitals/:vitalId", h.UpdateVital)        // ❌
	rg.DELETE("/:id/vitals/:vitalId", h.DeleteVital)       // ❌
}
```
**リソース**: `medical-records`

### 9. Daily Record Routes — `daily_record_handler.go:225-232`
```go
func (h *Handler) RegisterDailyRecordRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/daily-records", h.ListDailyRecords)
	rg.POST("/:id/daily-records", h.CreateDailyRecord)                    // ❌
	rg.GET("/:id/daily-records/:date", h.GetDailyRecord)
	rg.POST("/:id/daily-records/:date/vitals", h.AddVitalRecord)          // ❌
	rg.POST("/:id/daily-records/:date/care-logs", h.AddCareLogRecord)     // ❌
	rg.POST("/:id/daily-records/:date/staff-notes", h.AddStaffNoteRecord) // ❌
}
```
**リソース**: `hospitalization`

### 10. Care Plan Item Routes — `care_plan_item_handler.go:143-160`
```go
func (h *Handler) RegisterCarePlanItemRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/care-plan-items", h.ListCarePlanItems)
	rg.POST("/:id/care-plan-items", h.CreateCarePlanItem)                    // ❌
	rg.GET("/:id/care-plan-items/:itemID", h.GetCarePlanItem)
	rg.PATCH("/:id/care-plan-items/:itemID", h.UpdateCarePlanItem)           // ❌
	rg.DELETE("/:id/care-plan-items/:itemID", h.DeleteCarePlanItem)          // ❌
	rg.PUT("/:id/care-plan-items/:itemID/execute", h.ExecuteCarePlanItem)    // ❌
	rg.PUT("/:id/care-plan-items/:itemID/unexecute", h.UnexecuteCarePlanItem) // ❌
}
```
**リソース**: `hospitalization`

### 11. Checkup Routes (in Medical Records) — `checkup_handler.go:193-204`
```go
func (h *Handler) RegisterCheckupRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/checkups", h.ListCheckups)
	rg.POST("/:id/checkups", h.CreateCheckup)              // ❌
	rg.GET("/:id/checkups/:checkupID", h.GetCheckup)
	rg.PATCH("/:id/checkups/:checkupID", h.UpdateCheckup)  // ❌
	rg.DELETE("/:id/checkups/:checkupID", h.DeleteCheckup) // ❌
}
```
**リソース**: `medical-records`

### 12. `WithAuth` 系ルート（owners, medical-records, hospitalization 等）

`registerOwnerRoutesWithAuth`, `registerMedicalRecordRoutesWithAuth` 等も同様に
POST/PATCH/DELETE に `RequirePermission` が適用されていない。
"WithAuth" という名前は JWT 認証を意味するだけで、RBAC 権限チェックではない。

## リソースマッピング

| エンドポイントグループ | Resource 定数 | 文字列 |
|---------------------|--------------|--------|
| owners | `ResourceOwners` | `"owners"` |
| pets | `ResourceOwners` | `"owners"` |
| reservations | `ResourceReservations` | `"reservations"` |
| medical-records | `ResourceMedicalRecords` | `"medical-records"` |
| hospitalization | `ResourceHospitalization` | `"hospitalization"` |
| trimming | `ResourceTrimming` | `"trimming"` |
| examinations | `ResourceExaminations` | `"examinations"` |
| accounting | `ResourceAccounting` | `"accounting"` |
| vaccinations | `ResourceVaccinations` | `"vaccinations"` |
| inventory | `ResourceInventory` | `"inventory"` |
| estimates | `ResourceEstimates` | `"estimates"` |
| shifts | `ResourceShifts` | `"shifts"` |
| clinics / company | `ResourceHospitalSettings` | `"hospital-settings"` |
| checkup-types / checkups | `ResourceCheckups` | `"checkups"` |
| billing-items | `ResourceAccounting` | `"accounting"` |
| masters/* | 各マスタリソース | 各マスタ文字列 |

## 修正方針

各ハンドラのルート登録で、`staff_handler.go` の permission-groups パターンに準じて書き込み操作をガード:

```go
// 修正例: pet_handler.go
func (h *Handler) RegisterPetRoutes(rg *gin.RouterGroup) {
	pets := rg.Group("/pets")
	pets.GET("", h.ListPets)
	pets.GET("/:id", h.GetPet)

	// 書き込み操作に権限チェック
	pets.POST("", h.RequirePermission(string(model.ResourceOwners), "create"), h.CreatePet)
	pets.PATCH("/:id", h.RequirePermission(string(model.ResourceOwners), "edit"), h.UpdatePet)
	pets.DELETE("/:id", h.RequirePermission(string(model.ResourceOwners), "delete"), h.DeletePet)
}
```

または Group + Use パターン:
```go
petWrite := pets.Group("")
petWrite.Use(h.RequirePermission(string(model.ResourceOwners), "edit"))
petWrite.POST("", h.CreatePet)
petWrite.PATCH("/:id", h.UpdatePet)
petWrite.DELETE("/:id", h.DeletePet)
```

## 優先度
**Critical** — API レベルで RBAC をバイパス可能。認証済みユーザーが任意のデータを改ざん・削除できる。

## 関連チケット
- BUG-121: フロントエンド `/settings` 権限ガード
- BUG-123: `/settings` インデックスのマスタカード権限フィルタリング

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — バックエンド・アーキテクチャ規約
> **handler → service → repository の軽量レイヤードを徹底**

権限チェックは **handler 層（ミドルウェア）** の責務。Service/Repository 層に認可ロジックを
混入させてはならない。`RequirePermission` ミドルウェアをルート登録時に適用するのが正しいパターン。

### `.claude/CLAUDE.md` — エラー処理の統一 (MANDATORY)
> **Handler: `RespondError(c, err)` で統一レスポンス**
> **`c.JSON(http.StatusBadRequest, gin.H{"error": ...})` の直接使用は禁止**

現在の `RequirePermission` 実装（`response.go:177-188`）は `c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})` を
直接返している。本来は `apperrors.ErrForbidden` のようなセンチネルエラーを定義し、
`RespondError(c, err)` で統一レスポンスを返すべきだが、これは別途リファクタリング対象とする。
**現時点では既存の `RequirePermission` をそのまま適用することを優先する。**

### `.claude/rules/security.md` — Input Validation
> "Validate on both client and server"

クライアント側の `RequirePermission` ガード（BUG-121）だけでは不十分。
**サーバー側でも必ず認可チェックを実施する（多層防御）。**
フロントエンドは DevTools で改ざん可能であり、API は直接呼び出し可能であるため、
**バックエンド API が唯一の信頼できる認可境界**である。

### `.claude/rules/api.md` — Security
> "Validate all user input"
> "Use context for request-scoped values"

`RequirePermission` は `gin.Context` から `staff_id` と `is_system_admin` を取得し、
`permission_group_rules` テーブルでリソース×アクションの権限を判定する。
これは API ルール の「context for request-scoped values」パターンに準拠している。

### `.claude/rules/go-language.md` — Context 伝播（必須）
> 全関数の第一引数に `context.Context` を配置

`hasPermission()` 内部で `c.Request.Context()` を取得し、Repository 呼び出しに渡している。
新規ミドルウェア追加時もこのパターンを維持すること。

### `.claude/rules/error-handling.md` — HTTP Status マッピング
権限チェック失敗時は `403 Forbidden` を返す。これは以下のマッピングに準拠:
- 401 Unauthorized: 認証失敗（JWT 無効・期限切れ）
- 403 Forbidden: 認可失敗（権限不足）← **今回の対象**
- 404 Not Found: リソース未存在

### プロジェクト内参照実装
`backend/internal/handler/staff_handler.go:399-409` の permission-groups ルート登録が
唯一の正しい実装。**他の全ハンドラはこのパターンに統一すべき。**

```go
// ✅ 参照実装: permission-groups
pgWrite := masters.Group("/permission-groups")
pgWrite.Use(h.RequirePermission(string(model.ResourceMasterPermission), "edit"))
pgWrite.POST("", h.CreatePermissionGroup)
pgWrite.PATCH("", h.ReorderPermissionGroups)
pgWrite.PATCH("/:id", h.UpdatePermissionGroup)
pgWrite.DELETE("/:id", h.DeletePermissionGroup)
pgWrite.PUT("/:id/rules", h.SetPermissionGroupRules)
```

## 関連ファイル
- `backend/internal/handler/handler.go:46-83` — メインルート登録
- `backend/internal/handler/response.go:177-188` — RequirePermission 実装
- `backend/internal/handler/clinic_handler.go:50-83` — hasPermission 実装
- `backend/internal/handler/staff_handler.go:399-409` — 正しい実装パターン（参考）
- `backend/internal/model/permission.go` — Resource 定数定義
- 上記 11 ハンドラファイル — 各修正対象
