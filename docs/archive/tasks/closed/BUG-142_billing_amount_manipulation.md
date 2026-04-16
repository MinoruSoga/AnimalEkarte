# BUG-142: 会計データの金額をゼロ/負数に変更可能 + 請求超過リファンド

## 概要
`PATCH /api/v1/accountings/:id` で請求金額をゼロに変更可能。
また `POST /api/v1/accountings/:id/refunds` で請求額を超えるリファンドを作成可能。
金額の業務バリデーションが不足している。

## 脆弱性分類
- **CWE-20**: Improper Input Validation
- **影響**: 会計データの改ざん。請求額をゼロにして支払い回避、過剰リファンドで不正還付。

## 再現手順

### 1. 請求額をゼロに変更
```bash
curl -X PATCH /api/v1/accountings/1 \
  -H 'Content-Type: application/json' \
  -d '{"total_amount": 0}'
# → 200 OK
```

### 2. 負の請求額
```bash
curl -X PATCH /api/v1/accountings/1 \
  -H 'Content-Type: application/json' \
  -d '{"total_amount": -10000}'
# → 500 Internal Server Error
```

### 3. 請求額超過リファンド
```bash
# 請求額 ¥880 のレコードに ¥99,999,999 のリファンド
curl -X POST /api/v1/accountings/1/refunds \
  -H 'Content-Type: application/json' \
  -d '{"amount": 99999999}'
# → 201 Created ❌
```

## ブラウザテスト結果

| テスト | 期待 | 実際 |
|--------|------|------|
| total_amount = 0 | 400 | **200** ❌ |
| total_amount = -10000 | 400 | **500** ❌ |
| refund 99999999 (請求超過) | 400 | **201** ❌ |

## 期待する動作
- `total_amount` は正の数のみ許可（`> 0`）
- リファンド額は `total_amount - 既存リファンド合計` 以下
- 負の金額は 400 Bad Request

## 修正方針

### Service 層でバリデーション

```go
// accounting_service.go
func (s *AccountingService) Update(ctx context.Context, id uint64, input UpdateInput) error {
    if input.TotalAmount != nil && *input.TotalAmount < 0 {
        return apperrors.WrapInvalidInput("金額は0以上で指定してください")
    }
    // ...
}

func (s *AccountingService) CreateRefund(ctx context.Context, acctID uint64, amount int) error {
    acct, _ := s.repo.FindByID(ctx, acctID)
    existingRefunds, _ := s.repo.SumRefunds(ctx, acctID)
    if amount > acct.TotalAmount - existingRefunds {
        return apperrors.WrapInvalidInput("リファンド額が請求残高を超えています")
    }
    // ...
}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/api.md`
> "Include input validation on all endpoints"

金額フィールドの業務バリデーションが不足。

## 優先度
**Medium** — 不正な会計データが作成可能。業務データの整合性に影響。

## 関連ファイル
- `backend/internal/handler/accounting_handler.go` — UpdateAccounting
- `backend/internal/service/accounting_service.go` — バリデーション追加
