---
name: test-generation
description: テスト自動生成（ユニット・統合テスト、Go testify、React Testing Library）
---

# Test Generation & Automation

関数・メソッド・コンポーネントのテストケースを自動生成します。

## テスト作成の基本方針

### AAA (Arrange-Act-Assert) パターン

テストは Arrange（準備）→ Act（実行）→ Assert（検証）の3ブロックで構成する。

```typescript
it('should return expected result when given valid input', () => {
  // Arrange
  const input = 'test input'

  // Act
  const result = functionToTest(input)

  // Assert
  expect(result).toBe('expected output')
})
```

### Descriptive naming

テスト名は「何をしたら何が起きるか」を明示する（`should return X when Y` 形式）。`test1` や `it works` のような曖昧な名前は禁止。

### テストコマンド

| 種別 | ツール | スコープ限定（自動実行可） | 全体（ユーザー手動実行） |
|------|--------|--------------------------|------------------------|
| ユニット (FE) | Vitest + Testing Library + MSW | `docker compose exec frontend npx vitest run <spec>` | `pnpm test:run` |
| カバレッジ (FE) | Vitest | — | `pnpm test:coverage` |
| ユニット (BE) | go test + testify | `docker compose exec backend go test ./internal/<pkg>/... -v` | `go test ./... -v` |

全体コマンドは CLAUDE.md の自動実行禁止リスト — 生成後の全体確認はユーザーに実行を依頼する。

E2Eは `e2e-design` コマンド / `docs/ops/testing/E2E_TESTING_GUIDE.md` を参照

### モック方針

- モックは最小限に。過剰なモック化は実装との乖離を生む
- テストデータはファクトリパターンを使用
- 非同期テストは適切に await する

## 実行スコープ

### 1. Go Backend テスト生成

#### テスト構造（手書き fn-field モック — 本プロジェクトの正本パターン。詳細は `golang-testing` スキル参照）

testify/mock パッケージのマッチャー/期待値APIは使わない（本プロジェクトの実コードに1件も存在しない）。各テストケースで必要な fn だけ差し替える手書きモックが正本（実例: `backend/internal/reservation/liff_service_mock_test.go`）。詳細は `golang-testing`。

```go
func TestCreateOwner(t *testing.T) {
  tests := []struct {
    name      string
    input     *CreateOwnerInput
    createFn  func(ctx context.Context, owner *model.Owner) error
    wantError bool
    wantCode  string
  }{
    {
      name:  "happy path",
      input: &CreateOwnerInput{Name: "田中", Email: "test@example.com"},
      createFn: func(ctx context.Context, owner *model.Owner) error { return nil },
      wantError: false,
    },
    {
      name:      "invalid email",
      input:     &CreateOwnerInput{Email: "invalid"},
      wantError: true,
      wantCode:  "INVALID_INPUT",
    },
    {
      name:  "duplicate email",
      input: &CreateOwnerInput{Email: "existing@example.com"},
      createFn: func(ctx context.Context, owner *model.Owner) error { return apperrors.ErrDuplicate },
      wantError: true,
    },
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      repo := &mockOwnerRepository{createFn: tt.createFn}
      svc := NewOwnerService(repo)
      err := svc.CreateOwner(context.Background(), tt.input)
      if tt.wantError {
        require.Error(t, err)
        assert.Equal(t, tt.wantCode, err.Code)
      } else {
        require.NoError(t, err)
      }
    })
  }
}
```

#### 生成対象

**Happy path テスト:**
- 正常系：入力データが有効で、期待した結果が返される

**Edge case テスト:**
- Nil/Empty: `nil`、空文字、空スライス
- Boundary: 最小値、最大値、境界値
- 大文字小文字混在
- Unicode・絵文字

**Error case テスト:**
- バリデーションエラー
- Business logic エラー（重複、権限なし等）
- システムエラー（DB接続失敗等）

**Integration テスト:**
- Service + Repository + Database
- Context キャンセル、Timeout
- トランザクション成功・失敗

#### モック生成例（手書き fn-field — モック自動生成ツールは本プロジェクト未導入のため使わない）

```go
type OwnerRepository interface {
  Create(ctx context.Context, owner *Owner) error
  GetByID(ctx context.Context, id uint) (*Owner, error)
  Update(ctx context.Context, owner *Owner) error
}

// 手書き fn-field モック: 使う fn だけ差し替える（golang-testing スキル参照）
type mockOwnerRepository struct {
  createFn  func(ctx context.Context, owner *Owner) error
  getByIDFn func(ctx context.Context, id uint) (*Owner, error)
  updateFn  func(ctx context.Context, owner *Owner) error
}

func (m *mockOwnerRepository) Create(ctx context.Context, owner *Owner) error {
  if m.createFn != nil {
    return m.createFn(ctx, owner)
  }
  return nil
}
```

### 2. React Frontend テスト生成

