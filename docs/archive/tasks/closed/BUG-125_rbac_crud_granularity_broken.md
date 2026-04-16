# BUG-125: RBAC 権限チェックが CRUD 個別に機能していない

## 概要
BUG-122 修正（コミット `fa62e7cf`）で `RequirePermission` ミドルウェアが追加されたが、
**2つの重大な欠陥** があり、CRUD の個別制御が正しく機能していない。

### 欠陥1: POST/PATCH/DELETE が `"edit"` アクションで一括ガード
`WithAuth` 系ルート（owners, medical-records, hospitalization 等）で POST/PATCH/DELETE が
すべて `RequirePermission(resource, "edit")` で一括ガードされている。
`create` 権限と `delete` 権限が個別にチェックされない。

### 欠陥2: 7つのルートが未修正
`RegisterReservationRoutes`, `RegisterPetRoutes`, `RegisterShiftRoutes`,
`RegisterClinicRoutes`, `RegisterCompanyRoutes`, `RegisterGlobalCheckupRoutes`,
`RegisterBillingItemRoutes` — これらは `RequirePermission` が適用されていない。

## 脆弱性分類
- **CWE-863**: Incorrect Authorization
- **影響**: CRUD の粒度で権限制御ができない。create=true/edit=false のユーザーが作成不可、delete=false のユーザーが削除可能等

## 再現手順（RBAC検証用グループ ID=7 で検証済み）

### テスト設定
| リソース | view | create | edit | delete |
|---------|------|--------|------|--------|
| owners | T | **T** | **F** | F |
| reservations | T | **F** | T | **F** |
| inventory | T | F | F | **T** |

### 再現結果

| エンドポイント | 期待 | 実際 | 判定 |
|--------------|------|------|------|
| `POST /owners` | 許可（create=T） | **403** | ❌ edit=F で一括ブロック |
| `PATCH /owners/1` | 拒否（edit=F） | 403 | ✅ |
| `DELETE /owners/1` | 拒否（delete=F） | 403 | ✅（偶然。edit=F で一括ブロック） |
| `POST /reservations` | 拒否（create=F） | **400**（認可通過） | ❌ ガードなし |
| `DELETE /reservations/1` | 拒否（delete=F） | **204**（削除成功） | ❌ ガードなし |
| `DELETE /inventory/1` | 許可（delete=T） | **403** | ❌ edit=F で一括ブロック |

## 現状コード

### 欠陥1: 一括ガード — `backend/internal/handler/handler.go:86-97`
```go
func (h *Handler) registerOwnerRoutesWithAuth(rg *gin.RouterGroup) {
	owners := rg.Group("/owners")
	owners.GET("", h.ListOwners)
	owners.GET("/:id", h.GetOwner)

	ownerWrite := owners.Group("")
	ownerWrite.Use(h.RequirePermission(string(model.ResourceOwners), "edit"))  // ❌ 全部 "edit"
	ownerWrite.POST("", h.CreateOwner)     // ❌ create ではなく edit でチェック
	ownerWrite.PATCH("/:id", h.UpdateOwner)
	ownerWrite.DELETE("/:id", h.DeleteOwner)  // ❌ delete ではなく edit でチェック
}
```

同じパターンが全 `WithAuth` ルートに適用されている:
- `registerOwnerRoutesWithAuth` (handler.go:86)
- `registerMedicalRecordRoutesWithAuth` (handler.go:99)
- `registerHospitalizationRoutesWithAuth` (handler.go:121)
- `registerTrimmingRoutesWithAuth` (handler.go:139)
- `registerExaminationRoutesWithAuth` (handler.go:152)
- `registerVaccinationRoutesWithAuth` (handler.go:165)
- `registerAccountingRoutesWithAuth` (handler.go:178)
- `registerInventoryRoutesWithAuth` (handler.go:193)
- `registerEstimateRoutesWithAuth` (handler.go:213)

### 欠陥2: 未修正ルート — `backend/internal/handler/handler.go:68-83`
```go
h.RegisterPetRoutes(protected)            // ❌ RequirePermission なし
h.RegisterReservationRoutes(protected)    // ❌ RequirePermission なし
h.RegisterClinicRoutes(protected)         // ❌ RequirePermission なし
h.RegisterShiftRoutes(protected)          // ❌ RequirePermission なし
h.RegisterCompanyRoutes(protected)        // ❌ RequirePermission なし
h.RegisterGlobalCheckupRoutes(protected)  // ❌ RequirePermission なし
h.RegisterBillingItemRoutes(protected)    // ❌ RequirePermission なし
```

## 修正方針

### POST/PATCH/DELETE を個別アクションでガード

