# TASK-041: エンドポイント命名規則違反 3件

## 優先度

MEDIUM

---

## 概要

プロジェクト命名規則（`.claude/rules/naming-conventions.md`）では「API パスはテーブル名と 1:1・kebab-case」と定めているが、以下の 3 件で HTTP メソッドまたはパスが不統一になっている。

---

## 問題 1: `/masters/examination-types` — テーブル名 `exam_types` と不一致

### ファイル
`backend/internal/handler/staff_handler.go:544-550`

### 問題
```go
masters.GET("/examination-types", h.ListExaminationTypes)
masters.POST("/examination-types", ...)
masters.PATCH("/examination-types/reorder", ...)
masters.GET("/examination-types/:id", h.GetExaminationType)
masters.PATCH("/examination-types/:id", ...)
masters.DELETE("/examination-types/:id", ...)
```

テーブル名は `exam_types`（`backend/internal/model/exam_type.go` 参照）であり、kebab-case 変換すると `/exam-types` になるべき。一方 Go ファイル名も `exam_type_handler.go`・`exam_type_service.go` と `exam_type` プレフィックス。API パスだけが `examination-types` と拡張されており、命名規則「テーブル名と 1:1」に違反する。

また、他ドメイン（`checkup-types`, `diagnosis-types`, `chief-complaint-types`）はテーブル名そのまま kebab-case を使用しており不統一。

### 修正案（優先度：フロントエンドへの影響を考慮して慎重に実施）
- パスを `/exam-types` に変更し、フロントエンドの API 呼び出しパスを合わせて修正する。
- または、明示的に「`examination-types` を正規パスとして採用する」と `naming-conventions.md` に記載し、テーブル名との乖離を許容することを意思決定として残す。

---

## 問題 2: `PATCH /masters/permission-groups` で Reorder — `/reorder` サフィックスなし

### ファイル
`backend/internal/handler/staff_handler.go:588`

### 問題
```go
masters.PATCH("/permission-groups", perm(...), h.ReorderPermissionGroups)
```

全マスタの Reorder エンドポイントは `PATCH /xxx/reorder` のパターンを採用しているが、permission-groups だけ `/reorder` サフィックスがない。

| ドメイン | Reorder パス |
|---------|-------------|
| animal-species | `PATCH /animal-species/reorder` |
| cages | `PATCH /cages/reorder` |
| vaccines | `PATCH /vaccines/reorder` |
| **permission-groups** | `PATCH /permission-groups` ← **`/reorder` なし** |
| occupation | `PATCH /occupations/reorder` |

さらに、`PATCH /permission-groups` が Reorder であるため、List（`GET /permission-groups`）と同パスで異なる操作になっており、API の意図が読みとりにくい。

### 修正案
```go
// 変更前
masters.PATCH("/permission-groups", perm(...), h.ReorderPermissionGroups)

// 変更後
masters.PATCH("/permission-groups/reorder", perm(...), h.ReorderPermissionGroups)
```

フロントエンドの呼び出し箇所も合わせて修正が必要。

---

## 問題 3: `POST /masters/merchandise-items/reorder` — Reorder に POST を使用

### ファイル
`backend/internal/handler/staff_handler.go:611`

### 問題
```go
masters.POST("/merchandise-items/reorder", perm(...), h.ReorderMerchandiseItems)
```

全マスタの Reorder が `PATCH` を使う中、merchandise-items だけ `POST` を使用している。

| ドメイン | Reorder HTTP メソッド |
|---------|---------------------|
| animal-species | `PATCH` |
| cages | `PATCH` |
| vaccines | `PATCH` |
| **merchandise-items** | `POST` ← **不統一** |
| permission-groups | `PATCH`（ただし問題 2 のパス違反あり） |

`POST` は「リソースの新規作成」の意味を持つため、並び替えという冪等操作に POST を使うのは REST セマンティクス上も不適切。

### 修正案
```go
// 変更前
masters.POST("/merchandise-items/reorder", ...)

// 変更後
masters.PATCH("/merchandise-items/reorder", ...)
```

フロントエンドの呼び出し箇所も合わせて修正が必要。
