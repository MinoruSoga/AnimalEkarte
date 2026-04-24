# CODE-QUALITY-201: Handler 層のビジネスロジック分離（HIGH × 2件）

## 概要

2つのハンドラがビジネスロジック（ドメインモデル変換・サービスメソッド分岐）を持っており、
handler → service Input DTO の変換のみに徹する規約に違反している。

## 優先度

HIGH

## 影響ファイル

| ファイル | 問題箇所 | 重大度 |
|---------|---------|--------|
| `backend/internal/handler/diagnosis_handler.go` | L182-196 | HIGH |
| `backend/internal/handler/permission_group_handler.go` | L238-247 | HIGH |
| `backend/internal/service/diagnosis_service.go` | L197-203 | MEDIUM（派生） |
| `backend/internal/service/permission_group_service.go` | — | 修正対象 |

---

## 問題 1: diagnosis_handler.go:182-196 — Handler でのサービスメソッド分岐

### 現状コード

```go
// diagnosis_handler.go:186-196
if typeIDStr := c.Query("type_id"); typeIDStr != "" {
    catID, parseErr := strconv.ParseUint(typeIDStr, 10, 64)
    if parseErr != nil {
        RespondError(c, apperrors.WrapInvalidInput("invalid type_id"))
        return
    }
    names, total, err = h.svc.DiagnosisName.ListByCategoryID(ctx, clinicID, catID, page, limit)
} else {
    names, total, err = h.svc.DiagnosisName.List(ctx, clinicID, page, limit)
}
```

### 問題

`type_id` の有無に応じて `ListByCategoryID` / `List` を Handler が直接切り替えている。
「どのクエリロジックを使うか」はビジネスロジックであり Service 層が担うべき責務。

### 修正方針

`DiagnosisNameService.List` のシグネチャに `typeID *uint64` を追加し、
Service 内で分岐を吸収する。Handler は `strconv.ParseUint` の変換のみ担当。

```go
// handler — 修正後
var typeID *uint64
if s := c.Query("type_id"); s != "" {
    id, err := strconv.ParseUint(s, 10, 64)
    if err != nil {
        RespondError(c, apperrors.WrapInvalidInput("invalid type_id"))
        return
    }
    typeID = &id
}
names, total, err = h.svc.DiagnosisName.List(c.Request.Context(), clinicID, typeID, page, limit)

// service — 修正後
func (s *diagnosisNameService) List(ctx context.Context, clinicID uint64, typeID *uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
    if typeID != nil {
        return s.repo.FindByCategoryID(ctx, clinicID, *typeID, page, limit)
    }
    return s.repo.FindAll(ctx, clinicID, page, limit)
}
```

### 派生修正

`DiagnosisNameService` インターフェースから `ListByCategoryID` を削除し、
`List(typeID *uint64)` に統合後は外部公開不要なためインターフェースを縮小する。

---

## 問題 2: permission_group_handler.go:238-247 — model 変換が Handler に混入

### 現状コード

```go
// permission_group_handler.go:238-247
rules := make([]model.PermissionGroupRule, 0, len(req.Rules))
for _, r := range req.Rules {
    rules = append(rules, model.PermissionGroupRule{
        Resource:  r.Resource,
        CanView:   r.CanView,
        CanCreate: r.CanCreate,
        CanEdit:   r.CanEdit,
        CanDelete: r.CanDelete,
    })
}
h.svc.PermissionGroup.SetRules(ctx, id, rules, staffID)
```

### 問題

`model.PermissionGroupRule{}` の構築（Request DTO → Domain Model 変換）が Handler 内にある。
他ドメインは handler → service Input DTO 変換にとどめており、model 構築は service 内で行っている。

### 修正方針

```go
// service — SetPermissionGroupRulesInput DTO を追加
type SetPermissionGroupRulesInput struct {
    Resource  string
    CanView   bool
    CanCreate bool
    CanEdit   bool
    CanDelete bool
}

func (s *permissionGroupService) SetRules(ctx context.Context, groupID uint64,
    inputs []SetPermissionGroupRulesInput, actorStaffID uint64) error {
    rules := make([]model.PermissionGroupRule, 0, len(inputs))
    for _, inp := range inputs {
        rules = append(rules, model.PermissionGroupRule{
            Resource:  model.Resource(inp.Resource),
            CanView:   inp.CanView,
            CanCreate: inp.CanCreate,
            CanEdit:   inp.CanEdit,
            CanDelete: inp.CanDelete,
        })
    }
    // 以下は現行ロジックをそのまま移動
    ...
}

// handler — 修正後（Input DTO 変換のみ）
inputRules := make([]service.SetPermissionGroupRulesInput, 0, len(req.Rules))
for _, r := range req.Rules {
    inputRules = append(inputRules, service.SetPermissionGroupRulesInput{
        Resource:  r.Resource,
        CanView:   r.CanView,
        CanCreate: r.CanCreate,
        CanEdit:   r.CanEdit,
        CanDelete: r.CanDelete,
    })
}
if err := h.svc.PermissionGroup.SetRules(c.Request.Context(), id, inputRules, staffID); err != nil { ... }
```

---

## 規約参照

- `.claude/CLAUDE.md`: handler → service（Input DTO）→ repository の軽量レイヤードを徹底
- `.claude/rules/go-language.md`: インターフェース最小化

## テスト

- `diagnosis_handler.go`: `type_id` あり/なし両ケースでサービスが正しく呼ばれることを確認
- `permission_group_handler.go`: `SetRules` に Input DTO が正しく渡ることを確認