```go
func (h *Handler) registerOwnerRoutesWithAuth(rg *gin.RouterGroup) {
	owners := rg.Group("/owners")
	owners.GET("", h.ListOwners)
	owners.GET("/:id", h.GetOwner)

	// ✅ 各操作に対応するアクションで個別ガード
	owners.POST("",
		h.RequirePermission(string(model.ResourceOwners), "create"),
		h.CreateOwner)
	owners.PATCH("/:id",
		h.RequirePermission(string(model.ResourceOwners), "edit"),
		h.UpdateOwner)
	owners.DELETE("/:id",
		h.RequirePermission(string(model.ResourceOwners), "delete"),
		h.DeleteOwner)
}
```

### 未修正ルートも同様に修正

```go
// handler.go
h.registerPetRoutesWithAuth(protected)
h.registerReservationRoutesWithAuth(protected)
h.registerClinicRoutesWithAuth(protected)
h.registerShiftRoutesWithAuth(protected)
h.registerCompanyRoutesWithAuth(protected)
h.registerGlobalCheckupRoutesWithAuth(protected)
h.registerBillingItemRoutesWithAuth(protected)
```

各ハンドラファイルの `RegisterXxxRoutes` を `registerXxxRoutesWithAuth` にリネームし、
上記と同じ個別ガードパターンを適用。

### 子ルート（Vitals, DailyRecords 等）も同様

`RegisterVitalRoutes`, `RegisterDailyRecordRoutes` 等の子ルートも
親リソースのアクションで個別ガードする:

```go
func (h *Handler) RegisterVitalRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/vitals", h.ListVitals)
	rg.POST("/:id/vitals",
		h.RequirePermission(string(model.ResourceMedicalRecords), "edit"),
		h.CreateVital)
	// ...
}
```

## 修正対象ファイル一覧

| ファイル | 修正内容 |
|---------|---------|
| `backend/internal/handler/handler.go:86-230` | 全 `WithAuth` ルートの一括ガード → 個別ガードに変更 |
| `backend/internal/handler/handler.go:68-83` | 7つの未修正ルートを `WithAuth` 版に変更 |
| `backend/internal/handler/pet_handler.go:179` | `RegisterPetRoutes` → 個別ガード追加 |
| `backend/internal/handler/reservation_handler.go:296` | `RegisterReservationRoutes` → 個別ガード追加 |
| `backend/internal/handler/clinic_handler.go:228` | `RegisterClinicRoutes` → 個別ガード追加 |
| `backend/internal/handler/shift_handler.go:141` | `RegisterShiftRoutes` → 個別ガード追加 |
| `backend/internal/handler/company_handler.go:51` | `RegisterCompanyRoutes` → 個別ガード追加 |
| `backend/internal/handler/checkup_handler.go:188` | `RegisterGlobalCheckupRoutes` → 個別ガード追加 |
| `backend/internal/handler/billing_item_handler.go:168` | `RegisterBillingItemRoutes` → 個別ガード追加 |
| `backend/internal/handler/vital_handler.go:177` | 子ルートに個別ガード追加 |
| `backend/internal/handler/daily_record_handler.go:225` | 子ルートに個別ガード追加 |
| `backend/internal/handler/care_plan_item_handler.go:143` | 子ルートに個別ガード追加 |
| `backend/internal/handler/treatment_handler.go` | 子ルートに個別ガード追加 |
| `backend/internal/handler/treatment_plan_handler.go` | 子ルートに個別ガード追加 |
| `backend/internal/handler/clinical_plan_handler.go` | 子ルートに個別ガード追加 |
| `backend/internal/handler/billing_review_handler.go` | 子ルートに個別ガード追加 |
| `backend/internal/handler/record_image_handler.go` | 子ルートに個別ガード追加 |
| `backend/internal/handler/inquiry_handler.go` | 子ルートに個別ガード追加 |

## テスト検証方法と実施結果

RBAC検証用グループ（ID=7、八王子院 clinic_id=3）で48エンドポイントを網羅テスト。
`vet@example.com` に RBAC検証用グループを割り当てて実施。

### 実施結果: 48テスト中 34 PASS / 14 FAIL

#### パターンA: ALLOW 期待 → 403（edit 一括ガードで create/delete が誤ブロック）— 7件

| テスト | 権限 | 期待 | 実際 | 原因 |
|--------|------|------|------|------|
| `POST /owners` | C=T, E=**F** | ALLOW | **403** | POST が `edit` でガード |
| `POST /pets` | C=T, E=**F** | ALLOW | **403** | POST が `edit` でガード |
| `POST /vaccinations` | C=T, E=**F** | ALLOW | **403** | POST が `edit` でガード |
| `POST /estimates` | C=T, E=**F** | ALLOW | **403** | POST が `edit` でガード |
| `POST /masters/insurances` | C=T, E=**F** | ALLOW | **403** | POST が `edit` でガード |
| `DELETE /inventory/1` | D=T, E=**F** | ALLOW | **403** | DELETE が `edit` でガード |
| `DELETE /masters/insurances/1` | D=T, E=**F** | ALLOW | **403** | DELETE が `edit` でガード |