#### Component Rendering テスト
```typescript
import { render, screen } from '@testing-library/react'
import { OwnerCard } from './OwnerCard'

describe('OwnerCard', () => {
  const mockOwner = {
    id: 1,
    name: '田中太郎',
    email: 'test@example.com',
    pets: []
  }

  test('renders owner name and email', () => {
    render(<OwnerCard owner={mockOwner} />)

    expect(screen.getByText('田中太郎')).toBeInTheDocument()
    expect(screen.getByText('test@example.com')).toBeInTheDocument()
  })

  test('displays pet count', () => {
    const ownerWithPets = { ...mockOwner, pets: [{ id: 1, name: 'Fluffy' }] }
    render(<OwnerCard owner={ownerWithPets} />)

    expect(screen.getByText('1 pets')).toBeInTheDocument()
  })
})
```

#### ユーザーインタラクション テスト
```typescript
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { OwnerForm } from './OwnerForm'

test('submits form with valid data', async () => {
  const handleSubmit = vitest.fn()
  const user = userEvent.setup()

  render(<OwnerForm onSubmit={handleSubmit} />)

  await user.type(screen.getByLabelText('Name'), '田中')
  await user.type(screen.getByLabelText('Email'), 'test@example.com')
  await user.click(screen.getByRole('button', { name: 'Submit' }))

  expect(handleSubmit).toHaveBeenCalledWith({
    name: '田中',
    email: 'test@example.com'
  })
})
```

#### Props バリデーションテスト
```typescript
test('renders with required props only', () => {
  render(<OwnerCard owner={mockOwner} />)
  expect(screen.getByText(mockOwner.name)).toBeInTheDocument()
})

test('handles optional onClick prop', async () => {
  const onClick = vitest.fn()
  const user = userEvent.setup()

  render(<OwnerCard owner={mockOwner} onClick={onClick} />)
  await user.click(screen.getByRole('button'))

  expect(onClick).toHaveBeenCalledWith(mockOwner.id)
})

test('raises error with invalid owner prop', () => {
  expect(() => {
    render(<OwnerCard owner={null} />)
  }).toThrow('owner is required')
})
```

#### Hook テスト
```typescript
import { renderHook, act } from '@testing-library/react'
import { useOwnerForm } from './use-owner-form' // hooks ファイルは kebab-case 命名

test('useOwnerForm initializes with default values', () => {
  const { result } = renderHook(() => useOwnerForm())

  expect(result.current.formData).toEqual({
    name: '',
    email: ''
  })
})

test('useOwnerForm updates formData on input', () => {
  const { result } = renderHook(() => useOwnerForm())

  act(() => {
    result.current.setFormData({ name: '田中', email: 'test@example.com' })
  })

  expect(result.current.formData.name).toBe('田中')
})

test('useOwnerForm validates email format', () => {
  const { result } = renderHook(() => useOwnerForm())

  act(() => {
    result.current.validate({ email: 'invalid-email' })
  })

  expect(result.current.errors.email).toBe('Invalid email format')
})
```

### 3. 統合テスト

#### API エンドポイント テスト
```go
func TestCreateOwnerAPI(t *testing.T) {
  suite := NewAPITestSuite(t)
  defer suite.Cleanup()

  req := httptest.NewRequest("POST", "/api/owners", bytes.NewReader([]byte(
    `{"name":"田中","email":"test@example.com"}`)))
  w := httptest.NewRecorder()

  suite.handler.CreateOwner(w, req)

  assert.Equal(t, http.StatusCreated, w.Code)

  var resp CreateOwnerResponse
  json.NewDecoder(w.Body).Decode(&resp)
  assert.NotZero(t, resp.ID)
}
```

### 4. テストカバレッジ

```bash
# Coverage レポート生成（スコープ限定 — 全体カバレッジ計測はユーザー手動実行。CLAUDE.md 禁止コマンド）
docker compose exec backend go test -coverprofile=coverage.out ./internal/<対象パッケージ>/...
docker compose exec backend go tool cover -html=coverage.out

# 目標: 80% 以上
```

## テスト生成チェックリスト

### Happy Path
- [ ] 正常系：有効入力で期待結果が返される
- [ ] すべての happy path シナリオをカバー

### Edge Cases
- [ ] Nil / 空 / ゼロ値
- [ ] 最小値 / 最大値
- [ ] 境界値
- [ ] 長い入力 / 短い入力
- [ ] Unicode / 特殊文字

### Error Cases
- [ ] バリデーションエラー（各フィールド）
- [ ] 業務ロジックエラー（重複、権限など）
- [ ] システムエラー（DB接続失敗など）
- [ ] Context キャンセル / Timeout

### Integration
- [ ] Repository + Service + Handler
- [ ] データベース操作
- [ ] トランザクション

## 出力形式

```markdown
## Test Generation Report

### Generated Tests
- **CreateOwner**: 12 test cases
  - Happy path: 1
  - Edge cases: 5
  - Error cases: 6

- **OwnerCard Component**: 8 test cases
  - Rendering: 2
  - Interactions: 3
  - Props validation: 3

### Coverage
- handler/owner_handler.go: 92%
- service/owner_service.go: 88%
- components/OwnerCard.tsx: 85%
- Overall: 88%

### Next Steps
1. Review generated tests
2. Add custom test cases if needed
3. Run tests: `docker compose exec backend go test ./internal/<対象パッケージ>/...`
   （⚠️ 全体 `go test ./...` は CLAUDE.md の自動実行禁止コマンド。ユーザーに手動実行を依頼する）
```

## 関連スキル

- `/tdd-workflow`（コマンド） - テスト駆動開発フロー
