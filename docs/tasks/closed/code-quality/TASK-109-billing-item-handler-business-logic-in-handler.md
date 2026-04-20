# TASK-109: `billing_item_handler.go` — Handler 層にビジネスロジック混在

## 優先度

**High** — テナント所有権確認・デフォルト値設定・ドメインバリデーションが Handler 層に存在し、責務分離が崩れている。Service を通さずにロジックが変更された場合、セキュリティチェックが迂回されるリスクがある。

---

## 概要

`billing_item_handler.go` の `CreateBillingItem` ハンドラ（行 26-56）が以下の 3 つのビジネスロジックを担っている:

1. **テナント所有権確認**（行 27）: `h.svc.Accounting.GetByID` で billing が同一 clinic に属するか確認
2. **デフォルト値設定**（行 32-50）: `TaxType="excluded"`, `TaxRate=0.10`, `Source="manual"` のデフォルト値設定
3. **ドメインバリデーション**（行 34,46,53）: `validateTaxType`, `validateItemSource`, `validateItemCategory` の呼び出し

これらはすべて Service 層（`billing_item_service.go`）の責務である。Handler は HTTP 入力の
解析・バリデーション（`ShouldBindJSON`）と Service への委譲のみを担うべき。

---

## 問題箇所

### `handler/billing_item_handler.go:26-70`

```go
// ❌ テナント所有権確認が Handler に存在
// テナント分離: billing が同一クリニックに属することを確認
if _, err := h.svc.Accounting.GetByID(c.Request.Context(), clinicID, req.BillingID); err != nil {
    RespondError(c, err)
    return
}

// ❌ デフォルト値設定が Handler に存在
taxType := model.TaxTypeExcluded
if req.TaxType != "" {
    if err := validateTaxType(req.TaxType); err != nil {
        RespondError(c, err)
        return
    }
    taxType = model.TaxType(req.TaxType)
}
taxRate := 0.10
if req.TaxRate > 0 {
    taxRate = req.TaxRate
}
source := model.ItemSourceManual
if req.Source != "" {
    if err := validateItemSource(req.Source); err != nil {
        RespondError(c, err)
        return
    }
    source = model.ItemSource(req.Source)
}

// ❌ ドメインバリデーションが Handler に存在
if err := validateItemCategory(req.Category); err != nil {
    RespondError(c, err)
    return
}
```

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ handler/vaccine_handler.go — Handler はリクエスト解析と委譲のみ
func (h *Handler) CreateVaccine(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok { return }
    var req createVaccineRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
        return
    }
    vaccine, err := h.svc.Vaccine.Create(c.Request.Context(), clinicID, &service.CreateVaccineInput{
        Name:     req.Name,
        Species:  req.Species,
        ...
    })
    ...
}

