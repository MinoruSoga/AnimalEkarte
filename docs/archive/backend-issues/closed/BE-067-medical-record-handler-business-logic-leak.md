# BE-067: generateRecordNo がhandler層に漏出 + math/rand 使用（セキュリティ・責務違反）

**Status**: Closed
**Priority**: High
**Affects**: `backend/internal/handler/medical_record_handler.go`
**Date Created**: 2026-03-26
**Related**: -

## Summary

`medical_record_handler.go` に `generateRecordNo()` と `generateRandomString()` というビジネスロジック関数が定義されており、handler 層の責務に違反している。さらに乱数生成に `math/rand` を使用しており暗号学的に安全ではない（カルテ番号の推測が容易になる）。これらは service 層に移動し、`crypto/rand` を使用するよう修正が必要。

## 現状のコード

```go
// backend/internal/handler/medical_record_handler.go:6
"math/rand"  // ← crypto/rand に変更すべき

// backend/internal/handler/medical_record_handler.go:17-32
func generateRecordNo(visitDate time.Time) string {
	dateStr := visitDate.Format("20060102")
	return dateStr + "-" + generateRandomString(6)
}

func generateRandomString(n int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]  // ← math/rand（予測可能）
	}
	return string(b)
}
```

`CreateMedicalRecord` ハンドラ（:106-243）は 137 行に肥大化しており、日付解決・record_no 生成・ID型変換・バリデーション・ClinicalPlan 更新まで handler 内で行っている。

## 必要な変更

### 1. handler から generateRecordNo を削除

`medical_record_handler.go` から `generateRecordNo`, `generateRandomString` を削除し、`math/rand` import を削除。

### 2. service 層に移動 + crypto/rand 使用

```go
// backend/internal/service/medical_record_service.go に追加

import (
	"crypto/rand"
	"math/big"
)

func generateRecordNo(visitDate time.Time) string {
	dateStr := visitDate.Format("20060102")
	return dateStr + "-" + generateRandomString(6)
}

func generateRandomString(n int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			// fallback: 0 を使う（エラー時の安全側への倒し方）
			b[i] = letters[0]
			continue
		}
		b[i] = letters[num.Int64()]
	}
	return string(b)
}
```

### 3. CreateMedicalRecord の handler スリム化

record_no 生成ロジックを service の `CreateMedicalRecord` 内に統合し、handler は request のバインドと response の返却のみを行う構造にする。

## 完了条件

- [ ] `handler/medical_record_handler.go` から `generateRecordNo`, `generateRandomString`, `math/rand` import を削除
- [ ] `service/medical_record_service.go` に `crypto/rand` を使った実装を追加
- [ ] `docker compose exec backend go test ./... -v` がパス
- [ ] `docker compose exec backend golangci-lint run ./...` がパス（`math/rand` 使用の警告なし）
