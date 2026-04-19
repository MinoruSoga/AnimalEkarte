# BUG-421: trimming_request の Update フィールドに `**float64` 2重ポインタを使用

## 概要

`trimming_request.go` の `updateTrimmingRequest` 構造体で、体重（BW）・体温（BT）フィールドに
`**float64`（2重ポインタ）を使用している。Go の慣例に反し、`*float64` で十分に表現できる。

## 問題箇所

```go
// trimming_request.go:37, 39
type updateTrimmingRequest struct {
    // ...
    BW  **float64 `json:"bw"`   // ← 2重ポインタ（不要）
    BT  **float64 `json:"bt"`   // ← 2重ポインタ（不要）
    // ...
}
```

## 対比：createTrimmingRequest

```go
// trimming_request.go の createTrimmingRequest
type createTrimmingRequest struct {
    BW  *float64 `json:"bw"`   // ← 単一ポインタ（正しい）
    BT  *float64 `json:"bt"`   // ← 単一ポインタ（正しい）
}
```

Create と Update で同一フィールドの型が異なる。

## 問題の根本原因

2重ポインタは「値なし（変更しない）」と「0.0（明示的にゼロ値を設定）」を区別する意図と推測されるが:

- `*float64` で nil = 送信なし、`&0.0` = ゼロ値を明示的に設定、と表現可能
- 2重ポインタは JSON デシリアライズ時の挙動が複雑になり、バグの温床となる
- Go の慣例（Effective Go）は `**T` を避けることを推奨している

## 修正方針

```go
// 修正後の updateTrimmingRequest
type updateTrimmingRequest struct {
    // ...
    BW  *float64 `json:"bw"`   // 単一ポインタに変更
    BT  *float64 `json:"bt"`   // 単一ポインタに変更
    // ...
}
```

Service Input の `UpdateTrimmingInput` 側も合わせて確認し、型を統一する。

## 影響ファイル

- `backend/internal/handler/trimming_request.go` — 行 37, 39
- `backend/internal/service/trimming_service.go` — `UpdateTrimmingInput` の BW/BT フィールド（確認要）

## 優先度

**Medium** — Go 慣例違反。JSON デシリアライズのバグリスクあり。
