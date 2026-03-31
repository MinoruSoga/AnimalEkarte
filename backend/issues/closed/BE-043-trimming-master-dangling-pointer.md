# BE-043: トリミングマスタ ダングリングポインタ修正

**Status**: Open
**Priority**: High
**Affects**: トリミングマスタ設定ページ (`/settings/trimming`)
**Date Created**: 2026-03-18
**Related**: TASK-016

## Summary

`toTrimmingCourseResponse()` 内でローカル変数へのポインタを返しているため、JSONマーシャル時にダングリングポインタアクセスが発生し、ランタイムエラーまたは不正なデータが返される。

## 現状のコード

```go
// backend/internal/handler/trimming_master_response.go:9-41
type trimmingCourseResponse struct {
	// ...
	TargetSize  *string   `json:"target_size,omitempty"`  // ← ポインタ型
	// ...
}

func toTrimmingCourseResponse(c *model.TrimmingCourse) trimmingCourseResponse {
	var targetSize *string
	if c.TargetSize != nil {
		s := string(*c.TargetSize)  // ← ローカル変数（スタック上）
		targetSize = &s              // ← ダングリングポインタ！
	}
	return trimmingCourseResponse{
		// ...
		TargetSize: targetSize,      // ← 失効したメモリアドレスを返す
		// ...
	}
}
```

**問題**: `s` は関数のスタックフレーム上のローカル変数。関数が return した後、`targetSize` が指すメモリは無効になる。JSONマーシャラーがこのアドレスを参照すると undefined behavior が発生する。

## 必要な変更

### 1. 構造体フィールド型を値型に変更

```go
// backend/internal/handler/trimming_master_response.go
// Before
type trimmingCourseResponse struct {
	TargetSize  *string   `json:"target_size,omitempty"`
}

// After
type trimmingCourseResponse struct {
	TargetSize  string    `json:"target_size,omitempty"`
}
```

### 2. 変換関数を簡略化

```go
// Before
func toTrimmingCourseResponse(c *model.TrimmingCourse) trimmingCourseResponse {
	var targetSize *string
	if c.TargetSize != nil {
		s := string(*c.TargetSize)
		targetSize = &s
	}
	return trimmingCourseResponse{ TargetSize: targetSize, ... }
}

// After
func toTrimmingCourseResponse(c *model.TrimmingCourse) trimmingCourseResponse {
	targetSize := ""
	if c.TargetSize != nil {
		targetSize = string(*c.TargetSize)
	}
	return trimmingCourseResponse{ TargetSize: targetSize, ... }
}
```

## フロントエンド影響

- 変更なし。フロントエンドは `target_size?: TargetSize` として受け取っており、`omitempty` により空文字列は JSON から省略されるため互換性は維持される。

## 完了条件

- [ ] `toTrimmingCourseResponse()` がポインタではなく値型で `TargetSize` を返している
- [ ] `docker compose exec backend go build ./cmd/api` パス
- [ ] `/settings/trimming` ページでコース一覧が正常に表示される
- [ ] コースの作成・編集（target_size 指定あり/なし）が正常動作する
