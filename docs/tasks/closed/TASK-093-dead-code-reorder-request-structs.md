# TASK-093: handler — 未使用の reorder リクエスト struct（デッドコード）

## 優先度

LOW

---

## 概要

複数の `*_request.go` ファイルに、実際のハンドラでは使用されていない
ドメイン固有の reorder リクエスト struct が定義されている。

全ドメインのハンドラは共通の `reorderRequest`（`slice_helpers.go:15`）を使用しており、
ドメイン固有 struct は完全なデッドコードである。

---

## 問題箇所

### vaccine_request.go:26-28

```go
// ❌ 定義されているが使用されていない
type reorderVaccineRequest struct {
    IDs []uint64 `json:"ids" binding:"required"`
}

// 実際のハンドラ（vaccine_handler.go:129）では共通 struct を使用
var req reorderRequest  // ← slice_helpers.go の共通型
```

### reservation_type_group_request.go:17-19

```go
// ❌ 定義されているが使用されていない
type reorderReservationTypeGroupRequest struct {
    IDs []uint64 `json:"ids" binding:"required"`
}

// 実際のハンドラ（reservation_type_group_handler.go:121）では共通 struct を使用
var req reorderRequest  // ← slice_helpers.go の共通型
```

### shift_template_handler.go:44-47

```go
// ❌ 定義されているが使用されていない（TASK-092 でも言及）
type reorderShiftTemplateRequest struct {
    IDs []uint64 `json:"ids" binding:"required"`
}

// 実際のハンドラ（shift_template_handler.go:223）では共通 struct を使用
var req reorderRequest  // ← slice_helpers.go の共通型
```

---

## 参照実装（共通 struct）

```go
// ✅ slice_helpers.go:13-17
// reorderRequest は全ドメイン共通の Reorder リクエスト struct。
type reorderRequest struct {
    IDs []uint64 `json:"ids" binding:"required,min=1"`
}
```

---

## 修正方針

各ファイルからデッドコードとなっている struct 定義を削除するのみ。
ハンドラの動作には影響しない。

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `handler/vaccine_request.go` | `reorderVaccineRequest` struct を削除 |
| `handler/reservation_type_group_request.go` | `reorderReservationTypeGroupRequest` struct を削除 |
| `handler/shift_template_handler.go` | `reorderShiftTemplateRequest` struct を削除（TASK-092 と同時対応可） |
