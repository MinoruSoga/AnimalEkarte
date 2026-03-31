---
description: テスト規約（ユニット・統合テスト）
alwaysApply: false
globs: ["backend/**/*_test.go", "frontend/**/*.{test,spec}.{ts,tsx}"]
---

# Testing Rules

## Go (Backend)

### Test Structure
- Test files: `*_test.go` in same package
- Use table-driven tests for multiple cases
- Follow AAA pattern: Arrange, Act, Assert

### Naming
```go
func TestFunctionName_Scenario_ExpectedResult(t *testing.T)
func TestGetPatient_ValidID_ReturnsPatient(t *testing.T)
func TestGetPatient_InvalidID_ReturnsError(t *testing.T)
```

### Table-Driven Tests
```go
func TestValidateInput(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid input", "test", false},
        {"empty input", "", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateInput(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Running Tests
```bash
go test ./...                    # Run all tests
go test -v ./...                 # Verbose output
go test -cover ./...             # With coverage
go test -race ./...              # Race detection
```

---

## TypeScript / React 19 (Frontend)

### Test Structure
- Use `describe` for grouping related tests
- Use `it` or `test` for individual test cases
- Follow AAA pattern: Arrange, Act, Assert
- テストファイルはfeature内の`__tests__/`ディレクトリに配置

### Feature Test Structure（bulletproof-react準拠）

テストファイルは対象ファイルと**同階層**に配置する（`__tests__/` ディレクトリは使用しない）。

```
src/features/owners/
├── routes/
│   ├── OwnersList.tsx
│   └── OwnersList.test.tsx       ← 同階層
├── hooks/
│   ├── useOwnerForm.ts
│   └── useOwnerForm.test.ts      ← 同階層
└── api/
    ├── get-owners.ts
    └── get-owners.test.ts        ← 同階層
```

### Naming
- Test files: `*.test.ts` or `*.test.tsx`
- Test descriptions: Start with "should"

### React 19 Testing Notes
- `useActionState`のテスト: form actionをモックし、state遷移を検証
- `useOptimistic`のテスト: 楽観的更新とロールバックの両方を検証
- `ref` as prop: `forwardRef`なしでref受け渡しテスト可能

### Running Tests
```bash
docker compose exec frontend npm run test:run    # Run all tests
docker compose exec frontend npm run test:coverage  # With coverage
```

---

## Coverage Requirements

- New features: Minimum 80% coverage
- Bug fixes: Add regression test

## Mocking

- Mock external dependencies
- Use dependency injection for testability
- Use interfaces for mockable dependencies (Go)
- Feature内のAPIモックは`src/testing/`に共通モックを配置