// ✅ service/vaccine_service.go — バリデーション・デフォルト値は Service で処理
func (s *vaccineService) Create(ctx context.Context, clinicID uint64, input *CreateVaccineInput) (*model.Vaccine, error) {
    if err := validateRequiredName(input.Name); err != nil { return nil, err }
    if input.Species != nil {
        if err := validateVaccineSpecies(*input.Species); err != nil { return nil, err }
    }
    ...
}
```

---

## 修正方針

### 1. `handler/billing_item_handler.go` — シンプル化

Handler は入力解析のみ行い、raw 値をそのまま Service に渡す。

```go
// ✅ 修正後
func (h *Handler) CreateBillingItem(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok { return }

    var req createBillingItemRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
        return
    }

    item, err := h.svc.BillingItem.CreateItem(c.Request.Context(), &service.CreateBillingItemInput{
        ClinicID:              clinicID,
        BillingID:             req.BillingID,
        Category:              req.Category,    // string のまま渡す（変換は Service で）
        Name:                  req.Name,
        UnitPrice:             req.UnitPrice,
        Quantity:              req.Quantity,
        TaxType:               req.TaxType,     // "" のまま渡す（デフォルトは Service で）
        TaxRate:               req.TaxRate,     // 0 のまま渡す（デフォルトは Service で）
        IsInsuranceApplicable: req.IsInsuranceApplicable,
        Source:                req.Source,      // "" のまま渡す（デフォルトは Service で）
        SortOrder:             req.SortOrder,
    })
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusCreated, toBillingItemResponse(item))
}
```

### 2. `service/billing_item_service.go` — CreateItem にロジック移動

`CreateItem` に以下を追加:

```go
func (s *billingItemService) CreateItem(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error) {
    // ✅ テナント所有権確認を Service に移動
    if _, err := s.billingRepo.FindByID(ctx, input.ClinicID, input.BillingID); err != nil {
        return nil, apperrors.Wrap(err, "billing not found or belongs to different clinic")
    }

    // ✅ カテゴリバリデーション
    if err := validateItemCategory(string(input.Category)); err != nil {
        return nil, err
    }

    // ✅ TaxType デフォルト設定とバリデーション
    taxType := model.TaxTypeExcluded
    if input.TaxType != "" {
        if err := validateTaxType(string(input.TaxType)); err != nil {
            return nil, err
        }
        taxType = input.TaxType
    }

    // ✅ TaxRate デフォルト設定
    taxRate := 0.10
    if input.TaxRate > 0 {
        taxRate = input.TaxRate
    }

    // ✅ Source デフォルト設定とバリデーション
    source := model.ItemSourceManual
    if input.Source != "" {
        if err := validateItemSource(string(input.Source)); err != nil {
            return nil, err
        }
        source = input.Source
    }

    // 既存のバリデーション（Name, BillingID, UnitPrice）は維持
    if input.Name == "" {
        return nil, apperrors.WrapInvalidInput("name is required")
    }
    ...

    item := &model.BillingItem{
        ...
        TaxType: taxType,
        TaxRate: taxRate,
        Source:  source,
    }
    ...
}
```

### 3. `service/billing_item_service.go` — CreateBillingItemInput 型変更

Handler から raw string を受け取るため Input DTO の型を変更する:

```go
// 修正前
type CreateBillingItemInput struct {
    Category              model.ItemCategory
    TaxType               model.TaxType
    Source                model.ItemSource
    ...
}

// ✅ 修正後（raw string で受け取り Service 内で変換）
type CreateBillingItemInput struct {
    Category              string  // Service 内で model.ItemCategory に変換
    TaxType               string  // "" = デフォルト "excluded"
    Source                string  // "" = デフォルト "manual"
    ...
}
```

---

## 影響範囲

| ファイル | 対象 | 状態 |
|---------|------|------|
| `handler/billing_item_handler.go:26-70` | CreateBillingItem | ❌ ビジネスロジック混在 |
| `service/billing_item_service.go:87-128` | CreateItem | ロジック追加が必要 |
| `service/billing_item_service.go:26-38` | CreateBillingItemInput | 型変更が必要 |

---

## 準拠すべきプロジェクト規約

### `backend/CLAUDE.md` — 依存関係の方向

> handler (プレゼンテーション層) → service (ビジネスロジック層) → repository (データアクセス層)

Handler はリクエスト解析と Service への委譲のみを担う。ビジネスルール（テナント確認・デフォルト値・ドメインバリデーション）は Service 層が担う。

### `.claude/CLAUDE.md` — バックエンド・アーキテクチャ規約

> handler → service → repository の軽量レイヤードを徹底

### プロジェクト内参照実装

- `handler/vaccine_handler.go` — Handler がリクエスト解析のみ行い、バリデーションを Service に委譲
- `service/vaccine_service.go` — Service でバリデーションとデフォルト値処理を実施

---

## 関連ファイル

- `handler/billing_item_handler.go:26-70`
- `service/billing_item_service.go:87-128`
- `handler/vaccine_handler.go` — 参照実装
- `service/vaccine_service.go` — 参照実装
