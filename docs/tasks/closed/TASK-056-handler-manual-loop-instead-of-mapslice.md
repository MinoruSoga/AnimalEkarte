# TASK-056: procedure / cage handler — mapSlice 未使用の手動ループ

## 優先度

LOW

---

## 概要

`handler/slice_helpers.go` に `mapSlice` ジェネリック関数が定義されており、List 系レスポンス変換の共通ユーティリティとして全ハンドラが使うべき設計になっている。  
`procedure_handler.go` と `cage_handler.go` のみが `mapSlice` を使わず手動で `make + for range` ループを実装しており、冗長で不統一。

---

## 問題箇所

### procedure_handler.go（L44-48 概算）

```go
// ❌ 手動ループ（冗長）
resp := make([]procedureResponse, len(procedures))
for i := range procedures {
    resp[i] = toProcedureResponse(&procedures[i])
}
c.JSON(http.StatusOK, resp)
```

### cage_handler.go（L30-34 概算）

```go
// ❌ 手動ループ（冗長）
resp := make([]cageResponse, len(cages))
for i := range cages {
    resp[i] = toCageResponse(&cages[i])
}
c.JSON(http.StatusOK, resp)
```

---

## 修正方針

`mapSlice` ユーティリティ関数を使って 1 行に統一する。

```go
// ✅ 修正後 — procedure_handler.go
c.JSON(http.StatusOK, mapSlice(procedures, toProcedureResponse))

// ✅ 修正後 — cage_handler.go
c.JSON(http.StatusOK, mapSlice(cages, toCageResponse))
```

---

## 参照実装

`vaccine_handler.go` / `insurance_handler.go` / `exam_type_handler.go` / `hospitalization_plan_handler.go` がすべて `mapSlice` を使っている。

---

## 備考

`mapSlice` のシグネチャ: `func mapSlice[M, R any](items []M, f func(*M) R) []R`  
変換関数が `*T` を受け取る場合はそのまま渡せる。`T` を受け取る場合はラッパーが必要。
