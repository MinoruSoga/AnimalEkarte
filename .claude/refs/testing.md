---
description: Testing standards (unit/integration tests)
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
- Test files in `__tests__/` directory within feature

### Feature Test Structure (bulletproof-react compliant)

Place test files at **same level** as target file (no `__tests__/` directory).

```
src/features/owners/
├── routes/
│   ├── OwnersList.tsx
│   └── OwnersList.test.tsx       ← same level
├── hooks/
│   ├── useOwnerForm.ts
│   └── useOwnerForm.test.ts      ← same level
└── api/
    ├── get-owners.ts
    └── get-owners.test.ts        ← same level
```

### Naming
- Test files: `*.test.ts` or `*.test.tsx`
- Test descriptions: Start with "should"

### React 19 Testing Notes
- `useActionState` testing: mock form action, verify state transitions
- `useOptimistic` testing: verify both optimistic update and rollback
- `ref` as prop: can test ref pass-through without forwardRef

### Running Tests
```bash
docker compose exec frontend pppnpm test:run    # Run all tests
docker compose exec frontend pppnpm test:coverage  # With coverage
```

---

## Coverage Requirements

- New features: Minimum 80% coverage
- Bug fixes: Add regression test

## Mocking

- Mock external dependencies
- Use dependency injection for testability
- Use interfaces for mockable dependencies (Go)
- Place feature-internal API mocks in `src/testing/`