#### パターンB: DENY 期待 → 通過（ガードなし or edit=T で通過し create/delete チェックなし）— 7件

| テスト | 権限 | 期待 | 実際 | 原因 |
|--------|------|------|------|------|
| `POST /reservations` | C=**F** | DENY | **400**（通過） | ガードなし（未修正ルート） |
| `DELETE /reservations/1` | D=**F** | DENY | **404**（通過） | ガードなし（未修正ルート） |
| `POST /examinations` | C=**F**, E=T | DENY | **400**（通過） | edit=T で通過、create チェックなし |
| `DELETE /examinations/1` | D=**F**, E=T | DENY | **404**（通過） | edit=T で通過、delete チェックなし |
| `DELETE /trimmings/1` | D=**F**, E=T | DENY | **204**（削除成功） | edit=T で通過、delete チェックなし |
| `DELETE /masters/staffs/1` | D=**F**, E=T | DENY | **409**（通過） | edit=T で通過、delete チェックなし |

#### PASS したテスト — 34件

| リソース | POST(create) | PATCH(edit) | DELETE(delete) |
|---------|-------------|-------------|----------------|
| owners (C=T,E=F,D=F) | ❌ 403 | ✅ 403 | ✅ 403 |
| pets (C=T,E=F,D=F) | ❌ 403 | ✅ 403 | ✅ 403 |
| reservations (C=F,E=T,D=F) | ❌ 400 | ✅ 400 | ❌ 404 |
| medical-records (C=T,E=T,D=T) | ✅ 400 | ✅ 400 | ✅ 204 |
| accounting (C=F,E=F,D=F) | ✅ 403 | ✅ 403 | ✅ 403 |
| hospitalization (V=F,C=F,E=F,D=F) | ✅ 403 | — | ✅ 403 |
| shifts (C=T,E=T,D=T) | ✅ 400 | ✅ 400 | ✅ 404 |
| inventory (C=F,E=F,D=T) | ✅ 403 | ✅ 403 | ❌ 403 |
| vaccinations (C=T,E=F,D=F) | ❌ 403 | ✅ 403 | ✅ 403 |
| examinations (C=F,E=T,D=F) | ❌ 400 | ✅ 404 | ❌ 404 |
| trimming (C=T,E=T,D=F) | ✅ 201 | ✅ 200 | ❌ 204 |
| estimates (C=T,E=F,D=F) | ❌ 403 | ✅ 403 | ✅ 403 |
| master-staff (C=T,E=T,D=F) | ✅ 201 | ✅ 200 | ❌ 409 |
| master-permission (C=F,E=F,D=F) | ✅ 403 | — | ✅ 403 |
| master-insurance (C=T,E=F,D=T) | ❌ 403 | ✅ 403 | ❌ 403 |
| master-merchandise (C=F,E=F,D=F) | ✅ 403 | — | ✅ 403 |
| clinic (E=F) | — | ✅ 403 | ✅ 403 |
| company (E=F) | — | ✅ 403 | — |

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — バックエンド・アーキテクチャ規約
> handler → service → repository の軽量レイヤードを徹底

権限チェックは handler 層（ミドルウェア）の責務。HTTP メソッドと RBAC アクションは 1:1 で対応すべき:
- POST → `"create"`
- PATCH/PUT → `"edit"`
- DELETE → `"delete"`

### `.claude/rules/api.md` — Security
> "Validate all user input"
> "Use proper HTTP status codes"

HTTP メソッドに対応する正しいアクションで認可チェックを行い、不正アクセスには 403 を返す。

### `.claude/rules/security.md` — Input Validation
> "Validate on both client and server"

CRUD の粒度で権限を設計している以上、バックエンドでも同じ粒度でチェックしなければ、
権限設計の意味がない。

### プロジェクト内参照実装
`staff_handler.go:399-409` の permission-groups ルートは `"edit"` で一括ガードしているが、
これは permission-groups が create/edit/delete を分離する必要がないケース（管理者のみ操作）であり、
一般リソースには適用すべきでない。

## 優先度
**Critical** — RBAC の粒度制御が機能していない。権限設定画面で設定した create/delete の個別権限が
API レベルで反映されない。

## 関連チケット
- BUG-122（修正済みだが不完全）: バックエンド API 権限チェック
- BUG-124: マスタページ操作ボタンの表示制御（フロントエンド UI）
