---
description: テスト自動生成（ユニット・統合テスト）
argument-hint: "<path> (e.g. internal/service/owner_service.go)"
---

# /test-gen [path]

指定ファイルのテストケースを自動生成します。

## 使用法

```bash
# Backend Go テスト生成
/test-gen internal/service/owner_service.go

# Frontend React テスト生成
/test-gen features/owners/routes/OwnersList.tsx
```

## 生成内容

### Go (testify)
- Happy path テスト
- エッジケース（nil, empty, boundary）
- エラーケース
- Table-driven tests

### React (Vitest + RTL)
- Component rendering
- ユーザーインタラクション
- Props validation
- Error boundary

## 出力形式

```
Generated: XXX_test.go
Test cases: XX
Estimated coverage: XX%

// 自動生成されたテストコード
```

## 使用エージェント

`test-strategist` (Sonnet) を自動起動
