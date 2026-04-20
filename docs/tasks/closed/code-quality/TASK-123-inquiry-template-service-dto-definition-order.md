# TASK-123: `inquiry_template_service.go` — `UpdateInquiryTemplateInput` が `CreateInquiryTemplateInput` より前に定義 + 定数がメソッド実装後

## 優先度

**Low** — コードの一貫性の問題。機能には影響しない。

---

## 概要

`inquiry_template_service.go` において 2 つの定義順序の問題がある:

1. `UpdateInquiryTemplateInput`（行 17）が `CreateInquiryTemplateInput`（行 25）より前に定義されている
2. DB カラム定数（行 132 以降）がメソッド実装より後に定義されている

プロジェクト内の他のマスタサービスはすべて以下の統一順序に従っている:
```
CreateXxxInput → UpdateXxxInput → const → buildUpdateFields → interface → impl
```

---

## 問題箇所

### `service/inquiry_template_service.go:15-32`

```go
// ❌ UpdateDTO が CreateDTO より前に定義されている
type UpdateInquiryTemplateInput struct {  // 行 17
    Category  *string
    Title     *string
    Content   *string
    IsActive  *bool
    SortOrder *int
}

type CreateInquiryTemplateInput struct {  // 行 25（後に定義）
    Category  string
    Title     string
    Content   string
    IsActive  bool
    SortOrder int
}
```

### `service/inquiry_template_service.go:132` 以降（定数がメソッド実装後）

```go
// ❌ インターフェース（34行）・実装（43行）の後に定数が定義されている
type InquiryTemplateService interface { ... }  // 行 34
type inquiryTemplateService struct { ... }     // 行 43
func (s *inquiryTemplateService) List(...) {}  // 行 51+
// ...

const (  // 行 132 以降 — メソッド実装より後
    colInquiryTemplateXxx = "..."
    ...
)
func buildInquiryTemplateUpdateFields(...) { ... }
```

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ service/cage_service.go — Create → Update の順序
type CreateCageInput struct { ... }   // 先に Create
type UpdateCageInput struct { ... }   // 後に Update

// ✅ service/reservation_type_service.go — DTOs → const → builder → interface → impl
type CreateReservationTypeInput struct { ... }
type UpdateReservationTypeInput struct { ... }
const ( colReservationTypeName = "name" ... )
func buildReservationTypeUpdateFields(...) { ... }
type ReservationTypeService interface { ... }
```

---

## 修正方針

```
修正後の順序:
1. CreateInquiryTemplateInput（既存 25 行の内容を先に配置）
2. UpdateInquiryTemplateInput（既存 17 行の内容を後に移動）
3. const colInquiryTemplate*（行 132 以降から移動）
4. buildInquiryTemplateUpdateFields（移動）
5. InquiryTemplateService interface（既存 34 行）
6. impl（既存 43 行以降）
```

コードの移動のみ。ロジック変更なし。

---

## 影響範囲

| ファイル | 対象 | 状態 |
|---------|------|------|
| `service/inquiry_template_service.go:17` | UpdateDTO が CreateDTO より前 | ❌ 逆順 |
| `service/inquiry_template_service.go:132+` | const・builder がメソッド実装後 | ❌ 定義位置が後方 |

---

## 準拠すべきプロジェクト規約

### プロジェクト内参照実装

- `service/cage_service.go` — CreateDTO → UpdateDTO の正しい順序
- `service/reservation_type_service.go` — 完全に正しい順序の参照実装
- 関連タスク: TASK-120（occupation_service 同一問題）, TASK-116（checkup_type_service 同一問題）
