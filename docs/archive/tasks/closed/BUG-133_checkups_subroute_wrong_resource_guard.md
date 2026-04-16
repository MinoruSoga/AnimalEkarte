# BUG-133: カルテ内の定期健診サブリソースが checkups リソースでガードされている

## 概要
`POST /api/v1/medical-records/:id/checkups` が `medical-records` ではなく `checkups` リソースで
権限チェックされている。カルテのフルアクセス権限（medical-records: C=T, E=T, D=T）を持つユーザーが
checkups: C=F の場合、カルテ内の定期健診を作成できない。

## 脆弱性分類
- **設計上の不整合**（セキュリティバグではなく権限モデルの問題）
- **影響**: カルテ編集権限があるのに定期健診サブリソースを操作できない矛盾

## 再現手順
1. RBAC検証用グループ: `medical-records: C=T, E=T, D=T` / `checkups: C=F, E=F, D=F`
2. `POST /api/v1/medical-records/1/checkups` → **403 Forbidden**

## ブラウザテスト結果

| エンドポイント | 親リソース権限 | ガードリソース | 結果 |
|--------------|-------------|-------------|------|
| `POST /mr/:id/vitals` | medical-records E=T | medical-records | ✅ 通過 |
| `POST /mr/:id/treatments` | medical-records E=T | medical-records | ✅ 通過 |
| `POST /mr/:id/checkups` | medical-records E=T | **checkups** | ❌ 403 |
| `POST /mr/:id/billing-review/confirm` | — | accounting | ✅ 403（正しい） |

## 設計判断

これは「バグ」か「仕様」かの判断が必要:

**案A: medical-records リソースでガードする（子リソースは親に従う）**
- vitals, treatments, clinical-plan, treatment-plans, images, inquiries と統一
- カルテの編集権限があれば定期健診も操作可能
- **推奨**: 子リソースは親リソースの権限に従うのが一貫性がある

**案B: checkups リソースでガードする（現状維持）**
- 定期健診は独立したリソースとして管理
- カルテ内でも checkups 権限が必要
- billing-review が accounting リソースでガードされるのと同じパターン

## 現状コード

### `backend/internal/handler/checkup_handler.go`

```go
func (h *Handler) RegisterCheckupRoutes(rg *gin.RouterGroup) {
    rg.GET("/:id/checkups", h.ListCheckups)
    rg.POST("/:id/checkups",
        h.RequirePermission(string(model.ResourceCheckups), "create"),  // ← checkups リソース
        h.CreateCheckup)
    // ...
}
```

## 修正方針（案A を採用する場合）

```go
func (h *Handler) RegisterCheckupRoutes(rg *gin.RouterGroup) {
    rg.GET("/:id/checkups", h.ListCheckups)
    rg.POST("/:id/checkups",
        h.RequirePermission(string(model.ResourceMedicalRecords), "edit"),  // ← 親リソース
        h.CreateCheckup)
    // ...
}
```

## 優先度
**Low** — 権限モデルの設計判断。実運用で問題になる場合に対応。

## 関連ファイル
- `backend/internal/handler/checkup_handler.go` — RegisterCheckupRoutes
- `backend/internal/handler/handler.go` — ルート登録
